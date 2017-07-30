package tile_surge

import (
    "fmt"
    "github.com/golang/protobuf/proto"
    "io/ioutil"
    "log"
    "math/rand"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "vector-tile/2.1"
)

func add_filemap(searchDir string, filemap map[string][]string, dirmap map[string]string) (map[string][]string, map[string]string) {
    fileList := []string{}
    err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
        fileList = append(fileList, path)
        return nil
    })
    _ = err
    fmt.Print(fileList)
    for _, file := range fileList {
        vals := strings.Split(file, "/")
        if len(vals) == 4 {
            dirval := strings.Join(vals[1:3], "/")
            dirmap["tiles/"+dirval] = ""
            filemap["tiles"+file[len(searchDir):]] = append(filemap["tiles/"+file[len(searchDir):]], file)

        }
    }

    return filemap, dirmap
}

// getting filename tile
func read_vt_layer(filename string) *vector_tile.Tile_Layer {
    in, _ := ioutil.ReadFile(filename)
    tile := &vector_tile.Tile{}
    if err := proto.Unmarshal(in, tile); err != nil {
        log.Fatalln("Failed to parse address book:", err)
    }
    if len(tile.Layers) == 1 {
        return tile.Layers[0]

    } else {
        return &vector_tile.Tile_Layer{}
    }
}

// combines the layer prefixs to form one large tile containing all layers
func Combine_Layer_Prefixs(prefixs []string) {
    dirmap := map[string]string{}
    filemap := map[string][]string{}
    for _, prefix := range prefixs {
        filemap, dirmap = add_filemap(prefix, filemap, dirmap)
    }

    // creating all the files
    for k := range dirmap {
        fmt.Print(k, "\n")
        os.MkdirAll(k, os.ModePerm)
    }
    c := make(chan string)

    for k, v := range filemap {
        go func(k string, v []string, c chan string) {
            tile := &vector_tile.Tile{}

            tile.Layers = []*vector_tile.Tile_Layer{}
            for _, filename := range v {
                //fmt.Print(len(read_vt_layer(filename).Tile_Value), "\n")
                tile.Layers = append(tile.Layers, read_vt_layer(filename))
            }

            pbfdata, _ := proto.Marshal(tile)
            ioutil.WriteFile(k, []byte(pbfdata), 0644)
            c <- ""
        }(k, v, c)
    }

    count := 0
    for range filemap {
        select {
        case msg1 := <-c:
            fmt.Printf("[%d/%d]%s\n", count, len(filemap), msg1)
        }
        count += 1
    }

}

func Shuffle(src []string) []string {
    dest := make([]string, len(src))
    perm := rand.Perm(len(src))
    for i, v := range perm {
        dest[v] = src[i]
    }
    return dest
}

func Values_Map(areas []int) map[string][]*vector_tile.Tile_Value {
    areas = Shuffle(areas)
    colors := []string{"#0030E5", "#0042E4", "#0053E4", "#0064E4", "#0075E4", "#0186E4", "#0198E3", "#01A8E3", "#01B9E3", "#01CAE3", "#02DBE3", "#02E2D9", "#02E2C8", "#02E2B7", "#02E2A6", "#03E295", "#03E184", "#03E174", "#03E163", "#03E152", "#04E142", "#04E031", "#04E021", "#04E010", "#09E004", "#19E005", "#2ADF05", "#3BDF05", "#4BDF05", "#5BDF05", "#6CDF06", "#7CDE06", "#8CDE06", "#9DDE06", "#ADDE06", "#BDDE07", "#CDDD07", "#DDDD07", "#DDCD07", "#DDBD07", "#DCAD08", "#DC9D08", "#DC8D08", "#DC7D08", "#DC6D08", "#DB5D09", "#DB4D09", "#DB3D09", "#DB2E09", "#DB1E09", "#DB0F0A"}
    mymap := map[string][]*vector_tile.Tile_Value{}
    for _, area := range areas {
        area := strconv.Iota(area)
        color := colors[rand.Intn(50)]
        mymap[area] = []*vector_tile.Tile_Value{Make_Tv_String(color)}
    }
    return mymap
}

func shit(filemap map[string][]string) map[string][]string {
    newfilemap := map[string][]string{}
    count := 0
    for k, v := range filemap {
        if count == 0 {
            count = 1
            newfilemap[k] = v
        }
    }
    filemap = newfilemap
    return filemap
}

// going to keep it simple in the beginning
// one prefix at a time, and just going to throw in like a colorkey on zips
// thinking after read in is done on each tile construct tile_value_map
// and see what values need added
// this repo will currently assume all tags are the same the same order
// this should probably be put somewhere else however its to much duplicated code to not do this
func Update_Layer_Values(values map[string][]*vector_tile.Tile_Value, prefix string, columns []string) {
    // getting filemap
    filemap := map[string][]string{}
    filemap, _ = add_filemap(prefix, filemap, map[string]string{})

    layer := read_vt_layer("wv/12/1107/1572")
    keys := layer.Keys
    fmt.Print(keys)
    var keyint int
    var colints []int
    for i, k := range keys {
        if ("gid" == k) || ("area" == k) {
            keyint = i*2 + 1
        }
        for _, col := range columns {
            if col == k {
                colints = append(colints, i*2+1)
            }
        }
    }
    fmt.Print(keyint, colints, "\n")

    //filemap = shit(filemap)

    //fmt.Print(keyint, colints)
    c := make(chan string)
    for _, v := range filemap {
        go func(v []string, keyint int, colints []int, c chan string) {
            if len(v) == 1 {
                in, _ := ioutil.ReadFile(v[0])
                tile := &vector_tile.Tile{}
                if err := proto.Unmarshal(in, tile); err != nil {
                    fmt.Print(v[0], "\n")
                    log.Fatalln("Failed to parse address book:", err)
                } // getting the tile values map
                //fmt.Print(len(tile.Layers), "\n")

                for len(tile.Layers) == 0 {
                    in, _ := ioutil.ReadFile(v[0])
                    tile = &vector_tile.Tile{}
                    if err := proto.Unmarshal(in, tile); err != nil {
                        fmt.Print(v[0], "\n")
                        log.Fatalln("Failed to parse address book:", err)
                    } // getting the tile values map
                }

                if len(tile.Layers) > 0 {
                    tile_values := tile.Layers[0].Values
                    tile_values_map := map[uint64]uint32{}
                    for i, tv := range tile_values {
                        var hash uint64
                        hash = Hash_Tv(tv)
                        tile_values_map[hash] = uint32(i)
                    }

                    features := []*vector_tile.Tile_Feature{}
                    keys := tile.Layers[0].Keys
                    for _, feat := range tile.Layers[0].Features {
                        if len(tile_values) != 0 {
                            tags := feat.Tags
                            //oldtags := tags
                            //fmt.Println(tags, colints)
                            //keys := tile.Layers[0].Keys
                            //fmt.Println(tags)
                            //fmt.Print(tags, keyint, "\n")
                            //fmt.Print(len(tile_values))
                            //fmt.Print(tile_values[tags[keyint]])
                            vals := values[tile_values[tags[keyint]].GetStringValue()]
                            //fmt.Print(vals, tile_values[tags[keyint]].GetStringValue(), "\n")
                            //for _, ii := range colints {
                            //    valval := tags[ii]
                            //    headval := tags[ii-1]
                            //    realvalval := tile_values[valval]
                            //    realheadval := keys[headval]
                            //    fmt.Print(realheadval, realvalval, "\n")
                            //}
                            //newtags := []uint32{}
                            for i, val := range vals {
                                h := Hash_Tv(val)
                                value, bool := tile_values_map[h]
                                fmt.Print(tags, tile_values[tags[colints[i]]], tile_values[tags[keyint]], "before\n")

                                if bool == true {
                                    tags[colints[i]] = value
                                    //fmt.Print(tags, tile_values[value], "done\n")
                                } else {
                                    tile_values = append(tile_values, val)
                                    tile_values_map[h] = uint32(len(tile_values) - 1)

                                    value = uint32(len(tile_values) - 1)
                                    tags[colints[i]] = value
                                }
                                fmt.Print(tags, tile_values[tags[colints[i]]], tile_values[tags[keyint]], "done\n")

                            }

                            //for _, ii := range colints {
                            //    valval := tags[ii]
                            //    headval := tags[ii-1]
                            //    realvalval := tile_values[valval]
                            //    realheadval := keys[headval]
                            //    fmt.Print(realheadval, realvalval, "\n")
                            //}
                            //fmt.Println(tags, "\n\n")
                            //fmt.Print(oldtags, tags, "\n")
                            feat.Tags = tags
                            features = append(features, feat)
                        } else {
                            features = append(features, feat)

                        }

                    }
                    //layer.Values = tile_values
                    tile.ProtoMessage()
                    tile = &vector_tile.Tile{}
                    layerVersion := uint32(15)
                    extent := vector_tile.Default_Tile_Layer_Extent
                    //var bound []Bounds
                    layername := prefix
                    tile.Layers = []*vector_tile.Tile_Layer{
                        {
                            Version:  &layerVersion,
                            Name:     &layername,
                            Extent:   &extent,
                            Features: features,
                            Values:   tile_values,
                            Keys:     keys,
                        },
                    }

                    pbfdata, _ := proto.Marshal(tile)

                    ioutil.WriteFile(v[0], pbfdata, 0666)

                } else {
                    //fmt.Print(v[0], "here\n\n\n\n\n")
                }
            }

            c <- ""
        }(v, keyint, colints, c)
    }

    // collecting the channel
    count := 0
    for range filemap {
        select {
        case msg1 := <-c:
            fmt.Printf("[%i/%i]%s\n", count, len(filemap), msg1)
        }
        count += 1
    }
}

// going to keep it simple in the beginning
// one prefix at a time, and just going to throw in like a colorkey on zips
// thinking after read in is done on each tile construct tile_value_map
// and see what values need added
// this repo will currently assume all tags are the same the same order
// this should probably be put somewhere else however its to much duplicated code to not do this
func Update_Layer_Values2(values map[string][]*vector_tile.Tile_Value, prefix string, columns []string) {
    // getting filemap
    filemap := map[string][]string{}
    filemap, _ = add_filemap(prefix, filemap, map[string]string{})

    for _, v := range filemap {
        in, _ := ioutil.ReadFile(v[0])
        tile := &vector_tile.Tile{}
        if err := proto.Unmarshal(in, tile); err != nil {
            fmt.Print(v[0], "\n")
            log.Fatalln("Failed to parse address book:", err)
        } // getting the tile values map
        pbfdata, _ := proto.Marshal(tile)
        ioutil.WriteFile("a", pbfdata, 0666)
    }

}

func unique(data []string) []string {
    mymap := map[string]string{}
    for _, i := range data {
        mymap[i] = ""
    }
    newdata := []string{}
    for i := range mymap {
        newdata = append(newdata, i)
    }
    return newdata
}

// going to keep it simple in the beginning
// one prefix at a time, and just going to throw in like a colorkey on zips
// thinking after read in is done on each tile construct tile_value_map
// and see what values need added
// this repo will currently assume all tags are the same the same order
// this should probably be put somewhere else however its to much duplicated code to not do this
func Update_Layer_Values3(values map[string][]*vector_tile.Tile_Value, prefix string, columns []string) {
    // getting filemap
    filemap := map[string][]string{}
    filemap, _ = add_filemap(prefix, filemap, map[string]string{})

    for _, v := range filemap {
        in, _ := ioutil.ReadFile(v[0])
        tile := &vector_tile.Tile{}
        if err := proto.Unmarshal(in, tile); err != nil {
            fmt.Print(v[0], "\n")
            log.Fatalln("Failed to parse address book:", err)
        } // getting the tile values map
        tile_values := tile.Layers[0].Values
        tilemap := map[string][]string{}
        for _, feat := range tile.Layers[0].Features {
            //fmt.Print(tile_values[feat.Tags[3]])
            value := *tile_values[feat.Tags[3]].StringValue
            key := *tile_values[feat.Tags[1]].StringValue
            //fmt.Print(string(tile_values[feat.Tags[1]].StringValue))
            tilemap[key] = append(tilemap[key], value)
        }

        for k, v := range tilemap {
            fmt.Print(k, unique(v), "\n")
        }
        if len(tilemap) == 0 {
            fmt.Print(tile.Layers[0].Features)
        }

        fmt.Print("done\n\n\n")
    }

}

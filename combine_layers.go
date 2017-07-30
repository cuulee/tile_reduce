package tile_surge

import (
    "fmt"
    "github.com/golang/protobuf/proto"
    "io/ioutil"
    "log"
    "math/rand"
    "os"
    "path/filepath"
    "reflect"
    "strings"
    "vector-tile/2.1"
)

// creates or appends a file map normally
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

// returns a filemap and a single file to open
// this prevents having to iterate thrugh the entire map again
func add_filemap_file(searchDir string, filemap map[string][]string, dirmap map[string]string) (map[string][]string, string) {
    fileList := []string{}
    err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
        fileList = append(fileList, path)
        return nil
    })
    _ = err
    var myfile string
    for _, file := range fileList {
        vals := strings.Split(file, "/")
        if len(vals) == 4 {
            dirval := strings.Join(vals[1:3], "/")
            dirmap["tiles/"+dirval] = ""
            filemap["tiles"+file[len(searchDir):]] = append(filemap["tiles/"+file[len(searchDir):]], file)
            myfile = file

        }
    }

    return filemap, myfile
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

// get update and key values
func get_key_values(filename string, columns []string) (int, string, []int, []string, string) {
    // reading in vector tile
    in, _ := ioutil.ReadFile(filename)
    tile := &vector_tile.Tile{}
    if err := proto.Unmarshal(in, tile); err != nil {
        log.Fatalln("Failed to parse address book:", err)
    }

    // getting the keys
    keys := tile.Layers[0].Keys

    // iterating through the key columns
    // identifying the positional that the desired values fall under
    var keyint int
    var keysint int
    var colints []int
    update_columns := []string{}
    for i, k := range keys {
        if ("gid" == k) || ("area" == k) {
            keyint = i*2 + 1
            keysint = i
        }
        for _, col := range columns {
            if col == k {
                colints = append(colints, i*2+1)
                update_columns = append(update_columns, keys[i])
            }
        }
    }

    // getting keyva
    keyval := keys[keysint]

    return keyint, keyval, colints, update_columns, *tile.Layers[0].Name
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

func Shuffle(src []interface{}) []interface{} {
    dest := make([]interface{}, len(src))
    perm := rand.Perm(len(src))
    for i, v := range perm {
        dest[v] = src[i]
    }
    return dest
}

func Values_Map(areas []interface{}) map[uint64][]*vector_tile.Tile_Value {
    areas = Shuffle(areas)
    colors := []string{"#0030E5", "#0042E4", "#0053E4", "#0064E4", "#0075E4", "#0186E4", "#0198E3", "#01A8E3", "#01B9E3", "#01CAE3", "#02DBE3", "#02E2D9", "#02E2C8", "#02E2B7", "#02E2A6", "#03E295", "#03E184", "#03E174", "#03E163", "#03E152", "#04E142", "#04E031", "#04E021", "#04E010", "#09E004", "#19E005", "#2ADF05", "#3BDF05", "#4BDF05", "#5BDF05", "#6CDF06", "#7CDE06", "#8CDE06", "#9DDE06", "#ADDE06", "#BDDE07", "#CDDD07", "#DDDD07", "#DDCD07", "#DDBD07", "#DCAD08", "#DC9D08", "#DC8D08", "#DC7D08", "#DC6D08", "#DB5D09", "#DB4D09", "#DB3D09", "#DB2E09", "#DB1E09", "#DB0F0A"}
    mymap := map[uint64][]*vector_tile.Tile_Value{}
    for _, area := range areas {
        color := colors[rand.Intn(50)]
        vv := reflect.ValueOf(area)
        kd := vv.Kind()
        var tv *vector_tile.Tile_Value
        if (reflect.Float64 == kd) || (reflect.Float32 == kd) {
            //fmt.Print(v, "float", k)
            tv = Make_Tv_Float(float64(vv.Float()))
            //hash = Hash_Tv(tv)
        } else if (reflect.Int == kd) || (reflect.Int8 == kd) || (reflect.Int16 == kd) || (reflect.Int32 == kd) || (reflect.Int64 == kd) || (reflect.Uint8 == kd) || (reflect.Uint16 == kd) || (reflect.Uint32 == kd) || (reflect.Uint64 == kd) {
            //fmt.Print(v, "int", k)
            tv = Make_Tv_Int(int(vv.Int()))
            //hash = Hash_Tv(tv)
        } else if reflect.String == kd {
            //fmt.Print(v, "str", k)
            tv = Make_Tv_String(string(vv.String()))
            //hash = Hash_Tv(tv)

        } else {
            tv := new(vector_tile.Tile_Value)
            t := ""
            tv.StringValue = &t
        }
        hash := Hash_Tv(tv)
        mymap[hash] = []*vector_tile.Tile_Value{Make_Tv_String(color)}
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
func Update_Layer_Values(values map[uint64][]*vector_tile.Tile_Value, prefix string, columns []string) {
    fmt.Print("Starting Layer Updates.\n")
    // getting filemap
    filemap := map[string][]string{}
    filemap, filename := add_filemap_file(prefix, filemap, map[string]string{})

    keyint, keyval, colints, update_columns, layername := get_key_values(filename, columns)

    //fmt.Print(keyint, colints)
    c := make(chan string)
    for _, v := range filemap {
        go func(v []string, keyint int, colints []int, c chan string) {
            if len(v) == 1 {

                // reading in the file
                in, _ := ioutil.ReadFile(v[0])
                tile := &vector_tile.Tile{}
                if err := proto.Unmarshal(in, tile); err != nil {
                    log.Fatalln("Failed to parse address book:", err)
                }

                // if for some reason the file is misread read again
                for len(tile.Layers) == 0 {
                    in, _ := ioutil.ReadFile(v[0])
                    tile = &vector_tile.Tile{}
                    if err := proto.Unmarshal(in, tile); err != nil {
                        log.Fatalln("Failed to parse address book:", err)
                    }
                }

                // making sure the file has to layers
                if len(tile.Layers) > 0 {
                    // getting tile values
                    tile_values := tile.Layers[0].Values

                    // creating the tile_values_map
                    tile_values_map := map[uint64]uint32{}
                    for i, tv := range tile_values {
                        var hash uint64
                        hash = Hash_Tv(tv)
                        tile_values_map[hash] = uint32(i)
                    }

                    // getting the keys of the Layer
                    keys := tile.Layers[0].Keys

                    // collecting each feature after the tags have been changed
                    features := []*vector_tile.Tile_Feature{}
                    for _, feat := range tile.Layers[0].Features {
                        if len(tile_values) != 0 {
                            // getting the tags within each feature
                            tags := feat.Tags

                            // hashing the keypos value determined to be the identifying
                            // feature field i.e. like area or gid
                            hash := Hash_Tv(tile_values[tags[keyint]])

                            // looking up value in map entered with in function
                            vals := values[hash]

                            // iterating through the values that will be changed in
                            // the tags secton of features
                            // so this is iterating through each value to change in a feature
                            // hashing that value looking it up to get the tile_value pos
                            // and setting tags pos to that tile value position
                            // sketchy looking at it but the only way to do it.
                            for i, val := range vals {
                                // hasing the value to be changed
                                h := Hash_Tv(val)

                                // checking to see if the value exists in the tile values map
                                value, bool := tile_values_map[h]

                                // if false add value to tile_values and tile_Values map
                                // and continue on setting the value pos in tags to the size -1
                                if bool == true {
                                    tags[colints[i]] = value
                                } else {
                                    tile_values = append(tile_values, val)
                                    tile_values_map[h] = uint32(len(tile_values) - 1)

                                    value = uint32(len(tile_values) - 1)
                                    tags[colints[i]] = value
                                }

                            }

                            // doing a set on tags to the modified verrsion and appending
                            feat.Tags = tags
                            features = append(features, feat)
                        } else {
                            features = append(features, feat)

                        }

                    }

                    // finally creating the final vector tile
                    tile = &vector_tile.Tile{}
                    layerVersion := uint32(15)
                    extent := vector_tile.Default_Tile_Layer_Extent
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

                    // as well as outputting the data
                    pbfdata, _ := proto.Marshal(tile)
                    ioutil.WriteFile(v[0], pbfdata, 0666)

                } else {
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
            fmt.Printf("\tLayer: %s, ID: %s,Updating: %s [%d/%d]%s\r", layername, keyval, update_columns, count, len(filemap), msg1)
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

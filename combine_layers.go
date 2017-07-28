package tile_surge

import (
    "fmt"
    "github.com/golang/protobuf/proto"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
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

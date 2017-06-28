package tile_reduce

import (
	//geo "github.com/paulmach/go.geo"
	"vector-tile/2.1"
	//"os"
	//"strings"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
)

// reading a vector file and parsing the structure back out of it
func Read_Vector_Tile_Index(filename string) map[string][]Yrow {
	// Read the existing address book.
	in, _ := ioutil.ReadFile(filename)
	tile := &vector_tile.Tile{}
	if err := proto.Unmarshal(in, tile); err != nil {
		log.Fatalln("Failed to parse address book:", err)
	}

	values := tile.Layers[0].Values
	fmt.Print(len(values))
	features := tile.Layers[0].Features
	tilemap := map[string][]Yrow{}
	tge := []uint32{}
	yrows := []Yrow{}
	for _, feat := range features {
		//fmt.Print(feat, "\n")
		ghash := values[feat.Tags[1]].GetStringValue()

		// getting geometries
		geoms := feat.Geometry

		// getting geometries in 3s
		if len(geoms) == 3 {
			// do shit
			fmt.Print(geoms)
			aa := Yrow{Range: []float64{values[geoms[0]].GetDoubleValue(), values[geoms[1]].GetDoubleValue()}, Area: values[geoms[2]].GetStringValue()}
			yrows = []Yrow{aa}
		} else {
			for _, g := range geoms {
				tge = append(tge, g)

				if len(tge) == 3 {
					// do shit
					yrows = append(yrows, Yrow{Range: []float64{values[geoms[0]].GetDoubleValue(), values[geoms[1]].GetDoubleValue()}, Area: values[geoms[2]].GetStringValue()})

					tge = []uint32{}
				}
			}
		}
		tilemap[ghash] = yrows
	}
	return tilemap
}

// performs the sequential operation necessary to create a vector tile index
// it is then written to a pbf mapbox vector tile
func Make_Vector_Tile_Index(tilemap map[string][]Yrow, outfilename string) {
	// getting the keys that will be used
	keys := []string{"GEOHASH"}

	// creating the tile values slice
	tile_values := []*vector_tile.Tile_Value{}

	// creating the stringmap slice
	stringmap := map[string]int{}

	// setting up the maps to collect
	// the unique strings as well as the unique floats
	uniquestrings := map[string]string{}
	uniquefloats := map[float64]string{}

	// setting current add geohash
	current := 0
	for k, v := range tilemap {
		// adding each geohash
		tv := new(vector_tile.Tile_Value)
		str := string(k)
		tv.StringValue = &str
		tile_values = append(tile_values, tv)

		// adding value to stringmap
		stringmap[k] = current

		current += 1

		// iterating through each output config
		for _, i := range v {
			uniquestrings[i.Area] = ""
			uniquefloats[float64(i.Range[0])] = ""
			uniquefloats[float64(i.Range[1])] = ""
		}

	}
	// now adding whats left over of the unique strings
	for k := range uniquestrings {
		tv := new(vector_tile.Tile_Value)

		str := string(k)
		tv.StringValue = &str
		tile_values = append(tile_values, tv)

		// adding value to stringmap
		stringmap[k] = current

		current += 1
	}

	// finally adding the unique floats in
	floatmap := map[float64]int{}
	//fmt.Print(len(uniquefloats))
	for k := range uniquefloats {
		tv := new(vector_tile.Tile_Value)
		flo := float64(k)
		tv.DoubleValue = &flo
		tile_values = append(tile_values, tv)

		floatmap[k] = current

		current += 1
	}

	// now creating each feature and shit

	features := []*vector_tile.Tile_Feature{}
	feat_type := vector_tile.Tile_POLYGON
	for k, v := range tilemap {
		feat := vector_tile.Tile_Feature{}
		feat.Tags = []uint32{0, uint32(stringmap[k])} // this takes of geohash / the geohash value
		feat.Type = &feat_type                        // adding the correct feature type

		// now iterating through each v value
		geom := []uint32{}
		for _, val := range v {
			geom = append(geom, uint32(floatmap[val.Range[0]]))
			geom = append(geom, uint32(floatmap[val.Range[1]]))
			geom = append(geom, uint32(stringmap[val.Area]))
		}
		// adding geom on
		feat.Geometry = geom

		features = append(features, &feat)
	}

	tile := &vector_tile.Tile{}
	layerVersion := vector_tile.Default_Tile_Layer_Version
	extent := vector_tile.Default_Tile_Layer_Extent
	//var bound []Bounds
	layername := "TileIndex"

	tile.Layers = []*vector_tile.Tile_Layer{
		{
			Version:  &layerVersion,
			Name:     &layername,
			Extent:   &extent,
			Values:   tile_values,
			Keys:     keys,
			Features: features,
		},
	}
	//fmt.Print(tile)
	//tv := new(vector_tile.Tile_Value)
	//tile.Layers[0].Features = features
	pbfdata, _ := proto.Marshal(tile)
	ioutil.WriteFile(outfilename, []byte(pbfdata), 0644)

}

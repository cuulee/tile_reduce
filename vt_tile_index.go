package tile_surge

import (
	//geo "github.com/paulmach/go.geo"
	"vector-tile/2.1"
	//"os"
	//"strings"
	//"fmt"
	"github.com/golang/protobuf/proto"
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	geo "github.com/paulmach/go.geo"
	"io/ioutil"
	"log"
	"strconv"
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
	//fmt.Print(len(values))
	features := tile.Layers[0].Features
	tilemap := map[string][]Yrow{}
	for _, feat := range features {
		//fmt.Print(i, len(features), feat, "\n")
		//fmt.Print(feat, "\n")
		ghash := values[feat.Tags[0]].GetStringValue()

		// getting geometries
		geoms := feat.Geometry

		tge := []uint32{}
		yrows := []Yrow{}

		//fmt.Print(len(geoms), "\n")
		// getting geometries in 3s
		if len(geoms) == 3 {
			// do shit
			//fmt.Print(geoms)
			aa := Yrow{Range: []float64{values[geoms[0]].GetDoubleValue(), values[geoms[1]].GetDoubleValue()}, Area: values[geoms[2]].GetStringValue()}
			yrows = []Yrow{aa}
		} else {
			for _, g := range geoms {
				tge = append(tge, g)

				if len(tge) == 3 {
					//fmt.Print(tge, "\n")
					// do shit
					yrows = append(yrows, Yrow{Range: []float64{values[tge[0]].GetDoubleValue(), values[tge[1]].GetDoubleValue()}, Area: values[tge[2]].GetStringValue()})

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
	//keys := []string{"GEOHASH"}

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
		feat.Tags = []uint32{uint32(stringmap[k])} // this takes of geohash / the geohash value
		feat.Type = &feat_type                     // adding the correct feature type

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
			Keys:     []string{},
			Features: features,
		},
	}
	//fmt.Print(tile)
	//tv := new(vector_tile.Tile_Value)
	//tile.Layers[0].Features = features
	pbfdata, _ := proto.Marshal(tile)
	ioutil.WriteFile(outfilename, []byte(pbfdata), 0644)

}

func Make_Line_Geohash(line []pc.Point, tile_values []*vector_tile.Tile_Value, tile_values_map map[uint64]uint32) ([]*vector_tile.Tile_Value, map[uint64]uint32, []uint32) {
	var geometry []uint32
	for _, pt := range line {
		ghash := geo.NewPoint(pt.X, pt.Y).GeoHash(12)
		tv := new(vector_tile.Tile_Value)
		tv = Make_Tv_String(string(ghash))
		hash := Hash_Tv(tv)
		//tile_slice = append(tile_slice, tv)
		onetag, ok := tile_values_map[hash]
		if ok == false {
			tile_values = append(tile_values, tv)
			tile_values_map[hash] = uint32(len(tile_values) - 1)
			geometry = append(geometry, uint32(len(tile_values)-1))
			//current += 1
		} else {
			geometry = append(geometry, onetag)
		}
	}

	return tile_values, tile_values_map, geometry
}

// performs the sequential operation necessary to create a vector tile index
// it is then written to a pbf mapbox vector tile
func Make_Vector_Tile_Line_Index(k m.TileID, lines []Line_Edge, feats []string, keysmap map[string]uint32) {
	// getting filename
	filename := "tiles/" + strconv.Itoa(int(k.X)) + "_" + strconv.Itoa(int(k.Y)) + "_" + strconv.Itoa(int(k.Z)) + ".pbf"
	tile_values := []*vector_tile.Tile_Value{}
	tile_values_map := map[uint64]uint32{}
	var tags, geometry []uint32
	// iterating through each line
	features := []*vector_tile.Tile_Feature{}
	feat_type := vector_tile.Tile_LINESTRING

	//fmt.Print(tile)
	for _, line := range lines {
		properties := tile_values_slice(line.Properties)

		// adding the properties to tile values map
		tile_values_map, tile_values, _, tags, _ = Tile_Values_Add_Feature3(properties, feats, tile_values_map, tile_values, uint32(0), keysmap)

		// now creating the Line
		tile_values, tile_values_map, geometry = Make_Line_Geohash(line.Line, tile_values, tile_values_map)

		// creating the individual features
		feat := vector_tile.Tile_Feature{}
		feat.Tags = tags // this takes of geohash / the geohash value
		feat.Type = &feat_type
		feat.Geometry = geometry

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
			Keys:     feats,
			Features: features,
		},
	}
	pbfdata, _ := proto.Marshal(tile)
	ioutil.WriteFile(filename, []byte(pbfdata), 0644)
}

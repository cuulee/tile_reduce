package tile_surge

import (
	geo "github.com/paulmach/go.geo"
	"vector-tile/2.1"
	//"os"
	//"strings"
	//"fmt"
	"github.com/golang/protobuf/proto"
	//h "github.com/mitchellh/hashstructure"
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	"io/ioutil"
	//"log"

	"strconv"
)

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

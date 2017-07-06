package tile_surge

import (
	"encoding/json"
	"fmt"
	//"github.com/golang/protobuf/proto"
	"github.com/jackc/pgx"
	_ "github.com/lib/pq"
	h "github.com/mitchellh/hashstructure"
	//"io/ioutil"

	"github.com/golang/protobuf/proto"
	l "github.com/murphy214/layersplit"
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"vector-tile/2.1"
)

// hashs a given tv structure
func Hash_Tv(tv *vector_tile.Tile_Value) uint64 {
	hash, _ := h.Hash(tv, nil)
	return hash
}

// makes a tile_value string
func Make_Tv_String(stringval string) *vector_tile.Tile_Value {
	tv := new(vector_tile.Tile_Value)
	t := string(stringval)
	tv.StringValue = &t
	return tv
}

// makes a tile value float
func Make_Tv_Float(val float64) *vector_tile.Tile_Value {
	tv := new(vector_tile.Tile_Value)
	t := float64(val)
	tv.DoubleValue = &t
	return tv
}

// makes a tile value int
func Make_Tv_Int(val int) *vector_tile.Tile_Value {
	tv := new(vector_tile.Tile_Value)
	t := int64(val)
	tv.SintValue = &t
	return tv
}

// adds a single geohash feature and maintains position.
func Tile_Values_Add_Feature(tile_values_map map[uint64]uint32, tile_values []*vector_tile.Tile_Value, current uint32, val map[string]interface{}, keysmap map[string]uint32) (map[uint64]uint32, []*vector_tile.Tile_Value, uint32, []uint32, []string) {
	tags := []uint32{}
	klist := []string{}
	for k, v := range val {
		var tv *vector_tile.Tile_Value
		var hash uint64
		boolval := false
		//fmt.Print(v)
		vv := reflect.ValueOf(v)
		kd := vv.Kind()
		if (reflect.Float64 == kd) || (reflect.Float32 == kd) {
			//fmt.Print(v, "float", k)
			tv = Make_Tv_Float(float64(vv.Float()))
			hash = Hash_Tv(tv)
			boolval = true

		} else if (reflect.Int == kd) || (reflect.Int8 == kd) || (reflect.Int16 == kd) || (reflect.Int32 == kd) || (reflect.Int64 == kd) || (reflect.Uint8 == kd) || (reflect.Uint16 == kd) || (reflect.Uint32 == kd) || (reflect.Uint64 == kd) {
			//fmt.Print(v, "int", k)
			tv = Make_Tv_Int(int(vv.Int()))
			hash = Hash_Tv(tv)
			boolval = true

		} else if reflect.String == kd {
			//fmt.Print(v, "str", k)
			tv = Make_Tv_String(string(vv.String()))
			hash = Hash_Tv(tv)
			boolval = true

		}
		if boolval == true {
			onetag, ok := tile_values_map[hash]
			if ok == false {
				tile_values_map[hash] = current
				tile_values = append(tile_values, tv)
				tags = append(tags, keysmap[k])
				tags = append(tags, current)
				current += 1
			} else {
				tags = append(tags, keysmap[k])
				tags = append(tags, onetag)
			}
		}
		klist = append(klist, k)
	}
	return tile_values_map, tile_values, current, tags, klist
}

// adds a single geohash feature and maintains position.
func tile_values_slice(vals []interface{}) []*vector_tile.Tile_Value {
	tile_slice := []*vector_tile.Tile_Value{}
	for _, v := range vals {
		var tv *vector_tile.Tile_Value
		//fmt.Print(v)
		vv := reflect.ValueOf(v)
		kd := vv.Kind()
		if (reflect.Float64 == kd) || (reflect.Float32 == kd) {
			//fmt.Print(v, "float", k)
			tv = Make_Tv_Float(float64(vv.Float()))
			//hash = Hash_Tv(tv)
			tile_slice = append(tile_slice, tv)
		} else if (reflect.Int == kd) || (reflect.Int8 == kd) || (reflect.Int16 == kd) || (reflect.Int32 == kd) || (reflect.Int64 == kd) || (reflect.Uint8 == kd) || (reflect.Uint16 == kd) || (reflect.Uint32 == kd) || (reflect.Uint64 == kd) {
			//fmt.Print(v, "int", k)
			tv = Make_Tv_Int(int(vv.Int()))
			//hash = Hash_Tv(tv)
			tile_slice = append(tile_slice, tv)
		} else if reflect.String == kd {
			//fmt.Print(v, "str", k)
			tv = Make_Tv_String(string(vv.String()))
			//hash = Hash_Tv(tv)
			tile_slice = append(tile_slice, tv)

		} else {
			tv := new(vector_tile.Tile_Value)
			t := ""
			tv.StringValue = &t
			tile_slice = append(tile_slice, tv)
		}
		//fmt.Print(tile_slice, tv, "\n")
	}
	return tile_slice
}

// adds a single geohash feature and maintains position.
func Tile_Values_Add_Feature2(tempval []*vector_tile.Tile_Value, keylist []string, tile_values_map map[uint64]uint32, tile_values []*vector_tile.Tile_Value, current uint32, keysmap map[string]uint32) (map[uint64]uint32, []*vector_tile.Tile_Value, uint32, []uint32, []string) {
	tags := []uint32{}
	klist := []string{}
	for i, tv := range tempval {
		var hash uint64
		k := keylist[i]
		hash = Hash_Tv(tv)
		onetag, ok := tile_values_map[hash]
		if ok == false {
			tile_values = append(tile_values, tv)

			tile_values_map[hash] = uint32(len(tile_values) - 1)
			tags = append(tags, keysmap[k])
			tags = append(tags, uint32(len(tile_values)-1))
			current = uint32(len(tile_values))
			//current += 1
		} else {
			tags = append(tags, keysmap[k])
			tags = append(tags, onetag)
		}
		//fmt.Print(tags, tile_values, "tags current \n")

		klist = append(klist, k)
	}
	return tile_values_map, tile_values, current, tags, klist
}

// NOTES on geomfield
// geom field is a field in your database that contains a geojson representation of an alignment in string format
// however this representation can also encompass multi polygons
// by separating each polygon with a '|'
// so a polygon field may look something liek this
// "[[[lng1,lat1],[lng2,lat1]]]|[[[lng1,lat1],[lng2,lat1]]]"\
type Config struct {
	Database  string   // the name of the database
	Tablename string   // the name of the table from the db
	Port      uint16   // port your sql instance is on
	Host      string   // host your sql instance uses
	GeomField string   // geometry field within your table
	ID        string   // id field will usually be gid in postgis database
	Fields    []string // fields in table to be properties in a feature
	Zooms     []int    // the zooms you'd like the tiles created at
}

// creates a new default config
func NewConfig(idfield string, geomfield string, database string, geomtype string, fields []string, zooms []int) Config {
	a := Config{ID: idfield,
		GeomField: geomfield,
		Database:  database,
		Tablename: database,
		Port:      5432,
		Host:      "localhost",
		Fields:    fields,
		Zooms:     zooms}
	return a
}

// currently only makes the a single line layer from a configuration struct
// this is simply my first actual time doing a complete function like this
// once i know how to make both individually ill look into combining them.
func Make_Tile_Layer_Line(config Config) {
	a := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     config.Host,
			Port:     config.Port,
			Database: config.Database,
			User:     "postgres",
		},
		MaxConnections: 1,
	}
	///a := pgx.ConnPoolConfig{ConnConfig: {pgx.ConnConfig{Host: "localhost", Port: uint16(5432), Database: "philly", User: "postgres"}}, MaxConnections: 10}
	var val map[string]interface{}

	if err := json.Unmarshal([]byte("{}"), &val); err != nil {
		panic(err)
	}

	var klist []string
	current := uint32(0)
	keys := []string{}
	keysmap := map[string]uint32{} // setting up the thing to execute querries off of
	p, _ := pgx.NewConnPool(a)
	var geom string
	var gid int
	var tempval map[string]interface{}

	// getting the gid and geometr2y field given within the config
	// this will be used to create the tilemaps going across each zoom.
	rows, _ := p.Query(fmt.Sprintf("SELECT %s,%s FROM %s", config.ID, config.GeomField, config.Database))
	var geoms [][]string
	for rows.Next() {
		_ = rows.Scan(&gid, &geom)
		gidstr := strconv.Itoa(gid)
		geoms = append(geoms, []string{gidstr, geom})
	}
	var vallist []map[string]interface{}
	fmt.Print(geoms, "\n")
	if len(config.Fields) != 0 {
		rows, _ = p.Query(fmt.Sprintf("SELECT %s,%s FROM %s", config.ID, strings.Join(config.Fields, ","), config.Database))

		// getting key maps and shit
		fdescs := rows.FieldDescriptions()
		for _, i := range fdescs {
			keys = append(keys, i.Name)
			keysmap[i.Name] = current
			current += 1
		}

		// setting up enviromentals that will be used to create tiles
		current = uint32(0)
		var vv []int
		var vals []interface{}
		// getting the geometries so they can be passed in
		//features := []*vector_tile.Tile_Feature{}
		//feat_type := vector_tile.Tile_POLYGON
		for rows.Next() {
			tempval = val
			vv = []int{}
			vals, _ = rows.Values()
			for ii := range vals {
				if (vals[ii] != nil) && (fdescs[ii].Name == config.ID) {
					//fmt.Print(vals[ii])
					tempval[fdescs[ii].Name] = vals[ii]
					vv = append(vv, ii)

				}
			}
			vallist = append(vallist, tempval)
		}
	}
	c := make(chan string)
	for _, zoom := range config.Zooms {
		tempplist := vallist

		// this parrelizes the write out for each go function
		go func(geoms [][]string, tempplist []map[string]interface{}, zoom int, c chan string) {
			var tags []uint32
			//fmt.Print(len(tilemap), "\n")
			tilemap := Make_Tilemap_Lines(geoms, zoom)
			cc := make(chan string)
			for k, v := range tilemap {
				go func(k m.TileID, v []Line_Edge, cc chan string) {

					tile_values := []*vector_tile.Tile_Value{}
					tile_values_map := map[uint64]uint32{}
					current = uint32(0)
					// getting the filename location of the tile were building within
					filename := "tiles/" + strconv.Itoa(int(k.Z)) + "/" + strconv.Itoa(int(k.X)) + "/" + strconv.Itoa(int(k.Y))
					dir := "tiles/" + strconv.Itoa(int(k.Z)) + "/" + strconv.Itoa(int(k.X))
					os.MkdirAll(dir, os.ModePerm)

					// geometry initialization
					bd := m.Bounds(k)
					var geomtile []uint32
					geomtile = []uint32{}

					// feature initializaition
					features := []*vector_tile.Tile_Feature{}
					feat_type := vector_tile.Tile_LINESTRING
					var tempval map[string]interface{}

					for k, line := range v {
						// this is are tempvallist
						//fmt.Print(tempplist, "\n")
						if len(tempplist) != 0 {
							tempval = tempplist[k]
						}
						tile_values_map, tile_values, current, tags, klist = Tile_Values_Add_Feature(tile_values_map, tile_values, current, tempval, keysmap)
						// making the geometry
						geomtile, _ = Make_Line_Geom(Make_Coords(line.Line, bd), []int32{0, 0})

						if len(geomtile) != 0 {
							feat := vector_tile.Tile_Feature{}
							feat.Tags = tags       // this takes of geohash / the geohash value
							feat.Type = &feat_type // adding the correct feature type
							// now iterating through each v value
							// adding geom on
							feat.Geometry = geomtile
							features = append(features, &feat)

						}

					}
					//fmt.Print(tags, klist, "\n")
					//fmt.Print(k, "\n")

					tile := &vector_tile.Tile{}
					layerVersion := vector_tile.Default_Tile_Layer_Version
					extent := vector_tile.Default_Tile_Layer_Extent
					//var bound []Bounds
					layername := "lines"
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
					// writing out each tile
					pbfdata, _ := proto.Marshal(tile)
					ioutil.WriteFile(filename, []byte(pbfdata), 0644)
					cc <- ""
					//fmt.Print(tile, "\n")
				}(k, v, cc)

			}
			count := 0
			for count < len(tilemap) {
				select {
				case msg1 := <-cc:
					fmt.Printf("%s%d[%d/%d]\n", msg1, zoom, count, len(tilemap))
				}
				count += 1
			}
			c <- ""
		}(geoms, tempplist, zoom, c)
	}

	for range config.Zooms {
		select {
		case msg1 := <-c:
			fmt.Print(msg1)
		}
	}

}

// currently only makes the a single line layer from a configuration struct
// this is simply my first actual time doing a complete function like this
// once i know how to make both individually ill look into combining them.
func Make_Tile_Layer_Polygon(config Config) {
	a := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     config.Host,
			Port:     config.Port,
			Database: config.Database,
			User:     "postgres",
		},
		MaxConnections: 1,
	}
	///a := pgx.ConnPoolConfig{ConnConfig: {pgx.ConnConfig{Host: "localhost", Port: uint16(5432), Database: "philly", User: "postgres"}}, MaxConnections: 10}
	var val map[string]interface{}

	if err := json.Unmarshal([]byte("{}"), &val); err != nil {
		panic(err)
	}

	var klist []string
	current := uint32(0)
	keys := []string{}
	keysmap := map[string]uint32{} // setting up the thing to execute querries off of
	p, _ := pgx.NewConnPool(a)
	var geom string
	var gid string
	//var tempval map[string]interface{}

	// getting the gid and geometr2y field given within the config
	// this will be used to create the tilemaps going across each zoom.
	rows, _ := p.Query(fmt.Sprintf("SELECT %s,%s FROM %s", config.ID, config.GeomField, config.Database))
	var geoms [][]string
	for rows.Next() {
		_ = rows.Scan(&gid, &geom)
		geoms = append(geoms, []string{gid, geom})
	}
	//fmt.Print(geoms, "\n")
	//var vallist []map[string]interface{}
	//var valmap map[int]map[string]interface{}
	total_tile_values := [][]*vector_tile.Tile_Value{}
	if len(config.Fields) != 0 {
		rows, _ = p.Query(fmt.Sprintf("SELECT %s FROM %s", strings.Join(config.Fields, ","), config.Database))
		//fmt.Print(rows, "\n\n")
		// getting key maps and shit
		fdescs := rows.FieldDescriptions()
		for _, i := range fdescs {
			keys = append(keys, i.Name)
			keysmap[i.Name] = current
			current += 1
		}
		//fmt.Print(keysmap)

		// setting up enviromentals that will be used to create tiles
		current = uint32(0)
		//var vals []interface{}
		//var vallist []map[string]interface{}

		// getting the geometries so they can be passed in
		//features := []*vector_tile.Tile_Feature{}
		//feat_type := vector_tile.Tile_POLYGON
		//count := 0
		for rows.Next() {
			vals, _ := rows.Values()
			total_tile_values = append(total_tile_values, tile_values_slice(vals))
			//vallist = append(vallist, tempval)
		}
	}

	mymap := map[string]int{}
	for i, row := range geoms {
		mymap[row[0]] = i
	}
	//fmt.Print(mymap, "\n")
	//fmt.Print(total_tile_values, "\n")
	// makign the tile layer from the geometry gathered
	layer := l.Make_Layer(geoms, "STATES")

	//fmt.Print(layer[0], "\n")
	//fmt.Print(total_tile_values)
	c := make(chan string)
	//fmt.Print(layer, "layer")
	for _, zoom := range config.Zooms {
		tilemap := Make_Tilemap(layer, zoom)

		//fmt.Print(tilemap, "\n")
		// this parrelizes the write out for each go function
		go func(geoms [][]string, total_tile_values [][]*vector_tile.Tile_Value, zoom int, tilemap map[m.TileID][]int, c chan string) {
			//fmt.Print(len(tilemap), "\n")

			cc := make(chan string)
			for k, v := range tilemap {

				// getting polygons from the index positions in tilemap
				polys := []l.Polygon{}
				for _, i := range v {
					polys = append(polys, layer[i])
				}
				go func(k m.TileID, polys []l.Polygon, cc chan string) {
					var tags []uint32
					tile_values := []*vector_tile.Tile_Value{}
					tile_values_map := map[uint64]uint32{}
					current = uint32(0)
					// getting the filename location of the tile were building within
					filename := "tiles/" + strconv.Itoa(int(k.Z)) + "/" + strconv.Itoa(int(k.X)) + "/" + strconv.Itoa(int(k.Y))
					dir := "tiles/" + strconv.Itoa(int(k.Z)) + "/" + strconv.Itoa(int(k.X))
					os.MkdirAll(dir, os.ModePerm)

					// geometry initialization
					bd := m.Bounds(k)
					var geomtile []uint32
					geomtile = []uint32{}

					// feature initializaition
					features := []*vector_tile.Tile_Feature{}
					feat_type := vector_tile.Tile_POLYGON

					// getting alignment of current id layer
					boxpolygon := Make_Tile_Poly(k)

					// in this block of codee im iterating through v the index positons of polygons
					// fromt he original layer as well as the polygons themselves
					// this is beccause wwe still have to access the idnex pos to get the tempvals proerpties
					// for each polygon
					for _, polygon := range polys {
						// getting polygon
						// this is are tempvallist
						//fmt.Print(tempplist, "\n")

						tempval := total_tile_values[mymap[polygon.Area]]

						//fmt.Print(tempval, polygon.Area, polygon.Pos, mymap[polygon.Area], "\n")
						//fmt.Print(tempval, "\n")
						//fmt.Print(tempval, "tempval\n")
						// updating the features map
						//fmt.Print(current, len(tile_values), tile_values, "befoer\n")
						tile_values_map, tile_values, current, tags, klist = Tile_Values_Add_Feature2(tempval, config.Fields, tile_values_map, tile_values, current, keysmap)

						//fmt.Print(current, len(tile_values), tile_values, tags, "after\n")
						//fmt.Print(tags, "\n")
						// trimming the polygon by the box shit.
						polygon.Polygon = polygon.Polygon.Construct(pc.INTERSECTION, boxpolygon)

						// linting the polygon if contains more than two alignments
						polygons := Lint_Single_Polygon(polygon)

						// iterating through each polygon alignment
						for _, poly := range polygons {
							//fmt.Print(tags)
							// making the geometry
							geomtile, _ = Make_Polygon(Make_Coords_Polygon(poly.Polygon, bd), []int32{0, 0})
							//fmt.Print(tags, tile_values, "\n")
							if len(geomtile) != 0 {
								feat := vector_tile.Tile_Feature{}
								//fmt.Print(tile_values, "tabs\n")
								feat.Tags = tags       // this takes of geohash / the geohash value
								feat.Type = &feat_type // adding the correct feature type
								// now iterating through each v value
								// adding geom on
								feat.Geometry = geomtile
								features = append(features, &feat)
								//fmt.Print(tags, len(tile_values), "\n")

							}
						}
					}
					//fmt.Print(tags, klist, "\n")
					//fmt.Print(k, "\n")

					tile := &vector_tile.Tile{}
					layerVersion := uint32(15)
					extent := vector_tile.Default_Tile_Layer_Extent
					//var bound []Bounds
					layername := "lines"
					//fmt.Print(tile_values, "end\n")
					//fmt.Print("\n\n\n\n")
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

					// writing out each tile
					pbfdata, _ := proto.Marshal(tile)

					ioutil.WriteFile(filename, pbfdata, 0666)
					cc <- ""
					//fmt.Print(tile, "\n")
				}(k, polys, cc)

			}
			count := 0
			for count < len(tilemap) {
				select {
				case msg1 := <-cc:
					fmt.Printf("Size: %s%d [%d/%d]\n", msg1, zoom, count, len(tilemap))
				}
				count += 1
			}
			c <- ""
		}(geoms, total_tile_values, zoom, tilemap, c)
	}

	for range config.Zooms {
		select {
		case msg1 := <-c:
			fmt.Print(msg1)
		}
	}
}

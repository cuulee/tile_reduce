package tile_surge

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"strings"
)

type Config_Toml struct {
	Method string
	DB     DataSource `toml:"datasource"`
	L      Layers     `toml:"layers"`
	V      []Layers
}

// this data structure is representative of a single data set that will be pulled from multiple times
type DataSource struct {
	Name       string
	Host       string // postgis database host
	Port       int    // postgis database port
	Database   string // postgis database name
	User       string // postgis database user
	Password   string // postgis database password
	Tablename  string // tablename
	LayerType  string
	Coordbool  bool
	Geomheader string
}

// basically any filter were defaulting to hard sql querry
// easiest way to do it
type Layers struct {
	Layer  string
	Zooms  []int
	Fields []string
	Sql    string
}
type Total struct {
	DB     DataSource `toml:"datasource"`
	L      []Layers   `toml:"layers"`
	Method string     `toml:"method"`
	Zoom   int        `toml:"zoom"`
}

// evaluates the toml tile configurtion
func Eval_Toml(tomlname string) {

	// reading in the tomal tile
	config := read_toml(tomlname)

	// getting id
	prefixs := []string{}
	var id, sql string
	if config.DB.LayerType == "lines" {
		id = "gid"
	} else {
		id = "area"
	}

	// operation for if the db layer is linesx
	if config.DB.LayerType == "lines" {
		for _, L := range config.L {
			// getting where clause in sql_query
			if len(L.Sql) == 0 {
				sql = ""
			} else {
				if strings.Contains(L.Sql, "WHERE") {
					vals := strings.Split(L.Sql, "WHERE")
					sql = "WHERE" + vals[len(vals)-1]
				} else {
					vals := strings.Split(L.Sql, "where")
					sql = "WHERE" + vals[len(vals)-1]
				}
			}

			// getting config
			dbconfig := Config{
				Database:  config.DB.Database,     // the name of the database
				Tablename: config.DB.Tablename,    // the name of the table from the db
				Port:      uint16(config.DB.Port), // port your sql instance is on
				Host:      config.DB.Host,         // host your sql instance uses
				GeomField: config.DB.Geomheader,   // geometry field within your table
				ID:        id,                     // id field will usually be gid in postgis database
				Fields:    L.Fields,               // fields in table to be properties in a feature
				Zooms:     L.Zooms,                // the zooms you'd like the tiles created at
				SQL_Query: sql,                    // raw sql however will only extract where clause
			}

			// getting line layer with some configuration.s
			layerconfig := Make_Layer_DB_Line(dbconfig)
			layerconfig.Prefix = L.Layer
			prefixs = append(prefixs, L.Layer)

			// logic for using the write method
			if config.Method == "vector" {
				// making the tilelayer for each line
				Make_Tile_Layer_Line(layerconfig)
			} else if config.Method == "index" {
				Make_Tile_Line_Index(layerconfig, config.Zoom)
			}
		}
	} else if config.DB.LayerType == "polygons" {
		for _, P := range config.L {
			// getting where clause in sql_query
			if len(P.Sql) == 0 {
				sql = ""
			} else {
				if strings.Contains(P.Sql, "WHERE") {
					vals := strings.Split(P.Sql, "WHERE")
					sql = "WHERE" + vals[len(vals)-1]
				} else {
					vals := strings.Split(P.Sql, "where")
					sql = "WHERE" + vals[len(vals)-1]
				}
			}

			// getting config
			dbconfig := Config{
				Database:  config.DB.Database,     // the name of the database
				Tablename: config.DB.Tablename,    // the name of the table from the db
				Port:      uint16(config.DB.Port), // port your sql instance is on
				Host:      config.DB.Host,         // host your sql instance uses
				GeomField: config.DB.Geomheader,   // geometry field within your table
				ID:        id,                     // id field will usually be gid in postgis database
				Fields:    P.Fields,               // fields in table to be properties in a feature
				Zooms:     P.Zooms,                // the zooms you'd like the tiles created at
				SQL_Query: sql,                    // raw sql however will only extract where clause
			}

			// getting line layer with some configuration.s
			layerconfig := Make_Layer_DB_Polygon(dbconfig)
			layerconfig.Prefix = P.Layer
			prefixs = append(prefixs, P.Layer)

			// logic for using the write method
			if config.Method == "vector" {
				// making the tilelayer for each line
				Make_Tile_Layer_Polygon(layerconfig)
			} else if config.Method == "index" {
				Make_Tile_Polygon_Index(layerconfig, config.Zoom)
			}
		}
	}
	//t.Combine_Layer_Prefixs(prefixs)
}

func read_toml(tomlname string) Total {
	var total Total
	if _, err := toml.DecodeFile(tomlname, &total); err != nil {
		fmt.Println(err)
	}
	return total
}

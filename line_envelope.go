package tile_surge

import (
	"encoding/json"
	"fmt"
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	"sort"
	"strconv"
	"strings"
)

// gets the slope of two pc.Points along a line
// if statement logic accounts for undefined corner case
func get_slope2(pt1 pc.Point, pt2 pc.Point) float64 {
	if pt1.X == pt2.X {
		return 1000000.0
	}
	return (pt2.Y - pt1.Y) / (pt2.X - pt1.X)
}

// pc.Point represents a pc.Point in space.
type Size2 struct {
	deltaX float64
	deltaY float64
}

// iteroplates the position of y based on x of the location between two pc.Points
// this function accepts m the slope to keep it from recalculating
// what could be several hundred/thousand times between two pc.Points
func interp2(pt1 pc.Point, pt2 pc.Point, x float64) pc.Point {
	m := get_slope2(pt1, pt2)
	y := (x-pt1.X)*m + pt1.Y
	return pc.Point{x, y}
}

type ResponseCoords2 struct {
	Coords [][][]float64 `json:"coords"`
}

// gets the coordstring into a slice the easiest way I'm aware of
func get_coords_json2(stringcoords string) [][][]float64 {
	stringcoords = fmt.Sprintf(`{"coords":%s}`, stringcoords)
	res := ResponseCoords2{}
	json.Unmarshal([]byte(stringcoords), &res)

	return res.Coords
}

func distance_pts(oldpt pc.Point, pt pc.Point) Size2 {
	return Size2{pt.X - oldpt.X, pt.Y - oldpt.Y}

}

func distance_bounds(bds m.Extrema) Size2 {
	return Size2{bds.E - bds.W, bds.N - bds.S}
}

func which_plane(oldpt pc.Point, pt pc.Point, oldbds m.Extrema) string {
	xs := []float64{oldpt.X, pt.X}
	sort.Float64s(xs)

	ys := []float64{oldpt.Y, pt.Y}
	sort.Float64s(ys)

	mybds := m.Extrema{W: xs[0], E: xs[1], S: ys[0], N: ys[1]}

	//if (pt.Y == 39.371616) || (oldpt.Y == 39.371616) {
	//	fmt.Print("here!\n")
	//	fmt.Print(mybds.N, mybds.S, "mybds\n")
	//	fmt.Print(oldbds.N, oldbds.S, "oldextrema\n")
	//}

	if (mybds.N >= oldbds.N) && (mybds.S <= oldbds.N) {
		return "north"
	} else if (mybds.N >= oldbds.S) && (mybds.S <= oldbds.S) {
		return "south"
	} else if (mybds.E >= oldbds.E) && (mybds.W <= oldbds.E) {
		return "east"
	} else if (mybds.E >= oldbds.W) && (mybds.W <= oldbds.W) {
		return "west"
	} else {
		return "NONE"
	}
}

// given a pc.Point checks to see if the given pt is within the correct bounds
func check_bounds(oldpt pc.Point, pt pc.Point, intersectpt pc.Point, oldbds m.Extrema) bool {
	if (intersectpt.X >= oldbds.W) && (intersectpt.X <= oldbds.E) && (intersectpt.Y >= oldbds.S) && (intersectpt.Y <= oldbds.N) && (check_bb(oldpt, pt, intersectpt) == true) {
		//fmt.Print(check_bb(oldpt, pt, intersectpt), "\n")
		return true
	} else {
		return false
	}
}

// finding the pc.Point that intersects with a given y
func opp_interp(pt1 pc.Point, pt2 pc.Point, y float64) pc.Point {
	m := get_slope2(pt1, pt2)
	x := ((y - pt1.Y) / m) + pt1.X
	return pc.Point{x, y}
}

func check_bb(oldpt pc.Point, pt pc.Point, intersectpt pc.Point) bool {
	xs := []float64{oldpt.X, pt.X}
	sort.Float64s(xs)

	ys := []float64{oldpt.Y, pt.Y}
	sort.Float64s(xs)
	//fmt.Print(xs, ys, "\n")
	if (intersectpt.X >= xs[0]) && (intersectpt.X <= xs[1]) && (intersectpt.Y >= ys[0]) && (intersectpt.Y <= ys[1]) {
		return true
	} else {
		return false
	}

}

// this function gets the intersection pc.Point with a bb box
// it also returns a string of the axis it intersected with
func get_intersection_pt(oldpt pc.Point, pt pc.Point, oldbds m.Extrema) (pc.Point, string) {
	trypt := interp2(oldpt, pt, oldbds.W)
	axis := "west"
	//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)

	if check_bounds(oldpt, pt, trypt, oldbds) == false {
		trypt = interp2(oldpt, pt, oldbds.E)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)

		axis = "east"
	}
	if check_bounds(oldpt, pt, trypt, oldbds) == false {
		trypt = opp_interp(oldpt, pt, oldbds.S)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
		axis = "south"
	}
	if check_bounds(oldpt, pt, trypt, oldbds) == false {
		trypt = opp_interp(oldpt, pt, oldbds.N)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)

		axis = "north"
	}
	if axis == "north" {
		trypt = pc.Point{0, 0}
	}

	return trypt, axis
}
func itersection_pt(oldpt pc.Point, pt pc.Point, oldbds m.Extrema, axis string) pc.Point {
	//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
	if axis == "west" {
		trypt := interp2(oldpt, pt, oldbds.W)
		return trypt
	} else if axis == "east" {
		trypt := interp2(oldpt, pt, oldbds.E)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
		return trypt

	} else if axis == "south" {
		trypt := opp_interp(oldpt, pt, oldbds.S)
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
		return trypt

	} else if axis == "north" {
		trypt := opp_interp(oldpt, pt, oldbds.N)
		return trypt
		//fmt.Printf("%f,%f\n", trypt.X, trypt.Y)
	}
	return pc.Point{0, 0}
}

// convert the lines representing tile coods into lines readable by
// nlgeojson
func convert_tile_coords(total [][]pc.Point) {
	count := 0
	var totalstring []string
	for _, line := range total {
		totalstring = []string{}
		for _, pt := range line {
			totalstring = append(totalstring, fmt.Sprintf("[%f,%f]", pt.X, pt.Y))
		}
		//fmt.Printf(`%d,"[%s]"`, count, strings.Join(totalstring, ","))
		//fmt.Print("\n")
		count += 1
	}

}

// preliminary logic for getting out polygons from said lines
// Each line should have the followoing
// bounds (maybenot)
// axis beg
// axis end
// tilestr
// will probably put in struct called meta something
// which will then be placed in two different maps (maybe
// if mapped lists maintain there order

// pc.Point represents a pc.Point in space.
type TileMeta struct {
	AxisS  string
	AxisE  string
	Bounds m.Extrema
}

// is the number even
func Even(number int) bool {
	return number%2 == 0
}

// is the number odd?
func Odd(number int) bool {
	return !Even(number)
}

// pc.Point in polygon derived from the xmap data structure
func Pip_simple(pt pc.Point, topmap map[string][]float64, latconst float64, size int) bool {
	ghash := m.Tile_Geohash(pt.X, latconst, size)
	ycollisions := topmap[ghash]
	//fmt.Print(ycollisions)
	y := 0.0
	count := 0
	oldy := 0.0
	for i := range ycollisions {
		y = ycollisions[i]
		if count == 0 {
			count = 1
		} else {
			if (oldy < pt.Y) && (y > pt.Y) && (Odd(count) == true) {
				return true
			}
			count += 1

		}
		oldy = y

	}
	return false
}

// checks a single geohash for its occurance in an alignment
// checks at each of the four corners of the geohash for its existance in the polygon
// *		* < testing these corners
//
//     -	<- this is the normal decode pc.Point
//
// *		* < testing these corners
func Check_single_ghash(ghash string, totalmap map[string][]float64, zoom int) (map[string]pc.Point, []string) {
	extrema := m.Bounds(m.Strtile(ghash))

	///fmt.Printf("%f,%f\n", extrema.e, extrema.n)
	//fmt.Printf("%f,%f\n", extrema.w, extrema.n)
	//fmt.Printf("%f,%f\n", extrema.e, extrema.s)
	//fmt.Printf("%f,%f\n", extrema.w, extrema.s)
	latconst := totalmap["latconst"][0]

	// checking each cornerr pc.Point
	boolupperright := Pip_simple(pc.Point{extrema.E - .0000001, extrema.N - .0000001}, totalmap, latconst, zoom)
	boolupperleft := Pip_simple(pc.Point{extrema.W + .0000001, extrema.N - .0000001}, totalmap, latconst, zoom)
	boollowerright := Pip_simple(pc.Point{extrema.E - .0000001, extrema.S + .0000001}, totalmap, latconst, zoom)
	boollowerleft := Pip_simple(pc.Point{extrema.W + .0000001, extrema.S + .0000001}, totalmap, latconst, zoom)

	//fmt.Print("upperleft:", boolupperleft, ", lowerright:", boollowerright, " lowerleft:", boollowerleft, " upperright:", boolupperright, "\n")
	geomshouldinclude := map[string]pc.Point{}
	keys := []string{}
	if boolupperright == true {
		geomshouldinclude["upperright"] = pc.Point{extrema.E, extrema.N}
		//fmt.Printf("%f,%f\n", extrema.E, extrema.N)
		keys = append(keys, "upperright")
	}

	if boolupperleft == true {
		geomshouldinclude["upperleft"] = pc.Point{extrema.W, extrema.N}
		keys = append(keys, "upperleft")

		//fmt.Printf("%f,%f\n", extrema.W, extrema.N)
	}

	if boollowerright == true {
		geomshouldinclude["lowerright"] = pc.Point{extrema.E, extrema.S}
		//fmt.Printf("%f,%f\n", extrema.E, extrema.S)
		keys = append(keys, "lowerright")

	}

	if boollowerleft == true {
		geomshouldinclude["lowerleft"] = pc.Point{extrema.W, extrema.S}
		keys = append(keys, "lowerleft")

		//fmt.Printf("%f,%f\n", extrema.W, extrema.S)
	}
	//fmt.Print(ghash, "    ", Get_XY(pt[0], pt[1], float64(currentsize)), "\n")
	return geomshouldinclude, keys
}

func opp_axis(val string) string {
	if val == "north" {
		return "south"
	} else if val == "south" {
		return "north"
	} else if val == "west" {
		return "east"
	} else if val == "east" {
		return "west"
	}

	return val
}

// Point represents a point in space.
type Line_Edge struct {
	Line       []pc.Point
	Gid        string
	Properties []interface{}
}

// functionifying this section so it doesnt get massive pretty decent break point
func make_edges(line Line, gid string, zoom int) map[m.TileID][]Line_Edge {
	count := 0
	var pt, oldpt, intersectpt pc.Point
	var tileid, oldtileid m.TileID
	var bds, oldbds m.Extrema
	var axis string
	var tilecoords []pc.Point
	//var totaltilecoords [][]Point
	//var mapmeta
	//var oldrow []float64
	//axisbeg := "hee"
	tilemap := map[m.TileID][]Line_Edge{}
	for _, row := range line.Line {
		pt = pc.Point{row[0], row[1]}
		tileid = m.Tile(pt.X, pt.Y, zoom)
		bds = m.Bounds(tileid)
		if count == 0 {
			count = 1
			tilecoords = append(tilecoords, pt)

		} else {
			// shit goes down here
			// getting the distances between two coordinate points
			dist := distance_pts(oldpt, pt)

			if tileid != oldtileid {
				bnddist := distance_bounds(bds)

				// if one of the distances violates or is greater than the distance
				// for bounds it will be sent into a tile creation function
				if (bnddist.deltaX < dist.deltaX) || (bnddist.deltaY < dist.deltaY) {
					// send to tile generation function
				} else {
					// otherwise handle normally finding the intersection point and adding in the
					// the end of tile coords
					//intersectpt, axis = get_intersection_pt(oldpt, pt, oldbds)
					axis = which_plane(oldpt, pt, oldbds)
					intersectpt = itersection_pt(oldpt, pt, oldbds, axis)
					tilecoords = append(tilecoords, intersectpt)
					tilemap[oldtileid] = append(tilemap[oldtileid], Line_Edge{Line: tilecoords, Gid: gid, Properties: line.Properties})

					tilecoords = []pc.Point{intersectpt}
					//axisbeg := opp_axis(axis)
				}
			} else {
				tilecoords = append(tilecoords, oldpt)

			}

			//fmt.Print(tilecoords, "\n")
			//fmt.Print(distance_pts(oldpt, pt), "\n")
			//fmt.Print(oldpt, oldtileid, "\n")

		}

		oldpt = pt
		oldtileid = tileid
		oldbds = bds
		//oldrow = row
	}
	tilecoords = append(tilecoords, oldpt)

	// account for if all points were in the square
	tilemap[oldtileid] = append(tilemap[oldtileid], Line_Edge{Line: tilecoords, Gid: gid, Properties: line.Properties})

	return tilemap
}

func translate(val string) string {
	if "upper" == val {
		return "north"
	} else if "lower" == val {
		return "south"
	} else if "left" == val {
		return "west"
	} else if "right" == val {
		return "east"
	}
	return val
}

func opp_translate(val string) string {
	if "north" == val {
		return "upper"
	} else if "south" == val {
		return "lower"
	} else if "west" == val {
		return "left"
	} else if "right" == val {
		return "east"
	}
	return val
}

func Get_string(align []pc.Point) string {
	newlist := []string{}
	for _, i := range align {
		newlist = append(newlist, fmt.Sprintf("[%f,%f]", i.X, i.Y))
	}
	return fmt.Sprintf("[%s]", strings.Join(newlist, ","))
}

// Point represents a point in space.
type Output_Map struct {
	Map map[m.TileID][]Line_Edge
}

// makes the tilemap datastructure for lines
func Make_Tilemap_Lines(data []Line, size int) map[m.TileID][]Line_Edge {
	c := make(chan Output_Map)
	counter := 0
	totalmap := map[m.TileID][]Line_Edge{}
	total := 0
	// iterating through each line in csv file
	for i, row := range data {
		go func(row Line, size int, c chan Output_Map) {
			istr := strconv.Itoa(i)
			c <- Output_Map{make_edges(row, istr, size)}
		}(row, size, c)

		if (counter == 50000) || (len(data)-1 == i) {
			count := 0
			total += counter
			fmt.Printf("[%d/%d] Getting tilemap, Size: %d\n", total, len(data), size)
			for count < counter {
				select {
				case msg1 := <-c:
					mymap := msg1.Map
					for k, v := range mymap {
						totalmap[k] = append(totalmap[k], v...)
					}
				}
				count += 1
			}
			counter = 0
		}
		counter += 1
	}
	return totalmap
}

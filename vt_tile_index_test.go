package tile_reduce

import (
	//t "./tile_reduce"
	"fmt"
	"github.com/mitchellh/hashstructure"
	"testing"
)

func Test_Vtile_Read_Write(t *testing.T) {
	// creating a Xmap to test
	val := map[string][]Yrow{}
	ymaps := []Yrow{{Range: []float64{46.55886, 46.72452892002719}, Area: "{'COUNTY':'30107','STATES':'3'}"}, {Range: []float64{46.724528442737444, 46.800059}, Area: "{'COUNTY':'30045','STATES':'3'}"}}
	val["c83f5pwy0"] = ymaps

	// creating teststruct map
	hash, _ := hashstructure.Hash(val, nil)
	//fmt.Print(hash)
	//a := Test_Struct{val}
	//testmap[a] = ""

	// writing vt
	Make_Vector_Tile_Index(val, "shit.pbf")

	// reading vt
	val2 := Read_Vector_Tile_Index("shit.pbf")
	//fmt.Print(val, val2)

	hash2, _ := hashstructure.Hash(val2, nil)
	if hash != hash2 {
		t.Error("The hash values of the structures created are different")
	}
	//k, ok := testmap[Test_Struct{val}]
	//fmt.Print(k, ok)

}

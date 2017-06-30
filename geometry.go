package tile_reduce

func Pos() []int32 {
	return []int32{0, 0}
}

func moverow(row []int32, geometry []uint32) []uint32 {
	geometry = append(geometry, moveTo(1))
	geometry = append(geometry, uint32(paramEnc(row[0])))
	geometry = append(geometry, uint32(paramEnc(row[1])))
	return geometry

}
func linerow(row []int32, geometry []uint32) []uint32 {
	geometry = append(geometry, uint32(paramEnc(row[0])))
	geometry = append(geometry, uint32(paramEnc(row[1])))
	return geometry

}

func cmdEnc(id uint32, count uint32) uint32 {
	return (id & 0x7) | (count << 3)
}

func moveTo(count uint32) uint32 {
	return cmdEnc(1, count)
}

func lineTo(count uint32) uint32 {
	return cmdEnc(2, count)
}

func closePath(count uint32) uint32 {
	return cmdEnc(7, count)
}

func paramEnc(value int32) int32 {
	return (value << 1) ^ (value >> 31)
}

func Make_Line_Geom(coords [][]int32, position []int32) ([]uint32, []int32) {
	var count uint32
	count = 0
	var geometry []uint32
	var oldrow []int32
	//total := map[uint32][]int32{}
	//var linetocount uint32
	linetocount := uint32(len(coords) - 1)

	for _, row := range coords {
		if count == 0 {
			geometry = moverow([]int32{row[0] - position[0], row[1] - position[1]}, geometry)
			geometry = append(geometry, lineTo(linetocount))

			count = 1
		} else {
			geometry = linerow([]int32{row[0] - oldrow[0], row[1] - oldrow[1]}, geometry)
		}
		oldrow = row
	}

	return geometry, oldrow
}

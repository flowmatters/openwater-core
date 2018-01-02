package conv

func IntsToUints(ints []int) []uint {
	result := make([]uint, len(ints))
	for i, v := range ints {
		result[i] = uint(v)
	}
	return result
}

func UintsToInts(uints []uint) []int {
	result := make([]int, len(uints))
	for i, v := range uints {
		result[i] = int(v)
	}
	return result
}

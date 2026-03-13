package m

// type NumT generic.Number

type Number interface {
	float32 | float64 | int32 | uint32 | int64 | uint64 | int | uint | uint8 | int8 | int16 | uint16
}

func Min[T Number](a, b T) T {
	if a > b {
		return b
	}

	return a
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}

	return b
}

package data

// Number is a type constraint for array types.
type Number interface {
	float32 | float64 | int32 | uint32 | int64 | uint64 | int | uint
}

// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/joelrahman/genny

package data

func ApplyFunc1Float64(dest, source NDFloat64, fn func(val float64) float64) {
	if dest.Contiguous() && source.Contiguous() {
		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] = fn(sourceSlice[i])
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, fn(source.Get(idx)))
		Increment(idx, shape)
	}

}

func ScaleFloat64Array(dest, source NDFloat64, scale float64) {
	ApplyFunc1Float64(dest, source, func(v float64) float64 { return v * scale })
}

func AddToFloat64Array(dest, source NDFloat64) {
	if dest.Contiguous() && source.Contiguous() {

		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] += sourceSlice[i]
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, dest.Get(idx)+source.Get(idx))
		Increment(idx, shape)
	}
}

func ApplyFunc1Float32(dest, source NDFloat32, fn func(val float32) float32) {
	if dest.Contiguous() && source.Contiguous() {
		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] = fn(sourceSlice[i])
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, fn(source.Get(idx)))
		Increment(idx, shape)
	}

}

func ScaleFloat32Array(dest, source NDFloat32, scale float32) {
	ApplyFunc1Float32(dest, source, func(v float32) float32 { return v * scale })
}

func AddToFloat32Array(dest, source NDFloat32) {
	if dest.Contiguous() && source.Contiguous() {

		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] += sourceSlice[i]
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, dest.Get(idx)+source.Get(idx))
		Increment(idx, shape)
	}
}

func ApplyFunc1Int32(dest, source NDInt32, fn func(val int32) int32) {
	if dest.Contiguous() && source.Contiguous() {
		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] = fn(sourceSlice[i])
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, fn(source.Get(idx)))
		Increment(idx, shape)
	}

}

func ScaleInt32Array(dest, source NDInt32, scale int32) {
	ApplyFunc1Int32(dest, source, func(v int32) int32 { return v * scale })
}

func AddToInt32Array(dest, source NDInt32) {
	if dest.Contiguous() && source.Contiguous() {

		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] += sourceSlice[i]
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, dest.Get(idx)+source.Get(idx))
		Increment(idx, shape)
	}
}

func ApplyFunc1Uint32(dest, source NDUint32, fn func(val uint32) uint32) {
	if dest.Contiguous() && source.Contiguous() {
		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] = fn(sourceSlice[i])
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, fn(source.Get(idx)))
		Increment(idx, shape)
	}

}

func ScaleUint32Array(dest, source NDUint32, scale uint32) {
	ApplyFunc1Uint32(dest, source, func(v uint32) uint32 { return v * scale })
}

func AddToUint32Array(dest, source NDUint32) {
	if dest.Contiguous() && source.Contiguous() {

		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] += sourceSlice[i]
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, dest.Get(idx)+source.Get(idx))
		Increment(idx, shape)
	}
}

func ApplyFunc1Int64(dest, source NDInt64, fn func(val int64) int64) {
	if dest.Contiguous() && source.Contiguous() {
		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] = fn(sourceSlice[i])
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, fn(source.Get(idx)))
		Increment(idx, shape)
	}

}

func ScaleInt64Array(dest, source NDInt64, scale int64) {
	ApplyFunc1Int64(dest, source, func(v int64) int64 { return v * scale })
}

func AddToInt64Array(dest, source NDInt64) {
	if dest.Contiguous() && source.Contiguous() {

		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] += sourceSlice[i]
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, dest.Get(idx)+source.Get(idx))
		Increment(idx, shape)
	}
}

func ApplyFunc1Uint64(dest, source NDUint64, fn func(val uint64) uint64) {
	if dest.Contiguous() && source.Contiguous() {
		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] = fn(sourceSlice[i])
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, fn(source.Get(idx)))
		Increment(idx, shape)
	}

}

func ScaleUint64Array(dest, source NDUint64, scale uint64) {
	ApplyFunc1Uint64(dest, source, func(v uint64) uint64 { return v * scale })
}

func AddToUint64Array(dest, source NDUint64) {
	if dest.Contiguous() && source.Contiguous() {

		destSlice := dest.Unroll()
		sourceSlice := source.Unroll()
		for i := range destSlice {
			destSlice[i] += sourceSlice[i]
		}

		return
	}

	idx := dest.NewIndex(0)
	shape := dest.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, dest.Get(idx)+source.Get(idx))
		Increment(idx, shape)
	}
}

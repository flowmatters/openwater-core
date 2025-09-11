package data

func ApplyFunc1[T Number](dest, source ND[T], fn func(val T) T) {
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

func ScaleArray[T Number](dest, source ND[T], scale T) {
	ApplyFunc1[T](dest, source, func(v T) T { return v * scale })
}

func AddToArray[T Number](dest, source ND[T]) {
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

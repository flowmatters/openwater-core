package data

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "ArrayType=float64,float32,int32,uint32,int64,uint64"

func ApplyFunc1ArrayType(dest, source NDArrayType, fn func(val ArrayType) ArrayType) {
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
	size := product(shape)
	for pos := 0; pos < size; pos++ {
		dest.Set(idx, fn(source.Get(idx)))
		increment(idx, shape)
	}

}

func ScaleArrayTypeArray(dest, source NDArrayType, scale ArrayType) {
	ApplyFunc1ArrayType(dest, source, func(v ArrayType) ArrayType { return v * scale })
}

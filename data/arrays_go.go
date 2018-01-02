package data

import (
	//	"fmt"
	"errors"
)

type ndFloat64 struct {
	ndFloat64Common
	Impl []float64
}

// func (nd *ndFloat64) getUnderlying(i int) float64 {
// 	return nd.Impl[i]
// }

// func (nd *ndFloat64) setUnderlying(i int, v float64) {
// 	nd.Impl[i] = v
// }

// func (nd *ndFloat64) takeImplementation(other NDFloat64) error {
// 	like, err := other.(*ndFloat64)
// 	if !err {
// 		return errors.New("Can't take implementation...")
// 	}
// 	nd.Impl = like.Impl
// 	return nil
// }

func (nd *ndFloat64) Get(loc []int) float64 {
	return nd.Impl[nd.Index(loc)]
}

func (nd *ndFloat64) Set(loc []int, val float64) {
	nd.Impl[nd.Index(loc)] = val
}

func (nd *ndFloat64) Slice(loc []int, dims []int, step []int) NDFloat64 {
	result := ndFloat64{}
	nd.slice(&result.ndFloat64Common, loc, dims, step)
	result.Impl = nd.Impl
	return &result
}

func (nd *ndFloat64) Apply(loc []int, dim int, step int, vals []float64) {
	sliceDim := nd.NewIndex(1)
	sliceDim[dim] = len(vals)
	sliceStep := nd.NewIndex(1)
	sliceStep[dim] = step
	slice := nd.Slice(loc, sliceDim, sliceStep)

	if slice.Contiguous() {
		concrete := slice.(*ndFloat64)
		implSlice := concrete.Impl
		subset := implSlice[concrete.Start : concrete.Start+len(vals)]
		copy(subset, vals)
	} else {
		start := loc[dim]
		for i, v := range vals {
			loc[dim] = start + i*step
			nd.Set(loc, v)
		}
		loc[dim] = start
	}
}

func (nd *ndFloat64) ApplySlice(loc []int, step []int, vals NDFloat64) {
	shape := vals.Shape()
	slice := nd.Slice(loc, shape, step)
	if slice.Contiguous() {
		copy(slice.Unroll(), vals.Unroll())
		return
	}

	idx := slice.NewIndex(0)
	size := product(shape)
	for pos := 0; pos < size; pos++ {
		slice.Set(idx, vals.Get(idx))
		increment(idx, shape)
	}
	// How to speed up
}

func (nd *ndFloat64) Unroll() []float64 {
	if nd.Contiguous() {
		s := nd.Start
		e := nd.Index(decrement(nd.Dims))
		return nd.Impl[s : e+1]
	}

	//	fmt.Println(nd)

	length := product(nd.Shape())
	res := make([]float64, length)

	dimOffsets := offsets(nd.Dims)
	//fmt.Println(dimOffsets)
	for i := 0; i < length; i++ {
		loc := idivMod(i, dimOffsets, nd.Dims)
		//		fmt.Println(i, loc, nd.Index(loc))
		//		fmt.Println(loc,i)
		res[i] = nd.Get(loc)
	}
	return res
}

func (nd *ndFloat64) ReshapeFast(newShape []int) (NDFloat64, error) {
	if !nd.Contiguous() {
		return nil, errors.New("Array not contiguous")
	}

	return nd.Reshape(newShape)
}

func (nd *ndFloat64) Reshape(newShape []int) (NDFloat64, error) {
	result := ndFloat64{}
	size := product(newShape)
	currentSize := product(nd.Shape())

	if size != currentSize {
		return nil, errors.New("Size mismatch")
	}

	reshapeToSeries := (len(newShape) == 1) && (maximum(nd.Shape()) == len(newShape))

	if nd.Contiguous() || !reshapeToSeries {
		result.Start = 0
		result.Impl = nd.Unroll()
		result.OriginalDims = newShape
		result.Dims = newShape
		result.Step = ones(len(newShape))

		result.Offset = offsets(newShape)
		return &result, nil
	}

	seriesDim := argmax(nd.Shape())
	// Special case 1D
	result.Start = nd.Start
	//result.takeImplementation(nd)
	result.Impl = nd.Impl
	result.OriginalDims = nd.OriginalDims
	result.Dims = newShape
	result.Step = []int{nd.Step[seriesDim]}
	result.Offset = []int{nd.Offset[seriesDim]}
	return &result, nil
}

func (nd *ndFloat64) MustReshape(newShape []int) NDFloat64 {
	result, e := nd.Reshape(newShape)
	if e != nil {
		panic(e.Error())
	}
	return result
}

func (nd *ndFloat64) Get1(loc int) float64 {
	var idx []int

	if len(nd.Dims) == 1 {
		idx = []int{loc}
	} else {
		idx = nd.NewIndex(0)
		for i := 0; i < len(nd.Dims); i++ {
			if nd.Dims[i] > 1 {
				idx[i] = loc
				break
			}
		}
		//		fmt.Println("nDims>1",idx,nd.Dims,loc)
	}
	return nd.Get(idx)
}

func (nd *ndFloat64) Set1(loc int, val float64) {
	nd.Set([]int{loc}, val)
}

func (nd *ndFloat64) Apply1(loc int, step int, vals []float64) {
	for i := 0; i < len(vals); i++ {
		nd.Set1(loc+i*step, vals[i])
	}
}

func (nd *ndFloat64) Get2(loc1 int, loc2 int) float64 {
	return nd.Get([]int{loc1, loc2})
}

func (nd *ndFloat64) Set2(loc1 int, loc2 int, val float64) {
	nd.Set([]int{loc1, loc2}, val)
}

func (nd *ndFloat64) Get3(loc1 int, loc2 int, loc3 int) float64 {
	return nd.Get([]int{loc1, loc2, loc3})
}

func (nd *ndFloat64) Set3(loc1 int, loc2 int, loc3 int, val float64) {
	nd.Set([]int{loc1, loc2, loc3}, val)
}

func NewArray(dims []int) NDFloat64 {
	return newArray(dims)
}

func newArray(dims []int) *ndFloat64 {
	result := ndFloat64{}
	size := product(dims)
	result.Start = 0
	result.Impl = make([]float64, size)
	result.OriginalDims = dims
	result.Dims = dims
	result.Step = ones(len(dims))
	result.Offset = offsets(dims)
	return &result
}

func NewArray1D(dim int) ND1Float64 {
	return newArray([]int{dim})
}

func NewArray2D(dim1 int, dim2 int) ND2Float64 {
	return newArray([]int{dim1, dim2})
}

func NewArray3D(dim1 int, dim2 int, dim3 int) ND3Float64 {
	return newArray([]int{dim1, dim2, dim3})
}

func ARange(n int) NDFloat64 {
	arr := NewArray([]int{n})
	idx := arr.NewIndex(0)
	for i := 0; i < n; i++ {
		idx[0] = i
		arr.Set(idx, float64(i))
	}
	return arr
}

func idivMod(numerator int, denominators []int, modulator []int) []int {
	res := make([]int, len(denominators))
	for i := range denominators {
		res[i] = (numerator / denominators[i]) % modulator[i]
	}
	return res
}

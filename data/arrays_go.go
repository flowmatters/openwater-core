package data

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "ArrayType=float64,float32,int32,uint32,int64,uint64,int,uint"

import (
	//	"fmt"
	"errors"

	"github.com/flowmatters/openwater-core/util/slice"
)

type ndArrayType struct {
	ndArrayTypeCommon
	Impl []ArrayType
}

// func (nd *ndArrayType) getUnderlying(i int) float64 {
// 	return nd.Impl[i]
// }

// func (nd *ndArrayType) setUnderlying(i int, v float64) {
// 	nd.Impl[i] = v
// }

// func (nd *ndArrayType) takeImplementation(other ndArrayType) error {
// 	like, err := other.(*ndArrayType)
// 	if !err {
// 		return errors.New("Can't take implementation...")
// 	}
// 	nd.Impl = like.Impl
// 	return nil
// }

func (nd *ndArrayType) Get(loc []int) ArrayType {
	return nd.Impl[nd.Index(loc)]
}

func (nd *ndArrayType) Set(loc []int, val ArrayType) {
	nd.Impl[nd.Index(loc)] = val
}

func (nd *ndArrayType) Slice(loc []int, dims []int, step []int) NDArrayType {
	result := ndArrayType{}
	nd.slice(&result.ndArrayTypeCommon, loc, dims, step)
	result.Impl = nd.Impl
	return &result
}

func (nd *ndArrayType) Apply(loc []int, dim int, step int, vals []ArrayType) {
	sliceDim := nd.NewIndex(1)
	sliceDim[dim] = len(vals)
	sliceStep := nd.NewIndex(1)
	sliceStep[dim] = step
	slice := nd.Slice(loc, sliceDim, sliceStep)

	if slice.Contiguous() {
		concrete := slice.(*ndArrayType)
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

func (nd *ndArrayType) ApplySlice(loc []int, step []int, vals NDArrayType) {
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

func (nd *ndArrayType) Unroll() []ArrayType {
	if nd.Contiguous() {
		s := nd.Start
		e := nd.Index(decrement(nd.Dims))
		return nd.Impl[s : e+1]
	}

	//	fmt.Println(nd)

	length := product(nd.Shape())
	res := make([]ArrayType, length)

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

func (nd *ndArrayType) ReshapeFast(newShape []int) (NDArrayType, error) {
	if !nd.Contiguous() {
		return nil, errors.New("Array not contiguous")
	}

	return nd.Reshape(newShape)
}

func (nd *ndArrayType) Reshape(newShape []int) (NDArrayType, error) {
	result := ndArrayType{}
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
		result.Step = slice.Ones(len(newShape))
		result.Offset = offsets(newShape)
		result.OffsetStep = multiply(result.Step, result.Offset)
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
	result.OffsetStep = multiply(result.Step, result.Offset)
	return &result, nil
}

func (nd *ndArrayType) MustReshape(newShape []int) NDArrayType {
	result, e := nd.Reshape(newShape)
	if e != nil {
		panic(e.Error())
	}
	return result
}

func (nd *ndArrayType) Get1(loc int) ArrayType {
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

func (nd *ndArrayType) Set1(loc int, val ArrayType) {
	nd.Set([]int{loc}, val)
}

func (nd *ndArrayType) Apply1(loc int, step int, vals []ArrayType) {
	for i := 0; i < len(vals); i++ {
		nd.Set1(loc+i*step, vals[i])
	}
}

func (nd *ndArrayType) Get2(loc1 int, loc2 int) ArrayType {
	return nd.Get([]int{loc1, loc2})
}

func (nd *ndArrayType) Set2(loc1 int, loc2 int, val ArrayType) {
	nd.Set([]int{loc1, loc2}, val)
}

func (nd *ndArrayType) Get3(loc1 int, loc2 int, loc3 int) ArrayType {
	return nd.Get([]int{loc1, loc2, loc3})
}

func (nd *ndArrayType) Set3(loc1 int, loc2 int, loc3 int, val ArrayType) {
	nd.Set([]int{loc1, loc2, loc3}, val)
}

func NewArrayArrayType(dims []int) NDArrayType {
	return newArrayArrayType(dims)
}

func newArrayArrayType(dims []int) *ndArrayType {
	result := ndArrayType{}
	size := product(dims)
	result.Start = 0
	result.Impl = make([]ArrayType, size)
	result.OriginalDims = dims
	result.Dims = dims
	result.Step = slice.Ones(len(dims))
	result.Offset = offsets(dims)
	result.OffsetStep = multiply(result.Step, result.Offset)
	return &result
}

func NewArray1DArrayType(dim int) ND1ArrayType {
	return newArrayArrayType([]int{dim})
}

func NewArray2DArrayType(dim1 int, dim2 int) ND2ArrayType {
	return newArrayArrayType([]int{dim1, dim2})
}

func NewArray3DArrayType(dim1 int, dim2 int, dim3 int) ND3ArrayType {
	return newArrayArrayType([]int{dim1, dim2, dim3})
}

func ARangeArrayType(n int) NDArrayType {
	arr := NewArrayArrayType([]int{n})
	idx := arr.NewIndex(0)
	for i := 0; i < n; i++ {
		idx[0] = i
		arr.Set(idx, ArrayType(i))
	}
	return arr
}

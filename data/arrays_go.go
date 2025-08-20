package data

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "[T]=float64,float32,int32,uint32,int64,uint64,int,uint"

import (
	//	"fmt"
	"errors"
	"iter"

	"github.com/flowmatters/openwater-core/util/slice"
)

type nd[T Number] struct {
	NdCommon[T]
	Impl []T
}

func (nd *nd[T]) Get(loc []int) T {
	return nd.Impl[nd.Index(loc)]
}

func (nd *nd[T]) Set(loc []int, val T) {
	nd.Impl[nd.Index(loc)] = val
}

func (nd *nd[T]) SetRaw(loc int, val T) {
	nd.Impl[loc] = val
}

func (nd *nd[T]) GetRaw(loc int) T {
	return nd.Impl[loc]
}

func (nd *nd[T]) Values(from []int, axis int, by int) iter.Seq[T] {
	if nd.Contiguous() {
		values := nd.Unroll()
		return func(yield func(T) bool) {
			for _, v := range values {
				if !yield(v) {
					return
				}
			}
		}
	}
	return func(yield func(T) bool) {
		index := nd.RawIndices(from, axis, by)
		for i := range index {
			if !yield(nd.GetRaw(i)) {
				return
			}
		}
	}
}

func (nd *nd[T]) RawIndices(from []int, axis int, by int) iter.Seq[int] {
	return func(yield func(int) bool) {
		i0 := nd.Index(from)
		offset := nd.OffsetStep[axis]
		for i := 0; i < by; i++ {
			if !yield(i0) {
				return
			}
			i0 += offset
		}
	}
}

func (ndArray *nd[T]) Slice(loc []int, dims []int, step []int) ND[T] {
	result := nd[T]{}
	ndArray.SliceInto(&result.NdCommon, loc, dims, step)
	result.Impl = ndArray.Impl
	return &result
}

func (ndArray *nd[T]) Apply(loc []int, dim int, step int, vals []T) {
	sliceDim := ndArray.NewIndex(1)
	sliceDim[dim] = len(vals)
	sliceStep := ndArray.NewIndex(1)
	sliceStep[dim] = step
	slice := ndArray.Slice(loc, sliceDim, sliceStep)

	if slice.Contiguous() {
		concrete := slice.(*nd[T])
		implSlice := concrete.Impl
		subset := implSlice[concrete.Start : concrete.Start+len(vals)]
		copy(subset, vals)
	} else {
		start := loc[dim]
		for i, v := range vals {
			loc[dim] = start + i*step
			ndArray.Set(loc, v)
		}
		loc[dim] = start
	}
}

func (nd *nd[T]) ApplySlice(loc []int, step []int, vals ND[T]) {
	shape := vals.Shape()
	slice := nd.Slice(loc, shape, step)
	if slice.Contiguous() {
		copy(slice.Unroll(), vals.Unroll())
		return
	}

	idx := slice.NewIndex(0)
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		slice.Set(idx, vals.Get(idx))
		Increment(idx, shape)
	}
	// How to speed up
}

func (nd *nd[T]) CopyFrom(other ND[T]) {
	nd.ApplySlice(nd.NewIndex(0), nil, other)
}

func (nd *nd[T]) Unroll() []T {
	if nd.Contiguous() {
		s := nd.Start
		e := nd.Index(decrement(nd.Dims))
		return nd.Impl[s : e+1]
	}

	//	fmt.Println(nd)

	length := Product(nd.Shape())
	res := make([]T, length)

	dimOffsets := Offsets(nd.Dims)
	//fmt.Println(dimOffsets)
	for i := 0; i < length; i++ {
		loc := IDivMod(i, dimOffsets, nd.Dims)
		//		fmt.Println(i, loc, nd.Index(loc))
		//		fmt.Println(loc,i)
		res[i] = nd.Get(loc)
	}
	return res
}

func (nd *nd[T]) ReshapeFast(newShape []int) (ND[T], error) {
	if !nd.Contiguous() {
		return nil, errors.New("Array not contiguous")
	}

	return nd.Reshape(newShape)
}

func (ndArray *nd[T]) Reshape(newShape []int) (ND[T], error) {
	result := nd[T]{}
	size := Product(newShape)
	currentSize := Product(ndArray.Shape())

	if size != currentSize {
		return nil, errors.New("Size mismatch")
	}

	reshapeToSeries := (len(newShape) == 1) && (Maximum(ndArray.Shape()) == len(newShape))

	if ndArray.Contiguous() || !reshapeToSeries {
		result.Start = 0
		result.Impl = ndArray.Unroll()
		result.OriginalDims = newShape
		result.Dims = newShape
		result.Step = slice.Ones(len(newShape))
		result.Offset = Offsets(newShape)
		result.OffsetStep = Multiply(result.Step, result.Offset)
		return &result, nil
	}

	seriesDim := Argmax(ndArray.Shape())
	// Special case 1D
	result.Start = ndArray.Start
	//result.takeImplementation(nd)
	result.Impl = ndArray.Impl
	result.OriginalDims = ndArray.OriginalDims
	result.Dims = newShape
	result.Step = []int{ndArray.Step[seriesDim]}
	result.Offset = []int{ndArray.Offset[seriesDim]}
	result.OffsetStep = Multiply(result.Step, result.Offset)
	return &result, nil
}

func (ndArray *nd[T]) MustReshape(newShape []int) ND[T] {
	result, e := ndArray.Reshape(newShape)
	if e != nil {
		panic(e.Error())
	}
	return result
}

func (nd *nd[T]) Get1(loc int) T {
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

func (nd *nd[T]) Set1(loc int, val T) {
	nd.Set([]int{loc}, val)
}

func (nd *nd[T]) Apply1(loc int, step int, vals []T) {
	for i := 0; i < len(vals); i++ {
		nd.Set1(loc+i*step, vals[i])
	}
}

func (nd *nd[T]) Get2(loc1 int, loc2 int) T {
	return nd.Get([]int{loc1, loc2})
}

func (nd *nd[T]) Set2(loc1 int, loc2 int, val T) {
	nd.Set([]int{loc1, loc2}, val)
}

func (nd *nd[T]) Get3(loc1 int, loc2 int, loc3 int) T {
	return nd.Get([]int{loc1, loc2, loc3})
}

func (nd *nd[T]) Set3(loc1 int, loc2 int, loc3 int, val T) {
	nd.Set([]int{loc1, loc2, loc3}, val)
}

func (nd *nd[T]) Maximum() T {
	idx := nd.NewIndex(0)
	res := nd.Get(idx)

	shape := nd.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		v := nd.Get(idx)
		if v > res {
			res = v
		}
		Increment(idx, shape)
	}
	return res
}

func (nd *nd[T]) Minimum() T {
	idx := nd.NewIndex(0)
	res := nd.Get(idx)

	shape := nd.Shape()
	size := Product(shape)
	for pos := 0; pos < size; pos++ {
		v := nd.Get(idx)
		if v < res {
			res = v
		}
		Increment(idx, shape)
	}
	return res
}

func NewArray[T Number](dims []int) ND[T] {
	return newArray[T](dims)
}

func ArrayFromSlice[T Number](data []T, dims []int) ND[T] {
	return arrayFromSlice[T](data, dims)
}

func arrayFromSlice[T Number](data []T, dims []int) *nd[T] {
	result := nd[T]{}
	// size := Product(dims)
	result.Start = 0
	result.Impl = data
	result.OriginalDims = dims
	result.Dims = dims
	result.Step = slice.Ones(len(dims))
	result.Offset = Offsets(dims)
	result.OffsetStep = Multiply(result.Step, result.Offset)
	return &result
}

func newArray[T Number](dims []int) *nd[T] {
	size := Product(dims)
	impl := make([]T, size)
	return arrayFromSlice[T](impl, dims)
	// result := nd[T]{}
	// result.Start = 0
	// result.Impl = impl
	// result.OriginalDims = dims
	// result.Dims = dims
	// result.Step = slice.Ones(len(dims))
	// result.Offset = Offsets(dims)
	// result.OffsetStep = Multiply(result.Step, result.Offset)
	// return &result
}

func NewArray1D[T Number](dim int) ND1[T] {
	return newArray[T]([]int{dim})
}

func NewArray2D[T Number](dim1 int, dim2 int) ND2[T] {
	return newArray[T]([]int{dim1, dim2})
}

func NewArray3D[T Number](dim1 int, dim2 int, dim3 int) ND3[T] {
	return newArray[T]([]int{dim1, dim2, dim3})
}

func ARange[T Number](n int) ND[T] {
	arr := NewArray[T]([]int{n})
	idx := arr.NewIndex(0)
	for i := 0; i < n; i++ {
		idx[0] = i
		arr.Set(idx, T(i))
	}
	return arr
}

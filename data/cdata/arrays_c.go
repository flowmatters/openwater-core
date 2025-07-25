package cdata

import (
	"errors"
	"reflect"
	"unsafe"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/slice"
	"github.com/joelrahman/genny/generic/cgeneric"
)

type CArray[T data.Number] cgeneric.CNumber

type ndC[T data.Number] struct {
	data.NdCommon[T]
	//	Impl *C.double
	Impl *[1 << 30]CArray[T]
	//p2 := (*[1<<30]C.int)(unsafe.Pointer(p))
}

// func ArrayFromC(impl *C.double, shape []int) ND[float64] {
// 	res := ndC[T]{}
// 	res.Dims = shape
// 	res.OriginalDims = shape
// 	res.Start = 0
// 	res.Offset = uniform(len(shape), 0)
// 	res.Step = ones(len(shape))
// 	res.Impl = impl
// 	return &res
// }

func (nd *ndC[T]) Get(loc []int) T {
	return T(nd.Impl[nd.Index(loc)])
}

func (nd *ndC[T]) Set(loc []int, val T) {
	nd.Impl[nd.Index(loc)] = CArray[T](val)
}

func (nd *ndC[T]) Slice(loc []int, dims []int, step []int) data.ND[T] {
	result := ndC[T]{}
	nd.SliceInto(&result.NdCommon, loc, dims, step)
	result.Impl = nd.Impl
	return &result
}

func (nd *ndC[T]) Apply(loc []int, dim int, step int, vals []T) {
	sliceDim := nd.NewIndex(1)
	sliceDim[dim] = len(vals)
	sliceStep := nd.NewIndex(1)
	sliceStep[dim] = step
	//	slice := nd.Slice(loc, sliceDim, sliceStep)

	// if slice.Contiguous() {
	// 	concrete := slice.(*ndC[T])
	// 	implSlice := concrete.Impl
	// 	subset := implSlice[concrete.Start : concrete.Start+len(vals)]
	// 	copy(subset, vals)
	// } else {
	start := loc[dim]
	for i, v := range vals {
		loc[dim] = start + i*step
		nd.Set(loc, v)
	}
	loc[dim] = start
	// }
}

func (nd *ndC[T]) ApplySlice(loc []int, step []int, vals data.ND[T]) {
	shape := vals.Shape()
	slice := nd.Slice(loc, shape, step)

	idx := slice.NewIndex(0)
	size := data.Product(shape)
	for pos := 0; pos < size; pos++ {
		slice.Set(idx, vals.Get(idx))
		data.Increment(idx, shape)
	}
	// How to speed up
}

func (nd *ndC[T]) CopyFrom(other data.ND[T]) {
	nd.ApplySlice(nd.NewIndex(0), nil, other)
}

func (nd *ndC[T]) Unroll() []T {
	// if nd.Contiguous() {
	// 	s := nd.Start
	// 	e := nd.Index(decrement(nd.Dims))
	// 	return nd.Impl[s : e+1]
	// }

	//	fmt.Println(nd)

	length := data.Product(nd.Shape())
	res := make([]T, length)

	dimOffsets := data.Offsets(nd.Dims)
	//fmt.Println(dimOffsets)
	for i := 0; i < length; i++ {
		loc := data.IDivMod(i, dimOffsets, nd.Dims)
		//		fmt.Println(i, loc, nd.Index(loc))
		//		fmt.Println(loc,i)
		res[i] = nd.Get(loc)
	}
	return res
}

func (nd *ndC[T]) ReshapeFast(newShape []int) (data.ND[T], error) {
	if !nd.Contiguous() {
		return nil, errors.New("Array not contiguous")
	}

	return nd.Reshape(newShape)
}

func (nd *ndC[T]) Reshape(newShape []int) (data.ND[T], error) {
	result := ndC[T]{}
	size := data.Product(newShape)
	currentSize := data.Product(nd.Shape())

	if size != currentSize {
		return nil, errors.New("Size mismatch")
	}

	reshapeToSeries := (len(newShape) == 1) && (data.Maximum(nd.Shape()) == len(newShape))

	if nd.Contiguous() || !reshapeToSeries {
		result.Start = nd.Start
		result.Impl = nd.Impl
		result.OriginalDims = newShape
		result.Dims = newShape
		result.Step = slice.Ones(len(newShape))

		result.Offset = data.Offsets(newShape)
		result.OffsetStep = data.Multiply(result.Step, result.Offset)
		return &result, nil
	}

	seriesDim := data.Argmax(nd.Shape())
	// Special case 1D
	result.Start = nd.Start
	//result.takeImplementation(nd)
	result.Impl = nd.Impl
	result.OriginalDims = nd.OriginalDims
	result.Dims = newShape
	result.Step = []int{nd.Step[seriesDim]}
	result.Offset = []int{nd.Offset[seriesDim]}
	result.OffsetStep = data.Multiply(result.Step, result.Offset)
	return &result, nil
}

func (nd *ndC[T]) MustReshape(newShape []int) data.ND[T] {
	result, e := nd.Reshape(newShape)
	if e != nil {
		panic(e.Error())
	}
	return result
}

func (nd *ndC[T]) Get1(loc int) T {
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

func (nd *ndC[T]) Set1(loc int, val T) {
	nd.Set([]int{loc}, val)
}

func (nd *ndC[T]) Apply1(loc int, step int, vals []T) {
	for i := 0; i < len(vals); i++ {
		nd.Set1(loc+i*step, vals[i])
	}
}

func (nd *ndC[T]) Get2(loc1 int, loc2 int) T {
	return nd.Get([]int{loc1, loc2})
}

func (nd *ndC[T]) Set2(loc1 int, loc2 int, val T) {
	nd.Set([]int{loc1, loc2}, val)
}

func (nd *ndC[T]) Get3(loc1 int, loc2 int, loc3 int) T {
	return nd.Get([]int{loc1, loc2, loc3})
}

func (nd *ndC[T]) Set3(loc1 int, loc2 int, loc3 int, val T) {
	nd.Set([]int{loc1, loc2, loc3}, val)
}

func (nd *ndC[T]) Maximum() T {
	idx := nd.NewIndex(0)
	res := nd.Get(idx)

	shape := nd.Shape()
	size := data.Product(shape)
	for pos := 0; pos < size; pos++ {
		v := nd.Get(idx)
		if v > res {
			res = v
		}
		data.Increment(idx, shape)
	}
	return res
}

func (nd *ndC[T]) Minimum() T {
	idx := nd.NewIndex(0)
	res := nd.Get(idx)

	shape := nd.Shape()
	size := data.Product(shape)
	for pos := 0; pos < size; pos++ {
		v := nd.Get(idx)
		if v < res {
			res = v
		}
		data.Increment(idx, shape)
	}
	return res
}

func NewCArray[T data.Number](impl unsafe.Pointer, dims []int) data.ND[T] {
	return newCArray[T]((*[1 << 30]CArray[T])(impl), dims)
}

func newCArray[T data.Number](impl *[1 << 30]CArray[T], dims []int) *ndC[T] {
	result := ndC[T]{}
	//	size := product(dims)
	result.Start = 0
	result.Impl = impl
	result.OriginalDims = dims
	result.Dims = dims
	result.Step = slice.Ones(len(dims))
	result.Offset = data.Offsets(dims)
	result.OffsetStep = data.Multiply(result.Step, result.Offset)
	return &result
}

func makeCArrayForTest[T data.Number](shape []int) *ndC[T] {
	goArray := data.ARange[T](data.Product(shape)).MustReshape(shape)
	impl := goArray.Unroll()

	v := reflect.Indirect(reflect.ValueOf(&impl))
	slice := (*reflect.SliceHeader)(unsafe.Pointer(v.UnsafeAddr()))
	addr := (*[1 << 30]CArray[T])(unsafe.Pointer(slice.Data))

	return newCArray[T](addr, shape)
}

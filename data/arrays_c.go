package data

import (
	"C"
	"errors"
	"reflect"
	"unsafe"
)

type ndFloat64C struct {
	ndFloat64Common
	//	Impl *C.double
	Impl *[1 << 30]C.double
	//p2 := (*[1<<30]C.int)(unsafe.Pointer(p))
}

// func ArrayFromC(impl *C.double, shape []int) NDFloat64 {
// 	res := ndFloat64C{}
// 	res.Dims = shape
// 	res.OriginalDims = shape
// 	res.Start = 0
// 	res.Offset = uniform(len(shape), 0)
// 	res.Step = ones(len(shape))
// 	res.Impl = impl
// 	return &res
// }

func (nd *ndFloat64C) Get(loc []int) float64 {
	return float64(nd.Impl[nd.Index(loc)])
}

func (nd *ndFloat64C) Set(loc []int, val float64) {
	nd.Impl[nd.Index(loc)] = C.double(val)
}

func (nd *ndFloat64C) Slice(loc []int, dims []int, step []int) NDFloat64 {
	result := ndFloat64C{}
	nd.slice(&result.ndFloat64Common, loc, dims, step)
	result.Impl = nd.Impl
	return &result
}

func (nd *ndFloat64C) Apply(loc []int, dim int, step int, vals []float64) {
	sliceDim := nd.NewIndex(1)
	sliceDim[dim] = len(vals)
	sliceStep := nd.NewIndex(1)
	sliceStep[dim] = step
	//	slice := nd.Slice(loc, sliceDim, sliceStep)

	// if slice.Contiguous() {
	// 	concrete := slice.(*ndFloat64C)
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

func (nd *ndFloat64C) ApplySlice(loc []int, step []int, vals NDFloat64) {
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

func (nd *ndFloat64C) Unroll() []float64 {
	// if nd.Contiguous() {
	// 	s := nd.Start
	// 	e := nd.Index(decrement(nd.Dims))
	// 	return nd.Impl[s : e+1]
	// }

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

func (nd *ndFloat64C) ReshapeFast(newShape []int) (NDFloat64, error) {
	if !nd.Contiguous() {
		return nil, errors.New("Array not contiguous")
	}

	return nd.Reshape(newShape)
}

func (nd *ndFloat64C) Reshape(newShape []int) (NDFloat64, error) {
	result := ndFloat64C{}
	size := product(newShape)
	currentSize := product(nd.Shape())

	if size != currentSize {
		return nil, errors.New("Size mismatch")
	}

	reshapeToSeries := (len(newShape) == 1) && (maximum(nd.Shape()) == len(newShape))

	if nd.Contiguous() || !reshapeToSeries {
		result.Start = 0
		result.Impl = nd.Impl
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

func (nd *ndFloat64C) MustReshape(newShape []int) NDFloat64 {
	result, e := nd.Reshape(newShape)
	if e != nil {
		panic(e.Error())
	}
	return result
}

func (nd *ndFloat64C) Get1(loc int) float64 {
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

func (nd *ndFloat64C) Set1(loc int, val float64) {
	nd.Set([]int{loc}, val)
}

func (nd *ndFloat64C) Apply1(loc int, step int, vals []float64) {
	for i := 0; i < len(vals); i++ {
		nd.Set1(loc+i*step, vals[i])
	}
}

func (nd *ndFloat64C) Get2(loc1 int, loc2 int) float64 {
	return nd.Get([]int{loc1, loc2})
}

func (nd *ndFloat64C) Set2(loc1 int, loc2 int, val float64) {
	nd.Set([]int{loc1, loc2}, val)
}

func (nd *ndFloat64C) Get3(loc1 int, loc2 int, loc3 int) float64 {
	return nd.Get([]int{loc1, loc2, loc3})
}

func (nd *ndFloat64C) Set3(loc1 int, loc2 int, loc3 int, val float64) {
	nd.Set([]int{loc1, loc2, loc3}, val)
}

func NewCArray(impl unsafe.Pointer, dims []int) NDFloat64 {
	return newCArray((*[1 << 30]C.double)(impl), dims)
}

func newCArray(impl *[1 << 30]C.double, dims []int) *ndFloat64C {
	result := ndFloat64C{}
	//	size := product(dims)
	result.Start = 0
	result.Impl = impl
	result.OriginalDims = dims
	result.Dims = dims
	result.Step = ones(len(dims))
	result.Offset = offsets(dims)
	return &result
}

func makeCArrayForTest(shape []int) *ndFloat64C {
	goArray := ARange(product(shape)).MustReshape(shape)
	impl := goArray.Unroll()

	v := reflect.Indirect(reflect.ValueOf(&impl))
	slice := (*reflect.SliceHeader)(unsafe.Pointer(v.UnsafeAddr()))
	addr := (*[1 << 30]C.double)(unsafe.Pointer(slice.Data))

	return newCArray(addr, shape)
}

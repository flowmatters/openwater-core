package data

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "ArrayType=float64,float32,int32,uint32,int64,uint64,int,uint"

import (
	"errors"
	"reflect"
	"unsafe"

	"github.com/flowmatters/openwater-core/util/slice"
	"github.com/joelrahman/genny/generic"
)

type CArrayType generic.CNumber

type ndArrayTypeC struct {
	ndArrayTypeCommon
	//	Impl *C.double
	Impl *[1 << 30]CArrayType
	//p2 := (*[1<<30]C.int)(unsafe.Pointer(p))
}

// func ArrayFromC(impl *C.double, shape []int) NDFloat64 {
// 	res := ndArrayTypeC{}
// 	res.Dims = shape
// 	res.OriginalDims = shape
// 	res.Start = 0
// 	res.Offset = uniform(len(shape), 0)
// 	res.Step = ones(len(shape))
// 	res.Impl = impl
// 	return &res
// }

func (nd *ndArrayTypeC) Get(loc []int) ArrayType {
	return ArrayType(nd.Impl[nd.Index(loc)])
}

func (nd *ndArrayTypeC) Set(loc []int, val ArrayType) {
	nd.Impl[nd.Index(loc)] = CArrayType(val)
}

func (nd *ndArrayTypeC) Slice(loc []int, dims []int, step []int) NDArrayType {
	result := ndArrayTypeC{}
	nd.slice(&result.ndArrayTypeCommon, loc, dims, step)
	result.Impl = nd.Impl
	return &result
}

func (nd *ndArrayTypeC) Apply(loc []int, dim int, step int, vals []ArrayType) {
	sliceDim := nd.NewIndex(1)
	sliceDim[dim] = len(vals)
	sliceStep := nd.NewIndex(1)
	sliceStep[dim] = step
	//	slice := nd.Slice(loc, sliceDim, sliceStep)

	// if slice.Contiguous() {
	// 	concrete := slice.(*ndArrayTypeC)
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

func (nd *ndArrayTypeC) ApplySlice(loc []int, step []int, vals NDArrayType) {
	shape := vals.Shape()
	slice := nd.Slice(loc, shape, step)

	idx := slice.NewIndex(0)
	size := product(shape)
	for pos := 0; pos < size; pos++ {
		slice.Set(idx, vals.Get(idx))
		increment(idx, shape)
	}
	// How to speed up
}

func (nd *ndArrayTypeC) CopyFrom(other NDArrayType) {
	nd.ApplySlice(nd.NewIndex(0), nil, other)
}

func (nd *ndArrayTypeC) Unroll() []ArrayType {
	// if nd.Contiguous() {
	// 	s := nd.Start
	// 	e := nd.Index(decrement(nd.Dims))
	// 	return nd.Impl[s : e+1]
	// }

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

func (nd *ndArrayTypeC) ReshapeFast(newShape []int) (NDArrayType, error) {
	if !nd.Contiguous() {
		return nil, errors.New("Array not contiguous")
	}

	return nd.Reshape(newShape)
}

func (nd *ndArrayTypeC) Reshape(newShape []int) (NDArrayType, error) {
	result := ndArrayTypeC{}
	size := product(newShape)
	currentSize := product(nd.Shape())

	if size != currentSize {
		return nil, errors.New("Size mismatch")
	}

	reshapeToSeries := (len(newShape) == 1) && (maximum(nd.Shape()) == len(newShape))

	if nd.Contiguous() || !reshapeToSeries {
		result.Start = nd.Start
		result.Impl = nd.Impl
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

func (nd *ndArrayTypeC) MustReshape(newShape []int) NDArrayType {
	result, e := nd.Reshape(newShape)
	if e != nil {
		panic(e.Error())
	}
	return result
}

func (nd *ndArrayTypeC) Get1(loc int) ArrayType {
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

func (nd *ndArrayTypeC) Set1(loc int, val ArrayType) {
	nd.Set([]int{loc}, val)
}

func (nd *ndArrayTypeC) Apply1(loc int, step int, vals []ArrayType) {
	for i := 0; i < len(vals); i++ {
		nd.Set1(loc+i*step, vals[i])
	}
}

func (nd *ndArrayTypeC) Get2(loc1 int, loc2 int) ArrayType {
	return nd.Get([]int{loc1, loc2})
}

func (nd *ndArrayTypeC) Set2(loc1 int, loc2 int, val ArrayType) {
	nd.Set([]int{loc1, loc2}, val)
}

func (nd *ndArrayTypeC) Get3(loc1 int, loc2 int, loc3 int) ArrayType {
	return nd.Get([]int{loc1, loc2, loc3})
}

func (nd *ndArrayTypeC) Set3(loc1 int, loc2 int, loc3 int, val ArrayType) {
	nd.Set([]int{loc1, loc2, loc3}, val)
}

func NewArrayTypeCArray(impl unsafe.Pointer, dims []int) NDArrayType {
	return newArrayTypeCArray((*[1 << 30]CArrayType)(impl), dims)
}

func newArrayTypeCArray(impl *[1 << 30]CArrayType, dims []int) *ndArrayTypeC {
	result := ndArrayTypeC{}
	//	size := product(dims)
	result.Start = 0
	result.Impl = impl
	result.OriginalDims = dims
	result.Dims = dims
	result.Step = slice.Ones(len(dims))
	result.Offset = offsets(dims)
	result.OffsetStep = multiply(result.Step, result.Offset)
	return &result
}

func makeArrayTypeCArrayForTest(shape []int) *ndArrayTypeC {
	goArray := ARangeArrayType(product(shape)).MustReshape(shape)
	impl := goArray.Unroll()

	v := reflect.Indirect(reflect.ValueOf(&impl))
	slice := (*reflect.SliceHeader)(unsafe.Pointer(v.UnsafeAddr()))
	addr := (*[1 << 30]CArrayType)(unsafe.Pointer(slice.Data))

	return newArrayTypeCArray(addr, shape)
}

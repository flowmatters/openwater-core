package data

import (
	"github.com/flowmatters/openwater-core/util/slice"
	"github.com/joelrahman/genny/generic"
)

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "ArrayType=float64,float32,int32,uint32,int64,uint64,int,uint"

type ArrayType generic.Number

// type NDArray interface {
// }

type NDArrayType interface {
	Len(axis int) int
	Shape() []int
	NDims() int
	NewIndex(val int) []int

	Get(loc []int) ArrayType
	Set(loc []int, val ArrayType)
	Slice(loc []int, dims []int, step []int) NDArrayType
	Apply(loc []int, dim int, step int, vals []ArrayType)
	ApplySlice(loc []int, step []int, vals NDArrayType)
	CopyFrom(other NDArrayType)
	Contiguous() bool
	Unroll() []ArrayType
	Reshape(newShape []int) (NDArrayType, error)
	MustReshape(newShape []int) NDArrayType
	ReshapeFast(newShape []int) (NDArrayType, error)
	Maximum() ArrayType
}

type ND1ArrayType interface {
	NDArrayType
	Len1() int
	Get1(loc int) ArrayType
	Set1(loc int, val ArrayType)
	Apply1(loc int, step int, vals []ArrayType)
}

type ND2ArrayType interface {
	NDArrayType
	Len2() int
	Get2(loc1 int, loc2 int) ArrayType
	Set2(loc1 int, loc2 int, val ArrayType)
}

type ND3ArrayType interface {
	NDArrayType
	Len3() int
	Get3(loc1 int, loc2 int, loc3 int) ArrayType
	Set3(loc1 int, loc2 int, loc3 int, val ArrayType)
}

type NdArrayTypeCommon struct {
	OriginalDims []int
	Dims         []int
	Start        int
	Offset       []int
	Step         []int
	OffsetStep   []int
}

func (nd *NdArrayTypeCommon) Len(ax int) int {
	return nd.Dims[ax]
}

func (nd *NdArrayTypeCommon) Shape() []int {
	return nd.Dims
}

func (nd *NdArrayTypeCommon) NDims() int {
	return len(nd.Dims)
}

func (nd *NdArrayTypeCommon) NewIndex(val int) []int {
	return slice.Uniform(nd.NDims(), val)
}

func (nd *NdArrayTypeCommon) Index(loc []int) int {
	result := nd.Start
	for i := 0; i < len(loc); i++ {
		result += loc[i] * nd.OffsetStep[i]
	}
	return result

	//	return nd.Start + dotProduct(multiply(loc, nd.Step), nd.Offset)
}

func (nd *NdArrayTypeCommon) Contiguous() bool {
	// What about step!
	var i int
	contiguousOffset := 1
	dimsMustBeOne := false

	for i = len(nd.Dims) - 1; i >= 0; i-- {
		if nd.Dims[i] > 1 {
			if dimsMustBeOne {
				return false
			}

			if nd.Step[i] > 1 {
				return false
			}

			if nd.Offset[i] > contiguousOffset {
				return false
			}
		}

		if nd.Dims[i] != nd.OriginalDims[i] {
			dimsMustBeOne = true
		}

		contiguousOffset *= nd.Dims[i]
	}

	return true
}

func (nd *NdArrayTypeCommon) Len1() int {
	return nd.Dims[0]
}

func (nd *NdArrayTypeCommon) Len2() int {
	return nd.Dims[1]
}

func (nd *NdArrayTypeCommon) Len3() int {
	return nd.Dims[2]
}

func (nd *NdArrayTypeCommon) SliceInto(dest *NdArrayTypeCommon, loc []int, dims []int, step []int) {
	dest.OriginalDims = nd.OriginalDims
	dest.Dims = dims
	dest.Start = nd.Start + dotProduct(loc, nd.Offset)
	dest.Offset = Multiply(nd.Offset, nd.Step)

	if step == nil {
		dest.Step = nd.Step
	} else {
		dest.Step = Multiply(nd.Step, step)
	}
	dest.OffsetStep = Multiply(dest.Step, dest.Offset)
}

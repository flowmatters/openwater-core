package data

import (
	"github.com/flowmatters/openwater-core/util/slice"
)

// type T generic.Number

// type NDArray interface {
// }

type ND[T Number] interface {
	Len(axis int) int
	Shape() []int
	NDims() int
	NewIndex(val int) []int

	Get(loc []int) T
	// Values(from []int, axis int, by int) iter.Seq[T]
	// RawIndices(from []int, axis int, by int) iter.Seq[int]
	// GetRaw(loc int) T
	// SetRaw(loc int, val T)
	// Enumerate(from []int, axis int) iter.Seq2[int, T]
	// Positions(from []int, to []int) iter.Seq[*Pos[T]]

	Set(loc []int, val T)
	Slice(loc []int, dims []int, step []int) ND[T]
	Apply(loc []int, dim int, step int, vals []T)
	ApplySlice(loc []int, step []int, vals ND[T])
	CopyFrom(other ND[T])
	Contiguous() bool
	Unroll() []T
	Reroll()
	Reshape(newShape []int) (ND[T], error)
	MustReshape(newShape []int) ND[T]
	ReshapeFast(newShape []int) (ND[T], error)
	Maximum() T
	Minimum() T
}

type ND1[T Number] interface {
	ND[T]
	Len1() int
	Get1(loc int) T
	Set1(loc int, val T)
	Apply1(loc int, step int, vals []T)
}

type ND2[T Number] interface {
	ND[T]
	Len2() int
	Get2(loc1 int, loc2 int) T
	Set2(loc1 int, loc2 int, val T)
}

type ND3[T Number] interface {
	ND[T]
	Len3() int
	Get3(loc1 int, loc2 int, loc3 int) T
	Set3(loc1 int, loc2 int, loc3 int, val T)
}

type NdCommon[T Number] struct {
	OriginalDims []int
	Dims         []int
	Start        int
	Offset       []int
	Step         []int
	OffsetStep   []int
}

func (nd *NdCommon[T]) Len(ax int) int {
	return nd.Dims[ax]
}

func (nd *NdCommon[T]) Shape() []int {
	return nd.Dims
}

func (nd *NdCommon[T]) NDims() int {
	return len(nd.Dims)
}

func (nd *NdCommon[T]) NewIndex(val int) []int {
	return slice.Uniform(nd.NDims(), val)
}

func (nd *NdCommon[T]) Index(loc []int) int {
	result := nd.Start
	for i := 0; i < len(loc); i++ {
		result += loc[i] * nd.OffsetStep[i]
	}
	return result

	//	return nd.Start + dotProduct(multiply(loc, nd.Step), nd.Offset)
}

func (nd *NdCommon[T]) Contiguous() bool {
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

func (nd *NdCommon[T]) Len1() int {
	return nd.Dims[0]
}

func (nd *NdCommon[T]) Len2() int {
	return nd.Dims[1]
}

func (nd *NdCommon[T]) Len3() int {
	return nd.Dims[2]
}

func (nd *NdCommon[T]) SliceInto(dest *NdCommon[T], loc []int, dims []int, step []int) {
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

// For interface purposes. Does nothing.
func (nd *NdCommon[T]) Reroll() {

}

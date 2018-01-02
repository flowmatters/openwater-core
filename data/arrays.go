package data

// type NDArray interface {
// }

type NDFloat64 interface {
	Len(axis int) int
	Shape() []int
	NDims() int
	NewIndex(val int) []int

	Get(loc []int) float64
	Set(loc []int, val float64)
	Slice(loc []int, dims []int, step []int) NDFloat64
	Apply(loc []int, dim int, step int, vals []float64)
	ApplySlice(loc []int, step []int, vals NDFloat64)
	Contiguous() bool
	Unroll() []float64
	Reshape(newShape []int) (NDFloat64, error)
	MustReshape(newShape []int) NDFloat64
	ReshapeFast(newShape []int) (NDFloat64, error)
	// getUnderlying(i int) float64
	// setUnderlying(i int, v float64)
	// takeImplementation(other NDFloat64) error
}

type ND1Float64 interface {
	NDFloat64
	Len1() int
	Get1(loc int) float64
	Set1(loc int, val float64)
	Apply1(loc int, step int, vals []float64)
}

type ND2Float64 interface {
	NDFloat64
	Len2() int
	Get2(loc1 int, loc2 int) float64
	Set2(loc1 int, loc2 int, val float64)
}

type ND3Float64 interface {
	NDFloat64
	Len3() int
	Get3(loc1 int, loc2 int, loc3 int) float64
	Set3(loc1 int, loc2 int, loc3 int, val float64)
}

type ndFloat64Common struct {
	OriginalDims []int
	Dims         []int
	Start        int
	Offset       []int
	Step         []int
}

func (nd *ndFloat64Common) Len(ax int) int {
	return nd.Dims[ax]
}

func (nd *ndFloat64Common) Shape() []int {
	return nd.Dims
}

func (nd *ndFloat64Common) NDims() int {
	return len(nd.Dims)
}

func (nd *ndFloat64Common) NewIndex(val int) []int {
	return uniform(nd.NDims(), val)
}

func (nd *ndFloat64Common) Index(loc []int) int {
	return nd.Start + dotProduct(multiply(loc, nd.Step), nd.Offset)
}

func (nd *ndFloat64Common) Contiguous() bool {
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

func (nd *ndFloat64Common) Len1() int {
	return nd.Dims[0]
}

func (nd *ndFloat64Common) Len2() int {
	return nd.Dims[1]
}

func (nd *ndFloat64Common) Len3() int {
	return nd.Dims[2]
}

func offsets(dims []int) []int {
	res := make([]int, len(dims))
	res[len(dims)-1] = 1
	for i := len(dims) - 2; i >= 0; i-- {
		res[i] = res[i+1] * dims[i+1]
	}
	return res
}

func (nd *ndFloat64Common) slice(dest *ndFloat64Common, loc []int, dims []int, step []int) {
	dest.OriginalDims = nd.OriginalDims
	dest.Dims = dims
	dest.Start = nd.Start + dotProduct(loc, nd.Offset)
	dest.Offset = multiply(nd.Offset, nd.Step)

	if step == nil {
		dest.Step = nd.Step
	} else {
		dest.Step = multiply(nd.Step, step)
	}
}

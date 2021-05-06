package fn

import (
	"errors"
	"fmt"
	"github.com/flowmatters/openwater-core/data"
)

func brackets(x float64, xs data.ND1Float64) (i, j int) {
	i = -1
	j = -1
	idx := []int{0}
	n := xs.Len1()

	idx[0] = 0
	if x < xs.Get(idx) {
		return
	}

	idx[0] = n-1
	if x > xs.Get(idx) {
		return
	}

	i = 0
	for j = 1; j < n; j++ {
		idx[0] = j
		valueAtJ := xs.Get(idx)
		if valueAtJ >= x {
			return
		}
		i += 1
	}
	i = -1
	j = -1
	return
}

func Piecewise(x float64, xs, ys data.ND1Float64) (y float64, err error) {
	err = nil
	i, j := brackets(x, xs)
	if (i < 0) || (j<0)  {
		err = errors.New(fmt.Sprintf("Couldn't find brackets for %f in %s",x,xs))
		return
	}
	idx := []int{i}
	x0 := xs.Get(idx)
	idx[0] = j
	x1 := xs.Get(idx)

	frac := (x-x0)/(x1-x0) // What if x1==x0?

	idx[0] = i
	y0 := ys.Get(idx)
	idx[0] = j
	y1 := ys.Get(idx)
	y = y0 + frac * (y1-y0)
	return
}


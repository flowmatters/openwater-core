package fn

import (
	"github.com/flowmatters/openwater-core/data"
)

func brackets(x float64, xs data.ND1Float64) (i, j int) {
	i = -1
	j = -1
	n := xs.Len1()

	if x < xs.Get1(0) {
		return
	}

	if x > xs.Get1(n-1) {
		return
	}

	i = 0
	for j = 1; j < n; i++ {
		if xs.Get1(j) > x {
			break
		}
		i += 1
	}
	return
}

func Piecewise(x float64, xs, ys data.ND1Float64) (y float64) {
	i, j := brackets(x, xs)
	x0 := xs.Get1(i)
	x1 := xs.Get1(j)

	frac := (x-x0)/(x1-x0) // What if x1==x0?

	y0 := ys.Get1(i)
	y1 := ys.Get1(j)
	y = y0 + frac * (y1-y0)
	return
}



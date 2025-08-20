package fn

import (
	"fmt"
)

func brackets(x float64, xs []float64) (i, j int) {
	i = -1
	j = -1
	n := len(xs)

	if x < xs[0] {
		return
	}

	if x > xs[n-1] {
		return
	}

	i = 0
	for j = 1; j < n; j++ {
		valueAtJ := xs[j]
		if valueAtJ >= x {
			return
		}
		i++
	}
	i = -1
	j = -1
	return
}

func Piecewise(x float64, xs, ys []float64) (y float64, err error) {
	err = nil
	i, j := brackets(x, xs)
	if (i < 0) || (j < 0) {
		err = fmt.Errorf("Couldn't find brackets for %f in %v", x, xs)
		return
	}
	idx := i
	x0 := xs[idx]
	idx = j
	x1 := xs[idx]

	frac := (x - x0) / (x1 - x0) // What if x1==x0?

	idx = i
	y0 := ys[idx]
	idx = j
	y1 := ys[idx]
	y = y0 + frac*(y1-y0)
	return
}

package fn

import (
	"errors"
	"fmt"
	"sort"
)

// Returns the indicies of the bracket around x in xs
func brackets(x float64, xs []float64) (i, j int) {
	n := len(xs)
	i = -1
	j = -1
	if x < xs[0] || x > xs[n-1] {
		return
	}
	j = sort.Search(n-1, func(j int) bool { return xs[j+1] >= x })
	i = j
	j++
	return
}

func Piecewise(x float64, xs, ys []float64) (y float64, err error) {
	err = nil
	i, j := brackets(x, xs)
	if (i < 0) || (j < 0) {
		err = errors.New(fmt.Sprintf("Couldn't find brackets for %f in %v", x, xs))
		return
	}
	x0 := xs[i]
	x1 := xs[j]

	frac := (x - x0) / (x1 - x0) // What if x1==x0?

	y0 := ys[i]
	y1 := ys[j]
	y = y0 + frac*(y1-y0)
	return
}

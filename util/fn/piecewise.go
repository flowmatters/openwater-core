package fn

import (
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
		err = fmt.Errorf("Couldn't find brackets for %f in %v", x, xs)
		return
	}
	x0 := xs[i]
	if j >= len(xs) {
		fmt.Printf("x=%f, xs=%v, i=%d, j=%d\n", x, xs, i, j)
		err = fmt.Errorf("Index j=%d out of range for xs of length %d", j, len(xs))
		return
	}
	x1 := xs[j]

	if x1 == x0 {
		err = fmt.Errorf("Division by zero: x0 (%f) == x1 (%f)", x0, x1)
		return
	}
	frac := (x - x0) / (x1 - x0) // What if x1==x0?

	y0 := ys[i]
	y1 := ys[j]
	y = y0 + frac*(y1-y0)
	return
}

package functions

import (
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
Sum:
	inputs:
		i1:
		i2:
	states:
	parameters:
	outputs:
		out:
	implementation:
		function: sum
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		function
*/

func sum(i1, i2 data.ND1Float64,
	out data.ND1Float64) {

	n := i1.Len1()
	idx := []int{0}

	for day := 0; day < n; day++ {
		idx[0] = day
		s := i1.Get(idx) + i2.Get(idx)
		out.Set(idx, s)
	}
}

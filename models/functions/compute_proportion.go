package functions

import (
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
ComputeProportion:
	inputs:
		numerator:
		denominator:
	states:
	parameters:
		resultOnZeroDenominator: default=86400
	outputs:
		proportion:
	implementation:
		function: computeProportion
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		dates function
*/

func computeProportion(numerator, denominator data.ND1Float64,
	resultOnZeroDenominator float64,
	proportion data.ND1Float64) {
	n := numerator.Len1()
	idx := []int{0}

	for i := 0; i < n; i++ {
		idx[0] = i
		n := numerator.Get(idx)
		d := denominator.Get(idx)

		if d == 0.0 {
			proportion.Set(idx,resultOnZeroDenominator)
		} else {
			proportion.Set(idx,n/d)
		}
	}
}

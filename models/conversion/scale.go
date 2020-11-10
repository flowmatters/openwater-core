package conversion

import (
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
ApplyScalingFactor:
  inputs:
		input:
	states:
	parameters:
		scale: 'default=1'
	outputs:
		output:
	implementation:
		function: applyScaling
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		partition
*/

func applyScaling(input data.ND1Float64,
	scale float64,
	output data.ND1Float64) {

	if scale == 0.0 {
		return
	}
	
	
	nDays := input.Len1()
	idx := []int{0}

	for i := 0; i < nDays; i++ {
		idx[0] = i
		incoming := input.Get(idx)
		output.Set(idx, incoming*scale)
	}
}

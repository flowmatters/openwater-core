package conversion

import (
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
VariablePartition:
  inputs:
		input:
		fraction:
	states:
	parameters:
	outputs:
		output1:
		output2:
	implementation:
		function: variablePartition
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		partition
*/

func variablePartition(input, fraction data.ND1Float64,
	output1, output2 data.ND1Float64) {

	nDays := input.Len1()
	idx := []int{0}

	for i := 0; i < nDays; i++ {
		idx[0] = i
		incoming := input.Get(idx)
		frac := fraction.Get(idx)
		output1.Set(idx, incoming*frac)
		output2.Set(idx, incoming*(1-frac))
	}
}

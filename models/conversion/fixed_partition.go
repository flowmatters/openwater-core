package conversion

import (
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
FixedPartition:
  inputs:
		input:
	states:
	parameters:
		fraction:
	outputs:
		output1:
		output2:
	implementation:
		function: fixedPartition
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		partition
*/

func fixedPartition(input data.ND1Float64,
	fraction float64,
	output1, output2 data.ND1Float64) {

	nDays := input.Len1()
	idx := []int{0}

	for i := 0; i < nDays; i++ {
		idx[0] = i
		incoming := input.Get(idx)
		output1.Set(idx, incoming*fraction)
		output2.Set(idx, incoming*(1-fraction))
	}
}

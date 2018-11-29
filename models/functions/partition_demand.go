package functions

import (
	"math"

	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
PartitionDemand:
	inputs:
		input:
		demand:
	states:
	parameters:
	outputs:
		outflow:
		extraction:
	implementation:
		function: partitionDemand
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		dates function
*/

func partitionDemand(input, demand data.ND1Float64,
	outflow, extraction data.ND1Float64) {
	n := input.Len1()
	idx := []int{0}

	for i := 0; i < n; i++ {
		idx[0] = n
		dmd := demand.Get(idx)
		inp := input.Get(idx)

		ext := math.Min(dmd, inp)
		out := math.Max(inp-ext, 0.0)

		outflow.Set(idx, out)
		extraction.Set(idx, ext)
	}
}

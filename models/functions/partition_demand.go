package functions

import (
	"math"
)

/*OW-SPEC
PartitionDemand:
	symbol: PD
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

func partitionDemand(input, demand []float64,
	outflow, extraction []float64) {
	n := len(input)

	for i := 0; i < n; i++ {
		dmd := demand[i]
		inp := input[i]

		ext := math.Min(dmd, inp)
		out := math.Max(inp-ext, 0.0)

		outflow[i] = out
		extraction[i] = ext
	}
}

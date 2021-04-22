package functions

import (
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
Input:
	inputs:
		input:
	states:
	parameters:
	outputs:
		output:
	implementation:
		function: inputNode
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		dates function
*/

func inputNode(input data.ND1Float64,
	output data.ND1Float64) {
	output.CopyFrom(input)
}

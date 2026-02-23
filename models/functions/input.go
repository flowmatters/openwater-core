package functions

/*OW-SPEC
Input:
	symbol: In
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

func inputNode(input []float64,
	output []float64) {
	copy(output, input)
}

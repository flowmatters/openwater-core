package conversion

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

func variablePartition(input, fraction []float64,
	output1, output2 []float64) {

	nDays := len(input)

	for i := 0; i < nDays; i++ {
		incoming := input[i]
		frac := fraction[i]
		output1[i] = incoming * frac
		output2[i] = incoming * (1 - frac)
	}
}

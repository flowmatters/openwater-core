package conversion

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

func fixedPartition(input []float64,
	fraction float64,
	output1, output2 []float64) {

	nDays := len(input)

	for i := 0; i < nDays; i++ {
		incoming := input[i]
		output1[i] = incoming * fraction
		output2[i] = incoming * (1 - fraction)
	}
}

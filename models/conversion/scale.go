package conversion

/*OW-SPEC
ApplyScalingFactor:
  symbol: SF
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

func applyScaling(input []float64,
	scale float64,
	output []float64) {

	if scale == 0.0 {
		return
	}

	nDays := len(input)

	for i := 0; i < nDays; i++ {
		incoming := input[i]
		output[i] = incoming * scale
	}
}

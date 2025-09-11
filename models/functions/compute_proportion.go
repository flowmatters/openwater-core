package functions

/*OW-SPEC
ComputeProportion:
	inputs:
		numerator:
		denominator:
	states:
	parameters:
		resultOnZeroDenominator: default=86400
	outputs:
		proportion:
	implementation:
		function: computeProportion
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		dates function
*/

func computeProportion(numerator, denominator []float64,
	resultOnZeroDenominator float64,
	proportion []float64) {
	n := len(numerator)

	for i := 0; i < n; i++ {
		n := numerator[i]
		d := denominator[i]

		if d == 0.0 {
			proportion[i] = resultOnZeroDenominator
		} else {
			proportion[i] = n / d
		}
	}
}

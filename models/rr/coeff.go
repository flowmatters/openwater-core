package rr

/* OW-SPEC
RunoffCoefficient:
  symbol: RC
  inputs:
    rainfall: mm
  states:
  parameters:
    coeff: ''
	outputs:
		runoff: mm
	implementation:
		function: runoffCoefficient
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		rainfall runoff test

*/

func runoffCoefficient(rainfall []float64, coeff float64, runoff []float64) {
	n := len(rainfall)
	for i := 0; i < n; i++ {
		runoff[i] = coeff * rainfall[i]
	}
}

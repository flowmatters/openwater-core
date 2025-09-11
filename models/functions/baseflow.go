package functions

/*OW-SPEC
BaseflowFilter:
	inputs:
		streamflow:
	states:
	parameters:
	outputs:
		quickflow:
		baseflow:
	implementation:
		function: baseflowFilter
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		baseflow separation
		streamlow
*/

func baseflowFilter(streamflow []float64,
	quickflow, baseflow []float64) {
}

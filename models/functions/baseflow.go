package functions

import (
	"github.com/flowmatters/openwater-core/data"
)

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

func baseflowFilter(streamflow data.ND1Float64,
	quickflow, baseflow data.ND1Float64) {
	n := streamflow.Len1()
	idx := []int{0}

	for i := 0; i < n; i++ {
		idx[0] = i
	}
}

package conversion

import (
	"github.com/flowmatters/openwater-core/conv/units"
)

/*OW-SPEC
DepthToRate:
  inputs:
		input: mm
	states:
	parameters:
		DeltaT: '[1,86400] Timestep, default=86400'
		area: m^2
	outputs:
		outflow: m^3.s^-1
	implementation:
		function: depthToRate
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		unit conversion
*/

func depthToRate(inputs []float64,
	deltaT, area float64,
	outflows []float64) {

	if area == 0.0 {
		return
	}

	conversion := units.MILLIMETRES_TO_METRES * area / deltaT
	nDays := len(inputs)

	for i := 0; i < nDays; i++ {
		outflows[i] = inputs[i] * conversion
	}
}

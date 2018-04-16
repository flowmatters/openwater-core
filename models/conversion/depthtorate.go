package conversion

import (
	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/data"
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

func depthToRate(inputs data.ND1Float64,
	deltaT, area float64,
	outflows data.ND1Float64) {
	conversion := conv.MILLIMETRES_TO_METRES * area / deltaT
	data.ScaleFloat64Array(outflows, inputs, conversion)
}

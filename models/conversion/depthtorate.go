package conversion

import (
	"github.com/flowmatters/openwater-core/conv/units"
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

	if area == 0.0 {
		return
	}

	conversion := units.MILLIMETRES_TO_METRES * area / deltaT
	nDays := inputs.Len1()
	idx := []int{0}

	for i := 0; i < nDays; i++ {
		idx[0] = i
		outflows.Set(idx, inputs.Get(idx)*conversion)
	}
}

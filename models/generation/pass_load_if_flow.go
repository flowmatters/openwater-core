package generation

import (
	"github.com/flowmatters/openwater-core/data"
)

const (
	EFFECTIVELY_ZERO = 1e-8
)
/*OW-SPEC
PassLoadIfFlow:
	inputs:
		flow: m^3.s^-1
		inputLoad:
	states:
	parameters:
		scalingFactor:
	outputs:
		outputLoad: kg
	implementation:
		function: passLoadIfFlow
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		constituent generation
*/

func passLoadIfFlow(flow, inputLoad data.ND1Float64,
	scalingFactor float64,
	outputLoad data.ND1Float64) {

	if scalingFactor == 0.0 {
		return
	}

	n := flow.Len1()
	idx := []int{0}

	for day := 0; day < n; day++ {
		idx[0] = day
		f := flow.Get(idx)
		l := inputLoad.Get(idx)

		if f > EFFECTIVELY_ZERO {
			outputLoad.Set(idx, l*scalingFactor)
		} else {
			outputLoad.Set(idx, 0.0)
		}
	}
}

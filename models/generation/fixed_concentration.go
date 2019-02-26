package generation

import (
	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
FixedConcentration:
	inputs:
		flow: m^3.s^-1
  states:
  parameters:
		concentration: '[0.1,10000]mg.L^-1 Event Mean Concentration'
	outputs:
		load: kg.s^-1
	implementation:
		function: fixedConcentration
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		constituent generation
*/

func fixedConcentration(flow data.ND1Float64, conc float64, load data.ND1Float64) {
	nDays := flow.Len1()
	idx := []int{0}

	for i := 0; i < nDays; i++ {
		idx[0] = i
		f := flow.Get(idx)

		l := f * conc * units.MG_PER_LITRE_TO_KG_PER_M3

		load.Set(idx, l)
	}
}

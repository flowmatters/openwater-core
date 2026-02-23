package generation

import (
	"github.com/flowmatters/openwater-core/conv/units"
)

/*OW-SPEC
FixedConcentration:
	symbol: FC
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

func fixedConcentration(flow []float64, conc float64, load []float64) {
	nDays := len(flow)

	if conc == 0.0 {
		return
	}

	for i := 0; i < nDays; i++ {
		f := flow[i]

		l := f * conc * units.MG_PER_LITRE_TO_KG_PER_M3

		load[i] = l
	}
}

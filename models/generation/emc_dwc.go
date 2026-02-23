package generation

import (
	"github.com/flowmatters/openwater-core/conv/units"
)

/*OW-SPEC
EmcDwc:
	symbol: ED
	inputs:
		quickflow: m^3.s^-1
		baseflow: m^3.s^-1
  states:
  parameters:
		EMC: '[0.1,10000]mg.L^-1 Event Mean Concentration'
		DWC: '[0.1,10000]mg.L^-1 Dry Weather Concentration'
	outputs:
		quickLoad: kg.s^-1
		slowLoad: kg.s^-1
		totalLoad: kg.s^-1
	implementation:
		function: emcDWC
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		constituent generation
*/

func emcDWC(quickflow, slowflow []float64, emc, dwc float64, quickLoad, slowLoad, totalLoad []float64) {
	nDays := len(quickflow)

	if (emc == 0.0) && (dwc == 0.0) {
		return
	}

	for i := 0; i < nDays; i++ {
		qf := quickflow[i]
		sf := slowflow[i]

		ql := qf * emc * units.MG_PER_LITRE_TO_KG_PER_M3
		sl := sf * dwc * units.MG_PER_LITRE_TO_KG_PER_M3
		total := ql + sl

		quickLoad[i] = ql
		slowLoad[i] = sl
		totalLoad[i] = total
	}
}

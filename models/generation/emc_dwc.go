package generation

import (
	"github.com/flowmatters/openwater-core/conv"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
EmcDwc:
	inputs:
		quickFlow: m^3.s^-1
		slowFlow: m^3.s^-
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

func emcDWC(quickflow, slowflow data.ND1Float64, emc, dwc float64, quickLoad, slowLoad, totalLoad data.ND1Float64) {
	nDays := quickflow.Len1()

	for i := 0; i < nDays; i++ {
		qf := quickflow.Get1(i)
		sf := slowflow.Get1(i)

		ql := qf * emc * conv.MG_PER_LITER_TO_KG_PER_M3
		sl := sf * dwc * conv.MG_PER_LITER_TO_KG_PER_M3
		total := ql + sl

		quickLoad.Set1(i, ql)
		slowLoad.Set1(i, sl)
		totalLoad.Set1(i, total)
	}
}

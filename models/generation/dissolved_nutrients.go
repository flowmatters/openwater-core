package generation

import (
	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
SednetDissolvedNutrientGeneration:
  inputs:
		quickflow: m^3.s^-1
		slowflow: m^3.s^-1
  states:
	parameters:
		dissConst_EMC: mg.L^-1
		dissConst_DWC: mg.L^-1
	outputs:
		quickflowConstituent: kg.s^-1
		slowflowConstituent: kg.s^-1
	implementation:
		function: dissolvedNutrients
		type: scalar
		lang: go
		outputs: params
	init:
	  zero: true
	tags:
		nutrients

*/
func dissolvedNutrients(quickflow, slowflow data.ND1Float64,
	dissConst_EMC, dissConst_DWC float64,
	quickflowConstituent, slowflowConstituent data.ND1Float64) {
	//All calcs done in units / day then converted back to units per sec for E2 consumption
	n := quickflow.Len1()

	idx := []int{0}
	for day := 0; day < n; day++ {
		idx[0] = day

		DailyConstituent_EMC_mgL := dissConst_EMC // +(dissConst_EMC_Max * Math.Exp(-1 * b1 * daysSinceStart));

		cumecs_to_lpd := float64(units.SECONDS_PER_DAY) * units.CUBIC_METRES_TO_LITRES
		Quickflow_Litres := quickflow.Get(idx) * cumecs_to_lpd
		Slowflow_Litres := slowflow.Get(idx) * cumecs_to_lpd

		dissolvedConstituent_Quickflow_Load_kg := DailyConstituent_EMC_mgL * Quickflow_Litres * units.MILLIGRAM_TO_KG
		dissolvedConstituent_Slowflow_Load_kg := dissConst_DWC * Slowflow_Litres * units.MILLIGRAM_TO_KG

		// Daily_DissolvedConstituent_Concentration_mg_L := 0.0
		// if Quickflow_Litres > 0 && dissolvedConstituent_Quickflow_Load_kg > 0 {
		// 	Daily_DissolvedConstituent_Concentration_mg_L = dissolvedConstituent_Quickflow_Load_kg * units.KG_TO_MILLIGRAM / Quickflow_Litres
		// } else {
		// 	Daily_DissolvedConstituent_Concentration_mg_L = 0
		// }

		quickflowConstituent.Set(idx, dissolvedConstituent_Quickflow_Load_kg/units.SECONDS_PER_DAY)
		slowflowConstituent.Set(idx, dissolvedConstituent_Slowflow_Load_kg/units.SECONDS_PER_DAY)

		// total_Dissolved_Constituent_kg += (dissolvedConstituent_Quickflow_Load_kg + dissolvedConstituent_Slowflow_Load_kg)
	}
}

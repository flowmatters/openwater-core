package generation

import (
	"math"

	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
SednetParticulateNutrientGeneration:
	inputs:
		fineSedModelFineSheetGeneratedKg:
		fineSedModelCoarseSheetGeneratedKg:
		fineSedModelFineGullyGeneratedKg:
		fineSedModelCoarseGullyGeneratedKg:
		slowflow: m^3.s^-1
  states:
	parameters:
		area: m^2
		nutSurfSoilConc: kg.kg^-1
		hillDeliveryRatio: '%'
		Nutrient_Enrichment_Ratio:
		nutSubSoilConc: kg.kg^-1
		Nutrient_Enrichment_Ratio_Gully:
		gullyDeliveryRatio: '%'
		nutrientDWC: mg.L^-1
		Do_P_CREAMS_Enrichment: flag
	outputs:
		quickflowConstituent: kg.s^-1
		slowflowConstituent: kg.s^-1
	implementation:
		function: particulateNutrients
		type: scalar
		lang: go
		outputs: params
	init:
	  zero: true
	tags:
		particulate nutrients
*/

func particulateNutrients(fineSedModelFineSheetGeneratedKg, fineSedModelCoarseSheetGeneratedKg,
	fineSedModelFineGullyGeneratedKg, fineSedModelCoarseGullyGeneratedKg,
	slowflow data.ND1Float64,
	area, nutSurfSoilConc, hillDeliveryRatio, Nutrient_Enrichment_Ratio, nutSubSoilConc,
	Nutrient_Enrichment_Ratio_Gully, gullyDeliveryRatio,
	nutrientDWC, Do_P_CREAMS_Enrichment float64,
	quickflowConstituent, slowflowConstituent data.ND1Float64) {
	const CREAMS_CONSTANT = 1.2
	//All calcs done in units / day then converted back to units per sec for E2 consumption
	n := fineSedModelCoarseSheetGeneratedKg.Len1()

	idx := []int{0}
	for day := 0; day < n; day++ {
		idx[0] = day

		Hillslope_ErosionLoad_kg := 0.0
		Gully_ErosionLoad_kg := 0.0

		//fuAreaHa = areaInSquareMeters * ConversionConst.squareMetres_to_Hectares;

		Daily_Gully_Particulate_load_kg := 0.0
		Daily_Hillslope_Particulate_load_kg := 0.0
		Daily_Total_Particulate_load_kg := 0.0

		soilLoad := 0.0

		//All calcs done in kg/day then converted back to kg per sec for E2 consumption
		Hillslope_ErosionLoad_kg = fineSedModelFineSheetGeneratedKg.Get(idx) + fineSedModelCoarseSheetGeneratedKg.Get(idx)
		Gully_ErosionLoad_kg = fineSedModelFineGullyGeneratedKg.Get(idx) + fineSedModelCoarseGullyGeneratedKg.Get(idx)

		soilLoad = Hillslope_ErosionLoad_kg + Gully_ErosionLoad_kg

		if soilLoad > 0 {
			//This is only ever set to true during APSIM parameterisation
			if Do_P_CREAMS_Enrichment > 0 {

				areaHa := area * units.SQUARE_METRES_TO_HECTARES

				logComponent := 0.0
				//Enrichment is, for some reason, limited to just the hillslope contribution (spec doc quite clear about this)
				if Hillslope_ErosionLoad_kg > 0 && areaHa > 0 {
					logComponent = 2.4 - 0.27*math.Log(Hillslope_ErosionLoad_kg/areaHa)
				}

				PEnrichment := CREAMS_CONSTANT
				if logComponent > 0 {
					//equates to a soil load of 7250 hg/ha/day or less
					PEnrichment = CREAMS_CONSTANT * logComponent
				}

				//double PEnrichment = creamsConstant * (2.4 - 0.27 * Math.Log(soilLoad / areaHa));
				//RDS changed a suspected typo in the next line 27-9-2011 - changed hillDeliveryRatio * 0.1 to hillDeliveryRatio * 0.01 - meant to convert percent to ratio
				Daily_Hillslope_Particulate_load_kg = Hillslope_ErosionLoad_kg * nutSurfSoilConc * PEnrichment * hillDeliveryRatio * 0.01

			} else {
				//normal SedNet approach, where the NER itself will determine if enrichmemnt occurs
				//RDS changed a suspected typo in the next line 27-9-2011 - changed hillDeliveryRatio * 0.1 to hillDeliveryRatio * 0.01 - meant to convert percent to ratio
				Daily_Hillslope_Particulate_load_kg = Hillslope_ErosionLoad_kg * nutSurfSoilConc * Nutrient_Enrichment_Ratio * (hillDeliveryRatio * 0.01)
				Daily_Gully_Particulate_load_kg = Gully_ErosionLoad_kg * nutSubSoilConc * Nutrient_Enrichment_Ratio_Gully * (gullyDeliveryRatio * 0.01)
			}

			//Daily_Gully_Particulate_load_kg = Gully_ErosionLoad_kg * nutSubSoilConc * (gullyDeliveryRatio * 0.01);
		}

		Daily_Total_Particulate_load_kg = Daily_Hillslope_Particulate_load_kg + Daily_Gully_Particulate_load_kg // * ConversionConst.Grams_to_Kilograms;

		quickflowConstituent.Set(idx, Daily_Total_Particulate_load_kg/units.SECONDS_PER_DAY)
		slowflowConstituent.Set(idx, slowflow.Get(idx)*nutrientDWC*units.MG_PER_LITRE_TO_KG_PER_M3)

		// Total_Total_Particulate_Constituent_kg += Daily_Total_Particulate_load_kg
		// Total_Hillslope_Particulate_Constituent_kg += Daily_Hillslope_Particulate_load_kg
		// Total_Gully_Particulate_Constituent_kg += Daily_Gully_Particulate_load_kg
	}
}

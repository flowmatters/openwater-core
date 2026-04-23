package generation

import (
	"math"

	u "github.com/flowmatters/openwater-core/conv/units"
)

/*OW-SPEC
SednetParticulateNutrientGeneration:
	symbol: Pn
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
		totalLoad: kg.s^-1
		hillslopeContribution: kg.s^-1
		gullyContribution: kg.s^-1
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
	slowflow []float64,
	area,
	nutSurfSoilConc, hillDeliveryRatio, Nutrient_Enrichment_Ratio,
	nutSubSoilConc, Nutrient_Enrichment_Ratio_Gully, gullyDeliveryRatio,
	nutrientDWC, Do_P_CREAMS_Enrichment float64,
	quickflowConstituent, slowflowConstituent, totalLoad,
	hillslopeContribution, gullyContribution []float64) {
	const CREAMS_CONSTANT = 1.2
	//All calcs done in units / day then converted back to units per sec for E2 consumption
	n := len(fineSedModelCoarseSheetGeneratedKg)
	areaHa := area * u.SQUARE_METRES_TO_HECTARES

	for day := 0; day < n; day++ {

		Gully_Particulate_load_kg := 0.0
		Hillslope_Particulate_load_kg := 0.0
		Total_Particulate_load_kg := 0.0

		Hillslope_ErosionLoad_kg := fineSedModelFineSheetGeneratedKg[day] + fineSedModelCoarseSheetGeneratedKg[day]
		Gully_ErosionLoad_kg := fineSedModelFineGullyGeneratedKg[day] + fineSedModelCoarseGullyGeneratedKg[day]

		if Do_P_CREAMS_Enrichment > 0.5 {
			logComponent := 0.0
			if Hillslope_ErosionLoad_kg > 0 && areaHa > 0 {
				// Erosion inputs are in kg/s but CREAMS formula expects kg/day
				logComponent = 2.4 - 0.27*math.Log(Hillslope_ErosionLoad_kg*u.SECONDS_PER_DAY/areaHa)
			}

			PEnrichment := CREAMS_CONSTANT
			if logComponent > 0 {
				PEnrichment = CREAMS_CONSTANT * logComponent
			}

			Hillslope_Particulate_load_kg = Hillslope_ErosionLoad_kg * nutSurfSoilConc * PEnrichment * (hillDeliveryRatio * u.PERCENT_TO_PROPORTION)
			// TODO: Confirm expected behaviour - gully contribution not calculated in CREAMS branch. Commenting out to match Source for now.
			// Gully_Particulate_load_kg = Gully_ErosionLoad_kg * nutSubSoilConc * Nutrient_Enrichment_Ratio_Gully * (gullyDeliveryRatio * u.PERCENT_TO_PROPORTION)

		} else {
			//normal SedNet approach, where the NER itself will determine if enrichmemnt occurs
			//RDS changed a suspected typo in the next line 27-9-2011 - changed hillDeliveryRatio * 0.1 to hillDeliveryRatio * 0.01 - meant to convert percent to ratio
			Hillslope_Particulate_load_kg = Hillslope_ErosionLoad_kg * nutSurfSoilConc * Nutrient_Enrichment_Ratio * (hillDeliveryRatio * u.PERCENT_TO_PROPORTION)
			Gully_Particulate_load_kg = Gully_ErosionLoad_kg * nutSubSoilConc * Nutrient_Enrichment_Ratio_Gully * (gullyDeliveryRatio * u.PERCENT_TO_PROPORTION)
		}

		//Daily_Gully_Particulate_load_kg = Gully_ErosionLoad_kg * nutSubSoilConc * (gullyDeliveryRatio * 0.01);

		Total_Particulate_load_kg = Hillslope_Particulate_load_kg + Gully_Particulate_load_kg // * ConversionConst.Grams_to_Kilograms;

		quickLoad := Total_Particulate_load_kg
		quickflowConstituent[day] = quickLoad
		slowLoad := slowflow[day] * nutrientDWC * u.MG_PER_LITRE_TO_KG_PER_M3
		slowflowConstituent[day] = slowLoad
		totalLoad[day] = quickLoad + slowLoad

		hillslopeContribution[day] = Hillslope_Particulate_load_kg
		gullyContribution[day] = Gully_Particulate_load_kg

		// Total_Total_Particulate_Constituent_kg += Daily_Total_Particulate_load_kg
		// Total_Hillslope_Particulate_Constituent_kg += Daily_Hillslope_Particulate_load_kg
		// Total_Gully_Particulate_Constituent_kg += Daily_Gully_Particulate_load_kg
	}
}

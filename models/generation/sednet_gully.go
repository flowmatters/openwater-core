package generation

import (
	"math"

	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
DynamicSednetGully:
	inputs:
		quickflow: m^3.s^-1
		year: year
		AnnualRunoff: 'mm.yr^-1'
		annualLoad: ''
	states:
	parameters:
		YearDisturbance: ''
		GullyEndYear: ''
		Area: m^2
		averageGullyActivityFactor: '[0,3]'
		GullyAnnualAverageSedimentSupply: 't.yr^-1'
		GullyPercentFine: 'Average clay + silt percentage of gully material'
		managementPracticeFactor: ''
		longtermRunoffFactor: ''
		dailyRunoffPowerFactor: ''
		sdrFine: ''
		sdrCoarse: ''
		timeStepInSeconds: '[0,100000000]s Duration of timestep in seconds, default=86400'
	outputs:
		fineLoad: kg
		coarseLoad: kg
		generatedFine: kg
		generatedCoarse: kg
	implementation:
		function: sednetGullyOrig
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		constituent generation
		sediment
		gully
*/

type gullyExportFn func(float64, float64, float64, float64, float64, float64, float64, float64, float64, float64) (float64, float64)

func sednetGully(quickflow, year, annualRunoff_ts, annualLoad_ts data.ND1Float64,
	yearDisturbance, gullyEndYear, area, averageGullyActivityFactor,
	annualAverageSedimentSupply, percentFine,
	managementPracticeFactor, longtermRunoffFactor, dailyRunoffPowerFactor,
	sdrFine, sdrCoarse, timestepInSeconds float64, 
	fineLoad, coarseLoad, generatedFine, generatedCoarse data.ND1Float64, calc gullyExportFn) {
	n := quickflow.Len1()
	idx := []int{0}
	propFine := percentFine / 100

	for day := 0; day < n; day++ {
		idx[0] = day
		yr := year.Get(idx)
		annualLoad := annualLoad_ts.Get(idx)
		annualRunoff := annualRunoff_ts.Get(idx)

		runoffRate := quickflow.Get(idx)
		if yr < yearDisturbance {
			fineLoad.Set(idx, 0)
			coarseLoad.Set(idx, 0)
			continue
		}
		activityFactor := 1.0

		if yr > gullyEndYear {
			activityFactor = averageGullyActivityFactor
		}

		if runoffRate == 0 || annualRunoff == 0 { //|| annualAverageSedimentSupply == 0 {
			fineLoad.Set(idx, 0)
			coarseLoad.Set(idx, 0)
			continue
		}

		generated_gully_load_kg_fine, generated_gully_load_kg_coarse := calc(runoffRate, annualRunoff, area, propFine, activityFactor, managementPracticeFactor,
			annualLoad, annualAverageSedimentSupply, longtermRunoffFactor, dailyRunoffPowerFactor)

		generated_gully_load_kg_fine = generated_gully_load_kg_fine/timestepInSeconds
		generated_gully_load_kg_coarse = generated_gully_load_kg_coarse/timestepInSeconds

		fineLoad.Set(idx, generated_gully_load_kg_fine*(sdrFine*0.01))
		coarseLoad.Set(idx, generated_gully_load_kg_coarse*(sdrCoarse*0.01))
		generatedFine.Set(idx, generated_gully_load_kg_fine)
		generatedCoarse.Set(idx, generated_gully_load_kg_coarse)
	}
}

func sednetGullyOrig(quickflow, year, annualRunoff, annualLoad data.ND1Float64,
	yearDisturbance, gullyEndYear, area, averageGullyActivityFactor,
	annualAverageSedimentSupply, percentFine,
	managementPracticeFactor, longtermRunoffFactor, dailyRunoffPowerFactor,
	sdrFine, sdrCoarse, timestepInSeconds float64,
	fineLoad, coarseLoad, generatedFine, generatedCoarse data.ND1Float64) {
	sednetGully(quickflow, year, annualRunoff, annualLoad,
		yearDisturbance, gullyEndYear, area, averageGullyActivityFactor,
		annualAverageSedimentSupply, percentFine,
		managementPracticeFactor, longtermRunoffFactor, dailyRunoffPowerFactor,
		sdrFine, sdrCoarse, timestepInSeconds, fineLoad, coarseLoad, generatedFine, generatedCoarse, gullyLoadOrig)
}

func gullyLoadOrig(dailyRunoff, annualRunoff, area, propFine, activityFactor, managementPracticeFactor, annualLoad, annualSupply,
	longTermRunoffFactor, dailyRunoffPowerfactor float64) (float64, float64) {
	//Scott's simplified factor to break annual load into daily
	annualToDailyAdjustmentFactor := 1 / 365.25

	thisYearsSedimentSupply := annualSupply // math.Max(annualSupply, annualLoad)

	dailyRunoffFactor := 1.0
	//Stop NaN's on models that don't have the required longterm flow analysis
	if longTermRunoffFactor > 0 {
		if dailyRunoffPowerfactor <= 0 {
			dailyRunoffPowerfactor = 1
		}

		//Swap these over if reverting to Scott's power-based event-to-annual adjustment

		//Scott's complex version with stuffed raised to to a power
		//all cumecs
		dailyRunoffFactor = math.Pow(dailyRunoff, dailyRunoffPowerfactor) / longTermRunoffFactor
	}

	Gully_Daily_Load_kg_Fine := annualToDailyAdjustmentFactor * dailyRunoffFactor * propFine * activityFactor * managementPracticeFactor * thisYearsSedimentSupply * units.TONNES_TO_KG

	Gully_Daily_Load_kg_Coarse := annualToDailyAdjustmentFactor * dailyRunoffFactor * (1 - propFine) * thisYearsSedimentSupply * managementPracticeFactor * units.TONNES_TO_KG

	return Gully_Daily_Load_kg_Fine, Gully_Daily_Load_kg_Coarse
}

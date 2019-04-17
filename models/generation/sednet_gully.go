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
	states:
	parameters:
		YearDisturbance: ''
		GullyEndYear: ''
		Area: m^2
		averageGullyActivityFactor: '[0,3]'
		AnnualRunoff: 'mm.yr^-1'
		GullyAnnualAverageSedimentSupply: 't.yr^-1'
		GullyPercentFine: 'Average clay + silt percentage of gully material'
		managementPracticeFactor: ''
		annualLoad: ''
		longtermRunoffFactor: ''
		dailyRunoffPowerFactor: ''
		sdrFine: ''
		sdrCoarse: ''
	outputs:
		fineLoad: kg
		coarseLoad: kg
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

func sednetGully(quickflow, year data.ND1Float64,
	yearDisturbance, gullyEndYear, area, averageGullyActivityFactor,
	annualRunoff, annualAverageSedimentSupply, percentFine,
	managementPracticeFactor, annualLoad, longtermRunoffFactor, dailyRunoffPowerFactor,
	sdrFine, sdrCoarse float64, fineLoad, coarseLoad data.ND1Float64, calc gullyExportFn) {
	n := quickflow.Len1()
	idx := []int{0}
	propFine := percentFine / 100

	for day := 0; day < n; day++ {
		idx[0] = day
		yr := year.Get(idx)
		dailyRunoff := quickflow.Get(idx)
		if yr < yearDisturbance {
			fineLoad.Set(idx, 0)
			coarseLoad.Set(idx, 0)
		}
		activityFactor := 1.0

		if yr > gullyEndYear {
			activityFactor = averageGullyActivityFactor
		}

		if dailyRunoff == 0 || annualRunoff == 0 || annualAverageSedimentSupply == 0 {
			fineLoad.Set(idx, 0)
			coarseLoad.Set(idx, 0)
		}

		Gully_Daily_Load_kg_Fine, Gully_Daily_Load_kg_Coarse := calc(dailyRunoff, annualRunoff, area, propFine, activityFactor, managementPracticeFactor,
			annualLoad, annualAverageSedimentSupply, longtermRunoffFactor, dailyRunoffPowerFactor)

		fineLoad.Set(idx, Gully_Daily_Load_kg_Fine*(sdrFine*0.01))
		coarseLoad.Set(idx, Gully_Daily_Load_kg_Coarse*(sdrCoarse*0.01))
	}
}

func sednetGullyOrig(quickflow, year data.ND1Float64,
	yearDisturbance, gullyEndYear, area, averageGullyActivityFactor,
	annualRunoff, annualAverageSedimentSupply, percentFine,
	managementPracticeFactor, annualLoad, longtermRunoffFactor, dailyRunoffPowerFactor,
	sdrFine, sdrCoarse float64, fineLoad, coarseLoad data.ND1Float64) {
	sednetGully(quickflow, year,
		yearDisturbance, gullyEndYear, area, averageGullyActivityFactor,
		annualRunoff, annualAverageSedimentSupply, percentFine,
		managementPracticeFactor, annualLoad, longtermRunoffFactor, dailyRunoffPowerFactor,
		sdrFine, sdrCoarse, fineLoad, coarseLoad, gullyLoadOrig)
}

func gullyLoadOrig(dailyRunoff, annualRunoff, area, propFine, activityFactor, managementPracticeFactor, annualLoad, annualSupply,
	longTermRunoffFactor, dailyRunoffPowerfactor float64) (float64, float64) {
	//Scott's simplified factor to break annual load into daily
	annualToDailyAdjustmentFactor := 1 / 365.25

	thisYearsSedimentSupply := annualSupply

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

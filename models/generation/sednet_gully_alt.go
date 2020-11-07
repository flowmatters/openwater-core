package generation

import (
	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
DynamicSednetGullyAlt:
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
		function: sednetGullyDerm
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

func sednetGullyDerm(quickflow, year, annualRunoff, annualLoad data.ND1Float64,
	yearDisturbance, gullyEndYear, area, averageGullyActivityFactor,
	annualAverageSedimentSupply, percentFine,
	managementPracticeFactor, longtermRunoffFactor, dailyRunoffPowerFactor,
	sdrFine, sdrCoarse, timeStepInSeconds float64,
	fineLoad, coarseLoad, generatedFine, generatedCoarse data.ND1Float64) {
	sednetGully(quickflow, year, annualRunoff, annualLoad,
		yearDisturbance, gullyEndYear, area, averageGullyActivityFactor,
		annualAverageSedimentSupply, percentFine,
		managementPracticeFactor, longtermRunoffFactor, dailyRunoffPowerFactor,
		sdrFine, sdrCoarse, timeStepInSeconds, fineLoad, coarseLoad, generatedFine, generatedCoarse, gullyLoadDerm)
}

func gullyLoadDerm(runoffRate, annualRunoff, area, propFine, activityFactor, managementPracticeFactor, annualLoad, annualSupply,
	longTermRunoffFactor, dailyRunoffPowerfactor float64) (float64, float64) {
	//Annual_Gully_Load has already had the 'yearly proportion' taken into account (could be a non-linear calculation
	//and has also had the 'annual runoff magnitude compared to average annual runoff' adjustment made during parameterisation

	//Gully_Daily_Load_kg_Fine = (Daily_Runoff / Annual_Runoff) * Gully_Annual_Fine_Load;
	//Gully_Daily_Load_kg_Coarse = (Daily_Runoff / Annual_Runoff) * Gully_Annual_Coarse_Load;

	//double fact = (Event_Runoff / Annual_Runoff);

	//DateTime checker = new DateTime(1987, 1, 15);
	////if (quickflow > 0)
	dailyRunoffDepth := (runoffRate / area) * units.METRES_TO_MILLIMETRES * units.SECONDS_PER_DAY
	annualSupplyAfterManagement := managementPracticeFactor * annualLoad
	Gully_Daily_Load_kg_Fine := (dailyRunoffDepth / annualRunoff) * propFine * activityFactor * annualSupplyAfterManagement

	Gully_Daily_Load_kg_Coarse := (dailyRunoffDepth / annualRunoff) * (1 - propFine) * annualSupplyAfterManagement

	return Gully_Daily_Load_kg_Fine, Gully_Daily_Load_kg_Coarse
}

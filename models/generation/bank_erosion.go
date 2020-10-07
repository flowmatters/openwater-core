package generation

import (
	"math"

	"github.com/flowmatters/openwater-core/conv/rough"
	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
BankErosion:
	inputs:
		downstreamFlowVolume:
		totalVolume:
  states:
	parameters:
		riparianVegPercent:
		maxRiparianVegEffectiveness:
		soilErodibility:
		bankErosionCoeff:
		linkSlope:
		bankFullFlow:
		bankMgtFactor:
		sedBulkDensity:
		bankHeight:
		linkLength:
		dailyFlowPowerFactor:
		longTermAvDailyFlow:
		soilPercentFine:
		durationInSeconds: '[1,86400] Timestep, default=86400'
	outputs:
		bankErosionFine: kg
		bankErosionCoarse: kg
	implementation:
		function: bankErosion
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		sediment generation
*/

// Does this return the same value every timestep? (Name suggests it does!)
func meanAnnualBankErosion(riparianVegPercent, maxRiparianVegEffectiveness, soilErodibility, bankErosionCoeff,
	linkSlope, bankFullFlow, bankMgtFactor, sedBulkDensity, bankHeight, linkLength float64) float64 {
	densityWater := 1000.0 // kg.m^-3
	gravity := 9.81        // m.s^-2

	BankErodability := (1 - math.Min((riparianVegPercent/100), (maxRiparianVegEffectiveness/100))) * (soilErodibility / 100)
	RetreatRate_MperYr := bankErosionCoeff * densityWater * gravity * linkSlope * bankFullFlow * bankMgtFactor
	massConversion := sedBulkDensity * bankHeight * linkLength
	result := massConversion * RetreatRate_MperYr * BankErodability

	return result
}

func bankErosion(downstreamFlowVolume, totalVolume data.ND1Float64,
	riparianVegPercent, maxRiparianVegEffectiveness, soilErodibility, bankErosionCoeff,
	linkSlope, bankFullFlow, bankMgtFactor, sedBulkDensity, bankHeight, linkLength,
	dailyFlowPowerFactor, longTermAvDailyFlow, soilPercentFine, durationInSeconds float64,
	bankErosionFine, bankErosionCoarse data.ND1Float64) {
	idx := []int{0}
	n := downstreamFlowVolume.Len1()
	meanAnnual := meanAnnualBankErosion(riparianVegPercent, maxRiparianVegEffectiveness, soilErodibility, bankErosionCoeff,
		linkSlope, bankFullFlow, bankMgtFactor, sedBulkDensity, bankHeight, linkLength)

	//implementation of formula 4.20 in specifiction document - page 22
	//bank erosion calculated as per this formula is in tonnes per day
	for i := 0; i < n; i++ {
		idx[0] = i
		LinkDischargeFactor := 0.0
		outflow := downstreamFlowVolume.Get(idx)

		if totalVolume.Get(idx) <= 0 || outflow <= 0 || longTermAvDailyFlow <= 0 {
			LinkDischargeFactor = 0
		} else {
			//convert to daily m3 before rasing to power
			LinkDischargeFactor = math.Pow(outflow*durationInSeconds, dailyFlowPowerFactor) / longTermAvDailyFlow
		}

		//	mainChannelArea := /*mainChannelStreamDimensions.*/contribArea_Km

		BankErosion_TperDay := (meanAnnual * LinkDischargeFactor) / rough.DAYS_PER_YEAR

		BankErosionTotal_kg_per_Second := BankErosion_TperDay * units.TONNES_TO_KG / durationInSeconds

		bankErosionFine_Kg_per_Second := BankErosionTotal_kg_per_Second * (soilPercentFine * units.PERCENT_TO_PROPORTION)
		bankErosionCoarse_Kg_per_Second := BankErosionTotal_kg_per_Second * (1 - (soilPercentFine * units.PERCENT_TO_PROPORTION))

		bankErosionFine.Set(idx, bankErosionFine_Kg_per_Second)
		bankErosionCoarse.Set(idx, bankErosionCoarse_Kg_per_Second)
	}

}

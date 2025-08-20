package rr

/*OW-SPEC
Simhyd:
  inputs:
    rainfall: mm
    pet: mm
  states:
    SoilMoistureStore:
    Groundwater:
    TotalStore:
  parameters:
    baseflowCoefficient: ''
    imperviousThreshold: ''
    infiltrationCoefficient: ''
		infiltrationShape: ''
		interflowCoefficient: ''
		perviousFraction: ''
		rainfallInterceptionStoreCapacity: ''
		rechargeCoefficient: ''
		soilMoistureStoreCapacity: ''
	outputs:
		runoff: mm
		quickflow: mm
		baseflow: mm
		store: mm
	implementation:
		function: simhyd
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		rainfall runoff

*/

import (
	"math"
)

const SOIL_ET_CONST = 10.0

func simhyd(rainfall []float64, pet []float64,
	initialStore float64, initialGW float64, initialTotalStore float64,
	baseflowCoefficient float64, imperviousThreshold float64, infiltrationCoefficient float64,
	infiltrationShape float64, interflowCoefficient float64, perviousFraction float64,
	risc float64, rechargeCoefficient float64, smsc float64,
	runoff, quickflow, baseflow, store []float64) (
	float64, // final store
	float64, // final GW
	float64) { // final total store
	nDays := len(rainfall)

	soilMoistureStore := initialStore
	gw := initialGW
	totalStore := initialTotalStore

	for i := 0; i < nDays; i++ {
		rainToday := rainfall[i]
		petToday := pet[i]

		perviousIncident := rainToday
		imperviousIncident := rainToday

		imperviousEt := math.Min(imperviousThreshold, imperviousIncident)

		imperviousRunoff := imperviousIncident - imperviousEt

		interceptionEt := math.Min(perviousIncident, math.Min(petToday, risc))

		throughfall := perviousIncident - interceptionEt

		soilMoistureFraction := soilMoistureStore / smsc

		infiltrationCapacity := infiltrationCoefficient * math.Exp(-infiltrationShape*soilMoistureFraction)
		infiltration := math.Min(throughfall, infiltrationCapacity)
		infiltrationXsRunoff := throughfall - infiltration

		interflowRunoff := interflowCoefficient * soilMoistureFraction * infiltration
		infiltrationAfterInterflow := infiltration - interflowRunoff
		recharge := rechargeCoefficient * soilMoistureFraction * infiltrationAfterInterflow
		soilInput := infiltrationAfterInterflow - recharge
		soilMoistureStore += soilInput

		soilMoistureFraction = soilMoistureStore / smsc

		gw += recharge

		if soilMoistureFraction > 1 {
			gw += soilMoistureStore - smsc
			soilMoistureStore = smsc
			soilMoistureFraction = 1
		}

		baseflowRunoff := baseflowCoefficient * gw
		gw -= baseflowRunoff

		soilEt := math.Min(soilMoistureStore, math.Min(petToday-interceptionEt, soilMoistureFraction*SOIL_ET_CONST))
		soilMoistureStore -= soilEt

		totalStore = (soilMoistureStore + gw) * perviousFraction

		//totalEt := (1 - perviousFraction) * imperviousEt + perviousFraction * (interceptionEt + soilEt);

		eventRunoff := (1-perviousFraction)*imperviousRunoff +
			perviousFraction*(infiltrationXsRunoff+interflowRunoff)

		totalRunoff := eventRunoff + perviousFraction*baseflowRunoff

		//effectiveRainfall := rainToday - totalEt;
		store[i] = soilMoistureStore
		baseflow[i] = baseflowRunoff * perviousFraction
		runoff[i] = totalRunoff
		quickflow[i] = eventRunoff
	}
	return soilMoistureStore, gw, totalStore
}

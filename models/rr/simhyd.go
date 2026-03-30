package rr

/*OW-SPEC
Simhyd:
  symbol: Sh
  reference: Chiew, F.H.S., Peel, M.C., Western, A.W., 2002. Application and testing of the simple rainfall-runoff model SIMHYD. In Mathematical Models of Small Watershed Hydrology and Applications, edited by V.P. Singh and D.K. Frevert, Water Resources Publications, Littleton, Colorado, 335-367.
  inputs:
    rainfall: mm
    pet: mm
  states:
    SoilMoistureStore:
    Groundwater:
    TotalStore:
  parameters:
    baseflowCoefficient: '[0.003,0.3] Baseflow linear recession parameter - fraction of groundwater store released as baseflow each timestep, default=0.1'
    imperviousThreshold: '[0,5]mm Rainfall threshold for impervious area runoff - depression storage on impervious surfaces, default=1.0'
    infiltrationCoefficient: '[0,400]mm Maximum infiltration loss parameter - controls infiltration capacity when soil is dry, default=200.0'
    infiltrationShape: '[0,10] Infiltration loss exponent - controls how rapidly infiltration capacity decreases as soil moisture increases, default=1.5'
    interflowCoefficient: '[0,1] Constant of proportionality in interflow equation - fraction of infiltrated water that becomes interflow scaled by soil moisture fraction, default=0.3'
    perviousFraction: '[0,1] Fraction of catchment area that is pervious, default=0.9'
    rainfallInterceptionStoreCapacity: '[0.5,5]mm Interception store capacity - maximum depth of rainfall intercepted by vegetation and lost to evaporation, default=2.5'
    rechargeCoefficient: '[0,1] Constant of proportionality in groundwater recharge equation - fraction of infiltrated water after interflow that recharges groundwater, default=0.1'
    soilMoistureStoreCapacity: '[1,500]mm Soil moisture storage capacity - maximum depth of water the soil moisture store can hold, default=200.0'
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

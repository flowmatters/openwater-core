package rr

/*OW-SPEC
Surm:
  inputs:
    rainfall: mm
    pet: mm
  states:
    SoilMoistureStore:
    Groundwater:
    TotalStore:
  parameters:
    bfac: ''
    coeff: ''
    dseep: ''
    fcFrac: ''
    fimp: ''
    rfac: ''
    smax: ''
    sq: ''
    thres: ''
	outputs:
		runoff: mm
		quickflow: mm
		baseflow: mm
		store: mm
	implementation:
		function: surm
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
	"github.com/flowmatters/openwater-core/data"
)

// const SOIL_ET_CONST = 10.0

func surm(rainfall, pet data.ND1Float64,
	initialStore, initialGW, initialTotalStore float64,
	bfac, coeff, dseep, fcFrac, fimp, rfac, smax, sq, thres float64,
	runoffTS, quickflowTS, baseflowTS, storeTS data.ND1Float64) (
	float64, // final store
	float64, // final GW
	float64) { // final total store
	nTimesteps := rainfall.Len1()

	soilMoistureStore := initialStore
	gw := initialGW
	totalStore := initialTotalStore
	idx := []int{0}
	fperv := 1-fimp

	for i := 0; i < nTimesteps; i++ {
		idx[0] = i
		rainThisTS := rainfall.Get(idx)
		petThisTS := pet.Get(idx)
		quickflow := 0.0

		imperviousRunoff := math.Max(rainThisTS-thres,0.0)
		quickflow += imperviousRunoff * fimp

		maxInfiltration := fperv * coeff * math.Exp(-sq * soilMoistureStore / smax)
		infiltration := math.Min(maxInfiltration,rainThisTS) 
		infiltrationExcess := rainThisTS - infiltration
		soilMoistureStore += infiltration

		saturationExcess := math.Max(soilMoistureStore - smax, 0.0)
		if soilMoistureStore > smax {
			soilMoistureStore = smax
		}
		perviousQuickflow := infiltrationExcess + saturationExcess
		quickflow += perviousQuickflow

		et := math.Max(math.Min(10*soilMoistureStore/smax,petThisTS*fperv),0.0)
		soilMoistureStore -= et

		recharge := math.Max(rfac*(soilMoistureStore/smax-fcFrac),0.0) * fperv
		gw += recharge
		soilMoistureStore -= recharge

		baseflow := bfac * gw
		seep := dseep * gw
		gw = math.Max(gw - baseflow - seep,0.0)
		// baseflow = baseflow 
		baseflow = baseflow * fperv

		runoff := quickflow + baseflow

		totalStore = soilMoistureStore + gw

		runoffTS.Set(idx, runoff)
		quickflowTS.Set(idx,quickflow)
		baseflowTS.Set(idx,baseflow)
		storeTS.Set(idx,totalStore)
	}

	return soilMoistureStore, gw, totalStore
}
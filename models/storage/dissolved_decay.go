package storage

import (
	"math"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/models/routing"
)

/* OW-SPEC
StorageDissolvedDecay:
	inputs:
		inflowMass: kg.s^-1
		inflow: m^3.s^-1
		outflow: m^3.s^-1
		storageVolume: m^3
	states:
		storedMass: kg
	parameters:
		DeltaT: '[1,86400] Timestep, default=86400'
		doStorageDecay:
		annualReturnInterval:
		bankFullFlow:
		medianFloodResidenceTime:
	outputs:
		decayedMass: kg
		outflowMass: kg.s^-1
	implementation:
		function: storageDissolvedDecay
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		storage, sediment
*/

func storageDissolvedDecay(inflowMass, storageInflow, storageOutflow, storageVolume data.ND1Float64, // inputs
	initialStoredMass float64,
	deltaT, doStorageDecay, annualReturnInterval, bankFullFlow, medianFloodResidenceTime float64,
	decayedMass, outflowMass data.ND1Float64) (storedMass float64) {

	if doStorageDecay < 0.5 {
		storedMass = routing.LumpedConstituentTransport(
			inflowMass, nil, storageOutflow, storageVolume,
			initialStoredMass,
			0.0, 0.0, deltaT,
			outflowMass,nil)
		return
	}

	storedMass = initialStoredMass
	nDays := inflowMass.Len1()
	idx := []int{0}

	for i := 0; i < nDays; i++ {
		idx[0] = i
		upstreamFlowMass := inflowMass.Get(idx) * deltaT
		storageVol := storageVolume.Get(idx)
		outflowRate := storageOutflow.Get(idx)

		availLoadForOutflow := 0.0
		dailyDecayedConstituentLoad := 0.0

		if outflowRate < bankFullFlow {
			dailyDecayedConstituentLoad = storedMass
			availLoadForOutflow = upstreamFlowMass
		} else {
			totalConstsituentLoad := upstreamFlowMass + storedMass

			if medianFloodResidenceTime <= 0 {
				dailyDecayedConstituentLoad = 0
				availLoadForOutflow = upstreamFlowMass + storedMass
			} else {
				propLost := math.Min(1, medianFloodResidenceTime/5)
				dailyDecayedConstituentLoad = propLost * totalConstsituentLoad
				availLoadForOutflow = totalConstsituentLoad - dailyDecayedConstituentLoad
			}
		}

		concentration := availLoadForOutflow / storageVol
		constituentRateInOutflow := concentration * outflowRate
		outflowMass.Set(idx, constituentRateInOutflow)
		decayedMass.Set(idx, dailyDecayedConstituentLoad)

		storedMass -= dailyDecayedConstituentLoad
		storedMass -= deltaT * constituentRateInOutflow
	}

	return
}

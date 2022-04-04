package storage

import (
	"math"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/m"
)

/* OW-SPEC
StorageParticulateTrapping:
	inputs:
		inflowLoad: kg.s^-1
		inflow: m^3.s^-1
		outflow: m^3.s^-1
		storage: m^3
	states:
		storedMass: kg
	parameters:
		DeltaT: '[1,86400] Timestep, default=86400'
		reservoirCapacity:
		reservoirLength: 'Length (m) of reservoir from dam wall to longest impounded water at dam capacity'
		subtractor: 'default=112'
		multiplier: 'default=800'
		lengthDischargeFactor:
		lengthDischargePower:
	outputs:
		trappedMass: kg
		outflowLoad: kg.s^-1
	implementation:
		function: storageParticulateTrapping
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		storage, sediment
*/

func storageParticulateTrapping(inflowMass, storageInflow, storageOutflow, storageVolume data.ND1Float64, // inputs
	initialStoredMass float64,
	deltaT, reservoirCapacity, reservoirLength, subtractor, multiplier, lengthDischargeFactor, lengthDischargePower float64,
	trappedMass, outflowLoad data.ND1Float64) (storedMass float64) {
	storedMass = initialStoredMass
	n := inflowMass.Len1()
	idx := []int{0}
	for i := 0; i < n; i++ {
		idx[0] = i
		incomingMass := inflowMass.Get(idx) * deltaT
		inflowRate := storageInflow.Get(idx)

		damTrappingPC := 0.0

		if (inflowRate > 0) && (reservoirLength > 0) {
			sedimentationIndex := math.Pow(reservoirCapacity, 2.0) / (lengthDischargeFactor * reservoirLength * math.Pow(inflowRate, 2.0))
			damTrappingPC = subtractor - (multiplier * math.Pow(sedimentationIndex, lengthDischargePower))
			damTrappingPC = m.MinFloat64(100.0, m.MaxFloat64(0.0, damTrappingPC))
		}

		dailyTrappedConstituentLoad := incomingMass * damTrappingPC / 100.0
		trappedMass.Set(idx, dailyTrappedConstituentLoad)

		storedMass = storedMass + incomingMass - dailyTrappedConstituentLoad

		storageOutflowRate := storageOutflow.Get(idx)
		storageWorkingVolume := storageOutflowRate*deltaT + storageVolume.Get(idx)

		concentration := storedMass / storageWorkingVolume
		massOutRate := storageOutflowRate * concentration
		storedMass = math.Max(storedMass - (massOutRate*deltaT),0.0)
		outflowLoad.Set(idx, massOutRate)
	}

	return
}

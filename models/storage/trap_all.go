package storage

import (
	"github.com/flowmatters/openwater-core/data"
)

/* OW-SPEC
StorageTrapAll:
	inputs:
		inflowMass: kg.s^-1
		inflow: m^3.s^-1
		outflow: m^3.s^-1
		storageVolume: m^3
	states:
		storedMass: kg
	parameters:
	outputs:
		trappedMass: kg
		outflowMass: kg.s^-1
	implementation:
		function: storageTrapAll
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		storage, sediment
*/

func storageTrapAll(inflowMass, storageInflow, storageOutflow, storageVolume data.ND1Float64, // inputs
	initialStoredMass float64,
	trappedMass, outflowMass data.ND1Float64) (storedMass float64) {

	trappedMass.CopyFrom(inflowMass)

	idx := []int{0}
	trappedMass.Set(idx, trappedMass.Get(idx)+initialStoredMass)
	storedMass = 0.0

	return
}

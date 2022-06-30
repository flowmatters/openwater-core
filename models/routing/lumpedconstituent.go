package routing

import (
	"github.com/flowmatters/openwater-core/data"
	//	"fmt"
)

const MINIMUM_VOLUME = 1e-2

/*OW-SPEC
LumpedConstituentRouting:
  inputs:
		inflowLoad: kg.s^-1
		lateralLoad: kg.s^-1
		outflow: m^3.s^-1
		storage: m^3
	states:
		storedMass:
	parameters:
		X: '[0,1] Weighting'
		pointInput: kg.s^-1
		DeltaT: '[1,86400] Timestep, default=86400'
	outputs:
		outflowLoad: kg.s^-1
		pointSourceLoad: kg.s^-1
	implementation:
		function: LumpedConstituentTransport
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		constituent routing
*/

func LumpedConstituentTransport(inflowLoads, lateralLoads, outflows, storage data.ND1Float64,
	initialStoredMass float64,
	x, pointInput, deltaT float64,
	outflowLoads, pointSourceLoad data.ND1Float64) (storedMass float64) {
	storedMass = initialStoredMass
	nDays := inflowLoads.Len1()

	idx := []int{0}
	for i := 0; i < nDays; i++ {
		idx[0] = i
		inflowLoad := inflowLoads.Get(idx)
		lateralLoad := lateralLoads.Get(idx)
		totalLoadIn := (inflowLoad + lateralLoad + pointInput) * deltaT

		outflowR := outflows.Get(idx)
		outflowV := outflowR * deltaT
		storedV := storage.Get(idx)

		workingMass := storedMass + totalLoadIn
		workingVol := outflowV + storedV

		if workingVol < MINIMUM_VOLUME {
			storedMass = 0.0
			outflowLoads.Set(idx, 0.0)
			if pointSourceLoad != nil {
				pointSourceLoad.Set(idx,0.0)
			}
			continue
		}

		concentration := workingMass / workingVol
		storedMass = concentration * storedV
		outflowLoad := concentration * outflowR

		outflowLoads.Set(idx, outflowLoad)
		if pointSourceLoad != nil {
			pointSourceLoad.Set(idx,pointInput)
		}
	}
	return storedMass
}

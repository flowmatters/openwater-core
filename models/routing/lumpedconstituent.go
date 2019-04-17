package routing

import (
	"github.com/flowmatters/openwater-core/data"
	//	"fmt"
)

const MINIMUM_VOLUME = 1e-4

/*OW-SPEC
LumpedConstituentRouting:
  inputs:
		inflowLoad: kg.s^-1
		lateralLoad: kg.s^-1
    inflow: m^3.s^-1
		outflow: m^3.s^-1
		storage: m^3
	states:
		storedMass:
	parameters:
		X: '[0,1] Weighting'
		DeltaT: '[1,86400] Timestep, default=86400'
	outputs:
		outflowLoad: kg.s^-1
	implementation:
		function: lumpedConstituents
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		constituent routing
*/

func lumpedConstituents(inflowLoads, lateralLoads, inflows, outflows, storage data.ND1Float64,
	storedMass float64,
	x, deltaT float64,
	outflowLoads data.ND1Float64) float64 {
	nDays := inflows.Len1()

	idx := []int{0}
	for i := 0; i < nDays; i++ {
		idx[0] = i
		inflowLoad := inflowLoads.Get(idx) * deltaT
		lateralLoad := lateralLoads.Get(idx) * deltaT
		workingMass := storedMass + inflowLoad + lateralLoad

		outflowV := outflows.Get(idx) * deltaT
		storedV := storage.Get(idx)

		workingVol := outflowV + storedV
		if workingVol < MINIMUM_VOLUME {
			storedMass = workingMass
			outflowLoads.Set(idx, 0.0)
			continue
		}

		concentration := workingMass / workingVol
		storedMass = concentration * storedV
		outflowLoad := concentration * outflowV / deltaT

		outflowLoads.Set(idx, outflowLoad)
	}
	return storedMass
}

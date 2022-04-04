package routing

import (
	"math"

	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
ConstituentDecay:
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
		halfLife:
		DeltaT: '[1,86400] Timestep, default=86400'
	outputs:
		outflowLoad: kg.s^-1
	implementation:
		function: constituentDecay
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		constituent transport
*/

func constituentDecay(inflowLoads, lateralLoads, inflows, outflows, storage data.ND1Float64,
	storedMass float64,
	x, halflife, deltaT float64,
	outflowLoads data.ND1Float64) float64 {
	const MINIMUM_VOLUME=0.01
	n := inflowLoads.Len1()
	idx := []int{0}

	for day := 0; day < n; day++ {
		idx[0] = day

		if halflife > 0 {
			storedMass *= math.Pow(2.0, -deltaT/halflife)
		}

		inflowLoad := inflowLoads.Get(idx) * deltaT
		lateralLoad := lateralLoads.Get(idx) * deltaT
		workingMass := storedMass + inflowLoad + lateralLoad

		outflowR := outflows.Get(idx)
		outflowV := outflowR * deltaT
		storedV := storage.Get(idx)

		workingVol := outflowV + storedV
		if workingVol < MINIMUM_VOLUME {
			storedMass = 0.0
			outflowLoads.Set(idx, 0.0)
			continue
		}

		concentration := workingMass / workingVol
		outflowLoad := concentration * outflowR
		storedMass = workingMass - outflowLoad * deltaT

		outflowLoads.Set(idx, outflowLoad)
	}
	return storedMass
}

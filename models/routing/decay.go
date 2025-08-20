package routing

import (
	"math"
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
		decayedLoad: kg.s^-1
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

func constituentDecay(inflowLoads, lateralLoads, inflows, outflows, storage []float64,
	storedMass float64,
	x, halflife, deltaT float64,
	decayedLoad, outflowLoads []float64) float64 {
	const MINIMUM_VOLUME = 0.01
	n := len(inflowLoads)

	for day := 0; day < n; day++ {

		decayedAmount := 0.0
		if halflife > 0 {
			fraction := math.Pow(2.0, -deltaT/halflife)
			decayedAmount = (1 - fraction) * storedMass
			decayedLoad[day] = decayedAmount / deltaT
			storedMass *= fraction
		}

		inflowLoad := inflowLoads[day] * deltaT
		lateralLoad := lateralLoads[day] * deltaT
		workingMass := storedMass + inflowLoad + lateralLoad

		outflowR := outflows[day]
		outflowV := outflowR * deltaT
		storedV := storage[day]

		workingVol := outflowV + storedV
		if workingVol < MINIMUM_VOLUME {
			storedMass = 0.0
			outflowLoads[day] = 0.0
			continue
		}

		concentration := workingMass / workingVol
		outflowLoad := concentration * outflowR
		storedMass = workingMass - outflowLoad*deltaT

		outflowLoads[day] = outflowLoad
	}
	return storedMass
}

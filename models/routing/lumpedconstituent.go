package routing

//	"fmt"

const MINIMUM_VOLUME = 1e-2

/*OW-SPEC
LumpedConstituentRouting:
  symbol: LC
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

func LumpedConstituentTransport(inflowLoads, lateralLoads, outflows, storage []float64,
	initialStoredMass float64,
	x, pointInput, deltaT float64,
	outflowLoads, pointSourceLoad []float64) (storedMass float64) {
	storedMass = initialStoredMass
	nDays := len(inflowLoads)

	for i := 0; i < nDays; i++ {
		inflowLoad := inflowLoads[i]
		lateralLoad := lateralLoads[i]
		totalLoadIn := (inflowLoad + lateralLoad + pointInput) * deltaT

		outflowR := outflows[i]
		outflowV := outflowR * deltaT
		storedV := storage[i]

		workingMass := storedMass + totalLoadIn
		workingVol := outflowV + storedV

		if workingVol < MINIMUM_VOLUME {
			storedMass = 0.0
			outflowLoads[i] = 0.0
			if pointSourceLoad != nil {
				pointSourceLoad[i] = 0.0
			}
			continue
		}

		concentration := workingMass / workingVol
		storedMass = concentration * storedV
		outflowLoad := concentration * outflowR

		outflowLoads[i] = outflowLoad
		if pointSourceLoad != nil {
			pointSourceLoad[i] = pointInput
		}
	}
	return storedMass
}

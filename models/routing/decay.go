package routing

import "github.com/flowmatters/openwater-core/data"

/*OW-SPEC
ConstituentDecay:
	inputs:
		incoming:
	states:
		storedMass:
	parameters:
		halflife:
		durationInSeconds:
	outputs:
		outgoing:
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

func constituentDecay(incoming data.ND1Float64,
	storedMass float64,
	halflife, durationInSeconds float64,
	outgoing data.ND1Float64) float64 {
	n := incoming.Len1()
	idx := []int{0}

	for day := 0; day < n; day++ {
		idx[0] = day

		//		load :=
	}
	return 0.0
}

package functions

import (
	"github.com/flowmatters/openwater-core/data"
)

/*OW-SPEC
Gate:
	inputs:
		trigger:
		incoming:
	states:
	parameters:
	outputs:
		outgoing: kg
	implementation:
		function: gate
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		function
*/

func gate(trigger, incoming data.ND1Float64,
	outgoing data.ND1Float64) {

	n := trigger.Len1()
	idx := []int{0}

	for day := 0; day < n; day++ {
		idx[0] = day
		t := trigger.Get(idx)
		i := incoming.Get(idx)

		if t > 0 {
			outgoing.Set(idx, i)
		} else {
			outgoing.Set(idx, 0.0)
		}
	}
}

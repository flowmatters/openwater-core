package functions

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

func gate(trigger, incoming []float64,
	outgoing []float64) {

	n := len(trigger)

	for day := 0; day < n; day++ {
		t := trigger[day]
		i := incoming[day]

		if t > 0 {
			outgoing[day] = i
		} else {
			outgoing[day] = 0.0
		}
	}
}

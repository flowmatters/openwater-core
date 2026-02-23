package generation

const (
	EFFECTIVELY_ZERO = 1e-8
)

/*OW-SPEC
PassLoadIfFlow:
	symbol: PL
	inputs:
		flow: m^3.s^-1
		inputLoad:
	states:
	parameters:
		scalingFactor:
	outputs:
		outputLoad: kg
	implementation:
		function: passLoadIfFlow
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		constituent generation
*/

func passLoadIfFlow(flow, inputLoad []float64,
	scalingFactor float64,
	outputLoad []float64) {

	if scalingFactor == 0.0 {
		return
	}

	n := len(flow)

	for day := 0; day < n; day++ {
		f := flow[day]
		l := inputLoad[day]

		if f > EFFECTIVELY_ZERO {
			outputLoad[day] = l * scalingFactor
		} else {
			outputLoad[day] = 0.0
		}
	}
}

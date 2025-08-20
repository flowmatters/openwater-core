package routing

import (
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/m"
)

/*OW-SPEC
Lag:
	inputs:
		inflow: m^3.s^-1
	states:
		lagged:
	parameters:
		timeLag:
	outputs:
		outflow: m^3.s^-1
	implementation:
		function: lag
		type: scalar
		lang: go
		outputs: params
	init:
		function: initLag
		type: scalar
		lang: go
	extractstates:
		function: extractLagStates
		packfunc: packLagStates
		type: scalar
		lang: go
	tags:
		lag
*/

func initLag(timeLag float64) data.ND2[float64] {
	lags := make([]float64, int(timeLag))

	result := packLagStates(lags)
	return result
}

func extractLagStates(states []float64) []float64 {
	return states
}

func packLagStates(lagged []float64) data.ND2[float64] {
	result := data.NewArray2D[float64](1, len(lagged))
	return result
}

func lag(inflow []float64,
	lagged []float64,
	timeLag float64,
	outflow []float64) []float64 {

	lagSteps := int(timeLag)

	if lagSteps == 0 {
		outflow.CopyFrom(inflow)
		return lagged
	}

	n := len(outflow)
	for i := 0; i < m.Min[int](lagSteps, n); i++ {
		outflow[i] = lagged[i]
	}

	for i := lagSteps; i < n; i++ {
		outflow[i] = inflow[i-lagSteps]
	}
	m := len(inflow)
	if lagSteps > m {
		for i := m; i < lagSteps; i++ {
			lagged[i-m] = lagged[i]
		}

		for i := 0; i < m; i++ {
			lagged[i+m] = inflow[i]
		}
	} else {
		for i := 0; i < lagSteps; i++ {
			lagged[i] = inflow[m-lagSteps+i]
		}
	}

	return lagged
}

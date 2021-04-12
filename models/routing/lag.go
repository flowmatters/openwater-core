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

func initLag(timeLag float64) data.ND2Float64 {
	lags := make([]float64, int(timeLag))

	result := packLagStates(lags)
	return result
}

func extractLagStates(states data.ND1Float64) []float64 {
	return states.Unroll()
}

func packLagStates(lagged []float64) data.ND2Float64 {
	result := data.NewArray2DFloat64(1, len(lagged))
	return result
}

func lag(inflow data.ND1Float64,
	lagged []float64,
	timeLag float64,
	outflow data.ND1Float64) []float64 {

	lagSteps := int(timeLag)

	if lagSteps == 0 {
		outflow.CopyFrom(inflow)
		return lagged
	}

	idx := []int{0}
	for i := 0; i < m.MinInt(lagSteps, outflow.Len1()); i++ {
		idx[0] = i
		outflow.Set(idx, lagged[i])
	}

	idxInflow := []int{0}
	for i := lagSteps; i < outflow.Len1(); i++ {
		idx[0] = i
		idxInflow[0] = i - lagSteps
		outflow.Set(idx, inflow.Get(idxInflow))
	}

	if lagSteps > inflow.Len1() {
		for i := inflow.Len1(); i < lagSteps; i++ {
			lagged[i-inflow.Len1()] = lagged[i]
		}

		for i := 0; i < inflow.Len1(); i++ {
			idxInflow[0] = i
			lagged[i+inflow.Len1()] = inflow.Get(idxInflow)
		}
	} else {
		for i := 0; i < lagSteps; i++ {
			idxInflow[0] = inflow.Len1() - lagSteps + i
			lagged[i] = inflow.Get(idxInflow)
		}
	}

	return lagged
}

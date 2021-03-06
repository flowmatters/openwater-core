package routing

import (
	"github.com/flowmatters/openwater-core/data"
	//	"fmt"
)

/*OW-SPEC
Muskingum:
  inputs:
		inflow: m^3.s^-1
		lateral: m^3.s^-1
	states:
		S:
		prevInflow:
		prevOutflow:
	parameters:
		K: '[0,200000]s Constant'
		X: '[0,1] Weighting'
		DeltaT: '[1,86400] Timestep, default=86400'
	outputs:
		outflow: m^3.s^-1
	implementation:
		function: muskingum
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		flow routing
*/

func muskingum(inflows, laterals data.ND1Float64,
	s, prevInflow, prevOutflow float64,
	k, x, deltaT float64,
	outflows data.ND1Float64) (float64, float64, float64) {
	idx := []int{0}
	nDays := inflows.Len1()

	kx2 := 2 * k * x
	denom := (2*k*(1-x) + deltaT)
	a1 := (deltaT - kx2) / denom
	a2 := (deltaT + kx2) / denom
	a3 := (2*k*(1-x) - deltaT) / denom

	for i := 0; i < nDays; i++ {
		idx[0] = i
		inflow := inflows.Get(idx)
		lateral := laterals.Get(idx)

		outflow := a1*(inflow+lateral) + a2*prevInflow + a3*prevOutflow
		outflows.Set(idx, outflow)

		prevOutflow = outflow
		prevInflow = inflow
	}

	return s, prevInflow, prevOutflow
}

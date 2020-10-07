package routing

import (
	// "fmt"
	"math"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/fn"
)

const (
	massBalanceLimit = 1e-3
	convergenceLimit = 1e-8
	maxIterations    = 20
)

//	"fmt"

/*OW-SPEC
StorageRouting:
  inputs:
		inflow: m^3.s^-1
		lateral: m^3.s^-1
		rainfall: mm
		evap: mm
	states:
		S:
		prevInflow:
		prevOutflow:
	parameters:
		InflowBias:
		RoutingConstant:
		RoutingPower:
		area:
		deadStorage:
		DeltaT: '[1,86400] Timestep, default=86400'
	outputs:
		outflow: m^3.s^-1
		storage: m^3
	implementation:
		function: storageRouting
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		flow routing
*/

func storageRouting(inflows, laterals, rainfall, evap data.ND1Float64,
	s, prevInflow, prevOutflow float64,
	bias, k, x, area, deadStorage, deltaT float64,
	outflows, storages data.ND1Float64) (float64, float64, float64) {
	n := inflows.Len1()
	idx := []int{0}

	Klimit := 0.0
	Qlimit := 0.0
	Koffset := 0.0

	if math.Abs(bias) < 0.001 {
		bias = 0.0
		if x > 1.0 {
			Qlimit = 1e37
		}
		Klimit = k
	} else if math.Abs(x-1.0) < 0.001 {
		x = 1.0
		Klimit = k
	} else { // Non-linear index flow routing
		Klimit = deltaT / bias
		Qlimit = math.Pow((Klimit / (x * k)), (1.0 / (x - 1.0)))
		if x < 1.0 {
			Koffset = Qlimit * Klimit * (1.0 - x) / x
		}
	}
	qi := 0.0
	outflow := 0.0
	storage := 0.0
	// fmt.Printf("\n\n\n================= x=%f, bias=%f, kLimit=%f, qlimit=%f, koffset=%f ===============\n",
	// x, bias, Klimit, Qlimit, Koffset)

	for i := 0; i < n; i++ {
		idx[0] = i
		if bias > 0.999 && i < 10 {
			// fmt.Printf("\n==== TS %d, inflow=%f, laterial=%f ====\n", i, inflows.Get(idx), laterals.Get(idx))
		}
		inflow := inflows.Get(idx)
		lateral := laterals.Get(idx)
		evapRate := evap.Get(idx) - rainfall.Get(idx)/deltaT
		qi, outflow, storage = calcOutflow(i, inflow, lateral, bias, qi, outflow, storage,
			evapRate, area, deadStorage, deltaT, x, k, Qlimit, Klimit, Koffset)
		outflows.Set(idx, outflow)
		storages.Set(idx, storage)
	}

	return 0.0, 0.0, 0.0
}

func calcOutflow(timestep int, inflow, lateral, bias, prevQi, prevOutflow, prevStorage, netEvapRate,
	area, deadStorage, duration, routingPower, routingConstant, Qlimit, Klimit, Koffset float64) (qi, outflow, storage float64) {
	qi = prevQi
	outflow = prevOutflow
	storage = prevStorage
	totalInflow := inflow + lateral
	initialFluxMax := (math.Max(0.0, prevStorage) / duration) + totalInflow // inflow + lateral

	evaluateRouting := func(q float64) (massBalance, outflow, SIndex float64) {
		return runRouting(q, totalInflow, 0, initialFluxMax, storage, area, netEvapRate, deadStorage, duration,
			bias, routingPower, routingConstant, Qlimit, Klimit, Koffset)
	}
	evaluateRoutingMassBalance := func(q float64) float64 {
		delta, _, _ := evaluateRouting(q)
		return delta
	}

	slopeOfMassBalance := func(q float64) (dMdq float64) {
		dMdq = duration / (1.0 - bias)

		if (routingPower <= 1.0 && qi < Qlimit) || (routingPower > 1.0 && qi > Qlimit) {
			dMdq += Klimit
		} else if qi > 0 {
			dMdq += routingConstant * routingPower * math.Pow(qi, routingPower-1.0)
		}
		return
	}

	minQI := bias * (inflow + lateral)
	delta := evaluateRoutingMassBalance(minQI)
	// if bias > 0.999 && timestep < 10 {
	// 	fmt.Printf("calcOutflow-1, minQI=%f,delta=%f,massBalanceLimit=%f\n", minQI, delta, massBalanceLimit)
	// }
	if delta >= massBalanceLimit {
		// Qindexmin is not small enough with zero outflow, so lets call it zero outflow
		qi = minQI
		outflow = 0.0
		// fmt.Printf("calcOutflow-2, outflow=0, storage=%f\n", storage)
		return
	}

	if delta >= -massBalanceLimit {
		// Gets here if:
		//   InflowBias >= .999 in this case massbalance= 0 and ;
		//   No outflow is the solution
		qi = minQI

		delta, outflow, storage = evaluateRouting(qi)
		// if bias < 0.999 {
		// 	fmt.Printf("calcOutflow-3, qi=%f,delta=%f,outflow=%f,storage=%f\n", qi, delta, outflow, storage)
		// }
		return
	}

	// Save the initial minimum mass balance error
	// massBalanceMin := delta

	fluxmax := initialFluxMax

	//Determine the maximum net evaporation flux assuming no outflow
	netEvaporationFlux := math.Min(fluxmax, area*netEvapRate)
	fluxmax -= netEvaporationFlux

	// Maximum outflow is the inflow plus storage (fluxmax) less all fluxes
	// Therefore qindexmax= xI + (1-x)*Maximum Outflow
	maxQI := minQI + (1.0-bias)*math.Max(0.0, fluxmax)
	if maxQI <= minQI {
		//Fluxes exceed inflow and storage so outflow is zero
		qi = minQI
		outflow = 0.0
		delta = 0.0
		// fmt.Printf("calcOutflow-4, qi=%f,delta=%f,outflow=%f,storage=%f\n", qi, delta, outflow, storage)
		return
	}

	delta = evaluateRoutingMassBalance(maxQI)
	if delta < massBalanceLimit {
		//Solution is the maximum possible index flow
		qi = maxQI
		outflow = math.Max(0.0, initialFluxMax-netEvaporationFlux)
		delta = 0.0
		// fmt.Printf("calcOutflow-5, qi=%f,delta=%f,outflow=%f,storage=%f\n", qi, delta, outflow, storage)
		return
	}

	// massBalanceMax := delta
	//qindex is based on the current inflow and previous outflow

	// make sure it does not exceed the qindexmin and qindexmax limits
	if qi <= minQI || qi >= maxQI {
		qi = (minQI + maxQI) * 0.5
	}

	//Determine the associated storage, maximum storage and mass balance error
	delta = evaluateRoutingMassBalance(qi)

	if math.Abs(delta) < massBalanceLimit {
		// fmt.Printf("calcOutflow-6, qi=%f,delta=%f,outflow=%f,storage=%f\n", qi, delta, outflow, storage)
		return
	}

	qi, delta = fn.FindRoot(evaluateRoutingMassBalance, slopeOfMassBalance, minQI, minQI, maxQI, massBalanceLimit, convergenceLimit, maxIterations)

	delta, outflow, storage = evaluateRouting(qi)
	if math.Abs(delta) > massBalanceLimit {
		// fmt.Printf("Timestep = %d, delta=%f, outflow=%f, storage=%f\n", timestep, delta, outflow, storage)
	}
	return
}

func runRouting(qIndex, inflow, lateral, initialFluxMax, storage, area, netEvapRate, deadStorage, duration,
	bias, routingPower, routingConstant, Qlimit, Klimit, Koffset float64) (massBalance, outflow, SIndex float64) {
	// eqn := 0
	if qIndex <= 0.0 {
		SIndex = deadStorage
	} else if (routingPower <= 1.0 && qIndex < Qlimit) || (routingPower > 1.0 && qIndex > Qlimit) {
		// eqn = 1
		SIndex = Klimit*qIndex + deadStorage
		// fmt.Printf("SIndex = %f = Klimit * qIndex + deadStorage = %f * %f + %f\n", SIndex, Klimit, qIndex, deadStorage)
	} else {
		// eqn = 2
		SIndex = routingConstant*math.Pow(qIndex, routingPower) - Koffset + deadStorage
	}

	fluxmax := initialFluxMax
	netEvaporationFlux := math.Min(fluxmax, area*netEvapRate)
	// TODO Should account for change in area
	fluxmax -= netEvaporationFlux

	newStorage := math.Max((storage + (inflow+lateral-netEvaporationFlux)*duration), 0.0)
	// corrected := false
	// if newStorage < SIndex {
	// 	// fmt.Printf("Correcting SIndex. Used eqn %d. newStorage=%f, SIndex=%f, inflow=%f, bias=%f,x=%f\n",
	// 	// 	eqn, newStorage, SIndex, inflow, bias, routingPower)
	// 	// corrected = true
	// 	SIndex = newStorage
	// }

	if bias < .999 {
		massBalance = (qIndex-bias*(inflow+lateral))*duration/(1.0-bias) + SIndex - newStorage
	} else {
		massBalance = 0.0
	}

	outflow = math.Max(0, newStorage-SIndex) / duration
	//SIndex = SIndex - outflow

	// if corrected {
	// 	fmt.Printf("massBalance after correction = %f, SIndex=%f, outflow=%f\n", massBalance, SIndex, outflow)
	// }

	return
}

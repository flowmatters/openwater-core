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

//dsLaterals,
func storageRouting(inflows, laterals,  rainfall, evap data.ND1Float64,
	s, prevInflow, prevOutflow float64,
	bias, k /*RoutingConstant */, x /*RoutingPower*/, area, deadStorage, deltaT float64,
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
	} else {
		Klimit = deltaT / bias
		Qlimit = math.Pow((Klimit / (x * k)), (1.0 / (x - 1.0)))
		if x < 1.0 {
			Koffset = Qlimit * Klimit * (1.0 - x) / x
		}
	}
	qi := 0.0
	outflow := 0.0
	storage := 0.0

	for i := 0; i < n; i++ {
		idx[0] = i

		inflow := inflows.Get(idx)
		lateral := laterals.Get(idx)

		if math.IsNaN(inflow) {
			fmt.Printf("inflow=%f\n",inflow)
			fmt.Printf("lateral=%f\n",lateral)
			fmt.Printf("storage=%f\n",storage)
			fmt.Printf("deltaT=%f\n",deltaT)
		}
		evapRate := (evap.Get(idx) - rainfall.Get(idx))/deltaT

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
	// totalInflow := inflow + lateral
	initialFluxMax := (math.Max(0.0, prevStorage) / duration) + inflow // inflow + lateral

	evaluateRouting := func(q float64) (massBalance, outflow, SIndex float64) {
		return runRouting(q, inflow, lateral, initialFluxMax, prevStorage, area, netEvapRate, deadStorage, duration,
			bias, routingPower, routingConstant, Qlimit, Klimit, Koffset)
	}
	evaluateRoutingMassBalance := func(q float64) float64 {
		delta, _, _ := evaluateRouting(q)
		return delta
	}

	slopeOfMassBalance := func(q float64) (dMdq float64) {
		dMdq = duration / (1.0 - bias)

		if (routingPower <= 1.0 && q < Qlimit) || (routingPower > 1.0 && q > Qlimit) {
			dMdq += Klimit
		} else if q > 0 {
			dMdq += routingConstant * routingPower * math.Pow(q, routingPower-1.0)
		}

		return
	}

	if math.IsNaN(bias) || math.IsNaN(inflow) || math.IsNaN(lateral) {
		fmt.Printf("bias=%f\n",bias)
		fmt.Printf("inflow=%f\n",inflow)
		fmt.Printf("lateral=%f\n",lateral)
		panic("NAN!")
	}
	minQI := bias * (inflow + lateral)
	delta,outflow,storage := evaluateRouting(minQI)
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
	maxQI := minQI + (1.0-bias)*math.Max(0.0, fluxmax+lateral)
	if maxQI <= minQI {
		//Fluxes exceed inflow and storage so outflow is zero
		qi = minQI
		outflow = 0.0
		delta = 0.0
		// fmt.Printf("calcOutflow-4, qi=%f,delta=%f,outflow=%f,storage=%f\n", qi, delta, outflow, storage)
		return
	}

	delta, outflow, storage = evaluateRouting(maxQI)
	if delta < massBalanceLimit {
		//Solution is the maximum possible index flow
		qi = maxQI
		outflow = math.Max(0.0, initialFluxMax-netEvaporationFlux)

		delta = 0.0
		storage = math.Max((prevStorage + (inflow+lateral-netEvaporationFlux-outflow)*duration), 0.0)

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
	delta, outflow, storage = evaluateRouting(qi)

	if math.Abs(delta) < massBalanceLimit {
		// fmt.Printf("calcOutflow-6, qi=%f,delta=%f,outflow=%f,storage=%f\n", qi, delta, outflow, storage)
		return
	}

	qi, delta = fn.FindRoot(evaluateRoutingMassBalance, slopeOfMassBalance, minQI, minQI, maxQI, massBalanceLimit, convergenceLimit, maxIterations)

	if math.IsNaN(delta) {
		fmt.Printf("delta is NaN.\n")
		fmt.Printf("qi=%f\n",qi)
		fmt.Printf("minQI=%f\n",minQI)
		fmt.Printf("maxQI=%f\n",maxQI)
		fmt.Printf("massBalanceLimit=%f\n",massBalanceLimit)
		fmt.Printf("convergenceLimit=%f\n",convergenceLimit)
		fmt.Printf("maxIterations=%d\n",maxIterations)
		panic("delta is NaN")
	}
	delta, outflow, storage = evaluateRouting(qi)
	if math.Abs(delta) > massBalanceLimit {
		// fmt.Printf("Timestep = %d, delta=%f, outflow=%f, storage=%f\n", timestep, delta, outflow, storage)
	}
	if math.IsNaN(outflow){
		fmt.Printf("outflow=%f\n",outflow)
		fmt.Printf("storage=%f\n",storage)
		fmt.Printf("delta=%f\n",delta)
		panic("outflow is nan")
	}
	return
}

func runRouting(qIndex, inflow, lateral, initialFluxMax, storage, area, netEvapRate, deadStorage, duration,
	bias, routingPower, routingConstant, Qlimit, Klimit, Koffset float64) (massBalance, outflow, SIndex float64) {
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
	if math.IsNaN(outflow){
		fmt.Printf("outflow=%f\n",outflow)
		fmt.Printf("newStorage=%f\n",newStorage)
		fmt.Printf("SIndex=%f\n",SIndex)
		fmt.Printf("duration=%f\n",duration)
		panic("outflow is nan")
	}
	//SIndex = SIndex - outflow

	// if corrected {
	// 	fmt.Printf("massBalance after correction = %f, SIndex=%f, outflow=%f\n", massBalance, SIndex, outflow)
	// }

	return
}

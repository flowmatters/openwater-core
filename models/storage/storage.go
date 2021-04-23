package storage


import (
	"fmt"
	"math"
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/fn"
	"github.com/flowmatters/openwater-core/conv/units"
)

const (
	MIN_TIMESTEP = 60.0
	ALLOWED_REL_ERROR_RELEASE_RATE = 1e-4
	ALLOWED_ABS_ERROR_RELEASE_RATE = 1e-2
)

/* OW-SPEC
Storage:
  inputs:
		rainfall: mm
		pet: mm
		inflow: m^3.s^-1
		demand: m^3.s^-1
	states:
		currentVolume: m^3
		level: m
		area: m^2
	parameters:
		DeltaT: '[1,86400] Timestep, default=86400'
		nLVA: ''
		levels[nLVA]:
		volumes[nLVA]:
		areas[nLVA]:
		minRelease[nLVA]:
		maxRelease[nLVA]:
	outputs:
		volume: m^3
		outflow: m^3.s^-1
	implementation:
		function: storageWaterBalance
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		storage
*/

func storageWaterBalance(rainfallTS, petTS, inflowTS, demandTS data.ND1Float64, 
												 initialVolume, initialLevel, initialArea float64,
												 deltaT float64,
												 nLVA int,
												 levels, volumes, areas, minRelease, maxRelease data.ND1Float64,
												 volumeTS, outflowTS data.ND1Float64) (volume, level, area float64) {
	volume = initialVolume
	n := rainfallTS.Len1()
	nSubtimeSteps := 0
	idx := []int{0}
	for i := 0; i < n; i++ {
		idx[0] = i
		timeRemaining := deltaT
		outflowVolume := 0.0

		inflow := inflowTS.Get(idx)
		demand := demandTS.Get(idx)

		rainfall := rainfallTS.Get(idx)
		pet := petTS.Get(idx) / deltaT
		netAtmosphericFluxIn := (rainfall - pet) * units.MILLIMETRES_TO_METRES

		subtimestep := deltaT

		releaseRate := func (demand,vol float64) (release float64) {
			minRel := fn.Piecewise(vol,volumes,minRelease)
			maxRel := fn.Piecewise(vol,volumes,maxRelease)

			if demand < minRel {
				release = minRel
			} else if demand > maxRel {
				release = maxRel
			} else {
				release = demand
			}
			return
		}

		releaseRatesCloseEnough := func (a, b float64) bool {
			absError := math.Abs(a-b)
			if absError > ALLOWED_ABS_ERROR_RELEASE_RATE {
				return false
			}
			if (a == 0.0) || (b == 0.0) {
				return true
			}
			return (absError / a) > ALLOWED_REL_ERROR_RELEASE_RATE
		}

		for timeRemaining > 0 {
			nSubtimeSteps += 1
			subtimestep = math.Min(subtimestep,timeRemaining)

			estOutflow := releaseRate(demand,volume)
			netSurfaceFluxIn := inflow - estOutflow

			area = fn.Piecewise(volume,volumes,areas)
			netAtmosphericFluxInRate := netAtmosphericFluxIn * area

			netFluxIn := netSurfaceFluxIn + netAtmosphericFluxInRate
			testVol := volume + netFluxIn * subtimestep
			estOutflowAfter := releaseRate(demand,testVol)

			if releaseRatesCloseEnough(estOutflow,estOutflowAfter) || (subtimestep < MIN_TIMESTEP) {
				outflowVolume += estOutflow * subtimestep
				volume = testVol
				timeRemaining -= subtimestep
			} else {
				subtimestep /= 2.0
				if subtimestep < MIN_TIMESTEP {
					fmt.Println("Hit minimum timestep")
				}
			}
		}

		volumeTS.Set(idx,volume)

		outflowRate := outflowVolume / deltaT
		outflowTS.Set(idx,outflowRate)
	}

	if n > 0 {
		fmt.Printf("Storage ran %d timesteps in %d subtimesteps (average %f subtimesteps/timestep)\n",n,nSubtimeSteps,float64(nSubtimeSteps/n))
	}

	level = fn.Piecewise(volume,volumes,levels)
	area = fn.Piecewise(volume,volumes,areas)

	return
}



/*
  TODO
	* Level needed?
	


	*/
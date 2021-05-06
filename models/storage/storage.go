package storage


import (
	"errors"
	"fmt"
	"math"
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/fn"
	"github.com/flowmatters/openwater-core/conv/units"
)

const (
	MIN_TIMESTEP_SECONDS = 30
	ALLOWED_REL_ERROR_RELEASE_RATE = 1e-5
	ALLOWED_ABS_ERROR_RELEASE_RATE = 1e-4
	MAX_SUBTIMESTEPS = 600000
)

/* OW-SPEC
Storage:
  inputs:
		rainfall: mm
		pet: mm
		inflow: m^3.s^-1
		demand: m^3.s^-1
		targetMinimumVolume: m^3
		targetMinimumCapacity: m^3
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

func checkStorageConfiguration(nLVA int, volumes data.ND1Float64) (error) {
	if nLVA == 0 {
		return errors.New("No points in LVA table or release curves!" )
	}

	if volumes.Maximum() <= 0.0 {
		return errors.New("No volumes")
	}

	return nil
}

func storageWaterBalance(rainfallTS, petTS, inflowTS, demandTS, targetMinimumVolume, targetMinimumCapacity data.ND1Float64,
												 initialVolume, initialLevel, initialArea float64,
												 deltaT float64,
												 nLVA int,
												 levels, volumes, areas, minRelease, maxRelease data.ND1Float64,
												 volumeTS, outflowTS data.ND1Float64) (volume, level, area float64) {
	idxCurve0 := []int{0}
	idxCurveN := []int{nLVA-1}

	volCurveMin := volumes.Get(idxCurve0)
	volCurveMax := volumes.Get(idxCurveN)

	// Water balance functions
	cappedPiecewise := func(vol float64,ys data.ND1Float64) float64 {
		if vol < volCurveMin {
			return ys.Get(idxCurve0)
		}
		if vol > volCurveMax {
			return ys.Get(idxCurveN)
		}
		res,err := fn.Piecewise(vol,volumes,ys)
		if err != nil {
			panic(err)
		}
		return res
	}

	releaseRate := func (demand,vol float64) float64 {
		minRel := cappedPiecewise(vol,minRelease)

		if demand < minRel {
			return minRel
		}

		maxRel := cappedPiecewise(vol,maxRelease)
		if demand > maxRel {
			return maxRel
		}
		return demand
	}

	releaseRatesCloseEnough := func (a, b float64) bool {
		absError := math.Abs(a-b)
		if absError < ALLOWED_ABS_ERROR_RELEASE_RATE {
			return true
		}
		if (a == 0.0) || (b == 0.0) {
			return true
		}
		relError := (absError / a)
		if relError > ALLOWED_REL_ERROR_RELEASE_RATE {
			return false
		}
		return true
	}

	err := checkStorageConfiguration(nLVA,volumes)
	if err != nil {
		fmt.Println(err)
		return
	}

	volume = initialVolume
	n := rainfallTS.Len1()
	nSubtimeSteps := 0

	idx := []int{0}
	for i := 0; i < n; i++ {
		idx[0] = i
		timeRemaining := deltaT
		subtimestep := deltaT

		outflowVolume := 0.0

		inflow := inflowTS.Get(idx)
		demand := demandTS.Get(idx)

		rainfallPerSecond := rainfallTS.Get(idx) / deltaT
		petPerSecond := petTS.Get(idx) / deltaT
		netAtmosphericFluxInPerSecond := (rainfallPerSecond - petPerSecond) * units.MILLIMETRES_TO_METRES

		for timeRemaining > 0 {
			nSubtimeSteps += 1
			subtimestep = math.Min(timeRemaining,subtimestep*2)

			estOutflow := releaseRate(demand,volume)

			area = cappedPiecewise(volume,areas)

			netAtmosphericFluxInRate := netAtmosphericFluxInPerSecond * area

			netFluxInWithoutRelease := inflow + netAtmosphericFluxInRate
			avgOutflow := 0.0
			for {
				testVol := volume + ((inflow-estOutflow)+(netAtmosphericFluxInPerSecond*area)) * subtimestep

				if testVol < 0.0 {
					if subtimestep <= MIN_TIMESTEP_SECONDS {
						minR := cappedPiecewise(volume,minRelease)
						maxR := cappedPiecewise(volume,maxRelease)

						fmt.Printf("==== %d - Hit minimum timestep (%f) and volume out of bounds ====\n",i,subtimestep)
						fmt.Printf("demand = %f\n",demand)
						fmt.Printf("input = %f\n",inflow)
						fmt.Printf("volume = %f\n",volume)
						fmt.Printf("netAtmosphericFluxInRate=%f\n",netAtmosphericFluxInRate)
						fmt.Printf("netFluxInWithoutRelease=%f\n",netFluxInWithoutRelease)
						fmt.Printf("estOutflow=%f\n",estOutflow)
						fmt.Printf("minR=%f\n",minR)
						fmt.Printf("maxR=%f\n",maxR)
						fmt.Printf("testVol=%f\n",testVol)
						fmt.Printf("%d - testVol(%f) outside volume bounds [%f,%f] at minimum subtimestep\n",
								i,testVol,volCurveMin,volCurveMax)

						panic(err)
					}
					subtimestep = math.Max(subtimestep*0.5,MIN_TIMESTEP_SECONDS)
				} else {
					testArea := cappedPiecewise(testVol,areas)
					avgArea := (area+testArea)/2.0
					testVol = volume + ((inflow-estOutflow)+(netAtmosphericFluxInPerSecond*avgArea)) * subtimestep

					estOutflowAfter := releaseRate(demand,testVol)

					avgOutflow = (estOutflowAfter + estOutflow) / 2.0
					if releaseRatesCloseEnough(estOutflow,avgOutflow) {
						break
					}

					if subtimestep <= MIN_TIMESTEP_SECONDS {
						break
					}
				}

				subtimestep = math.Max(subtimestep*0.5,MIN_TIMESTEP_SECONDS)
			}

			outflowVolume += avgOutflow * subtimestep
			volume = volume + (netFluxInWithoutRelease-avgOutflow) * subtimestep
			timeRemaining -= subtimestep
		}

		volumeTS.Set(idx,volume)

		outflowRate := outflowVolume / deltaT
		outflowTS.Set(idx,outflowRate)
	}

	level = cappedPiecewise(volume,levels)
	area = cappedPiecewise(volume,areas)

	return
}


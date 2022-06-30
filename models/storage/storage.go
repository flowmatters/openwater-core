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
	MIN_TIMESTEP_SECONDS_NEGATIVE = 6
	MIN_TIMESTEP_SECONDS_POSITIVE = 60
	ALLOWED_REL_ERROR_RELEASE_RATE = 1e-5
	ALLOWED_ABS_ERROR_RELEASE_RATE = 1e-4
	ESSENTIALLY_ZERO_RELEASE_RATE = 1e-4
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
		rainfallVolume: m^3.s^-1
		evaporationVolume: m^3.s^-1
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
												 volumeTS, outflowTS, rainfallVolume, evaporationVolume data.ND1Float64) (volume, level, area float64) {
	idxCurve0 := []int{0}
	idxCurveN := []int{nLVA-1}

	volCurveMin := volumes.Get(idxCurve0)
	volCurveMax := volumes.Get(idxCurveN)
	maxSpill := minRelease.Get(idxCurveN)

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
		if (math.Abs(a) < ESSENTIALLY_ZERO_RELEASE_RATE) && (math.Abs(b) < ESSENTIALLY_ZERO_RELEASE_RATE) {
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

		targetMinCap := targetMinimumCapacity.Get(idx)
		targetMaxVol := volCurveMax - targetMinCap
		autoAdjustDemand := false//true

		inflow := inflowTS.Get(idx)
		origDemand := demandTS.Get(idx)
		demand := origDemand

		// volumeOverTarget := math.Max(0.0, volume - targetMaxVol)
		// if volumeOverTarget > 0.0 {
		// 	// Adjust once per timestep
		// 	rateToTarget := volumeOverTarget/subtimestep
		// 	x := origDemand + 0.0001 * rateToTarget * rateToTarget
		// 	y := math.Min(x,rateToTarget)
		// 	z := math.Max(y,origDemand)	
		// 	demand = z
		// 	// demand = math.Max(origDemand,math.Log(volumeOverTarget)/subtimestep)
		// }

		// if volume > volCurveMax {
		// 	outflowVolume += volume - volCurveMax
		// 	volume = volCurveMax
		// }

		rainfallVolForTimestep := 0.0
		evaporationVolForTimestep := 0.0

		rainfallPerSecond := rainfallTS.Get(idx) / deltaT
		petPerSecond := petTS.Get(idx) / deltaT
		netAtmosphericFluxDepthPerSecond := (rainfallPerSecond - petPerSecond) * units.MILLIMETRES_TO_METRES

		for timeRemaining > 0 {
			nSubtimeSteps += 1

			subtimestep = math.Min(timeRemaining,subtimestep*2)

			demand = origDemand
			if autoAdjustDemand && (volume > targetMaxVol) {
				demand = math.Max(origDemand,0.0001*(volume-targetMaxVol)/subtimestep)
			}

			// demand = origDemand
			// if (volume) > volCurveMax {
			// 	demand = math.Max(origDemand,(volume-volCurveMax)/(timeRemaining))
			// }

			// demand = origDemand
			// if (volume+inflow*timeRemaining) > volCurveMax {
			// 	demand = math.Max(origDemand,((volume+inflow*subtimestep)-volCurveMax)/(timeRemaining))
			// }

			// if volume > volCurveMax {
			// 	outflowVolume += volume - volCurveMax
			// 	volume = volCurveMax
			// }

			estOutflow := releaseRate(demand,volume)

			area = cappedPiecewise(volume,areas)
			// netAtmosphericFluxInRate := netAtmosphericFluxDepthPerSecond * area

			// netFluxInWithoutRelease := inflow + netAtmosphericFluxInRate
			testVol := 0.0
			avgOutflow := 0.0
			avgArea := 0.0

			for {
				testVol = volume + ((inflow-estOutflow)+(netAtmosphericFluxDepthPerSecond*area)) * subtimestep
				// demand = origDemand
				// if volume > targetMaxVol {
				// 	overAmount := volume - targetMaxVol
				// 	extraReleaseRequired := overAmount / subtimestep
				// 	demand = origDemand + extraReleaseRequired
				// 	// demand = math.Max(origDemand,0.05*(volume-targetMaxVol)/subtimestep)
				// }

				// demand = origDemand
				// if testVol > volCurveMax {
				// 	// drainTime := math.Max(subtimestep,0.25*timeRemaining)
				// 	drainTime := 0.4*timeRemaining
				// 	// drainTime := timeRemaining
				// 	demand = math.Max(origDemand,(testVol-volCurveMax)/drainTime)
				// 	estOutflow = releaseRate(demand,testVol)
				// }

				if testVol < 0.0 {
					if subtimestep <= MIN_TIMESTEP_SECONDS_NEGATIVE {
						// report()
						panic("testVol < 0.0 and subtimestep <= MIN_TIMESTEP_SECONDS")
						// fmt.Println("testVol < 0.0 and subtimestep <= MIN_TIMESTEP_SECONDS_NEGATIVE")
						// return
					}
					subtimestep = math.Max(subtimestep*0.5,MIN_TIMESTEP_SECONDS_NEGATIVE)
				} else {
					// testArea = cappedPiecewise(testVol,areas)
					// avgArea = (area+testArea)/2.0
					avgArea = cappedPiecewise((testVol+volume)/2.0,areas)
					testVol = volume + ((inflow-estOutflow)+(netAtmosphericFluxDepthPerSecond*avgArea)) * subtimestep

					estOutflowAfter := releaseRate(demand,testVol)

					avgOutflow = (estOutflowAfter + estOutflow) / 2.0

					testVol = volume + ((inflow-avgOutflow)+(netAtmosphericFluxDepthPerSecond*avgArea)) * subtimestep
					if testVol >= 0.0 {
						if releaseRatesCloseEnough(estOutflow,avgOutflow) {
							break
						}

						if subtimestep <= MIN_TIMESTEP_SECONDS_POSITIVE {
							break
						}
					} else if subtimestep <= MIN_TIMESTEP_SECONDS_NEGATIVE {
						panic("testVol < 0.0 and subtimestep <= MIN_TIMESTEP_SECONDS_NEGATIVE")
						// fmt.Println("testVol < 0.0 and subtimestep <= MIN_TIMESTEP_SECONDS_NEGATIVE")
						// return
					}
				}
				rainfallVolForTimestep += rainfallPerSecond * avgArea * subtimestep
				evaporationVolForTimestep += petPerSecond * avgArea * subtimestep
				subtimestep = math.Max(subtimestep*0.5,MIN_TIMESTEP_SECONDS_NEGATIVE)
			}

			outflowVolume += avgOutflow * subtimestep
			// testVol = volume + (inflow+(netAtmosphericFluxDepthPerSecond*avgArea)-avgOutflow) * subtimestep
			// testArea := cappedPiecewise(testVol,areas)
			// netAtmosphericFluxInRate = netAtmosphericFluxDepthPerSecond * (area+testArea)/2.0
			// netFluxInWithoutRelease = inflow + netAtmosphericFluxInRate
			volume = volume + (inflow+(netAtmosphericFluxDepthPerSecond*avgArea)-avgOutflow) * subtimestep
			if volume < 0 {
				// report()
				panic(err)
			}

			if volume > volCurveMax {
				overTopRatio := math.Min(volume/volCurveMax,2.0)
				excessOutflow := math.Max((overTopRatio*maxSpill) - avgOutflow,0.0)
				excessOutflowVolume := excessOutflow * subtimestep
				excessOutflowVolume = math.Max(math.Min(excessOutflowVolume,volume-volCurveMax),0.0)
				outflowVolume += excessOutflowVolume
				volume = volume - excessOutflowVolume
			}
	
			timeRemaining -= subtimestep
		}

		volumeTS.Set(idx,volume)

		outflowRate := outflowVolume / deltaT
		outflowTS.Set(idx,outflowRate)

		rainfallVolume.Set(idx,rainfallVolForTimestep/deltaT)
		evaporationVolume.Set(idx,evaporationVolForTimestep/deltaT)
	}

	level = cappedPiecewise(volume,levels)
	area = cappedPiecewise(volume,areas)

	return
}


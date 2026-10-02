package storage

import (
	"errors"
	"math"
	"slices"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/conv/units"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/util/fn"
)

const (
	MIN_TIMESTEP_SECONDS_NEGATIVE  = 6
	MIN_TIMESTEP_SECONDS_POSITIVE  = 60
	ALLOWED_REL_ERROR_RELEASE_RATE = 1e-5
	ALLOWED_ABS_ERROR_RELEASE_RATE = 1e-4
	ESSENTIALLY_ZERO_RELEASE_RATE  = 1e-4
	MAX_SUBTIMESTEPS               = 600000
	// Relative discrepancy in the release rate worth reporting. Well above
	// ALLOWED_REL_ERROR_RELEASE_RATE: the solver routinely stops refining
	// slightly outside its convergence tolerance, and that on its own is not
	// a problem worth telling anyone about.
	REPORTABLE_REL_ERROR_RELEASE_RATE = 1e-2
)

/* OW-SPEC
Storage:
  symbol: WB
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
	reports:
		badConfiguration: '[error] storage is unusable: {0} points in the LVA table, maximum volume {1} m^3'
		subtimestepFloor: '[error] sub-timestep floor of {0} s reached with volume still negative ({1} m^3)'
		releaseNotConverged: '[warning] release rate did not converge: estimated {0} vs averaged {1} m^3.s^-1 at sub-timestep {2} s'
		subtimestepBudget: '[warning] sub-timestep budget exhausted: {0} sub-timesteps used against a budget of {1}'
		notANumber: '[error] computation produced a value that is not a number (volume {0} m^3, inflow {1}, demand {2} m^3.s^-1)'
		volumeClamped: '[error] storage volume went negative ({0} m^3) and was clamped to zero'
	tags:
		storage
*/

func checkStorageConfiguration(nLVA int, volumes []float64) error {
	if nLVA == 0 {
		return errors.New("No points in LVA table or release curves!")
	}

	if slices.Max(volumes) <= 0.0 {
		return errors.New("No volumes")
	}

	return nil
}

// abandonCell marks the rest of a cell's outputs as unknown.
//
// Filling with NaN rather than leaving zeros is deliberate. A zero is
// indistinguishable from a real answer, and a storage that silently produced
// zeros for a whole run is the failure this reporting work set out to make
// visible. A NaN propagates, shows up in the results, and cannot be mistaken
// for a number anyone computed.
func abandonCell(from int, series ...[]float64) {
	nan := math.NaN()
	for _, s := range series {
		for i := from; i < len(s); i++ {
			s[i] = nan
		}
	}
}

// releaseRateDiscrepancyMatters reports whether the gap left when the solver
// stops at the sub-timestep floor is big enough to affect the answer.
func releaseRateDiscrepancyMatters(estimated, averaged float64) bool {
	scale := math.Max(math.Abs(averaged), ESSENTIALLY_ZERO_RELEASE_RATE)
	return math.Abs(estimated-averaged) > REPORTABLE_REL_ERROR_RELEASE_RATE*scale
}

func storageWaterBalance(rainfallTS, petTS, inflowTS, demandTS, targetMinimumVolume, targetMinimumCapacity []float64,
	initialVolume, initialLevel, initialArea float64,
	deltaT float64,
	nLVA int,
	levels, volumes, areas, minRelease, maxRelease []float64,
	volumeTS, outflowTS, rainfallVolume, evaporationVolume []float64,
	rep *sim.Reporter) (volume, level, area float64) {
	err := checkStorageConfiguration(nLVA, volumes)
	if err != nil {
		// Timestep -1: a configuration fault has no timestep.
		maxVolume := 0.0
		if len(volumes) > 0 {
			maxVolume = slices.Max(volumes)
		}
		rep.Err2(StorageReportBadConfiguration, -1, float64(nLVA), maxVolume)
		log.Error().Stack().Err(err).Msg("")
		abandonCell(0, volumeTS, outflowTS, rainfallVolume, evaporationVolume)
		return math.NaN(), math.NaN(), math.NaN()
	}

	idxCurve0 := 0
	idxCurveN := nLVA - 1

	volCurveMin := volumes[idxCurve0]
	volCurveMax := volumes[idxCurveN]
	maxSpill := minRelease[idxCurveN]

	// Water balance functions
	cappedPiecewise := func(vol float64, ys []float64) float64 {
		if vol < volCurveMin {
			return ys[idxCurve0]
		}
		if vol > volCurveMax {
			return ys[idxCurveN]
		}
		res, err := fn.Piecewise(vol, volumes, ys)
		if err != nil {
			// A malformed curve (non-monotonic volumes, say). Returning NaN
			// hands this to the NaN checks below, which report it against the
			// cell and abandon it, rather than taking down the whole run.
			return math.NaN()
		}
		return res
	}

	releaseRate := func(demand, vol float64) float64 {
		minRel := cappedPiecewise(vol, minRelease)

		if demand < minRel {
			return minRel
		}

		maxRel := cappedPiecewise(vol, maxRelease)
		if demand > maxRel {
			return maxRel
		}
		return demand
	}

	releaseRatesCloseEnough := func(a, b float64) bool {
		absError := math.Abs(a - b)
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

	volume = initialVolume
	n := len(rainfallTS)
	nSubtimeSteps := 0
	budgetReported := false

	for i := 0; i < n; i++ {
		timeRemaining := deltaT
		subtimestep := deltaT

		outflowVolume := 0.0
		releaseReported := false

		targetMinCap := targetMinimumCapacity[i]
		targetMaxVol := volCurveMax - targetMinCap
		autoAdjustDemand := true

		inflow := inflowTS[i]
		origDemand := demandTS[i]
		demand := origDemand

		if math.IsNaN(demand) || math.IsNaN(inflow) {
			rep.Err3(StorageReportNotANumber, int32(i), volume, inflow, demand)
			abandonCell(i, volumeTS, outflowTS, rainfallVolume, evaporationVolume)
			return math.NaN(), math.NaN(), math.NaN()
		}
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

		rainfallPerSecond := rainfallTS[i] / deltaT
		petPerSecond := petTS[i] / deltaT
		netAtmosphericFluxDepthPerSecond := (rainfallPerSecond - petPerSecond) * units.MILLIMETRES_TO_METRES

		for timeRemaining > 0 {
			nSubtimeSteps++

			subtimestep = math.Min(timeRemaining, subtimestep*2)

			demand = origDemand
			if autoAdjustDemand && (volume > targetMaxVol) {
				demand = math.Max(origDemand, (volume-targetMaxVol)/timeRemaining)
			}
			// demand = origDemand
			// if (volume) > volCurveMax {
			// 	demand = math.Max(origDemand,(volume-volCurveMax)/(timeRemaining))
			// }
			if math.IsNaN(demand) {
				rep.Err3(StorageReportNotANumber, int32(i), volume, inflow, demand)
				abandonCell(i, volumeTS, outflowTS, rainfallVolume, evaporationVolume)
				return math.NaN(), math.NaN(), math.NaN()
			}

			// demand = origDemand
			// if (volume+inflow*timeRemaining) > volCurveMax {
			// 	demand = math.Max(origDemand,((volume+inflow*subtimestep)-volCurveMax)/(timeRemaining))
			// }

			// if volume > volCurveMax {
			// 	outflowVolume += volume - volCurveMax
			// 	volume = volCurveMax
			// }

			estOutflow := releaseRate(demand, volume)
			if math.IsNaN(estOutflow) {
				rep.Err3(StorageReportNotANumber, int32(i), volume, inflow, demand)
				abandonCell(i, volumeTS, outflowTS, rainfallVolume, evaporationVolume)
				return math.NaN(), math.NaN(), math.NaN()
			}
			area = cappedPiecewise(volume, areas)
			// netAtmosphericFluxInRate := netAtmosphericFluxDepthPerSecond * area

			// netFluxInWithoutRelease := inflow + netAtmosphericFluxInRate
			testVol := 0.0
			avgOutflow := 0.0
			avgArea := 0.0

			for {
				testVol = volume + ((inflow-estOutflow)+(netAtmosphericFluxDepthPerSecond*area))*subtimestep
				// Check for NaNs
				if math.IsNaN(testVol) {
					rep.Err3(StorageReportNotANumber, int32(i), volume, inflow, demand)
					abandonCell(i, volumeTS, outflowTS, rainfallVolume, evaporationVolume)
					return math.NaN(), math.NaN(), math.NaN()
				}

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
						// The sub-timestep cannot shrink further and the
						// storage still wants to go negative. Empty it and
						// carry on: the volume is bounded and the report says
						// the answer is not to be trusted.
						rep.Err2(StorageReportSubtimestepFloor, int32(i), subtimestep, testVol)
						testVol = 0.0
						break
					}
					subtimestep = math.Max(subtimestep*0.5, MIN_TIMESTEP_SECONDS_NEGATIVE)
				} else {
					// testArea = cappedPiecewise(testVol,areas)
					// avgArea = (area+testArea)/2.0
					avgArea = cappedPiecewise((testVol+volume)/2.0, areas)
					testVol = volume + ((inflow-estOutflow)+(netAtmosphericFluxDepthPerSecond*avgArea))*subtimestep

					estOutflowAfter := releaseRate(demand, testVol)

					avgOutflow = (estOutflowAfter + estOutflow) / 2.0

					testVol = volume + ((inflow-avgOutflow)+(netAtmosphericFluxDepthPerSecond*avgArea))*subtimestep
					if testVol >= 0.0 {
						if releaseRatesCloseEnough(estOutflow, avgOutflow) {
							break
						}

						if subtimestep <= MIN_TIMESTEP_SECONDS_POSITIVE {
							// Once per timestep at most: this loop runs many
							// times per timestep, and "the solver stopped here"
							// is a property of the day, not of each attempt.
							if !releaseReported && releaseRateDiscrepancyMatters(estOutflow, avgOutflow) {
								rep.Warn3(StorageReportReleaseNotConverged, int32(i),
									estOutflow, avgOutflow, subtimestep)
								releaseReported = true
							}
							break
						}
					} else if subtimestep <= MIN_TIMESTEP_SECONDS_NEGATIVE {
						rep.Err2(StorageReportSubtimestepFloor, int32(i), subtimestep, testVol)
						break
					}
				}
				subtimestep = math.Max(subtimestep*0.5, MIN_TIMESTEP_SECONDS_NEGATIVE)
			}

			outflowVolume += avgOutflow * subtimestep
			rainfallVolForTimestep += rainfallPerSecond * units.MILLIMETRES_TO_METRES * avgArea * subtimestep
			evaporationVolForTimestep += petPerSecond * units.MILLIMETRES_TO_METRES * avgArea * subtimestep
			// testVol = volume + (inflow+(netAtmosphericFluxDepthPerSecond*avgArea)-avgOutflow) * subtimestep
			// testArea := cappedPiecewise(testVol,areas)
			// netAtmosphericFluxInRate = netAtmosphericFluxDepthPerSecond * (area+testArea)/2.0
			// netFluxInWithoutRelease = inflow + netAtmosphericFluxInRate
			volume = volume + (inflow+(netAtmosphericFluxDepthPerSecond*avgArea)-avgOutflow)*subtimestep
			if volume < 0 {
				// Bounded failure: an empty storage is wrong but usable, and
				// the report carries the volume that was discarded.
				rep.Err1(StorageReportVolumeClamped, int32(i), volume)
				volume = 0.0
			}

			if volume > volCurveMax {
				overTopRatio := volume / volCurveMax
				excessOutflow := math.Max((overTopRatio*maxSpill)-avgOutflow, 0.0)
				excessOutflowVolume := excessOutflow * subtimestep
				excessOutflowVolume = math.Max(math.Min(excessOutflowVolume, volume-volCurveMax), 0.0)
				outflowVolume += excessOutflowVolume
				volume = volume - excessOutflowVolume
			}

			timeRemaining -= subtimestep
		}

		if nSubtimeSteps > MAX_SUBTIMESTEPS && !budgetReported {
			rep.Warn2(StorageReportSubtimestepBudget, int32(i), float64(nSubtimeSteps), MAX_SUBTIMESTEPS)
			budgetReported = true
		}

		volumeTS[i] = volume
		outflowRate := outflowVolume / deltaT
		outflowTS[i] = outflowRate
		rainfallVolume[i] = rainfallVolForTimestep / deltaT
		evaporationVolume[i] = evaporationVolForTimestep / deltaT
	}

	level = cappedPiecewise(volume, levels)
	area = cappedPiecewise(volume, areas)

	return
}

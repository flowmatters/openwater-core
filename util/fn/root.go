package fn

import (
	"math"
)

func FindRoot(fn func(x float64) float64,
fn_dx func(x float64) float64,
initialX, minX, maxX, tolerance, convergenceLimit float64, maxIterations int) (x, delta float64) {
	delta = fn(initialX)
	x = initialX
	maxDelta := fn(maxX)
	minDelta := fn(minX)
	if minDelta > 0 || maxDelta < 0 {
		panic("Invalid range")
	}
	for iteration := 0; iteration < maxIterations; iteration++ {
		var trialXs []float64
		var trialDeltas []float64

		halvingX := maxX - (maxX-minX)*0.5
		bisectionX := maxX - (maxX-minX)*maxDelta/(maxDelta-minDelta)

		trialXs = append(trialXs, halvingX, bisectionX)

		if fn_dx != nil {
			deriv := fn_dx(x)
			if deriv != 0.0 {
				newtonRaphsonX := x - delta/deriv

				if newtonRaphsonX > minX && newtonRaphsonX < maxX {
					trialXs = append(trialXs, newtonRaphsonX)
				}
			}
		}

		//initialise previous bounds
		var minTrialX = minX
		var minTrialDelta = minDelta
		var maxTrialX = maxX
		var maxTrialDelta = maxDelta
		hitConvergenceLimit := 0
		for _, trial := range trialXs {
			if math.Abs(x-trial) < convergenceLimit {
				hitConvergenceLimit++
			}

			trialDelta := fn(trial)
			if math.Abs(trialDelta) < tolerance {
				x = trial
				delta = trialDelta
				return
			}

			if trialDelta < 0.0 {
				if trial > minTrialX && trial <= maxTrialX {
					minTrialX = trial
					minTrialDelta = trialDelta
				}
			} else {
				if trial < maxTrialX && trial >= minTrialX {
					maxTrialX = trial
					maxTrialDelta = trialDelta
				}
			}

			trialDeltas = append(trialDeltas, trialDelta)
		}

		maxX = maxTrialX
		maxDelta = maxTrialDelta

		minX = minTrialX
		minDelta = minTrialDelta

		// if minTrialX == minX {
		// 	// minTrialX hasn't been updated

		// 	x = maxTrialX
		// 	delta = maxTrialDelta

		// 	// maxX = maxTrialX
		// 	// maxDelta = maxTrialDelta

		// 	// minX = minTrialX
		// 	// minDelta = minTrialDelta

		// } else if maxTrialX == maxX {
		// 	// maxTrialX hasn't been updated

		// 	x = minTrialX
		// 	delta = minTrialDelta

		// 	// minX = minTrialX
		// 	// minDelta = minTrialDelta

		// 	// maxX = maxTrialX
		// 	// maxDelta = maxTrialDelta
		// } else {

		// }

		if math.Abs(minTrialDelta) <= maxTrialDelta {
			x = minTrialX
			delta = minTrialDelta

			// maxX = maxTrialX
			// maxDelta = maxTrialDelta
		} else {
			x = maxTrialX
			delta = maxTrialDelta

			// minX = minTrialX
			// minDelta = minTrialDelta
		}

		if hitConvergenceLimit == len(trialXs) {
			return
		}
	}

	return
}


package storage

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testDeltaT = 86400.0

type storageRun struct {
	volume, outflow, rainfallVolume, evaporationVolume []float64
}

// Flat-sided storage: constant surface area at all volumes.
func runStorage(t *testing.T, rainfall, pet, inflow []float64, initialVolume float64, minRelease, maxRelease []float64) storageRun {
	n := len(rainfall)
	zeros := make([]float64, n)
	levels := []float64{0, 10}
	volumes := []float64{0, 1e7}
	areas := []float64{1e6, 1e6}

	r := storageRun{make([]float64, n), make([]float64, n), make([]float64, n), make([]float64, n)}
	storageWaterBalance(rainfall, pet, inflow, zeros, zeros, zeros,
		initialVolume, 0, 0, testDeltaT, len(levels),
		levels, volumes, areas, minRelease, maxRelease,
		r.volume, r.outflow, r.rainfallVolume, r.evaporationVolume)
	return r
}

func filled(n int, v float64) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = v
	}
	return res
}

func TestRainfallAndEvaporationVolumesAreInCubicMetresPerSecond(t *testing.T) {
	n := 5
	// No release, so the storage only changes through rainfall and evaporation
	r := runStorage(t, filled(n, 10.0), filled(n, 4.0), filled(n, 0.0), 5e6, []float64{0, 0}, []float64{0, 0})

	// 10mm and 4mm over 1e6 m^2, per day
	expectedRain := 0.010 * 1e6 / testDeltaT
	expectedEvap := 0.004 * 1e6 / testDeltaT
	for i := 0; i < n; i++ {
		assert.InDelta(t, expectedRain, r.rainfallVolume[i], 1e-9*expectedRain, "rainfallVolume[%d]", i)
		assert.InDelta(t, expectedEvap, r.evaporationVolume[i], 1e-9*expectedEvap, "evaporationVolume[%d]", i)
	}
}

func TestWaterBalanceClosesIncludingAtmosphericFluxes(t *testing.T) {
	n := 30
	rainfall := make([]float64, n)
	pet := filled(n, 5.0)
	inflow := make([]float64, n)
	for i := 0; i < n; i++ {
		if i%7 == 0 {
			rainfall[i] = 60.0
			inflow[i] = 200.0
		}
	}
	// Release depends on volume, which forces sub-timestepping
	minRelease := []float64{0, 150}
	maxRelease := []float64{0, 150}
	initialVolume := 2e6
	r := runStorage(t, rainfall, pet, inflow, initialVolume, minRelease, maxRelease)

	prev := initialVolume
	for i := 0; i < n; i++ {
		netIn := (inflow[i] - r.outflow[i] + r.rainfallVolume[i] - r.evaporationVolume[i]) * testDeltaT
		change := r.volume[i] - prev
		tolerance := math.Max(1e-6*math.Abs(r.volume[i]), 1e-3)
		assert.InDelta(t, change, netIn, tolerance, "step %d: volume change %f, net inflow %f", i, change, netIn)
		prev = r.volume[i]
	}
}

package storage

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flowmatters/openwater-core/sim"
)

// A workable storage: a three-point level/volume/area curve with a fixed
// release range.
func workingCurve() (levels, volumes, areas, minRelease, maxRelease []float64) {
	levels = []float64{0, 5, 10}
	volumes = []float64{0, 5e5, 2e6}
	areas = []float64{1e4, 5e4, 1e5}
	minRelease = []float64{0, 0, 50}
	maxRelease = []float64{0, 10, 100}
	return
}

// storageReportCount is the size of Storage's report catalogue, read from the
// model itself so adding a report doesn't leave these tests sizing sinks too
// small and silently dropping records.
func storageReportCount() int {
	return len(sim.Catalog["Storage"]().(sim.Reporting).ReportCatalog())
}

func constant(n int, v float64) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = v
	}
	return s
}

func runStorage(n int, nLVA int, levels, volumes, areas, minRelease, maxRelease []float64,
	inflow, demand float64) (sim.ReportSet, []float64) {
	volumeTS := make([]float64, n)
	outflowTS := make([]float64, n)
	rainVol := make([]float64, n)
	evapVol := make([]float64, n)

	sink := sim.NewReportSink(1, storageReportCount(), sim.SevWarning, 100)
	storageWaterBalance(
		constant(n, 0), constant(n, 2), constant(n, inflow), constant(n, demand),
		constant(n, 0), constant(n, 0),
		1e5, 0, 0,
		86400, nLVA,
		levels, volumes, areas, minRelease, maxRelease,
		volumeTS, outflowTS, rainVol, evapVol,
		sink.Cell(0))

	return sink.Flush(), volumeTS
}

// The headline case. An unusable LVA table used to produce one context-free log
// line for the entire model and zeros for the whole run -- a silently wrong
// answer. It is now reported against the cell, and the outputs are NaN.
func TestBadConfigurationIsReported(t *testing.T) {
	levels, _, areas, minRelease, maxRelease := workingCurve()
	zeroVolumes := []float64{0, 0, 0}

	rs, volumes := runStorage(10, 3, levels, zeroVolumes, areas, minRelease, maxRelease, 5, 1)

	require.Equal(t, 1, rs.Len(), "an unusable configuration must be reported once")
	assert.EqualValues(t, StorageReportBadConfiguration, rs.Codes[0])
	assert.EqualValues(t, sim.SevError, rs.Severities[0])
	assert.EqualValues(t, -1, rs.Timesteps[0],
		"a configuration fault has no timestep, and must not claim one")
	assert.Equal(t, 3.0, rs.Args[0], "reports the number of LVA points")
	assert.Equal(t, 0.0, rs.Args[1], "reports the maximum volume")

	// See TestBadConfigurationYieldsNaNNotZeros in degradation_test.go: the
	// outputs must not look like a real answer.
	for i, v := range volumes {
		assert.True(t, math.IsNaN(v), "timestep %d should be NaN", i)
	}
}

func TestEmptyCurveIsReported(t *testing.T) {
	rs, _ := runStorage(5, 0, nil, nil, nil, nil, nil, 5, 1)
	require.Equal(t, 1, rs.Len())
	assert.EqualValues(t, StorageReportBadConfiguration, rs.Codes[0])
	assert.Equal(t, 0.0, rs.Args[0])
}

func TestWorkingStorageIsSilent(t *testing.T) {
	levels, volumes, areas, minRelease, maxRelease := workingCurve()
	rs, volumeTS := runStorage(50, 3, levels, volumes, areas, minRelease, maxRelease, 5, 1)

	assert.Zero(t, rs.Len(), "an ordinary storage run must report nothing: %v", rs.Codes)
	for i, v := range volumeTS {
		require.False(t, math.IsNaN(v), "NaN volume at step %d", i)
	}
}

// The release-rate report must stay reachable. Breaking out of the solver at
// the sub-timestep floor is common and harmless, so it is filtered by a
// materiality threshold -- these cases check the filter did not turn the report
// into dead code.
func TestReleaseNotConvergedIsReported(t *testing.T) {
	t.Run("volume crossing a release discontinuity", func(t *testing.T) {
		// The maximum release jumps from 0 to 5000 m^3.s^-1 over a 1% change in
		// volume, so the solver cannot pin the rate down on the day it crosses.
		rs, _ := runStorage(50, 3,
			[]float64{0, 5, 10}, []float64{0, 1e5, 1.01e5}, []float64{1e4, 5e4, 1e5},
			[]float64{0, 0, 0}, []float64{0, 0, 5000}, 50, 1)

		require.Equal(t, 1, rs.Len(), "should report on the one day it crosses")
		assert.EqualValues(t, StorageReportReleaseNotConverged, rs.Codes[0])
		assert.EqualValues(t, sim.SevWarning, rs.Severities[0])
		assert.GreaterOrEqual(t, rs.Timesteps[0], int32(0), "a timestep event must carry its timestep")
	})

	t.Run("storage sitting on a discontinuity", func(t *testing.T) {
		// A minimum release that jumps to 2000 m^3.s^-1 keeps the storage
		// pinned to the cliff, so every day is affected.
		rs, _ := runStorage(50, 3,
			[]float64{0, 5, 10}, []float64{0, 2e5, 2.05e5}, []float64{1e4, 5e4, 1e5},
			[]float64{0, 0, 2000}, []float64{0, 0, 5000}, 30, 1)

		counts := rs.Counts[StorageReportReleaseNotConverged]
		assert.EqualValues(t, 50, counts, "counts stay complete for a condition that fires daily")
		assert.EqualValues(t, 50, rs.Len(), "and the detail cap here is high enough to keep them all")
	})
}

// At most one report per timestep, however many sub-timesteps a day needs.
func TestReleaseReportIsOncePerTimestep(t *testing.T) {
	const n = 50
	rs, _ := runStorage(n, 3,
		[]float64{0, 5, 10}, []float64{0, 2e5, 2.05e5}, []float64{1e4, 5e4, 1e5},
		[]float64{0, 0, 2000}, []float64{0, 0, 5000}, 30, 1)

	seen := map[int32]bool{}
	for i := 0; i < rs.Len(); i++ {
		if rs.Codes[i] != StorageReportReleaseNotConverged {
			continue
		}
		assert.False(t, seen[rs.Timesteps[i]], "timestep %d reported more than once", rs.Timesteps[i])
		seen[rs.Timesteps[i]] = true
	}
	assert.LessOrEqual(t, len(seen), n)
}

func TestStorageRunsWithoutAReporter(t *testing.T) {
	levels, volumes, areas, minRelease, maxRelease := workingCurve()
	n := 20
	volumeTS := make([]float64, n)
	outflowTS := make([]float64, n)
	rainVol := make([]float64, n)
	evapVol := make([]float64, n)

	assert.NotPanics(t, func() {
		storageWaterBalance(
			constant(n, 0), constant(n, 2), constant(n, 5), constant(n, 1),
			constant(n, 0), constant(n, 0),
			1e5, 0, 0,
			86400, 3,
			levels, volumes, areas, minRelease, maxRelease,
			volumeTS, outflowTS, rainVol, evapVol,
			nil)
	}, "reporting off is a nil Reporter, and must not change how the model runs")

	assert.NotZero(t, volumeTS[n-1])
}

// The two runs must agree exactly: instrumentation reports, it does not alter
// the water balance.
func TestReportingDoesNotChangeResults(t *testing.T) {
	levels, volumes, areas, minRelease, maxRelease := workingCurve()
	n := 100

	run := func(rep *sim.Reporter) []float64 {
		volumeTS := make([]float64, n)
		outflowTS := make([]float64, n)
		rainVol := make([]float64, n)
		evapVol := make([]float64, n)
		storageWaterBalance(
			constant(n, 3), constant(n, 2), constant(n, 8), constant(n, 2),
			constant(n, 0), constant(n, 0),
			1e5, 0, 0,
			86400, 3,
			levels, volumes, areas, minRelease, maxRelease,
			volumeTS, outflowTS, rainVol, evapVol,
			rep)
		return append(volumeTS, outflowTS...)
	}

	withReporting := run(sim.NewReportSink(1, storageReportCount(), sim.SevWarning, 100).Cell(0))
	without := run(nil)
	assert.Equal(t, without, withReporting)
}

func TestStorageReportCatalog(t *testing.T) {
	model := sim.Catalog["Storage"]()
	reporting, ok := model.(sim.Reporting)
	require.True(t, ok, "Storage declares a reports: block, so it must implement sim.Reporting")

	catalog := reporting.ReportCatalog()
	require.Len(t, catalog, storageReportCount())

	// Every declared code must line up with its catalogue entry, since the code
	// is what identifies a report in the results file.
	for code, expected := range map[int32]struct {
		name     string
		severity sim.Severity
		nArgs    int
	}{
		StorageReportBadConfiguration:    {"badConfiguration", sim.SevError, 2},
		StorageReportSubtimestepFloor:    {"subtimestepFloor", sim.SevError, 2},
		StorageReportReleaseNotConverged: {"releaseNotConverged", sim.SevWarning, 3},
		StorageReportSubtimestepBudget:   {"subtimestepBudget", sim.SevWarning, 2},
		StorageReportNotANumber:          {"notANumber", sim.SevError, 3},
		StorageReportVolumeClamped:       {"volumeClamped", sim.SevError, 1},
	} {
		entry := catalog[code]
		assert.Equal(t, expected.name, entry.Name)
		assert.Equal(t, expected.severity, entry.Severity, "%s severity", entry.Name)
		assert.Equal(t, expected.nArgs, entry.NArgs, "%s argument count", entry.Name)
		assert.NotEmpty(t, entry.Template)
	}
}

func BenchmarkStorageReportingOff(b *testing.B) {
	levels, volumes, areas, minRelease, maxRelease := workingCurve()
	n := 13514
	volumeTS := make([]float64, n)
	outflowTS := make([]float64, n)
	rainVol := make([]float64, n)
	evapVol := make([]float64, n)
	rain, pet := constant(n, 3), constant(n, 2)
	inflow, demand := constant(n, 8), constant(n, 2)
	zero := constant(n, 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storageWaterBalance(rain, pet, inflow, demand, zero, zero, 1e5, 0, 0,
			86400, 3, levels, volumes, areas, minRelease, maxRelease,
			volumeTS, outflowTS, rainVol, evapVol, nil)
	}
}

func BenchmarkStorageReportingOn(b *testing.B) {
	levels, volumes, areas, minRelease, maxRelease := workingCurve()
	n := 13514
	volumeTS := make([]float64, n)
	outflowTS := make([]float64, n)
	rainVol := make([]float64, n)
	evapVol := make([]float64, n)
	rain, pet := constant(n, 3), constant(n, 2)
	inflow, demand := constant(n, 8), constant(n, 2)
	zero := constant(n, 0)
	rep := sim.NewReportSink(1, storageReportCount(), sim.SevWarning, 3).Cell(0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storageWaterBalance(rain, pet, inflow, demand, zero, zero, 1e5, 0, 0,
			86400, 3, levels, volumes, areas, minRelease, maxRelease,
			volumeTS, outflowTS, rainVol, evapVol, rep)
	}
}

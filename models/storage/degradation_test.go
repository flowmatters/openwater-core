package storage

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/sim"
)

// Phase 6: a cell that cannot be computed reports the fault and yields NaN,
// instead of panicking and taking the whole simulation with it.

func runWithNaNInput(t *testing.T, nanAt int, series string) (sim.ReportSet, []float64) {
	t.Helper()
	levels, volumes, areas, minRelease, maxRelease := workingCurve()

	const n = 20
	inflow := constant(n, 8)
	demand := constant(n, 2)
	switch series {
	case "inflow":
		inflow[nanAt] = math.NaN()
	case "demand":
		demand[nanAt] = math.NaN()
	default:
		t.Fatalf("unknown series %s", series)
	}

	volumeTS := make([]float64, n)
	outflowTS := make([]float64, n)
	rainVol := make([]float64, n)
	evapVol := make([]float64, n)

	sink := sim.NewReportSink(1, storageReportCount(), sim.SevWarning, 100)

	require.NotPanics(t, func() {
		storageWaterBalance(constant(n, 3), constant(n, 2), inflow, demand,
			constant(n, 0), constant(n, 0), 1e5, 0, 0, 86400, 3,
			levels, volumes, areas, minRelease, maxRelease,
			volumeTS, outflowTS, rainVol, evapVol, sink.Cell(0))
	}, "a NaN input must not bring down the run")

	return sink.Flush(), volumeTS
}

func TestNaNInputIsReportedNotFatal(t *testing.T) {
	for _, series := range []string{"inflow", "demand"} {
		t.Run(series, func(t *testing.T) {
			rs, volumeTS := runWithNaNInput(t, 5, series)

			require.NotZero(t, rs.Len())
			assert.EqualValues(t, StorageReportNotANumber, rs.Codes[0])
			assert.EqualValues(t, sim.SevError, rs.Severities[0])
			assert.EqualValues(t, 5, rs.Timesteps[0], "reported against the timestep that broke")

			// Everything before the fault is real; everything after is unknown.
			for i := 0; i < 5; i++ {
				assert.False(t, math.IsNaN(volumeTS[i]), "timestep %d should still hold a real value", i)
			}
			for i := 5; i < len(volumeTS); i++ {
				assert.True(t, math.IsNaN(volumeTS[i]), "timestep %d should be NaN, not a plausible number", i)
			}
		})
	}
}

// An unusable configuration used to yield zeros for the whole run, which is
// indistinguishable from a real answer. It now yields NaN.
func TestBadConfigurationYieldsNaNNotZeros(t *testing.T) {
	levels, _, areas, minRelease, maxRelease := workingCurve()
	rs, volumeTS := runStorage(10, 3, levels, []float64{0, 0, 0}, areas, minRelease, maxRelease, 5, 1)

	require.Equal(t, 1, rs.Len())
	assert.EqualValues(t, StorageReportBadConfiguration, rs.Codes[0])
	for i, v := range volumeTS {
		assert.True(t, math.IsNaN(v), "timestep %d should be NaN, not zero", i)
	}
}

func TestDegradationWorksWithoutAReporter(t *testing.T) {
	levels, volumes, areas, minRelease, maxRelease := workingCurve()
	const n = 10
	inflow := constant(n, 8)
	inflow[3] = math.NaN()

	volumeTS := make([]float64, n)
	outflowTS := make([]float64, n)
	rainVol := make([]float64, n)
	evapVol := make([]float64, n)

	assert.NotPanics(t, func() {
		storageWaterBalance(constant(n, 3), constant(n, 2), inflow, constant(n, 2),
			constant(n, 0), constant(n, 0), 1e5, 0, 0, 86400, 3,
			levels, volumes, areas, minRelease, maxRelease,
			volumeTS, outflowTS, rainVol, evapVol, nil)
	}, "degradation must not depend on reporting being switched on")

	assert.True(t, math.IsNaN(volumeTS[n-1]))
}

// The whole point: one broken cell must not affect its neighbours, and the
// generated wrapper runs them concurrently.
func TestOneBadCellDoesNotAffectTheOthers(t *testing.T) {
	const (
		cells = 3
		n     = 20
	)

	model := sim.Catalog["Storage"]()
	model.InitialiseDimensions([]int{3})

	// DeltaT, nLVA, then levels/volumes/areas/minRelease/maxRelease, 3 each.
	levels, volumes, areas, minRelease, maxRelease := workingCurve()
	params := data.NewArray2D[float64](2+5*3, cells)
	for c := 0; c < cells; c++ {
		params.Set2(0, c, 86400)
		params.Set2(1, c, 3)
		for i := 0; i < 3; i++ {
			params.Set2(2+i, c, levels[i])
			params.Set2(5+i, c, volumes[i])
			params.Set2(8+i, c, areas[i])
			params.Set2(11+i, c, minRelease[i])
			params.Set2(14+i, c, maxRelease[i])
		}
	}
	model.ApplyParameters(params)

	// Inputs: rainfall, pet, inflow, demand, targetMinimumVolume, targetMinimumCapacity
	inputs := data.NewArray3D[float64](cells, 6, n)
	for c := 0; c < cells; c++ {
		for ts := 0; ts < n; ts++ {
			inputs.Set3(c, 0, ts, 3)
			inputs.Set3(c, 1, ts, 2)
			inputs.Set3(c, 2, ts, 8)
			inputs.Set3(c, 3, ts, 2)
		}
	}
	// Only cell 1 is broken.
	inputs.Set3(1, 2, 7, math.NaN())

	states := model.InitialiseStates(cells)
	for c := 0; c < cells; c++ {
		states.Set2(c, 0, 1e5)
	}
	outputs := sim.InitialiseOutputs(model, n, cells)

	sink := sim.NewReportSink(cells, storageReportCount(), sim.SevWarning, 10)
	require.True(t, sim.AttachReportSink(model, sink))

	require.NotPanics(t, func() { model.Run(inputs, states, outputs) },
		"one bad cell must not take down the generation")

	rs := sink.Flush()
	counts := func(cell, code int) uint32 { return rs.Counts[cell*rs.NCodes+code] }
	notANumber := int(StorageReportNotANumber)

	assert.Zero(t, counts(0, notANumber), "cell 0 is fine")
	assert.NotZero(t, counts(1, notANumber), "cell 1 is the broken one")
	assert.Zero(t, counts(2, notANumber), "cell 2 is fine")

	// Cells 0 and 2 produced real numbers throughout; cell 1 went NaN at the
	// fault and stayed there.
	for _, cell := range []int{0, 2} {
		for ts := 0; ts < n; ts++ {
			assert.False(t, math.IsNaN(outputs.Get3(cell, 0, ts)),
				"cell %d timestep %d should be unaffected", cell, ts)
		}
	}
	assert.False(t, math.IsNaN(outputs.Get3(1, 0, 6)), "cell 1 before the fault")
	assert.True(t, math.IsNaN(outputs.Get3(1, 0, 7)), "cell 1 from the fault on")
	assert.True(t, math.IsNaN(outputs.Get3(1, 0, n-1)))
}

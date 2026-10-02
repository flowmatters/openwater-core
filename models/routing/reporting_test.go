package routing

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/sim"
)

func TestStorageRoutingReportCatalog(t *testing.T) {
	model := sim.Catalog["StorageRouting"]()
	reporting, ok := model.(sim.Reporting)
	require.True(t, ok, "StorageRouting declares a reports: block, so it must implement sim.Reporting")

	catalog := reporting.ReportCatalog()
	require.Len(t, catalog, 3)

	for code, expected := range map[int32]struct {
		name     string
		severity sim.Severity
		nArgs    int
	}{
		StorageRoutingReportMassBalanceNotConverged: {"massBalanceNotConverged", sim.SevWarning, 2},
		StorageRoutingReportNanInput:                {"nanInput", sim.SevError, 2},
		StorageRoutingReportNanOutflow:              {"nanOutflow", sim.SevError, 2},
	} {
		entry := catalog[code]
		assert.Equal(t, expected.name, entry.Name)
		assert.Equal(t, expected.severity, entry.Severity, "%s severity", entry.Name)
		assert.Equal(t, expected.nArgs, entry.NArgs, "%s argument count", entry.Name)
		assert.NotEmpty(t, entry.Template)
	}
}

// The report that storage_routing.go had an empty if-block waiting for. A
// strongly non-linear store defeats the 20-iteration root finder.
func TestMassBalanceNotConvergedIsReported(t *testing.T) {
	const n = 200
	out := make([]float64, n)
	store := make([]float64, n)
	sink := sim.NewReportSink(1, 3, sim.SevWarning, 5)

	stiffCase(n).run(out, store, sink.Cell(0))
	rs := sink.Flush()

	counts := rs.Counts[StorageRoutingReportMassBalanceNotConverged]
	assert.Greater(t, counts, uint32(n/2), "the stiff case should fail to converge on most timesteps")

	// Counts stay complete while the detail records are capped at five.
	assert.Equal(t, 5, rs.Len(), "detail records are capped")
	assert.EqualValues(t, int64(counts)-5, rs.Suppressed)

	// The kept records are the onset of the condition, and carry the mass
	// balance error and the limit it exceeded.
	for i := 0; i < rs.Len(); i++ {
		assert.EqualValues(t, sim.SevWarning, rs.Severities[i])
		assert.Equal(t, massBalanceLimit, rs.Args[i*sim.MaxReportArgs+1])
		assert.Greater(t, absFloat(rs.Args[i*sim.MaxReportArgs]), massBalanceLimit)
	}
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func TestWellBehavedRoutingIsSilent(t *testing.T) {
	const n = 200
	out := make([]float64, n)
	store := make([]float64, n)
	sink := sim.NewReportSink(1, 3, sim.SevWarning, 5)

	wellBehavedCase(n).run(out, store, sink.Cell(0))
	assert.Zero(t, sink.Flush().Len(), "a linear store converges, and should report nothing")
}

func TestRoutingRunsWithoutAReporter(t *testing.T) {
	const n = 200
	withReports := make([]float64, n)
	storeA := make([]float64, n)
	without := make([]float64, n)
	storeB := make([]float64, n)

	c := stiffCase(n)
	c.run(withReports, storeA, sim.NewReportSink(1, 3, sim.SevWarning, 5).Cell(0))
	assert.NotPanics(t, func() { c.run(without, storeB, nil) })

	// Instrumentation reports; it does not change the routing.
	assert.Equal(t, without, withReports)
	assert.Equal(t, storeB, storeA)
}

// Reports must be attributed to the right cell when the generated wrapper runs
// a whole generation concurrently. Cell 0 is linear and silent; cell 1 is stiff
// and reports on most timesteps.
func TestGenerationAttributesReportsPerCell(t *testing.T) {
	const (
		cells = 2
		n     = 200
	)

	model := sim.Catalog["StorageRouting"]()

	// Parameters, in spec order: InflowBias, RoutingConstant, RoutingPower,
	// area, deadStorage, DeltaT.
	linear, stiff := wellBehavedCase(n), stiffCase(n)
	params := data.NewArray2D[float64](6, cells)
	for cell, c := range []routingCase{linear, stiff} {
		for row, v := range []float64{c.bias, c.k, c.x, c.area, c.deadStorage, c.deltaT} {
			params.Set2(row, cell, v)
		}
	}
	model.ApplyParameters(params)

	// Inputs, in spec order: inflow, lateral, rainfall, evap.
	inputs := data.NewArray3D[float64](cells, 4, n)
	for cell := 0; cell < cells; cell++ {
		for ts := 0; ts < n; ts++ {
			inputs.Set3(cell, 0, ts, linear.inflows[ts])
			inputs.Set3(cell, 1, ts, linear.laterals[ts])
			inputs.Set3(cell, 2, ts, linear.rainfall[ts])
			inputs.Set3(cell, 3, ts, linear.evap[ts])
		}
	}

	states := model.InitialiseStates(cells)
	outputs := sim.InitialiseOutputs(model, n, cells)

	sink := sim.NewReportSink(cells, 3, sim.SevWarning, 5)
	sink.BaseIndex = 700 // this generation starts at global run 700
	require.True(t, sim.AttachReportSink(model, sink))

	model.Run(inputs, states, outputs)
	rs := sink.Flush()

	counts := func(cell, code int) uint32 { return rs.Counts[cell*rs.NCodes+code] }
	mb := int(StorageRoutingReportMassBalanceNotConverged)

	assert.Zero(t, counts(0, mb), "the linear cell must stay silent")
	assert.Greater(t, counts(1, mb), uint32(n/2), "the stiff cell must report")

	// Every detail record belongs to the stiff cell, at its global run index.
	require.NotZero(t, rs.Len())
	for i := 0; i < rs.Len(); i++ {
		assert.EqualValues(t, 701, rs.Runs[i], "report %d attributed to the wrong cell", i)
	}
}

// Phase 6: a link that cannot be routed reports the fault and yields NaN,
// rather than panicking and taking the whole simulation with it.

func TestNaNInputAbandonsTheLinkNotTheRun(t *testing.T) {
	const n = 50
	c := wellBehavedCase(n)
	c.inflows[10] = math.NaN()

	out := make([]float64, n)
	store := make([]float64, n)
	sink := sim.NewReportSink(1, 3, sim.SevWarning, 10)

	require.NotPanics(t, func() { c.run(out, store, sink.Cell(0)) },
		"a NaN inflow must not bring down the run")

	rs := sink.Flush()
	require.NotZero(t, rs.Len())
	assert.EqualValues(t, StorageRoutingReportNanInput, rs.Codes[0])
	assert.EqualValues(t, sim.SevError, rs.Severities[0])
	assert.EqualValues(t, 10, rs.Timesteps[0])

	for i := 0; i < 10; i++ {
		assert.False(t, math.IsNaN(out[i]), "timestep %d predates the fault", i)
	}
	for i := 10; i < n; i++ {
		assert.True(t, math.IsNaN(out[i]), "timestep %d should be NaN, not a plausible number", i)
		assert.True(t, math.IsNaN(store[i]))
	}
}

func TestNaNLateralIsAlsoCaught(t *testing.T) {
	const n = 30
	c := wellBehavedCase(n)
	c.laterals[4] = math.NaN()

	out := make([]float64, n)
	store := make([]float64, n)
	sink := sim.NewReportSink(1, 3, sim.SevWarning, 10)

	require.NotPanics(t, func() { c.run(out, store, sink.Cell(0)) })
	assert.EqualValues(t, 1, sink.Flush().Counts[StorageRoutingReportNanInput])
	assert.True(t, math.IsNaN(out[n-1]))
}

func TestRoutingDegradationWorksWithoutAReporter(t *testing.T) {
	const n = 30
	c := wellBehavedCase(n)
	c.inflows[3] = math.NaN()

	out := make([]float64, n)
	store := make([]float64, n)
	assert.NotPanics(t, func() { c.run(out, store, nil) },
		"degradation must not depend on reporting being switched on")
	assert.True(t, math.IsNaN(out[n-1]))
}

// One broken link must not affect the others in the same generation.
func TestOneBadLinkDoesNotAffectTheOthers(t *testing.T) {
	const (
		cells = 3
		n     = 40
	)

	model := sim.Catalog["StorageRouting"]()
	base := wellBehavedCase(n)

	params := data.NewArray2D[float64](6, cells)
	for cell := 0; cell < cells; cell++ {
		for row, v := range []float64{base.bias, base.k, base.x, base.area, base.deadStorage, base.deltaT} {
			params.Set2(row, cell, v)
		}
	}
	model.ApplyParameters(params)

	inputs := data.NewArray3D[float64](cells, 4, n)
	for cell := 0; cell < cells; cell++ {
		for ts := 0; ts < n; ts++ {
			inputs.Set3(cell, 0, ts, base.inflows[ts])
			inputs.Set3(cell, 1, ts, base.laterals[ts])
			inputs.Set3(cell, 2, ts, base.rainfall[ts])
			inputs.Set3(cell, 3, ts, base.evap[ts])
		}
	}
	inputs.Set3(1, 0, 12, math.NaN()) // only the middle link is broken

	states := model.InitialiseStates(cells)
	outputs := sim.InitialiseOutputs(model, n, cells)

	sink := sim.NewReportSink(cells, 3, sim.SevWarning, 10)
	require.True(t, sim.AttachReportSink(model, sink))
	require.NotPanics(t, func() { model.Run(inputs, states, outputs) },
		"one bad link must not take down the generation")

	rs := sink.Flush()
	counts := func(cell int) uint32 {
		return rs.Counts[cell*rs.NCodes+int(StorageRoutingReportNanInput)]
	}
	assert.Zero(t, counts(0))
	assert.NotZero(t, counts(1))
	assert.Zero(t, counts(2))

	for _, cell := range []int{0, 2} {
		for ts := 0; ts < n; ts++ {
			assert.False(t, math.IsNaN(outputs.Get3(cell, 0, ts)),
				"cell %d timestep %d should be unaffected", cell, ts)
		}
	}
	assert.False(t, math.IsNaN(outputs.Get3(1, 0, 11)))
	assert.True(t, math.IsNaN(outputs.Get3(1, 0, 12)))
}

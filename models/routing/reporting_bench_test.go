package routing

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flowmatters/openwater-core/sim"
)

// Benchmarks for StorageRouting's reporting, against the real instrumented
// kernel. Phase 1 measured the mechanism itself against an uninstrumented copy;
// what matters now is that the shipped kernel costs nothing extra when
// reporting is off, and little when it is on.

const (
	benchTimesteps = 13514 // matches the Burdekin model's run length
	benchCells     = 96    // StorageRouting's widest generation in that model
	benchNumCodes  = 3
)

type routingCase struct {
	inflows, laterals, rainfall, evap []float64
	bias, k, x, area, deadStorage     float64
	deltaT                            float64
}

func newCase(n int, bias, k, x float64) routingCase {
	c := routingCase{
		inflows:  make([]float64, n),
		laterals: make([]float64, n),
		rainfall: make([]float64, n),
		evap:     make([]float64, n),
		bias:     bias,
		k:        k,
		x:        x,
		area:     10000.0,
		deltaT:   86400.0,
	}
	for i := range c.inflows {
		c.inflows[i] = 5.0 + 4.0*math.Sin(float64(i)/7.0)
		c.laterals[i] = 0.5
		c.evap[i] = 3.0
	}
	return c
}

// wellBehavedCase routes water with a linear store, where the root finder
// converges on every timestep and nothing is reported.
func wellBehavedCase(n int) routingCase { return newCase(n, 0.0, 43200.0, 1.0) }

// stiffCase uses a strongly non-linear store (routing power 2.5) with a small
// inflow bias. These are legal parameters, and the root finder fails to hit the
// mass balance limit within its 20 iterations on most timesteps -- which is
// exactly the condition storage_routing.go had an empty if-block waiting for.
func stiffCase(n int) routingCase { return newCase(n, 0.1, 86400.0, 2.5) }

func (c routingCase) run(outflows, storages []float64, rep *sim.Reporter) {
	storageRouting(c.inflows, c.laterals, c.rainfall, c.evap,
		0, 0, 0,
		c.bias, c.k, c.x, c.area, c.deadStorage, c.deltaT,
		outflows, storages, rep)
}

// TestBenchCasesBehaveAsIntended pins the premise of the benchmarks: one case
// must be silent and the other must report on most timesteps.
func TestBenchCasesBehaveAsIntended(t *testing.T) {
	const n = 200

	for name, tc := range map[string]struct {
		c       routingCase
		wantMin uint32
		wantMax uint32
	}{
		"well behaved": {wellBehavedCase(n), 0, 0},
		"stiff":        {stiffCase(n), n / 2, n},
	} {
		out := make([]float64, n)
		store := make([]float64, n)
		sink := sim.NewReportSink(1, benchNumCodes, sim.SevWarning, 0)

		tc.c.run(out, store, sink.Cell(0))
		rs := sink.Flush()

		fired := rs.Counts[StorageRoutingReportMassBalanceNotConverged]
		assert.GreaterOrEqual(t, fired, tc.wantMin, "%s fired too rarely", name)
		assert.LessOrEqual(t, fired, tc.wantMax, "%s fired too often", name)

		// Neither case should be producing NaNs.
		assert.Zero(t, rs.Counts[StorageRoutingReportNanInput], "%s: NaN input", name)
		assert.Zero(t, rs.Counts[StorageRoutingReportNanOutflow], "%s: NaN outflow", name)
		for i, v := range out {
			require.False(t, math.IsNaN(v), "%s: NaN outflow at step %d", name, i)
		}
	}
}

func benchmarkRouting(b *testing.B, c routingCase, rep *sim.Reporter) {
	outflows := make([]float64, len(c.inflows))
	storages := make([]float64, len(c.inflows))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.run(outflows, storages, rep)
	}
}

func enabledReporter() *sim.Reporter {
	return sim.NewReportSink(1, benchNumCodes, sim.SevWarning, 3).Cell(0)
}

func BenchmarkRoutingSilent_ReportingOff(b *testing.B) {
	benchmarkRouting(b, wellBehavedCase(benchTimesteps), nil)
}

func BenchmarkRoutingSilent_ReportingOn(b *testing.B) {
	benchmarkRouting(b, wellBehavedCase(benchTimesteps), enabledReporter())
}

func BenchmarkRoutingStiff_ReportingOff(b *testing.B) {
	benchmarkRouting(b, stiffCase(benchTimesteps), nil)
}

func BenchmarkRoutingStiff_ReportingOn(b *testing.B) {
	benchmarkRouting(b, stiffCase(benchTimesteps), enabledReporter())
}

// One goroutine per cell, as the generated Run method does. This is where any
// false sharing between neighbouring cells would show up.
func benchmarkConcurrent(b *testing.B, c routingCase, sink *sim.ReportSink) {
	outflows := make([][]float64, benchCells)
	storages := make([][]float64, benchCells)
	for i := range outflows {
		outflows[i] = make([]float64, len(c.inflows))
		storages[i] = make([]float64, len(c.inflows))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for cell := 0; cell < benchCells; cell++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				c.run(outflows[j], storages[j], sink.Cell(j))
			}(cell)
		}
		wg.Wait()
	}
}

func BenchmarkRoutingConcurrentStiff_ReportingOff(b *testing.B) {
	benchmarkConcurrent(b, stiffCase(benchTimesteps), nil)
}

func BenchmarkRoutingConcurrentStiff_ReportingOn(b *testing.B) {
	benchmarkConcurrent(b, stiffCase(benchTimesteps),
		sim.NewReportSink(benchCells, benchNumCodes, sim.SevWarning, 3))
}

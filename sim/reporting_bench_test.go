package sim

import (
	"sync"
	"testing"
)

// The per-call cost of the reporting API, isolated from any model.

func BenchmarkReporterDisabled(b *testing.B) {
	var r *Reporter // reporting off is expressed as a nil Reporter
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Warn2(0, int32(i), 1.5, 2.5)
	}
}

func BenchmarkReporterBelowThreshold(b *testing.B) {
	r := NewReportSink(1, 1, SevError, 3).Cell(0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Warn2(0, int32(i), 1.5, 2.5)
	}
}

// The steady state for a condition that fires repeatedly: past the detail cap,
// so it costs a call plus a counter increment.
func BenchmarkReporterCounting(b *testing.B) {
	r := NewReportSink(1, 1, SevWarning, 3).Cell(0)
	r.Warn2(0, 0, 1, 2)
	r.Warn2(0, 1, 1, 2)
	r.Warn2(0, 2, 1, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Warn2(0, int32(i), 1.5, 2.5)
	}
}

func BenchmarkReporterRecording(b *testing.B) {
	r := NewReportSink(1, 1, SevWarning, 0).Cell(0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Warn2(0, int32(i), 1.5, 2.5)
	}
}

// Does the cache-line padding in NewReportSink earn its keep? These increment
// per-cell counters from one goroutine per cell, exactly as the generated Run
// methods do, with and without the padded stride.
func benchmarkCounterStride(b *testing.B, cells, codes, stride int) {
	counts := make([]uint32, cells*stride)
	const perGoroutine = 20000

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for c := 0; c < cells; c++ {
			wg.Add(1)
			go func(cell int) {
				defer wg.Done()
				slot := &counts[cell*stride]
				for j := 0; j < perGoroutine; j++ {
					*slot++
				}
			}(c)
		}
		wg.Wait()
	}
}

const (
	benchCells = 96 // StorageRouting's widest generation in the Burdekin model
	benchCodes = 3
)

func BenchmarkCounterStridePacked(b *testing.B) {
	benchmarkCounterStride(b, benchCells, benchCodes, benchCodes)
}

func BenchmarkCounterStridePadded(b *testing.B) {
	stride := ((benchCodes + countsPerLine - 1) / countsPerLine) * countsPerLine
	benchmarkCounterStride(b, benchCells, benchCodes, stride)
}

func BenchmarkSinkAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewReportSink(1908, 8, SevWarning, 3) // the widest generation seen
	}
}

func BenchmarkSinkFlush(b *testing.B) {
	s := NewReportSink(benchCells, benchCodes, SevWarning, 3)
	for c := 0; c < benchCells; c++ {
		for code := int32(0); code < benchCodes; code++ {
			for n := 0; n < 5; n++ {
				s.Cell(c).Warn2(code, int32(n), 1, 2)
			}
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Flush()
	}
}

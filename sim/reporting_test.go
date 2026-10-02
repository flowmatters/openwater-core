package sim

import (
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/flowmatters/openwater-core/data"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporterSizeIsCacheLineMultiple(t *testing.T) {
	size := unsafe.Sizeof(Reporter{})
	assert.Zero(t, size%cacheLineBytes,
		"Reporter must be a whole number of cache lines, got %d bytes -- adjust reporterFieldBytes", size)

	// Guards the hand-computed padding constant: if a field is added or
	// removed, this catches it before the padding silently stops working.
	fields := unsafe.Sizeof(Reporter{}.counts) +
		unsafe.Sizeof(Reporter{}.details) +
		unsafe.Sizeof(Reporter{}.minSeverity) +
		unsafe.Sizeof(Reporter{}.detailLimit) +
		unsafe.Sizeof(Reporter{}.suppressed)
	assert.EqualValues(t, reporterFieldBytes, fields, "reporterFieldBytes is stale")
}

func TestNeighbouringCellsDoNotShareACacheLine(t *testing.T) {
	s := NewReportSink(8, 3, SevWarning, 10)
	require.NotNil(t, s)

	// Each cell's counters must start at least a cache line apart, otherwise
	// concurrent increments from adjacent cells contend.
	for i := 1; i < s.NCells(); i++ {
		prev := uintptr(unsafe.Pointer(&s.cells[i-1].counts[0]))
		cur := uintptr(unsafe.Pointer(&s.cells[i].counts[0]))
		assert.GreaterOrEqual(t, uint64(cur-prev), uint64(cacheLineBytes),
			"counters for cells %d and %d are less than a cache line apart", i-1, i)
	}
}

func TestNilReporterIsInert(t *testing.T) {
	var r *Reporter

	assert.NotPanics(t, func() {
		r.Warn0(0, 1)
		r.Warn1(0, 1, 1)
		r.Warn2(0, 1, 1, 2)
		r.Warn3(0, 1, 1, 2, 3)
		r.Err0(0, 1)
		r.Err1(0, 1, 1)
		r.Err2(0, 1, 1, 2)
		r.Err3(0, 1, 1, 2, 3)
	})
	assert.False(t, r.Enabled())
	assert.Zero(t, r.Suppressed())
}

func TestNilSinkIsInert(t *testing.T) {
	var s *ReportSink

	assert.Nil(t, s.Cell(0))
	assert.Zero(t, s.NCells())
	assert.Zero(t, s.Flush().Len())

	// Reporting off is expressed as a nil sink, so these must all hold.
	assert.Nil(t, NewReportSink(0, 3, SevWarning, 5))
	assert.Nil(t, NewReportSink(4, 0, SevWarning, 5))
	assert.Nil(t, NewReportSink(4, 3, SevNone, 5))
}

func TestSeverityThreshold(t *testing.T) {
	s := NewReportSink(1, 2, SevError, 10)
	r := s.Cell(0)

	r.Warn1(0, 5, 1.5) // below threshold, dropped
	r.Err1(1, 6, 2.5)

	rs := s.Flush()
	require.Equal(t, 1, rs.Len())
	assert.EqualValues(t, SevError, rs.Severities[0])
	assert.EqualValues(t, 1, rs.Codes[0])
	assert.EqualValues(t, 6, rs.Timesteps[0])
	assert.Equal(t, uint32(0), rs.Counts[0], "a report below the threshold must not be counted either")
	assert.Equal(t, uint32(1), rs.Counts[1])
}

func TestDetailCapKeepsCountsComplete(t *testing.T) {
	const limit = 3
	s := NewReportSink(1, 1, SevWarning, limit)
	r := s.Cell(0)

	for ts := int32(0); ts < 1000; ts++ {
		r.Warn1(0, ts, float64(ts))
	}

	rs := s.Flush()
	assert.Equal(t, limit, rs.Len(), "detail records are capped")
	assert.Equal(t, uint32(1000), rs.Counts[0], "counts stay complete past the cap")
	assert.EqualValues(t, 1000-limit, rs.Suppressed)

	// The records kept are the first ones, which is what you want for
	// diagnosis -- the onset of the condition, not its tail.
	for i := 0; i < limit; i++ {
		assert.EqualValues(t, i, rs.Timesteps[i])
	}
}

func TestUnlimitedDetailWhenCapIsZero(t *testing.T) {
	s := NewReportSink(1, 1, SevWarning, 0)
	r := s.Cell(0)
	for ts := int32(0); ts < 50; ts++ {
		r.Warn0(0, ts)
	}
	rs := s.Flush()
	assert.Equal(t, 50, rs.Len())
	assert.Zero(t, rs.Suppressed)
}

func TestUndeclaredCodeIsDroppedNotRecorded(t *testing.T) {
	s := NewReportSink(1, 2, SevWarning, 10)
	r := s.Cell(0)

	r.Warn0(7, 1)  // beyond the catalogue
	r.Warn0(-1, 2) // negative

	rs := s.Flush()
	assert.Zero(t, rs.Len(), "an undeclared code must not grow the detail slice without bound")
	assert.EqualValues(t, 2, rs.Suppressed)
}

func TestFlushIsOrderedByCellAndCarriesRunIndex(t *testing.T) {
	s := NewReportSink(3, 1, SevWarning, 10)
	s.BaseIndex = 100

	// Report out of cell order; Flush must still come back in cell order.
	s.Cell(2).Warn1(0, 20, 2)
	s.Cell(0).Warn1(0, 10, 0)
	s.Cell(1).Warn1(0, 15, 1)

	rs := s.Flush()
	require.Equal(t, 3, rs.Len())
	assert.Equal(t, []int32{100, 101, 102}, rs.Runs)
	assert.Equal(t, []int32{10, 15, 20}, rs.Timesteps)
	assert.Equal(t, 3*MaxReportArgs, len(rs.Args))
	assert.Equal(t, 0.0, rs.Args[0])
	assert.Equal(t, 1.0, rs.Args[MaxReportArgs])
	assert.Equal(t, 2.0, rs.Args[2*MaxReportArgs])
}

func TestArgsAreRecordedPerArity(t *testing.T) {
	s := NewReportSink(1, 4, SevWarning, 10)
	r := s.Cell(0)
	r.Warn0(0, 1)
	r.Warn1(1, 2, 1.5)
	r.Warn2(2, 3, 1.5, 2.5)
	r.Warn3(3, 4, 1.5, 2.5, 3.5)

	rs := s.Flush()
	require.Equal(t, 4, rs.Len())
	assert.Equal(t, []float64{
		0, 0, 0,
		1.5, 0, 0,
		1.5, 2.5, 0,
		1.5, 2.5, 3.5,
	}, rs.Args)
}

// Reporters are written by one goroutine per cell and read after those
// goroutines are joined -- exactly how the generated Run methods work. Run
// under -race to check the claim that no synchronisation is needed.
func TestConcurrentPerCellReportingIsRaceFree(t *testing.T) {
	const (
		cells     = 64
		codes     = 4
		timesteps = 500
	)
	s := NewReportSink(cells, codes, SevWarning, 1000000)

	var wg sync.WaitGroup
	for c := 0; c < cells; c++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := s.Cell(i)
			for ts := int32(0); ts < timesteps; ts++ {
				r.Warn2(int32(i%codes), ts, float64(i), float64(ts))
			}
		}(c)
	}
	wg.Wait()

	rs := s.Flush()
	assert.Equal(t, cells*timesteps, rs.Len())
	for c := 0; c < cells; c++ {
		assert.Equal(t, uint32(timesteps), rs.Counts[c*codes+(c%codes)],
			"cell %d lost counts", c)
	}
}

func TestReportSetAppend(t *testing.T) {
	first := NewReportSink(2, 2, SevWarning, 10)
	first.BaseIndex = 0
	first.Cell(0).Warn0(0, 1)

	second := NewReportSink(2, 2, SevWarning, 10)
	second.BaseIndex = 2
	second.Cell(1).Err0(1, 9)

	rs := first.Flush()
	rs.Append(second.Flush())

	assert.Equal(t, 2, rs.Len())
	assert.Equal(t, []int32{0, 3}, rs.Runs)
	assert.Equal(t, 4, rs.NCells)
	assert.Equal(t, 4*2, len(rs.Counts))
	assert.Equal(t, uint32(1), rs.Counts[0*2+0]) // run 0, code 0
	assert.Equal(t, uint32(1), rs.Counts[3*2+1]) // run 3, code 1
}

func TestAppendEmptyIsNoop(t *testing.T) {
	rs := ReportSet{}
	rs.Append(ReportSet{})
	assert.Zero(t, rs.Len())
	assert.Zero(t, rs.NCells)
}

func TestRenderSubstitutesArgs(t *testing.T) {
	d := ReportDescription{
		Name:     "massBalance",
		Severity: SevWarning,
		Template: "mass balance error {0} m^3 exceeds limit {1} after {2} iterations",
		NArgs:    3,
	}
	msg := d.Render(Report{NArgs: 3, Args: [MaxReportArgs]float64{0.0025, 0.001, 20}})
	assert.Equal(t, "mass balance error 0.0025 m^3 exceeds limit 0.001 after 20 iterations", msg)
}

func TestParseSeverity(t *testing.T) {
	for text, expected := range map[string]Severity{
		"warning":   SevWarning,
		"error":     SevError,
		" warning ": SevWarning,
	} {
		got, err := ParseSeverity(text)
		assert.NoError(t, err)
		assert.Equal(t, expected, got, "for %q", text)
	}

	_, err := ParseSeverity("info")
	assert.Error(t, err, "reporting is warning-and-above only")
	_, err = ParseSeverity("")
	assert.Error(t, err)
}

func TestSeverityString(t *testing.T) {
	assert.Equal(t, "warning", SevWarning.String())
	assert.Equal(t, "error", SevError.String())
	assert.Equal(t, "none", SevNone.String())
}

type stubModel struct{}

func (m *stubModel) Description() ModelDescription          { return ModelDescription{} }
func (m *stubModel) InitialiseDimensions(dims []int)        {}
func (m *stubModel) FindDimensions(data.ND2[float64]) []int { return nil }
func (m *stubModel) ApplyParameters(data.ND2[float64])      {}
func (m *stubModel) InitialiseStates(n int) data.ND2[float64] {
	return data.NewArray2D[float64](n, 0)
}
func (m *stubModel) Run(data.ND3[float64], data.ND2[float64], data.ND3[float64]) {}

type reportingModel struct {
	stubModel
	sink *ReportSink
}

func (m *reportingModel) SetReportSink(s *ReportSink)        { m.sink = s }
func (m *reportingModel) ReportCatalog() []ReportDescription { return nil }

func TestAttachReportSinkIsOptional(t *testing.T) {
	sink := NewReportSink(1, 1, SevWarning, 1)

	// A model that never declared a reports: block is left completely alone.
	// That is the whole point of opting in.
	plain := &stubModel{}
	assert.False(t, AttachReportSink(plain, sink))

	reporting := &reportingModel{}
	assert.True(t, AttachReportSink(reporting, sink))
	assert.Same(t, sink, reporting.sink)
}

// The Warn/Err methods are only cheap when reporting is off if they inline
// down to a nil check. record() is marked //go:noinline precisely so that
// append() does not blow their inline budget; this asserts the arrangement
// still holds, because a future edit could silently undo it.
func TestReporterCallsInline(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the compiler")
	}

	out, err := exec.Command("go", "build", "-gcflags=-m", "./").CombinedOutput()
	require.NoError(t, err, "go build -gcflags=-m failed: %s", out)

	text := string(out)
	for _, method := range []string{"Warn0", "Warn1", "Warn2", "Warn3", "Err0", "Err1", "Err2", "Err3"} {
		pattern := regexp.MustCompile(`can inline \(\*Reporter\)\.` + method + `\b`)
		assert.True(t, pattern.MatchString(text),
			"(*Reporter).%s is no longer inlinable -- the disabled path now costs a call per timestep", method)
	}

	// A //go:noinline function is simply absent from the "can inline" list.
	assert.False(t, strings.Contains(text, "can inline (*Reporter).record"),
		"record must stay out of line so the Warn/Err guards keep inlining")
}

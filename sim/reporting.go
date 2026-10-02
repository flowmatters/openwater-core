package sim

import (
	"fmt"
	"strconv"
	"strings"
)

// Model reporting.
//
// Model kernels run per cell per timestep and have no idea which subcatchment,
// functional unit or date they represent -- so they can't produce a message a
// user could act on. What they can do is say "this happened, at timestep N,
// with these numbers". Everything else (which place, which day, what wording)
// is added later by code that does have the model file's metadata.
//
// Three properties drive the design:
//
//   - No formatting in the timestep loop. A report is a code plus up to three
//     float64s; the message text lives in the model's catalogue and is rendered
//     only when someone reads the reports.
//   - No allocation on the disabled path. Reporter methods take fixed arity
//     rather than ...float64, because a variadic call builds its slice at the
//     call site -- before the nil check can short-circuit it.
//   - No synchronisation. Generated Run methods spawn one goroutine per cell,
//     and each gets its own Reporter.

// MaxReportArgs is the number of numeric arguments a single report can carry.
const MaxReportArgs = 3

// Severity classifies a report. Reporting is deliberately limited to warning
// and above: at one report per cell per timestep, anything more verbose floods
// faster than it informs.
type Severity int32

const (
	// SevNone is the zero value, and means "no reporting". A sink with
	// SevNone records nothing.
	SevNone Severity = 0
	// SevWarning marks a result that is still usable but questionable --
	// a solver that stopped on an iteration limit rather than converging.
	SevWarning Severity = 1
	// SevError marks a result that is wrong or was forced into range --
	// a value clamped to keep the run going.
	SevError Severity = 2
)

func (s Severity) String() string {
	switch s {
	case SevNone:
		return "none"
	case SevWarning:
		return "warning"
	case SevError:
		return "error"
	}
	return "Severity(" + strconv.Itoa(int(s)) + ")"
}

// ParseSeverity converts the severity keyword used in an OW-SPEC reports:
// block into a Severity.
func ParseSeverity(text string) (Severity, error) {
	switch strings.TrimSpace(text) {
	case "warning":
		return SevWarning, nil
	case "error":
		return SevError, nil
	}
	return SevNone, fmt.Errorf("unknown report severity %q (expected warning or error)", text)
}

// Report is one recorded event from one cell. Fixed size and pointer-free, so
// a slice of them is one allocation and the GC never has to scan it.
type Report struct {
	Timestep int32
	Code     int32
	Severity Severity
	NArgs    int32
	Args     [MaxReportArgs]float64
}

// ReportDescription is a model's catalogue entry for one report code,
// generated from the model's OW-SPEC reports: block.
type ReportDescription struct {
	Name     string
	Severity Severity
	Template string
	NArgs    int
}

// Render substitutes a report's arguments into the description's template,
// replacing {0}, {1} and {2}.
func (d ReportDescription) Render(r Report) string {
	msg := d.Template
	for i := 0; i < int(r.NArgs) && i < MaxReportArgs; i++ {
		msg = strings.ReplaceAll(msg,
			"{"+strconv.Itoa(i)+"}",
			strconv.FormatFloat(r.Args[i], 'g', 6, 64))
	}
	return msg
}

// Reporting is implemented by models that declare a reports: block. It is an
// optional interface -- like Hotstartable -- so models that don't report are
// completely unaffected, and neither the TimeSteppingModel interface nor the
// model signature hash changes.
type Reporting interface {
	SetReportSink(*ReportSink)
	ReportCatalog() []ReportDescription
}

// AttachReportSink gives a model a sink to report into, if it reports at all.
// It returns whether the model accepted one.
func AttachReportSink(m TimeSteppingModel, s *ReportSink) bool {
	r, ok := m.(Reporting)
	if !ok {
		return false
	}
	r.SetReportSink(s)
	return true
}

const (
	cacheLineBytes = 64
	countsPerLine  = cacheLineBytes / 4 // uint32s per cache line

	// Fields of Reporter, in bytes on a 64-bit platform. TestReporterSize
	// asserts this stays right; getting it wrong is a compile error here.
	reporterFieldBytes = 60
)

// Reporter collects the reports for a single cell.
//
// A Reporter is owned by exactly one goroutine for the duration of a Run, so
// nothing here synchronises. The padding matters though: Reporters live in one
// slice written concurrently by per-cell goroutines, and without it
// neighbouring cells share a cache line.
type Reporter struct {
	counts      []uint32 // one entry per code; a window onto the sink's backing array
	details     []Report
	minSeverity Severity
	detailLimit int32
	suppressed  int32

	_ [2*cacheLineBytes - reporterFieldBytes]byte
}

// Enabled reports whether this Reporter will record anything. Use it to guard
// a report whose arguments are themselves expensive to compute; for ordinary
// reports just call Warn/Err, which check internally.
func (r *Reporter) Enabled() bool {
	return r != nil && r.minSeverity != SevNone
}

// Suppressed is the number of reports dropped by the per-code detail cap. The
// counts are still complete -- only the detail records are capped.
func (r *Reporter) Suppressed() int32 {
	if r == nil {
		return 0
	}
	return r.suppressed
}

// The Warn/Err families are the hot-path API. Each is a bare guard plus a call
// into record, which is deliberately kept out of line so these stay inlinable
// -- see TestReporterCallsInline.

func (r *Reporter) Warn0(code, timestep int32) {
	if r == nil || SevWarning < r.minSeverity {
		return
	}
	r.record(SevWarning, code, timestep, 0, 0, 0, 0)
}

func (r *Reporter) Warn1(code, timestep int32, a float64) {
	if r == nil || SevWarning < r.minSeverity {
		return
	}
	r.record(SevWarning, code, timestep, 1, a, 0, 0)
}

func (r *Reporter) Warn2(code, timestep int32, a, b float64) {
	if r == nil || SevWarning < r.minSeverity {
		return
	}
	r.record(SevWarning, code, timestep, 2, a, b, 0)
}

func (r *Reporter) Warn3(code, timestep int32, a, b, c float64) {
	if r == nil || SevWarning < r.minSeverity {
		return
	}
	r.record(SevWarning, code, timestep, 3, a, b, c)
}

func (r *Reporter) Err0(code, timestep int32) {
	if r == nil || SevError < r.minSeverity {
		return
	}
	r.record(SevError, code, timestep, 0, 0, 0, 0)
}

func (r *Reporter) Err1(code, timestep int32, a float64) {
	if r == nil || SevError < r.minSeverity {
		return
	}
	r.record(SevError, code, timestep, 1, a, 0, 0)
}

func (r *Reporter) Err2(code, timestep int32, a, b float64) {
	if r == nil || SevError < r.minSeverity {
		return
	}
	r.record(SevError, code, timestep, 2, a, b, 0)
}

func (r *Reporter) Err3(code, timestep int32, a, b, c float64) {
	if r == nil || SevError < r.minSeverity {
		return
	}
	r.record(SevError, code, timestep, 3, a, b, c)
}

// record is the out-of-line slow path. Keeping append() out of the Warn/Err
// methods is what lets them inline down to a nil check when reporting is off.
//
//go:noinline
func (r *Reporter) record(sev Severity, code, timestep, nargs int32, a, b, c float64) {
	if code < 0 || int(code) >= len(r.counts) {
		// Not a code this model declared. Generated wrappers make this
		// unreachable; drop it rather than growing details without bound.
		r.suppressed++
		return
	}

	r.counts[code]++

	// The count is complete for every event. Only the detail records -- the
	// ones carrying timesteps and arguments -- are capped, so a condition that
	// fires every timestep on every cell costs a handful of records plus an
	// increment, not a record per timestep.
	if r.detailLimit > 0 && r.counts[code] > uint32(r.detailLimit) {
		r.suppressed++
		return
	}

	r.details = append(r.details, Report{
		Timestep: timestep,
		Code:     code,
		Severity: sev,
		NArgs:    nargs,
		Args:     [MaxReportArgs]float64{a, b, c},
	})
}

// ReportSink holds the Reporters for one model generation.
type ReportSink struct {
	// BaseIndex is the global run index of cell 0, so cell i is run
	// BaseIndex+i. Generations are contiguous slices of a model's runs.
	BaseIndex   int
	MinSeverity Severity
	DetailLimit int32
	NCodes      int

	cells  []Reporter
	counts []uint32
	stride int
}

// NewReportSink builds a sink for nCells cells over a catalogue of nCodes
// report codes. detailLimit caps the detail records kept per (cell, code);
// zero or less means no cap. It returns nil -- a usable, entirely inert sink --
// when reporting is off or there is nothing to report on.
func NewReportSink(nCells, nCodes int, minSeverity Severity, detailLimit int32) *ReportSink {
	if nCells <= 0 || nCodes <= 0 || minSeverity == SevNone {
		return nil
	}

	// Pad each cell's counters out to a whole number of cache lines so that
	// concurrent increments from neighbouring cells don't contend.
	stride := ((nCodes + countsPerLine - 1) / countsPerLine) * countsPerLine

	s := &ReportSink{
		MinSeverity: minSeverity,
		DetailLimit: detailLimit,
		NCodes:      nCodes,
		stride:      stride,
		cells:       make([]Reporter, nCells),
		counts:      make([]uint32, nCells*stride),
	}

	for i := range s.cells {
		lo := i * stride
		s.cells[i].counts = s.counts[lo : lo+nCodes : lo+nCodes]
		s.cells[i].minSeverity = minSeverity
		s.cells[i].detailLimit = detailLimit
	}

	return s
}

// Cell returns the Reporter for cell i, or nil if the sink is nil. Passing the
// nil back into a model kernel is the intended way to run with reporting off.
func (s *ReportSink) Cell(i int) *Reporter {
	if s == nil {
		return nil
	}
	return &s.cells[i]
}

// NCells is the number of cells this sink covers.
func (s *ReportSink) NCells() int {
	if s == nil {
		return 0
	}
	return len(s.cells)
}

// ReportSet is the flattened, storage-ready form of a sink's contents: struct
// of arrays, one entry per detail record, plus the complete per-cell counts.
// This is the shape that goes to HDF5.
type ReportSet struct {
	Runs       []int32
	Timesteps  []int32
	Codes      []int32
	Severities []int32
	Args       []float64 // MaxReportArgs per record

	// Counts is dense, cell-major, NCells x NCodes, and complete: it covers
	// every event, including those whose detail records were capped.
	Counts []uint32
	NCells int
	NCodes int

	Suppressed int64
}

// Len is the number of detail records.
func (rs ReportSet) Len() int { return len(rs.Runs) }

// Flush flattens the sink in cell order. Ordering by cell index rather than by
// goroutine completion keeps results reproducible across runs.
//
// Call it only after every per-cell goroutine has been joined: the channel
// receive that joins them is what makes their writes visible here.
func (s *ReportSink) Flush() ReportSet {
	if s == nil {
		return ReportSet{}
	}

	total := 0
	for i := range s.cells {
		total += len(s.cells[i].details)
	}

	rs := ReportSet{
		Runs:       make([]int32, 0, total),
		Timesteps:  make([]int32, 0, total),
		Codes:      make([]int32, 0, total),
		Severities: make([]int32, 0, total),
		Args:       make([]float64, 0, total*MaxReportArgs),
		Counts:     make([]uint32, len(s.cells)*s.NCodes),
		NCells:     len(s.cells),
		NCodes:     s.NCodes,
	}

	for i := range s.cells {
		c := &s.cells[i]
		run := int32(s.BaseIndex + i)

		for _, d := range c.details {
			rs.Runs = append(rs.Runs, run)
			rs.Timesteps = append(rs.Timesteps, d.Timestep)
			rs.Codes = append(rs.Codes, d.Code)
			rs.Severities = append(rs.Severities, int32(d.Severity))
			rs.Args = append(rs.Args, d.Args[0], d.Args[1], d.Args[2])
		}

		copy(rs.Counts[i*s.NCodes:(i+1)*s.NCodes], c.counts)
		rs.Suppressed += int64(c.suppressed)
	}

	return rs
}

// Append concatenates another ReportSet onto this one, for accumulating
// generations across a whole run. Counts are stacked cell-major, so the
// combined Counts stays aligned with the combined run indices.
func (rs *ReportSet) Append(other ReportSet) {
	if other.NCells == 0 && other.Len() == 0 {
		return
	}
	if rs.NCodes == 0 {
		rs.NCodes = other.NCodes
	}
	if other.NCodes != 0 && other.NCodes != rs.NCodes {
		panic(fmt.Sprintf("ReportSet.Append: code count mismatch (%d vs %d)", rs.NCodes, other.NCodes))
	}

	rs.Runs = append(rs.Runs, other.Runs...)
	rs.Timesteps = append(rs.Timesteps, other.Timesteps...)
	rs.Codes = append(rs.Codes, other.Codes...)
	rs.Severities = append(rs.Severities, other.Severities...)
	rs.Args = append(rs.Args, other.Args...)
	rs.Counts = append(rs.Counts, other.Counts...)
	rs.NCells += other.NCells
	rs.Suppressed += other.Suppressed
}

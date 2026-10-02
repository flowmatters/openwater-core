package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flowmatters/openwater-core/io"
	"github.com/flowmatters/openwater-core/sim"
)

func testCatalog() []sim.ReportDescription {
	return []sim.ReportDescription{
		{Name: "converged", Severity: sim.SevWarning, Template: "did not converge: {0} vs {1}", NArgs: 2},
		{Name: "broken", Severity: sim.SevError, Template: "unusable configuration ({0} points)", NArgs: 1},
	}
}

// refWithReports builds a modelReference carrying the reports two generations
// would have produced, without running a simulation.
func refWithReports(name string, gens ...*sim.ReportSink) *modelReference {
	mr := &modelReference{ModelName: name, reportCatalog: testCatalog()}
	for _, sink := range gens {
		mr.CollectReports(sink)
	}
	return mr
}

func TestCollectReportsAccumulatesAcrossGenerations(t *testing.T) {
	// Generation 0: two cells starting at run 0. Generation 1: one cell at run 2.
	first := sim.NewReportSink(2, 2, sim.SevWarning, 3)
	first.BaseIndex = 0
	first.Cell(1).Warn2(0, 5, 1.5, 2.5)

	second := sim.NewReportSink(1, 2, sim.SevWarning, 3)
	second.BaseIndex = 2
	second.Cell(0).Err1(1, -1, 3)

	mr := refWithReports("Widget", first, second)
	set := mr.Reports()

	assert.Equal(t, 3, set.NCells, "cells accumulate across generations")
	require.Equal(t, 2, set.Len())
	assert.Equal(t, []int32{1, 2}, set.Runs, "run indices must be global, not per generation")
	assert.Equal(t, []int32{5, -1}, set.Timesteps)
	assert.Equal(t, uint32(1), set.Counts[1*set.NCodes+0])
	assert.Equal(t, uint32(1), set.Counts[2*set.NCodes+1])
}

// Reports must survive PurgeGeneration, which is why they are collected onto
// the modelReference rather than left on the generation.
func TestReportsSurviveGenerationPurge(t *testing.T) {
	sink := sim.NewReportSink(1, 2, sim.SevWarning, 3)
	sink.Cell(0).Warn2(0, 7, 1, 2)

	mr := refWithReports("Widget", sink)
	mr.Generations = []*modelGeneration{{}}
	mr.PurgeGeneration(0)

	assert.Equal(t, 1, mr.Reports().Len(), "purging a generation must not discard its reports")
}

func TestSummariseReports(t *testing.T) {
	sink := sim.NewReportSink(3, 2, sim.SevWarning, 2)
	// Cell 0 trips the warning many times; only two details survive the cap.
	for ts := int32(0); ts < 10; ts++ {
		sink.Cell(0).Warn2(0, 20-ts, 1, 2)
	}
	// Cell 2 raises the error once.
	sink.Cell(2).Err1(1, -1, 4)

	mr := refWithReports("Widget", sink)
	models := map[string]*modelReference{"Widget": mr}

	summaries := summariseReports(models, []string{"Widget"})
	require.Len(t, summaries, 2)

	// Most severe first.
	assert.Equal(t, "broken", summaries[0].Name)
	assert.Equal(t, sim.SevError, summaries[0].Severity)
	assert.EqualValues(t, 1, summaries[0].Total)
	assert.Equal(t, 1, summaries[0].Cells)
	assert.EqualValues(t, -1, summaries[0].FirstStep, "a configuration fault has no timestep")

	assert.Equal(t, "converged", summaries[1].Name)
	assert.EqualValues(t, 10, summaries[1].Total, "counts are complete despite the detail cap")
	assert.Equal(t, 1, summaries[1].Cells)
	assert.EqualValues(t, 19, summaries[1].FirstStep, "earliest surviving timestep")
}

func TestSummariseSkipsModelsThatDoNotReport(t *testing.T) {
	plain := &modelReference{ModelName: "Plain"}
	models := map[string]*modelReference{"Plain": plain}
	assert.Empty(t, summariseReports(models, []string{"Plain"}))
}

func TestExitCodeForReports(t *testing.T) {
	warnOnly := []reportSummary{{Severity: sim.SevWarning, Total: 5}}
	withError := []reportSummary{{Severity: sim.SevError, Total: 1}}

	assert.Zero(t, exitCodeForReports(nil))
	assert.Zero(t, exitCodeForReports(warnOnly), "warnings alone must not fail a run")
	assert.Equal(t, 1, exitCodeForReports(withError),
		"an error-severity report must fail the run, or converting panics to reports would hide breakage")

	original := *reportsFatal
	defer func() { *reportsFatal = original }()
	*reportsFatal = "off"
	assert.Zero(t, exitCodeForReports(withError), "-reports-fatal=off disables the exit code")
}

func TestWriteReportsRoundTrip(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "results.h5")

	sink := sim.NewReportSink(2, 2, sim.SevWarning, 3)
	sink.BaseIndex = 10
	for ts := int32(0); ts < 5; ts++ {
		sink.Cell(1).Warn2(0, ts, float64(ts), 99)
	}

	mr := refWithReports("Widget", sink)
	models := map[string]*modelReference{"Widget": mr}
	require.NoError(t, writeReports(models, []string{"Widget"}, fn))

	read := func(name string) []int32 {
		ref := io.H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/Widget/" + name}
		loaded, err := ref.Load()
		require.NoError(t, err, name)
		return loaded.Unroll()
	}

	// The detail cap keeps three records; the counts keep all five.
	assert.Equal(t, []int32{11, 11, 11}, read("run"))
	assert.Equal(t, []int32{0, 1, 2}, read("timestep"))
	assert.Equal(t, []int32{0, 0, 0}, read("code"))
	assert.Equal(t, []int32{0, 0, 5, 0}, read("counts"), "cell-major counts, complete past the cap")

	args := io.H5Ref[float64]{Filename: fn, Dataset: "/REPORTS/Widget/args"}
	loadedArgs, err := args.Load()
	require.NoError(t, err)
	assert.Equal(t, []int{3, sim.MaxReportArgs}, loadedArgs.Shape())
	assert.Equal(t, 99.0, loadedArgs.Unroll()[1])

	// The catalogue travels with the data, so a results file can be read
	// without the engine build that produced it.
	names := io.H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/Widget/names"}
	loadedNames, err := names.LoadText()
	require.NoError(t, err)
	assert.Equal(t, []string{"converged", "broken"}, loadedNames)

	messages := io.H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/Widget/messages"}
	loadedMessages, err := messages.LoadText()
	require.NoError(t, err)
	assert.Equal(t, "did not converge: {0} vs {1}", loadedMessages[0])

	// And the signature stamp, so a reader can detect a mismatched model file.
	sig := io.H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/signature"}
	loadedSig, err := sig.LoadText()
	require.NoError(t, err)
	require.Len(t, loadedSig, 1)
	assert.NotEmpty(t, loadedSig[0])
}

// A model that reported nothing still gets its counts written, so a reader can
// tell "checked, all clear" from "never ran".
func TestWriteReportsForASilentModel(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "silent.h5")

	sink := sim.NewReportSink(2, 2, sim.SevWarning, 3)
	mr := refWithReports("Widget", sink)
	require.NoError(t, writeReports(map[string]*modelReference{"Widget": mr}, []string{"Widget"}, fn))

	counts := io.H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/Widget/counts"}
	loaded, err := counts.Load()
	require.NoError(t, err)
	assert.Equal(t, []int32{0, 0, 0, 0}, loaded.Unroll())

	detail := io.H5Ref[int32]{Filename: fn, Dataset: "/REPORTS/Widget/run"}
	assert.False(t, detail.Exists(), "no detail records means no detail datasets")
}

func TestWriteReportsSkippedWithoutAnOutputFile(t *testing.T) {
	sink := sim.NewReportSink(1, 2, sim.SevWarning, 3)
	sink.Cell(0).Warn2(0, 1, 1, 2)
	mr := refWithReports("Widget", sink)

	// Running without an output file is legal; reports are summarised to the
	// log but there is nowhere to write them.
	assert.NoError(t, writeReports(map[string]*modelReference{"Widget": mr}, []string{"Widget"}, ""))
}

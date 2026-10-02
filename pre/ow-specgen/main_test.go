package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/flowmatters/openwater-core/data"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/pre/ow-specgen/testdata/reportingprobe"
	"github.com/flowmatters/openwater-core/sim"
)

// specFrom parses an OW-SPEC body the way processFile does, so these tests
// exercise the real YAML path rather than hand-built structs.
func specFrom(t *testing.T, body string) ModelSpec {
	t.Helper()
	desc := make(ModelSpecs)
	require.NoError(t, yaml.Unmarshal([]byte(body), &desc))
	require.Len(t, desc, 1)
	for name, spec := range desc {
		spec.Name = name
		return spec
	}
	return ModelSpec{}
}

const minimalImplementation = `
  implementation:
    function: kernel
    type: scalar
    lang: go
    outputs: params
`

func TestParseReports(t *testing.T) {
	spec := specFrom(t, `
Widget:`+minimalImplementation+`
  reports:
    started: '[warning] kernel started'
    aboveThreshold: '[warning] inflow {0} exceeds threshold {1}'
    broken: '[error] value {0} out of range [{1}, {2}]'
`)

	reports, err := parseReports(spec)
	require.NoError(t, err)
	require.Len(t, reports, 3)

	// Codes are assigned in declaration order -- yaml.MapSlice preserves it,
	// which matters because the code is what gets written to the results file.
	assert.Equal(t, []int{0, 1, 2}, []int{reports[0].Code, reports[1].Code, reports[2].Code})
	assert.Equal(t, "started", reports[0].Name)
	assert.Equal(t, "WidgetReportStarted", reports[0].ConstName)
	assert.Equal(t, "sim.SevWarning", reports[0].Severity)
	assert.Equal(t, "kernel started", reports[0].Template)
	assert.Equal(t, 0, reports[0].NArgs)

	assert.Equal(t, "WidgetReportAboveThreshold", reports[1].ConstName)
	assert.Equal(t, "inflow {0} exceeds threshold {1}", reports[1].Template)
	assert.Equal(t, 2, reports[1].NArgs)

	assert.Equal(t, "sim.SevError", reports[2].Severity)
	assert.Equal(t, 3, reports[2].NArgs, "argument count comes from the highest placeholder")
}

func TestParseReportsAbsent(t *testing.T) {
	reports, err := parseReports(specFrom(t, "Widget:"+minimalImplementation))
	assert.NoError(t, err)
	assert.Empty(t, reports, "a model without a reports: block opts out entirely")
}

// The message templates are free-form English, which makes them by far the most
// parse-hazardous thing in an OW-SPEC block. These are the mistakes a modeller
// will actually make.
func TestParseReportsRejectsBadSpecs(t *testing.T) {
	cases := map[string]struct {
		body    string
		wantErr string
	}{
		"no severity": {
			body:    "    badConfig: 'LVA table invalid'",
			wantErr: "bracketed severity",
		},
		"unknown severity": {
			body:    "    chatty: '[info] just so you know'",
			wantErr: "unknown report severity",
		},
		"severity but no message": {
			body:    "    empty: '[warning]   '",
			wantErr: "no message",
		},
		"placeholder beyond the argument limit": {
			body:    "    tooMany: '[warning] {0} {1} {2} {3}'",
			wantErr: "at most 3 arguments",
		},
		"name that is not an identifier": {
			body:    "    'bad name': '[warning] whatever'",
			wantErr: "Go-style identifier",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			spec := specFrom(t, "Widget:"+minimalImplementation+"  reports:\n"+tc.body+"\n")
			_, err := parseReports(spec)
			require.Error(t, err, "expected %q to be rejected", tc.body)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// Report messages are free-form English, and the two mistakes below are not
// rejected by parseReports at all -- they are invalid YAML, so the spec block
// never parses. That is only safe because processFile now propagates the error
// and main exits non-zero; before that, the build silently kept a stale wrapper.
func TestUnquotedReportMessagesFailTheBuild(t *testing.T) {
	cases := map[string]string{
		"message containing a colon":    "    badConfig: error: LVA table invalid",
		"message starting with a brace": "    badConfig: {0} points seen",
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			spec := "Widget:" + minimalImplementation + "  reports:\n" + body + "\n"

			// Straight YAML: these do not parse.
			desc := make(ModelSpecs)
			assert.Error(t, yaml.Unmarshal([]byte(spec), &desc),
				"an unquoted report message should be invalid YAML")

			// And the whole file therefore fails to process, rather than
			// quietly generating nothing.
			dir := t.TempDir()
			fn := dir + "/widget.go"
			source := "package widget\n\n/*OW-SPEC\n" + spec + "*/\n"
			require.NoError(t, os.WriteFile(fn, []byte(source), 0o644))

			assert.Error(t, processFile(fn),
				"a malformed OW-SPEC block must fail the build")
			_, statErr := os.Stat(dir + "/generated_Widget.go")
			assert.True(t, os.IsNotExist(statErr), "no wrapper should have been written")
		})
	}
}

func TestParseReportsRequiresAnImplementation(t *testing.T) {
	spec := specFrom(t, `
Widget:
  reports:
    started: '[warning] kernel started'
`)
	_, err := parseReports(spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no implementation block")
}

// A duplicate key is dropped by YAML itself before parseReports sees it, so
// this checks the surviving behaviour rather than an error.
func TestParseReportsDuplicateNames(t *testing.T) {
	spec := ModelSpec{
		Name:           "Widget",
		Implementation: yaml.MapSlice{{Key: "function", Value: "kernel"}},
		Reports: yaml.MapSlice{
			{Key: "same", Value: "[warning] first"},
			{Key: "same", Value: "[warning] second"},
		},
	}
	_, err := parseReports(spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate report")
}

// A bad reports: block must fail the build rather than leaving a stale wrapper,
// which is what phase 0 made possible.
func TestBadReportsBlockFailsGeneration(t *testing.T) {
	spec := specFrom(t, "Widget:"+minimalImplementation+`
  reports:
    chatty: '[info] just so you know'
`)
	spec.Filename = t.TempDir() + "/widget.go"

	err := generateWrapper(spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown report severity")

	out, statErr := exec.Command("ls", spec.Filename).CombinedOutput()
	assert.Error(t, statErr, "no wrapper should have been written: %s", out)
}

func TestGeneratedCatalogMatchesTheSpec(t *testing.T) {
	model := sim.Catalog["ReportingProbe"]
	require.NotNil(t, model, "the fixture model should have registered itself")

	reporting, ok := model().(sim.Reporting)
	require.True(t, ok, "a model with a reports: block must implement sim.Reporting")

	catalog := reporting.ReportCatalog()
	require.Len(t, catalog, 3)
	assert.Equal(t, "started", catalog[0].Name)
	assert.Equal(t, sim.SevWarning, catalog[0].Severity)
	assert.Equal(t, 0, catalog[0].NArgs)
	assert.Equal(t, "aboveThreshold", catalog[1].Name)
	assert.Equal(t, 2, catalog[1].NArgs)
	assert.Equal(t, sim.SevError, catalog[2].Severity)

	// The generated constants must line up with catalogue positions, since the
	// code is what identifies a report in the results file.
	assert.EqualValues(t, 0, reportingprobe.ReportingProbeReportStarted)
	assert.EqualValues(t, 1, reportingprobe.ReportingProbeReportAboveThreshold)
	assert.EqualValues(t, 2, reportingprobe.ReportingProbeReportNegativeInflow)
}

func TestNonReportingModelsDoNotImplementReporting(t *testing.T) {
	// Opting out has to mean opting out. Most models declare no reports: block
	// and must be entirely untouched by the reporting machinery.
	for _, name := range []string{"Simhyd", "GR4J", "EmcDwc", "Muskingum"} {
		factory := sim.Catalog[name]
		require.NotNil(t, factory, "%s should be registered", name)
		_, ok := factory().(sim.Reporting)
		assert.False(t, ok, "%s declares no reports: block, so it must not implement sim.Reporting", name)
	}

	// And opting in has to mean opting in.
	for _, name := range []string{"Storage", "StorageRouting"} {
		factory := sim.Catalog[name]
		require.NotNil(t, factory, "%s should be registered", name)
		_, ok := factory().(sim.Reporting)
		assert.True(t, ok, "%s declares a reports: block, so it must implement sim.Reporting", name)
	}
}

// End to end: run the generated wrapper and check reports come back attributed
// to the right cell and timestep.
func TestGeneratedWrapperRoutesReportsPerCell(t *testing.T) {
	const (
		cells     = 3
		timesteps = 6
	)

	model := sim.Catalog["ReportingProbe"]()

	params := data.NewArray2D[float64](1, cells)
	for c := 0; c < cells; c++ {
		params.Set2(0, c, 10) // threshold
	}
	model.ApplyParameters(params)

	inputs := data.NewArray3D[float64](cells, 1, timesteps)
	for ts := 0; ts < timesteps; ts++ {
		inputs.Set3(0, 0, ts, 1)   // cell 0: always under the threshold
		inputs.Set3(1, 0, ts, 100) // cell 1: always over it
		inputs.Set3(2, 0, ts, -1)  // cell 2: always negative
	}

	states := model.InitialiseStates(cells)
	outputs := sim.InitialiseOutputs(model, timesteps, cells)

	sink := sim.NewReportSink(cells, len(model.(sim.Reporting).ReportCatalog()), sim.SevWarning, 100)
	sink.BaseIndex = 40 // this generation starts at global run 40
	require.True(t, sim.AttachReportSink(model, sink))

	model.Run(inputs, states, outputs)
	rs := sink.Flush()

	counts := func(cell, code int) uint32 { return rs.Counts[cell*rs.NCodes+code] }

	// Every cell reports "started" exactly once.
	for c := 0; c < cells; c++ {
		assert.EqualValues(t, 1, counts(c, 0), "cell %d started", c)
	}
	assert.EqualValues(t, 0, counts(0, 1), "cell 0 never crosses the threshold")
	assert.EqualValues(t, timesteps, counts(1, 1), "cell 1 always crosses it")
	assert.EqualValues(t, timesteps, counts(2, 2), "cell 2 is always negative")

	// Run indices must be global, not per generation.
	for _, run := range rs.Runs {
		assert.GreaterOrEqual(t, run, int32(40))
		assert.Less(t, run, int32(40+cells))
	}

	// And the reports must render against the catalogue.
	catalog := model.(sim.Reporting).ReportCatalog()
	var rendered []string
	for i := 0; i < rs.Len(); i++ {
		if rs.Runs[i] == 41 && rs.Codes[i] == 1 && rs.Timesteps[i] == 0 {
			rendered = append(rendered, catalog[1].Render(sim.Report{
				NArgs: 2,
				Args:  [sim.MaxReportArgs]float64{rs.Args[i*3], rs.Args[i*3+1], rs.Args[i*3+2]},
			}))
		}
	}
	require.Len(t, rendered, 1)
	assert.Equal(t, "inflow 100 exceeds threshold 10", rendered[0])
}

// Reporting off must stay off: a nil sink leaves the model running normally.
func TestGeneratedWrapperRunsWithoutASink(t *testing.T) {
	model := sim.Catalog["ReportingProbe"]()
	params := data.NewArray2D[float64](1, 1)
	params.Set2(0, 0, 10)
	model.ApplyParameters(params)

	inputs := data.NewArray3D[float64](1, 1, 4)
	for ts := 0; ts < 4; ts++ {
		inputs.Set3(0, 0, ts, 100)
	}
	states := model.InitialiseStates(1)
	outputs := sim.InitialiseOutputs(model, 4, 1)

	assert.NotPanics(t, func() { model.Run(inputs, states, outputs) },
		"a model with reports must run with no sink attached")

	var total float64
	for ts := 0; ts < 4; ts++ {
		total += outputs.Get3(0, 0, ts)
	}
	assert.Equal(t, 400.0, total, fmt.Sprintf("outputs: %v", outputs.Unroll()))
}

package main

import (
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/sim"
)

var verbose bool
var quiet bool

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var overwrite = flag.Bool("overwrite", false, "overwrite existing output files")
var outputsFor = flag.String("outputs-for", "", "only write model outputs for specified models. Specify as command separated list of model names")
var inputsFor = flag.String("inputs-for", "", "only write final model inputs for specified models. Specify as command separated list of model names")
var noOutputsFor = flag.String("no-outputs-for", "", "do not write model outputs for specified models. Specify as command separated list of model names")
var noInputsFor = flag.String("no-inputs-for", "", "do not write final model inputs for specified models. Specify as command separated list of model names")
var noOutputs = flag.Bool("no-outputs", false, "do not write model outputs for any models")
var noInputs = flag.Bool("no-inputs", false, "do not write final model inputs for any models")
var onlyOutputsFor = flag.String("only-outputs-for", "", "only write model outputs for the specified models (strict whitelist). Specify as comma separated list of model names")
var onlyInputsFor = flag.String("only-inputs-for", "", "only write final model inputs for the specified models (strict whitelist). Specify as comma separated list of model names")

var parameterInputFile = flag.String("parameters", "", "specify file for model parameters")
var statesInputFile = flag.String("initial-states", "", "specify file for initial states")
var timeseriesInputFile = flag.String("input-timeseries", "", "specify file for input timeseries")
var statesOutputFile = flag.String("final-states", "", "specify file for final states")

var splitOutputs = flag.String("outputs", "", "split output files by model type. Specify as <model>:<fn>,<model>:<fn>,...")
var writerMode = flag.Bool("writer", false, "operate as an output writer for another simulation process")
var showVersion = flag.Bool("version", false, "display version information")
var linkWorkers = flag.Int("link-workers", 0, "number of worker goroutines for parallel link processing; 0 = auto (min(NumCPU, 8))")
var compressOutputs = flag.Int("compress-outputs", 0, "compress output datasets with deflate (gzip) at the given level (1=fastest, 9=best, 0=off). Reduces output file size at some CPU cost")
var maxWriteAhead = flag.Int("max-write-ahead", 4, "max generations the simulation can run ahead of the output writer. Lower values reduce peak memory; higher values allow more overlap. 0 = unlimited")

func init() {
	const usage = "show progress of simulation generations"
	flag.BoolVar(&verbose, "verbose", false, usage)
	flag.BoolVar(&verbose, "v", false, usage+" (shorthand)")
	const usageQuiet = "Only log errors and warnings"
	flag.BoolVar(&quiet, "quiet", false, usageQuiet)
	flag.BoolVar(&quiet, "q", false, usageQuiet+" (shorthand)")
}

// Model reporting. Models opt in via a reports: block in their OW-SPEC; see
// docs/MODEL_REPORTING.md. Reporting is on by default because the disabled
// path costs ~0.1ns per call and a silent run is the current problem.
var reportsFlag = flag.String("reports", "warning", "collect model reports at this severity and above: off, warning or error")
var reportDetailLimit = flag.Int("report-detail-limit", 3, "detail records kept per model, cell and report code. Counts are always complete; 0 = keep every record")
var reportsFatal = flag.String("reports-fatal", "error", "exit non-zero when a report of this severity or worse was raised: off or error")

func reportSeverity() sim.Severity {
	switch *reportsFlag {
	case "off":
		return sim.SevNone
	case "warning":
		return sim.SevWarning
	case "error":
		return sim.SevError
	}
	log.Fatal().Str("Value", *reportsFlag).Msg("Invalid -reports value; expected off, warning or error")
	return sim.SevNone
}

// reportsFatalSeverity is the severity at which a raised report makes ow-sim
// exit non-zero. Without it, converting a model's log.Panic() sites into
// reports would turn a loud crash into a wrong-but-complete run with exit
// code 0, which is how batch and CI runs currently detect breakage.
func reportsFatalSeverity() sim.Severity {
	switch *reportsFatal {
	case "off":
		return sim.SevNone
	case "error":
		return sim.SevError
	}
	log.Fatal().Str("Value", *reportsFatal).Msg("Invalid -reports-fatal value; expected off or error")
	return sim.SevNone
}

package main

import (
	"flag"
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

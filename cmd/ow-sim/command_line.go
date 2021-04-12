package main

import (
	"flag"
)

var verbose bool

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var overwrite = flag.Bool("overwrite", false, "overwrite existing output files")
var outputsFor = flag.String("outputs-for", "", "only write model outputs for specified models. Specify as command separated list of model names")
var inputsFor = flag.String("inputs-for", "", "only write final model inputs for specified models. Specify as command separated list of model names")
var noOutputsFor = flag.String("no-outputs-for", "", "do not write model outputs for specified models. Specify as command separated list of model names")
var noInputsFor = flag.String("no-inputs-for", "", "do not write final model inputs for specified models. Specify as command separated list of model names")

var splitOutputs = flag.String("outputs", "", "split output files by model type. Specify as <model>:<fn>,<model>:<fn>,...")
var writerMode = flag.Bool("writer", false, "operate as an output writer for another simulation process")

func init() {
	const usage = "show progress of simulation generations"
	flag.BoolVar(&verbose, "verbose", false, usage)
	flag.BoolVar(&verbose, "v", false, usage+" (shorthand)")
}

package main

import (
	"flag"
)

var verbose bool

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var overwrite = flag.Bool("overwrite", false, "overwrite existing output files")
var splitOutputs = flag.String("outputs","","split output files by model type. Specify as <model>:<fn>,<model>:<fn>,...")
var writerMode = flag.Bool("writer",false,"operate as an output writer for another simulation process")

func init() {
	const usage = "show progress of simulation generations"
	flag.BoolVar(&verbose, "verbose", false, usage)
	flag.BoolVar(&verbose, "v", false, usage+" (shorthand)")
}


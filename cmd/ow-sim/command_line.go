package main

import (
	"flag"
)

var verbose bool

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var overwrite = flag.Bool("overwrite", false, "overwrite existing output files")
//var split = flag.Bool("split",false,"split output files by model type")

func init() {
	const usage = "show progress of simulation generations"
	flag.BoolVar(&verbose, "verbose", false, usage)
	flag.BoolVar(&verbose, "v", false, usage+" (shorthand)")
}


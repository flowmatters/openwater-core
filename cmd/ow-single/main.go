package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/util"
)

var showVersion = flag.Bool("version", false, "display version information")

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("ow-single %s\n", util.FullVersion())
		fmt.Printf("Signature: %s\n", util.GetSignatureHash())
		fmt.Printf("Build: %s (%s)\n", util.BuildTime, util.BuildSHA)
		os.Exit(0)
	}

	sim.RunSingleModelJSON(os.Stdin, os.Stdout, true)
}

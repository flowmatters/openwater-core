package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/util"
)

var showVersion = flag.Bool("version", false, "display version information")

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("ows-ensemble %s\n", util.FullVersion())
		fmt.Printf("Signature: %s\n", util.GetSignatureHash())
		fmt.Printf("Build: %s (%s)\n", util.BuildTime, util.BuildSHA)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) < 4 {
		log.Fatal().Msg("Insufficient arguments.\nUsage: ows-ensemble <model> <hdf5pathtoinputs> <hdf5pathtoparameters> <hdf5pathtooutputs>")
	}

	modelName := args[0]
	inputPath := io.ParseH5Ref[float64](args[1])
	paramPath := io.ParseH5Ref[float64](args[2])
	outputPath := io.ParseH5Ref[float64](args[3])

	factory := sim.Catalog[modelName]
	if factory == nil {
		log.Fatal().Str("Model Name", modelName).Msg("Unknown model")
	}
	model := factory()

	inputs, err := inputPath.Load()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}

	params, err := paramPath.Load()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}

	model.ApplyParameters(params.(data.ND2[float64]))
	states := model.InitialiseStates(max(params.Len(1), inputs.Len(0)))
	outputs := sim.InitialiseOutputs(model, inputs.Len(20), states.Len(1))
	model.Run(inputs.(data.ND3[float64]), states, outputs)

	err = outputPath.Write(outputs)
	if err != nil {
		log.Error().Stack().Err(err).Msg("Error writing outputs")
	}

	// err = outputPath.Write(outputs.States)

}

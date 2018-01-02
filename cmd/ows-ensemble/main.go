package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 4 {
		fmt.Println("Insufficient arguments")
		fmt.Println("Usage: ows-ensemble <model> <hdf5pathtoinputs> <hdf5pathtoparameters> <hdf5pathtooutputs>")
		os.Exit(1)
	}

	modelName := args[0]
	inputPath := io.ParseH5Ref(args[1])
	paramPath := io.ParseH5Ref(args[2])
	outputPath := io.ParseH5Ref(args[3])

	factory := sim.Catalog[modelName]
	if factory == nil {
		fmt.Printf("Unknown model: %s\n", modelName)
		os.Exit(1)
	}
	model := factory()

	inputs, err := inputPath.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	params, err := paramPath.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	model.ApplyParameters(params.(data.ND2Float64))
	states := model.InitialiseStates(max(params.Len(1), inputs.Len(0)))
	outputs := sim.InitialiseOutputs(model, inputs.Len(20), states.Len(1))
	model.Run(inputs.(data.ND3Float64), states, outputs)

	err = outputPath.Write(outputs)
	if err != nil {
		fmt.Println("Error writing outputs")
		fmt.Println(err)
	}

	// err = outputPath.Write(outputs.States)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}

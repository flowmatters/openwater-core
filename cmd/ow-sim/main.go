package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
)

type modelReference struct {
	Filename  string
	ModelName string
	Batches   []int32
}

func initModel(fn, model string) (*modelReference, error) {
	modelRef := io.H5RefInt32{Filename: fn, Dataset: "/MODELS/" + model + "/batches"}
	batchesArray, err := modelRef.Load()
	if err != nil {
		return nil, err
	}
	batches := batchesArray.Unroll()
	result := modelReference{Filename: fn, ModelName: model, Batches: batches}
	return &result, nil
}

func loadGeneration(modelType string) {
	// Needs to be a member of the type...
	// For given model, load inputs, parameters, initial states for a given generation
}

func main() {
	flag.Parse()
	args := flag.Args()
	fn := args[0]
	modelsRef := io.H5RefFloat64{Filename: fn, Dataset: "/META/models"}
	dimsRef := io.H5RefFloat64{Filename: fn, Dataset: "/DIMENSIONS"}
	//	procRef := io.H5Ref{Filename: fn, Dataset: "/PROCESSES"}

	models, err := modelsRef.LoadText()
	if err != nil {
		fmt.Println("Couldn't read model metadata")
		os.Exit(1)
	}
	fmt.Println("Models", models)

	dims, err := dimsRef.GetDatasets()
	if err != nil {
		fmt.Println("Couldn't read model dimensions")
		os.Exit(1)
	}
	fmt.Println("Dimensions", dims)

	aModel, err := initModel(fn, models[0])
	if err != nil {
		fmt.Println("Couldn't initialise model", models[0])
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Processes", processes)

}

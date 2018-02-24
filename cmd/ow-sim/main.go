package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
)

func main() {
	flag.Parse()
	args := flag.Args()
	fn := args[0]
	modelsRef := io.H5Ref{Filename: fn, Dataset: "/META/models"}
	dimsRef := io.H5Ref{Filename: fn, Dataset: "/DIMENSIONS"}
	procRef := io.H5Ref{Filename: fn, Dataset: "/PROCESSES"}

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

	processes, err := procRef.GetDatasets()
	if err != nil {
		fmt.Println("Couldn't read model processes")
		os.Exit(1)
	}
	fmt.Println("Processes", processes)

}

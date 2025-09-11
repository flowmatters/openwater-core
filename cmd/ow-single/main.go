package main

import (
	"os"

	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
)

func main() {
	sim.RunSingleModelJSON(os.Stdin, os.Stdout, true)
}

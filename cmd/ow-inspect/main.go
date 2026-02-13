package main

import (
	"encoding/json"
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
		fmt.Printf("ow-inspect %s\n", util.FullVersion())
		fmt.Printf("Signature: %s\n", util.GetSignatureHash())
		fmt.Printf("Build: %s (%s)\n", util.BuildTime, util.BuildSHA)
		os.Exit(0)
	}

	args := flag.Args()

	var models []string
	if len(args) == 0 {
		models = make([]string, len(sim.Catalog))
		i := 0
		for k := range sim.Catalog {
			models[i] = k
			i++
		}
	} else {
		models = args
	}

	allModels := make(map[string]sim.ModelDescription)
	for _, modelName := range models {
		model := sim.Catalog[modelName]()
		allModels[modelName] = model.Description()
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	encoder.Encode(allModels)
}

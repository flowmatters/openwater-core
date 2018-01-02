package main

import (
	"flag"
  _ "github.com/flowmatters/openwater-core/models"
  "github.com/flowmatters/openwater-core/sim"
	"os"
	"encoding/json"
)

func main(){
//	jsonFormat := flag.Bool("json",false,"Output in JSON format")

  flag.Parse()
  args := flag.Args()

	var models []string
  if len(args)==0 {
		models = make([]string,len(sim.Catalog))
		i := 0
		for k := range sim.Catalog {
			models[i] = k
			i++
		}
  } else {
		models = args
	}

	allModels := make(map[string]sim.ModelDescription)
	for _,modelName := range models {
		model := sim.Catalog[modelName]();
		allModels[modelName] = model.Description();
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent(""," ")
	encoder.Encode(allModels)
}

package main

import (
	"fmt"
	"time"
	"os"
)


func runGeneration(i int, models map[string]*modelReference, modelNames []string) float64 {
	genTotal := 0
	fmt.Printf("==== Generation %d ====\n", i)
	simulationDone := make(chan string)
	genStart := time.Now()
	modelCount := 0

	for _, modelName := range modelNames {
		gen, err := models[modelName].GetGeneration(i)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if gen.Count == 0 {
			continue
		}
		verbosePrintf("* %d x %s\n", gen.Count, modelName)
		modelCount++

		genTotal += gen.Count
		go func(g *modelGeneration, name string) {
			if g.Count > 0 {
				g.Run()
				outputs := g.Outputs
				if outputs == nil {
					fmt.Printf("No outputs from %s in generation %d\n", name, i)
				}
			}

			simulationDone <- name
		}(gen, modelName)
	}

	for i := 0; i < modelCount; i++ {
		mn := <-simulationDone
		verbosePrintf("%d: %s finished\n", i, mn)
	}

	genSimulationEnd := time.Now()
	genSimulationElapsed := genSimulationEnd.Sub(genStart).Seconds()
	verbosePrintf("= %d runs in %f seconds\n", genTotal, genSimulationElapsed)

	return genSimulationElapsed
}

package main

import (
	"fmt"
	"os"
	"time"
)

func runGeneration(i int, models map[string]*modelReference, modelNames []string) float64 {
	genTotal := 0
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
				simulationDone <- name
			} else {
				simulationDone <- ""
			}
		}(gen, modelName)
	}

	for i := 0; i < modelCount; i++ {
		mn := <-simulationDone
		if mn != "" {
			verbosePrintf("%d: %s finished\n", i, mn)
		}
	}

	genSimulationEnd := time.Now()
	genSimulationElapsed := genSimulationEnd.Sub(genStart).Seconds()
	verbosePrintf("= %d runs in %f seconds\n", genTotal, genSimulationElapsed)

	return genSimulationElapsed
}

func writeGeneration(g int, models map[string]*modelReference, modelNames []string) {
	genWriteStart := time.Now()
	verbosePrintf("Writing results for generation %d...\n", g)
	for _, modelName := range modelNames {
		modelRef := models[modelName]
		err := modelRef.WriteData(g)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	genWriteEnd := time.Now()
	genWriteElapsed := genWriteEnd.Sub(genWriteStart)
	verbosePrintf("Results for generation %d written in %f seconds\n", g, genWriteElapsed.Seconds())
}

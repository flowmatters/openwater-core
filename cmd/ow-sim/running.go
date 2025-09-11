package main

import (
	"time"

	"github.com/rs/zerolog/log"
)

func runGeneration(i int, models map[string]*modelReference, modelNames []string) (elapsedTime float64, nodesRun int) {
	nodesRun = 0
	simulationDone := make(chan string)
	genStart := time.Now()
	modelCount := 0

	for _, modelName := range modelNames {
		gen, err := models[modelName].GetGeneration(i)
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}
		if gen.Count == 0 {
			continue
		}
		log.Debug().Int("Count", gen.Count).Str("Model Name", modelName).Msg("Running model generation")
		modelCount++

		nodesRun += gen.Count
		go func(g *modelGeneration, name string) {
			if g.Count > 0 {
				g.Run()
				outputs := g.Outputs
				if outputs == nil {
					log.Info().Str("Model Name", name).Int("Generation", i).Msg("No outputs from model in generation")
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
			log.Debug().Int("Index", i).Str("Model Name", mn).Msg("Model finished")
		}
	}

	genSimulationEnd := time.Now()
	elapsedTime = genSimulationEnd.Sub(genStart).Seconds()
	log.Debug().Int("Runs", nodesRun).Float64("Elapsed Seconds", elapsedTime).Msg("Generation runs completed")
	return
}

func writeGeneration(g int, models map[string]*modelReference, modelNames []string) {
	genWriteStart := time.Now()
	log.Debug().Int("Generation", g).Msg("Writing results for generation")
	for _, modelName := range modelNames {
		modelRef := models[modelName]
		err := modelRef.WriteData(g)
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}
	}
	genWriteEnd := time.Now()
	genWriteElapsed := genWriteEnd.Sub(genWriteStart)
	log.Debug().Int("Generation", g).Float64("Elapsed Seconds", genWriteElapsed.Seconds()).Msg("Results written for generation")
}

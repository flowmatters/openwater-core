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
		log.Debug().Msgf("* %d x %s", gen.Count, modelName)
		modelCount++

		nodesRun += gen.Count
		go func(g *modelGeneration, name string) {
			if g.Count > 0 {
				g.Run()
				outputs := g.Outputs
				if outputs == nil {
					log.Info().Msgf("No outputs from %s in generation %d", name, i)
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
			log.Debug().Msgf("%d: %s finished", i, mn)
		}
	}

	genSimulationEnd := time.Now()
	elapsedTime = genSimulationEnd.Sub(genStart).Seconds()
	log.Debug().Msgf("= %d runs in %f seconds", nodesRun, elapsedTime)
	return
}

func writeGeneration(g int, models map[string]*modelReference, modelNames []string) {
	genWriteStart := time.Now()
	log.Debug().Msgf("Writing results for generation %d...", g)
	for _, modelName := range modelNames {
		modelRef := models[modelName]
		err := modelRef.WriteData(g)
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}
	}
	genWriteEnd := time.Now()
	genWriteElapsed := genWriteEnd.Sub(genWriteStart)
	log.Debug().Msgf("Results for generation %d written in %f seconds", g, genWriteElapsed.Seconds())
}

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/rs/zerolog/log"
	"gonum.org/v1/hdf5"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/util"
)

const (
	LINK_SRC_GENERATION  = 0
	LINK_SRC_MODEL       = 1
	LINK_SRC_NODE        = 2
	LINK_SRC_GEN_NODE    = 3
	LINK_SRC_VAR         = 4
	LINK_DEST_GENERATION = 5
	LINK_DEST_MODEL      = 6
	LINK_DEST_NODE       = 7
	LINK_DEST_GEN_NODE   = 8
	LINK_DEST_VAR        = 9
)

func main() {
	flag.Parse()
	setupLogger()

	if *showVersion {
		fmt.Printf("ow-sim %s\n", util.FullVersion())
		fmt.Printf("Signature: %s\n", util.GetSignatureHash())
		fmt.Printf("Build: %s (%s)\n", util.BuildTime, util.BuildSHA)
		os.Exit(0)
	}

	hdf5.DisplayErrors(false)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()

	if *writerMode {
		run_writer(args)
	} else {
		run_simulation(args)
	}
}

func run_simulation(args []string) {
	if len(args) == 0 {
		log.Fatal().Msg("Must specify an input file")
	}
	fn := args[0]
	var outputFn string = ""
	if len(args) > 1 {
		outputFn = args[1]

		if _, err := os.Stat(outputFn); err == nil {
			if *overwrite {
				os.Remove(outputFn)
			} else {
				log.Fatal().Str("Output Filename", outputFn).Msg("Output file exists and overwrite not set. Delete file or use -overwrite")
			}
		}
	}

	modelsRef := io.H5Ref[float64]{Filename: fn, Dataset: "/META/models"}
	dimsRef := io.H5Ref[float64]{Filename: fn, Dataset: "/DIMENSIONS"}
	//	procRef := io.H5Ref{Filename: fn, Dataset: "/PROCESSES"}

	modelNames, err := modelsRef.LoadText()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("Couldn't read model metadata")
	}
	log.Debug().Any("Models", modelNames).Msg("")

	dims, err := dimsRef.GetDatasets()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("Couldn't read model dimensions")
	}
	log.Debug().Any("Dimensions", dims).Msg("")

	linksRef := io.H5Ref[uint32]{Filename: fn, Dataset: "/LINKS"}
	linksND, err := linksRef.Load()
	links := linksND.(data.ND2[uint32])
	linkSliceDim := []int{1, LINK_DEST_VAR + 1}
	linkSliceStep := []int{1, 1}
	nLinks := links.Len(0)
	nextLink := 0
	simStart := time.Now()

	totalTimeSimulation := 0.0
	totalTimeFinalWrite := 0.0
	totalTimeLinks := 0.0

	var genCount int
	writingDone := make(chan int)

	models, genCount, nodeCount := makeModelRefs(modelNames, fn, outputFn)
	nodesCompleted := 0

	for i := 0; i < genCount; i++ {
		pcComplete := 100.0 * float64(nodesCompleted) / float64(nodeCount)
		log.Info().Float64("Percent Complete", pcComplete).Int("Generation", i+1).Int("Total Generations", genCount).Msg("Generation progress")

		// === RUN GENERATION ===
		genSimulationTime, nodesInGeneration := runGeneration(i, models, modelNames) // synchronous
		nodesCompleted += nodesInGeneration
		totalTimeSimulation += genSimulationTime
		// === /RUN GENERATION ===

		// === WRITE GENERATION OUTPUTS ===
		// asynchronous
		if outputFn != "" {
			go func(g int) {
				if g > 0 {
					prevG := -1
					for {
						prevG = <-writingDone

						for _, modelName := range modelNames {
							modelRef := models[modelName]
							modelRef.PurgeGeneration(prevG)
						}

						if prevG == (g - 1) {
							break
						}
						log.Debug().Int("Expected Generation", g).Int("Current Generation", prevG).Msg("Waiting for generation to finish (sleeping)")
						writingDone <- prevG
						time.Sleep(time.Duration(1000 * 1000 * 500)) // Half a second
					}
				}

				writeGeneration(g, models, modelNames)
				writingDone <- g
			}(i)
		}
		// === /WRITE GENERATION OUTPUTS ===

		// === PROCESS LINKS ===
		// synchronous
		genLinkStart := time.Now()
		currentLink := nextLink
		for {
			if nextLink >= nLinks {
				break
			}

			linkND := links.Slice([]int{nextLink, 0}, linkSliceDim, linkSliceStep)
			link := linkND.(data.ND1[uint32])
			linkGen := link.Get1(LINK_SRC_GENERATION)

			if linkGen > uint32(i) {
				break
			}

			// Copy data from output to input...
			srcModelNumber := link.Get1(LINK_SRC_MODEL)
			srcModelName := modelNames[srcModelNumber]
			srcModel, _ := models[srcModelName].GetGeneration(int(linkGen))

			destGen := link.Get1(LINK_DEST_GENERATION)
			destModelNumber := link.Get1(LINK_DEST_MODEL)
			destModelName := modelNames[destModelNumber]
			destModel, err := models[destModelName].GetGeneration(int(destGen))
			if err != nil {
				log.Fatal().Stack().Err(err).Msg("")
			}

			nTimesteps := srcModel.Outputs.Len(sim.DIMO_TIMESTEP)
			srcVar := link.Get1(LINK_SRC_VAR)
			if srcVar < 0 {
				continue
			}
			srcIdx := link.Get1(LINK_SRC_GEN_NODE)
			srcData := srcModel.Outputs.Slice([]int{int(srcIdx), int(srcVar), 0}, []int{1, 1, nTimesteps}, []int{1, 1, 1})

			destVar := link.Get1(LINK_DEST_VAR)
			if destVar < 0 {
				continue
			}
			destIdx := link.Get1(LINK_DEST_GEN_NODE)
			destData := destModel.Inputs.Slice([]int{int(destIdx), int(destVar), 0}, []int{1, 1, nTimesteps}, []int{1, 1, 1})

			data.AddToArray[float64](destData, srcData)
			nextLink++
		}
		genLinkEnd := time.Now()
		genLinkElapsed := genLinkEnd.Sub(genLinkStart).Seconds()
		totalTimeLinks += genLinkElapsed
		log.Debug().Int("Links Processed", nextLink-currentLink).Int("Start Link", currentLink).Int("End Link", nextLink).Float64("Elapsed Time (seconds)", genLinkElapsed).Msg("Processed links")
		// === /PROCESS LINKS ===

		genElapsed := genLinkElapsed + genSimulationTime
		log.Debug().Float64("Elapsed Seconds", genElapsed).Msg("Generation completed")
	}

	log.Debug().Msg("Simulation finished. Waiting for results to be written")
	generationsEnd := time.Now()

	if outputFn != "" {
		for {
			genFinished := <-writingDone
			if genFinished == (genCount - 1) {
				log.Debug().Int("Generation", genFinished).Msg("Generation Finished Writing")
				break
			}
			log.Debug().Int("Expected Final Generation", genCount-1).Int("Received Generation", genFinished).Msg("Waiting for final generation to finish (sleeping)")
			writingDone <- genFinished
			time.Sleep(time.Duration(500 * 1000 * 1000))
		}
	}

	simEnd := time.Now()
	finalWriteElapsed := simEnd.Sub(generationsEnd)
	totalTimeFinalWrite = finalWriteElapsed.Seconds()
	simElapsed := simEnd.Sub(simStart)
	log.Info().
		Float64("Total Run Time", simElapsed.Seconds()).
		Float64("Total Simulation Time", totalTimeSimulation).
		Float64("Total Link Time", totalTimeLinks).
		Float64("Total Final Write Time", totalTimeFinalWrite).
		Msg("Simulation complete")
}

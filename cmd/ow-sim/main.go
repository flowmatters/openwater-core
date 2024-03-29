package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
	"gonum.org/v1/hdf5"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
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
	hdf5.DisplayErrors(false)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
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

	fn := args[0]
	var outputFn string = ""
	if len(args) > 1 {
		outputFn = args[1]

		if _, err := os.Stat(outputFn); err == nil {
			if *overwrite {
				os.Remove(outputFn)
			} else {
				fmt.Printf("Output file (%s) exists and overwrite not set. Delete file or use -overwrite\n", outputFn)
				os.Exit(1)
			}
		}
	}

	modelsRef := io.H5RefFloat64{Filename: fn, Dataset: "/META/models"}
	dimsRef := io.H5RefFloat64{Filename: fn, Dataset: "/DIMENSIONS"}
	//	procRef := io.H5Ref{Filename: fn, Dataset: "/PROCESSES"}

	modelNames, err := modelsRef.LoadText()
	if err != nil {
		fmt.Println("Couldn't read model metadata")
		os.Exit(1)
	}
	verbosePrintln("Models", modelNames)

	dims, err := dimsRef.GetDatasets()
	if err != nil {
		fmt.Println("Couldn't read model dimensions")
		os.Exit(1)
	}
	verbosePrintln("Dimensions", dims)

	linksRef := io.H5RefUint32{Filename: fn, Dataset: "/LINKS"}
	linksND, err := linksRef.Load()
	links := linksND.(data.ND2Uint32)
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

	fmt.Println()
	for i := 0; i < genCount; i++ {
		pcComplete := 100.0 * float64(nodesCompleted) / float64(nodeCount)
		fmt.Printf("==== %.1f%% - Generation %d / %d ====\n", pcComplete, i+1, genCount)

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
						verbosePrintf("Waiting for generation %d, got generation %d, sleeping\n", g, prevG)
						writingDone <- prevG
						time.Sleep(time.Duration(1000 * 1000 * 500)) // Half a second
					}
				}

				writeGeneration(g, models, modelNames)
				writingDone <- g
			}(i)
		}
		// fmt.Printf("Results written in %f seconds\n", genWriteElapsed.Seconds())
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
			link := linkND.(data.ND1Uint32)
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
				fmt.Println(err)
				os.Exit(1)
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

			data.AddToFloat64Array(destData, srcData)
			nextLink++
		}
		genLinkEnd := time.Now()
		genLinkElapsed := genLinkEnd.Sub(genLinkStart).Seconds()
		totalTimeLinks += genLinkElapsed
		verbosePrintf("%d links (%d to %d), processed in %f seconds\n", nextLink-currentLink, currentLink, nextLink, genLinkElapsed)
		// === /PROCESS LINKS ===

		genElapsed := genLinkElapsed + genSimulationTime
		verbosePrintf("Generation completed in %f seconds\n", genElapsed)
		verbosePrintln()
	}

	verbosePrintln("Simulation finished. Waiting for results to be written")
	generationsEnd := time.Now()

	if outputFn != "" {
		for {
			genFinished := <-writingDone
			if genFinished == (genCount - 1) {
				verbosePrintf("Generation %d finished writing\n", genFinished)
				break
			}
			verbosePrintf("Waiting for final generation (%d), got generation %d, sleeping\n", genCount-1, genFinished)
			writingDone <- genFinished
			time.Sleep(time.Duration(500 * 1000 * 1000))
		}
	}

	simEnd := time.Now()
	finalWriteElapsed := simEnd.Sub(generationsEnd)
	totalTimeFinalWrite = finalWriteElapsed.Seconds()
	simElapsed := simEnd.Sub(simStart)
	fmt.Printf("Simulation completed in %f seconds\n", simElapsed.Seconds())
	fmt.Printf("Total Simulation Time: %f\n", totalTimeSimulation)
	fmt.Printf("Total Link Time: %f\n", totalTimeLinks)
	fmt.Printf("Total Final Write Time: %f\n", totalTimeFinalWrite)

	//	fmt.Println("Add stats")
}

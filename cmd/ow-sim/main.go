package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"time"

	"github.com/flowmatters/openwater-core/data"

	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
)

var verbose bool

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var overwrite = flag.Bool("overwrite", false, "overwrite existing output files")

func init() {
	const usage = "show progress of simulation generations"
	flag.BoolVar(&verbose, "verbose", false, usage)
	flag.BoolVar(&verbose, "v", false, usage+" (shorthand)")
}

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

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func prefix(msg string, e error) error {
	return &errorString{msg + e.Error()}
}

type modelGeneration struct {
	Model      sim.TimeSteppingModel
	Count      int
	Inputs     data.ND3Float64
	States     data.ND2Float64
	Parameters data.ND2Float64
	Outputs    data.ND3Float64
}

func (g *modelGeneration) Run() {
	if g.Count == 0 {
		return
	}

	g.Model.ApplyParameters(g.Parameters)
	g.Outputs = sim.InitialiseOutputs(g.Model, g.Inputs.Len(2), g.Inputs.Len(0))
	g.Model.Run(g.Inputs, g.States, g.Outputs)
}

type modelReference struct {
	Filename           string
	OutputFilename     string
	WriteInputs        bool
	WriteOutputs       bool
	ModelName          string
	Batches            []int32
	Generations        []*modelGeneration
	outputsInitialised bool
}

func initModel(fn, model string) (*modelReference, error) {
	modelRef := io.H5RefInt32{Filename: fn, Dataset: "/MODELS/" + model + "/batches"}
	batchesArray, err := modelRef.Load()
	if err != nil {
		return nil, err
	}
	batches := batchesArray.Unroll()
	result := modelReference{Filename: fn, ModelName: model, Batches: batches}
	result.Generations = make([]*modelGeneration, len(batches))
	return &result, nil
}

func (mr *modelReference) GetReference(genSlice []int, element string) io.H5RefFloat64 {
	result := io.H5RefFloat64{}
	result.Filename = mr.Filename
	result.Dataset = "/MODELS/" + mr.ModelName + "/" + element

	if element == "parameters" {
		result.Slice = [][]int{nil, genSlice}
	} else {
		result.Slice = [][]int{genSlice, nil}
	}

	if element == "inputs" {
		result.Slice = append(result.Slice, nil)
	}

	return result
}

func (mr *modelReference) GetGeneration(i int) (*modelGeneration, error) {
	if mr.Generations[i] == nil {
		gen := modelGeneration{}
		gen.Model = sim.Catalog[mr.ModelName]()
		mr.Generations[i] = &gen
		genSlice := []int{0, int(mr.Batches[i]), 1}
		if i > 0 {
			genSlice[0] = int(mr.Batches[i-1])
		}
		gen.Count = genSlice[1] - genSlice[0]

		if gen.Count == 0 {
			gen.Inputs = data.NewArray3DFloat64(0, 0, 0)
			gen.Parameters = data.NewArray2DFloat64(0, 0)
			gen.States = data.NewArray2DFloat64(0, 0)
			return mr.Generations[i], nil
		}

		inputRef := mr.GetReference(genSlice, "inputs")
		inputs, err := inputRef.Load()
		if err != nil {
			return nil, err
		}
		gen.Inputs = inputs.(data.ND3Float64)

		paramRef := mr.GetReference(genSlice, "parameters")
		parameters, err := paramRef.Load()
		if err != nil {
			return nil, err
		}
		gen.Parameters = parameters.(data.ND2Float64)

		stateRef := mr.GetReference(genSlice, "states")
		states, err := stateRef.Load()
		if err != nil {
			return nil, err
		}
		gen.States = states.(data.ND2Float64)
	}
	return mr.Generations[i], nil
}

func (mr *modelReference) PurgeGeneration(i int) {
	mr.Generations[i] = nil
}

func (mr *modelReference) TotalRuns() int {
	return int(mr.Batches[len(mr.Batches)-1])
}

func (mr *modelReference) initialiseDataset(label string, refShape []int) error {
	ref := io.H5RefFloat64{}
	ref.Filename = mr.OutputFilename
	ref.Dataset = "/MODELS/" + mr.ModelName + "/" + label
	count := mr.TotalRuns()
	return ref.Create([]int{count, refShape[1], refShape[2]}, math.NaN())
}

func (mr *modelReference) InitialiseOutputs(refGeneration int) error {
	gen, err := mr.GetGeneration(refGeneration)
	if err != nil {
		return prefix(fmt.Sprintf("Couldn't get generation for %s: ", mr.ModelName), err)
	}

	if mr.WriteOutputs {
		err = mr.initialiseDataset("outputs", gen.Outputs.Shape())
		if err != nil {
			return prefix("Couldn't init dataset for outputs: ", err)
		}
	}

	if mr.WriteInputs {
		err = mr.initialiseDataset("inputs", gen.Inputs.Shape())
		if err != nil {
			return prefix("Couldn't init dataset for inputs: ", err)
		}
	}

	mr.outputsInitialised = true
	return nil
}

func (mr *modelReference) writeData(label string, data data.ND3Float64, loc int32) error {
	ref := io.H5RefFloat64{}
	ref.Filename = mr.OutputFilename
	ref.Dataset = "/MODELS/" + mr.ModelName + "/" + label
	return ref.WriteSlice(data, []int{int(loc), 0, 0})
}

func (mr *modelReference) WriteData(generation int) error {
	gen, err := mr.GetGeneration(generation)
	if err != nil {
		return prefix("Cannot open generation: ", err)
	}

	if gen.Count == 0 {
		return nil
	}

	if !mr.outputsInitialised {
		if gen.Outputs.Len(0) > 0 {
			err = mr.InitialiseOutputs(generation)
			if err != nil {
				return prefix("Couldn't initialise outputs: ", err)
			}
		} else {
			return nil
		}
	}

	var loc int32 = 0
	if generation > 0 {
		loc = mr.Batches[generation-1]
	}

	if mr.WriteInputs {
		err = mr.writeData("inputs", gen.Inputs, loc)
		if err != nil {
			return prefix("Writing inputs ", err)
		}
	}

	if mr.WriteOutputs {
		err = mr.writeData("outputs", gen.Outputs, loc)
		if err != nil {
			return prefix("Writing outputs ", err)
		}
	}

	return nil
}

func verbosePrintln(a ...interface{}) {
	if verbose {
		fmt.Println(a...)
	}
}

func verbosePrintf(s string, a ...interface{}) {
	if verbose {
		fmt.Printf(s, a...)
	}
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()
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
	models := make(map[string]*modelReference)
	writingDone := make(chan int)

	for _, modelName := range modelNames {
		ref, err := initModel(fn, modelName)
		if err != nil {
			fmt.Println("Couldn't initialise model", modelName)
			fmt.Println(err)
			os.Exit(1)
		}

		if outputFn != "" {
			ref.OutputFilename = outputFn
			ref.WriteOutputs = true

			if ref.Batches[0] == 0 {
				ref.WriteInputs = true
			}
		}

		verbosePrintln("Batches for ", ref.ModelName, ref.Batches)
		verbosePrintln("Generations for ", ref.ModelName, ref.Generations)
		models[modelName] = ref
		genCount = len(ref.Generations)
	}

	fmt.Println()
	for i := 0; i < genCount; i++ {
		genTotal := 0
		fmt.Printf("==== Generation %d ====\n", i)
		simulationDone := make(chan string)
		genStart := time.Now()

		for _, modelName := range modelNames {
			gen, err := models[modelName].GetGeneration(i)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if gen.Count > 0 {
				verbosePrintf("* %d x %s\n", gen.Count, modelName)
			}
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

		for i, _ := range modelNames {
			mn := <-simulationDone
			verbosePrintf("%d: %s finished\n", i, mn)
		}

		genSimulationEnd := time.Now()
		genSimulationElapsed := genSimulationEnd.Sub(genStart)
		totalTimeSimulation += genSimulationElapsed.Seconds()

		verbosePrintf("= %d runs in %f seconds\n", genTotal, genSimulationElapsed.Seconds())

		if outputFn != "" {
			go func(g int) {
				if g > 0 {
					prevG := -1
					for {
						prevG = <-writingDone
						if prevG == (g - 1) {
							break
						}
						verbosePrintf("Waiting for generation %d, got generation %d, sleeping\n", g, prevG)
						writingDone <- prevG
						time.Sleep(time.Duration(1000 * 1000 * 500)) // Half a second
					}
				}

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
				writingDone <- g
			}(i)
		}
		// fmt.Printf("Results written in %f seconds\n", genWriteElapsed.Seconds())

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
		genLinkElapsed := genLinkEnd.Sub(genSimulationEnd)
		totalTimeLinks += genLinkElapsed.Seconds()

		verbosePrintf("%d links (%d to %d), processed in %f seconds\n", nextLink-currentLink, currentLink, nextLink, genLinkElapsed.Seconds())
		genElapsed := genLinkEnd.Sub(genStart)
		verbosePrintf("Generation completed in %f seconds\n", genElapsed.Seconds())
		verbosePrintln()
	}

	verbosePrintln("Simulation finished. Waiting for results to be written")
	generationsEnd := time.Now()

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

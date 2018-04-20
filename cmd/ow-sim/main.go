package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/flowmatters/openwater-core/data"

	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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

type modelGeneration struct {
	Model      sim.TimeSteppingModel
	Count      int
	Inputs     data.ND3Float64
	States     data.ND2Float64
	Parameters data.ND2Float64
	Outputs    data.ND3Float64
}

func (g *modelGeneration) Run() {
	g.Model.ApplyParameters(g.Parameters)
	g.Outputs = sim.InitialiseOutputs(g.Model, g.Inputs.Len(2), g.Inputs.Len(0))
	g.Model.Run(g.Inputs, g.States, g.Outputs)
}

type modelReference struct {
	Filename    string
	ModelName   string
	Batches     []int32
	Generations []*modelGeneration
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
		inputs, _ := inputRef.Load()
		gen.Inputs = inputs.(data.ND3Float64)

		paramRef := mr.GetReference(genSlice, "parameters")
		parameters, _ := paramRef.Load()
		gen.Parameters = parameters.(data.ND2Float64)

		stateRef := mr.GetReference(genSlice, "states")
		states, _ := stateRef.Load()
		gen.States = states.(data.ND2Float64)
	}
	return mr.Generations[i], nil
}

func (mr *modelReference) PurgeGeneration(i int) {
	mr.Generations[i] = nil
}

func loadGeneration(modelType string) {
	// Needs to be a member of the type...
	// For given model, load inputs, parameters, initial states for a given generation
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
	modelsRef := io.H5RefFloat64{Filename: fn, Dataset: "/META/models"}
	dimsRef := io.H5RefFloat64{Filename: fn, Dataset: "/DIMENSIONS"}
	//	procRef := io.H5Ref{Filename: fn, Dataset: "/PROCESSES"}

	modelNames, err := modelsRef.LoadText()
	if err != nil {
		fmt.Println("Couldn't read model metadata")
		os.Exit(1)
	}
	fmt.Println("Models", modelNames)

	dims, err := dimsRef.GetDatasets()
	if err != nil {
		fmt.Println("Couldn't read model dimensions")
		os.Exit(1)
	}
	fmt.Println("Dimensions", dims)

	linksRef := io.H5RefUint32{Filename: fn, Dataset: "/LINKS"}
	linksND, err := linksRef.Load()
	links := linksND.(data.ND2Uint32)
	linkSliceDim := []int{1, LINK_DEST_VAR + 1}
	linkSliceStep := []int{1, 1}
	nLinks := links.Len(0)
	nextLink := 0

	var genCount int
	models := make(map[string]*modelReference)
	for _, modelName := range modelNames {
		ref, err := initModel(fn, modelName)
		if err != nil {
			fmt.Println("Couldn't initialise model", modelName)
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Batches for ", ref.ModelName, ref.Batches)
		fmt.Println("Generations for ", ref.ModelName, ref.Generations)
		models[modelName] = ref
		genCount = len(ref.Generations)
	}

	fmt.Println()
	for i := 0; i < genCount; i++ {
		genTotal := 0
		fmt.Printf("==== Generation %d ====\n", i)
		for _, modelName := range modelNames {
			gen, err := models[modelName].GetGeneration(i)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if gen.Count > 0 {
				fmt.Printf("* %d x %s\n", gen.Count, modelName)
			}
			genTotal += gen.Count
			gen.Run()
			outputs := gen.Outputs
			if outputs == nil {
				fmt.Printf("No outputs from %s in generation %d\n", modelName, i)
			}
		}

		fmt.Printf("= %d runs\n", genTotal)

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
			srcIdx := link.Get1(LINK_SRC_GEN_NODE)
			srcData := srcModel.Outputs.Slice([]int{int(srcIdx), int(srcVar), 0}, []int{1, 1, nTimesteps}, []int{1, 1, 1})

			destVar := link.Get1(LINK_DEST_VAR)
			destIdx := link.Get1(LINK_DEST_GEN_NODE)
			destData := destModel.Inputs.Slice([]int{int(destIdx), int(destVar), 0}, []int{1, 1, nTimesteps}, []int{1, 1, 1})

			data.AddToFloat64Array(destData, srcData)
			nextLink++
		}
		fmt.Printf("%d links (%d to %d)\n", nextLink-currentLink, currentLink, nextLink)
		fmt.Println()
	}

	fmt.Println("Add stats")
}

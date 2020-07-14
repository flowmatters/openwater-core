package main

import (
	"fmt"
	"math"

	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
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
		verbosePrintf("Initialising Generation %d for %s\n",i+1,mr.ModelName)
		gen := modelGeneration{}
		modelRef := sim.Catalog[mr.ModelName]
		if modelRef == nil {
			errorMsg := fmt.Sprintf("Unknown model: %s", mr.ModelName)
			return nil, &errorString{errorMsg}
		}

		gen.Model = modelRef()
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
	verbosePrintf("Purging Generation %d for %s\n",i+1,mr.ModelName)
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

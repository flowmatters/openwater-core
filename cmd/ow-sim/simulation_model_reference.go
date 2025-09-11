package main

import (
	"encoding/binary"
	"fmt"
	gio "io"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	"github.com/flowmatters/openwater-core/io/protobuf"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/golang/protobuf/proto"
	"github.com/kardianos/osext"
)

type modelGeneration struct {
	Model      sim.TimeSteppingModel
	Count      int
	Inputs     data.ND3[float64]
	States     data.ND2[float64]
	Parameters data.ND2[float64]
	Outputs    data.ND3[float64]
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
	StructureFilename     string
	TimeSeriesFilename    string
	ParametersFilename    string
	InitialStatesFilename string
	OutputFilename        string
	FinalStatesFilename   string
	WriteInputs           bool
	WriteOutputs          bool
	WriteStates           bool
	ModelName             string
	Batches               []int32
	SimLength             int
	Dimensions            []int
	Generations           []*modelGeneration
	OutputWriter          *gio.PipeWriter
	OutputProcess         *exec.Cmd
	outputsInitialised    bool
}

func initModel(fn, model, paramFn string) (*modelReference, error) {
	modelRef := io.H5Ref[int32]{Filename: fn, Dataset: "/MODELS/" + model + "/batches"}
	batchesArray, err := modelRef.Load()
	if err != nil {
		log.Error().Msgf("Couldn't load batches for %s", model)
		return nil, err
	}
	batches := batchesArray.Unroll()
	result := modelReference{StructureFilename: fn, ModelName: model, Batches: batches}
	result.TimeSeriesFilename = fn
	result.ParametersFilename = paramFn
	result.InitialStatesFilename = fn
	dimensions, err := result.initDimensions()
	if err != nil {
		log.Error().Msgf("Couldn't initialise dimensions for %s", model)
		return nil, err
	}
	result.Dimensions = dimensions
	result.Generations = make([]*modelGeneration, len(batches))
	return &result, nil
}

func (mr *modelReference) makeModel() (sim.TimeSteppingModel, error) {
	modelRef := sim.Catalog[mr.ModelName]
	if modelRef == nil {
		errorMsg := fmt.Sprintf("Unknown model: %s", mr.ModelName)
		return nil, &errorString{errorMsg}
	}
	return modelRef(), nil
}

func (mr *modelReference) initDimensions() ([]int, error) {
	modelInstance, err := mr.makeModel()
	if err != nil {
		log.Error().Msgf("Couldn't make model")
		return nil, err
	}

	dims := modelInstance.Description().Dimensions
	if len(dims) == 0 {
		return []int{}, nil
	}

	h5Ref := io.H5Ref[float64]{}
	h5Ref.Dataset = "/MODELS/" + mr.ModelName + "/parameters"
	h5Ref.Filename = mr.ParametersFilename

	allParameters, err := h5Ref.Load()
	if err != nil {
		log.Error().Msgf("Couldn't load parameters for model %s from %s", mr.ModelName, mr.ParametersFilename)
		return nil, err
	}

	dimSizes := modelInstance.FindDimensions(allParameters.(data.ND2[float64]))

	log.Debug().Msgf("===== Simulation dimension sizes for %s =====", mr.ModelName)
	for ix, dim := range dims {
		log.Debug().Msgf("\t%s=%d", dim, dimSizes[ix])
	}

	return dimSizes, err
}

func (mr *modelReference) GetReference(genSlice []int, element string) io.H5Ref[float64] {
	result := io.H5Ref[float64]{}
	result.Dataset = "/MODELS/" + mr.ModelName + "/" + element

	if element == "parameters" {
		result.Slice = [][]int{nil, genSlice}
		result.Filename = mr.ParametersFilename
	} else {
		result.Slice = [][]int{genSlice, nil}
		result.Filename = mr.InitialStatesFilename
	}

	if element == "inputs" {
		result.Slice = append(result.Slice, nil)
		result.Filename = mr.TimeSeriesFilename
	}

	return result
}

func (mr *modelReference) GetGeneration(i int) (*modelGeneration, error) {
	if mr.Generations[i] == nil {
		log.Debug().Msgf("Initialising Generation %d for %s", i, mr.ModelName)
		gen := modelGeneration{}
		modelInstance, err := mr.makeModel()
		if err != nil {
			return nil, err
		}
		gen.Model = modelInstance
		if mr.Dimensions != nil {
			gen.Model.InitialiseDimensions(mr.Dimensions)
		}
		mr.Generations[i] = &gen
		genSlice := []int{0, int(mr.Batches[i]), 1}
		if i > 0 {
			genSlice[0] = int(mr.Batches[i-1])
		}
		gen.Count = genSlice[1] - genSlice[0]

		if gen.Count == 0 {
			gen.Inputs = data.NewArray3D[float64](0, 0, 0)
			gen.Parameters = data.NewArray2D[float64](0, 0)
			gen.States = data.NewArray2D[float64](0, 0)
			return mr.Generations[i], nil
		}

		inputRef := mr.GetReference(genSlice, "inputs")
		inputsExist := inputRef.Exists()
		if inputsExist {
			inputs, err := inputRef.Load()
			if err == nil {
				gen.Inputs = inputs.(data.ND3[float64])
				mr.SimLength = inputs.Len(sim.DIMI_TIMESTEP)
			} else {
				inputsExist = false
			}
		}

		if !inputsExist {
			log.Debug().Msgf("No inputs saved for %s. Initialising", mr.ModelName)
			gen.Inputs = data.NewArray3D[float64](gen.Count, len(gen.Model.Description().Inputs), mr.SimLength)
		}

		paramRef := mr.GetReference(genSlice, "parameters")
		parameters, err := paramRef.Load()
		if err != nil {
			return nil, err
		}
		gen.Parameters = parameters.(data.ND2[float64])

		stateRef := mr.GetReference(genSlice, "states")
		states, err := stateRef.Load()
		if err != nil {
			return nil, err
		}
		gen.States = states.(data.ND2[float64])
	}
	return mr.Generations[i], nil
}

func (mr *modelReference) PurgeGeneration(i int) {
	log.Debug().Msgf("Purging Generation %d for %s", i, mr.ModelName)
	mr.Generations[i] = nil
}

func (mr *modelReference) TotalRuns() int {
	return int(mr.Batches[len(mr.Batches)-1])
}

func (mr *modelReference) initialiseTimeSeriesDataset(label string, refShape []int) error {
	ref := io.H5Ref[float64]{}
	ref.Filename = mr.OutputFilename
	ref.Dataset = "/MODELS/" + mr.ModelName + "/" + label
	count := mr.TotalRuns()
	return ref.Create([]int{count, refShape[1], refShape[2]}, math.NaN(), false)
}

func (mr *modelReference) initialiseStatesDataset(label string, refShape []int) error {
	ref := io.H5Ref[float64]{}
	ref.Filename = mr.FinalStatesFilename
	ref.Dataset = "/MODELS/" + mr.ModelName + "/" + label
	count := mr.TotalRuns()
	return ref.Create([]int{count, refShape[1]}, math.NaN(), false)
}

func (mr *modelReference) InitialiseOutputs(refGeneration int) error {
	gen, err := mr.GetGeneration(refGeneration)
	if err != nil {
		return prefix(fmt.Sprintf("Couldn't get generation for %s: ", mr.ModelName), err)
	}

	if mr.WriteOutputs {
		err = mr.initialiseTimeSeriesDataset("outputs", gen.Outputs.Shape())
		if err != nil {
			return prefix("Couldn't init dataset for outputs: ", err)
		}
	}

	if mr.WriteInputs {
		err = mr.initialiseTimeSeriesDataset("inputs", gen.Inputs.Shape())
		if err != nil {
			return prefix("Couldn't init dataset for inputs: ", err)
		}
	}

	if mr.WriteStates {
		err = mr.initialiseStatesDataset("states", gen.States.Shape())
		if err != nil {
			return prefix("Couldn't init dataset for states: ", err)
		}
	}

	mr.outputsInitialised = true
	return nil
}

func (mr *modelReference) writeTimeSeries(label string, data data.ND3[float64], loc int32) error {
	ref := io.H5Ref[float64]{}
	ref.Filename = mr.OutputFilename
	ref.Dataset = "/MODELS/" + mr.ModelName + "/" + label
	return ref.WriteSlice(data, []int{int(loc), 0, 0})
}

func (mr *modelReference) writeStates(label string, data data.ND2[float64], loc int32) error {
	ref := io.H5Ref[float64]{}
	ref.Filename = mr.FinalStatesFilename
	ref.Dataset = "/MODELS/" + mr.ModelName + "/" + label
	return ref.WriteSlice(data, []int{int(loc), 0})
}

func (mr *modelReference) writeProtobuf(generation int) error {
	// TODO
	// * Possibly push writing to a goroutine controlled by a mutex
	//   (stored in modelreference?)
	data := &protobuf.ModelOutput{}
	data.Model = mr.ModelName

	gen, err := mr.GetGeneration(generation)
	if err != nil {
		return err
	}

	data.Cells = int32(gen.Count)
	data.TotalCells = int32(mr.TotalRuns())
	data.StartingLocation = mr.generationLocation(generation)

	if mr.WriteInputs {
		shp := gen.Inputs.Shape()
		data.Length = int32(shp[sim.DIMI_TIMESTEP])
		data.InputColumns = int32(shp[sim.DIMI_INPUT])
		data.InputValues = gen.Inputs.Unroll()
	}

	if mr.WriteOutputs {
		shp := gen.Outputs.Shape()
		data.Length = int32(shp[sim.DIMO_TIMESTEP])
		data.OutputColumns = int32(shp[sim.DIMO_OUTPUT])
		data.OutputValues = gen.Outputs.Unroll()
	}

	// Write the new address book back to disk.
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(msg)))
	log.Info().Msgf("Writing gen %d of %s", generation, mr.ModelName)
	if _, err := mr.OutputWriter.Write(buf); err != nil {
		return err
	}

	if _, err := mr.OutputWriter.Write(msg); err != nil {
		return err
	}
	log.Info().Msgf("Sent gen %d of %s", generation, mr.ModelName)

	if generation == len(mr.Batches)-1 {
		log.Info().Msgf("Waiting for output writer for %s to close", mr.ModelName)
		mr.OutputWriter.Close()
		mr.OutputProcess.Wait()
		log.Info().Msgf("Output writer for %s closed", mr.ModelName)
	}

	return nil
}

func (mr *modelReference) generationLocation(generation int) int32 {
	if generation == 0 {
		return 0
	}

	return mr.Batches[generation-1]

}

func (mr *modelReference) WriteData(generation int) error {
	gen, err := mr.GetGeneration(generation)
	if err != nil {
		return prefix("Cannot open generation: ", err)
	}

	if gen.Count == 0 {
		return nil
	}

	if mr.OutputProcess != nil {
		return mr.writeProtobuf(generation)
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

	loc := mr.generationLocation(generation)
	if generation > 0 {
		loc = mr.Batches[generation-1]
	}

	if mr.WriteInputs {
		err = mr.writeTimeSeries("inputs", gen.Inputs, loc)
		if err != nil {
			return prefix("Writing inputs ", err)
		}
	}

	if mr.WriteOutputs {
		err = mr.writeTimeSeries("outputs", gen.Outputs, loc)
		if err != nil {
			return prefix("Writing outputs ", err)
		}
	}

	if mr.WriteStates {
		err = mr.writeStates("states", gen.States, loc)
		if err != nil {
			return prefix("Writing states", err)
		}
	}

	return nil
}

func writeFor(modelName, includeFlag, excludeFlag string, defaultVal bool) bool {
	if includeFlag != "" {
		includedModels := strings.Split(includeFlag, ",")
		for _, im := range includedModels {
			if im == modelName {
				return true
			}
		}
	}

	if excludeFlag != "" {
		excludedModels := strings.Split(excludeFlag, ",")
		for _, exm := range excludedModels {
			if exm == modelName {
				return false
			}
		}
	}

	return defaultVal
}

func writeInputs(modelName string, defaultVal bool) bool {
	return writeFor(modelName, *inputsFor, *noInputsFor, defaultVal)
}

func writeOutputs(modelName string) bool {
	return writeFor(modelName, *outputsFor, *noOutputsFor, true)
}

func filenameOrDefault(flag *string, defaultFn string) string {
	if (flag == nil) || (*flag == "") {
		return defaultFn
	}

	return *flag
}

func makeModelRefs(modelNames []string, inputFn, defaultOutputFn string) (models map[string]*modelReference, genCount, nodeCount int) {
	simLength := 0

	outputPaths := make(map[string]string)
	if *splitOutputs != "" {
		pairs := strings.Split(*splitOutputs, ",")
		for _, pair := range pairs {
			elements := strings.Split(pair, "=")
			outputPaths[elements[0]] = elements[1]
		}
	}

	tsFilename := filenameOrDefault(timeseriesInputFile, inputFn)
	paramFilename := filenameOrDefault(parameterInputFile, inputFn)
	initStatesFilename := filenameOrDefault(statesInputFile, inputFn)

	nodeCount = 0
	models = make(map[string]*modelReference)
	for _, modelName := range modelNames {
		ref, err := initModel(inputFn, modelName, paramFilename)

		if err != nil {
			log.Fatal().Stack().Err(err).Str("Model name", modelName).Msg("Couldn't initialise model")
		}

		ref.TimeSeriesFilename = tsFilename
		ref.InitialStatesFilename = initStatesFilename

		if simLength == 0 {
			log.Debug().Msg("Trying to establish simulation length...")
			inputRef := io.H5Ref[float64]{}
			inputRef.Filename = tsFilename
			inputRef.Dataset = "/MODELS/" + modelName + "/inputs"
			if inputRef.Exists() {
				inputShp, err := inputRef.Shape()
				if err != nil {
					log.Fatal().Stack().Err(err).Str("Model Name", modelName).Msg("Couldn't query input dimensions")
				}
				simLength = inputShp[sim.DIMI_TIMESTEP]
			}
			log.Debug().Msgf("Simulation has %d timesteps", simLength)
		}

		destFn := outputPaths[modelName]
		if destFn == "" {
			destFn = defaultOutputFn
		}

		if destFn != "" {
			ref.FinalStatesFilename = filenameOrDefault(statesOutputFile, destFn)
			ref.OutputFilename = destFn
			ref.WriteOutputs = writeOutputs(modelName)
			ref.WriteStates = true
			ref.WriteInputs = writeInputs(modelName, ref.Batches[0] == 0)

			if destFn != defaultOutputFn {
				exe_path, _ := osext.Executable()
				log.Info().Msgf("Configuring external write process: %s", exe_path)
				write_cmd := exec.Command(exe_path, "-writer", destFn)
				reader, writer := gio.Pipe()
				write_cmd.Stdin = reader
				write_cmd.Stdout = os.Stdout
				ref.OutputWriter = writer
				write_cmd.Start()
				ref.OutputProcess = write_cmd
			}
		}

		log.Debug().Str("Model Name", ref.ModelName).Any("Batches:", ref.Batches).Any("Generations", ref.Generations).Msg("")
		models[modelName] = ref
		genCount = len(ref.Generations)
		nodeCount += int(ref.Batches[len(ref.Batches)-1])
	}

	for _, r := range models {
		r.SimLength = simLength
	}

	return
}

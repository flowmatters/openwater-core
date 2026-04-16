package main

import (
	"encoding/binary"
	"fmt"
	gio "io"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	"github.com/flowmatters/openwater-core/io/protobuf"
	"github.com/flowmatters/openwater-core/sim"
	"gonum.org/v1/hdf5"
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

	// genMu serializes access to Generations[]. GetGeneration may be called
	// concurrently from parallel link-loop workers: the original implementation
	// was write-before-init (publishes &gen before populating gen.Inputs), so
	// without a lock two workers racing on the same uninitialized (model, gen)
	// could see a partially-constructed modelGeneration and nil-deref on
	// Inputs/Outputs. The lock is taken only around the nil-check and
	// initialization path; fully-initialized generations are effectively
	// cache-hits that still take+release the lock but do no HDF5 work.
	genMu sync.Mutex

	// Cached existence checks for the input file's per-model datasets.
	// The input file structure is read-only during a run, so these only need
	// to be checked once per model (in initModel), avoiding 2 x Exists()
	// calls per GetGeneration init — each of which opens+reads the HDF5 file
	// and contends with the writer on the global HDF5 mutex.
	hasInputs bool
	hasStates bool
}

func initModel(fn, model, paramFn string) (*modelReference, error) {
	modelRef := io.H5Ref[int32]{Filename: fn, Dataset: "/MODELS/" + model + "/batches"}
	batchesArray, err := modelRef.Load()
	if err != nil {
		log.Error().Str("Model Name", model).Msg("Couldn't load batches")
		return nil, err
	}
	batches := batchesArray.Unroll()
	result := modelReference{StructureFilename: fn, ModelName: model, Batches: batches}
	result.TimeSeriesFilename = fn
	result.ParametersFilename = paramFn
	result.InitialStatesFilename = fn
	dimensions, err := result.initDimensions()
	if err != nil {
		log.Error().Str("Model Name", model).Msg("Couldn't initialise dimensions")
		return nil, err
	}
	result.Dimensions = dimensions
	result.Generations = make([]*modelGeneration, len(batches))

	// Cache whether this model has stored inputs/states datasets.
	// Checked once here to avoid per-generation Exists() calls inside
	// GetGeneration which are expensive (file-open + H5Lexists + close)
	// and contend with the writer on the global HDF5 mutex.
	inputRef := io.H5Ref[float64]{Filename: fn, Dataset: "/MODELS/" + model + "/inputs"}
	result.hasInputs = inputRef.Exists()
	stateRef := io.H5Ref[float64]{Filename: fn, Dataset: "/MODELS/" + model + "/states"}
	result.hasStates = stateRef.Exists()

	return &result, nil
}

func (mr *modelReference) makeModel() (sim.TimeSteppingModel, error) {
	modelRef := sim.Catalog[mr.ModelName]
	if modelRef == nil {
		log.Error().Str("Model Name", mr.ModelName).Msg("Unknown model")
		errorMsg := fmt.Sprintf("Unknown model: %s", mr.ModelName)
		return nil, &errorString{errorMsg}
	}
	return modelRef(), nil
}

func (mr *modelReference) initDimensions() ([]int, error) {
	modelInstance, err := mr.makeModel()
	if err != nil {
		log.Error().Str("Model Name", mr.ModelName).Msg("Couldn't make model")
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
		log.Error().Str("Model Name", mr.ModelName).Str("Parameters Filename", mr.ParametersFilename).Msg("Couldn't load parameters for model")
		return nil, err
	}

	dimSizes := modelInstance.FindDimensions(allParameters.(data.ND2[float64]))

	log.Debug().Str("Model Name", mr.ModelName).Msg("Simulation dimension sizes")
	for ix, dim := range dims {
		log.Debug().Str("Dimension", dim).Int("Size", dimSizes[ix]).Msg("Dimension size")
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
	mr.genMu.Lock()
	defer mr.genMu.Unlock()
	if mr.Generations[i] == nil {
		log.Debug().Int("Generation", i).Str("Model Name", mr.ModelName).Msg("Initialising generation for model")
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
		paramRef := mr.GetReference(genSlice, "parameters")
		stateRef := mr.GetReference(genSlice, "states")

		// Batch HDF5 reads: when inputs, parameters, and states all live in
		// the same file (the common case), open the file once under a single
		// global-mutex acquisition. This reduces contention with the writer
		// goroutine from 2-3 separate lock+open+close cycles per init down
		// to one. Falls back to individual Load() calls when files differ.
		sameFile := inputRef.Filename == paramRef.Filename && paramRef.Filename == stateRef.Filename
		if sameFile {
			err = io.WithReadFile(paramRef.Filename, func(f *hdf5.File) error {
				if mr.hasInputs {
					inputs, err := inputRef.LoadFromFile(f)
					if err == nil {
						gen.Inputs = inputs.(data.ND3[float64])
						mr.SimLength = inputs.Len(sim.DIMI_TIMESTEP)
					}
				}
				if gen.Inputs == nil {
					gen.Inputs = data.NewArray3D[float64](gen.Count, len(gen.Model.Description().Inputs), mr.SimLength)
				}

				parameters, err := paramRef.LoadFromFile(f)
				if err != nil {
					return prefix("loading parameters: ", err)
				}
				gen.Parameters = parameters.(data.ND2[float64])

				if mr.hasStates {
					states, err := stateRef.LoadFromFile(f)
					if err != nil {
						return prefix("loading states: ", err)
					}
					gen.States = states.(data.ND2[float64])
				} else {
					gen.States = gen.Model.InitialiseStates(gen.Count)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			// Rare case: data spread across multiple files.
			inputsLoaded := false
			if mr.hasInputs {
				inputs, loadErr := inputRef.Load()
				if loadErr == nil {
					gen.Inputs = inputs.(data.ND3[float64])
					mr.SimLength = inputs.Len(sim.DIMI_TIMESTEP)
					inputsLoaded = true
				}
			}
			if !inputsLoaded {
				gen.Inputs = data.NewArray3D[float64](gen.Count, len(gen.Model.Description().Inputs), mr.SimLength)
			}

			parameters, loadErr := paramRef.Load()
			if loadErr != nil {
				return nil, loadErr
			}
			gen.Parameters = parameters.(data.ND2[float64])

			if mr.hasStates {
				states, loadErr := stateRef.Load()
				if loadErr != nil {
					return nil, loadErr
				}
				gen.States = states.(data.ND2[float64])
			} else {
				gen.States = gen.Model.InitialiseStates(gen.Count)
			}
		}
	}
	return mr.Generations[i], nil
}

func (mr *modelReference) PurgeGeneration(i int) {
	log.Debug().Int("Generation", i).Str("Model Name", mr.ModelName).Msg("Purging generation for model")
	mr.Generations[i] = nil
}

func (mr *modelReference) TotalRuns() int {
	return int(mr.Batches[len(mr.Batches)-1])
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
	log.Info().Int("Generation", generation).Str("Model Name", mr.ModelName).Msg("Writing protobuf for generation")
	if _, err := mr.OutputWriter.Write(buf); err != nil {
		return err
	}

	if _, err := mr.OutputWriter.Write(msg); err != nil {
		return err
	}
	log.Info().Int("Generation", generation).Str("Model Name", mr.ModelName).Msg("Sent protobuf for generation")

	if generation == len(mr.Batches)-1 {
		log.Info().Str("Model Name", mr.ModelName).Msg("Waiting for output writer to close")
		mr.OutputWriter.Close()
		mr.OutputProcess.Wait()
		log.Info().Str("Model Name", mr.ModelName).Msg("Output writer closed")
	}

	return nil
}

func (mr *modelReference) generationLocation(generation int) int32 {
	if generation == 0 {
		return 0
	}

	return mr.Batches[generation-1]

}

// WriteData writes generation outputs/inputs/states. Opens the output
// file once for all datasets in this model's generation (rather than once
// per dataset).
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

	// Initialise output datasets on first write (needs separate lock
	// acquisitions because Create opens/creates the file).
	if !mr.outputsInitialised {
		if gen.Outputs.Len(0) > 0 {
			err = mr.initialiseOutputsViaCreate(generation)
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

	// Batch the actual data writes under a single file-open.
	if mr.OutputFilename == mr.FinalStatesFilename {
		return io.WithWriteFile(mr.OutputFilename, func(f *hdf5.File) error {
			return mr.writeAllDatasets(f, f, gen, loc)
		})
	}
	return io.WithWriteFile(mr.OutputFilename, func(outF *hdf5.File) error {
		return io.WithWriteFile(mr.FinalStatesFilename, func(stF *hdf5.File) error {
			return mr.writeAllDatasets(outF, stF, gen, loc)
		})
	})
}

func (mr *modelReference) initialiseOutputsViaCreate(refGeneration int) error {
	gen, err := mr.GetGeneration(refGeneration)
	if err != nil {
		return prefix(fmt.Sprintf("Couldn't get generation for %s: ", mr.ModelName), err)
	}

	compressLevel := *compressOutputs

	if mr.WriteOutputs {
		ref := io.H5Ref[float64]{Filename: mr.OutputFilename, Dataset: "/MODELS/" + mr.ModelName + "/outputs"}
		if err := ref.Create([]int{mr.TotalRuns(), gen.Outputs.Shape()[1], gen.Outputs.Shape()[2]}, math.NaN(), compressLevel); err != nil {
			return prefix("Couldn't init dataset for outputs: ", err)
		}
	}

	if mr.WriteInputs {
		ref := io.H5Ref[float64]{Filename: mr.OutputFilename, Dataset: "/MODELS/" + mr.ModelName + "/inputs"}
		if err := ref.Create([]int{mr.TotalRuns(), gen.Inputs.Shape()[1], gen.Inputs.Shape()[2]}, math.NaN(), compressLevel); err != nil {
			return prefix("Couldn't init dataset for inputs: ", err)
		}
	}

	if mr.WriteStates {
		// States are 2D [nodes, vars] and typically tiny — don't compress.
		// Also avoids issues with zero-dimension states (e.g. models with
		// no state variables produce [N, 0] which can't be chunked).
		ref := io.H5Ref[float64]{Filename: mr.FinalStatesFilename, Dataset: "/MODELS/" + mr.ModelName + "/states"}
		if err := ref.Create([]int{mr.TotalRuns(), gen.States.Shape()[1]}, math.NaN(), 0); err != nil {
			return prefix("Couldn't init dataset for states: ", err)
		}
	}

	mr.outputsInitialised = true
	return nil
}

func (mr *modelReference) writeAllDatasets(outputFile, statesFile *hdf5.File, gen *modelGeneration, loc int32) error {
	if mr.WriteInputs {
		ref := io.H5Ref[float64]{Filename: mr.OutputFilename, Dataset: "/MODELS/" + mr.ModelName + "/inputs"}
		if err := ref.WriteSliceToFile(outputFile, gen.Inputs, []int{int(loc), 0, 0}); err != nil {
			return prefix("Writing inputs ", err)
		}
	}

	if mr.WriteOutputs {
		ref := io.H5Ref[float64]{Filename: mr.OutputFilename, Dataset: "/MODELS/" + mr.ModelName + "/outputs"}
		if err := ref.WriteSliceToFile(outputFile, gen.Outputs, []int{int(loc), 0, 0}); err != nil {
			return prefix("Writing outputs ", err)
		}
	}

	if mr.WriteStates {
		ref := io.H5Ref[float64]{Filename: mr.FinalStatesFilename, Dataset: "/MODELS/" + mr.ModelName + "/states"}
		if err := ref.WriteSliceToFile(statesFile, gen.States, []int{int(loc), 0}); err != nil {
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

func inList(csv, name string) bool {
	if csv == "" {
		return false
	}
	for _, item := range strings.Split(csv, ",") {
		if item == name {
			return true
		}
	}
	return false
}

func writeInputs(modelName string, defaultVal bool) bool {
	if *noInputs {
		return false
	}
	if *onlyInputsFor != "" {
		return inList(*onlyInputsFor, modelName)
	}
	return writeFor(modelName, *inputsFor, *noInputsFor, defaultVal)
}

func writeOutputs(modelName string) bool {
	if *noOutputs {
		return false
	}
	if *onlyOutputsFor != "" {
		return inList(*onlyOutputsFor, modelName)
	}
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
			log.Fatal().Stack().Err(err).Str("Model Name", modelName).Msg("Couldn't initialise model")
		}

		ref.TimeSeriesFilename = tsFilename
		ref.InitialStatesFilename = initStatesFilename

		if simLength == 0 {
			log.Debug().Str("Model Name", modelName).Msg("Trying to establish simulation length")
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
			log.Debug().Str("Model Name", modelName).Int("Timesteps", simLength).Msg("Simulation length established")
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
				log.Info().Str("Executable Path", exe_path).Msg("Configuring external write process")
				write_cmd := exec.Command(exe_path, "-writer", destFn)
				reader, writer := gio.Pipe()
				write_cmd.Stdin = reader
				write_cmd.Stdout = os.Stdout
				ref.OutputWriter = writer
				write_cmd.Start()
				ref.OutputProcess = write_cmd
			}
		}

		log.Debug().Str("Model Name", ref.ModelName).Any("Batches", ref.Batches).Any("Generations", ref.Generations).Msg("Model reference initialised")
		models[modelName] = ref
		genCount = len(ref.Generations)
		nodeCount += int(ref.Batches[len(ref.Batches)-1])
	}

	for _, r := range models {
		r.SimLength = simLength
	}

	return
}

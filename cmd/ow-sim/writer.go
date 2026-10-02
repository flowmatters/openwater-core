package main

import (
	"encoding/binary"
	gio "io"
	"math"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	"github.com/flowmatters/openwater-core/io/protobuf"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/golang/protobuf/proto"
)

func initialiseDataset(fn, modelName, label string, shape []int) error {
	ref := io.H5Ref[float64]{}
	ref.Filename = fn
	ref.Dataset = "/MODELS/" + modelName + "/" + label
	if err := ref.Create(shape, math.NaN(), 0); err != nil {
		return err
	}

	if factory := sim.Catalog[modelName]; factory != nil {
		description := factory().Description()
		variables := description.Outputs
		if label == "inputs" {
			variables = description.Inputs
		}
		nameVariables(ref, sim.VariableNames(variables), shape[1])
	}
	return nil
}

func writeData(
	fn, modelName, label string,
	values []float64,
	loc, cells, columns, rows int32) error {

	ref := io.H5Ref[float64]{}
	ref.Filename = fn
	ref.Dataset = "/MODELS/" + modelName + "/" + label

	shp := []int{int(cells), int(columns), int(rows)}
	arr := data.ArrayFromSlice[float64](values, shp)

	return ref.WriteSlice(arr, []int{int(loc), 0, 0})
}

func run_writer(args []string) {
	fn := args[0]
	log.Info().Str("Target File", fn).Msg("Writing results from stdin")

	// The simulating process passes the model file so the results can be
	// traced back to it.
	modelStamp := map[string]string{}
	if len(args) > 1 {
		modelStamp = loadModelStamp(args[1])
	}
	stampResultsFile(fn, modelStamp)
	input := os.Stdin

	for {
		buf := make([]byte, 4)
		if _, err := gio.ReadFull(input, buf); err != nil {
			return
		}

		size := binary.LittleEndian.Uint32(buf)

		msg := make([]byte, size)
		if _, err := gio.ReadFull(input, msg); err != nil {
			log.Error().Stack().Err(err).Msg("")
			return
		}

		data := &protobuf.ModelOutput{}
		if err := proto.Unmarshal(msg, data); err != nil {
			log.Fatal().Stack().Err(err).Msg("Failed to parse model data")
		}

		if data.Cells == 0 {
			continue
		}

		log.Info().Str("Model", data.Model).Str("Filename", fn).Int("Cells", int(data.Cells)).Msg("Writing data from model to file")

		if data.InputColumns > 0 {
			if data.StartingLocation == 0 {
				shp := []int{int(data.TotalCells), int(data.InputColumns), int(data.Length)}
				initialiseDataset(fn, data.Model, "inputs", shp)
			}

			err := writeData(fn, data.Model, "inputs", data.InputValues,
				data.StartingLocation, data.Cells, data.InputColumns, data.Length)
			if err != nil {
				log.Error().Stack().Err(err).Msg("")
				return
			}
		}

		if data.OutputColumns > 0 {
			if data.StartingLocation == 0 {
				shp := []int{int(data.TotalCells), int(data.OutputColumns), int(data.Length)}
				initialiseDataset(fn, data.Model, "outputs", shp)
			}

			err := writeData(fn, data.Model, "outputs", data.OutputValues,
				data.StartingLocation, data.Cells, data.OutputColumns, data.Length)
			if err != nil {
				log.Error().Stack().Err(err).Msg("")
				return
			}
		}

		// Construct Arrays

		// data.Cells = int32(gen.Count);
		// data.TotalCells = int32(mr.TotalRuns())
		// data.StartingLocation = mr.generationLocation(generation)

		// if mr.WriteInputs {
		// 	shp := gen.Inputs.Shape()
		// 	data.Length = int32(shp[sim.DIMI_TIMESTEP])
		// 	data.InputColumns = int32(shp[sim.DIMI_INPUT])
		// 	data.InputValues = gen.Inputs.Unroll()
		// }

		// if mr.WriteOutputs {
		// 	shp := gen.Outputs.Shape()
		// 	data.Length = int32(shp[sim.DIMO_TIMESTEP])
		// 	data.OutputColumns = int32(shp[sim.DIMO_OUTPUT])
		// 	data.OutputValues = gen.Outputs.Unroll()
		// }

	}
}

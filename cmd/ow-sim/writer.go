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
	"github.com/golang/protobuf/proto"
)

func initialiseDataset(fn, modelName, label string, shape []int) error {
	ref := io.H5Ref[float64]{}
	ref.Filename = fn
	ref.Dataset = "/MODELS/" + modelName + "/" + label
	return ref.Create(shape, math.NaN(), 0)
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

package main

import (
	"fmt"
	"math"
	gio "io"
	"os"
	"log"
	"encoding/binary"
	"github.com/flowmatters/openwater-core/io/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
)

func initialiseDataset(fn, modelName, label string, shape []int) error {
	ref := io.H5RefFloat64{}
	ref.Filename = fn
	ref.Dataset = "/MODELS/" + modelName + "/" + label
	return ref.Create(shape, math.NaN(),false)
}

func writeData(
	fn, modelName, label string, 
	values []float64,
	loc, cells, columns, rows int32) error {

	ref := io.H5RefFloat64{}
	ref.Filename = fn
	ref.Dataset = "/MODELS/" + modelName + "/" + label

	shp := []int{int(cells),int(columns),int(rows)}
	arr := data.ArrayFromSliceFloat64(values,shp)

	return ref.WriteSlice(arr, []int{int(loc), 0, 0})
}

func run_writer(args []string){
	fn := args[0]
	fmt.Printf("Writing results from stdin to %s\n",fn)
	input := os.Stdin

	for ;; {
    buf := make([]byte, 4)
    if _, err := gio.ReadFull(input, buf); err != nil {
        return
    }

    size := binary.LittleEndian.Uint32(buf)

    msg := make([]byte, size)
    if _, err := gio.ReadFull(input, msg); err != nil {
				fmt.Println(err)
        return
    }

		data := &protobuf.ModelOutput{}
		if err := proto.Unmarshal(msg, data); err != nil {
			log.Fatalln("Failed to parse model data:", err)
		}

		if data.Cells == 0 {
			continue
		}

		fmt.Printf("Writing data from %s to %s (%d cells)\n",data.Model,fn,data.Cells)

		if data.InputColumns > 0 {
			if data.StartingLocation == 0 {
				shp := []int{int(data.TotalCells), int(data.InputColumns), int(data.Length)}
				initialiseDataset(fn,data.Model,"inputs",shp)
			}

			err := writeData(fn,data.Model,"inputs",data.InputValues,
								data.StartingLocation,data.Cells,data.InputColumns,data.Length)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		if data.OutputColumns > 0 {
			if data.StartingLocation == 0 {
				shp := []int{int(data.TotalCells), int(data.OutputColumns), int(data.Length)}
				initialiseDataset(fn,data.Model,"outputs",shp)
			}

			err := writeData(fn,data.Model,"outputs",data.OutputValues,
											data.StartingLocation,data.Cells,data.OutputColumns,data.Length)
			if err != nil {
				fmt.Println(err)
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
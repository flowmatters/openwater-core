package main

import (
	"fmt"

	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
)

func uniformParameters(params []float64, n int) data.ND2Float64 {
	res := data.NewArray2D(len(params), 1)
	for i := 0; i < len(params); i++ {
		//    res[i]=make([]float64,n);
		for j := 0; j < n; j++ {
			res.Set2(i, j, params[i])
		}
	}

	return res
}

func main() {
	fmt.Println("hello worldly")
	fn := "/Users/joelrahman/Documents/FlowMatters/Projects/045_RainfallRunoffApplication_ALCOA/FromJustin/Lewis_RR_input.csv"
	inputTS, err := io.ReadTimeSeriesCSV(fn)
	if err != nil {
		fmt.Println(err)
		return
	}
	rainfall := inputTS["P"]
	pet := inputTS["E"]

	model := sim.Catalog["GR4J"]()
	//model.X1 =

	var params = []float64{350.0, 0.5, 40.0, 4}
	paramsEnumerated := uniformParameters(params, 1)
	model.ApplyParameters(paramsEnumerated)

	//if model.X1[0] != 350.0{
	//  panic(model);
	//};
	//
	//if model.X2[0] != 0 {
	//  panic(model);
	//}
	//
	//if model.X3[0] != 40.0 {
	//  panic(model);
	//}
	//
	//if model.X4[0] != 0.5 {
	//  panic(model);
	//}

	states := model.InitialiseStates(1)

	inputs := data.NewArray3D(1, 2, len(rainfall))
	inputs.Apply([]int{0, 0, 0}, 2, 1, rainfall)
	inputs.Apply([]int{0, 0, 1}, 2, 1, pet)

	//  fmt.Println(len(rainfall));

	fmt.Println(states)
	outputs := sim.InitialiseOutputs(model, len(rainfall), 1)
	model.Run(inputs, states, outputs)
	modelResults := sim.RunResults{outputs, states}
	//  runoff := modelResults.Outputs[0][0];

	fmt.Println(modelResults.States)

	//  fmt.Println(runoff);
}

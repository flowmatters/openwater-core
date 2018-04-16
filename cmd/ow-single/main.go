package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/io"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
)

type (
	singleModel struct {
		Name       string
		Inputs     modelInputs
		States     modelValues
		Parameters modelValues
	}

	modelInputs []modelInput

	modelInput struct {
		Name   string
		Values []float64
	}

	modelValues []modelValue
	modelValue  struct {
		Name  string
		Value float64
	}

	overallResults struct {
		Log        []string
		RunResults struct {
			Outputs []interface{}
			States  []interface{}
		}
	}
)

func (vals modelInputs) Find(name string) []float64 {
	for _, v := range vals {
		if v.Name == name {
			return v.Values
		}
	}

	return nil
}

func (vals modelValues) Find(name string) float64 {
	for _, v := range vals {
		if v.Name == name {
			return v.Value
		}
	}

	return 0
}

func (m singleModel) Initialise() (error, sim.TimeSteppingModel, data.ND3Float64, data.ND2Float64) {
	if m.Name == "" {
		return errors.New("No model name provided"), nil, nil, nil
	}
	factory := sim.Catalog[m.Name]
	if factory == nil {
		return errors.New(fmt.Sprintf("Unknown model: %s", m.Name)), nil, nil, nil
	}
	model := factory()
	desc := model.Description()

	params := make([]float64, len(desc.Parameters))
	for i, p := range desc.Parameters {
		params[i] = m.Parameters.Find(p.Name)
	}

	model.ApplyParameters(uniformParameters(params, 1))
	var states data.ND2Float64
	if len(m.States) == len(desc.States) {
		log("TODO. Use supplied states")
		states = model.InitialiseStates(1)
	} else {
		states = model.InitialiseStates(1)
	}
	log(fmt.Sprintf("%#v", states))
	//log(fmt.Sprint(m.Inputs))
	var inputs data.ND3Float64 = nil
	for i, p := range desc.Inputs {
		thisInput := m.Inputs.Find(p)
		//fmt.Println(p,thisInput)
		if thisInput == nil {
			return errors.New(fmt.Sprintf("Missing input: %s", p)), nil, nil, nil
		}

		if inputs == nil {
			inputs = data.NewArray3DFloat64(1, len(desc.Inputs), len(thisInput))
		}

		inputs.Apply([]int{0, i, 0}, 2, 1, thisInput)
	}

	return nil, model, inputs, states
}

func uniformParameters(params []float64, n int) data.ND2Float64 {
	res := data.NewArray2DFloat64(len(params), n)
	for i := 0; i < len(params); i++ {
		for j := 0; j < n; j++ {
			res.Set2(i, j, params[i])
		}
	}

	return res
}

var runLogs []string
var results sim.RunResults

// func jsonSafe2D(vals data.NDFloat64, shiftDim int) [][]interface{} {
//   result := make([][]interface{}, len(vals))
//   for i,v := range(vals){
//     result[i] = jsonSafe(v)
//   }
//   return result
// }

// func jsonSafe(vals data.NDFloat64) []interface{} {
//   result := make([]interface{}, len(vals))
//   for i,v := range(vals){
//     if math.IsNaN(v) {
//       result[i] = fmt.Sprint(v)
//     } else if math.IsInf(v,0) {
//       result[i] = fmt.Sprint(v)
//     } else {
//       result[i] = v
//     }
//   }
//   return result
// }

func encodeResults() {
	//  outputs := jsonSafe3D(results.Outputs)
	// results.Outputs := outputs
	// results.States := jsonSafe2D(results.States)

	var overall overallResults
	overall.Log = runLogs
	//  fmt.Println("O",results.Outputs.Shape(),"S",results.States.Shape())
	//fmt.Println(results.Outputs.Shape())
	overall.RunResults.Outputs = io.JsonSafeArray(results.Outputs.MustReshape(results.Outputs.Shape()[1:]), 0)
	overall.RunResults.States = io.JsonSafeArray(results.States.MustReshape(results.States.Shape()[1:]), 0)
	encoder := json.NewEncoder(os.Stdout)

	err := encoder.Encode(overall)
	if err != nil {
		fmt.Println(overall.Log)
		fmt.Println(err)
	}
}

func log(s string) {
	runLogs = append(runLogs, s)
	//fmt.Println(s)
}

func main() {

	var modelDescription singleModel
	decoder := json.NewDecoder(os.Stdin)
	err := decoder.Decode(&modelDescription)
	if err != nil {
		log(err.Error())
		return
	}

	defer encodeResults()
	err, model, inputs, states := modelDescription.Initialise()
	//  fmt.Println(inputs)

	if err != nil {
		log(err.Error())
		return
	}

	log(fmt.Sprint(states))
	log(fmt.Sprintln(states.Len(0)))
	outputs := sim.InitialiseOutputs(model, inputs.Len3(), 1)
	//fmt.Println(outputs.Shape())
	model.Run(inputs, states, outputs)
	results = sim.RunResults{}
	results.Outputs = outputs
	results.States = states
	//  runoff := modelResults.Outputs[0][0];

	log(fmt.Sprintf("%#v", results.States))

	//  fmt.Println(runoff);
}

package sim

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/flowmatters/openwater-core/data"
	owjs "github.com/flowmatters/openwater-core/io/json"
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

	singleModelResults struct {
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

func (m singleModel) Initialise() (error, TimeSteppingModel, data.ND3Float64, data.ND2Float64) {
	if m.Name == "" {
		return errors.New("No model name provided"), nil, nil, nil
	}
	factory := Catalog[m.Name]
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
		//TODO: log("TODO. Use supplied states")
		states = model.InitialiseStates(1)
	} else {
		states = model.InitialiseStates(1)
	}
	var inputs data.ND3Float64 = nil
	for i, p := range desc.Inputs {
		thisInput := m.Inputs.Find(p)
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

func RunSingleModelJSON(r io.Reader, w io.Writer) {
	var runLogs []string
	var results = RunResults{}

	log := func(s string) {
		runLogs = append(runLogs, s)
	}

	defer func() {
		encodeResults(w, runLogs, results)
	}()

	var modelDescription singleModel
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&modelDescription)
	if err != nil {
		log(err.Error())
		return
	}

	err, model, inputs, states := modelDescription.Initialise()

	if err != nil {
		log(err.Error())
		return
	}

	log(fmt.Sprint(states))
	log(fmt.Sprintln(states.Len(0)))
	outputs := InitialiseOutputs(model, inputs.Len3(), 1)
	//fmt.Println(outputs.Shape())
	model.Run(inputs, states, outputs)
	results.Outputs = outputs
	results.States = states
	//  runoff := modelResults.Outputs[0][0];

	log(fmt.Sprintf("%#v", results.States))

	//  fmt.Println(runoff);
}

func encodeResults(w io.Writer, runLogs []string, results RunResults) {
	//  outputs := jsonSafe3D(results.Outputs)
	// results.Outputs := outputs
	// results.States := jsonSafe2D(results.States)

	var overall singleModelResults
	overall.Log = runLogs
	//  fmt.Println("O",results.Outputs.Shape(),"S",results.States.Shape())
	//fmt.Println(results.Outputs.Shape())
	if results.Outputs != nil {
		overall.RunResults.Outputs = owjs.JsonSafeArray(results.Outputs.MustReshape(results.Outputs.Shape()[1:]), 0)
	}

	if results.States != nil {
		overall.RunResults.States = owjs.JsonSafeArray(results.States.MustReshape(results.States.Shape()[1:]), 0)
	}

	encoder := json.NewEncoder(w)

	err := encoder.Encode(overall)
	if err != nil {
		fmt.Println(overall.Log)
		fmt.Println(err)
	}
}

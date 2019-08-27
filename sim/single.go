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
			Outputs interface{}
			States  interface{}
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

func (vals modelValues) Find(name string, defaultValue float64) (float64, string) {
	for _, v := range vals {
		if v.Name == name {
			return v.Value, ""
		}
	}

	return defaultValue, fmt.Sprintf("%s not found, using default=%f", name, defaultValue)
}

func (m singleModel) Initialise() (error, TimeSteppingModel, data.ND3Float64, data.ND2Float64, []string) {
	warnings := make([]string, 1)

	if m.Name == "" {
		return errors.New("No model name provided"), nil, nil, nil, warnings
	}
	factory := Catalog[m.Name]
	if factory == nil {
		return errors.New(fmt.Sprintf("Unknown model: %s", m.Name)), nil, nil, nil, warnings
	}
	model := factory()
	desc := model.Description()

	params := make([]float64, len(desc.Parameters))
	for i, p := range desc.Parameters {
		paramValue, msg := m.Parameters.Find(p.Name, p.Default)
		params[i] = paramValue
		if msg != "" {
			warnings = append(warnings, msg)
		}
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
			warnings = append(warnings, fmt.Sprintf("Missing input: %s, using 0", p))
			continue
		}

		if inputs == nil {
			inputs = data.NewArray3DFloat64(1, len(desc.Inputs), len(thisInput))
		}

		inputs.Apply([]int{0, i, 0}, 2, 1, thisInput)
	}

	return nil, model, inputs, states, warnings
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

func RunSingleModelJSON(r io.Reader, w io.Writer, splitOutputs bool) {
	var runLogs []string
	var results = RunResults{}
	var description ModelDescription

	log := func(s string) {
		runLogs = append(runLogs, s)
	}

	defer func() {
		encodeResults(w, runLogs, results, description, splitOutputs)
	}()

	var modelDescription singleModel
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&modelDescription)
	if err != nil {
		log(err.Error())
		return
	}

	err, model, inputs, states, warnings := modelDescription.Initialise()

	if err != nil {
		log(err.Error())
		return
	}

	for _, w := range warnings {
		log(w)
	}

	description = model.Description()

	outputs := InitialiseOutputs(model, inputs.Len3(), 1)

	model.Run(inputs, states, outputs)
	results.Outputs = outputs
	results.States = states

	log(fmt.Sprintf("%#v", results.States))
}

func encodeResults(w io.Writer, runLogs []string, results RunResults,
	description ModelDescription, splitOutputs bool) {
	//  outputs := jsonSafe3D(results.Outputs)
	// results.Outputs := outputs
	// results.States := jsonSafe2D(results.States)

	var overall singleModelResults
	overall.Log = runLogs

	if results.Outputs != nil {
		outputArray := results.Outputs.MustReshape(results.Outputs.Shape()[1:])
		if splitOutputs {
			outputMap := make(map[string]interface{})
			length := outputArray.Len(1)
			for i, output := range description.Outputs {
				singleOutput := outputArray.Slice([]int{i, 0}, []int{1, length}, []int{1, 1}).MustReshape([]int{length})
				outputMap[output] = owjs.JsonSafeArray(singleOutput, 0)
			}
			overall.RunResults.Outputs = outputMap
		} else {
			overall.RunResults.Outputs = owjs.JsonSafeArray(outputArray, 0)
		}
	}

	if results.States != nil {
		stateArray := results.States.MustReshape(results.States.Shape()[1:])
		if splitOutputs {
			stateMap := make(map[string]interface{})
			for i, state := range description.States {
				singleState := stateArray.Get([]int{i})
				stateMap[state] = singleState
			}
			overall.RunResults.States = stateMap
		} else {
			overall.RunResults.States = owjs.JsonSafeArray(stateArray, 0)
		}
	}

	encoder := json.NewEncoder(w)

	err := encoder.Encode(overall)
	if err != nil {
		fmt.Println(overall.Log)
		fmt.Println(err)
	}
}

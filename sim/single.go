package sim

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"

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

func (m singleModel) Initialise() (error, TimeSteppingModel, data.ND3[float64], data.ND2[float64], []string) {
	warnings := make([]string, 1)

	if m.Name == "" {
		return errors.New("No model name provided"), nil, nil, nil, warnings
	}
	factory := Catalog[m.Name]
	if factory == nil {
		return fmt.Errorf("Unknown model: %s", m.Name), nil, nil, nil, warnings
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
	var states data.ND2[float64]
	if len(m.States) == len(desc.States) {
		//TODO: log("TODO. Use supplied states")
		states = model.InitialiseStates(1)
	} else {
		states = model.InitialiseStates(1)
	}
	var inputs data.ND3[float64] = nil
	for i, p := range desc.Inputs {
		thisInput := m.Inputs.Find(p.Name)
		if thisInput == nil {
			warnings = append(warnings, fmt.Sprintf("Missing input: %s, using 0", p))
			continue
		}

		if inputs == nil {
			inputs = data.NewArray3D[float64](1, len(desc.Inputs), len(thisInput))
		}

		inputs.Apply([]int{0, i, 0}, 2, 1, thisInput)
	}

	return nil, model, inputs, states, warnings
}

func uniformParameters(params []float64, n int) data.ND2[float64] {
	res := data.NewArray2D[float64](len(params), n)
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

	// One cell and no performance pressure, so the cap is generous rather than
	// absent: a long run of a misbehaving model would otherwise put one log
	// line per timestep into the response. The per-code totals below stay
	// complete either way.
	var sink *ReportSink
	if reporting, ok := model.(Reporting); ok {
		catalog := reporting.ReportCatalog()
		sink = NewReportSink(1, len(catalog), SevWarning, singleRunDetailLimit)
		reporting.SetReportSink(sink)
	}

	model.Run(inputs, states, outputs)
	results.Outputs = outputs
	results.States = states

	for _, line := range renderReports(model, sink) {
		log(line)
	}
}

// singleRunDetailLimit is how many detail records ow-single reports per code
// before falling back to a total.
const singleRunDetailLimit = 20

// renderReports turns a single run's reports into readable log lines. The
// timestep is reported as an index: a model kernel has no idea what date that
// is, and ow-single has no model file to ask.
func renderReports(model TimeSteppingModel, sink *ReportSink) []string {
	if sink == nil {
		return nil
	}

	reporting, ok := model.(Reporting)
	if !ok {
		return nil
	}
	catalog := reporting.ReportCatalog()

	set := sink.Flush()
	lines := make([]string, 0, set.Len())

	for i := 0; i < set.Len(); i++ {
		code := set.Codes[i]
		if code < 0 || int(code) >= len(catalog) {
			continue
		}
		entry := catalog[code]

		report := Report{NArgs: int32(entry.NArgs)}
		copy(report.Args[:], set.Args[i*MaxReportArgs:(i+1)*MaxReportArgs])

		if ts := set.Timesteps[i]; ts >= 0 {
			lines = append(lines, fmt.Sprintf("%s: timestep %d: %s",
				entry.Severity, ts, entry.Render(report)))
		} else {
			lines = append(lines, fmt.Sprintf("%s: %s", entry.Severity, entry.Render(report)))
		}
	}

	// Counts are complete even where detail records were capped. Say so per
	// code, rather than letting the log imply a condition happened only as
	// often as it was recorded.
	detailsPerCode := make([]uint32, set.NCodes)
	for i := 0; i < set.Len(); i++ {
		if code := set.Codes[i]; code >= 0 && int(code) < set.NCodes {
			detailsPerCode[code]++
		}
	}

	for code := 0; code < set.NCodes && code < len(catalog); code++ {
		total := set.Counts[code]
		if total > detailsPerCode[code] {
			lines = append(lines, fmt.Sprintf("%s: %s occurred %d times in total (%d shown)",
				catalog[code].Severity, catalog[code].Name, total, detailsPerCode[code]))
		}
	}

	return lines
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
				outputMap[output.Name] = owjs.JsonSafeArray(singleOutput, 0)
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
				stateMap[state] = owjs.JsonSafeValue(singleState)
			}
			overall.RunResults.States = stateMap
		} else {
			overall.RunResults.States = owjs.JsonSafeArray(stateArray, 0)
		}
	}

	encoder := json.NewEncoder(w)

	err := encoder.Encode(overall)
	if err != nil {
		log.Error().Stack().Err(err).Any("Overall Log", overall.Log).Msg("")
	}
}

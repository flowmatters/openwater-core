package sim

import (
	"github.com/flowmatters/openwater-core/data"
)

const (
	DIMP_PARAMETER int = 0
	DIMP_CELL      int = 1

	DIMI_CELL     int = 0
	DIMI_INPUT    int = 1
	DIMI_TIMESTEP int = 2

	DIMS_CELL  int = 0
	DIMS_STATE int = 1

	DIMO_CELL     int = 0
	DIMO_OUTPUT   int = 1
	DIMO_TIMESTEP int = 2
)

type ModelDescription struct {
	Parameters []ParameterDescription
	States     []string
	Inputs     []string
	Outputs    []string
}

type ParameterDescription struct {
	Name        string
	Default     float64
	Description string
	Range       [2]float64
	RangeOpen   [2]bool
	Units       string
}

type TimeSteppingModel interface {
	Description() ModelDescription

	ApplyParameters(params data.ND2Float64)
	InitialiseStates(n int) data.ND2Float64
	Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64)
}

type Hotstartable interface {
	GetStates() []float64
	SetStates(states []float64)
}

type Series data.ND1Float64

//type InputSet data.ND2Float64
//type OutputSet data.ND3Float64
//type StateSet data.ND2Float64

type RunResults struct {
	Outputs data.ND3Float64
	States  data.ND2Float64
}

func DescribeParameters(names []string) []ParameterDescription {
	result := make([]ParameterDescription, len(names))
	for i, val := range names {
		result[i] = NewParameter(val)
	}
	return result
}

func DescribeParameter(name string, defaultValue float64, description string,
	paramRange []float64, units string) ParameterDescription {
	var result ParameterDescription
	result.Name = name
	result.Default = defaultValue
	result.Description = description
	result.Range[0] = paramRange[0]
	result.Range[1] = paramRange[1]
	result.Units = units
	return result
}

func NewParameter(name string) ParameterDescription {
	dummyRange := make([]float64, 2)
	return DescribeParameter(name, 0, "", dummyRange, "")
}

func InitialiseOutputs(model TimeSteppingModel, nTimeSteps int, nCells int) data.ND3Float64 {
	return data.NewArray3DFloat64(nCells, len(model.Description().Outputs), nTimeSteps)
}

/*
  How do we want it run?

  * Provide inputs for a given window of time
  * Receive outputs for corresponding window of time
  * Receive initial states, return final states

  * How is time specified? Implicitly?

  * How to specify linked models (conceptually and spatially)
    * Generic? or
    * Specific to a model problem (eg a Dyanmic Sednet stragegy?
  *
*/

package rr

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from src/github.com/flowmatters/openwater-core/models/rr/coeff.go
 * 
 * Don't edit this file. Edit src/github.com/flowmatters/openwater-core/models/rr/coeff.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type RunoffCoefficient struct {
  coeff data.ND1Float64
  
}

func (m *RunoffCoefficient) ApplyParameters(parameters data.ND2Float64) {
  // fmt.Println(parameters)
  // fmt.Println(parameters.Shape())
  nSets := parameters.Len(sim.DIMP_CELL)
  // fmt.Println(nSets)
  m.coeff = parameters.Slice([]int{ 0, 0}, []int{ 1, nSets}, nil).(data.ND1Float64)
  
}


func buildRunoffCoefficient() sim.TimeSteppingModel {
	result := RunoffCoefficient{}
	return &result
}

func init() {
	sim.Catalog["RunoffCoefficient"] = buildRunoffCoefficient
}

func (m *RunoffCoefficient)  Description() sim.ModelDescription{
	var result sim.ModelDescription
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("coeff",0,""),}

  result.Inputs = []string{
  "rainfall",}
  result.Outputs = []string{
  "runoff",}

  result.States = []string{
  }

	return result
}




func (m *RunoffCoefficient) InitialiseStates(n int) data.ND2Float64 {
  // Zero states
	var result = data.NewArray2DFloat64(n,0)

	// for i := 0; i < n; i++ {
  //   stateSet := make(sim.StateSet,0)
  //   

  //   if result==nil {
  //     result = data.NewArray2DFloat64(stateSet.Len(0),n)
  //   }
  //   result.Apply([]int{0,i},[]int{1,1},stateSet)
	// }
 
	return result
}



func (m *RunoffCoefficient) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

  // Loop over all cells
  inputDims := inputs.Shape()
  numCells := states.Len(sim.DIMS_CELL)
  numStates := states.Len(sim.DIMS_STATE)
  numInputSequences := inputs.Len(sim.DIMI_CELL)

  //  fmt.Println("num cells",lenStates,"num states",numStates)
  // fmt.Println("states shape",states.Shape())
  // fmt.Println("states",states) 
  inputLen := inputDims[sim.DIMI_TIMESTEP]
  cellInputsShape := inputDims[1:]
  inputNewShape := []int{inputLen}

  outputPosSlice := outputs.NewIndex(0)
  outputStepSlice := outputs.NewIndex(1)
  outputSizeSlice := outputs.NewIndex(1)
  outputSizeSlice[sim.DIMO_TIMESTEP] = inputLen

  statesPosSlice := states.NewIndex(0)
  statesSizeSlice := states.NewIndex(1)
  statesSizeSlice[sim.DIMS_STATE] = numStates

  inputsPosSlice := inputs.NewIndex(0)
  inputsSizeSlice := inputs.NewIndex(1)
  inputsSizeSlice[sim.DIMI_INPUT] = inputDims[sim.DIMI_INPUT]
  inputsSizeSlice[sim.DIMI_TIMESTEP] = inputLen

//  var result sim.RunResults
//	result.Outputs = data.NewArray3DFloat64( 1, inputLen, numCells)
//	result.States = states  //clone? make([]sim.StateSet, len(states))

  // fmt.Println("Running RunoffCoefficient for ",numCells,"cells")
  for i := 0; i < numCells; i++ {
    outputPosSlice[sim.DIMO_CELL] = i
    statesPosSlice[sim.DIMS_CELL] = i
    inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

    
    // fmt.Println("coeff=",m.coeff)
		coeff := m.coeff.Get1(i%m.coeff.Len1())
    

    // fmt.Println("i",i)
    // fmt.Println("States",states.Shape())
    // fmt.Println("Tmp2",tmp2.Shape())
    

    
    
    

//    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

		cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
//    fmt.Println("cellInputs Shape",cellInputs.Shape())
    
//    fmt.Println("{rainfall mm}",tmpTS.Shape())
		rainfall := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
    

    

    
    
    outputPosSlice[sim.DIMO_OUTPUT] = 0
    runoff := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
    
    

		 runoffCoefficient(rainfall,coeff,runoff)

    
    
    

//		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 1)
    
	}

//	return result
}

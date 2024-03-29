package routing

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/routing/lumpedconstituent.go
 * 
 * Don't edit this file. Edit ./models/routing/lumpedconstituent.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type LumpedConstituentRouting struct {
  X data.ND1Float64
  pointInput data.ND1Float64
  DeltaT data.ND1Float64
  

  
}

func (m *LumpedConstituentRouting) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  var newShape []int
  paramIdx := 0
  paramSize := 1


  paramSize = 1
  newShape = []int{ nSets}

  m.X = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.pointInput = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.DeltaT = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  
}


func buildLumpedConstituentRouting() sim.TimeSteppingModel {
	result := LumpedConstituentRouting{}
	return &result
}

func init() {
	sim.Catalog["LumpedConstituentRouting"] = buildLumpedConstituentRouting
}

func (m *LumpedConstituentRouting)  Description() sim.ModelDescription{
	var result sim.ModelDescription
  
  XDims := []string{
      }
  
  pointInputDims := []string{
      }
  
  DeltaTDims := []string{
      }
  
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("X",0,"Weighting",[]float64{ 0, 1 }," ",XDims),
  sim.DescribeParameter("pointInput",0,"kg.s^-1",[]float64{ 0, 0 },"",pointInputDims),
  sim.DescribeParameter("DeltaT",86400,"Timestep",[]float64{ 1, 86400 }," ",DeltaTDims),}

  result.Inputs = []string{
  "inflowLoad","lateralLoad","outflow","storage",}
  result.Outputs = []string{
  "outflowLoad","pointSourceLoad",}

  result.States = []string{
  "storedMass",}

  result.Dimensions = []string{
      }
	return result
}

func (m *LumpedConstituentRouting) InitialiseDimensions(dims []int) {
  
}

func (m *LumpedConstituentRouting) FindDimensions(parameters data.ND2Float64) []int {
  
  return []int{}
  
}




func (m *LumpedConstituentRouting) InitialiseStates(n int) data.ND2Float64 {
  // Zero states
	var result = data.NewArray2DFloat64(n,1)

	// for i := 0; i < n; i++ {
  //   stateSet := make(sim.StateSet,1)
  //   
	// 	stateSet[0] = 0 // storedMass
  //   

  //   if result==nil {
  //     result = data.NewArray2DFloat64(stateSet.Len(0),n)
  //   }
  //   result.Apply([]int{0,i},[]int{1,1},stateSet)
	// }
 
	return result
}



func (m *LumpedConstituentRouting) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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

//  outputPosSlice := outputs.NewIndex(0)
  outputStepSlice := outputs.NewIndex(1)
  outputSizeSlice := outputs.NewIndex(1)
  outputSizeSlice[sim.DIMO_TIMESTEP] = inputLen

//  statesPosSlice := states.NewIndex(0)
  statesSizeSlice := states.NewIndex(1)
  statesSizeSlice[sim.DIMS_STATE] = numStates

//  inputsPosSlice := inputs.NewIndex(0)
  inputsSizeSlice := inputs.NewIndex(1)
  inputsSizeSlice[sim.DIMI_INPUT] = inputDims[sim.DIMI_INPUT]
  inputsSizeSlice[sim.DIMI_TIMESTEP] = inputLen

//  var result sim.RunResults
//	result.Outputs = data.NewArray3DFloat64( 2, inputLen, numCells)
//	result.States = states  //clone? make([]sim.StateSet, len(states))

  doneChan := make(chan int)
  // fmt.Println("Running LumpedConstituentRouting for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      x := m.X.Get1(i%m.X.Len1())
      pointinput := m.pointInput.Get1(i%m.pointInput.Len1())
      deltat := m.DeltaT.Get1(i%m.DeltaT.Len1())
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      
      initialStates := states.Slice(statesPosSlice,statesSizeSlice,nil).MustReshape([]int{numStates}).(data.ND1Float64)
      

      
      
      storedmass := initialStates.Get1(0)
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{inflowLoad kg.s^-1}",tmpTS.Shape())
      inflowload := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{lateralLoad kg.s^-1}",tmpTS.Shape())
      lateralload := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{outflow m^3.s^-1}",tmpTS.Shape())
      outflow := cellInputs.Slice([]int{ 2,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{storage m^3}",tmpTS.Shape())
      storage := cellInputs.Slice([]int{ 3,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      outflowload := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      pointsourceload := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

      storedmass= LumpedConstituentTransport(inflowload,lateralload,outflow,storage,storedmass,x,pointinput,deltat,outflowload,pointsourceload)

      
      
      initialStates.Set1(0, storedmass)
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 2)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}

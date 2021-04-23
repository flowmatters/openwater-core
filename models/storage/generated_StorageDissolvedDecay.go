package storage

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/storage/dissolved_decay.go
 * 
 * Don't edit this file. Edit ./models/storage/dissolved_decay.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type StorageDissolvedDecay struct {
  DeltaT data.ND1Float64
  doStorageDecay data.ND1Float64
  annualReturnInterval data.ND1Float64
  bankFullFlow data.ND1Float64
  medianFloodResidenceTime data.ND1Float64
  
}

func (m *StorageDissolvedDecay) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  var newShape []int
  paramIdx := 0
  paramSize := 1


  paramSize = 1
  newShape = []int{ nSets}

  m.DeltaT = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.doStorageDecay = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.annualReturnInterval = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.bankFullFlow = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.medianFloodResidenceTime = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  
}


func buildStorageDissolvedDecay() sim.TimeSteppingModel {
	result := StorageDissolvedDecay{}
	return &result
}

func init() {
	sim.Catalog["StorageDissolvedDecay"] = buildStorageDissolvedDecay
}

func (m *StorageDissolvedDecay)  Description() sim.ModelDescription{
	var result sim.ModelDescription
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("DeltaT",86400,"Timestep",[]float64{ 1, 86400 }," "),
  sim.DescribeParameter("doStorageDecay",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("annualReturnInterval",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("bankFullFlow",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("medianFloodResidenceTime",0,"",[]float64{ 0, 0 },""),}

  result.Inputs = []string{
  "inflowMass","inflow","outflow","storageVolume",}
  result.Outputs = []string{
  "decayedMass","outflowMass",}

  result.States = []string{
  "storedMass",}

	return result
}




func (m *StorageDissolvedDecay) InitialiseStates(n int) data.ND2Float64 {
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



func (m *StorageDissolvedDecay) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
  // fmt.Println("Running StorageDissolvedDecay for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      deltat := m.DeltaT.Get1(i%m.DeltaT.Len1())
      dostoragedecay := m.doStorageDecay.Get1(i%m.doStorageDecay.Len1())
      annualreturninterval := m.annualReturnInterval.Get1(i%m.annualReturnInterval.Len1())
      bankfullflow := m.bankFullFlow.Get1(i%m.bankFullFlow.Len1())
      medianfloodresidencetime := m.medianFloodResidenceTime.Get1(i%m.medianFloodResidenceTime.Len1())
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      
      initialStates := states.Slice(statesPosSlice,statesSizeSlice,nil).MustReshape([]int{numStates}).(data.ND1Float64)
      

      
      
      storedmass := initialStates.Get1(0)
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{inflowMass kg.s^-1}",tmpTS.Shape())
      inflowmass := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{inflow m^3.s^-1}",tmpTS.Shape())
      inflow := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{outflow m^3.s^-1}",tmpTS.Shape())
      outflow := cellInputs.Slice([]int{ 2,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{storageVolume m^3}",tmpTS.Shape())
      storagevolume := cellInputs.Slice([]int{ 3,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      decayedmass := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      outflowmass := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

      storedmass= storageDissolvedDecay(inflowmass,inflow,outflow,storagevolume,storedmass,deltat,dostoragedecay,annualreturninterval,bankfullflow,medianfloodresidencetime,decayedmass,outflowmass)

      
      
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

package generation

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/generation/muslefine.go
 * 
 * Don't edit this file. Edit ./models/generation/muslefine.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type CoreUSLE struct {
  area data.ND1Float64
  avK data.ND1Float64
  avLS data.ND1Float64
  avFines data.ND1Float64
  DWC data.ND1Float64
  maxConc data.ND1Float64
  usleHSDRFine data.ND1Float64
  usleHSDRCoarse data.ND1Float64
  timeStepInSeconds data.ND1Float64
  

  
}

func (m *CoreUSLE) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  var newShape []int
  paramIdx := 0
  paramSize := 1


  paramSize = 1
  newShape = []int{ nSets}

  m.area = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.avK = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.avLS = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.avFines = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.DWC = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.maxConc = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.usleHSDRFine = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.usleHSDRCoarse = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.timeStepInSeconds = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  
}


func buildCoreUSLE() sim.TimeSteppingModel {
	result := CoreUSLE{}
	return &result
}

func init() {
	sim.Catalog["CoreUSLE"] = buildCoreUSLE
}

func (m *CoreUSLE)  Description() sim.ModelDescription{
	var result sim.ModelDescription
  
  areaDims := []string{
      }
  
  avKDims := []string{
      }
  
  avLSDims := []string{
      }
  
  avFinesDims := []string{
      }
  
  DWCDims := []string{
      }
  
  maxConcDims := []string{
      }
  
  usleHSDRFineDims := []string{
      }
  
  usleHSDRCoarseDims := []string{
      }
  
  timeStepInSecondsDims := []string{
      }
  
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("area",0,"[0",[]float64{ 0, 0 },"",areaDims),
  sim.DescribeParameter("avK",0,"",[]float64{ 0, 0 },"",avKDims),
  sim.DescribeParameter("avLS",0,"",[]float64{ 0, 0 },"",avLSDims),
  sim.DescribeParameter("avFines",0,"% of fine sediment in soil",[]float64{ 0, 0 },"",avFinesDims),
  sim.DescribeParameter("DWC",0,"Dry Weather Concentration",[]float64{ 0.1, 10000 }," ",DWCDims),
  sim.DescribeParameter("maxConc",0,"mg.L^-1 USLE Maximum Fine Sediment Allowable Runoff Concentration",[]float64{ 0, 10000 },"",maxConcDims),
  sim.DescribeParameter("usleHSDRFine",0,"% Hillslope Fine Sediment Delivery Ratio",[]float64{ 0, 100 },"",usleHSDRFineDims),
  sim.DescribeParameter("usleHSDRCoarse",0,"% Hillslope Coarse Sediment Delivery Ratio",[]float64{ 0, 100 },"",usleHSDRCoarseDims),
  sim.DescribeParameter("timeStepInSeconds",86400,"s Duration of timestep in seconds",[]float64{ 0, 1e+08 },"",timeStepInSecondsDims),}

  result.Inputs = []string{
  "rFactor","quickflow","baseflow","cFactor",}
  result.Outputs = []string{
  "quickLoadFine","slowLoadFine","quickLoadCoarse","slowLoadCoarse","totalFineLoad","totalCoarseLoad","generatedLoadFine","generatedLoadCoarse",}

  result.States = []string{
  }

  result.Dimensions = []string{
      }
	return result
}

func (m *CoreUSLE) InitialiseDimensions(dims []int) {
  
}

func (m *CoreUSLE) FindDimensions(parameters data.ND2Float64) []int {
  
  return []int{}
  
}




func (m *CoreUSLE) InitialiseStates(n int) data.ND2Float64 {
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



func (m *CoreUSLE) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
//	result.Outputs = data.NewArray3DFloat64( 8, inputLen, numCells)
//	result.States = states  //clone? make([]sim.StateSet, len(states))

  doneChan := make(chan int)
  // fmt.Println("Running CoreUSLE for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      area := m.area.Get1(i%m.area.Len1())
      avk := m.avK.Get1(i%m.avK.Len1())
      avls := m.avLS.Get1(i%m.avLS.Len1())
      avfines := m.avFines.Get1(i%m.avFines.Len1())
      dwc := m.DWC.Get1(i%m.DWC.Len1())
      maxconc := m.maxConc.Get1(i%m.maxConc.Len1())
      uslehsdrfine := m.usleHSDRFine.Get1(i%m.usleHSDRFine.Len1())
      uslehsdrcoarse := m.usleHSDRCoarse.Get1(i%m.usleHSDRCoarse.Len1())
      timestepinseconds := m.timeStepInSeconds.Get1(i%m.timeStepInSeconds.Len1())
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      

      
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{rFactor unitless}",tmpTS.Shape())
      rfactor := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{quickflow m^3.s^-1}",tmpTS.Shape())
      quickflow := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{baseflow m^3.s^-}",tmpTS.Shape())
      baseflow := cellInputs.Slice([]int{ 2,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{cFactor [0,1] Cover Factor}",tmpTS.Shape())
      cfactor := cellInputs.Slice([]int{ 3,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      quickloadfine := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      slowloadfine := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 2
      quickloadcoarse := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 3
      slowloadcoarse := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 4
      totalfineload := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 5
      totalcoarseload := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 6
      generatedloadfine := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 7
      generatedloadcoarse := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

       coreUSLE(rfactor,quickflow,baseflow,cfactor,area,avk,avls,avfines,dwc,maxconc,uslehsdrfine,uslehsdrcoarse,timestepinseconds,quickloadfine,slowloadfine,quickloadcoarse,slowloadcoarse,totalfineload,totalcoarseload,generatedloadfine,generatedloadcoarse)

      
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 8)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}

package generation

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/generation/uslefine.go
 * 
 * Don't edit this file. Edit ./models/generation/uslefine.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type USLEFineSedimentGeneration struct {
  S data.ND1Float64
  P data.ND1Float64
  RainThreshold data.ND1Float64
  Alpha data.ND1Float64
  Beta data.ND1Float64
  Eta data.ND1Float64
  A1 data.ND1Float64
  A2 data.ND1Float64
  A3 data.ND1Float64
  DWC data.ND1Float64
  avK data.ND1Float64
  avLS data.ND1Float64
  avFines data.ND1Float64
  area data.ND1Float64
  maxConc data.ND1Float64
  usleHSDRFine data.ND1Float64
  usleHSDRCoarse data.ND1Float64
  timeStepInSeconds data.ND1Float64
  
}

func (m *USLEFineSedimentGeneration) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  var newShape []int
  paramIdx := 0
  paramSize := 1


  paramSize = 1
  newShape = []int{ nSets}

  m.S = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.P = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.RainThreshold = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.Alpha = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.Beta = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.Eta = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.A1 = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.A2 = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.A3 = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.DWC = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
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

  m.area = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
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


func buildUSLEFineSedimentGeneration() sim.TimeSteppingModel {
	result := USLEFineSedimentGeneration{}
	return &result
}

func init() {
	sim.Catalog["USLEFineSedimentGeneration"] = buildUSLEFineSedimentGeneration
}

func (m *USLEFineSedimentGeneration)  Description() sim.ModelDescription{
	var result sim.ModelDescription
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("S",0,"mm Mean Summer Rainfall",[]float64{ 0, 5000 },""),
  sim.DescribeParameter("P",0,"mm Mean Annual Rainfall",[]float64{ 0, 5000 },""),
  sim.DescribeParameter("RainThreshold",0,"mm R Factor Rainfall Threshold",[]float64{ 0, 12.7 },""),
  sim.DescribeParameter("Alpha",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("Beta",0,"Monthly EI30 Parameter",[]float64{ 0.1, 10 }," "),
  sim.DescribeParameter("Eta",0,"Monthly EI30 Parameter",[]float64{ 0.1, 10 }," "),
  sim.DescribeParameter("A1",0,"Alpha term 1",[]float64{ 0.001, 10 }," "),
  sim.DescribeParameter("A2",0,"Alpha term 2",[]float64{ 0.001, 10 }," "),
  sim.DescribeParameter("A3",0,"Alpha term 3",[]float64{ 0.001, 100 }," "),
  sim.DescribeParameter("DWC",0,"Dry Weather Concentration",[]float64{ 0.1, 10000 }," "),
  sim.DescribeParameter("avK",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("avLS",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("avFines",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("area",0,"[0",[]float64{ 0, 0 },""),
  sim.DescribeParameter("maxConc",0,"mg.L^-1 USLE Maximum Fine Sediment Allowable Runoff Concentration",[]float64{ 0, 10000 },""),
  sim.DescribeParameter("usleHSDRFine",0,"% Hillslope Fine Sediment Delivery Ratio",[]float64{ 0, 100 },""),
  sim.DescribeParameter("usleHSDRCoarse",0,"% Hillslope Coarse Sediment Delivery Ratio",[]float64{ 0, 100 },""),
  sim.DescribeParameter("timeStepInSeconds",86400,"s Duration of timestep in seconds",[]float64{ 0, 1e+08 },""),}

  result.Inputs = []string{
  "quickflow","baseflow","rainfall","KLSC","KLSC_Fine","CovOrCFact","dayOfYear",}
  result.Outputs = []string{
  "quickLoadFine","slowLoadFine","quickLoadCoarse","slowLoadCoarse","totalLoad",}

  result.States = []string{
  }

	return result
}




func (m *USLEFineSedimentGeneration) InitialiseStates(n int) data.ND2Float64 {
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



func (m *USLEFineSedimentGeneration) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
//	result.Outputs = data.NewArray3DFloat64( 5, inputLen, numCells)
//	result.States = states  //clone? make([]sim.StateSet, len(states))

  doneChan := make(chan int)
  // fmt.Println("Running USLEFineSedimentGeneration for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      s := m.S.Get1(i%m.S.Len1())
      p := m.P.Get1(i%m.P.Len1())
      rainthreshold := m.RainThreshold.Get1(i%m.RainThreshold.Len1())
      alpha := m.Alpha.Get1(i%m.Alpha.Len1())
      beta := m.Beta.Get1(i%m.Beta.Len1())
      eta := m.Eta.Get1(i%m.Eta.Len1())
      a1 := m.A1.Get1(i%m.A1.Len1())
      a2 := m.A2.Get1(i%m.A2.Len1())
      a3 := m.A3.Get1(i%m.A3.Len1())
      dwc := m.DWC.Get1(i%m.DWC.Len1())
      avk := m.avK.Get1(i%m.avK.Len1())
      avls := m.avLS.Get1(i%m.avLS.Len1())
      avfines := m.avFines.Get1(i%m.avFines.Len1())
      area := m.area.Get1(i%m.area.Len1())
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
      
  //    fmt.Println("{quickflow m^3.s^-1}",tmpTS.Shape())
      quickflow := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{baseflow m^3.s^-}",tmpTS.Shape())
      baseflow := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{rainfall mm}",tmpTS.Shape())
      rainfall := cellInputs.Slice([]int{ 2,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{KLSC [0,100000000] KLSC}",tmpTS.Shape())
      klsc := cellInputs.Slice([]int{ 3,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{KLSC_Fine [0,100000000] KLSC}",tmpTS.Shape())
      klsc_fine := cellInputs.Slice([]int{ 4,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{CovOrCFact [] Average C Factor}",tmpTS.Shape())
      covorcfact := cellInputs.Slice([]int{ 5,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{dayOfYear dayOfYear}",tmpTS.Shape())
      dayofyear := cellInputs.Slice([]int{ 6,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      quickloadfine := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      slowloadfine := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 2
      quickloadcoarse := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 3
      slowloadcoarse := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 4
      totalload := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

       usleFine(quickflow,baseflow,rainfall,klsc,klsc_fine,covorcfact,dayofyear,s,p,rainthreshold,alpha,beta,eta,a1,a2,a3,dwc,avk,avls,avfines,area,maxconc,uslehsdrfine,uslehsdrcoarse,timestepinseconds,quickloadfine,slowloadfine,quickloadcoarse,slowloadcoarse,totalload)

      
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 5)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}

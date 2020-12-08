package routing

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/routing/instream_fine_sediment.go
 * 
 * Don't edit this file. Edit ./models/routing/instream_fine_sediment.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type InstreamFineSediment struct {
  bankFullFlow data.ND1Float64
  fineSedSettVelocityFlood data.ND1Float64
  floodPlainArea data.ND1Float64
  linkWidth data.ND1Float64
  linkLength data.ND1Float64
  linkSlope data.ND1Float64
  bankHeight data.ND1Float64
  propBankHeightForFineDep data.ND1Float64
  sedBulkDensity data.ND1Float64
  manningsN data.ND1Float64
  fineSedSettVelocity data.ND1Float64
  fineSedReMobVelocity data.ND1Float64
  durationInSeconds data.ND1Float64
  
}

func (m *InstreamFineSediment) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  var newShape []int
  paramIdx := 0
  paramSize := 1


  paramSize = 1
  newShape = []int{ nSets}

  m.bankFullFlow = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.fineSedSettVelocityFlood = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.floodPlainArea = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.linkWidth = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.linkLength = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.linkSlope = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.bankHeight = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.propBankHeightForFineDep = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.sedBulkDensity = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.manningsN = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.fineSedSettVelocity = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.fineSedReMobVelocity = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.durationInSeconds = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  
}


func buildInstreamFineSediment() sim.TimeSteppingModel {
	result := InstreamFineSediment{}
	return &result
}

func init() {
	sim.Catalog["InstreamFineSediment"] = buildInstreamFineSediment
}

func (m *InstreamFineSediment)  Description() sim.ModelDescription{
	var result sim.ModelDescription
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("bankFullFlow",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("fineSedSettVelocityFlood",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("floodPlainArea",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("linkWidth",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("linkLength",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("linkSlope",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("bankHeight",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("propBankHeightForFineDep",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("sedBulkDensity",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("manningsN",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("fineSedSettVelocity",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("fineSedReMobVelocity",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("durationInSeconds",86400,"Timestep",[]float64{ 1, 86400 }," "),}

  result.Inputs = []string{
  "incomingMass","reachVolume","outflow",}
  result.Outputs = []string{
  "loadDownstream","loadToFloodplain","floodplainDepositionFraction","channelDepositionFraction",}

  result.States = []string{
  "channelStoreFine","totalStoredMass",}

	return result
}




func (m *InstreamFineSediment) InitialiseStates(n int) data.ND2Float64 {
  // Zero states
	var result = data.NewArray2DFloat64(n,2)

	// for i := 0; i < n; i++ {
  //   stateSet := make(sim.StateSet,2)
  //   
	// 	stateSet[0] = 0 // channelStoreFine
  //   
	// 	stateSet[1] = 0 // totalStoredMass
  //   

  //   if result==nil {
  //     result = data.NewArray2DFloat64(stateSet.Len(0),n)
  //   }
  //   result.Apply([]int{0,i},[]int{1,1},stateSet)
	// }
 
	return result
}



func (m *InstreamFineSediment) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
//	result.Outputs = data.NewArray3DFloat64( 4, inputLen, numCells)
//	result.States = states  //clone? make([]sim.StateSet, len(states))

  doneChan := make(chan int)
  // fmt.Println("Running InstreamFineSediment for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      bankfullflow := m.bankFullFlow.Get1(i%m.bankFullFlow.Len1())
      finesedsettvelocityflood := m.fineSedSettVelocityFlood.Get1(i%m.fineSedSettVelocityFlood.Len1())
      floodplainarea := m.floodPlainArea.Get1(i%m.floodPlainArea.Len1())
      linkwidth := m.linkWidth.Get1(i%m.linkWidth.Len1())
      linklength := m.linkLength.Get1(i%m.linkLength.Len1())
      linkslope := m.linkSlope.Get1(i%m.linkSlope.Len1())
      bankheight := m.bankHeight.Get1(i%m.bankHeight.Len1())
      propbankheightforfinedep := m.propBankHeightForFineDep.Get1(i%m.propBankHeightForFineDep.Len1())
      sedbulkdensity := m.sedBulkDensity.Get1(i%m.sedBulkDensity.Len1())
      manningsn := m.manningsN.Get1(i%m.manningsN.Len1())
      finesedsettvelocity := m.fineSedSettVelocity.Get1(i%m.fineSedSettVelocity.Len1())
      finesedremobvelocity := m.fineSedReMobVelocity.Get1(i%m.fineSedReMobVelocity.Len1())
      durationinseconds := m.durationInSeconds.Get1(i%m.durationInSeconds.Len1())
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      
      initialStates := states.Slice(statesPosSlice,statesSizeSlice,nil).MustReshape([]int{numStates}).(data.ND1Float64)
      

      
      
      channelstorefine := initialStates.Get1(0)
      
      totalstoredmass := initialStates.Get1(1)
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{incomingMass <nil>}",tmpTS.Shape())
      incomingmass := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{reachVolume <nil>}",tmpTS.Shape())
      reachvolume := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{outflow <nil>}",tmpTS.Shape())
      outflow := cellInputs.Slice([]int{ 2,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      loaddownstream := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      loadtofloodplain := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 2
      floodplaindepositionfraction := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 3
      channeldepositionfraction := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

      channelstorefine,totalstoredmass= instreamFineSediment(incomingmass,reachvolume,outflow,channelstorefine,totalstoredmass,bankfullflow,finesedsettvelocityflood,floodplainarea,linkwidth,linklength,linkslope,bankheight,propbankheightforfinedep,sedbulkdensity,manningsn,finesedsettvelocity,finesedremobvelocity,durationinseconds,loaddownstream,loadtofloodplain,floodplaindepositionfraction,channeldepositionfraction)

      
      
      initialStates.Set1(0, channelstorefine)
      
      initialStates.Set1(1, totalstoredmass)
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 4)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}

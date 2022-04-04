package routing

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/routing/instream_particulate_nutrient.go
 * 
 * Don't edit this file. Edit ./models/routing/instream_particulate_nutrient.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type InstreamParticulateNutrient struct {
  particulateNutrientConcentration data.ND1Float64
  soilPercentFine data.ND1Float64
  durationInSeconds data.ND1Float64
  

  
}

func (m *InstreamParticulateNutrient) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  var newShape []int
  paramIdx := 0
  paramSize := 1


  paramSize = 1
  newShape = []int{ nSets}

  m.particulateNutrientConcentration = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.soilPercentFine = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.durationInSeconds = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  
}


func buildInstreamParticulateNutrient() sim.TimeSteppingModel {
	result := InstreamParticulateNutrient{}
	return &result
}

func init() {
	sim.Catalog["InstreamParticulateNutrient"] = buildInstreamParticulateNutrient
}

func (m *InstreamParticulateNutrient)  Description() sim.ModelDescription{
	var result sim.ModelDescription
  
  particulateNutrientConcentrationDims := []string{
      }
  
  soilPercentFineDims := []string{
      }
  
  durationInSecondsDims := []string{
      }
  
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("particulateNutrientConcentration",0,"Proportion of sediment mass",[]float64{ 0, 1 }," ",particulateNutrientConcentrationDims),
  sim.DescribeParameter("soilPercentFine",0,"",[]float64{ 0, 0 },"",soilPercentFineDims),
  sim.DescribeParameter("durationInSeconds",86400,"Timestep",[]float64{ 1, 86400 }," ",durationInSecondsDims),}

  result.Inputs = []string{
  "incomingMassUpstream","incomingMassLateral","reachVolume","outflow","streambankErosion","lateralSediment","floodplainDepositionFraction","channelDepositionFraction",}
  result.Outputs = []string{
  "loadDownstream","loadToFloodplain",}

  result.States = []string{
  "instreamStoredMass","channelStoredMass",}

  result.Dimensions = []string{
      }
	return result
}

func (m *InstreamParticulateNutrient) InitialiseDimensions(dims []int) {
  
}

func (m *InstreamParticulateNutrient) FindDimensions(parameters data.ND2Float64) []int {
  
  return []int{}
  
}




func (m *InstreamParticulateNutrient) InitialiseStates(n int) data.ND2Float64 {
  // Zero states
	var result = data.NewArray2DFloat64(n,2)

	// for i := 0; i < n; i++ {
  //   stateSet := make(sim.StateSet,2)
  //   
	// 	stateSet[0] = 0 // instreamStoredMass
  //   
	// 	stateSet[1] = 0 // channelStoredMass
  //   

  //   if result==nil {
  //     result = data.NewArray2DFloat64(stateSet.Len(0),n)
  //   }
  //   result.Apply([]int{0,i},[]int{1,1},stateSet)
	// }
 
	return result
}



func (m *InstreamParticulateNutrient) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
  // fmt.Println("Running InstreamParticulateNutrient for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      particulatenutrientconcentration := m.particulateNutrientConcentration.Get1(i%m.particulateNutrientConcentration.Len1())
      soilpercentfine := m.soilPercentFine.Get1(i%m.soilPercentFine.Len1())
      durationinseconds := m.durationInSeconds.Get1(i%m.durationInSeconds.Len1())
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      
      initialStates := states.Slice(statesPosSlice,statesSizeSlice,nil).MustReshape([]int{numStates}).(data.ND1Float64)
      

      
      
      instreamstoredmass := initialStates.Get1(0)
      
      channelstoredmass := initialStates.Get1(1)
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{incomingMassUpstream <nil>}",tmpTS.Shape())
      incomingmassupstream := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{incomingMassLateral <nil>}",tmpTS.Shape())
      incomingmasslateral := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{reachVolume <nil>}",tmpTS.Shape())
      reachvolume := cellInputs.Slice([]int{ 2,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{outflow <nil>}",tmpTS.Shape())
      outflow := cellInputs.Slice([]int{ 3,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{streambankErosion <nil>}",tmpTS.Shape())
      streambankerosion := cellInputs.Slice([]int{ 4,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{lateralSediment <nil>}",tmpTS.Shape())
      lateralsediment := cellInputs.Slice([]int{ 5,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{floodplainDepositionFraction <nil>}",tmpTS.Shape())
      floodplaindepositionfraction := cellInputs.Slice([]int{ 6,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{channelDepositionFraction <nil>}",tmpTS.Shape())
      channeldepositionfraction := cellInputs.Slice([]int{ 7,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      loaddownstream := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      loadtofloodplain := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

      instreamstoredmass,channelstoredmass= instreamParticulateNutrient(incomingmassupstream,incomingmasslateral,reachvolume,outflow,streambankerosion,lateralsediment,floodplaindepositionfraction,channeldepositionfraction,instreamstoredmass,channelstoredmass,particulatenutrientconcentration,soilpercentfine,durationinseconds,loaddownstream,loadtofloodplain)

      
      
      initialStates.Set1(0, instreamstoredmass)
      
      initialStates.Set1(1, channelstoredmass)
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 2)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}

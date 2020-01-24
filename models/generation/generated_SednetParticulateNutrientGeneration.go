package generation

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/generation/particulate_nutrients.go
 * 
 * Don't edit this file. Edit ./models/generation/particulate_nutrients.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type SednetParticulateNutrientGeneration struct {
  area data.ND1Float64
  nutSurfSoilConc data.ND1Float64
  hillDeliveryRatio data.ND1Float64
  Nutrient_Enrichment_Ratio data.ND1Float64
  nutSubSoilConc data.ND1Float64
  Nutrient_Enrichment_Ratio_Gully data.ND1Float64
  gullyDeliveryRatio data.ND1Float64
  nutrientDWC data.ND1Float64
  Do_P_CREAMS_Enrichment data.ND1Float64
  
}

func (m *SednetParticulateNutrientGeneration) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  newShape := []int{nSets}

  m.area = parameters.Slice([]int{ 0, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  m.nutSurfSoilConc = parameters.Slice([]int{ 1, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  m.hillDeliveryRatio = parameters.Slice([]int{ 2, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  m.Nutrient_Enrichment_Ratio = parameters.Slice([]int{ 3, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  m.nutSubSoilConc = parameters.Slice([]int{ 4, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  m.Nutrient_Enrichment_Ratio_Gully = parameters.Slice([]int{ 5, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  m.gullyDeliveryRatio = parameters.Slice([]int{ 6, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  m.nutrientDWC = parameters.Slice([]int{ 7, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  m.Do_P_CREAMS_Enrichment = parameters.Slice([]int{ 8, 0}, []int{ 1, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  
}


func buildSednetParticulateNutrientGeneration() sim.TimeSteppingModel {
	result := SednetParticulateNutrientGeneration{}
	return &result
}

func init() {
	sim.Catalog["SednetParticulateNutrientGeneration"] = buildSednetParticulateNutrientGeneration
}

func (m *SednetParticulateNutrientGeneration)  Description() sim.ModelDescription{
	var result sim.ModelDescription
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("area",0,"m^2",[]float64{ 0, 0 },""),
  sim.DescribeParameter("nutSurfSoilConc",0,"kg.kg^-1",[]float64{ 0, 0 },""),
  sim.DescribeParameter("hillDeliveryRatio",0,"%",[]float64{ 0, 0 },""),
  sim.DescribeParameter("Nutrient_Enrichment_Ratio",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("nutSubSoilConc",0,"kg.kg^-1",[]float64{ 0, 0 },""),
  sim.DescribeParameter("Nutrient_Enrichment_Ratio_Gully",0,"",[]float64{ 0, 0 },""),
  sim.DescribeParameter("gullyDeliveryRatio",0,"%",[]float64{ 0, 0 },""),
  sim.DescribeParameter("nutrientDWC",0,"mg.L^-1",[]float64{ 0, 0 },""),
  sim.DescribeParameter("Do_P_CREAMS_Enrichment",0,"flag",[]float64{ 0, 0 },""),}

  result.Inputs = []string{
  "fineSedModelFineSheetGeneratedKg","fineSedModelCoarseSheetGeneratedKg","fineSedModelFineGullyGeneratedKg","fineSedModelCoarseGullyGeneratedKg","slowflow",}
  result.Outputs = []string{
  "quickflowConstituent","slowflowConstituent","totalLoad",}

  result.States = []string{
  }

	return result
}




func (m *SednetParticulateNutrientGeneration) InitialiseStates(n int) data.ND2Float64 {
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



func (m *SednetParticulateNutrientGeneration) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
//	result.Outputs = data.NewArray3DFloat64( 3, inputLen, numCells)
//	result.States = states  //clone? make([]sim.StateSet, len(states))

  doneChan := make(chan int)
  // fmt.Println("Running SednetParticulateNutrientGeneration for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      
      // fmt.Println("area=",m.area)
      area := m.area.Get1(i%m.area.Len1())
      
      // fmt.Println("nutSurfSoilConc=",m.nutSurfSoilConc)
      nutsurfsoilconc := m.nutSurfSoilConc.Get1(i%m.nutSurfSoilConc.Len1())
      
      // fmt.Println("hillDeliveryRatio=",m.hillDeliveryRatio)
      hilldeliveryratio := m.hillDeliveryRatio.Get1(i%m.hillDeliveryRatio.Len1())
      
      // fmt.Println("Nutrient_Enrichment_Ratio=",m.Nutrient_Enrichment_Ratio)
      nutrient_enrichment_ratio := m.Nutrient_Enrichment_Ratio.Get1(i%m.Nutrient_Enrichment_Ratio.Len1())
      
      // fmt.Println("nutSubSoilConc=",m.nutSubSoilConc)
      nutsubsoilconc := m.nutSubSoilConc.Get1(i%m.nutSubSoilConc.Len1())
      
      // fmt.Println("Nutrient_Enrichment_Ratio_Gully=",m.Nutrient_Enrichment_Ratio_Gully)
      nutrient_enrichment_ratio_gully := m.Nutrient_Enrichment_Ratio_Gully.Get1(i%m.Nutrient_Enrichment_Ratio_Gully.Len1())
      
      // fmt.Println("gullyDeliveryRatio=",m.gullyDeliveryRatio)
      gullydeliveryratio := m.gullyDeliveryRatio.Get1(i%m.gullyDeliveryRatio.Len1())
      
      // fmt.Println("nutrientDWC=",m.nutrientDWC)
      nutrientdwc := m.nutrientDWC.Get1(i%m.nutrientDWC.Len1())
      
      // fmt.Println("Do_P_CREAMS_Enrichment=",m.Do_P_CREAMS_Enrichment)
      do_p_creams_enrichment := m.Do_P_CREAMS_Enrichment.Get1(i%m.Do_P_CREAMS_Enrichment.Len1())
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      

      
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{fineSedModelFineSheetGeneratedKg <nil>}",tmpTS.Shape())
      finesedmodelfinesheetgeneratedkg := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{fineSedModelCoarseSheetGeneratedKg <nil>}",tmpTS.Shape())
      finesedmodelcoarsesheetgeneratedkg := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{fineSedModelFineGullyGeneratedKg <nil>}",tmpTS.Shape())
      finesedmodelfinegullygeneratedkg := cellInputs.Slice([]int{ 2,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{fineSedModelCoarseGullyGeneratedKg <nil>}",tmpTS.Shape())
      finesedmodelcoarsegullygeneratedkg := cellInputs.Slice([]int{ 3,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{slowflow m^3.s^-1}",tmpTS.Shape())
      slowflow := cellInputs.Slice([]int{ 4,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      quickflowconstituent := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      slowflowconstituent := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 2
      totalload := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

       particulateNutrients(finesedmodelfinesheetgeneratedkg,finesedmodelcoarsesheetgeneratedkg,finesedmodelfinegullygeneratedkg,finesedmodelcoarsegullygeneratedkg,slowflow,area,nutsurfsoilconc,hilldeliveryratio,nutrient_enrichment_ratio,nutsubsoilconc,nutrient_enrichment_ratio_gully,gullydeliveryratio,nutrientdwc,do_p_creams_enrichment,quickflowconstituent,slowflowconstituent,totalload)

      
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 3)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}

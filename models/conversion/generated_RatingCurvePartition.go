package conversion

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/conversion/rating_partition.go
 * 
 * Don't edit this file. Edit ./models/conversion/rating_partition.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type RatingCurvePartition struct {
  nPts data.ND1Float64
  inputAmount data.ND2Float64
  proportion data.ND2Float64
  

  maxnPts int
  
}

func (m *RatingCurvePartition) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  var newShape []int
  paramIdx := 0
  paramSize := 1


  paramSize = 1
  newShape = []int{ nSets}

  m.nPts = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1* m.maxnPts
  newShape = []int{m.maxnPts, nSets}

  m.inputAmount = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND2Float64)
  paramIdx += paramSize

  paramSize = 1* m.maxnPts
  newShape = []int{m.maxnPts, nSets}

  m.proportion = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND2Float64)
  paramIdx += paramSize

  
}


func buildRatingCurvePartition() sim.TimeSteppingModel {
	result := RatingCurvePartition{}
	return &result
}

func init() {
	sim.Catalog["RatingCurvePartition"] = buildRatingCurvePartition
}

func (m *RatingCurvePartition)  Description() sim.ModelDescription{
	var result sim.ModelDescription
  
  nPtsDims := []string{
      }
  
  inputAmountDims := []string{
    "nPts",  }
  
  proportionDims := []string{
    "nPts",  }
  
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("nPts",0,"",[]float64{ 0, 0 },"",nPtsDims),
  sim.DescribeParameter("inputAmount",0,"",[]float64{ 0, 0 },"",inputAmountDims),
  sim.DescribeParameter("proportion",0,"",[]float64{ 0, 0 },"",proportionDims),}

  result.Inputs = []string{
  "input",}
  result.Outputs = []string{
  "output1","output2",}

  result.States = []string{
  }

  result.Dimensions = []string{
    "nPts",  }
	return result
}

func (m *RatingCurvePartition) InitialiseDimensions(dims []int) {
  m.maxnPts = dims[0]
  
}

func (m *RatingCurvePartition) FindDimensions(parameters data.ND2Float64) []int {
  
  nSets := parameters.Len(sim.DIMP_CELL)
  paramIdx := 0
  paramSize := 1
  maxValues := make(map[string]float64)

  paramSize = 1

  maxValues["nPts"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  paramSize = 1* int(maxValues["nPts"])

  maxValues["inputAmount"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  paramSize = 1* int(maxValues["nPts"])

  maxValues["proportion"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  
  dims := []int{
    int(maxValues["nPts"]),
    
  }
  
  return dims
  
}




func (m *RatingCurvePartition) InitialiseStates(n int) data.ND2Float64 {
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



func (m *RatingCurvePartition) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
  // fmt.Println("Running RatingCurvePartition for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      // Dimension parameter
      npts := int(m.nPts.Get1(i%m.nPts.Len1()))
      inputamountShape := m.inputAmount.Shape()
      inputamountNSets := inputamountShape[len(inputamountShape)-1]
      inputamountFrom := []int{  0,  i%inputamountNSets }
      inputamountSliceShape := []int{  npts,  }
      // WAS inputamountSliceShape := []int{  m.maxnPts,  }
      inputamount := m.inputAmount.Slice(inputamountFrom, inputamountSliceShape, nil).(data.ND1Float64)
      proportionShape := m.proportion.Shape()
      proportionNSets := proportionShape[len(proportionShape)-1]
      proportionFrom := []int{  0,  i%proportionNSets }
      proportionSliceShape := []int{  npts,  }
      // WAS proportionSliceShape := []int{  m.maxnPts,  }
      proportion := m.proportion.Slice(proportionFrom, proportionSliceShape, nil).(data.ND1Float64)
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      

      
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{input <nil>}",tmpTS.Shape())
      input := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      output1 := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      output2 := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

       ratingPartition(input,npts,inputamount,proportion,output1,output2)

      
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 2)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}

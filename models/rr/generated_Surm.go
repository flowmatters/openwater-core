package rr

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/rr/surm.go
 * 
 * Don't edit this file. Edit ./models/rr/surm.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type Surm struct {
  bfac data.ND1Float64
  coeff data.ND1Float64
  dseep data.ND1Float64
  fcFrac data.ND1Float64
  fimp data.ND1Float64
  rfac data.ND1Float64
  smax data.ND1Float64
  sq data.ND1Float64
  thres data.ND1Float64
  

  
}

func (m *Surm) ApplyParameters(parameters data.ND2Float64) {

  nSets := parameters.Len(sim.DIMP_CELL)
  var newShape []int
  paramIdx := 0
  paramSize := 1


  paramSize = 1
  newShape = []int{ nSets}

  m.bfac = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.coeff = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.dseep = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.fcFrac = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.fimp = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.rfac = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.smax = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.sq = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1
  newShape = []int{ nSets}

  m.thres = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  
}


func buildSurm() sim.TimeSteppingModel {
	result := Surm{}
	return &result
}

func init() {
	sim.Catalog["Surm"] = buildSurm
}

func (m *Surm)  Description() sim.ModelDescription{
	var result sim.ModelDescription
  
  bfacDims := []string{
      }
  
  coeffDims := []string{
      }
  
  dseepDims := []string{
      }
  
  fcFracDims := []string{
      }
  
  fimpDims := []string{
      }
  
  rfacDims := []string{
      }
  
  smaxDims := []string{
      }
  
  sqDims := []string{
      }
  
  thresDims := []string{
      }
  
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("bfac",0,"",[]float64{ 0, 0 },"",bfacDims),
  sim.DescribeParameter("coeff",0,"",[]float64{ 0, 0 },"",coeffDims),
  sim.DescribeParameter("dseep",0,"",[]float64{ 0, 0 },"",dseepDims),
  sim.DescribeParameter("fcFrac",0,"",[]float64{ 0, 0 },"",fcFracDims),
  sim.DescribeParameter("fimp",0,"",[]float64{ 0, 0 },"",fimpDims),
  sim.DescribeParameter("rfac",0,"",[]float64{ 0, 0 },"",rfacDims),
  sim.DescribeParameter("smax",0,"",[]float64{ 0, 0 },"",smaxDims),
  sim.DescribeParameter("sq",0,"",[]float64{ 0, 0 },"",sqDims),
  sim.DescribeParameter("thres",0,"",[]float64{ 0, 0 },"",thresDims),}

  result.Inputs = []string{
  "rainfall","pet",}
  result.Outputs = []string{
  "runoff","quickflow","baseflow","store",}

  result.States = []string{
  "SoilMoistureStore","Groundwater","TotalStore",}

  result.Dimensions = []string{
      }
	return result
}

func (m *Surm) InitialiseDimensions(dims []int) {
  
}

func (m *Surm) FindDimensions(parameters data.ND2Float64) []int {
  
  return []int{}
  
}




func (m *Surm) InitialiseStates(n int) data.ND2Float64 {
  // Zero states
	var result = data.NewArray2DFloat64(n,3)

	// for i := 0; i < n; i++ {
  //   stateSet := make(sim.StateSet,3)
  //   
	// 	stateSet[0] = 0 // SoilMoistureStore
  //   
	// 	stateSet[1] = 0 // Groundwater
  //   
	// 	stateSet[2] = 0 // TotalStore
  //   

  //   if result==nil {
  //     result = data.NewArray2DFloat64(stateSet.Len(0),n)
  //   }
  //   result.Apply([]int{0,i},[]int{1,1},stateSet)
	// }
 
	return result
}



func (m *Surm) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
  // fmt.Println("Running Surm for ",numCells,"cells")
//  for i := 0; i < numCells; i++ {
  for j := 0; j < numCells; j++ {
    go func(i int){
      outputPosSlice := outputs.NewIndex(0)
      statesPosSlice := states.NewIndex(0)
      inputsPosSlice := inputs.NewIndex(0)

      outputPosSlice[sim.DIMO_CELL] = i
      statesPosSlice[sim.DIMS_CELL] = i
      inputsPosSlice[sim.DIMI_CELL] = i%numInputSequences

      bfac := m.bfac.Get1(i%m.bfac.Len1())
      coeff := m.coeff.Get1(i%m.coeff.Len1())
      dseep := m.dseep.Get1(i%m.dseep.Len1())
      fcfrac := m.fcFrac.Get1(i%m.fcFrac.Len1())
      fimp := m.fimp.Get1(i%m.fimp.Len1())
      rfac := m.rfac.Get1(i%m.rfac.Len1())
      smax := m.smax.Get1(i%m.smax.Len1())
      sq := m.sq.Get1(i%m.sq.Len1())
      thres := m.thres.Get1(i%m.thres.Len1())
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      
      initialStates := states.Slice(statesPosSlice,statesSizeSlice,nil).MustReshape([]int{numStates}).(data.ND1Float64)
      

      
      
      soilmoisturestore := initialStates.Get1(0)
      
      groundwater := initialStates.Get1(1)
      
      totalstore := initialStates.Get1(2)
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{rainfall mm}",tmpTS.Shape())
      rainfall := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{pet mm}",tmpTS.Shape())
      pet := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      runoff := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      quickflow := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 2
      baseflow := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 3
      store := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

      soilmoisturestore,groundwater,totalstore= surm(rainfall,pet,soilmoisturestore,groundwater,totalstore,bfac,coeff,dseep,fcfrac,fimp,rfac,smax,sq,thres,runoff,quickflow,baseflow,store)

      
      
      initialStates.Set1(0, soilmoisturestore)
      
      initialStates.Set1(1, groundwater)
      
      initialStates.Set1(2, totalstore)
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 4)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}

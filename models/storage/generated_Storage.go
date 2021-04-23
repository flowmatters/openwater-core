package storage

/* WARNING: GENERATED CODE
 *
 * This file is generated by ow-specgen using metadata from ./models/storage/storage.go
 * 
 * Don't edit this file. Edit ./models/storage/storage.go instead!
 */
import (
//  "fmt"
  "github.com/flowmatters/openwater-core/sim"
  "github.com/flowmatters/openwater-core/data"
)


type Storage struct {
  DeltaT data.ND1Float64
  nLVA data.ND1Float64
  levels data.ND2Float64
  volumes data.ND2Float64
  areas data.ND2Float64
  minRelease data.ND2Float64
  maxRelease data.ND2Float64
  

  maxnLVA int
  
}

func (m *Storage) ApplyParameters(parameters data.ND2Float64) {

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

  m.nLVA = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND1Float64)
  paramIdx += paramSize

  paramSize = 1* m.maxnLVA
  newShape = []int{m.maxnLVA, nSets}

  m.levels = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND2Float64)
  paramIdx += paramSize

  paramSize = 1* m.maxnLVA
  newShape = []int{m.maxnLVA, nSets}

  m.volumes = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND2Float64)
  paramIdx += paramSize

  paramSize = 1* m.maxnLVA
  newShape = []int{m.maxnLVA, nSets}

  m.areas = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND2Float64)
  paramIdx += paramSize

  paramSize = 1* m.maxnLVA
  newShape = []int{m.maxnLVA, nSets}

  m.minRelease = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND2Float64)
  paramIdx += paramSize

  paramSize = 1* m.maxnLVA
  newShape = []int{m.maxnLVA, nSets}

  m.maxRelease = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).MustReshape(newShape).(data.ND2Float64)
  paramIdx += paramSize

  
}


func buildStorage() sim.TimeSteppingModel {
	result := Storage{}
	return &result
}

func init() {
	sim.Catalog["Storage"] = buildStorage
}

func (m *Storage)  Description() sim.ModelDescription{
	var result sim.ModelDescription
  
  DeltaTDims := []string{
      }
  
  nLVADims := []string{
      }
  
  levelsDims := []string{
    "nLVA",  }
  
  volumesDims := []string{
    "nLVA",  }
  
  areasDims := []string{
    "nLVA",  }
  
  minReleaseDims := []string{
    "nLVA",  }
  
  maxReleaseDims := []string{
    "nLVA",  }
  
	result.Parameters = []sim.ParameterDescription{
  
  sim.DescribeParameter("DeltaT",86400,"Timestep",[]float64{ 1, 86400 }," ",DeltaTDims),
  sim.DescribeParameter("nLVA",0,"",[]float64{ 0, 0 },"",nLVADims),
  sim.DescribeParameter("levels",0,"",[]float64{ 0, 0 },"",levelsDims),
  sim.DescribeParameter("volumes",0,"",[]float64{ 0, 0 },"",volumesDims),
  sim.DescribeParameter("areas",0,"",[]float64{ 0, 0 },"",areasDims),
  sim.DescribeParameter("minRelease",0,"",[]float64{ 0, 0 },"",minReleaseDims),
  sim.DescribeParameter("maxRelease",0,"",[]float64{ 0, 0 },"",maxReleaseDims),}

  result.Inputs = []string{
  "rainfall","pet","inflow","demand",}
  result.Outputs = []string{
  "volume","outflow",}

  result.States = []string{
  "currentVolume","level","area",}

  result.Dimensions = []string{
    "nLVA",  }
	return result
}

func (m *Storage) InitialiseDimensions(dims []int) {
  m.maxnLVA = dims[0]
  
}

func (m *Storage) FindDimensions(parameters data.ND2Float64) []int {
  
  nSets := parameters.Len(sim.DIMP_CELL)
  paramIdx := 0
  paramSize := 1
  maxValues := make(map[string]float64)

  paramSize = 1

  maxValues["DeltaT"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  paramSize = 1

  maxValues["nLVA"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  paramSize = 1* int(maxValues["nLVA"])

  maxValues["levels"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  paramSize = 1* int(maxValues["nLVA"])

  maxValues["volumes"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  paramSize = 1* int(maxValues["nLVA"])

  maxValues["areas"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  paramSize = 1* int(maxValues["nLVA"])

  maxValues["minRelease"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  paramSize = 1* int(maxValues["nLVA"])

  maxValues["maxRelease"] = parameters.Slice([]int{ paramIdx, 0}, []int{ paramSize, nSets}, nil).Maximum()
  paramIdx += paramSize

  
  dims := []int{
    int(maxValues["nLVA"]),
    
  }
  
  return dims
  
}




func (m *Storage) InitialiseStates(n int) data.ND2Float64 {
  // Zero states
	var result = data.NewArray2DFloat64(n,3)

	// for i := 0; i < n; i++ {
  //   stateSet := make(sim.StateSet,3)
  //   
	// 	stateSet[0] = 0 // currentVolume
  //   
	// 	stateSet[1] = 0 // level
  //   
	// 	stateSet[2] = 0 // area
  //   

  //   if result==nil {
  //     result = data.NewArray2DFloat64(stateSet.Len(0),n)
  //   }
  //   result.Apply([]int{0,i},[]int{1,1},stateSet)
	// }
 
	return result
}



func (m *Storage) Run(inputs data.ND3Float64, states data.ND2Float64, outputs data.ND3Float64) {

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
  // fmt.Println("Running Storage for ",numCells,"cells")
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
      // Dimension parameter
      nlva := int(m.nLVA.Get1(i%m.nLVA.Len1()))
      levelsShape := m.levels.Shape()
      levelsNSets := levelsShape[len(levelsShape)-1]
      levelsFrom := []int{  0,  i%levelsNSets }
      levelsSliceShape := []int{  m.maxnLVA,  }
      levels := m.levels.Slice(levelsFrom, levelsSliceShape, nil).(data.ND1Float64)
      volumesShape := m.volumes.Shape()
      volumesNSets := volumesShape[len(volumesShape)-1]
      volumesFrom := []int{  0,  i%volumesNSets }
      volumesSliceShape := []int{  m.maxnLVA,  }
      volumes := m.volumes.Slice(volumesFrom, volumesSliceShape, nil).(data.ND1Float64)
      areasShape := m.areas.Shape()
      areasNSets := areasShape[len(areasShape)-1]
      areasFrom := []int{  0,  i%areasNSets }
      areasSliceShape := []int{  m.maxnLVA,  }
      areas := m.areas.Slice(areasFrom, areasSliceShape, nil).(data.ND1Float64)
      minreleaseShape := m.minRelease.Shape()
      minreleaseNSets := minreleaseShape[len(minreleaseShape)-1]
      minreleaseFrom := []int{  0,  i%minreleaseNSets }
      minreleaseSliceShape := []int{  m.maxnLVA,  }
      minrelease := m.minRelease.Slice(minreleaseFrom, minreleaseSliceShape, nil).(data.ND1Float64)
      maxreleaseShape := m.maxRelease.Shape()
      maxreleaseNSets := maxreleaseShape[len(maxreleaseShape)-1]
      maxreleaseFrom := []int{  0,  i%maxreleaseNSets }
      maxreleaseSliceShape := []int{  m.maxnLVA,  }
      maxrelease := m.maxRelease.Slice(maxreleaseFrom, maxreleaseSliceShape, nil).(data.ND1Float64)
      

      // fmt.Println("i",i)
      // fmt.Println("States",states.Shape())
      // fmt.Println("Tmp2",tmp2.Shape())
      
      initialStates := states.Slice(statesPosSlice,statesSizeSlice,nil).MustReshape([]int{numStates}).(data.ND1Float64)
      

      
      
      currentvolume := initialStates.Get1(0)
      
      level := initialStates.Get1(1)
      
      area := initialStates.Get1(2)
      
      

  //    fmt.Println("is",inputDims,"tmpShape",tmpCI.Shape(),"cis",cellInputsShape)

      cellInputs := inputs.Slice(inputsPosSlice,inputsSizeSlice,nil).MustReshape(cellInputsShape)
  //    fmt.Println("cellInputs Shape",cellInputs.Shape())
      
  //    fmt.Println("{rainfall mm}",tmpTS.Shape())
      rainfall := cellInputs.Slice([]int{ 0,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{pet mm}",tmpTS.Shape())
      pet := cellInputs.Slice([]int{ 1,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{inflow m^3.s^-1}",tmpTS.Shape())
      inflow := cellInputs.Slice([]int{ 2,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      
  //    fmt.Println("{demand m^3.s^-1}",tmpTS.Shape())
      demand := cellInputs.Slice([]int{ 3,0}, []int{ 1,inputLen}, nil).MustReshape(inputNewShape).(data.ND1Float64)
      

      

      
      
      outputPosSlice[sim.DIMO_OUTPUT] = 0
      volume := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      outputPosSlice[sim.DIMO_OUTPUT] = 1
      outflow := outputs.Slice(outputPosSlice,outputSizeSlice,outputStepSlice).MustReshape([]int{inputLen}).(data.ND1Float64)
      
      

      currentvolume,level,area= storageWaterBalance(rainfall,pet,inflow,demand,currentvolume,level,area,deltat,nlva,levels,volumes,areas,minrelease,maxrelease,volume,outflow)

      
      
      initialStates.Set1(0, currentvolume)
      
      initialStates.Set1(1, level)
      
      initialStates.Set1(2, area)
      
      

  //		result.Outputs.ApplySpice([]int{i,0,0},[]int = make([]sim.Series, 2)
      

      doneChan <- i
    }(j)
	}

  for j := 0; j < numCells; j++ {
    <- doneChan
  }
//	return result
}
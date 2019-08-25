package main

import (
	"C"
	"unsafe"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/data/cdata"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
)

//export RunSingleModel
func RunSingleModel(
	// nInputSets, nInputs, nTimesteps, nParameters, nParameterSets, nCells, nStates, nOutputCells, nOutputs, nOutputTimesteps C.int,
	// inputs, params, states, outputs *C.double,
	modelName *C.char,
	inputs *C.double, nInputSets, nInputs, nTimesteps C.int,
	params *C.double, nParameters, nParameterSets C.int,
	states *C.double, nCells, nStates C.int,
	outputs *C.double, nOutputCells, nOutputs, nOutputTimesteps C.int,
	initStates bool) {
	gName := C.GoString(modelName)
	model := sim.Catalog[gName]()

	// fmt.Println(nInputSets, nInputs, nTimesteps)
	// fmt.Println(nParameters, nParameterSets)
	// fmt.Println(nOutputCells, nOutputs, nOutputTimesteps)

	iArray := cdata.NewFloat64CArray(unsafe.Pointer(inputs), []int{int(nInputSets), int(nInputs), int(nTimesteps)}).(data.ND3Float64)
	pArray := cdata.NewFloat64CArray(unsafe.Pointer(params), []int{int(nParameters), int(nParameterSets)}).(data.ND2Float64)
	oArray := cdata.NewFloat64CArray(unsafe.Pointer(outputs), []int{int(nOutputCells), int(nOutputs), int(nOutputTimesteps)}).(data.ND3Float64)

	model.ApplyParameters(pArray)
	// coeffSlice := pArray.Slice([]int{0, 0}, []int{1, int(nParameterSets)}, nil).(data.ND1Float64)
	// i := 0
	// coeff := coeffSlice.Get1(i % coeffSlice.Len1())
	// fmt.Println("Runoff coefficient is", coeff)
	var sArray data.ND2Float64
	// fmt.Println("initStates", initStates)
	if initStates {
		sArray = model.InitialiseStates(int(nCells))
	} else {
		sArray = cdata.NewFloat64CArray(unsafe.Pointer(states), []int{int(nCells), int(nStates)}).(data.ND2Float64)
	}
	// fmt.Println("i", iArray.Shape())
	// fmt.Println("p", pArray.Shape())
	// fmt.Println("s", sArray.Shape())
	// fmt.Println("o", oArray.Shape())

	// fmt.Printf("Running model: %s!\n", gName)
	model.Run(iArray, sArray, oArray)
	// for i := 0; i < iArray.Len3(); i += 10 {
	// 	fmt.Println(i, iArray.Get3(0, 0, i))
	// }

	// fmt.Println("Params")
	// fmt.Println(pArray.Get2(0, 0))

	// if initStates Copy data back into provided states array...
	if initStates && (states != nil) {
		sOrig := cdata.NewFloat64CArray(unsafe.Pointer(states), []int{int(nCells), int(nStates)}).(data.ND2Float64)
		sOrig.CopyFrom(sArray)
	}
}

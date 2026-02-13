package main

import (
	"C"
	"unsafe"

	"log"
	"os"
	"runtime/pprof"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/data/cdata"
	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
	"github.com/flowmatters/openwater-core/util"
)

//export ow_version
func ow_version() *C.char {
	return C.CString(util.FullVersion())
}

//export ow_short_version
func ow_short_version() *C.char {
	return C.CString(util.ShortVersion())
}

//export ow_signature_hash
func ow_signature_hash() *C.char {
	return C.CString(util.GetSignatureHash())
}

//export RunSingleModel
func RunSingleModel(
	// nInputSets, nInputs, nTimesteps, nParameters, nParameterSets, nCells, nStates, nOutputCells, nOutputs, nOutputTimesteps C.int,
	// inputs, params, states, outputs *C.double,
	modelName *C.char,
	inputs *C.double, nInputSets, nInputs, nTimesteps C.int,
	params *C.double, nParameters, nParameterSets C.int,
	states *C.double, nCells, nStates C.int,
	outputs *C.double, nOutputCells, nOutputs, nOutputTimesteps C.int,
	initStates bool, cpuprofile *C.char) {

	if cpuprofile != nil && C.GoString(cpuprofile) != "" {
		f, err := os.Create(C.GoString(cpuprofile))
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	gName := C.GoString(modelName)
	model := sim.Catalog[gName]()

	iArray := cdata.NewCArray[float64](unsafe.Pointer(inputs), []int{int(nInputSets), int(nInputs), int(nTimesteps)}).(data.ND3[float64])
	pArray := cdata.NewCArray[float64](unsafe.Pointer(params), []int{int(nParameters), int(nParameterSets)}).(data.ND2[float64])
	oArray := cdata.NewCArray[float64](unsafe.Pointer(outputs), []int{int(nOutputCells), int(nOutputs), int(nOutputTimesteps)}).(data.ND3[float64])

	dimSizes := model.FindDimensions(pArray.(data.ND2[float64]))
	if len(dimSizes) > 0 {
		model.InitialiseDimensions(dimSizes)
	}

	model.ApplyParameters(pArray)
	// coeffSlice := pArray.Slice([]int{0, 0}, []int{1, int(nParameterSets)}, nil).(data.ND1[float64])
	// i := 0
	// coeff := coeffSlice.Get1(i % coeffSlice.Len1())
	var sArray data.ND2[float64]
	if initStates {
		sArray = model.InitialiseStates(int(nCells))
	} else {
		sArray = cdata.NewCArray[float64](unsafe.Pointer(states), []int{int(nCells), int(nStates)}).(data.ND2[float64])
	}

	// fmt.Errorf("Running model: %s!", gName)
	model.Run(iArray, sArray, oArray)
	// for i := 0; i < iArray.Len3(); i += 10 {
	// }

	// if initStates Copy data back into provided states array...
	if initStates && (states != nil) {
		sOrig := cdata.NewCArray[float64](unsafe.Pointer(states), []int{int(nCells), int(nStates)}).(data.ND2[float64])
		sOrig.CopyFrom(sArray)
	}
}

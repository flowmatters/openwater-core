package arraybench

// import (
// 	"testing"

// 	"github.com/flowmatters/openwater-core/data"
// )

// const SUB_ARRAY_SIZE_F64 = 1000
// const SUB_ARRAY_COUNT_F64 = 10
// const ARRAY_SIZE_F64 = SUB_ARRAY_COUNT_F64 * SUB_ARRAY_SIZE_F64

// func BenchmarkND3GetSet3_run_f64(b *testing.B) {
// 	as := data.NewArray3D[float64](300, 400, ARRAY_SIZE_F64)
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE_F64; j++ {
// 			v := as.Get3(0, 0, j)
// 			as.Set3(0, 0, j, v+1)
// 		}
// 	}
// }

// func BenchmarkND3GetSetReuseIndex_run_f64(b *testing.B) {
// 	as := data.NewArray3D[float64](300, 400, ARRAY_SIZE_F64)
// 	b.ResetTimer()
// 	idx := []int{0, 0, 0}
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE_F64; j++ {
// 			idx[2] = j
// 			v := as.Get(idx)
// 			as.Set(idx, v+1)
// 		}
// 	}
// }

// func BenchmarkND3GetSet31D_run_f64(b *testing.B) {
// 	as := data.NewArray1D[float64](ARRAY_SIZE_F64)
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE_F64; j++ {
// 			v := as.Get1(j)
// 			as.Set1(j, v+1)
// 		}
// 	}
// }

// func BenchmarkND3GetSetReuseIndex1D_run_f64(b *testing.B) {
// 	as := data.NewArray1D[float64](ARRAY_SIZE_F64)
// 	b.ResetTimer()
// 	idx := []int{0}
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE_F64; j++ {
// 			idx[0] = j
// 			v := as.Get(idx)
// 			as.Set(idx, v+1)
// 		}
// 	}
// }

// func BenchmarkND3GetSet1_run_f64(b *testing.B) {
// 	as := data.NewArray3D[float64](ARRAY_SIZE_F64, 300, 400)
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE_F64; j++ {
// 			v := as.Get3(j, 0, 0)
// 			as.Set3(j, 0, 0, v+1)
// 		}
// 	}
// }

// func BenchmarkND3AsSlice_run_f64(b *testing.B) {
// 	as := data.NewArray3D[float64](300, 400, ARRAY_SIZE_F64).Slice([]int{0, 0, 0}, []int{1, 1, ARRAY_SIZE_F64}, nil).MustReshape([]int{ARRAY_SIZE_F64})
// 	if !as.Contiguous() {
// 		b.Fatalf("Array is not contiguous")
// 	}
// 	b.ResetTimer()
// 	idx := []int{0}
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE_F64; j++ {
// 			idx[0] = j
// 			v := as.Get(idx)
// 			as.Set(idx, v+1)
// 		}
// 	}
// }

// func makeArraySliceF64() []([SUB_ARRAY_SIZE_F64]float64) {
// 	return make([][SUB_ARRAY_SIZE_F64]float64, SUB_ARRAY_COUNT_F64)
// }

// func BenchmarkArraySliceIterator_run_f64(b *testing.B) {
// 	as := makeArraySliceF64()
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for k := range as {
// 			for j := range as[k] {
// 				as[k][j] += 1
// 			}
// 		}
// 	}
// }

// func BenchmarkArray_run_f64(b *testing.B) {
// 	var arr [ARRAY_SIZE_F64]float64
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for j := range arr {
// 			arr[j] += 1
// 		}
// 	}
// }

// func BenchmarkPreallocatedSlice_run_f64(b *testing.B) {
// 	slice := make([]float64, ARRAY_SIZE_F64)
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for j := range slice {
// 			slice[j] += 1
// 		}
// 	}
// }

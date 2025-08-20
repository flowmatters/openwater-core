package arraybench

import (
	// "runtime/pprof",

	"testing"

	"github.com/flowmatters/openwater-core/data"
)

const SUB_ARRAY_SIZE = 1000
const SUB_ARRAY_COUNT = 10
const ARRAY_SIZE = SUB_ARRAY_COUNT * SUB_ARRAY_SIZE

// func BenchmarkND3GetSet3_run(b *testing.B) {
// 	as := data.NewArray3D[int](300, 400, ARRAY_SIZE)
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE; j++ {
// 			v := as.Get3(0, 0, j)
// 			as.Set3(0, 0, j, v+1)
// 		}
// 	}
// }

// func BenchmarkND3GetSetReuseIndex_run(b *testing.B) {
// 	as := data.NewArray3D[int](300, 400, ARRAY_SIZE)
// 	b.ResetTimer()
// 	idx := []int{0, 0, 0}
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE; j++ {
// 			idx[2] = j
// 			v := as.Get(idx)
// 			as.Set(idx, v+1)
// 		}
// 	}
// }

// func BenchmarkND3GetSet31D_run(b *testing.B) {
// 	as := data.NewArray1D[int](ARRAY_SIZE)
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE; j++ {
// 			v := as.Get1(j)
// 			as.Set1(j, v+1)
// 		}
// 	}
// }

// func BenchmarkND3GetSetReuseIndex1D_run(b *testing.B) {
// 	as := data.NewArray1D[int](ARRAY_SIZE)
// 	b.ResetTimer()
// 	idx := []int{0}
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE; j++ {
// 			idx[0] = j
// 			v := as.Get(idx)
// 			as.Set(idx, v+1)
// 		}
// 	}
// }

// func BenchmarkND3GetSet1_run(b *testing.B) {
// 	as := data.NewArray3D[int](ARRAY_SIZE, 300, 400)
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for j := 0; j < ARRAY_SIZE; j++ {
// 			v := as.Get3(j, 0, 0)
// 			as.Set3(j, 0, 0, v+1)
// 		}
// 	}
// }

// func BenchmarkND3AsSliceIterGetPull_run(b *testing.B) {
// 	as := data.NewArray3D[int](300, 400, ARRAY_SIZE).Slice([]int{0, 0, 0}, []int{1, 1, ARRAY_SIZE}, nil).MustReshape([]int{ARRAY_SIZE})
// 	if !as.Contiguous() {
// 		b.Fatalf("Array is not contiguous")
// 	}
// 	b.ResetTimer()
// 	idx := []int{0}
// 	for b.Loop() {
// 		values := as.Values(idx, 0, ARRAY_SIZE)
// 		next, _ := iter.Pull(values)
// 		// for v := range values {
// 		for j := 0; j < ARRAY_SIZE; j++ {
// 			v, _ := next()
// 			as.Set(idx, v+1)
// 		}
// 		// }
// 	}
// }

// func BenchmarkND3AsSliceIterGetRange_run(b *testing.B) {
// 	as := data.NewArray3D[int](300, 400, ARRAY_SIZE).Slice([]int{0, 0, 0}, []int{1, 1, ARRAY_SIZE}, nil).MustReshape([]int{ARRAY_SIZE})
// 	if !as.Contiguous() {
// 		b.Fatalf("Array is not contiguous")
// 	}
// 	b.ResetTimer()
// 	idx := []int{0}
// 	for b.Loop() {
// 		values := as.Values(idx, 0, ARRAY_SIZE)
// 		// for v := range values {
// 		for v := range values {
// 			as.Set(idx, v+1)
// 		}
// 		// }
// 	}
// }

func BenchmarkND3AsSliceRawIterGet_run(b *testing.B) {
	as := data.NewArray3D[int](300, 400, ARRAY_SIZE).Slice([]int{0, 0, 0}, []int{1, 1, ARRAY_SIZE}, nil).MustReshape([]int{ARRAY_SIZE})
	if !as.Contiguous() {
		b.Fatalf("Array is not contiguous")
	}
	b.ResetTimer()
	idx := []int{0}
	for b.Loop() {
		for v := range as.Unroll() {
			as.Set(idx, v+1)
		}
	}
}

// func BenchmarkFullRawIO_run(b *testing.B) {
// 	as := data.NewArray3D[int](300, 400, ARRAY_SIZE).Slice([]int{0, 0, 0}, []int{1, 1, ARRAY_SIZE}, nil).MustReshape([]int{ARRAY_SIZE})
// 	if !as.Contiguous() {
// 		b.Fatalf("Array is not contiguous")
// 	}
// 	b.ResetTimer()
// 	idx := []int{0}
// 	for b.Loop() {
// 		index := as.RawIndices(idx, 0, ARRAY_SIZE)
// 		for i := range index {
// 			as.SetRaw(i, as.GetRaw(i)+1)
// 		}
// 	}
// }

// func BenchmarkFullRawIONoIter_run(b *testing.B) {
// 	as := data.NewArray3D[int](300, 400, ARRAY_SIZE).Slice([]int{0, 0, 0}, []int{1, 1, ARRAY_SIZE}, nil).MustReshape([]int{ARRAY_SIZE})
// 	if !as.Contiguous() {
// 		b.Fatalf("Array is not contiguous")
// 	}
// 	b.ResetTimer()
// 	for b.Loop() {
// 		offset := as.OffsetStep[0]
// 		for i := 0; i < ARRAY_SIZE; i += offset {
// 			as.SetRaw(i, as.GetRaw(i)+1)
// 		}
// 	}
// }

func BenchmarkND3AsSlice_run(b *testing.B) {
	as := data.NewArray3D[int](300, 400, ARRAY_SIZE).Slice([]int{0, 0, 0}, []int{1, 1, ARRAY_SIZE}, nil).MustReshape([]int{ARRAY_SIZE})
	if !as.Contiguous() {
		b.Fatalf("Array is not contiguous")
	}
	b.ResetTimer()
	idx := []int{0}
	for b.Loop() {
		for j := 0; j < ARRAY_SIZE; j++ {
			idx[0] = j
			v := as.Get(idx)
			as.Set(idx, v+1)
		}
	}
}

// func makeArraySlice[T data.Number]() []([SUB_ARRAY_SIZE]T) {
// 	return make([][SUB_ARRAY_SIZE]T, SUB_ARRAY_COUNT)
// }

// func BenchmarkArraySliceIterator_run(b *testing.B) {
// 	as := makeArraySlice[int]()
// 	b.ResetTimer()
// 	for b.Loop() {
// 		for k := range as {
// 			for j := range as[k] {
// 				as[k][j] += 1
// 			}
// 		}
// 		// as = makeArraySlice[int]()
// 	}
// }

// From https://go-benchmarks.com/array-vs-slice
// BenchmarkArray_run seems ridiculously fast?
// func BenchmarkArray_run(b *testing.B) {
// 	var arr [ARRAY_SIZE]int // Define an array of fixed size

// 	b.ResetTimer()

// 	for b.Loop() {
// 		for j := range arr {
// 			arr[j] += 1
// 		}
// 		// arr = [ARRAY_SIZE]int{} // Reset the array
// 	}
// }

// func BenchmarkPreallocatedSlice_run(b *testing.B) {
// 	slice := make([]int, ARRAY_SIZE) // Define a slice with the same size as the array

// 	b.ResetTimer()

// 	for b.Loop() {
// 		for j := range slice {
// 			slice[j] += 1
// 		}
// 		// slice = make([]int, ARRAY_SIZE) // Reset the slice
// 	}
// }

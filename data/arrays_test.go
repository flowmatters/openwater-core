package data

import (
	"testing"

	"github.com/flowmatters/openwater-core/util/slice"
	"github.com/stretchr/testify/assert"
)

//import "fmt"
// import "fmt"

func TestOffset(t *testing.T) {
	lenI := 3
	lenJ := 2
	lenK := 4
	arr := newArray[float64]([]int{lenI, lenJ, lenK})
	expOffset := []int{8, 4, 1}
	if !slice.Equal(expOffset, arr.Offset) {
		t.Errorf("Incorrect Offset. Expected %v, got %v", expOffset, arr.Offset)
	}
}

func testData3D() ND3[float64] {
	lenI := 3
	lenJ := 2
	lenK := 4
	arr := NewArray3D[float64](lenI, lenJ, lenK)

	a := 0
	for i := 0; i < lenI; i++ {
		for j := 0; j < lenJ; j++ {
			for k := 0; k < lenK; k++ {
				arr.Set3(i, j, k, float64(a))
				a++
			}
		}
	}

	return arr
}

func testData2D() ND2[float64] {
	arr := NewArray2D[float64](2, 4)
	arr.Set2(0, 0, 0)
	arr.Set2(0, 1, 1)
	arr.Set2(0, 2, 35)
	arr.Set2(0, 3, 3)

	arr.Set2(1, 0, 5)
	arr.Set2(1, 1, 7)
	arr.Set2(1, 2, 75)
	arr.Set2(1, 3, 13)
	return arr
}

func TestNewAndAccess(t *testing.T) {
	arr := testData3D()
	testGet3(t, arr, 0, 0, 1, 1.0)
	testGet3(t, arr, 0, 1, 0, 4.0)
	testGet3(t, arr, 1, 0, 0, 8.0)
	testGet3(t, arr, 1, 1, 0, 12.0)
	testGet3(t, arr, 2, 1, 0, 20.0)
}

func TestSliceAndAccess(t *testing.T) {
	arr := testData3D()
	arrSlice := arr.Slice([]int{1, 1, 1}, []int{2, 1, 2}, []int{1, 1, 1}).(ND3[float64])
	//arrNative := arrSlice.(*nd[float64])

	expShape := []int{2, 1, 2}
	if !slice.Equal(expShape, arrSlice.Shape()) {
		t.Errorf("Slice shape should be %v. Got %v", expShape, arrSlice.Shape())
	}

	testGet3(t, arrSlice, 0, 0, 0, 13.0)
	testGet3(t, arrSlice, 1, 0, 0, 21.0)
	testGet3(t, arrSlice, 1, 0, 1, 22.0)
}

func TestContiguous(t *testing.T) {
	arr := testData3D()
	contigSlice1 := arr.Slice([]int{1, 0, 0}, []int{1, 1, 3}, []int{1, 1, 1})
	contigSlice2 := arr.Slice([]int{2, 0, 0}, []int{1, 2, 4}, []int{1, 1, 1})
	disContigSlice1 := arr.Slice([]int{1, 0, 0}, []int{1, 1, 2}, []int{1, 1, 2})
	disContigSlice2 := arr.Slice([]int{2, 0, 0}, []int{1, 2, 3}, []int{1, 1, 1})

	testContig(t, contigSlice1, true)
	testContig(t, contigSlice2, true)
	testContig(t, disContigSlice1, false)
	testContig(t, disContigSlice2, false)
}

func TestContiguousBig(t *testing.T) {
	arr := NewArray3D[float64](20, 30, 10)
	contig1 := arr.Slice([]int{5, 0, 0}, []int{3, 30, 10}, []int{1, 1, 1})
	disContig1 := arr.Slice([]int{5, 0, 0}, []int{3, 30, 10}, []int{1, 1, 2})
	disContig2 := arr.Slice([]int{5, 0, 0}, []int{3, 30, 9}, []int{1, 1, 1})

	testContig(t, contig1, true)
	testContig(t, disContig1, false)
	testContig(t, disContig2, false)
}

func TestUnroll(t *testing.T) {
	arr := testData3D()

	arrSlice1 := arr.Slice([]int{1, 0, 0}, []int{1, 1, 4}, []int{1, 1, 1}).Unroll()
	testSlice(t, arrSlice1, []float64{8.0, 9.0, 10.0, 11.0})

	arrSlice2 := arr.Slice([]int{1, 0, 0}, []int{1, 2, 3}, []int{1, 1, 1}).Unroll()
	testSlice(t, arrSlice2, []float64{8.0, 9.0, 10.0, 12.0, 13.0, 14.0})

	arrSlice3 := arr.Slice([]int{0, 1, 0}, []int{3, 1, 3}, []int{1, 1, 1}).Unroll()
	testSlice(t, arrSlice3, []float64{4.0, 5.0, 6.0, 12.0, 13.0, 14.0, 20.0, 21.0, 22.0})

}

func testGet3(t *testing.T, arr ND3[float64], loc1 int, loc2 int, loc3 int, exp float64) {
	res := arr.Get3(loc1, loc2, loc3)
	if res != exp {
		t.Errorf("arr[%d,%d,%d] expected %f, got %f", loc1, loc2, loc3, exp, res)
	}
}

func testContig(t *testing.T, arr ND[float64], expected bool) {
	res := arr.Contiguous()
	if res != expected {
		t.Errorf("Expected slice (%v).Contiguous()==%t but was %t", arr, expected, res)
	}
}

func testSlice(t *testing.T, fSlice []float64, expected []float64) {
	if len(fSlice) != len(expected) {
		t.Errorf("Length mismatch (exp %v (%d), got %v (%d)).", expected, len(expected), fSlice, len(fSlice))
		return
	}

	for i := range expected {
		if expected[i] != fSlice[i] {
			t.Errorf("Mismatch at %d. Expected %f, got %f", i, expected[i], fSlice[i])
		}
	}
}

func TestReshape(t *testing.T) {
	arr := testData2D()
	sliced, err := arr.Slice([]int{0, 0}, []int{2, 1}, nil).Reshape([]int{2})

	if assert.Nil(t, err) {
		arr1D := sliced.(ND1[float64])
		assert.Equal(t, 1, arr1D.NDims())
		assert.Equal(t, 0.0, arr1D.Get1(0))
		assert.Equal(t, 5.0, arr1D.Get1(1))
	}
}

func TestReshapeFast(t *testing.T) {
	arr := testData2D()
	sliced, err := arr.Slice([]int{0, 0}, []int{1, 4}, nil).ReshapeFast([]int{4})

	if assert.Nil(t, err) {
		arr1D := sliced.(ND1[float64])
		assert.Equal(t, 1, arr1D.NDims())
		assert.Equal(t, 0.0, arr1D.Get1(0))
		assert.Equal(t, 1.0, arr1D.Get1(1))
		assert.Equal(t, 35.0, arr1D.Get1(2))
		assert.Equal(t, 3.0, arr1D.Get1(3))
	}
}

func TestTreatAs1D(t *testing.T) {
	arr := testData2D()
	arr1D := arr.Slice([]int{0, 0}, []int{2, 1}, nil).(ND1[float64])

	//	assert.Equal(t,1,arr1D.NDims())
	assert.Equal(t, 0.0, arr1D.Get1(0))
	assert.Equal(t, 5.0, arr1D.Get1(1))

	alt1D := arr.Slice([]int{0, 0}, []int{1, 2}, nil).(ND1[float64])
	assert.Equal(t, 0.0, alt1D.Get([]int{0, 0}))
	assert.Equal(t, 1.0, alt1D.Get([]int{0, 1}))

	assert.Equal(t, 0.0, alt1D.Get1(0))
	assert.Equal(t, 1.0, alt1D.Get1(1))
}

func TestApplySlice(t *testing.T) {
	assert := assert.New(t)

	arr := testData2D()
	subst1D := NewArray1D[float64](2)
	subst1D.Set1(0, 21.0)
	subst1D.Set1(1, 22.0)
	subst2D, e := subst1D.Reshape([]int{1, 2})

	assert.Nil(e)
	arr.ApplySlice([]int{1, 1}, nil, subst2D)

	assert.Equal(21.0, arr.Get2(1, 1))
	assert.Equal(22.0, arr.Get2(1, 2))

	subst2D, e = subst1D.Reshape([]int{2, 1})
	assert.Nil(e)
	arr.ApplySlice([]int{0, 0}, nil, subst2D)

	assert.Equal(21.0, arr.Get2(0, 0))
	assert.Equal(22.0, arr.Get2(1, 0))

}

func TestApply(t *testing.T) {
	assert := assert.New(t)

	arr := testData2D()

	slice2 := []float64{99.0, 101.0}
	arr.Apply([]int{0, 1}, 0, 1, slice2)
	assert.Equal(99.0, arr.Get2(0, 1))
	assert.Equal(101.0, arr.Get2(1, 1))

	slice3 := []float64{423.0, 404.0, 500.0}
	arr.Apply([]int{1, 1}, 1, 1, slice3)
	assert.Equal(423.0, arr.Get2(1, 1))
	assert.Equal(404.0, arr.Get2(1, 2))
	assert.Equal(500.0, arr.Get2(1, 3))
}

func TestARange(t *testing.T) {
	assert := assert.New(t)

	arr := ARange[float64](12.0).MustReshape([]int{3, 4}).(ND2[float64])

	expShape := []int{3, 4}
	assert.True(slice.Equal(expShape, arr.Shape()), "Slice shape should be %v. Got %v", expShape, arr.Shape())
	assert.Equal(0.0, arr.Get2(0, 0))
	assert.Equal(3.0, arr.Get2(0, 3))
	assert.Equal(4.0, arr.Get2(1, 0))
	assert.Equal(11.0, arr.Get2(2, 3))

}

func testCopyFrom(t *testing.T, from, to ND2[float64]) {
	assert := assert.New(t)
	rows := from.Len(0)
	cols := from.Len(1)

	i := 0

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			to.Set2(r, c, 0.0)
			from.Set2(r, c, float64(i))
			i++
		}
	}

	to.CopyFrom(from)
	i = 0

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			assert.Equalf(float64(i), to.Get2(r, c), "Error at [%d,%d] in [%d,%d] array", r, c, rows, cols)
			i++
		}
	}

}

func TestCopyNativeToNative(t *testing.T) {
	rows := 9
	cols := 2

	dest := NewArray2D[float64](rows, cols)
	src := NewArray2D[float64](rows, cols)

	testCopyFrom(t, src, dest)
}

func TestMaximumArray(t *testing.T) {
	assert := assert.New(t)

	arr := NewArray1D[float64](4)
	arr.Set1(0, 3.0)
	arr.Set1(1, 7.0)
	arr.Set1(2, 1.0)
	arr.Set1(3, 5.0)
	assert.Equal(7.0, arr.Maximum())

	arr2D := testData2D()
	assert.Equal(75.0, arr2D.Maximum())
}

func TestMinimumArray(t *testing.T) {
	assert := assert.New(t)

	arr := NewArray1D[float64](4)
	arr.Set1(0, 3.0)
	arr.Set1(1, -2.0)
	arr.Set1(2, 1.0)
	arr.Set1(3, 5.0)
	assert.Equal(-2.0, arr.Minimum())

	arr2D := testData2D()
	assert.Equal(0.0, arr2D.Minimum())
}

func TestArrayFromSlice(t *testing.T) {
	assert := assert.New(t)

	data := []float64{1, 2, 3, 4, 5, 6}
	arr := ArrayFromSlice(data, []int{2, 3}).(ND2[float64])

	assert.Equal(1.0, arr.Get2(0, 0))
	assert.Equal(3.0, arr.Get2(0, 2))
	assert.Equal(4.0, arr.Get2(1, 0))
	assert.Equal(6.0, arr.Get2(1, 2))

	// Verify it shares the backing slice
	data[0] = 99.0
	assert.Equal(99.0, arr.Get2(0, 0))
}

func TestSliceWithStepValues(t *testing.T) {
	assert := assert.New(t)

	arr := ARange[float64](12).MustReshape([]int{3, 4})

	// Every other column from row 1: values 4,6
	sliced := arr.Slice([]int{1, 0}, []int{1, 2}, []int{1, 2})
	assert.Equal(4.0, sliced.Get([]int{0, 0}))
	assert.Equal(6.0, sliced.Get([]int{0, 1}))

	// Every other row, all columns: rows 0 and 2
	sliced2 := arr.Slice([]int{0, 0}, []int{2, 4}, []int{2, 1})
	assert.Equal(0.0, sliced2.Get([]int{0, 0}))
	assert.Equal(3.0, sliced2.Get([]int{0, 3}))
	assert.Equal(8.0, sliced2.Get([]int{1, 0}))
	assert.Equal(11.0, sliced2.Get([]int{1, 3}))
}

func TestApplyWithStep(t *testing.T) {
	assert := assert.New(t)

	arr := NewArray1D[float64](6)
	for i := 0; i < 6; i++ {
		arr.Set1(i, 0.0)
	}

	// Apply every other element
	arr.Apply([]int{0}, 0, 2, []float64{10.0, 20.0, 30.0})
	assert.Equal(10.0, arr.Get1(0))
	assert.Equal(0.0, arr.Get1(1))
	assert.Equal(20.0, arr.Get1(2))
	assert.Equal(0.0, arr.Get1(3))
	assert.Equal(30.0, arr.Get1(4))
	assert.Equal(0.0, arr.Get1(5))
}

func TestReshapeSizeMismatch(t *testing.T) {
	arr := NewArray2D[float64](2, 3)
	_, err := arr.Reshape([]int{4, 4})
	assert.NotNil(t, err)
}

func TestReshapeFastNonContiguous(t *testing.T) {
	arr := ARange[float64](12).MustReshape([]int{3, 4})
	// Non-contiguous slice
	sliced := arr.Slice([]int{0, 0}, []int{3, 2}, []int{1, 1})
	_, err := sliced.ReshapeFast([]int{6})
	assert.NotNil(t, err)
}

func TestCopyFromNonContiguous(t *testing.T) {
	assert := assert.New(t)

	src := ARange[float64](12).MustReshape([]int{3, 4})
	// Non-contiguous: every other column
	srcSlice := src.Slice([]int{0, 0}, []int{3, 2}, []int{1, 2})

	dst := NewArray2D[float64](3, 2)
	dst.CopyFrom(srcSlice)

	// src column 0: 0, 4, 8; src column 2: 2, 6, 10
	assert.Equal(0.0, dst.Get2(0, 0))
	assert.Equal(2.0, dst.Get2(0, 1))
	assert.Equal(4.0, dst.Get2(1, 0))
	assert.Equal(6.0, dst.Get2(1, 1))
	assert.Equal(8.0, dst.Get2(2, 0))
	assert.Equal(10.0, dst.Get2(2, 1))
}

func TestApplyOnNonContiguousSlice(t *testing.T) {
	assert := assert.New(t)

	arr := NewArray2D[float64](3, 4)
	// Non-contiguous slice: every other column
	sliced := arr.Slice([]int{0, 0}, []int{3, 2}, []int{1, 2})

	sliced.Apply([]int{0, 0}, 0, 1, []float64{10.0, 20.0, 30.0})

	// Should have written to columns 0 of the original
	assert.Equal(10.0, arr.Get2(0, 0))
	assert.Equal(20.0, arr.Get2(1, 0))
	assert.Equal(30.0, arr.Get2(2, 0))
	// Column 1 untouched
	assert.Equal(0.0, arr.Get2(0, 1))
}

func TestNewArray1DOperations(t *testing.T) {
	assert := assert.New(t)

	arr := NewArray1D[float64](5)
	for i := 0; i < 5; i++ {
		arr.Set1(i, float64(i*10))
	}

	assert.Equal(0.0, arr.Get1(0))
	assert.Equal(20.0, arr.Get1(2))
	assert.Equal(40.0, arr.Get1(4))
	assert.Equal(5, arr.Len1())

	arr.Apply1(1, 2, []float64{99.0, 88.0})
	assert.Equal(99.0, arr.Get1(1))
	assert.Equal(20.0, arr.Get1(2)) // untouched
	assert.Equal(88.0, arr.Get1(3))
}

func TestIntArrays(t *testing.T) {
	assert := assert.New(t)

	arr := NewArray2D[int](2, 3)
	arr.Set2(0, 0, 1)
	arr.Set2(0, 1, 2)
	arr.Set2(0, 2, 3)
	arr.Set2(1, 0, 4)
	arr.Set2(1, 1, 5)
	arr.Set2(1, 2, 6)

	assert.Equal(1, arr.Get2(0, 0))
	assert.Equal(6, arr.Get2(1, 2))
	assert.Equal(6, arr.Maximum())
	assert.Equal(1, arr.Minimum())

	unrolled := arr.Unroll()
	assert.Equal([]int{1, 2, 3, 4, 5, 6}, unrolled)
}

func TestInt32Arrays(t *testing.T) {
	assert := assert.New(t)

	arr := NewArray1D[int32](3)
	arr.Set1(0, 100)
	arr.Set1(1, 200)
	arr.Set1(2, 300)

	assert.Equal(int32(100), arr.Get1(0))
	assert.Equal(int32(300), arr.Maximum())
	assert.Equal(int32(100), arr.Minimum())
}

func TestFloat32Arrays(t *testing.T) {
	assert := assert.New(t)

	arr := NewArray1D[float32](3)
	arr.Set1(0, 1.5)
	arr.Set1(1, 2.5)
	arr.Set1(2, 0.5)

	assert.Equal(float32(1.5), arr.Get1(0))
	assert.Equal(float32(2.5), arr.Maximum())
	assert.Equal(float32(0.5), arr.Minimum())
}

func TestValuesContiguous(t *testing.T) {
	assert := assert.New(t)

	arr := newArray[float64]([]int{4})
	for i := 0; i < 4; i++ {
		arr.Set1(i, float64(i*10))
	}

	var collected []float64
	for v := range arr.Values([]int{0}, 0, 4) {
		collected = append(collected, v)
	}
	assert.Equal([]float64{0, 10, 20, 30}, collected)
}

func TestValuesNonContiguous(t *testing.T) {
	assert := assert.New(t)

	full := newArray[float64]([]int{3, 4})
	for i := 0; i < 12; i++ {
		full.Set([]int{i / 4, i % 4}, float64(i))
	}

	// Slice column 0: non-contiguous
	sliced := full.Slice([]int{0, 0}, []int{3, 1}, nil).(*nd[float64])

	var collected []float64
	for v := range sliced.Values([]int{0, 0}, 0, 3) {
		collected = append(collected, v)
	}
	assert.Equal([]float64{0, 4, 8}, collected)
}

func TestRawIndices(t *testing.T) {
	assert := assert.New(t)

	arr := newArray[float64]([]int{3, 4})
	// Row 1 starts at raw index 4, step along axis 1 is 1
	var indices []int
	for idx := range arr.RawIndices([]int{1, 0}, 1, 4) {
		indices = append(indices, idx)
	}
	assert.Equal([]int{4, 5, 6, 7}, indices)
}

func TestReshapeNonContiguousTo1D(t *testing.T) {
	assert := assert.New(t)

	// 2D array where the "series" dim is axis 1 (columns)
	arr := newArray[float64]([]int{1, 4})
	for i := 0; i < 4; i++ {
		arr.Set([]int{0, i}, float64(i*10))
	}

	// Slice to a non-contiguous view with shape [1, 4]
	sliced := arr.Slice([]int{0, 0}, []int{1, 4}, nil)

	reshaped, err := sliced.Reshape([]int{4})
	assert.Nil(err)

	arr1D := reshaped.(ND1[float64])
	assert.Equal(0.0, arr1D.Get1(0))
	assert.Equal(10.0, arr1D.Get1(1))
	assert.Equal(20.0, arr1D.Get1(2))
	assert.Equal(30.0, arr1D.Get1(3))
}

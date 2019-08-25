package cdata

import (
	"testing"

	"github.com/flowmatters/openwater-core/data"
	"github.com/flowmatters/openwater-core/util/slice"
	"github.com/stretchr/testify/assert"
)

//import "fmt"
// import "fmt"

func testData3D() data.ND3Float64 {
	lenI := 3
	lenJ := 2
	lenK := 4
	arr := data.NewArray3DFloat64(lenI, lenJ, lenK)

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

func testData2D() data.ND2Float64 {
	arr := data.NewArray2DFloat64(2, 4)
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
	arrSlice := arr.Slice([]int{1, 1, 1}, []int{2, 1, 2}, []int{1, 1, 1}).(data.ND3Float64)
	//arrNative := arrSlice.(*ndFloat64)

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
	arr := data.NewArray3DFloat64(20, 30, 10)
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

func testGet3(t *testing.T, arr data.ND3Float64, loc1, loc2, loc3 int, exp float64) {
	res := arr.Get3(loc1, loc2, loc3)
	if res != exp {
		t.Errorf("arr[%d,%d,%d] expected %f, got %f", loc1, loc2, loc3, exp, res)
	}
}

func testContig(t *testing.T, arr data.NDFloat64, expected bool) {
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
		arr1D := sliced.(data.ND1Float64)
		assert.Equal(t, int(1), arr1D.NDims())
		assert.Equal(t, 0.0, arr1D.Get1(0))
		assert.Equal(t, 5.0, arr1D.Get1(1))
	}
}

func TestReshapeFast(t *testing.T) {
	arr := testData2D()
	sliced, err := arr.Slice([]int{0, 0}, []int{1, 4}, nil).ReshapeFast([]int{4})

	if assert.Nil(t, err) {
		arr1D := sliced.(data.ND1Float64)
		assert.Equal(t, int(1), arr1D.NDims())
		assert.Equal(t, 0.0, arr1D.Get1(0))
		assert.Equal(t, 1.0, arr1D.Get1(1))
		assert.Equal(t, 35.0, arr1D.Get1(2))
		assert.Equal(t, 3.0, arr1D.Get1(3))
	}
}

func TestTreatAs1D(t *testing.T) {
	arr := testData2D()
	arr1D := arr.Slice([]int{0, 0}, []int{2, 1}, nil).(data.ND1Float64)

	//	assert.Equal(t,1,arr1D.NDims())
	assert.Equal(t, 0.0, arr1D.Get1(0))
	assert.Equal(t, 5.0, arr1D.Get1(1))

	alt1D := arr.Slice([]int{0, 0}, []int{1, 2}, nil).(data.ND1Float64)
	assert.Equal(t, 0.0, alt1D.Get([]int{0, 0}))
	assert.Equal(t, 1.0, alt1D.Get([]int{0, 1}))

	assert.Equal(t, 0.0, alt1D.Get1(0))
	assert.Equal(t, 1.0, alt1D.Get1(1))
}

func TestApplySlice(t *testing.T) {
	assert := assert.New(t)

	arr := testData2D()
	subst1D := data.NewArray1DFloat64(2)
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

	arr := data.ARangeFloat64(12.0).MustReshape([]int{3, 4}).(data.ND2Float64)

	expShape := []int{3, 4}
	assert.True(slice.Equal(expShape, arr.Shape()), "Slice shape should be %v. Got %v", expShape, arr.Shape())
	assert.Equal(0.0, arr.Get2(0, 0))
	assert.Equal(3.0, arr.Get2(0, 3))
	assert.Equal(4.0, arr.Get2(1, 0))
	assert.Equal(11.0, arr.Get2(2, 3))

}

func TestCArrayBasic(t *testing.T) {
	assert := assert.New(t)
	shape := []int{10, 5, 2}

	cArray := makefloat64CArrayForTest(shape)

	assert.Equal(0.0, cArray.Get([]int{0, 0, 0}))
	assert.Equal(1.0, cArray.Get([]int{0, 0, 1}))
	assert.Equal(2.0, cArray.Get([]int{0, 1, 0}))
	assert.Equal(10.0, cArray.Get([]int{1, 0, 0}))

	assert.Equal(99.0, cArray.Get([]int{9, 4, 1}))
}

func TestCArraySlice(t *testing.T) {
	assert := assert.New(t)
	shape := []int{10, 5, 2}

	cArray := makefloat64CArrayForTest(shape)

	sliced := cArray.Slice([]int{9, 3, 0}, []int{1, 2, 2}, nil)

	assert.Equal(96.0, sliced.Get([]int{0, 0, 0}))
	assert.Equal(97.0, sliced.Get([]int{0, 0, 1}))
	assert.Equal(98.0, sliced.Get([]int{0, 1, 0}))
	assert.Equal(99.0, sliced.Get([]int{0, 1, 1}))

	sliced2 := cArray.Slice([]int{0, 0, 1}, []int{10, 1, 1}, nil)
	assert.Equal(1.0, sliced2.Get([]int{0, 0, 0}))
	assert.Equal(11.0, sliced2.Get([]int{1, 0, 0}))
	assert.Equal(51.0, sliced2.Get([]int{5, 0, 0}))
	assert.Equal(91.0, sliced2.Get([]int{9, 0, 0}))
}

func TestCArraySliceUpdatesOriginal(t *testing.T) {
	assert := assert.New(t)
	shape := []int{10, 5, 2}

	cArray := makefloat64CArrayForTest(shape)

	sliced := cArray.Slice([]int{9, 3, 0}, []int{1, 2, 2}, nil)

	sliced.Set([]int{0, 0, 0}, 500.0)
	assert.Equal(500.0, cArray.Get([]int{9, 3, 0}))

	sliced.Set([]int{0, 1, 1}, 1000.0)
	assert.Equal(1000.0, cArray.Get([]int{9, 4, 1}))
}

func TestCArraySliceAndReshape(t *testing.T) {
	assert := assert.New(t)
	shape := []int{9, 1}

	cArray := makefloat64CArrayForTest(shape)

	sliced := cArray.Slice([]int{8, 0}, []int{1, 1}, nil)
	assert.Equal(8.0, sliced.Get([]int{0, 0}))

	reshaped := sliced.MustReshape([]int{1}).(data.ND1Float64)
	assert.Equal(8.0, reshaped.Get([]int{0}))
}

func testCopyFrom(t *testing.T, from, to data.ND2Float64) {
	assert := assert.New(t)
	rows := from.Len(0)
	cols := from.Len(1)

	i := 0

	var r, c int
	for r = 0; r < rows; r++ {
		for c = 0; c < cols; c++ {
			to.Set2(r, c, 0.0)
			from.Set2(r, c, float64(i))
			i++
		}
	}

	to.CopyFrom(from)
	i = 0

	for r = 0; r < rows; r++ {
		for c = 0; c < cols; c++ {
			assert.Equalf(float64(i), to.Get2(r, c), "Error at [%d,%d] in [%d,%d] array", r, c, rows, cols)
			i++
		}
	}

}
func TestCopyNativeToCArray(t *testing.T) {
	rows := int(9)
	cols := int(2)
	shape := []int{rows, cols}

	dest := makefloat64CArrayForTest(shape)
	src := data.NewArray2DFloat64(rows, cols)

	testCopyFrom(t, src, dest)
}

func TestCopyCArrayToNative(t *testing.T) {
	rows := int(9)
	cols := int(2)
	shape := []int{rows, cols}

	dest := data.NewArray2DFloat64(rows, cols)
	src := makefloat64CArrayForTest(shape)

	testCopyFrom(t, src, dest)
}

func TestCopyNativeToNative(t *testing.T) {
	rows := int(9)
	cols := int(2)

	dest := data.NewArray2DFloat64(rows, cols)
	src := data.NewArray2DFloat64(rows, cols)

	testCopyFrom(t, src, dest)
}

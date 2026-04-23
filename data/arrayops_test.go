package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyFunc1Contiguous(t *testing.T) {
	assert := assert.New(t)

	src := NewArray2D[float64](2, 3)
	dst := NewArray2D[float64](2, 3)
	for i := 0; i < 6; i++ {
		src.Set([]int{i / 3, i % 3}, float64(i))
	}

	ApplyFunc1[float64](dst, src, func(v float64) float64 { return v * 2 })

	assert.Equal(0.0, dst.Get2(0, 0))
	assert.Equal(2.0, dst.Get2(0, 1))
	assert.Equal(4.0, dst.Get2(0, 2))
	assert.Equal(6.0, dst.Get2(1, 0))
	assert.Equal(8.0, dst.Get2(1, 1))
	assert.Equal(10.0, dst.Get2(1, 2))
}

func TestApplyFunc1NonContiguous(t *testing.T) {
	assert := assert.New(t)

	full := ARange[float64](24).MustReshape([]int{4, 6})
	// Strided slice: rows 0,2 cols 0,2,4 — non-contiguous
	src := full.Slice([]int{0, 0}, []int{2, 3}, []int{2, 2})
	dst := NewArray2D[float64](2, 3)

	ApplyFunc1[float64](dst, src, func(v float64) float64 { return v + 100 })

	// src values: [0,0]=0, [0,2]=2, [0,4]=4, [2,0]=12, [2,2]=14, [2,4]=16
	assert.Equal(100.0, dst.Get2(0, 0))
	assert.Equal(102.0, dst.Get2(0, 1))
	assert.Equal(104.0, dst.Get2(0, 2))
	assert.Equal(112.0, dst.Get2(1, 0))
	assert.Equal(114.0, dst.Get2(1, 1))
	assert.Equal(116.0, dst.Get2(1, 2))
}

func TestScaleArray(t *testing.T) {
	assert := assert.New(t)

	src := NewArray1D[float64](3)
	dst := NewArray1D[float64](3)
	src.Set1(0, 2.0)
	src.Set1(1, 3.0)
	src.Set1(2, 5.0)

	ScaleArray[float64](dst, src, 10.0)

	assert.Equal(20.0, dst.Get1(0))
	assert.Equal(30.0, dst.Get1(1))
	assert.Equal(50.0, dst.Get1(2))
}

func TestAddToArrayContiguous(t *testing.T) {
	assert := assert.New(t)

	dst := NewArray1D[float64](4)
	src := NewArray1D[float64](4)
	for i := 0; i < 4; i++ {
		dst.Set1(i, float64(i*10))
		src.Set1(i, float64(i))
	}

	AddToArray[float64](dst, src)

	assert.Equal(0.0, dst.Get1(0))
	assert.Equal(11.0, dst.Get1(1))
	assert.Equal(22.0, dst.Get1(2))
	assert.Equal(33.0, dst.Get1(3))
}

func TestAddToArrayNonContiguous(t *testing.T) {
	assert := assert.New(t)

	full := ARange[float64](12).MustReshape([]int{3, 4})
	src := full.Slice([]int{0, 0}, []int{3, 1}, nil) // column 0: 0, 4, 8

	dst := NewArray[float64]([]int{3, 1})
	dst.Set([]int{0, 0}, 100.0)
	dst.Set([]int{1, 0}, 200.0)
	dst.Set([]int{2, 0}, 300.0)

	AddToArray[float64](dst, src)

	assert.Equal(100.0, dst.Get([]int{0, 0}))
	assert.Equal(204.0, dst.Get([]int{1, 0}))
	assert.Equal(308.0, dst.Get([]int{2, 0}))
}

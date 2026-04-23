package data

import (
	"testing"

	"github.com/flowmatters/openwater-core/util/slice"
	"github.com/stretchr/testify/assert"
)

func TestProduct(t *testing.T) {
	input := []int{1, 2, 3, 4}

	if Product(input) != 24 {
		t.Errorf("expect product(%q) == 24, was %d", input, Product(input))
	}
}

func TestCumulativeProduct(t *testing.T) {
	input := []int{1, 2, 3, 4}
	res := cumulProduct(input)
	expected := []int{1, 2, 6, 24}
	if !slice.Equal(res, expected) {
		t.Errorf("expect cumulProduct(%q) == %q, was %q", input, expected, res)
	}
}

func TestDotProduct(t *testing.T) {
	lhs := []int{1, 3, 5}
	rhs := []int{2, 6, 10}
	exp := 70
	res := dotProduct(lhs, rhs)

	if res != exp {
		t.Errorf("dotProduct(%q,%q) should equal %d, but was %d", lhs, rhs, exp, res)
	}
}

func TestMultiply(t *testing.T) {
	lhs := []int{1, 3, 5}
	rhs := []int{2, 6, 10}
	exp := []int{2, 18, 50}
	res := Multiply(lhs, rhs)

	if !slice.Equal(res, exp) {
		t.Errorf("multiply(%q,%q) should equal %q, but was %q", lhs, rhs, exp, res)
	}
}

func TestDecrement(t *testing.T) {
	test := []int{10, 12, 9}
	exp := []int{9, 11, 8}
	res := decrement(test)

	if !slice.Equal(res, exp) {
		t.Errorf("decr(%v) should be %v, but was %v", test, exp, res)
	}
}

func TestIncrement(t *testing.T) {
	assert := assert.New(t)

	shape := []int{9, 2}

	vec := []int{0, 0}
	Increment(vec, shape)
	assert.Equal([]int{0, 1}, vec)

	vec = []int{3, 1}
	Increment(vec, shape)
	assert.Equal([]int{4, 0}, vec)

	vec = []int{5, 1}
	Increment(vec, shape)
	assert.Equal([]int{6, 0}, vec)

	shape = []int{5, 3, 4}

	vec = []int{0, 0, 0}
	Increment(vec, shape)
	assert.Equal([]int{0, 0, 1}, vec)

	vec = []int{3, 1, 1}
	Increment(vec, shape)
	assert.Equal([]int{3, 1, 2}, vec)

	vec = []int{3, 1, 3}
	Increment(vec, shape)
	assert.Equal([]int{3, 2, 0}, vec)

	vec = []int{3, 2, 3}
	Increment(vec, shape)
	assert.Equal([]int{4, 0, 0}, vec)
}

func TestIncrementWrapsToZero(t *testing.T) {
	assert := assert.New(t)

	shape := []int{2, 3}
	vec := []int{1, 2} // last valid index
	Increment(vec, shape)
	assert.Equal([]int{0, 0}, vec)
}

func TestOffsets(t *testing.T) {
	assert := assert.New(t)

	assert.Equal([]int{1}, Offsets([]int{5}))
	assert.Equal([]int{4, 1}, Offsets([]int{3, 4}))
	assert.Equal([]int{12, 4, 1}, Offsets([]int{2, 3, 4}))
}

func TestIDivMod(t *testing.T) {
	assert := assert.New(t)

	// For a 2x3x4 array, offsets are [12, 4, 1], dims are [2, 3, 4]
	dims := []int{2, 3, 4}
	offsets := Offsets(dims)

	// Linear index 0 -> [0,0,0]
	assert.Equal([]int{0, 0, 0}, IDivMod(0, offsets, dims))
	// Linear index 5 -> [0,1,1]
	assert.Equal([]int{0, 1, 1}, IDivMod(5, offsets, dims))
	// Linear index 23 -> [1,2,3]
	assert.Equal([]int{1, 2, 3}, IDivMod(23, offsets, dims))
}

func TestArgmax(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(0, Argmax([]int{5, 1, 2}))
	assert.Equal(1, Argmax([]int{1, 5, 2}))
	assert.Equal(2, Argmax([]int{1, 2, 5}))
	assert.Equal(0, Argmax([]int{5}))
	assert.Equal(2, Argmax([]int{1, 1, 4}))
}

func TestMaximumSlice(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(5, Maximum([]int{5, 1, 2}))
	assert.Equal(5, Maximum([]int{1, 5, 2}))
	assert.Equal(5, Maximum([]int{1, 2, 5}))
}

package fn

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrackets(t *testing.T) {
	assert := assert.New(t)

	xs := []float64{0.0, 1.0, 2.0, 5.0, 10.0}

	// Test value in first segment
	i, j := brackets(0.5, xs)
	assert.Equal(0, i)
	assert.Equal(1, j)

	// Test value in middle segment
	i, j = brackets(3.0, xs)
	assert.Equal(2, i)
	assert.Equal(3, j)

	// Test value in last segment
	i, j = brackets(7.5, xs)
	assert.Equal(3, i)
	assert.Equal(4, j)

	// Test exact match at first point
	i, j = brackets(0.0, xs)
	assert.Equal(0, i)
	assert.Equal(1, j)

	// Test exact match at last point
	i, j = brackets(10.0, xs)
	assert.Equal(3, i)
	assert.Equal(4, j)

	// Test exact match at middle point
	i, j = brackets(2.0, xs)
	assert.Equal(1, i)
	assert.Equal(2, j)

	// Test value below range
	i, j = brackets(-1.0, xs)
	assert.Equal(-1, i)
	assert.Equal(-1, j)

	// Test value above range
	i, j = brackets(15.0, xs)
	assert.Equal(-1, i)
	assert.Equal(-1, j)
}

func TestPiecewise_SimpleLinearInterpolation(t *testing.T) {
	assert := assert.New(t)

	// Simple two-point line: y = 2*x
	xs := []float64{0.0, 10.0}
	ys := []float64{0.0, 20.0}

	// Test midpoint
	y, err := Piecewise(5.0, xs, ys)
	assert.Nil(err)
	assert.Equal(10.0, y)

	// Test at exact points
	y, err = Piecewise(0.0, xs, ys)
	assert.Nil(err)
	assert.Equal(0.0, y)

	y, err = Piecewise(10.0, xs, ys)
	assert.Nil(err)
	assert.Equal(20.0, y)

	// Test 1/4 point
	y, err = Piecewise(2.5, xs, ys)
	assert.Nil(err)
	assert.Equal(5.0, y)

	// Test 3/4 point
	y, err = Piecewise(7.5, xs, ys)
	assert.Nil(err)
	assert.Equal(15.0, y)
}

func TestPiecewise_MultipleSegments(t *testing.T) {
	assert := assert.New(t)

	// Piecewise linear function with different slopes
	xs := []float64{0.0, 2.0, 5.0, 10.0}
	ys := []float64{0.0, 4.0, 7.0, 12.0}

	// Test in first segment (0 to 2): slope = 2
	y, err := Piecewise(1.0, xs, ys)
	assert.Nil(err)
	assert.Equal(2.0, y)

	// Test in second segment (2 to 5): slope = 1
	y, err = Piecewise(3.5, xs, ys)
	assert.Nil(err)
	assert.Equal(5.5, y)

	// Test in third segment (5 to 10): slope = 1
	y, err = Piecewise(7.5, xs, ys)
	assert.Nil(err)
	assert.Equal(9.5, y)

	// Test at segment boundaries
	y, err = Piecewise(2.0, xs, ys)
	assert.Nil(err)
	assert.Equal(4.0, y)

	y, err = Piecewise(5.0, xs, ys)
	assert.Nil(err)
	assert.Equal(7.0, y)
}

func TestPiecewise_ConstantSegment(t *testing.T) {
	assert := assert.New(t)

	// Function with a flat segment
	xs := []float64{0.0, 5.0, 10.0}
	ys := []float64{0.0, 10.0, 10.0}

	// Test in increasing segment
	y, err := Piecewise(2.5, xs, ys)
	assert.Nil(err)
	assert.Equal(5.0, y)

	// Test in flat segment
	y, err = Piecewise(7.5, xs, ys)
	assert.Nil(err)
	assert.Equal(10.0, y)

	y, err = Piecewise(5.0, xs, ys)
	assert.Nil(err)
	assert.Equal(10.0, y)

	y, err = Piecewise(10.0, xs, ys)
	assert.Nil(err)
	assert.Equal(10.0, y)
}

func TestPiecewise_DecreasingFunction(t *testing.T) {
	assert := assert.New(t)

	// Decreasing function
	xs := []float64{0.0, 5.0, 10.0}
	ys := []float64{100.0, 50.0, 0.0}

	// Test in first segment
	y, err := Piecewise(2.5, xs, ys)
	assert.Nil(err)
	assert.Equal(75.0, y)

	// Test in second segment
	y, err = Piecewise(7.5, xs, ys)
	assert.Nil(err)
	assert.Equal(25.0, y)
}

func TestPiecewise_OutOfBounds(t *testing.T) {
	assert := assert.New(t)

	xs := []float64{0.0, 5.0, 10.0}
	ys := []float64{0.0, 10.0, 20.0}

	// Test below range
	_, err := Piecewise(-1.0, xs, ys)
	assert.NotNil(err)
	assert.Contains(err.Error(), "Couldn't find brackets")

	// Test above range
	_, err = Piecewise(15.0, xs, ys)
	assert.NotNil(err)
	assert.Contains(err.Error(), "Couldn't find brackets")
}

func TestPiecewise_NegativeValues(t *testing.T) {
	assert := assert.New(t)

	// Function with negative x and y values
	xs := []float64{-10.0, -5.0, 0.0, 5.0}
	ys := []float64{-20.0, -10.0, 0.0, 10.0}

	y, err := Piecewise(-7.5, xs, ys)
	assert.Nil(err)
	assert.Equal(-15.0, y)

	y, err = Piecewise(-2.5, xs, ys)
	assert.Nil(err)
	assert.Equal(-5.0, y)

	y, err = Piecewise(2.5, xs, ys)
	assert.Nil(err)
	assert.Equal(5.0, y)
}

func TestPiecewise_SmallIntervals(t *testing.T) {
	assert := assert.New(t)

	// Very small intervals
	xs := []float64{0.0, 0.001, 0.002}
	ys := []float64{0.0, 1.0, 2.0}

	y, err := Piecewise(0.0005, xs, ys)
	assert.Nil(err)
	assert.InDelta(0.5, y, 1e-10)

	y, err = Piecewise(0.0015, xs, ys)
	assert.Nil(err)
	assert.InDelta(1.5, y, 1e-10)
}

func TestPiecewise_LargeIntervals(t *testing.T) {
	assert := assert.New(t)

	// Very large intervals
	xs := []float64{0.0, 1e6, 2e6}
	ys := []float64{0.0, 1e3, 2e3}

	y, err := Piecewise(5e5, xs, ys)
	assert.Nil(err)
	assert.InDelta(500.0, y, 1e-6)

	y, err = Piecewise(1.5e6, xs, ys)
	assert.Nil(err)
	assert.InDelta(1500.0, y, 1e-6)
}

func TestPiecewise_NonUniformSpacing(t *testing.T) {
	assert := assert.New(t)

	// Non-uniform spacing between points
	xs := []float64{0.0, 1.0, 2.0, 10.0, 11.0}
	ys := []float64{0.0, 10.0, 15.0, 20.0, 30.0}

	// Test in short segment
	y, err := Piecewise(0.5, xs, ys)
	assert.Nil(err)
	assert.Equal(5.0, y)

	// Test in long segment
	y, err = Piecewise(6.0, xs, ys)
	assert.Nil(err)
	assert.Equal(17.5, y)

	// Test in another short segment
	y, err = Piecewise(10.5, xs, ys)
	assert.Nil(err)
	assert.Equal(25.0, y)
}

func TestPiecewise_AccuracyCheck(t *testing.T) {
	assert := assert.New(t)

	// Test known mathematical function approximation
	// Approximate sine function with piecewise linear
	xs := []float64{0.0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4, math.Pi}
	ys := []float64{0.0, math.Sqrt(2) / 2, 1.0, math.Sqrt(2) / 2, 0.0}

	// Test at quarter points
	y, err := Piecewise(math.Pi/4, xs, ys)
	assert.Nil(err)
	assert.InDelta(math.Sqrt(2)/2, y, 1e-10)

	// Test interpolation between quarter and half
	y, err = Piecewise(3*math.Pi/8, xs, ys)
	assert.Nil(err)
	expectedY := (math.Sqrt(2)/2 + 1.0) / 2
	assert.InDelta(expectedY, y, 1e-10)
}

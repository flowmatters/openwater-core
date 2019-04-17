package fn

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolynomal(t *testing.T) {
	assert := assert.New(t)

	the_fn := func(x float64) float64 {
		return 0.5*math.Pow(x, 2) + 4*x - 3
	}

	the_deriv := func(x float64) float64 {
		return x - 4
	}

	root, _ := FindRoot(the_fn, the_deriv, 0.5, 0.0, 2.0, 1e-6, 1e-15, 10)

	assert.True(math.Abs(0.690415-root) < 1e-6)
}

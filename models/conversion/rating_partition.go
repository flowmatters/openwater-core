package conversion

import (
	"fmt"
	"math"

	"github.com/flowmatters/openwater-core/util/fn"
)

/*OW-SPEC
RatingCurvePartition:
  inputs:
		input:
	states:
	parameters:
		nPts: ''
		inputAmount[nPts]:
		proportion[nPts]:
	outputs:
		output1:
		output2:
	implementation:
		function: ratingPartition
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		partition
*/

func ratingPartition(input []float64,
	nPts int,
	inputAmount, proportion []float64,
	output1, output2 []float64) {

	nDays := len(input)

	for i := 0; i < nDays; i++ {
		incoming := input[i]
		frac, err := fn.Piecewise(incoming, inputAmount, proportion)
		if err != nil {
			panic(err)
		}

		if math.IsNaN(frac) || math.IsNaN(incoming) {
			fmt.Printf("timestep=%d/%d\n", i, nDays)
			fmt.Printf("frac=%f\n", frac)
			fmt.Printf("incoming=%f\n", incoming)
			fmt.Printf("inputAmount=%v\n", inputAmount)
			fmt.Printf("proportion=%v\n", proportion)
			panic("nan")
		}

		output1[i] = incoming * frac
		output2[i] = incoming * (1 - frac)
	}
}

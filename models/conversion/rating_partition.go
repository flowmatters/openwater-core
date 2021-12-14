package conversion

import (
	"github.com/flowmatters/openwater-core/data"
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

func ratingPartition(input data.ND1Float64,
	nPts int,
	inputAmount, proportion data.ND1Float64,
	output1, output2 data.ND1Float64) {

	nDays := input.Len1()
	idx := []int{0}

	for i := 0; i < nDays; i++ {
		idx[0] = i
		incoming := input.Get(idx)
		frac,err := fn.Piecewise(incoming,inputAmount,proportion)
		if err != nil {
			panic(err)
		}

		output1.Set(idx, incoming*frac)
		output2.Set(idx, incoming*(1-frac))
	}
}

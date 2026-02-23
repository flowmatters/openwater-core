package conversion

import (
	"math"

	"github.com/rs/zerolog/log"

	"github.com/flowmatters/openwater-core/util/fn"
)

/*OW-SPEC
RatingCurvePartition:
  symbol: RP
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
			log.Panic().Stack().Err(err).Msg("")
		}

		if math.IsNaN(frac) || math.IsNaN(incoming) {
			log.Panic().Int("timestep", i).
				Int("nDays", nDays).
				Float64("frac", frac).
				Float64("incoming", incoming).
				Any("inputAmount", inputAmount).
				Any("proportion", proportion).
				Err(err).Msg("NAN")
		}

		output1[i] = incoming * frac
		output2[i] = incoming * (1 - frac)
	}
}

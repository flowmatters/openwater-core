package rr

import (
	"fmt"

	"github.com/flowmatters/openwater-core/data"
)

/* OW-SPEC
RunoffCoefficient:
  inputs:
    rainfall: mm
  states:
  parameters:
    coeff: ''
	outputs:
		runoff: mm
	implementation:
		function: runoffCoefficient
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
		lang: go
	tags:
		rainfall runoff test

*/

func runoffCoefficient(rainfall data.ND1Float64, coeff float64, runoff data.ND1Float64) {
	n := rainfall.Len1()
	fmt.Println("Running for ", n, "days")
	for i := 0; i < n; i++ {
		runoff.Set1(i, coeff*rainfall.Get1(i))
	}
}

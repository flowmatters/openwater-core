package io

import (
	"fmt"
	"math"

	"github.com/flowmatters/openwater-core/data"
)

func JsonSafeArray(vals data.NDFloat64, shiftDim int) []interface{} {
	shape := vals.Shape()
	length := vals.Len(shiftDim)
	ndims := vals.NDims()
	from := vals.NewIndex(0)
	to := vals.NewIndex(0)
	step := vals.NewIndex(1)

	for i := (shiftDim + 1); i < ndims; i++ {
		to[i] = shape[i]
	}

	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		from[shiftDim] = i
		if shiftDim == (ndims - 1) {
			v := vals.Get(from)
			result[i] = JsonSafeValue(v)
		} else {
			to[shiftDim] = i
			result[i] = JsonSafeArray(vals.Slice(from, to, step), shiftDim+1)
		}
	}
	return result
}

func JsonSafeValue(val float64) interface{} {
	if math.IsNaN(val) {
		return fmt.Sprint(val)
	} else if math.IsInf(val, 0) {
		return fmt.Sprint(val)
	}
	return val
}
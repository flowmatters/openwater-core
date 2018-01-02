package io

import (
	"github.com/flowmatters/openwater-core/data"
	"math"
  "fmt"
)

func JsonSafeArray(vals data.NDFloat64,shiftDim int) []interface{} {
  shape := vals.Shape()
  length := vals.Len(shiftDim)
  ndims := vals.NDims()
  from := vals.NewIndex(0)
  to := vals.NewIndex(0)
  step := vals.NewIndex(1)
  
  for i := (shiftDim+1); i < ndims; i++ {
    to[i] = shape[i]
  }

  result := make([]interface{},length)
  for i := 0; i < length; i++ {
    from[shiftDim] = i
    if shiftDim==(ndims-1){
      v := vals.Get(from)
      if math.IsNaN(v) {
        result[i] = fmt.Sprint(v)
      } else if math.IsInf(v,0) {
        result[i] = fmt.Sprint(v)
      } else {
        result[i] = v
      }
    } else {
      to[shiftDim] = i
      result[i] = JsonSafeArray(vals.Slice(from,to,step),shiftDim+1)
    }
  } 
  return result
}

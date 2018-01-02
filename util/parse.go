package util

import (
//	"fmt"
  "strconv"
  "math"
)

func ParseFloatNaN(s string) float64 {
  res,err := strconv.ParseFloat(s,64)

  if err != nil{
//    fmt.Println(s)
    return math.NaN()
  }

  return res
}

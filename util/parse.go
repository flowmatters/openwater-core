package util

import (
	//	"fmt"
	"math"
	"strconv"
)

func ParseFloatNaN(s string) float64 {
	res, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return math.NaN()
	}

	return res
}

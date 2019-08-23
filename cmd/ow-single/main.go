package main

import (
	"os"

	_ "github.com/flowmatters/openwater-core/models"
	"github.com/flowmatters/openwater-core/sim"
)

// func jsonSafe2D(vals data.NDFloat64, shiftDim int) [][]interface{} {
//   result := make([][]interface{}, len(vals))
//   for i,v := range(vals){
//     result[i] = jsonSafe(v)
//   }
//   return result
// }

// func jsonSafe(vals data.NDFloat64) []interface{} {
//   result := make([]interface{}, len(vals))
//   for i,v := range(vals){
//     if math.IsNaN(v) {
//       result[i] = fmt.Sprint(v)
//     } else if math.IsInf(v,0) {
//       result[i] = fmt.Sprint(v)
//     } else {
//       result[i] = v
//     }
//   }
//   return result
// }

func main() {
	sim.RunSingleModelJSON(os.Stdin, os.Stdout)
}

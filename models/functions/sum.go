package functions

/*OW-SPEC
Sum:
	inputs:
		i1:
		i2:
	states:
	parameters:
	outputs:
		out:
	implementation:
		function: sum
		type: scalar
		lang: go
		outputs: params
	init:
		zero: true
	tags:
		function
*/

func sum(i1, i2 []float64,
	out []float64) {

	n := len(i1)

	for day := 0; day < n; day++ {
		s := i1[day] + i2[day]
		out[day] = s
	}
}

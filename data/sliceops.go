package data

func product(ix []int) int {
	result := 1
	for _, v := range ix {
		result *= v
	}
	return result
}

func cumulProduct(ix []int) []int {
	result := make([]int, len(ix))
	product := 1
	for i, v := range ix {
		product *= v
		result[i] = product
	}
	return result
}

func dotProduct(lhs []int, rhs []int) int {
	result := 0
	for i := 0; i < len(lhs); i++ {
		result += lhs[i] * rhs[i]
	}
	return result
}

func multiply(lhs []int, rhs []int) []int {
	result := make([]int, len(lhs))
	for i := 0; i < len(lhs); i++ {
		result[i] = lhs[i] * rhs[i]

	}
	return result
}

func decrement(vector []int) []int {
	result := make([]int, len(vector))
	for i := 0; i < len(vector); i++ {
		result[i] = vector[i] - 1
	}

	return result
}

func idiv(numerator int, denominators []int) []int {
	res := make([]int, len(denominators))
	for i := range denominators {
		res[i] = numerator / denominators[i]
	}
	return res
}

func mod(numerator int, denominators []int) []int {
	res := make([]int, len(denominators))
	for i := range denominators {
		res[i] = numerator % denominators[i]
	}
	return res
}

func increment(vector, wrt []int) {
	dims := len(wrt)
	for i := (dims - 1); i >= 0; i-- {
		vector[i]++
		if vector[i] >= wrt[i] {
			vector[i]--
		} else {
			return
		}
	}
}

func argmax(vector []int) int {
	res := 0
	maxFound := vector[0]
	for i, v := range vector[1:] {
		if v > maxFound {
			maxFound = v
			res = i
		}
	}
	return res
}

func maximum(vector []int) int {
	res := vector[0]
	for _, v := range vector[1:] {
		res = max(res, v)
	}
	return res
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copyF(dest []float64, src []float64) {
	n := min(len(dest), len(src))
	for i := 0; i < n; i++ {
		dest[i] = src[i]
	}
}

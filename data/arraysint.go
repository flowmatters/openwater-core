package data

func offsets(dims []int) []int {
	res := make([]int, len(dims))
	res[len(dims)-1] = 1
	for i := len(dims) - 2; i >= 0; i-- {
		res[i] = res[i+1] * dims[i+1]
	}
	return res
}

func idivMod(numerator int, denominators []int, modulator []int) []int {
	res := make([]int, len(denominators))
	for i := range denominators {
		res[i] = (numerator / denominators[i]) % modulator[i]
	}
	return res
}

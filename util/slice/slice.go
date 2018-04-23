package slice

func Equal(lhs []int, rhs []int) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i := range lhs {
		if lhs[i] != rhs[i] {
			return false
		}
	}

	return true
}

func Uniform(ndims int, val int) []int {
	result := make([]int, ndims)
	if val == 0 {
		return result
	}

	for i := 0; i < ndims; i++ {
		result[i] = val
	}
	return result
}

func Ones(ndims int) []int {
	return Uniform(ndims, 1)
}

package data

import (
	"testing"

	"github.com/flowmatters/openwater-core/util/slice"
)

func TestProduct(t *testing.T) {
	input := []int{1, 2, 3, 4}

	if product(input) != 24 {
		t.Errorf("expect product(%q) == 24, was %d", input, product(input))
	}
}

func TestCumulativeProduct(t *testing.T) {
	input := []int{1, 2, 3, 4}
	res := cumulProduct(input)
	expected := []int{1, 2, 6, 24}
	if !slice.Equal(res, expected) {
		t.Errorf("expect cumulProduct(%q) == %q, was %q", input, expected, res)
	}
}

func TestDotProduct(t *testing.T) {
	lhs := []int{1, 3, 5}
	rhs := []int{2, 6, 10}
	exp := 70
	res := dotProduct(lhs, rhs)

	if res != exp {
		t.Errorf("dotProduct(%q,%q) should equal %d, but was %d", lhs, rhs, exp, res)
	}
}

func TestMultiply(t *testing.T) {
	lhs := []int{1, 3, 5}
	rhs := []int{2, 6, 10}
	exp := []int{2, 18, 50}
	res := multiply(lhs, rhs)

	if !slice.Equal(res, exp) {
		t.Errorf("multiply(%q,%q) should equal %q, but was %q", lhs, rhs, exp, res)
	}
}

func TestDecrement(t *testing.T) {
	test := []int{10, 12, 9}
	exp := []int{9, 11, 8}
	res := decrement(test)

	if !slice.Equal(res, exp) {
		t.Errorf("decr(%v) should be %v, but was %v", test, exp, res)
	}
}

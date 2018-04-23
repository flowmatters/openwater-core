package slice

import (
	"testing"
)

func TestEqual(t *testing.T) {
	lhs := []int{1, 2, 3, 4}
	eq := []int{1, 2, 3, 4}
	ne := []int{1, 2, 3, 0}
	diffLen := []int{1, 2, 3}

	if !Equal(lhs, eq) {
		t.Errorf("%q should equal %q", lhs, eq)
	}

	if Equal(lhs, ne) {
		t.Errorf("%q should not equal %q", lhs, ne)
	}

	if Equal(lhs, diffLen) {
		t.Errorf("%q should not equal %q", lhs, diffLen)
	}
}

func TestOnes(t *testing.T) {
	n := 5
	expected := []int{1, 1, 1, 1, 1}
	res := Ones(n)
	if !Equal(expected, res) {
		t.Errorf("ones(%d) should equal %q, but was %q", n, expected, res)
	}
}

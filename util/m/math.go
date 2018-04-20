package m

import "github.com/joelrahman/genny/generic"

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "NumT=NUMBERS"

type NumT generic.Number

func MinNumT(a, b NumT) NumT {
	if a > b {
		return b
	}

	return a
}

func MaxNumT(a, b NumT) NumT {
	if a > b {
		return a
	}

	return b
}

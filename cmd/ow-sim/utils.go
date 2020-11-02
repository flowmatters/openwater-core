package main

import (
	"fmt"
)

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func prefix(msg string, e error) error {
	return &errorString{msg + e.Error()}
}


func verbosePrintln(a ...interface{}) {
	if verbose {
		fmt.Println(a...)
	}
}

func verbosePrintf(s string, a ...interface{}) {
	if verbose {
		fmt.Printf(s, a...)
	}
}

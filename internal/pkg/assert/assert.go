package assert

import (
	"runtime"

	"go.octolab.org/errors"
)

var disabled bool

const (
	AssertionIsNotATrue = errors.Message("assertion is not a true")
)

type BugReport struct {
	error
	*stack
}

func True(check func() bool, optional ...error) {
	if disabled {
		return
	}

	if check() {
		return
	}

	var err error = AssertionIsNotATrue
	if len(optional) > 0 {
		err = optional[0]
	}
	panic(BugReport{err, callers()})
}

type stack []uintptr

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

package io

import "unix"

type Error struct {
	message string
}

var EOF *Error

func init() {
	EOF = &Error{
		message: "*EOF*",
	}
}

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

var alreadyCalledInit bool

func init() {
	// This has nothing to do with io, but I want to be sure this problem does not recur.
	if alreadyCalledInit {
		panic("io.init() ran twice")
	}
	alreadyCalledInit = true
}

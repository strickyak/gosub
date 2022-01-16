package io

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

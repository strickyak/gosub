package io

type Error struct{}

var EOF *Error

func init() {
	EOF = &Error{}
}

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

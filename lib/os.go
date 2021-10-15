package os

type File struct{}

var Stdin *File
var Stdout *File
var Stderr *File

func (f *File) Read(p []byte) (n int, err error)
func (f *File) Write(p []byte) (n int, err error)

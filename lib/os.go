package os

type File struct {
	fd int
}

var Stdin *File
var Stdout *File
var Stderr *File

func (f *File) Read(p []byte) (n int, err error)
func (f *File) Write(p []byte) (n int, err error)

func init() {
	Stdin = &File{fd: 0}
	Stdout = &File{fd: 1}
	Stderr = &File{fd: 2}
}

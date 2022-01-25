package os

import "errors"
import "io"
import "unix"
import "unsafe"

type File struct {
	fd int
}

var Stdin *File = &File{fd: 0}
var Stdout *File
var Stderr *File

func (f *File) Read(p []byte) (n int, err error) {
	start := unsafe.AddrOfFirstElement(p)
	cc, errno := unix.Read(f.fd, start, len(p))
	if errno != 0 {
		return cc, errors.New("cannot read")
	}
	if cc == 0 {
		return 0, io.EOF
	}
	return cc, nil
}
func (f *File) Write(p []byte) (n int, err error) {
	start := unsafe.AddrOfFirstElement(p)
	cc, errno := unix.Write(f.fd, start, len(p))
	if errno != 0 {
		return cc, errors.New("cannot read")
	}
	if cc == 0 {
		panic("unix.Write succeeded, but cc==0")
	}
	return cc, nil
}

func init() {
	Stdin = &File{fd: 0}
	Stdout = &File{fd: 1}
	Stderr = &File{fd: 2}
}

package os

import "errors"
import "io"
import "low"
import "unsafe"

type File struct {
	fd int
}

var Stdin *File = &File{fd: 0}
var Stdout *File = &File{fd: 1}
var Stderr *File = &File{fd: 2}

func (f *File) Read(p []byte) (n int, err error) {
	start := unsafe.AddrOfFirstElement(p)
	cc, errno := low.Read(f.fd, start, len(p))
	// log.Printf("^Read^%d^%d^", cc, errno)
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
	cc, errno := low.Write(f.fd, start, len(p))
	// log.Printf("^Write^%d^%d^", cc, errno)
	if errno != 0 {
		return cc, errors.New("cannot read")
	}
	if cc == 0 {
		panic("low.Write succeeded, but cc==0")
	}
	return cc, nil
}

func init() {
	if io.EOF == nil {
		panic("did not init io.EOF")
	}
}

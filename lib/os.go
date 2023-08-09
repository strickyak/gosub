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
	if errno != 0 {
		return cc, errors.New("cannot write")
	}
	if cc == 0 {
		panic("low.Write succeeded, but cc==0")
	}
	return cc, nil
}

func (f *File) ReadLine() (s string, err error) {
	start := low.StaticBufferAddress()
	cc, errno := low.ReadLn(f.fd, start, 250)
	if errno != 0 {
		return "", errors.New("cannot ReadLn")
	}
	if cc == 0 {
		return "", io.EOF
	}
	return low.StaticBufferToString(), nil
}
func (f *File) WriteLine(line string) (err error) {
	start := unsafe.AddrOfFirstElement(line)
	cc, errno := low.WritLn(f.fd, start, len(line))
	if errno != 0 {
		return errors.New("cannot write")
	}
	if cc == 0 {
		panic("low.WritLn succeeded, but cc==0")
	}
	return nil
}

func init() {
	if io.EOF == nil {
		panic("did not init io.EOF")
	}
}

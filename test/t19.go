package main

import "log"

// const k = 5

func MkStrings(s string) []string {
	var z []string
	z = append(z, s+"-one")
	z = append(z, s+"-two")
	z = append(z, s+"-three")
	z = append(z, s+"-four")
	return z
}

func MkSlices(s string) [][]string {
	var z [][]string
	z = append(z, MkStrings(s+"=cat"))
	z = append(z, MkStrings(s+"=dog"))
	z = append(z, MkStrings(s+"=bird"))
	return z
}

func MkDeepSlices() [][][]string {
	var z [][][]string
	z = append(z, MkSlices("boron"))
	z = append(z, MkSlices("carbon"))
	return z
}

func main() {
	ds := MkDeepSlices()
	// log.Printf("Nando len(ds)=%d T=%T, lenT=%T", len(ds), ds, len(ds))
	for i, e := range ds {
		// log.Printf("Nando i, eT = %d, %T", i, e)
		for j, f := range e {
			// log.Printf("Nando j, fT = %d, %T", i, e)
			for k, g := range f {
				// log.Printf("Nando [%d %d %d] T=%T %v", i, j, k, g, g)
				println(i, j, k, g)
			}
		}
	}
}

// expect: 0 0 0 boron=cat-one
// expect: 0 0 1 boron=cat-two
// expect: 0 0 2 boron=cat-three
// expect: 0 0 3 boron=cat-four
// expect: 0 1 0 boron=dog-one
// expect: 0 1 1 boron=dog-two
// expect: 0 1 2 boron=dog-three
// expect: 0 1 3 boron=dog-four
// expect: 0 2 0 boron=bird-one
// expect: 0 2 1 boron=bird-two
// expect: 0 2 2 boron=bird-three
// expect: 0 2 3 boron=bird-four
// expect: 1 0 0 carbon=cat-one
// expect: 1 0 1 carbon=cat-two
// expect: 1 0 2 carbon=cat-three
// expect: 1 0 3 carbon=cat-four
// expect: 1 1 0 carbon=dog-one
// expect: 1 1 1 carbon=dog-two
// expect: 1 1 2 carbon=dog-three
// expect: 1 1 3 carbon=dog-four
// expect: 1 2 0 carbon=bird-one
// expect: 1 2 1 carbon=bird-two
// expect: 1 2 2 carbon=bird-three
// expect: 1 2 3 carbon=bird-four

package main

// import "fmt"

// const k = 5

func MkStrings(s string) []string {
	var z []string
	z = append(z, s+"-one")
	z = append(z, s+"-two")
	z = append(z, s+"-three")
	return z
}

func MkSlices() [][]string {
	var z [][]string
	z = append(z, MkStrings("cat"))
	z = append(z, MkStrings("dog"))
	z = append(z, MkStrings("bird"))
	return z
}

func MkDeepSlices() [][][]string {
	var z [][][]string
	z = append(z, MkSlices())
	z = append(z, MkSlices())
	z = append(z, MkSlices())
	return z
}

func main() {
	a := MkDeepSlices()
	println(a)
}

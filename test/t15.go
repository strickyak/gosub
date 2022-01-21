package main

import "fmt"

func main() {
	println(fmt.Sprintf("abc%sxyz", "lmnop"))
}

// expect: abclmnopxyz

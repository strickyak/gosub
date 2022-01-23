package main

import "fmt"

func main() {
	println(fmt.Sprintf("abc%sxyz", "lmnop"))
	println(fmt.Sprintf("abc%dxyz", 12345))
	println(fmt.Sprintf("abc%dxyz", -12345))
}

// expect: abclmnopxyz
// expect: abc12345xyz
// expect: abc-12345xyz

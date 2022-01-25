package main

import "fmt"
import "os"

func main() {
	println(fmt.Sprintf("abc%sxyz", "lmnop"))
	println(fmt.Sprintf("abc%dxyz", 12345))
	println(fmt.Sprintf("abc%dxyz", -12345))
	fmt.Fprintf(os.Stdout, "abc%dxyz\n", 0)
}

// expect: abclmnopxyz
// expect: abc12345xyz
// expect: abc-12345xyz
// expect: abc0xyz

package main

import "fmt"
import "os"

func main() {
	x := fmt.Sprintf("abc%sxyz", "lmnop")
	y := fmt.Sprintf("abc%dxyz", 12345)
	z := fmt.Sprintf("abc%dxyz", -12345)
	fmt.Fprintf(os.Stdout, "%s\n", x)
	fmt.Fprintf(os.Stdout, "%s\n", y)
	fmt.Fprintf(os.Stdout, "%s\n", z)
	fmt.Fprintf(os.Stdout, "abc%dxyz\n", 0)
}

// expect: abclmnopxyz
// expect: abc12345xyz
// expect: abc-12345xyz
// expect: abc0xyz

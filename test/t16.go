package main

import "fmt"

var Answer int = 42
var Zero int

func main() {
	fmt.Printf("a %d\n", Answer)
	fmt.Printf("b %d\n", Answer+1)
	fmt.Printf("c %d\n", 44+Zero)
}

// expect: a 42
// expect: b 43
// expect: c 44

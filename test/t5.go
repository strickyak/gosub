package main

func compute() int {
	a := 10
	b := 100
	c := a + b
	return 2 * c
}

func main() {
	println(compute())
}

// expect: 220

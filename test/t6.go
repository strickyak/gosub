package main

func compute(x int) int {
	a := 15
	b := 100
	c := a + b
	return 2*c + x
}

func main() {
	println(compute(1000))
}

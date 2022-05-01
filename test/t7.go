package main

func triple() (x int, y int, z int) {
	return 111, 222, 333
}

func compute() int {
	a, b, c := triple()
	return 100*a + 10*b + c
}

func main() {
	println(compute())
}

// expect: 13653

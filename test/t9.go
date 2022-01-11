package main

func main() {
	i := 1
	var sum int
	for i <= 100 {
		sum = sum + i
		i = i + 1
	}
	println("Total:", sum)
}

// expect: Total: 5050

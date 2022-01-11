package main

func main() {
	i := 1
	var sum int
	for {
		sum = sum + i
		i = i + 1
		if i > 100 {
			break
		}
	}
	println("Total:", sum)
}

// expect: Total: 5050

package main

func main() {
	sum := 0
	for i := 1; i <= 100; i++ {
		sum = sum + i
	}
	println(sum)

	sum = 0
	for i := 1; i <= 999; i++ {
		if i%2 == 1 {
			continue
		}
		if i > 100 {
			break
		}
		sum = sum + i
	}
	println(sum)
}

// expect: 5050
// expect: 2550

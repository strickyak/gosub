package main

func main() {
	i := 0
	for {
		if i > 20 {
			break
		}
		i = i + 1
		if i%15 == 0 {
			println("fizzbuzz")
			continue
		} else if i%5 == 0 {
			println("buzz")
			continue
		} else if i%3 == 0 {
			println("fizz")
			continue
		} else {
			println(i)
			continue
		}
	}
}

// expect: 1
// expect: 2
// expect: fizz
// expect: 4
// expect: buzz
// expect: fizz
// expect: 7
// expect: 8
// expect: fizz
// expect: buzz
// expect: 11
// expect: fizz
// expect: 13
// expect: 14
// expect: fizzbuzz
// expect: 16
// expect: 17
// expect: fizz
// expect: 19
// expect: buzz
// expect: fizz

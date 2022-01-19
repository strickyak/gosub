package main

func divides(p int, q int) bool {
	return p%q == 0
}
func prime(p int) bool {
	i := 1
	divisions := 0
	for i <= p {
		if divides(p, i) {
			divisions++
		}
		i++
	}
	return divisions == 2
}

func main() {
	i := 2
	for i <= 100 {
		if prime(i) {
			println(i, "prime")
		} else {
			println(i, "composite")
		}
		i = i + 1
	}
}

// expect: 2 prime
// expect: 3 prime
// expect: 4 composite
// expect: 5 prime
// expect: 6 composite
// expect: 7 prime
// expect: 8 composite
// expect: 9 composite
// expect: 10 composite
// expect: 11 prime
// expect: 12 composite
// expect: 13 prime
// expect: 14 composite
// expect: 15 composite
// expect: 16 composite
// expect: 17 prime
// expect: 18 composite
// expect: 19 prime
// expect: 20 composite
// expect: 21 composite
// expect: 22 composite
// expect: 23 prime
// expect: 24 composite
// expect: 25 composite
// expect: 26 composite
// expect: 27 composite
// expect: 28 composite
// expect: 29 prime
// expect: 30 composite
// expect: 31 prime
// expect: 32 composite
// expect: 33 composite
// expect: 34 composite
// expect: 35 composite
// expect: 36 composite
// expect: 37 prime
// expect: 38 composite
// expect: 39 composite
// expect: 40 composite
// expect: 41 prime
// expect: 42 composite
// expect: 43 prime
// expect: 44 composite
// expect: 45 composite
// expect: 46 composite
// expect: 47 prime
// expect: 48 composite
// expect: 49 composite
// expect: 50 composite
// expect: 51 composite
// expect: 52 composite
// expect: 53 prime
// expect: 54 composite
// expect: 55 composite
// expect: 56 composite
// expect: 57 composite
// expect: 58 composite
// expect: 59 prime
// expect: 60 composite
// expect: 61 prime
// expect: 62 composite
// expect: 63 composite
// expect: 64 composite
// expect: 65 composite
// expect: 66 composite
// expect: 67 prime
// expect: 68 composite
// expect: 69 composite
// expect: 70 composite
// expect: 71 prime
// expect: 72 composite
// expect: 73 prime
// expect: 74 composite
// expect: 75 composite
// expect: 76 composite
// expect: 77 composite
// expect: 78 composite
// expect: 79 prime
// expect: 80 composite
// expect: 81 composite
// expect: 82 composite
// expect: 83 prime
// expect: 84 composite
// expect: 85 composite
// expect: 86 composite
// expect: 87 composite
// expect: 88 composite
// expect: 89 prime
// expect: 90 composite
// expect: 91 composite
// expect: 92 composite
// expect: 93 composite
// expect: 94 composite
// expect: 95 composite
// expect: 96 composite
// expect: 97 prime
// expect: 98 composite
// expect: 99 composite
// expect: 100 composite

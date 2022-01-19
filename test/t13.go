package main

func main() {
	s := "hello"
	n := len(s)
	println("len", n)
	println(s[0], s[1], s[2], s[3], s[4])
	println(s + " " + "world!")

	bb := make([]byte, n)
	bb[0] = 'H'
	println(bb[0])

	for j := range bb {
		bb[j] = s[j] + 1
	}
	for i, e := range bb {
		println(i, e)
	}

	// Test truncation to byte, in the slice.
	for j := range bb {
		bb[j] = s[j] * 100
	}
	for i, e := range bb {
		println(i, e)
	}

	// ii := make([]int, n)
	// for j := range ii {
	// ii[j] = s[j] * 100
	// }
	// for i, e := range ii {
	// println(i, e)
	// }
}

// expect: len 5
// expect: 104 101 108 108 111
// expect: hello world!
// expect: 72
// expect: 0 105
// expect: 1 102
// expect: 2 109
// expect: 3 109
// expect: 4 112
// expect: 0 160
// expect: 1 116
// expect: 2 48
// expect: 3 48
// expect: 4 92

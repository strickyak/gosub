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
}

// expect: len 5
// expect: 104 101 108 108 111
// expect: hello world!
// expect: 72
// 0 105
// 1 102
// 2 109
// 3 109
// 4 112

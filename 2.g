package main

func Triangle(n int) int {
    sum := 0
    for n>0 {
        sum = sum + n
        n = n - 1
    }
    return sum
}

func Hyp(a int, b int) int {
    return a*a + b*b
}

func main() {
    a := 3
    b := 4
    c := Hyp(a,b)
    println(c)
    println(Triangle(10))
}

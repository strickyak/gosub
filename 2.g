package main

func Hyp(a int, b int) int {
    return a*a + b*b
}

func main() {
    a := 3
    b := 4
    c := Hyp(a,b)
    println(c)
}

package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		t := sc.Text()
		h := Hash(t)
		fmt.Printf("%d\t%d\t%d\t%q\n", h, h%63, h%15, t)
	}
}

func Hash(s string) byte {
	var h byte
	for i := 0; i < len(s); i++ {
		h += s[i]
	}
	return h
}

package main

import "fmt"

type Stringer interface {
	String() string
}
type Frobber interface {
	Frob(x int) Stringer
}

type Apple struct {
	x int
	y int
}

type Grape struct {
	a int
	b int
	c int
}

func (p *Apple) String() string {
	return fmt.Sprintf("Apple(%d,%d)", p.x, p.y)
}
func (p *Grape) String() string {
	return fmt.Sprintf("Grape(%d,%d,%d)", p.a, p.b, p.c)
}

func (p *Apple) Frob(x int) Stringer {
	p.x = p.x + x
	p.y = p.y + x
	return p
}

func (p *Grape) Frob(x int) Stringer {
	p.a = p.a + x
	p.b = p.b + x
	p.c = p.c + x
	return p
}

func main() {
	apple := &Apple{}
	grape := &Grape{}
	println(apple.Frob(1).String())
	println(grape.Frob(10).String())
	var face Frobber
	face = apple
	println(face.Frob(100).String())
	face = grape
	println(face.Frob(100).String())
}

// expect: 2
// expect: 30
// expect: 202

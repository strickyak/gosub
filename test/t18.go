package main

import "fmt"

// TODO: swap these two interface types, so Stringer result is a forward reference.
// This is a current limitation that needs to be fixed.
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
	println(apple.String())
	println(grape.String())
	println(apple.Frob(1).String())
	println(grape.Frob(10).String())
	var face Frobber
	face = apple
	println(face.Frob(100).String())
	face = grape
	println(face.Frob(100).String())
}

// expect: Apple(0,0)
// expect: Grape(0,0,0)
// expect: Apple(1,1)
// expect: Grape(10,10,10)
// expect: Apple(101,101)
// expect: Grape(110,110,110)

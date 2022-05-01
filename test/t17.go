package main

type Frobber interface {
	Frob(x int) int
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

func (p *Apple) Frob(x int) int {
	p.x = p.x + x
	p.y = p.y + x
	return p.x + p.y
}

func (p *Grape) Frob(x int) int {
	p.a = p.a + x
	p.b = p.b + x
	p.c = p.c + x
	return p.a + p.b + p.c
}

func main() {
	apple := &Apple{}
	grape := &Grape{}
	println(apple.Frob(1))
	println(grape.Frob(10))
	var face Frobber
	face = apple
	println(face.Frob(100))
}

// expect: 2
// expect: 30
// expect: 202

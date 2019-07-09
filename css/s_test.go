package css

import (
	"fmt"
	"testing"
)

type A struct {
	i int
}

type B interface {
	Self() *A
}

func (a *A) Self() *A {
	return a
}

func TestI(t *testing.T) {
	var a A
	b := B(&a)
	p1 := b.Self()
	p2 := &a
	fmt.Printf("%p %p", p1, p2)
}

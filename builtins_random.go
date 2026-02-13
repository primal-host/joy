package main

import "math/rand"

var joyRand = rand.New(rand.NewSource(0))

func init() {
	// srand: I -> — seed the random number generator
	register("srand", func(m *Machine) {
		m.NeedStack(1, "srand")
		a := m.Pop()
		joyRand = rand.New(rand.NewSource(a.Int))
	})

	// rand: -> I — push a random non-negative integer
	register("rand", func(m *Machine) {
		m.Push(IntVal(joyRand.Int63()))
	})
}

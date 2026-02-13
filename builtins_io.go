package main

import "fmt"

func init() {
	// put: print top of stack followed by space
	register("put", func(m *Machine) {
		m.NeedStack(1, "put")
		a := m.Pop()
		fmt.Print(a.String())
	})

	// putch: print character value (no quotes)
	register("putch", func(m *Machine) {
		m.NeedStack(1, "putch")
		a := m.Pop()
		if a.Typ == TypeChar || a.Typ == TypeInteger {
			fmt.Print(string(rune(a.Int)))
		} else {
			fmt.Print(a.String())
		}
	})

	// putchars: print string without quotes
	register("putchars", func(m *Machine) {
		m.NeedStack(1, "putchars")
		a := m.Pop()
		if a.Typ == TypeString {
			fmt.Print(a.Str)
		} else {
			fmt.Print(a.String())
		}
	})

	// . (autoput): pop and print top value with newline
	register(".", func(m *Machine) {
		m.NeedStack(1, ".")
		a := m.Pop()
		fmt.Println(a.String())
	})

	// .s : print stack without consuming
	register(".s", func(m *Machine) {
		fmt.Println(m.PrintStack())
	})

	// newline
	register("newline", func(m *Machine) {
		fmt.Println()
	})

	// get: -> X — read a line from stdin, parse as Joy, push result
	register("get", func(m *Machine) {
		if m.Input == nil {
			joyErr("get: no input source available")
		}
		if !m.Input.Scan() {
			joyErr("get: end of input")
		}
		line := m.Input.Text()
		tokens := NewScanner(line).ScanAll()
		program := NewParser(tokens, m).Parse()
		// Push each parsed value onto the stack
		for _, v := range program {
			m.Push(v)
		}
	})
}

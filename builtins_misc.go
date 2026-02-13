package main

import (
	"fmt"
	"math"
	"os"
	"time"
)

func init() {
	register("true", func(m *Machine) {
		m.Push(BoolVal(true))
	})

	register("false", func(m *Machine) {
		m.Push(BoolVal(false))
	})

	register("maxint", func(m *Machine) {
		m.Push(IntVal(math.MaxInt64))
	})

	register("setsize", func(m *Machine) {
		m.Push(IntVal(SetSize))
	})

	register("clock", func(m *Machine) {
		// microseconds since program start
		m.Push(IntVal(time.Now().UnixMicro()))
	})

	register("time", func(m *Machine) {
		m.Push(IntVal(time.Now().Unix()))
	})

	register("argc", func(m *Machine) {
		m.Push(IntVal(int64(len(os.Args))))
	})

	register("argv", func(m *Machine) {
		var args []Value
		for _, a := range os.Args {
			args = append(args, StringVal(a))
		}
		m.Push(ListVal(args))
	})

	register("quit", func(m *Machine) {
		os.Exit(0)
	})

	register("abort", func(m *Machine) {
		joyErr("abort")
	})

	register("typeof", func(m *Machine) {
		m.NeedStack(1, "typeof")
		a := m.Pop()
		m.Push(IntVal(int64(a.Typ)))
	})

	register("sametype", func(m *Machine) {
		m.NeedStack(2, "sametype")
		b := m.Pop()
		a := m.Pop()
		m.Push(BoolVal(a.Typ == b.Typ))
	})

	register("equal", func(m *Machine) {
		m.NeedStack(2, "equal")
		b := m.Pop()
		a := m.Pop()
		m.Push(BoolVal(a.Equal(b)))
	})

	register("uncons2", func(m *Machine) {
		m.NeedStack(1, "uncons2")
		builtins["uncons"](m)
		n := len(m.Stack)
		m.Stack[n-1], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-1]
		builtins["uncons"](m)
		n = len(m.Stack)
		m.Stack[n-1], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-1]
	})

	register("uncons3", func(m *Machine) {
		m.NeedStack(1, "uncons3")
		builtins["uncons2"](m)
		n := len(m.Stack)
		m.Stack[n-1], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-1]
		builtins["uncons"](m)
		n = len(m.Stack)
		m.Stack[n-1], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-1]
	})

	register("opcase", func(m *Machine) {
		m.NeedStack(2, "opcase")
		cases := m.Pop()
		x := m.Peek()
		if cases.Typ != TypeList {
			joyErr("opcase: list expected")
		}
		for _, c := range cases.List {
			if c.Typ == TypeList && len(c.List) > 0 {
				if c.List[0].Equal(x) {
					m.Push(ListVal(c.List[1:]))
					return
				}
			}
		}
		// default: last element
		if len(cases.List) > 0 {
			last := cases.List[len(cases.List)-1]
			if last.Typ == TypeList {
				m.Push(last)
			} else {
				m.Push(ListVal([]Value{last}))
			}
		}
	})

	register("case", func(m *Machine) {
		m.NeedStack(2, "case")
		cases := m.Pop()
		x := m.Pop()
		if cases.Typ != TypeList {
			joyErr("case: list expected")
		}
		for _, c := range cases.List {
			if c.Typ == TypeList && len(c.List) >= 2 {
				if c.List[0].Equal(x) {
					m.Execute(c.List[1:])
					return
				}
			}
		}
		// default: last element
		if len(cases.List) > 0 {
			last := cases.List[len(cases.List)-1]
			if last.Typ == TypeList {
				m.Execute(last.List)
			}
		}
	})

	// REPL control
	register("setautoput", func(m *Machine) {
		m.NeedStack(1, "setautoput")
		a := m.Pop()
		m.Autoput = int(a.Int)
	})

	register("setecho", func(m *Machine) {
		m.NeedStack(1, "setecho")
		a := m.Pop()
		m.Echo = int(a.Int)
	})

	// GC trace toggle — no-op in this implementation, pops one arg
	register("__settracegc", func(m *Machine) {
		m.NeedStack(1, "__settracegc")
		m.Pop()
	})

	register("setundeferror", func(m *Machine) {
		m.NeedStack(1, "setundeferror")
		a := m.Pop()
		m.UndefError = int(a.Int)
	})

	register("help", func(m *Machine) {
		fmt.Println("Joy interpreter — built-in operators:")
		count := 0
		for name := range builtins {
			fmt.Printf("  %-16s", name)
			count++
			if count%5 == 0 {
				fmt.Println()
			}
		}
		if count%5 != 0 {
			fmt.Println()
		}
		fmt.Printf("\nTotal: %d built-in operators\n", count)
	})

	register("helpdetail", func(m *Machine) {
		m.NeedStack(1, "helpdetail")
		a := m.Pop()
		if a.Typ == TypeList {
			for _, item := range a.List {
				if _, ok := builtins[item.Str]; ok {
					fmt.Printf("%s : built-in\n", item.Str)
				} else if _, ok := m.Dict[item.Str]; ok {
					fmt.Printf("%s : user-defined\n", item.Str)
				} else {
					fmt.Printf("%s : unknown\n", item.Str)
				}
			}
		}
	})
}

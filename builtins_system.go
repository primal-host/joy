package main

import "os"

func init() {
	// getenv: S -> S2 — get environment variable
	register("getenv", func(m *Machine) {
		m.NeedStack(1, "getenv")
		a := m.Pop()
		if a.Typ != TypeString {
			joyErr("getenv: string expected")
		}
		m.Push(StringVal(os.Getenv(a.Str)))
	})

	// undefs: -> L — list undefined references in user definitions
	register("undefs", func(m *Machine) {
		seen := map[string]bool{}
		var undefs []Value
		for _, body := range m.Dict {
			for _, v := range body {
				if v.Typ == TypeUserDef {
					if _, inDict := m.Dict[v.Str]; !inDict {
						if _, inBuiltins := builtins[v.Str]; !inBuiltins {
							if !seen[v.Str] {
								seen[v.Str] = true
								undefs = append(undefs, StringVal(v.Str))
							}
						}
					}
				}
			}
		}
		if undefs == nil {
			undefs = []Value{}
		}
		m.Push(ListVal(undefs))
	})
}

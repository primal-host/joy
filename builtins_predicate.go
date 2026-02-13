package main

func init() {
	register("integer", func(m *Machine) {
		m.NeedStack(1, "integer")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeInteger))
	})

	register("char", func(m *Machine) {
		m.NeedStack(1, "char")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeChar))
	})

	register("logical", func(m *Machine) {
		m.NeedStack(1, "logical")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeBoolean))
	})

	register("float", func(m *Machine) {
		m.NeedStack(1, "float")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeFloat))
	})

	register("string", func(m *Machine) {
		m.NeedStack(1, "string")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeString))
	})

	register("list", func(m *Machine) {
		m.NeedStack(1, "list")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeList))
	})

	register("set", func(m *Machine) {
		m.NeedStack(1, "set")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeSet))
	})

	register("leaf", func(m *Machine) {
		m.NeedStack(1, "leaf")
		a := m.Pop()
		m.Push(BoolVal(a.Typ != TypeList))
	})

	register("user", func(m *Machine) {
		m.NeedStack(1, "user")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeUserDef))
	})

	register("file", func(m *Machine) {
		m.NeedStack(1, "file")
		a := m.Pop()
		m.Push(BoolVal(a.Typ == TypeFile))
	})
}

package main

func init() {
	register("and", func(m *Machine) {
		m.NeedStack(2, "and")
		b := m.Pop()
		a := m.Pop()
		if a.Typ == TypeSet && b.Typ == TypeSet {
			m.Push(SetVal(a.Int & b.Int))
		} else {
			m.Push(BoolVal(a.IsTruthy() && b.IsTruthy()))
		}
	})

	register("or", func(m *Machine) {
		m.NeedStack(2, "or")
		b := m.Pop()
		a := m.Pop()
		if a.Typ == TypeSet && b.Typ == TypeSet {
			m.Push(SetVal(a.Int | b.Int))
		} else {
			m.Push(BoolVal(a.IsTruthy() || b.IsTruthy()))
		}
	})

	register("xor", func(m *Machine) {
		m.NeedStack(2, "xor")
		b := m.Pop()
		a := m.Pop()
		if a.Typ == TypeSet && b.Typ == TypeSet {
			m.Push(SetVal(a.Int ^ b.Int))
		} else {
			m.Push(BoolVal(a.IsTruthy() != b.IsTruthy()))
		}
	})

	register("not", func(m *Machine) {
		m.NeedStack(1, "not")
		a := m.Pop()
		if a.Typ == TypeSet {
			m.Push(SetVal(^a.Int & ((1 << SetSize) - 1)))
		} else {
			m.Push(BoolVal(!a.IsTruthy()))
		}
	})
}

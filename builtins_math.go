package main

import "math"

func init() {
	// Arithmetic
	register("+", func(m *Machine) {
		m.NeedStack(2, "+")
		b := m.Pop()
		a := m.Pop()
		if a.Typ == TypeFloat || b.Typ == TypeFloat {
			m.Push(FloatVal(a.NumericVal() + b.NumericVal()))
		} else {
			m.Push(IntVal(a.Int + b.Int))
		}
	})

	register("-", func(m *Machine) {
		m.NeedStack(2, "-")
		b := m.Pop()
		a := m.Pop()
		if a.Typ == TypeFloat || b.Typ == TypeFloat {
			m.Push(FloatVal(a.NumericVal() - b.NumericVal()))
		} else {
			m.Push(IntVal(a.Int - b.Int))
		}
	})

	register("*", func(m *Machine) {
		m.NeedStack(2, "*")
		b := m.Pop()
		a := m.Pop()
		if a.Typ == TypeFloat || b.Typ == TypeFloat {
			m.Push(FloatVal(a.NumericVal() * b.NumericVal()))
		} else {
			m.Push(IntVal(a.Int * b.Int))
		}
	})

	register("/", func(m *Machine) {
		m.NeedStack(2, "/")
		b := m.Pop()
		a := m.Pop()
		if a.Typ == TypeFloat || b.Typ == TypeFloat {
			bv := b.NumericVal()
			if bv == 0 {
				joyErr("/: division by zero")
			}
			m.Push(FloatVal(a.NumericVal() / bv))
		} else {
			if b.Int == 0 {
				joyErr("/: division by zero")
			}
			m.Push(IntVal(a.Int / b.Int))
		}
	})

	register("rem", func(m *Machine) {
		m.NeedStack(2, "rem")
		b := m.Pop()
		a := m.Pop()
		if b.Int == 0 {
			joyErr("rem: division by zero")
		}
		m.Push(IntVal(a.Int % b.Int))
	})

	register("div", func(m *Machine) {
		m.NeedStack(2, "div")
		b := m.Pop()
		a := m.Pop()
		if b.Int == 0 {
			joyErr("div: division by zero")
		}
		m.Push(IntVal(a.Int / b.Int))
	})

	register("succ", func(m *Machine) {
		m.NeedStack(1, "succ")
		a := m.Pop()
		m.Push(IntVal(a.Int + 1))
	})

	register("pred", func(m *Machine) {
		m.NeedStack(1, "pred")
		a := m.Pop()
		m.Push(IntVal(a.Int - 1))
	})

	register("neg", func(m *Machine) {
		m.NeedStack(1, "neg")
		a := m.Pop()
		if a.Typ == TypeFloat {
			m.Push(FloatVal(-a.Flt))
		} else {
			m.Push(IntVal(-a.Int))
		}
	})

	register("abs", func(m *Machine) {
		m.NeedStack(1, "abs")
		a := m.Pop()
		if a.Typ == TypeFloat {
			m.Push(FloatVal(math.Abs(a.Flt)))
		} else {
			if a.Int < 0 {
				m.Push(IntVal(-a.Int))
			} else {
				m.Push(a)
			}
		}
	})

	register("sign", func(m *Machine) {
		m.NeedStack(1, "sign")
		a := m.Pop()
		v := a.NumericVal()
		if v < 0 {
			m.Push(IntVal(-1))
		} else if v > 0 {
			m.Push(IntVal(1))
		} else {
			m.Push(IntVal(0))
		}
	})

	register("max", func(m *Machine) {
		m.NeedStack(2, "max")
		b := m.Pop()
		a := m.Pop()
		if a.Compare(b) >= 0 {
			m.Push(a)
		} else {
			m.Push(b)
		}
	})

	register("min", func(m *Machine) {
		m.NeedStack(2, "min")
		b := m.Pop()
		a := m.Pop()
		if a.Compare(b) <= 0 {
			m.Push(a)
		} else {
			m.Push(b)
		}
	})

	register("ord", func(m *Machine) {
		m.NeedStack(1, "ord")
		a := m.Pop()
		switch a.Typ {
		case TypeChar:
			m.Push(IntVal(a.Int))
		case TypeInteger:
			m.Push(a)
		default:
			joyErr("ord: char or integer expected")
		}
	})

	register("chr", func(m *Machine) {
		m.NeedStack(1, "chr")
		a := m.Pop()
		m.Push(CharVal(a.Int))
	})

	// Comparisons
	register("<", func(m *Machine) {
		m.NeedStack(2, "<")
		b := m.Pop()
		a := m.Pop()
		m.Push(BoolVal(a.Compare(b) < 0))
	})

	register("<=", func(m *Machine) {
		m.NeedStack(2, "<=")
		b := m.Pop()
		a := m.Pop()
		m.Push(BoolVal(a.Compare(b) <= 0))
	})

	register(">", func(m *Machine) {
		m.NeedStack(2, ">")
		b := m.Pop()
		a := m.Pop()
		m.Push(BoolVal(a.Compare(b) > 0))
	})

	register(">=", func(m *Machine) {
		m.NeedStack(2, ">=")
		b := m.Pop()
		a := m.Pop()
		m.Push(BoolVal(a.Compare(b) >= 0))
	})

	register("=", func(m *Machine) {
		m.NeedStack(2, "=")
		b := m.Pop()
		a := m.Pop()
		m.Push(BoolVal(a.Equal(b)))
	})

	register("!=", func(m *Machine) {
		m.NeedStack(2, "!=")
		b := m.Pop()
		a := m.Pop()
		m.Push(BoolVal(!a.Equal(b)))
	})

	register("compare", func(m *Machine) {
		m.NeedStack(2, "compare")
		b := m.Pop()
		a := m.Pop()
		m.Push(IntVal(int64(a.Compare(b))))
	})

	// Float math
	register("sqrt", func(m *Machine) {
		m.NeedStack(1, "sqrt")
		a := m.Pop()
		m.Push(FloatVal(math.Sqrt(a.NumericVal())))
	})

	register("floor", func(m *Machine) {
		m.NeedStack(1, "floor")
		a := m.Pop()
		m.Push(IntVal(int64(math.Floor(a.NumericVal()))))
	})

	register("ceil", func(m *Machine) {
		m.NeedStack(1, "ceil")
		a := m.Pop()
		m.Push(IntVal(int64(math.Ceil(a.NumericVal()))))
	})

	register("trunc", func(m *Machine) {
		m.NeedStack(1, "trunc")
		a := m.Pop()
		m.Push(IntVal(int64(math.Trunc(a.NumericVal()))))
	})
}

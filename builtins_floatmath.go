package main

import "math"

func init() {
	// sin: F -> G — sine
	register("sin", func(m *Machine) {
		m.NeedStack(1, "sin")
		a := m.Pop()
		m.Push(FloatVal(math.Sin(a.NumericVal())))
	})

	// cos: F -> G — cosine
	register("cos", func(m *Machine) {
		m.NeedStack(1, "cos")
		a := m.Pop()
		m.Push(FloatVal(math.Cos(a.NumericVal())))
	})

	// tan: F -> G — tangent
	register("tan", func(m *Machine) {
		m.NeedStack(1, "tan")
		a := m.Pop()
		m.Push(FloatVal(math.Tan(a.NumericVal())))
	})

	// asin: F -> G — arc sine
	register("asin", func(m *Machine) {
		m.NeedStack(1, "asin")
		a := m.Pop()
		m.Push(FloatVal(math.Asin(a.NumericVal())))
	})

	// acos: F -> G — arc cosine
	register("acos", func(m *Machine) {
		m.NeedStack(1, "acos")
		a := m.Pop()
		m.Push(FloatVal(math.Acos(a.NumericVal())))
	})

	// atan: F -> G — arc tangent
	register("atan", func(m *Machine) {
		m.NeedStack(1, "atan")
		a := m.Pop()
		m.Push(FloatVal(math.Atan(a.NumericVal())))
	})

	// atan2: F G -> H — two-argument arc tangent
	register("atan2", func(m *Machine) {
		m.NeedStack(2, "atan2")
		g := m.Pop()
		f := m.Pop()
		m.Push(FloatVal(math.Atan2(f.NumericVal(), g.NumericVal())))
	})

	// log: F -> G — natural logarithm
	register("log", func(m *Machine) {
		m.NeedStack(1, "log")
		a := m.Pop()
		m.Push(FloatVal(math.Log(a.NumericVal())))
	})

	// log10: F -> G — base-10 logarithm
	register("log10", func(m *Machine) {
		m.NeedStack(1, "log10")
		a := m.Pop()
		m.Push(FloatVal(math.Log10(a.NumericVal())))
	})

	// exp: F -> G — e^F
	register("exp", func(m *Machine) {
		m.NeedStack(1, "exp")
		a := m.Pop()
		m.Push(FloatVal(math.Exp(a.NumericVal())))
	})

	// pow: F G -> H — F raised to the power G
	register("pow", func(m *Machine) {
		m.NeedStack(2, "pow")
		g := m.Pop()
		f := m.Pop()
		m.Push(FloatVal(math.Pow(f.NumericVal(), g.NumericVal())))
	})

	// ldexp: F I -> G — F * 2^I
	register("ldexp", func(m *Machine) {
		m.NeedStack(2, "ldexp")
		i := m.Pop()
		f := m.Pop()
		m.Push(FloatVal(math.Ldexp(f.NumericVal(), int(i.Int))))
	})

	// frexp: F -> G I — split F into fraction G and exponent I
	register("frexp", func(m *Machine) {
		m.NeedStack(1, "frexp")
		a := m.Pop()
		frac, exp := math.Frexp(a.NumericVal())
		m.Push(FloatVal(frac))
		m.Push(IntVal(int64(exp)))
	})

	// modf: F -> G H — split F into integer H and fractional G parts
	register("modf", func(m *Machine) {
		m.NeedStack(1, "modf")
		a := m.Pop()
		integer, frac := math.Modf(a.NumericVal())
		m.Push(FloatVal(frac))
		m.Push(FloatVal(integer))
	})
}

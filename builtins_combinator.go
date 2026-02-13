package main

func init() {
	// i: [P] -> ... — execute quotation
	register("i", func(m *Machine) {
		m.NeedStack(1, "i")
		q := m.Pop()
		if q.Typ != TypeList {
			joyErr("i: quotation expected")
		}
		m.Execute(q.List)
	})

	// x: [P] -> [P] ... — execute quotation without removing it
	register("x", func(m *Machine) {
		m.NeedStack(1, "x")
		q := m.Peek()
		if q.Typ != TypeList {
			joyErr("x: quotation expected")
		}
		m.Execute(q.List)
	})

	// dip: X [P] -> ... X — execute P under X
	register("dip", func(m *Machine) {
		m.NeedStack(2, "dip")
		q := m.Pop()
		x := m.Pop()
		if q.Typ != TypeList {
			joyErr("dip: quotation expected")
		}
		m.Execute(q.List)
		m.Push(x)
	})

	// dipd: Y X [P] -> ... Y X — execute P under two values
	register("dipd", func(m *Machine) {
		m.NeedStack(3, "dipd")
		q := m.Pop()
		x := m.Pop()
		y := m.Pop()
		if q.Typ != TypeList {
			joyErr("dipd: quotation expected")
		}
		m.Execute(q.List)
		m.Push(y)
		m.Push(x)
	})

	// dipdd: Z Y X [P] -> ... Z Y X
	register("dipdd", func(m *Machine) {
		m.NeedStack(4, "dipdd")
		q := m.Pop()
		x := m.Pop()
		y := m.Pop()
		z := m.Pop()
		if q.Typ != TypeList {
			joyErr("dipdd: quotation expected")
		}
		m.Execute(q.List)
		m.Push(z)
		m.Push(y)
		m.Push(x)
	})

	// app1: X [P] -> R — apply P to X
	register("app1", func(m *Machine) {
		m.NeedStack(2, "app1")
		q := m.Pop()
		if q.Typ != TypeList {
			joyErr("app1: quotation expected")
		}
		m.Execute(q.List)
	})

	// app2: X Y [P] -> Rx Ry
	register("app2", func(m *Machine) {
		m.NeedStack(3, "app2")
		q := m.Pop()
		y := m.Pop()
		x := m.Pop()
		if q.Typ != TypeList {
			joyErr("app2: quotation expected")
		}
		m.Push(x)
		m.Execute(q.List)
		rx := m.Pop()
		m.Push(y)
		m.Execute(q.List)
		ry := m.Pop()
		m.Push(rx)
		m.Push(ry)
	})

	// app3: X Y Z [P] -> Rx Ry Rz
	register("app3", func(m *Machine) {
		m.NeedStack(4, "app3")
		q := m.Pop()
		z := m.Pop()
		y := m.Pop()
		x := m.Pop()
		if q.Typ != TypeList {
			joyErr("app3: quotation expected")
		}
		m.Push(x)
		m.Execute(q.List)
		rx := m.Pop()
		m.Push(y)
		m.Execute(q.List)
		ry := m.Pop()
		m.Push(z)
		m.Execute(q.List)
		rz := m.Pop()
		m.Push(rx)
		m.Push(ry)
		m.Push(rz)
	})

	// branch: B [T] [F] -> ... — if B then T else F
	register("branch", func(m *Machine) {
		m.NeedStack(3, "branch")
		fBranch := m.Pop()
		tBranch := m.Pop()
		cond := m.Pop()
		if tBranch.Typ != TypeList || fBranch.Typ != TypeList {
			joyErr("branch: two quotations expected")
		}
		if cond.IsTruthy() {
			m.Execute(tBranch.List)
		} else {
			m.Execute(fBranch.List)
		}
	})

	// ifte: [B] [T] [F] -> ... — if-then-else preserving stack
	// Also supports: B [T] [F] -> ... (non-quotation condition)
	register("ifte", func(m *Machine) {
		m.NeedStack(3, "ifte")
		fBranch := m.Pop()
		tBranch := m.Pop()
		test := m.Pop()
		if fBranch.Typ != TypeList || tBranch.Typ != TypeList {
			joyErr("ifte: two quotation branches expected")
		}
		if test.Typ == TypeList {
			// save stack, run test, restore stack, then branch
			savedStack := make([]Value, len(m.Stack))
			copy(savedStack, m.Stack)
			m.Execute(test.List)
			cond := m.Pop()
			m.Stack = savedStack
			if cond.IsTruthy() {
				m.Execute(tBranch.List)
			} else {
				m.Execute(fBranch.List)
			}
		} else {
			// non-quotation condition: use directly
			if test.IsTruthy() {
				m.Execute(tBranch.List)
			} else {
				m.Execute(fBranch.List)
			}
		}
	})

	// cond: [[T1] [B1] [T2] [B2] ... [Def]] -> ...
	register("cond", func(m *Machine) {
		m.NeedStack(1, "cond")
		clauses := m.Pop()
		if clauses.Typ != TypeList {
			joyErr("cond: list of clauses expected")
		}
		for _, clause := range clauses.List {
			if clause.Typ != TypeList || len(clause.List) == 0 {
				joyErr("cond: each clause must be a non-empty list")
			}
			if len(clause.List) == 1 {
				// default clause
				m.Execute(clause.List)
				return
			}
			// [Test Body]
			test := clause.List[0]
			body := clause.List[1:]
			if test.Typ == TypeList {
				savedStack := make([]Value, len(m.Stack))
				copy(savedStack, m.Stack)
				m.Execute(test.List)
				cond := m.Pop()
				m.Stack = savedStack
				if cond.IsTruthy() {
					m.Execute(body)
					return
				}
			} else if test.IsTruthy() {
				m.Execute(body)
				return
			}
		}
	})

	// times: N [P] -> ... — execute P, N times
	register("times", func(m *Machine) {
		m.NeedStack(2, "times")
		q := m.Pop()
		n := m.Pop()
		if q.Typ != TypeList {
			joyErr("times: quotation expected")
		}
		count := n.Int
		for i := int64(0); i < count; i++ {
			m.Execute(q.List)
		}
	})

	// step: A [P] -> ... — execute P for each element of aggregate
	register("step", func(m *Machine) {
		m.NeedStack(2, "step")
		q := m.Pop()
		agg := m.Pop()
		if q.Typ != TypeList {
			joyErr("step: quotation expected")
		}
		switch agg.Typ {
		case TypeList:
			for _, item := range agg.List {
				m.Push(item)
				m.Execute(q.List)
			}
		case TypeString:
			for _, ch := range agg.Str {
				m.Push(CharVal(int64(ch)))
				m.Execute(q.List)
			}
		case TypeSet:
			for i := 0; i < SetSize; i++ {
				if agg.Int&(1<<i) != 0 {
					m.Push(IntVal(int64(i)))
					m.Execute(q.List)
				}
			}
		default:
			joyErr("step: aggregate expected")
		}
	})

	// map: A [P] -> B — apply P to each element
	register("map", func(m *Machine) {
		m.NeedStack(2, "map")
		q := m.Pop()
		agg := m.Pop()
		if q.Typ != TypeList {
			joyErr("map: quotation expected")
		}
		switch agg.Typ {
		case TypeList:
			result := make([]Value, 0, len(agg.List))
			for _, item := range agg.List {
				m.Push(item)
				m.Execute(q.List)
				result = append(result, m.Pop())
			}
			m.Push(ListVal(result))
		case TypeString:
			var result []byte
			for _, ch := range agg.Str {
				m.Push(CharVal(int64(ch)))
				m.Execute(q.List)
				r := m.Pop()
				if r.Typ == TypeChar || r.Typ == TypeInteger {
					result = append(result, byte(r.Int))
				}
			}
			m.Push(StringVal(string(result)))
		case TypeSet:
			var bits int64
			for i := 0; i < SetSize; i++ {
				if agg.Int&(1<<i) != 0 {
					m.Push(IntVal(int64(i)))
					m.Execute(q.List)
					r := m.Pop()
					if r.Int >= 0 && r.Int < SetSize {
						bits |= 1 << r.Int
					}
				}
			}
			m.Push(SetVal(bits))
		default:
			joyErr("map: aggregate expected")
		}
	})

	// filter: A [P] -> B — keep elements where P is true
	register("filter", func(m *Machine) {
		m.NeedStack(2, "filter")
		q := m.Pop()
		agg := m.Pop()
		if q.Typ != TypeList {
			joyErr("filter: quotation expected")
		}
		switch agg.Typ {
		case TypeList:
			var result []Value
			for _, item := range agg.List {
				m.Push(item)
				m.Execute(q.List)
				cond := m.Pop()
				if cond.IsTruthy() {
					result = append(result, item)
				}
			}
			if result == nil {
				result = []Value{}
			}
			m.Push(ListVal(result))
		case TypeString:
			var result []byte
			for _, ch := range agg.Str {
				m.Push(CharVal(int64(ch)))
				m.Execute(q.List)
				cond := m.Pop()
				if cond.IsTruthy() {
					result = append(result, byte(ch))
				}
			}
			m.Push(StringVal(string(result)))
		case TypeSet:
			var bits int64
			for i := 0; i < SetSize; i++ {
				if agg.Int&(1<<i) != 0 {
					m.Push(IntVal(int64(i)))
					m.Execute(q.List)
					cond := m.Pop()
					if cond.IsTruthy() {
						bits |= 1 << i
					}
				}
			}
			m.Push(SetVal(bits))
		default:
			joyErr("filter: aggregate expected")
		}
	})

	// fold: V0 A [P] -> V — left fold
	register("fold", func(m *Machine) {
		m.NeedStack(3, "fold")
		q := m.Pop()
		agg := m.Pop()
		// V0 is already on stack
		if q.Typ != TypeList {
			joyErr("fold: quotation expected")
		}
		switch agg.Typ {
		case TypeList:
			for _, item := range agg.List {
				m.Push(item)
				m.Execute(q.List)
			}
		case TypeString:
			for _, ch := range agg.Str {
				m.Push(CharVal(int64(ch)))
				m.Execute(q.List)
			}
		case TypeSet:
			for i := 0; i < SetSize; i++ {
				if agg.Int&(1<<i) != 0 {
					m.Push(IntVal(int64(i)))
					m.Execute(q.List)
				}
			}
		default:
			joyErr("fold: aggregate expected")
		}
	})

	// construct: [P] [[Q1] [Q2] ...] -> ... — apply each Qi, collect results
	register("construct", func(m *Machine) {
		m.NeedStack(2, "construct")
		specs := m.Pop()
		test := m.Pop()
		if test.Typ != TypeList || specs.Typ != TypeList {
			joyErr("construct: two quotations expected")
		}
		// Apply test first
		savedStack := make([]Value, len(m.Stack))
		copy(savedStack, m.Stack)
		m.Execute(test.List)
		postStack := make([]Value, len(m.Stack))
		copy(postStack, m.Stack)
		// Apply each spec
		results := make([]Value, len(specs.List))
		for i, spec := range specs.List {
			m.Stack = make([]Value, len(postStack))
			copy(m.Stack, postStack)
			if spec.Typ == TypeList {
				m.Execute(spec.List)
			}
			results[i] = m.Pop()
		}
		m.Stack = savedStack
		m.Push(ListVal(results))
	})

	// nullary: [P] -> R — execute P, push single result
	register("nullary", func(m *Machine) {
		m.NeedStack(1, "nullary")
		q := m.Pop()
		if q.Typ != TypeList {
			joyErr("nullary: quotation expected")
		}
		savedStack := make([]Value, len(m.Stack))
		copy(savedStack, m.Stack)
		m.Execute(q.List)
		result := m.Pop()
		m.Stack = savedStack
		m.Push(result)
	})

	// unary: X [P] -> R
	register("unary", func(m *Machine) {
		m.NeedStack(2, "unary")
		q := m.Pop()
		if q.Typ != TypeList {
			joyErr("unary: quotation expected")
		}
		savedStack := make([]Value, len(m.Stack))
		copy(savedStack, m.Stack)
		m.Execute(q.List)
		result := m.Pop()
		m.Stack = savedStack[:len(savedStack)-1] // remove the argument
		m.Push(result)
	})

	// binary: X Y [P] -> R
	register("binary", func(m *Machine) {
		m.NeedStack(3, "binary")
		q := m.Pop()
		if q.Typ != TypeList {
			joyErr("binary: quotation expected")
		}
		savedStack := make([]Value, len(m.Stack))
		copy(savedStack, m.Stack)
		m.Execute(q.List)
		result := m.Pop()
		m.Stack = savedStack[:len(savedStack)-2] // remove the two arguments
		m.Push(result)
	})

	// ternary: X Y Z [P] -> R
	register("ternary", func(m *Machine) {
		m.NeedStack(4, "ternary")
		q := m.Pop()
		if q.Typ != TypeList {
			joyErr("ternary: quotation expected")
		}
		savedStack := make([]Value, len(m.Stack))
		copy(savedStack, m.Stack)
		m.Execute(q.List)
		result := m.Pop()
		m.Stack = savedStack[:len(savedStack)-3]
		m.Push(result)
	})

	// cleave: X [P] [Q] -> ... — apply P and Q each to X
	register("cleave", func(m *Machine) {
		m.NeedStack(3, "cleave")
		q2 := m.Pop()
		q1 := m.Pop()
		if q1.Typ != TypeList || q2.Typ != TypeList {
			joyErr("cleave: two quotations expected")
		}
		savedStack := make([]Value, len(m.Stack))
		copy(savedStack, m.Stack)
		m.Execute(q1.List)
		r1 := m.Pop()
		m.Stack = make([]Value, len(savedStack))
		copy(m.Stack, savedStack)
		m.Execute(q2.List)
		r2 := m.Pop()
		m.Stack = savedStack[:len(savedStack)-1] // remove X
		m.Push(r1)
		m.Push(r2)
	})

	// infra: L1 [P] -> L2 — execute P within the list as a stack
	register("infra", func(m *Machine) {
		m.NeedStack(2, "infra")
		q := m.Pop()
		agg := m.Pop()
		if q.Typ != TypeList || agg.Typ != TypeList {
			joyErr("infra: list and quotation expected")
		}
		savedStack := m.Stack
		// Use the list as the stack (reverse: first element on top)
		m.Stack = make([]Value, len(agg.List))
		for i, v := range agg.List {
			m.Stack[len(agg.List)-1-i] = v
		}
		m.Execute(q.List)
		// Convert stack back to list
		result := make([]Value, len(m.Stack))
		for i, v := range m.Stack {
			result[len(m.Stack)-1-i] = v
		}
		m.Stack = savedStack
		m.Push(ListVal(result))
	})

	// treestep: T [P] treestep — depth-first leaf traversal
	// If leaf → push and execute P. If list → recurse on each child.
	register("treestep", func(m *Machine) {
		m.NeedStack(2, "treestep")
		p := m.Pop()
		t := m.Pop()
		if p.Typ != TypeList {
			joyErr("treestep: quotation expected")
		}
		treestepAux(m, t, p.List)
	})

	// treerec: T [O] [C] treerec — tree recursion
	// If leaf → push T, execute O. If branch → recurse each child, execute C.
	register("treerec", func(m *Machine) {
		m.NeedStack(3, "treerec")
		c := m.Pop()
		o := m.Pop()
		t := m.Pop()
		if o.Typ != TypeList || c.Typ != TypeList {
			joyErr("treerec: two quotations expected")
		}
		treerecAux(m, t, o.List, c.List)
	})

	// treegenrec: T [O1] [O2] [C] treegenrec — general tree recursion
	// If leaf → push T, execute O1. If branch → push T, execute O2,
	// push self-reference [O1 O2 C treegenrec], execute C.
	var treegenrecFn BuiltinFunc
	treegenrecFn = func(m *Machine) {
		m.NeedStack(4, "treegenrec")
		c := m.Pop()
		o2 := m.Pop()
		o1 := m.Pop()
		t := m.Pop()
		if o1.Typ != TypeList || o2.Typ != TypeList || c.Typ != TypeList {
			joyErr("treegenrec: three quotations expected")
		}
		if t.Typ != TypeList {
			// leaf
			m.Push(t)
			m.Execute(o1.List)
		} else {
			// branch
			m.Push(t)
			m.Execute(o2.List)
			selfQuot := []Value{o1, o2, c, BuiltinVal("treegenrec", treegenrecFn)}
			m.Push(ListVal(selfQuot))
			m.Execute(c.List)
		}
	}
	register("treegenrec", treegenrecFn)

	// some: A [B] -> X — true if any aggregate member satisfies B
	register("some", func(m *Machine) {
		m.NeedStack(2, "some")
		b := m.Pop()
		a := m.Pop()
		if b.Typ != TypeList {
			joyErr("some: quotation expected")
		}
		switch a.Typ {
		case TypeList:
			for _, item := range a.List {
				m.Push(item)
				m.Execute(b.List)
				if m.Pop().IsTruthy() {
					m.Push(BoolVal(true))
					return
				}
			}
		case TypeString:
			for _, ch := range a.Str {
				m.Push(CharVal(int64(ch)))
				m.Execute(b.List)
				if m.Pop().IsTruthy() {
					m.Push(BoolVal(true))
					return
				}
			}
		case TypeSet:
			for i := 0; i < SetSize; i++ {
				if a.Int&(1<<i) != 0 {
					m.Push(IntVal(int64(i)))
					m.Execute(b.List)
					if m.Pop().IsTruthy() {
						m.Push(BoolVal(true))
						return
					}
				}
			}
		default:
			joyErr("some: aggregate expected")
		}
		m.Push(BoolVal(false))
	})

	// all: A [B] -> X — true if all aggregate members satisfy B
	register("all", func(m *Machine) {
		m.NeedStack(2, "all")
		b := m.Pop()
		a := m.Pop()
		if b.Typ != TypeList {
			joyErr("all: quotation expected")
		}
		switch a.Typ {
		case TypeList:
			for _, item := range a.List {
				m.Push(item)
				m.Execute(b.List)
				if !m.Pop().IsTruthy() {
					m.Push(BoolVal(false))
					return
				}
			}
		case TypeString:
			for _, ch := range a.Str {
				m.Push(CharVal(int64(ch)))
				m.Execute(b.List)
				if !m.Pop().IsTruthy() {
					m.Push(BoolVal(false))
					return
				}
			}
		case TypeSet:
			for i := 0; i < SetSize; i++ {
				if a.Int&(1<<i) != 0 {
					m.Push(IntVal(int64(i)))
					m.Execute(b.List)
					if !m.Pop().IsTruthy() {
						m.Push(BoolVal(false))
						return
					}
				}
			}
		default:
			joyErr("all: aggregate expected")
		}
		m.Push(BoolVal(true))
	})

	// unary2: X Y [P] -> R S — apply P to X (yielding R), then to Y (yielding S)
	register("unary2", func(m *Machine) {
		m.NeedStack(3, "unary2")
		q := m.Pop()
		y := m.Pop()
		// X is on top of stack now
		if q.Typ != TypeList {
			joyErr("unary2: quotation expected")
		}
		savedStack := make([]Value, len(m.Stack))
		copy(savedStack, m.Stack)
		// Execute P on X (X is on stack top)
		m.Execute(q.List)
		r := m.Pop()
		// Restore stack minus X, push Y
		m.Stack = savedStack[:len(savedStack)-1]
		m.Push(y)
		m.Execute(q.List)
		s := m.Pop()
		m.Stack = savedStack[:len(savedStack)-1]
		m.Push(r)
		m.Push(s)
	})

	// while: [B] [P] -> ... — while B is true, execute P
	register("while", func(m *Machine) {
		m.NeedStack(2, "while")
		body := m.Pop()
		test := m.Pop()
		if test.Typ != TypeList || body.Typ != TypeList {
			joyErr("while: two quotations expected")
		}
		for {
			savedStack := make([]Value, len(m.Stack))
			copy(savedStack, m.Stack)
			m.Execute(test.List)
			cond := m.Pop()
			m.Stack = savedStack
			if !cond.IsTruthy() {
				break
			}
			m.Execute(body.List)
		}
	})
}

func treestepAux(m *Machine, t Value, p []Value) {
	if t.Typ != TypeList {
		// leaf
		m.Push(t)
		m.Execute(p)
	} else {
		// branch — recurse on each child
		for _, child := range t.List {
			treestepAux(m, child, p)
		}
	}
}

func treerecAux(m *Machine, t Value, o, c []Value) {
	if t.Typ != TypeList {
		// leaf
		m.Push(t)
		m.Execute(o)
	} else {
		// branch — recurse on each child, then combine
		for _, child := range t.List {
			treerecAux(m, child, o, c)
		}
		m.Execute(c)
	}
}

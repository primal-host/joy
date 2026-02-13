package main

func init() {
	// tailrec: [P] [T] [R] tailrec
	// Tail-recursive combinator. Go for-loop, no Go recursion.
	// Save stack, test P, restore. If truthy → T, done. If falsy → R, loop.
	register("tailrec", func(m *Machine) {
		m.NeedStack(3, "tailrec")
		r := m.Pop()
		t := m.Pop()
		p := m.Pop()
		if p.Typ != TypeList || t.Typ != TypeList || r.Typ != TypeList {
			joyErr("tailrec: three quotations expected")
		}
		for {
			saved := make([]Value, len(m.Stack))
			copy(saved, m.Stack)
			m.Execute(p.List)
			cond := m.Pop()
			m.Stack = saved
			if cond.IsTruthy() {
				m.Execute(t.List)
				return
			}
			m.Execute(r.List)
		}
	})

	// linrec: [P] [T] [R1] [R2] linrec
	// Linear recursion. If P → T. Else → R1, recurse, R2.
	register("linrec", func(m *Machine) {
		m.NeedStack(4, "linrec")
		r2 := m.Pop()
		r1 := m.Pop()
		t := m.Pop()
		p := m.Pop()
		if p.Typ != TypeList || t.Typ != TypeList || r1.Typ != TypeList || r2.Typ != TypeList {
			joyErr("linrec: four quotations expected")
		}
		linrecAux(m, p.List, t.List, r1.List, r2.List)
	})

	// binrec: [P] [T] [R1] [R2] binrec
	// Binary recursion. If P → T. Else → R1 (split), recurse both halves, R2 (combine).
	register("binrec", func(m *Machine) {
		m.NeedStack(4, "binrec")
		r2 := m.Pop()
		r1 := m.Pop()
		t := m.Pop()
		p := m.Pop()
		if p.Typ != TypeList || t.Typ != TypeList || r1.Typ != TypeList || r2.Typ != TypeList {
			joyErr("binrec: four quotations expected")
		}
		binrecAux(m, p.List, t.List, r1.List, r2.List)
	})

	// genrec: [P] [T] [R1] [R2] genrec
	// General recursion. If P → T. Else → R1, push [P T R1 R2 genrec], R2.
	var genrecFn BuiltinFunc
	genrecFn = func(m *Machine) {
		m.NeedStack(4, "genrec")
		r2 := m.Pop()
		r1 := m.Pop()
		t := m.Pop()
		p := m.Pop()
		if p.Typ != TypeList || t.Typ != TypeList || r1.Typ != TypeList || r2.Typ != TypeList {
			joyErr("genrec: four quotations expected")
		}
		saved := make([]Value, len(m.Stack))
		copy(saved, m.Stack)
		m.Execute(p.List)
		cond := m.Pop()
		m.Stack = saved
		if cond.IsTruthy() {
			m.Execute(t.List)
		} else {
			m.Execute(r1.List)
			// Push [P T R1 R2 genrec] quotation
			selfQuot := make([]Value, 0, 5)
			selfQuot = append(selfQuot, p, t, r1, r2)
			selfQuot = append(selfQuot, BuiltinVal("genrec", genrecFn))
			m.Push(ListVal(selfQuot))
			m.Execute(r2.List)
		}
	}
	register("genrec", genrecFn)

	// condlinrec: [[C1 B1 B1post] [C2 B2 B2post] ... [Dbase Dpost?]] condlinrec
	// Conditional linear recursion.
	register("condlinrec", func(m *Machine) {
		m.NeedStack(1, "condlinrec")
		clauses := m.Pop()
		if clauses.Typ != TypeList {
			joyErr("condlinrec: list of clauses expected")
		}
		condlinrecAux(m, clauses.List)
	})

	// primrec: X [I] [C] primrec — primitive recursion
	// Decompose X into constituents, execute I (initial), then C for each.
	register("primrec", func(m *Machine) {
		m.NeedStack(3, "primrec")
		c := m.Pop()
		i := m.Pop()
		x := m.Pop()
		if i.Typ != TypeList || c.Typ != TypeList {
			joyErr("primrec: two quotations expected")
		}
		switch x.Typ {
		case TypeInteger:
			// Push x, x-1, ..., 1
			n := x.Int
			if n < 0 {
				joyErr("primrec: non-negative integer expected")
			}
			for k := n; k >= 1; k-- {
				m.Push(IntVal(k))
			}
			m.Execute(i.List)
			for k := int64(0); k < n; k++ {
				m.Execute(c.List)
			}
		case TypeList:
			// Push each element (first to last)
			for idx := len(x.List) - 1; idx >= 0; idx-- {
				m.Push(x.List[idx])
			}
			m.Execute(i.List)
			for range x.List {
				m.Execute(c.List)
			}
		case TypeString:
			runes := []rune(x.Str)
			for idx := len(runes) - 1; idx >= 0; idx-- {
				m.Push(CharVal(int64(runes[idx])))
			}
			m.Execute(i.List)
			for range runes {
				m.Execute(c.List)
			}
		case TypeSet:
			var members []int
			for bit := 0; bit < SetSize; bit++ {
				if x.Int&(1<<bit) != 0 {
					members = append(members, bit)
				}
			}
			for idx := len(members) - 1; idx >= 0; idx-- {
				m.Push(IntVal(int64(members[idx])))
			}
			m.Execute(i.List)
			for range members {
				m.Execute(c.List)
			}
		default:
			joyErr("primrec: aggregate or integer expected")
		}
	})

	// condnestrec: [[C1 R1 R2 ...] [C2 ...] ... [D ...]] condnestrec
	// Conditional nested recursion.
	register("condnestrec", func(m *Machine) {
		m.NeedStack(1, "condnestrec")
		clauses := m.Pop()
		if clauses.Typ != TypeList {
			joyErr("condnestrec: list of clauses expected")
		}
		condnestrecAux(m, clauses.List)
	})
}

func linrecAux(m *Machine, p, t, r1, r2 []Value) {
	saved := make([]Value, len(m.Stack))
	copy(saved, m.Stack)
	m.Execute(p)
	cond := m.Pop()
	m.Stack = saved
	if cond.IsTruthy() {
		m.Execute(t)
	} else {
		m.Execute(r1)
		linrecAux(m, p, t, r1, r2)
		m.Execute(r2)
	}
}

func binrecAux(m *Machine, p, t, r1, r2 []Value) {
	saved := make([]Value, len(m.Stack))
	copy(saved, m.Stack)
	m.Execute(p)
	cond := m.Pop()
	m.Stack = saved
	if cond.IsTruthy() {
		m.Execute(t)
	} else {
		m.Execute(r1)
		second := m.Pop()
		binrecAux(m, p, t, r1, r2)
		m.Push(second)
		binrecAux(m, p, t, r1, r2)
		m.Execute(r2)
	}
}

func condlinrecAux(m *Machine, clauses []Value) {
	for i, clause := range clauses {
		if clause.Typ != TypeList || len(clause.List) == 0 {
			joyErr("condlinrec: each clause must be a non-empty list")
		}
		isLast := i == len(clauses)-1

		if !isLast {
			// Non-default clause: [Condition Body PostBody?]
			cond := clause.List[0]
			if cond.Typ != TypeList {
				joyErr("condlinrec: condition must be a quotation")
			}
			saved := make([]Value, len(m.Stack))
			copy(saved, m.Stack)
			m.Execute(cond.List)
			result := m.Pop()
			m.Stack = saved
			if result.IsTruthy() {
				if len(clause.List) > 1 {
					body := clause.List[1]
					if body.Typ == TypeList {
						m.Execute(body.List)
					}
				}
				if len(clause.List) > 2 {
					post := clause.List[2]
					if post.Typ == TypeList && len(post.List) > 0 {
						condlinrecAux(m, clauses)
						m.Execute(post.List)
					}
				}
				return
			}
		} else {
			// Default clause: [Body PostBody?]
			if len(clause.List) > 0 {
				body := clause.List[0]
				if body.Typ == TypeList {
					m.Execute(body.List)
				}
			}
			if len(clause.List) > 1 {
				post := clause.List[1]
				if post.Typ == TypeList && len(post.List) > 0 {
					condlinrecAux(m, clauses)
					m.Execute(post.List)
				}
			}
			return
		}
	}
}

func condnestrecAux(m *Machine, clauses []Value) {
	for i, clause := range clauses {
		if clause.Typ != TypeList || len(clause.List) == 0 {
			joyErr("condnestrec: each clause must be a non-empty list")
		}
		isLast := i == len(clauses)-1

		if !isLast {
			// Non-default clause: [Condition R1 R2 R3 ...]
			cond := clause.List[0]
			if cond.Typ != TypeList {
				joyErr("condnestrec: condition must be a quotation")
			}
			saved := make([]Value, len(m.Stack))
			copy(saved, m.Stack)
			m.Execute(cond.List)
			result := m.Pop()
			m.Stack = saved
			if result.IsTruthy() {
				parts := clause.List[1:]
				executeNested(m, clauses, parts)
				return
			}
		} else {
			// Default clause: [R1 R2 R3 ...]
			parts := clause.List
			executeNested(m, clauses, parts)
			return
		}
	}
}

// executeNested runs parts with recursive calls between each pair.
// [r1 r2 r3] → execute r1, recurse, r2, recurse, r3
func executeNested(m *Machine, clauses []Value, parts []Value) {
	for i, part := range parts {
		if part.Typ == TypeList {
			m.Execute(part.List)
		}
		if i < len(parts)-1 {
			condnestrecAux(m, clauses)
		}
	}
}

package main

func init() {
	register("pop", func(m *Machine) {
		m.NeedStack(1, "pop")
		m.Pop()
	})

	register("dup", func(m *Machine) {
		m.NeedStack(1, "dup")
		m.Push(m.Peek())
	})

	register("swap", func(m *Machine) {
		m.NeedStack(2, "swap")
		n := len(m.Stack)
		m.Stack[n-1], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-1]
	})

	register("rollup", func(m *Machine) {
		m.NeedStack(3, "rollup")
		n := len(m.Stack)
		// X Y Z -> Z X Y
		m.Stack[n-3], m.Stack[n-2], m.Stack[n-1] = m.Stack[n-1], m.Stack[n-3], m.Stack[n-2]
	})

	register("rolldown", func(m *Machine) {
		m.NeedStack(3, "rolldown")
		n := len(m.Stack)
		// X Y Z -> Y Z X
		m.Stack[n-3], m.Stack[n-2], m.Stack[n-1] = m.Stack[n-2], m.Stack[n-1], m.Stack[n-3]
	})

	register("rotate", func(m *Machine) {
		m.NeedStack(3, "rotate")
		n := len(m.Stack)
		// X Y Z -> Z Y X
		m.Stack[n-3], m.Stack[n-1] = m.Stack[n-1], m.Stack[n-3]
	})

	register("id", func(m *Machine) {
		// identity — does nothing
	})

	register("newstack", func(m *Machine) {
		m.Stack = nil
	})

	register("stack", func(m *Machine) {
		// Push a list of the current stack (top on front)
		items := make([]Value, len(m.Stack))
		for i, v := range m.Stack {
			items[len(m.Stack)-1-i] = v
		}
		m.Push(ListVal(items))
	})

	register("unstack", func(m *Machine) {
		m.NeedStack(1, "unstack")
		top := m.Pop()
		if top.Typ != TypeList {
			joyErr("unstack: list expected")
		}
		m.Stack = nil
		// list is top-first, push in reverse so first element ends on top
		for i := len(top.List) - 1; i >= 0; i-- {
			m.Push(top.List[i])
		}
	})

	register("popd", func(m *Machine) {
		m.NeedStack(2, "popd")
		n := len(m.Stack)
		// remove second element
		m.Stack = append(m.Stack[:n-2], m.Stack[n-1])
	})

	register("dupd", func(m *Machine) {
		m.NeedStack(2, "dupd")
		n := len(m.Stack)
		second := m.Stack[n-2]
		m.Stack = append(m.Stack[:n-1], second, m.Stack[n-1])
	})

	register("swapd", func(m *Machine) {
		m.NeedStack(3, "swapd")
		n := len(m.Stack)
		m.Stack[n-2], m.Stack[n-3] = m.Stack[n-3], m.Stack[n-2]
	})

	register("rollupd", func(m *Machine) {
		m.NeedStack(4, "rollupd")
		n := len(m.Stack)
		m.Stack[n-4], m.Stack[n-3], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-4], m.Stack[n-3]
	})

	register("rolldownd", func(m *Machine) {
		m.NeedStack(4, "rolldownd")
		n := len(m.Stack)
		m.Stack[n-4], m.Stack[n-3], m.Stack[n-2] = m.Stack[n-3], m.Stack[n-2], m.Stack[n-4]
	})

	register("rotated", func(m *Machine) {
		m.NeedStack(4, "rotated")
		n := len(m.Stack)
		m.Stack[n-4], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-4]
	})

	register("choice", func(m *Machine) {
		m.NeedStack(3, "choice")
		ifFalse := m.Pop()
		ifTrue := m.Pop()
		cond := m.Pop()
		if cond.IsTruthy() {
			m.Push(ifTrue)
		} else {
			m.Push(ifFalse)
		}
	})
}

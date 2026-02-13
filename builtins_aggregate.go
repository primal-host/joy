package main

import (
	"sort"
	"strconv"
	"strings"
)

func init() {
	register("cons", func(m *Machine) {
		m.NeedStack(2, "cons")
		agg := m.Pop()
		item := m.Pop()
		switch agg.Typ {
		case TypeList:
			newList := make([]Value, 0, len(agg.List)+1)
			newList = append(newList, item)
			newList = append(newList, agg.List...)
			m.Push(ListVal(newList))
		case TypeString:
			if item.Typ != TypeChar {
				joyErr("cons: char expected for string aggregate")
			}
			m.Push(StringVal(string(rune(item.Int)) + agg.Str))
		case TypeSet:
			if item.Int < 0 || item.Int >= SetSize {
				joyErr("cons: set member out of range")
			}
			m.Push(SetVal(agg.Int | (1 << item.Int)))
		default:
			joyErr("cons: aggregate expected")
		}
	})

	register("swons", func(m *Machine) {
		m.NeedStack(2, "swons")
		n := len(m.Stack)
		m.Stack[n-1], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-1]
		builtins["cons"](m)
	})

	register("first", func(m *Machine) {
		m.NeedStack(1, "first")
		a := m.Pop()
		switch a.Typ {
		case TypeList:
			if len(a.List) == 0 {
				joyErr("first: empty list")
			}
			m.Push(a.List[0])
		case TypeString:
			if a.Str == "" {
				joyErr("first: empty string")
			}
			m.Push(CharVal(int64(a.Str[0])))
		case TypeSet:
			if a.Int == 0 {
				joyErr("first: empty set")
			}
			for i := 0; i < SetSize; i++ {
				if a.Int&(1<<i) != 0 {
					m.Push(IntVal(int64(i)))
					return
				}
			}
		default:
			joyErr("first: aggregate expected")
		}
	})

	register("rest", func(m *Machine) {
		m.NeedStack(1, "rest")
		a := m.Pop()
		switch a.Typ {
		case TypeList:
			if len(a.List) == 0 {
				joyErr("rest: empty list")
			}
			m.Push(ListVal(a.List[1:]))
		case TypeString:
			if a.Str == "" {
				joyErr("rest: empty string")
			}
			m.Push(StringVal(a.Str[1:]))
		case TypeSet:
			if a.Int == 0 {
				joyErr("rest: empty set")
			}
			// remove lowest bit
			for i := 0; i < SetSize; i++ {
				if a.Int&(1<<i) != 0 {
					m.Push(SetVal(a.Int &^ (1 << i)))
					return
				}
			}
		default:
			joyErr("rest: aggregate expected")
		}
	})

	register("uncons", func(m *Machine) {
		m.NeedStack(1, "uncons")
		a := m.Peek()
		// first then rest
		builtins["first"](m)
		first := m.Pop()
		m.Push(a)
		builtins["rest"](m)
		rest := m.Pop()
		m.Push(first)
		m.Push(rest)
	})

	register("unswons", func(m *Machine) {
		m.NeedStack(1, "unswons")
		a := m.Peek()
		builtins["first"](m)
		first := m.Pop()
		m.Push(a)
		builtins["rest"](m)
		rest := m.Pop()
		m.Push(rest)
		m.Push(first)
	})

	register("size", func(m *Machine) {
		m.NeedStack(1, "size")
		a := m.Pop()
		switch a.Typ {
		case TypeList:
			m.Push(IntVal(int64(len(a.List))))
		case TypeString:
			m.Push(IntVal(int64(len(a.Str))))
		case TypeSet:
			count := int64(0)
			bits := a.Int
			for bits != 0 {
				count++
				bits &= bits - 1
			}
			m.Push(IntVal(count))
		default:
			joyErr("size: aggregate expected")
		}
	})

	register("at", func(m *Machine) {
		m.NeedStack(2, "at")
		idx := m.Pop()
		agg := m.Pop()
		i := int(idx.Int)
		switch agg.Typ {
		case TypeList:
			if i < 0 || i >= len(agg.List) {
				joyErr("at: index %d out of range", i)
			}
			m.Push(agg.List[i])
		case TypeString:
			if i < 0 || i >= len(agg.Str) {
				joyErr("at: index %d out of range", i)
			}
			m.Push(CharVal(int64(agg.Str[i])))
		default:
			joyErr("at: list or string expected")
		}
	})

	register("of", func(m *Machine) {
		m.NeedStack(2, "of")
		// of is swap at
		n := len(m.Stack)
		m.Stack[n-1], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-1]
		builtins["at"](m)
	})

	register("concat", func(m *Machine) {
		m.NeedStack(2, "concat")
		b := m.Pop()
		a := m.Pop()
		switch a.Typ {
		case TypeList:
			if b.Typ != TypeList {
				joyErr("concat: two lists expected")
			}
			newList := make([]Value, 0, len(a.List)+len(b.List))
			newList = append(newList, a.List...)
			newList = append(newList, b.List...)
			m.Push(ListVal(newList))
		case TypeString:
			if b.Typ != TypeString {
				joyErr("concat: two strings expected")
			}
			m.Push(StringVal(a.Str + b.Str))
		case TypeSet:
			if b.Typ != TypeSet {
				joyErr("concat: two sets expected")
			}
			m.Push(SetVal(a.Int | b.Int))
		default:
			joyErr("concat: aggregate expected")
		}
	})

	register("enconcat", func(m *Machine) {
		m.NeedStack(3, "enconcat")
		// X S T -> S concat [X] concat T concat — swapd cons concat
		n := len(m.Stack)
		m.Stack[n-2], m.Stack[n-3] = m.Stack[n-3], m.Stack[n-2]
		builtins["cons"](m)
		builtins["concat"](m)
	})

	register("take", func(m *Machine) {
		m.NeedStack(2, "take")
		n := m.Pop()
		a := m.Pop()
		count := int(n.Int)
		switch a.Typ {
		case TypeList:
			if count > len(a.List) {
				count = len(a.List)
			}
			if count < 0 {
				count = 0
			}
			m.Push(ListVal(a.List[:count]))
		case TypeString:
			if count > len(a.Str) {
				count = len(a.Str)
			}
			if count < 0 {
				count = 0
			}
			m.Push(StringVal(a.Str[:count]))
		default:
			joyErr("take: list or string expected")
		}
	})

	register("drop", func(m *Machine) {
		m.NeedStack(2, "drop")
		n := m.Pop()
		a := m.Pop()
		count := int(n.Int)
		switch a.Typ {
		case TypeList:
			if count > len(a.List) {
				count = len(a.List)
			}
			if count < 0 {
				count = 0
			}
			m.Push(ListVal(a.List[count:]))
		case TypeString:
			if count > len(a.Str) {
				count = len(a.Str)
			}
			if count < 0 {
				count = 0
			}
			m.Push(StringVal(a.Str[count:]))
		default:
			joyErr("drop: list or string expected")
		}
	})

	register("has", func(m *Machine) {
		m.NeedStack(2, "has")
		item := m.Pop()
		agg := m.Pop()
		switch agg.Typ {
		case TypeList:
			for _, v := range agg.List {
				if v.Equal(item) {
					m.Push(BoolVal(true))
					return
				}
			}
			m.Push(BoolVal(false))
		case TypeSet:
			if item.Int >= 0 && item.Int < SetSize {
				m.Push(BoolVal(agg.Int&(1<<item.Int) != 0))
			} else {
				m.Push(BoolVal(false))
			}
		default:
			joyErr("has: aggregate expected")
		}
	})

	register("in", func(m *Machine) {
		m.NeedStack(2, "in")
		// in is swap has
		n := len(m.Stack)
		m.Stack[n-1], m.Stack[n-2] = m.Stack[n-2], m.Stack[n-1]
		builtins["has"](m)
	})

	register("reverse", func(m *Machine) {
		m.NeedStack(1, "reverse")
		a := m.Pop()
		switch a.Typ {
		case TypeList:
			n := len(a.List)
			rev := make([]Value, n)
			for i, v := range a.List {
				rev[n-1-i] = v
			}
			m.Push(ListVal(rev))
		case TypeString:
			runes := []rune(a.Str)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			m.Push(StringVal(string(runes)))
		default:
			joyErr("reverse: list or string expected")
		}
	})

	register("name", func(m *Machine) {
		m.NeedStack(1, "name")
		a := m.Pop()
		switch a.Typ {
		case TypeBuiltin:
			m.Push(StringVal(a.Str))
		case TypeUserDef:
			m.Push(StringVal(a.Str))
		default:
			joyErr("name: function expected")
		}
	})

	register("body", func(m *Machine) {
		m.NeedStack(1, "body")
		a := m.Pop()
		if a.Typ != TypeUserDef {
			joyErr("body: user-defined symbol expected")
		}
		body, ok := m.Dict[a.Str]
		if !ok {
			joyErr("body: undefined: %s", a.Str)
		}
		m.Push(ListVal(body))
	})

	register("null", func(m *Machine) {
		m.NeedStack(1, "null")
		a := m.Pop()
		switch a.Typ {
		case TypeList:
			m.Push(BoolVal(len(a.List) == 0))
		case TypeString:
			m.Push(BoolVal(a.Str == ""))
		case TypeSet:
			m.Push(BoolVal(a.Int == 0))
		case TypeInteger, TypeChar:
			m.Push(BoolVal(a.Int == 0))
		case TypeFloat:
			m.Push(BoolVal(a.Flt == 0))
		case TypeBoolean:
			m.Push(BoolVal(a.Int == 0))
		default:
			m.Push(BoolVal(false))
		}
	})

	register("small", func(m *Machine) {
		m.NeedStack(1, "small")
		a := m.Pop()
		switch a.Typ {
		case TypeList:
			m.Push(BoolVal(len(a.List) <= 1))
		case TypeString:
			m.Push(BoolVal(len(a.Str) <= 1))
		case TypeSet:
			// 0 or 1 bits
			m.Push(BoolVal(a.Int == 0 || (a.Int&(a.Int-1)) == 0))
		case TypeInteger, TypeChar:
			m.Push(BoolVal(a.Int == 0 || a.Int == 1))
		case TypeFloat:
			m.Push(BoolVal(a.Flt == 0 || a.Flt == 1))
		case TypeBoolean:
			m.Push(BoolVal(true)) // true and false are both "small"
		default:
			m.Push(BoolVal(false))
		}
	})

	register("split", func(m *Machine) {
		m.NeedStack(2, "split")
		quot := m.Pop()
		agg := m.Pop()
		if quot.Typ != TypeList {
			joyErr("split: quotation expected")
		}
		if agg.Typ != TypeList {
			joyErr("split: list expected as second parameter")
		}
		var yes, no []Value
		for _, item := range agg.List {
			m.Push(item)
			m.Execute(quot.List)
			result := m.Pop()
			if result.IsTruthy() {
				yes = append(yes, item)
			} else {
				no = append(no, item)
			}
		}
		if yes == nil {
			yes = []Value{}
		}
		if no == nil {
			no = []Value{}
		}
		m.Push(ListVal(yes))
		m.Push(ListVal(no))
	})

	register("shunt", func(m *Machine) {
		m.NeedStack(2, "shunt")
		source := m.Pop()
		target := m.Pop()
		if source.Typ != TypeList || target.Typ != TypeList {
			joyErr("shunt: two lists expected")
		}
		result := make([]Value, len(target.List))
		copy(result, target.List)
		for _, v := range source.List {
			result = append(result, v)
		}
		m.Push(ListVal(result))
	})

	register("zip", func(m *Machine) {
		m.NeedStack(2, "zip")
		b := m.Pop()
		a := m.Pop()
		if a.Typ != TypeList || b.Typ != TypeList {
			joyErr("zip: two lists expected")
		}
		n := len(a.List)
		if len(b.List) < n {
			n = len(b.List)
		}
		result := make([]Value, n)
		for i := 0; i < n; i++ {
			result[i] = ListVal([]Value{a.List[i], b.List[i]})
		}
		m.Push(ListVal(result))
	})

	register("unit", func(m *Machine) {
		m.NeedStack(1, "unit")
		a := m.Pop()
		m.Push(ListVal([]Value{a}))
	})

	register("pair", func(m *Machine) {
		m.NeedStack(2, "pair")
		b := m.Pop()
		a := m.Pop()
		m.Push(ListVal([]Value{a, b}))
	})

	register("unpair", func(m *Machine) {
		m.NeedStack(1, "unpair")
		a := m.Pop()
		if a.Typ != TypeList || len(a.List) < 2 {
			joyErr("unpair: list of at least 2 elements expected")
		}
		m.Push(a.List[0])
		m.Push(a.List[1])
	})

	register("flatten", func(m *Machine) {
		m.NeedStack(1, "flatten")
		a := m.Pop()
		if a.Typ != TypeList {
			joyErr("flatten: list expected")
		}
		var flat []Value
		for _, item := range a.List {
			if item.Typ == TypeList {
				flat = append(flat, item.List...)
			} else {
				flat = append(flat, item)
			}
		}
		if flat == nil {
			flat = []Value{}
		}
		m.Push(ListVal(flat))
	})

	register("intern", func(m *Machine) {
		m.NeedStack(1, "intern")
		a := m.Pop()
		if a.Typ != TypeString {
			joyErr("intern: string expected")
		}
		// Parse the string as Joy source and return the first token as a value
		tokens := NewScanner(a.Str).ScanAll()
		if len(tokens) > 0 && tokens[0].Typ == TokAtom {
			m.Push(UserDefVal(tokens[0].Str))
		} else {
			joyErr("intern: atom expected")
		}
	})

	register("format", func(m *Machine) {
		m.NeedStack(2, "format")
		_factor := m.Pop() // ignored for basic formatting
		a := m.Pop()
		_ = _factor
		m.Push(StringVal(a.String()))
	})

	register("strtol", func(m *Machine) {
		m.NeedStack(2, "strtol")
		base := m.Pop()
		s := m.Pop()
		if s.Typ != TypeString {
			joyErr("strtol: string expected")
		}
		n := int64(0)
		b := base.Int
		for _, ch := range s.Str {
			var digit int64
			if ch >= '0' && ch <= '9' {
				digit = int64(ch - '0')
			} else if ch >= 'a' && ch <= 'z' {
				digit = int64(ch-'a') + 10
			} else if ch >= 'A' && ch <= 'Z' {
				digit = int64(ch-'A') + 10
			} else {
				break
			}
			if digit >= b {
				break
			}
			n = n*b + digit
		}
		m.Push(IntVal(n))
	})

	// sort: A -> B — sort list by Value.Compare or string by rune
	register("sort", func(m *Machine) {
		m.NeedStack(1, "sort")
		a := m.Pop()
		switch a.Typ {
		case TypeList:
			sorted := make([]Value, len(a.List))
			copy(sorted, a.List)
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Compare(sorted[j]) < 0
			})
			m.Push(ListVal(sorted))
		case TypeString:
			runes := []rune(a.Str)
			sort.Slice(runes, func(i, j int) bool {
				return runes[i] < runes[j]
			})
			m.Push(StringVal(string(runes)))
		default:
			joyErr("sort: list or string expected")
		}
	})

	register("strtod", func(m *Machine) {
		m.NeedStack(1, "strtod")
		s := m.Pop()
		if s.Typ != TypeString {
			joyErr("strtod: string expected")
		}
		s.Str = strings.TrimSpace(s.Str)
		var f float64
		for i, ch := range s.Str {
			if (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' || ch == '+' || ch == 'e' || ch == 'E' {
				continue
			}
			s.Str = s.Str[:i]
			break
		}
		_, _ = parseFloat(s.Str, &f)
		m.Push(FloatVal(f))
	})
}

func parseFloat(s string, f *float64) (bool, error) {
	var err error
	*f, err = strconv.ParseFloat(s, 64)
	return err == nil, err
}

package main

import (
	"fmt"
	"math"
	"strings"
)

type ValueType int

const (
	TypeBoolean ValueType = iota
	TypeChar
	TypeInteger
	TypeFloat
	TypeString
	TypeSet     // 32-bit bitmask stored in Int
	TypeList    // quotations are lists
	TypeBuiltin // carries Fn + Name
	TypeUserDef // carries Name, resolved at execution time
)

const SetSize = 32

type BuiltinFunc func(m *Machine)

type Value struct {
	Typ  ValueType
	Int  int64       // Boolean, Char, Integer, Set
	Flt  float64     // Float
	Str  string      // String, UserDef name, Builtin name
	List []Value     // List / Quotation
	Fn   BuiltinFunc // Builtin
}

func BoolVal(b bool) Value {
	v := Value{Typ: TypeBoolean}
	if b {
		v.Int = 1
	}
	return v
}

func CharVal(c int64) Value {
	return Value{Typ: TypeChar, Int: c}
}

func IntVal(n int64) Value {
	return Value{Typ: TypeInteger, Int: n}
}

func FloatVal(f float64) Value {
	return Value{Typ: TypeFloat, Flt: f}
}

func StringVal(s string) Value {
	return Value{Typ: TypeString, Str: s}
}

func SetVal(bits int64) Value {
	return Value{Typ: TypeSet, Int: bits}
}

func ListVal(items []Value) Value {
	return Value{Typ: TypeList, List: items}
}

func BuiltinVal(name string, fn BuiltinFunc) Value {
	return Value{Typ: TypeBuiltin, Str: name, Fn: fn}
}

func UserDefVal(name string) Value {
	return Value{Typ: TypeUserDef, Str: name}
}

func (v Value) IsTruthy() bool {
	switch v.Typ {
	case TypeBoolean:
		return v.Int != 0
	case TypeInteger, TypeChar:
		return v.Int != 0
	case TypeFloat:
		return v.Flt != 0
	case TypeString:
		return v.Str != ""
	case TypeList:
		return len(v.List) > 0
	case TypeSet:
		return v.Int != 0
	default:
		return true
	}
}

func (v Value) Equal(other Value) bool {
	if v.Typ != other.Typ {
		return false
	}
	switch v.Typ {
	case TypeBoolean, TypeChar, TypeInteger, TypeSet:
		return v.Int == other.Int
	case TypeFloat:
		return v.Flt == other.Flt
	case TypeString:
		return v.Str == other.Str
	case TypeList:
		if len(v.List) != len(other.List) {
			return false
		}
		for i := range v.List {
			if !v.List[i].Equal(other.List[i]) {
				return false
			}
		}
		return true
	case TypeBuiltin:
		return v.Str == other.Str
	case TypeUserDef:
		return v.Str == other.Str
	default:
		return false
	}
}

func (v Value) Compare(other Value) int {
	switch v.Typ {
	case TypeBoolean, TypeChar, TypeInteger:
		switch other.Typ {
		case TypeBoolean, TypeChar, TypeInteger:
			if v.Int < other.Int {
				return -1
			}
			if v.Int > other.Int {
				return 1
			}
			return 0
		case TypeFloat:
			fv := float64(v.Int)
			if fv < other.Flt {
				return -1
			}
			if fv > other.Flt {
				return 1
			}
			return 0
		}
	case TypeFloat:
		var ov float64
		switch other.Typ {
		case TypeBoolean, TypeChar, TypeInteger:
			ov = float64(other.Int)
		case TypeFloat:
			ov = other.Flt
		default:
			joyErr("compare: incompatible types")
		}
		if v.Flt < ov {
			return -1
		}
		if v.Flt > ov {
			return 1
		}
		return 0
	case TypeString:
		if other.Typ != TypeString {
			joyErr("compare: incompatible types")
		}
		if v.Str < other.Str {
			return -1
		}
		if v.Str > other.Str {
			return 1
		}
		return 0
	case TypeSet:
		if other.Typ != TypeSet {
			joyErr("compare: incompatible types")
		}
		if v.Int < other.Int {
			return -1
		}
		if v.Int > other.Int {
			return 1
		}
		return 0
	}
	joyErr("compare: unsupported types")
	return 0
}

func (v Value) NumericVal() float64 {
	switch v.Typ {
	case TypeInteger, TypeChar, TypeBoolean:
		return float64(v.Int)
	case TypeFloat:
		return v.Flt
	default:
		joyErr("numeric value expected")
		return 0
	}
}

func (v Value) String() string {
	switch v.Typ {
	case TypeBoolean:
		if v.Int != 0 {
			return "true"
		}
		return "false"
	case TypeChar:
		return fmt.Sprintf("'%c", rune(v.Int))
	case TypeInteger:
		return fmt.Sprintf("%d", v.Int)
	case TypeFloat:
		s := fmt.Sprintf("%g", v.Flt)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
			if !math.IsInf(v.Flt, 0) && !math.IsNaN(v.Flt) {
				s += ".0"
			}
		}
		return s
	case TypeString:
		return fmt.Sprintf("%q", v.Str)
	case TypeSet:
		var parts []string
		for i := 0; i < SetSize; i++ {
			if v.Int&(1<<i) != 0 {
				parts = append(parts, fmt.Sprintf("%d", i))
			}
		}
		return "{" + strings.Join(parts, " ") + "}"
	case TypeList:
		var parts []string
		for _, item := range v.List {
			parts = append(parts, item.String())
		}
		return "[" + strings.Join(parts, " ") + "]"
	case TypeBuiltin:
		return v.Str
	case TypeUserDef:
		return v.Str
	default:
		return "???"
	}
}

type JoyError struct {
	Msg string
}

func (e JoyError) Error() string {
	return e.Msg
}

func joyErr(format string, args ...any) {
	panic(JoyError{Msg: fmt.Sprintf(format, args...)})
}

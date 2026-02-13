package main

import (
	"fmt"
	"os"
	"strings"
)

type Machine struct {
	Stack   []Value
	Dict    map[string][]Value
	Autoput int // 0=off, 1=. (print top), 2=.. (print stack)
	ScopeID int // counter for HIDE/IN/END scope name mangling
}

func NewMachine() *Machine {
	return &Machine{
		Stack: nil,
		Dict:  make(map[string][]Value),
	}
}

func (m *Machine) Push(v Value) {
	m.Stack = append(m.Stack, v)
}

func (m *Machine) Pop() Value {
	n := len(m.Stack)
	if n == 0 {
		joyErr("stack underflow")
	}
	v := m.Stack[n-1]
	m.Stack = m.Stack[:n-1]
	return v
}

func (m *Machine) Peek() Value {
	n := len(m.Stack)
	if n == 0 {
		joyErr("stack underflow")
	}
	return m.Stack[n-1]
}

func (m *Machine) NeedStack(n int, name string) {
	if len(m.Stack) < n {
		joyErr("%s: expected %d parameters, got %d", name, n, len(m.Stack))
	}
}

func (m *Machine) Execute(program []Value) {
	for _, v := range program {
		switch v.Typ {
		case TypeBuiltin:
			v.Fn(m)
		case TypeUserDef:
			body, ok := m.Dict[v.Str]
			if !ok {
				joyErr("undefined: %s", v.Str)
			}
			m.Execute(body)
		default:
			// literal — push onto stack
			m.Push(v)
		}
	}
}

func (m *Machine) RunSafe(program []Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(JoyError); ok {
				err = je
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	m.Execute(program)
	return nil
}

// RunLine parses and executes a single line of Joy source.
func (m *Machine) RunLine(line string) error {
	tokens := NewScanner(line).ScanAll()
	program := NewParser(tokens, m).Parse()
	return m.RunSafe(program)
}

// RunSource parses and executes a complete Joy source string.
func (m *Machine) RunSource(source string) error {
	return m.RunLine(source)
}

// RunFile reads and executes a Joy source file.
func (m *Machine) RunFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	return m.RunSource(string(data))
}

// PrintStack prints the current stack (bottom to top).
func (m *Machine) PrintStack() string {
	var parts []string
	for _, v := range m.Stack {
		parts = append(parts, v.String())
	}
	return strings.Join(parts, " ")
}

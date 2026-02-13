package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Machine struct {
	Stack    []Value
	Dict     map[string][]Value
	Autoput  int              // 0=off, 1=. (print top), 2=.. (print stack)
	ScopeID  int              // counter for HIDE/IN/END scope name mangling
	LibPaths []string         // search directories for .joy files
	Included map[string]bool  // include guard (resolved path → loaded)
	Input    *bufio.Scanner   // input scanner for get builtin
}

func NewMachine() *Machine {
	return &Machine{
		Stack:    make([]Value, 0, 256),
		Dict:     make(map[string][]Value),
		Included: make(map[string]bool),
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
	for {
		for i, v := range program {
			switch v.Typ {
			case TypeBuiltin:
				v.Fn(m)
			case TypeUserDef:
				body, ok := m.Dict[v.Str]
				if !ok {
					joyErr("undefined: %s", v.Str)
				}
				if i == len(program)-1 {
					// Tail-call optimization: reuse loop instead of recursing
					program = body
					goto tailcall
				}
				m.Execute(body)
			default:
				// literal — push onto stack
				m.Push(v)
			}
		}
		return
	tailcall:
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

// ReadFile searches for a Joy source file and returns its contents.
// Search order: absolute/relative path, current dir, LibPaths, embedded FS.
func (m *Machine) ReadFile(name string) ([]byte, string, error) {
	// Absolute or relative path — try directly
	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		data, err := os.ReadFile(name)
		if err == nil {
			abs, _ := filepath.Abs(name)
			return data, abs, nil
		}
		return nil, "", fmt.Errorf("cannot read %s: %w", name, err)
	}

	// Try current directory
	if data, err := os.ReadFile(name); err == nil {
		abs, _ := filepath.Abs(name)
		return data, abs, nil
	}

	// Try LibPaths
	for _, dir := range m.LibPaths {
		full := filepath.Join(dir, name)
		if data, err := os.ReadFile(full); err == nil {
			abs, _ := filepath.Abs(full)
			return data, abs, nil
		}
	}

	// Try embedded FS
	data, err := readEmbeddedLib(name)
	if err == nil {
		return data, "embedded:" + name, nil
	}

	return nil, "", fmt.Errorf("cannot find %s", name)
}

// RunFile reads and executes a Joy source file with include guard.
func (m *Machine) RunFile(path string) error {
	data, resolved, err := m.ReadFile(path)
	if err != nil {
		return err
	}
	if m.Included[resolved] {
		return nil // already included
	}
	m.Included[resolved] = true
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

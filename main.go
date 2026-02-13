package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

func main() {
	m := NewMachine()
	m.Input = bufio.NewScanner(os.Stdin)

	// Set up library paths
	// 1. Exe-relative lib/ directory
	if exe, err := os.Executable(); err == nil {
		m.LibPaths = append(m.LibPaths, filepath.Join(filepath.Dir(exe), "lib"))
	}
	// 2. JOYLIB env var (colon-separated)
	if joylib := os.Getenv("JOYLIB"); joylib != "" {
		for _, p := range strings.Split(joylib, ":") {
			if p != "" {
				m.LibPaths = append(m.LibPaths, p)
			}
		}
	}

	// Parse flags
	noStdlib := false
	var files []string
	for _, arg := range os.Args[1:] {
		if arg == "--no-stdlib" {
			noStdlib = true
		} else {
			files = append(files, arg)
		}
	}

	// Auto-load standard library
	if !noStdlib {
		if err := m.RunFile("inilib.joy"); err != nil {
			// Not fatal — inilib may not exist in all environments
			_ = err
		}
	}

	if len(files) > 0 {
		// File execution mode
		for _, path := range files {
			if err := m.RunFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
		return
	}

	// REPL mode
	fmt.Println("Joy interpreter (Go) — type 'quit' to exit")

	if readline.IsTerminal(int(os.Stdin.Fd())) {
		replReadline(m)
	} else {
		replPipe(m)
	}
}

func replReadline(m *Machine) {
	homeDir, _ := os.UserHomeDir()
	histFile := filepath.Join(homeDir, ".joy_history")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "joy> ",
		HistoryFile:       histFile,
		HistorySearchFold: true,
	})
	if err != nil {
		replPipe(m)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if m.Echo > 0 {
			fmt.Println(line)
		}
		if err := m.RunLine(line); err != nil {
			reportError(err)
		} else if m.Autoput == 1 && len(m.Stack) > 0 {
			fmt.Println(m.Stack[len(m.Stack)-1].String())
		} else if m.Autoput == 2 && len(m.Stack) > 0 {
			fmt.Println(m.PrintStack())
		}
	}
	fmt.Println()
}

func replPipe(m *Machine) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("joy> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if m.Echo > 0 {
			fmt.Println(line)
		}
		if err := m.RunLine(line); err != nil {
			reportError(err)
		} else if m.Autoput == 1 && len(m.Stack) > 0 {
			fmt.Println(m.Stack[len(m.Stack)-1].String())
		} else if m.Autoput == 2 && len(m.Stack) > 0 {
			fmt.Println(m.PrintStack())
		}
	}
	fmt.Println()
}

func reportError(err error) {
	if je, ok := err.(JoyError); ok && je.Col > 0 {
		fmt.Fprintf(os.Stderr, "error at col %d: %s\n", je.Col, je.Msg)
	} else {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

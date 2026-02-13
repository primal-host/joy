package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	m := NewMachine()

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
		if err := m.RunLine(line); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
	fmt.Println()
}

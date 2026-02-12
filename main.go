package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	m := NewMachine()

	if len(os.Args) > 1 {
		// File execution mode
		for _, path := range os.Args[1:] {
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

package main

import "testing"

func benchMachine(b *testing.B) *Machine {
	b.Helper()
	m := NewMachine()
	if err := m.RunFile("inilib.joy"); err != nil {
		b.Fatalf("load inilib: %v", err)
	}
	return m
}

func BenchmarkFibonacci(b *testing.B) {
	m := benchMachine(b)
	if err := m.RunLine(`"numlib" libload`); err != nil {
		b.Fatalf("load numlib: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Stack = m.Stack[:0]
		if err := m.RunLine("20 fib"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFactorial(b *testing.B) {
	m := benchMachine(b)
	if err := m.RunLine(`"numlib" libload`); err != nil {
		b.Fatalf("load numlib: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Stack = m.Stack[:0]
		if err := m.RunLine("100 fact"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQsort(b *testing.B) {
	m := benchMachine(b)
	if err := m.RunLine(`"seqlib" libload`); err != nil {
		b.Fatalf("load seqlib: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Stack = m.Stack[:0]
		if err := m.RunLine("[10 3 7 1 8 4 2 9 6 5 15 13 11 14 12 20 18 16 19 17] qsort"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkArithmeticLoop(b *testing.B) {
	m := benchMachine(b)
	// Sum 1..100 using step
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Stack = m.Stack[:0]
		if err := m.RunLine("0 [1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20] [+] step"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrime(b *testing.B) {
	m := benchMachine(b)
	if err := m.RunLine(`"numlib" libload`); err != nil {
		b.Fatalf("load numlib: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Stack = m.Stack[:0]
		if err := m.RunLine("97 prime"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLazyTake(b *testing.B) {
	m := benchMachine(b)
	if err := m.RunLine(`"lazlib" libload`); err != nil {
		b.Fatalf("load lazlib: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Stack = m.Stack[:0]
		if err := m.RunLine("Naturals 50 Take"); err != nil {
			b.Fatal(err)
		}
	}
}

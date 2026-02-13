package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput runs fn and captures what it prints to stdout.
func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestBasicArithmetic(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"2 3 + .", "5\n"},
		{"7 6 * .", "42\n"},
		{"10 3 - .", "7\n"},
		{"20 4 / .", "5\n"},
		{"17 5 rem .", "2\n"},
		{"5 succ .", "6\n"},
		{"5 pred .", "4\n"},
		{"-3 abs .", "3\n"},
		{"-7 neg .", "7\n"},
		{"3 5 max .", "5\n"},
		{"3 5 min .", "3\n"},
		{"-1 sign .", "-1\n"},
		{"5 sign .", "1\n"},
		{"0 sign .", "0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestComparisons(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"3 5 < .", "true\n"},
		{"5 3 < .", "false\n"},
		{"3 3 = .", "true\n"},
		{"3 4 = .", "false\n"},
		{"5 3 > .", "true\n"},
		{"3 5 >= .", "false\n"},
		{"5 5 >= .", "true\n"},
		{"3 5 != .", "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestStackOps(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"5 dup + .", "10\n"},
		{"1 2 swap - .", "1\n"},
		{"1 2 3 rollup .s", "3 1 2\n"},
		{"1 2 3 rolldown .s", "2 3 1\n"},
		{"1 2 3 rotate .s", "3 2 1\n"},
		{"1 2 pop .", "1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestListOps(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"[1 2 3] first .", "1\n"},
		{"[1 2 3] rest .", "[2 3]\n"},
		{"5 [1 2 3] cons .", "[5 1 2 3]\n"},
		{"[1 2] [3 4] concat .", "[1 2 3 4]\n"},
		{"[1 2 3] size .", "3\n"},
		{"[1 2 3] 1 at .", "2\n"},
		{"[1 2 3] reverse .", "[3 2 1]\n"},
		{"[1 2 3] uncons . .", "[2 3]\n1\n"},
		{"5 unit .", "[5]\n"},
		{"1 2 pair .", "[1 2]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestCombinators(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"[3 2 +] i .", "5\n"},
		{"5 [dup *] i .", "25\n"},
		{"[3 2 4] [dup +] map .", "[6 4 8]\n"},
		{"0 [1 2 3 4 5] [+] fold .", "15\n"},
		{"[5 16 3 7 14] [10 <] filter .", "[5 3 7]\n"},
		{"1 2 3 [+] dip .s", "3 3\n"},       // dip saves 3, runs + on [1 2]=3, pushes 3 back
		{"3 [2 +] i .", "5\n"},
		{"true [1] [2] ifte .", "1\n"},
		{"false [1] [2] ifte .", "2\n"},
		{"true [1] [2] branch .", "1\n"},
		{"false [1] [2] branch .", "2\n"},
		{"3 [1 +] [dup *] cleave . .", "9\n4\n"}, // P=1+=4, Q=dup*=9; . pops 9, . pops 4
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestTimesAndStep(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"0 5 [1 +] times .", "5\n"},
		{"0 [1 2 3] [+] step .", "6\n"},
		{"2 3 [dup *] times .", "256\n"}, // 2->4->16->256 (dup * three times)
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestDefine(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"DEFINE sq == dup * .\n5 sq .", "25\n"},
		{"DEFINE double == 2 * .\n7 double .", "14\n"},
		{"DEFINE sq == dup * ; cube == dup dup * * .\n3 cube .", "27\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				// run all lines
				for _, line := range strings.Split(tt.input, "\n") {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					if err := m.RunLine(line); err != nil {
						t.Fatalf("error on %q: %v", line, err)
					}
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestLogic(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"true true and .", "true\n"},
		{"true false and .", "false\n"},
		{"true false or .", "true\n"},
		{"false false or .", "false\n"},
		{"true not .", "false\n"},
		{"false not .", "true\n"},
		{"true false xor .", "true\n"},
		{"true true xor .", "false\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestPredicates(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"42 integer .", "true\n"},
		{"3.14 integer .", "false\n"},
		{"3.14 float .", "true\n"},
		{"\"hello\" string .", "true\n"},
		{"[1 2] list .", "true\n"},
		{"42 list .", "false\n"},
		{"42 leaf .", "true\n"},
		{"[1] leaf .", "false\n"},
		{"true logical .", "true\n"},
		{"0 null .", "true\n"},
		{"[] null .", "true\n"},
		{"\"\" null .", "true\n"},
		{"1 null .", "false\n"},
		{"0 small .", "true\n"},
		{"1 small .", "true\n"},
		{"2 small .", "false\n"},
		{"[] small .", "true\n"},
		{"[1] small .", "true\n"},
		{"[1 2] small .", "false\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestSets(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"{1 2 3} .", "{1 2 3}\n"},
		{"{1 2 3} {2 3 4} and .", "{2 3}\n"},
		{"{1 2} {3 4} or .", "{1 2 3 4}\n"},
		{"{1 2 3} size .", "3\n"},
		{"{1 2 3} 5 has .", "false\n"},
		{"{1 2 3} 2 has .", "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`"hello" size .`, "5\n"},
		{`"hello" "world" concat .`, "\"helloworld\"\n"},
		{`"hello" reverse .`, "\"olleh\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestChars(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("'A ord ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "65\n" {
		t.Errorf("got %q, want %q", out, "65\n")
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"2.5 3.5 + .", "6.0\n"},  // 2.5+3.5=6.0
		{"9.0 sqrt .", "3.0\n"},
		{"3.7 floor .", "3\n"},
		{"3.2 ceil .", "4\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestNullary(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("5 [dup *] nullary .s"); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	// nullary should preserve 5, push 25
	if out != "5 25\n" {
		t.Errorf("got %q, want %q", out, "5 25\n")
	}
}

func TestUnary(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("5 [dup *] unary ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "25\n" {
		t.Errorf("got %q, want %q", out, "25\n")
	}
}

func TestIfteWithQuotation(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"5 [0 >] [dup *] [neg] ifte .", "25\n"},
		{"-3 [0 >] [dup *] [neg] ifte .", "3\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestRecursiveDefine(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("DEFINE factorial == [0 =] [pop 1] [dup 1 - factorial *] ifte ."); err != nil {
			t.Fatalf("error: %v", err)
		}
		if err := m.RunLine("5 factorial ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "120\n" {
		t.Errorf("got %q, want %q", out, "120\n")
	}
}

func TestWhile(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("1 [10 <] [2 *] while ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "16\n" {
		t.Errorf("got %q, want %q", out, "16\n")
	}
}

func TestInfra(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("[1 2 3] [+ +] infra ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "[6]\n" {
		t.Errorf("got %q, want %q", out, "[6]\n")
	}
}

func TestDip(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("1 2 3 [+] dip .s"); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	// 1 2 3: stack is [1 2 3]. dip pops [+] and 3. Runs + on [1 2] = 3. Push 3 back. Stack: [3 3]
	if out != "3 3\n" {
		t.Errorf("got %q, want %q", out, "3 3\n")
	}
}

func TestChoice(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"true 1 2 choice .", "1\n"},
		{"false 1 2 choice .", "2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestComments(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"2 3 + # this is a comment\n.", "5\n"},
		{"(* block comment *) 42 .", "42\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestErrorRecovery(t *testing.T) {
	m := NewMachine()
	err := m.RunLine("1 2 + pop pop pop")
	if err == nil {
		t.Error("expected stack underflow error")
	}
	// Machine should still be usable after error
	out := captureOutput(func() {
		if err := m.RunLine("42 ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "42\n" {
		t.Errorf("got %q, want %q", out, "42\n")
	}
}

func TestPlanExamples(t *testing.T) {
	// The exact test cases from the plan
	tests := []struct {
		input  string
		expect string
	}{
		{"2 3 + .", "5\n"},
		{"7 6 * .", "42\n"},
		{"[1 2 3] first .", "1\n"},
		{"[1 2 3] rest .", "[2 3]\n"},
		{"5 [1 2 3] cons .", "[5 1 2 3]\n"},
		{"[3 2 4] [dup +] map .", "[6 4 8]\n"},
		{"0 [1 2 3 4 5] [+] fold .", "15\n"},
		{"true [1] [2] ifte .", "1\n"},
		{"[5 16 3 7 14] [10 <] filter .", "[5 3 7]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestPlanDefineExample(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("DEFINE sq == dup * ."); err != nil {
			t.Fatalf("error: %v", err)
		}
		if err := m.RunLine("5 sq ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "25\n" {
		t.Errorf("got %q, want %q", out, "25\n")
	}
}

func TestTailrec(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		// countdown to 0
		{"10 [0 =] [] [1 -] tailrec .", "0\n"},
		// factorial via accumulator: 1 5 [0 =] [pop] [dup [*] dip pred] tailrec
		{"1 5 [0 =] [pop] [dup [*] dip pred] tailrec .", "120\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestLinrec(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		// factorial: 5 [null] [succ] [dup pred] [*] linrec
		{"5 [null] [succ] [dup pred] [*] linrec .", "120\n"},
		// factorial variant: 5 [0 =] [pop 1] [dup 1 -] [*] linrec
		{"5 [0 =] [pop 1] [dup 1 -] [*] linrec .", "120\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestBinrec(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		// fibonacci
		{"7 [small] [] [pred dup pred] [+] binrec .", "13\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestGenrec(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		// factorial: 5 [null] [succ] [dup pred] [i *] genrec
		{"5 [null] [succ] [dup pred] [i *] genrec .", "120\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestCondlinrec(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		// factorial: 5 [[[null] [succ]] [[dup pred] [*]]] condlinrec
		{"5 [[[null] [succ]] [[dup pred] [*]]] condlinrec .", "120\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestCondnestrec(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		// factorial
		{"5 [[[null] [pop 1]] [[dup pred] [*]]] condnestrec .", "120\n"},
		// McCarthy 91
		{"99 [[[100 >] [10 -]] [[11 +] [] []]] condnestrec .", "91\n"},
		{"100 [[[100 >] [10 -]] [[11 +] [] []]] condnestrec .", "91\n"},
		{"101 [[[100 >] [10 -]] [[11 +] [] []]] condnestrec .", "91\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestHideInEnd(t *testing.T) {
	// Basic: hidden helper accessible inside IN, invisible outside
	t.Run("basic", func(t *testing.T) {
		m := NewMachine()
		src := `HIDE helper == 2 * IN double == helper END`
		if err := m.RunLine(src); err != nil {
			t.Fatalf("error: %v", err)
		}
		out := captureOutput(func() {
			if err := m.RunLine("5 double ."); err != nil {
				t.Fatalf("error: %v", err)
			}
		})
		if out != "10\n" {
			t.Errorf("got %q, want %q", out, "10\n")
		}
		// hidden name should not be directly accessible
		err := m.RunLine("5 helper .")
		if err == nil {
			t.Error("expected error accessing hidden name 'helper'")
		}
	})

	// Nested HIDE
	t.Run("nested", func(t *testing.T) {
		m := NewMachine()
		src := `HIDE
			inner == 10 +
		IN
			HIDE
				secret == 100 +
			IN
				add100 == secret
			END
			add10 == inner
		END`
		if err := m.RunLine(src); err != nil {
			t.Fatalf("error: %v", err)
		}
		out := captureOutput(func() {
			if err := m.RunLine("5 add10 ."); err != nil {
				t.Fatalf("error: %v", err)
			}
		})
		if out != "15\n" {
			t.Errorf("add10: got %q, want %q", out, "15\n")
		}
		out = captureOutput(func() {
			if err := m.RunLine("5 add100 ."); err != nil {
				t.Fatalf("error: %v", err)
			}
		})
		if out != "105\n" {
			t.Errorf("add100: got %q, want %q", out, "105\n")
		}
		// inner and secret should be inaccessible
		err := m.RunLine("5 inner .")
		if err == nil {
			t.Error("expected error accessing hidden name 'inner'")
		}
		err = m.RunLine("5 secret .")
		if err == nil {
			t.Error("expected error accessing hidden name 'secret'")
		}
	})

	// HIDE inside DEFINE block
	t.Run("hide_in_define", func(t *testing.T) {
		m := NewMachine()
		src := `DEFINE
			HIDE
				impl == dup *
			IN
				square == impl
			END
		.`
		if err := m.RunLine(src); err != nil {
			t.Fatalf("error: %v", err)
		}
		out := captureOutput(func() {
			if err := m.RunLine("7 square ."); err != nil {
				t.Fatalf("error: %v", err)
			}
		})
		if out != "49\n" {
			t.Errorf("got %q, want %q", out, "49\n")
		}
	})
}

func TestFloatMath(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"0.0 sin .", "0.0\n"},
		{"0.0 cos .", "1.0\n"},
		{"0.0 tan .", "0.0\n"},
		{"100.0 log10 .", "2.0\n"},
		{"2.0 10.0 pow .", "1024.0\n"},
		{"1.0 exp .", "2.718281828459045\n"},
		{"1.0 log .", "0.0\n"},
		{"0.0 asin .", "0.0\n"},
		{"1.0 acos .", "0.0\n"},
		{"0.0 atan .", "0.0\n"},
		{"1.0 1.0 atan2 .", "0.7853981633974483\n"},
		{"0.5 3 ldexp .", "4.0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestFrexpModf(t *testing.T) {
	// frexp: 8.0 -> 0.5 3
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("8.0 frexp . ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "4\n0.5\n" {
		t.Errorf("frexp: got %q, want %q", out, "4\n0.5\n")
	}

	// modf: 3.75 -> frac=0.75, int=3.0  (frac pushed first, then int on top)
	m = NewMachine()
	out = captureOutput(func() {
		if err := m.RunLine("3.75 modf . ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "3.0\n0.75\n" {
		t.Errorf("modf: got %q, want %q", out, "3.0\n0.75\n")
	}
}

func TestRandom(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		// Seed with 42, get first random, seed again, should get same
		if err := m.RunLine("42 srand rand ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	first := out
	m = NewMachine()
	out = captureOutput(func() {
		if err := m.RunLine("42 srand rand ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != first {
		t.Errorf("srand/rand not deterministic: got %q then %q", first, out)
	}
}

func TestPrimrec(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		// factorial: 5! = 120
		{"5 [1] [*] primrec .", "120\n"},
		// sum: 1+2+3 = 6
		{"3 [0] [+] primrec .", "6\n"},
		// list primrec: rebuild list
		{"[1 2 3] [[]] [cons] primrec .", "[1 2 3]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestTreestep(t *testing.T) {
	// Sum all leaves: [1 [2 3]] -> 1+2+3 = 6
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("0 [1 [2 3]] [+] treestep ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "6\n" {
		t.Errorf("got %q, want %q", out, "6\n")
	}
}

func TestTreerec(t *testing.T) {
	// Count leaves: [1 [2 [3 4]]] — 4 leaves
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("[1 [2 [3 4]]] [pop 1] [+] treerec ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "4\n" {
		t.Errorf("got %q, want %q", out, "4\n")
	}
}

func TestSomeAll(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"[1 2 3] [3 =] some .", "true\n"},
		{"[1 2 3] [5 =] some .", "false\n"},
		{"[2 4 6] [2 rem 0 =] all .", "true\n"},
		{"[2 4 5] [2 rem 0 =] all .", "false\n"},
		{"[] [true] some .", "false\n"},
		{"[] [false] all .", "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"[3 1 2] sort .", "[1 2 3]\n"},
		{"[5 3 8 1 4] sort .", "[1 3 4 5 8]\n"},
		{"[] sort .", "[]\n"},
		{`"cab" sort .`, "\"abc\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestTimeBuiltins(t *testing.T) {
	// 0 gmtime first -> 1970 (year)
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("0 gmtime first ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "1970\n" {
		t.Errorf("gmtime year: got %q, want %q", out, "1970\n")
	}

	// Round-trip: mktime(gmtime(0)) should preserve timestamp
	// (may differ due to timezone, so use gmtime -> mktime with UTC)
	m = NewMachine()
	out = captureOutput(func() {
		if err := m.RunLine("0 gmtime size ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "9\n" {
		t.Errorf("gmtime list size: got %q, want %q", out, "9\n")
	}
}

func TestGetenv(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine(`"HOME" getenv size 0 > .`); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "true\n" {
		t.Errorf("getenv HOME: got %q, want %q", out, "true\n")
	}
}

func TestUndefs(t *testing.T) {
	m := NewMachine()
	if err := m.RunLine("DEFINE foo == bar baz ."); err != nil {
		t.Fatalf("error: %v", err)
	}
	out := captureOutput(func() {
		if err := m.RunLine("undefs size ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "2\n" {
		t.Errorf("undefs: got %q, want %q", out, "2\n")
	}
}

func TestFileIO(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.txt"

	// Write content to file, close, reopen, read back
	m := NewMachine()
	// Store path in a define for convenience
	if err := m.RunLine(fmt.Sprintf(`DEFINE testpath == "%s" .`, path)); err != nil {
		t.Fatalf("define: %v", err)
	}

	// Write
	if err := m.RunLine(`testpath "w" fopen`); err != nil {
		t.Fatalf("fopen: %v", err)
	}
	if err := m.RunLine(`"hello" fputchars`); err != nil {
		t.Fatalf("fputchars: %v", err)
	}
	if err := m.RunLine("fclose"); err != nil {
		t.Fatalf("fclose: %v", err)
	}

	// Read back
	if err := m.RunLine(`testpath "r" fopen`); err != nil {
		t.Fatalf("fopen read: %v", err)
	}
	if err := m.RunLine("5 fread"); err != nil {
		t.Fatalf("fread: %v", err)
	}
	out := captureOutput(func() {
		if err := m.RunLine("size ."); err != nil {
			t.Fatalf("size: %v", err)
		}
	})
	if out != "5\n" {
		t.Errorf("fread size: got %q, want %q", out, "5\n")
	}
	if err := m.RunLine("fclose"); err != nil {
		t.Fatalf("fclose read: %v", err)
	}
}

func TestFileSeekTell(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/seek.txt"

	m := NewMachine()
	// Write "abcde"
	if err := m.RunLine(fmt.Sprintf(`"%s" "w" fopen "abcde" fputchars fclose`, path)); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Open, seek to position 2, read 1 byte
	if err := m.RunLine(fmt.Sprintf(`"%s" "r" fopen`, path)); err != nil {
		t.Fatalf("fopen: %v", err)
	}
	if err := m.RunLine("2 0 fseek"); err != nil {
		t.Fatalf("fseek: %v", err)
	}
	out := captureOutput(func() {
		if err := m.RunLine("ftell ."); err != nil {
			t.Fatalf("ftell: %v", err)
		}
	})
	if out != "2\n" {
		t.Errorf("ftell: got %q, want %q", out, "2\n")
	}
	if err := m.RunLine("fclose"); err != nil {
		t.Fatalf("fclose: %v", err)
	}
}

func TestFremove(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/remove.txt"

	m := NewMachine()
	// Create file
	if err := m.RunLine(fmt.Sprintf(`"%s" "w" fopen fclose`, path)); err != nil {
		t.Fatalf("create: %v", err)
	}
	// Remove it
	out := captureOutput(func() {
		if err := m.RunLine(fmt.Sprintf(`"%s" fremove .`, path)); err != nil {
			t.Fatalf("fremove: %v", err)
		}
	})
	if out != "true\n" {
		t.Errorf("fremove: got %q, want %q", out, "true\n")
	}
	// Try to remove again — should fail
	out = captureOutput(func() {
		if err := m.RunLine(fmt.Sprintf(`"%s" fremove .`, path)); err != nil {
			t.Fatalf("fremove2: %v", err)
		}
	})
	if out != "false\n" {
		t.Errorf("fremove non-existent: got %q, want %q", out, "false\n")
	}
}

func TestFilePredicate(t *testing.T) {
	m := NewMachine()
	out := captureOutput(func() {
		if err := m.RunLine("stdin file ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "true\n" {
		t.Errorf("file predicate: got %q, want %q", out, "true\n")
	}
	out = captureOutput(func() {
		if err := m.RunLine("42 file ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "false\n" {
		t.Errorf("file predicate int: got %q, want %q", out, "false\n")
	}
}

func TestInclude(t *testing.T) {
	dir := t.TempDir()
	// Write a library file
	libPath := dir + "/mylib.joy"
	os.WriteFile(libPath, []byte("DEFINE double == 2 * ."), 0644)

	m := NewMachine()
	m.LibPaths = append(m.LibPaths, dir)

	// Include the library
	if err := m.RunLine(`"mylib.joy" include`); err != nil {
		t.Fatalf("include: %v", err)
	}

	// Use the definition
	out := captureOutput(func() {
		if err := m.RunLine("7 double ."); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "14\n" {
		t.Errorf("got %q, want %q", out, "14\n")
	}
}

func TestIncludeGuard(t *testing.T) {
	dir := t.TempDir()
	// Write a library that pushes 42 each time it runs
	libPath := dir + "/counter.joy"
	os.WriteFile(libPath, []byte("42"), 0644)

	m := NewMachine()
	m.LibPaths = append(m.LibPaths, dir)

	// Include twice
	if err := m.RunLine(`"counter.joy" include`); err != nil {
		t.Fatalf("include 1: %v", err)
	}
	if err := m.RunLine(`"counter.joy" include`); err != nil {
		t.Fatalf("include 2: %v", err)
	}

	// Stack should only have one 42 (guard prevented second load)
	out := captureOutput(func() {
		if err := m.RunLine(".s"); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	if out != "42\n" {
		t.Errorf("include guard: got %q, want %q (double include should be prevented)", out, "42\n")
	}
}

func TestFormatf(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"3.14159 'f 8 4 formatf .", "\"  3.1416\"\n"},
		{"3.14159 'e 12 4 formatf .", "\"  3.1416e+00\"\n"},
		{"3.14159 'g 8 4 formatf .", "\"   3.142\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := NewMachine()
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestEmbeddedInilib(t *testing.T) {
	m := NewMachine()
	// Load inilib from embedded FS
	if err := m.RunFile("inilib.joy"); err != nil {
		t.Fatalf("include inilib: %v", err)
	}

	// Test some inilib definitions
	tests := []struct {
		input  string
		expect string
	}{
		{"5 sqr .", "25\n"},
		{"3 cube .", "27\n"},
		{"4 even .", "true\n"},
		{"3 odd .", "true\n"},
		{"5 positive .", "true\n"},
		{"-3 negative .", "true\n"},
		{"[1 2 3] sum .", "6\n"},
		{"[2 3 4] product .", "24\n"},
		{"[1 2 3] second .", "2\n"},
		{"[1 2 3] third .", "3\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

// newMachineWithStdlib creates a machine with inilib pre-loaded.
func newMachineWithStdlib(t *testing.T) *Machine {
	t.Helper()
	m := NewMachine()
	if err := m.RunFile("inilib.joy"); err != nil {
		t.Fatalf("load inilib: %v", err)
	}
	return m
}

func TestLibAgglib(t *testing.T) {
	m := newMachineWithStdlib(t)
	if err := m.RunLine(`"agglib" libload`); err != nil {
		t.Fatalf("load agglib: %v", err)
	}
	tests := []struct {
		input  string
		expect string
	}{
		{"1 2 pairlist .", "[1 2]\n"},
		{"[1 2 3] unpair .", "2\n"},
		{"1 5 from-to .", "[1 2 3 4 5]\n"},
		{"[10 20 30] average .", "20\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestLibSeqlib(t *testing.T) {
	m := newMachineWithStdlib(t)
	if err := m.RunLine(`"seqlib" libload`); err != nil {
		t.Fatalf("load seqlib: %v", err)
	}
	tests := []struct {
		input  string
		expect string
	}{
		{"[5 3 1 4 2] qsort .", "[1 2 3 4 5]\n"},
		{"[1 2 3] reverselist .", "[3 2 1]\n"},
		{`"abc" reversestring .`, "\"cba\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestLibNumlib(t *testing.T) {
	m := newMachineWithStdlib(t)
	if err := m.RunLine(`"numlib" libload`); err != nil {
		t.Fatalf("load numlib: %v", err)
	}
	tests := []struct {
		input  string
		expect string
	}{
		{"5 fact .", "120\n"},
		{"7 prime .", "true\n"},
		{"20 prime .", "false\n"},
		{"4 even .", "true\n"},
		{"3 odd .", "true\n"},
		{"12 8 gcd .", "4\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestLibLazlib(t *testing.T) {
	m := newMachineWithStdlib(t)
	if err := m.RunLine(`"lazlib" libload`); err != nil {
		t.Fatalf("load lazlib: %v", err)
	}
	tests := []struct {
		input  string
		expect string
	}{
		{"3 From 5 Take .", "[3 4 5 6 7]\n"},
		{"Naturals 10 N-th .", "10\n"},
		{"Ones 3 Take .", "[1 1 1]\n"},
		{"Naturals 3 Drop 4 Take .", "[3 4 5 6]\n"},
		{"Naturals [dup *] Map 5 Take .", "[0 1 4 9 16]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestLibLazlibFilter(t *testing.T) {
	m := newMachineWithStdlib(t)
	if err := m.RunLine(`"lazlib" libload`); err != nil {
		t.Fatalf("load lazlib: %v", err)
	}
	if err := m.RunLine(`"numlib" libload`); err != nil {
		t.Fatalf("load numlib: %v", err)
	}
	tests := []struct {
		input  string
		expect string
	}{
		{"Positives [even] Filter 5 Take .", "[2 4 6 8 10]\n"},
		{"Positives [prime] Filter 5 Take .", "[2 3 5 7 11]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestLibTyplib(t *testing.T) {
	m := newMachineWithStdlib(t)
	if err := m.RunLine(`"typlib" libload`); err != nil {
		t.Fatalf("load typlib: %v", err)
	}
	tests := []struct {
		input  string
		expect string
	}{
		// Stack ADT
		{"st_new 1 st_push 2 st_push 3 st_push st_top .", "3\n"},
		// Big sets
		{"bs_new 5 bs_insert 3 bs_insert 7 bs_insert 1 bs_insert .", "[1 3 5 7]\n"},
		{"[1 3 5 7] 3 bs_member .", "true\n"},
		{"[1 3 5 7] 4 bs_member .", "false\n"},
		{"[1 3 5] [3 5 7] bs_union .", "[1 3 5 7]\n"},
		// Dictionary
		{"d_new [1 \"one\"] d_add [2 \"two\"] d_add dup 2 d_look .", "[2 \"two\"]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestModule(t *testing.T) {
	// Basic MODULE with PRIVATE/PUBLIC
	t.Run("basic", func(t *testing.T) {
		m := NewMachine()
		src := `MODULE m1 PRIVATE a == "a"; b == "b" PUBLIC ab == a b concat; ba == b a concat; abba == ab ba concat END`
		if err := m.RunLine(src); err != nil {
			t.Fatalf("error: %v", err)
		}
		tests := []struct {
			input  string
			expect string
		}{
			{`m1.ab .`, "\"ab\"\n"},
			{`m1.ba .`, "\"ba\"\n"},
			{`m1.abba .`, "\"abba\"\n"},
		}
		for _, tt := range tests {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error on %q: %v", tt.input, err)
				}
			})
			if out != tt.expect {
				t.Errorf("%s: got %q, want %q", tt.input, out, tt.expect)
			}
		}
		// Private fields should NOT be accessible
		err := m.RunLine("m1.a .")
		if err == nil {
			t.Error("expected error accessing private field m1.a")
		}
		err = m.RunLine("m1.b .")
		if err == nil {
			t.Error("expected error accessing private field m1.b")
		}
	})

	// MODULE with nested HIDE in PUBLIC section
	t.Run("nested_hide", func(t *testing.T) {
		m := NewMachine()
		src := `MODULE m2
			PRIVATE a == "A"; b == "B"
			PUBLIC
				ab == a b concat; ba == b a concat;
				abba == ab ba concat;
				HIDE c == "C"; d == "D"
				IN
					cd == c d concat;
					abc == a b concat c concat
				END;
				bcd == b cd concat
			END`
		if err := m.RunLine(src); err != nil {
			t.Fatalf("error: %v", err)
		}
		tests := []struct {
			input  string
			expect string
		}{
			{`m2.ab .`, "\"AB\"\n"},
			{`m2.ba .`, "\"BA\"\n"},
			{`m2.abba .`, "\"ABBA\"\n"},
			{`m2.cd .`, "\"CD\"\n"},
			{`m2.abc .`, "\"ABC\"\n"},
			{`m2.bcd .`, "\"BCD\"\n"},
		}
		for _, tt := range tests {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error on %q: %v", tt.input, err)
				}
			})
			if out != tt.expect {
				t.Errorf("%s: got %q, want %q", tt.input, out, tt.expect)
			}
		}
		// Private fields should NOT be accessible
		for _, field := range []string{"m2.a", "m2.b", "m2.c", "m2.d"} {
			err := m.RunLine(field + " .")
			if err == nil {
				t.Errorf("expected error accessing private field %s", field)
			}
		}
	})

	// Two MODULEs with same private names don't conflict
	t.Run("no_conflict", func(t *testing.T) {
		m := NewMachine()
		src1 := `MODULE x PRIVATE val == 10 PUBLIC get == val END`
		src2 := `MODULE y PRIVATE val == 20 PUBLIC get == val END`
		if err := m.RunLine(src1); err != nil {
			t.Fatalf("error: %v", err)
		}
		if err := m.RunLine(src2); err != nil {
			t.Fatalf("error: %v", err)
		}
		out := captureOutput(func() {
			if err := m.RunLine("x.get ."); err != nil {
				t.Fatalf("error: %v", err)
			}
		})
		if out != "10\n" {
			t.Errorf("x.get: got %q, want %q", out, "10\n")
		}
		out = captureOutput(func() {
			if err := m.RunLine("y.get ."); err != nil {
				t.Fatalf("error: %v", err)
			}
		})
		if out != "20\n" {
			t.Errorf("y.get: got %q, want %q", out, "20\n")
		}
	})
}

func TestInilibSequor(t *testing.T) {
	m := newMachineWithStdlib(t)
	tests := []struct {
		input  string
		expect string
	}{
		{"42 numerical .", "true\n"},
		{"3.14 numerical .", "true\n"},
		{`"hello" numerical .`, "false\n"},
		{"true boolean .", "true\n"},
		{"{1 2 3} boolean .", "true\n"},
		{"42 boolean .", "false\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

// ============== Reference Library Tests ==============
// Tests for Phase 5b libraries, derived from reference test files.

func loadLib(t *testing.T, m *Machine, lib string) {
	t.Helper()
	if err := m.RunLine(fmt.Sprintf(`"%s" libload`, lib)); err != nil {
		t.Fatalf("load %s: %v", lib, err)
	}
}

func TestRefMtrlib(t *testing.T) {
	m := newMachineWithStdlib(t)
	loadLib(t, m, "mtrlib")
	loadLib(t, m, "numlib")
	tests := []struct {
		input  string
		expect string
	}{
		{`5 "hello" n-e-vector .`, "[\"hello\" \"hello\" \"hello\" \"hello\" \"hello\"]\n"},
		{`5 1 n-e-vector .`, "[1 1 1 1 1]\n"},
		{`5 [1 2 3] [+]sv-bin-v .`, "[6 7 8]\n"},
		{`[1 2 3] [10 20 30] [+]vv-bin-v .`, "[11 22 33]\n"},
		{`[1 2 3] [1 2 3] [*][+]vv-2bin-s .`, "14\n"},
		{`[1 2 3 4] v-1row-m .`, "[[1 2 3 4]]\n"},
		{`[1 2 3 4] v-1col-m .`, "[[1] [2] [3] [4]]\n"},
		{`[1 2 3 4] v-zdiag-m .`, "[[1 0 0 0] [0 2 0 0] [0 0 3 0] [0 0 0 4]]\n"},
		{`[[1 2] [3 4]] [[5 6] [7 8]] mm-vercat-m .`, "[[1 2] [3 4] [5 6] [7 8]]\n"},
		{`[[1 2] [3 4]] [[5 6] [7 8]] mm-horcat-m .`, "[[1 2 5 6] [3 4 7 8]]\n"},
		{`[[1 2 3] [4 5 6] [7 8 9] [10 11 12]] m-transpose-m .`, "[[1 4 7 10] [2 5 8 11] [3 6 9 12]]\n"},
		{`[[1 2] [3 4] [5 6]] 10 [+]ms-cbin-m .`, "[[11 12] [13 14] [15 16]]\n"},
		{`[[1 2 3] [4 5 6]] [[10 20 30] [40 50 60]] mm-add-m .`, "[[11 22 33] [44 55 66]]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestRefMthlib(t *testing.T) {
	m := newMachineWithStdlib(t)
	loadLib(t, m, "mthlib")
	tests := []struct {
		input  string
		expect string
	}{
		// calc uses Cambridge Polish (prefix) notation
		{`[+ 3 4] calc .`, "7\n"},
		{`[* 3 4] calc .`, "12\n"},
		{`[- 10 3] calc .`, "7\n"},
		{`[+ [* 2 3] [- 10 4]] calc .`, "12\n"},
		// diff: differentiate with respect to variable
		{`[a] [* 3 a] diff .`, "3\n"},
		{`[a] [+ a a] diff .`, "2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestRefSymlib(t *testing.T) {
	m := newMachineWithStdlib(t)
	loadLib(t, m, "symlib")
	if err := m.RunLine(`DEFINE unops == [not succ pred fact fib first rest reverse i intern]`); err != nil {
		t.Fatal(err)
	}
	if err := m.RunLine(`DEFINE binops == [and or + - * / = < > cons concat map filter]`); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		input  string
		expect string
	}{
		// Rev2Cam: Reverse Polish to Cambridge
		{`[2 3 * 4 5 + -] Rev2Cam .`, "[- [* 2 3] [+ 4 5]]\n"},
		// Rev2Tre: Reverse Polish to Tree (atoms wrapped in lists)
		{`[2 3 * 4 5 + -] Rev2Tre .`, "[- [* [2] [3]] [+ [4] [5]]]\n"},
		// Cam2Rev: Cambridge to Reverse Polish
		{`[+ 2 3] Cam2Rev .`, "[2 3 +]\n"},
		{`[- [* 2 3] [+ 4 5]] Cam2Rev .`, "[2 3 * 4 5 + -]\n"},
		// Cam2Tre: Cambridge to Tree
		{`[+ 2 3] Cam2Tre .`, "[+ [2] [3]]\n"},
		// Rev2Inf: Reverse Polish to Infix (binaries wrapped in lists)
		{`[2 3 +] Rev2Inf .`, "[[2 + 3]]\n"},
		{`[2 3 * 4 5 + -] Rev2Inf .`, "[[[2 * 3] - [4 + 5]]]\n"},
		// Cam2Rev followed by eval
		{`[- [* 2 3] [+ 4 5]] Cam2Rev i .`, "-3\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestRefGrmlib(t *testing.T) {
	m := newMachineWithStdlib(t)
	loadLib(t, m, "grmlib")
	// Define grammars from reference grmtst.joy
	setup := []string{
		`DEFINE prs-trace == pop`,
		`DEFINE tree == [ "big" _ "tree" ]`,
		`DEFINE names == [ "peter" _ "smith" | "paul" _ "jones" | "mary" _ "robinson" ]`,
		`DEFINE anyname == [ ["peter" | "paul" | "mary"] _ ["smith" | "jones" | "robinson"] ]`,
		`DEFINE arith == [ "x" | "(" _ $ arith _ "+" _ $ arith _ ")" ]`,
	}
	for _, s := range setup {
		if err := m.RunLine(s); err != nil {
			t.Fatalf("setup %q: %v", s, err)
		}
	}
	tests := []struct {
		input  string
		expect string
	}{
		// Parsing tests from grmtst.joy
		{`["big" "tree"] tree prs-test .`, "true\n"},
		{`["peter" "smith"] names prs-test .`, "true\n"},
		{`["paul" "jones"] names prs-test .`, "true\n"},
		{`["mary" "robinson"] names prs-test .`, "true\n"},
		{`["fred"] names prs-test .`, "false\n"},
		{`["peter" "robinson"] anyname prs-test .`, "true\n"},
		{`["paul" "smith"] anyname prs-test .`, "true\n"},
		{`["paul" "nurks"] anyname prs-test .`, "false\n"},
		{`["fred" "nurks"] anyname prs-test .`, "false\n"},
		// $ (non-terminal call) in parsing
		{`["mary" "robinson"] [$ anyname] prs-test .`, "true\n"},
		{`["mary" "robertson"] [$ anyname] prs-test .`, "false\n"},
		// prs-count: count matching prefixes
		{`["*" "*" "*" "*" "*" "."] [* "*"] prs-count .`, "6\n"},
		{`["*" "*" "*" "." "*" "*"] [* "*"] prs-count .`, "4\n"},
		{`["*" "*" "*" "." "*" "*"] [+ "*"] prs-count .`, "3\n"},
		// Generation: gen-accumulate collects generated strings
		{`8 ["The" _ ["cat" | "dog"] _ "sat" _ "on" _ "the" _ ["mat" | "lawn"]] gen-accumulate size .`, "4\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestRefPlglib(t *testing.T) {
	m := newMachineWithStdlib(t)
	loadLib(t, m, "plglib")
	tests := []struct {
		input  string
		expect string
	}{
		{`[p or not p] Min2Tre taut-test .`, "true\n"},
		{`[p and not p] Min2Tre taut-test .`, "false\n"},
		{`[p imp p] Min2Tre taut-test .`, "true\n"},
		{`[[p imp q] imp [not q imp not p]] Min2Tre taut-test .`, "true\n"},
		{`[[p and q] imp p] Min2Tre taut-test .`, "true\n"},
		{`[p imp [p or q]] Min2Tre taut-test .`, "true\n"},
		{`[[[p imp q] and [q imp r]] imp [p imp r]] Min2Tre taut-test .`, "true\n"},
		{`[[p iff q] iff [q iff p]] Min2Tre taut-test .`, "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestRefLsplib(t *testing.T) {
	m := newMachineWithStdlib(t)
	loadLib(t, m, "lsplib")
	// Each test runs eval with lib0 as environment
	if err := m.RunLine("DEFINE lenv == lib0"); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		input  string
		expect string
	}{
		{`lenv [QUOTE [Peter Paul Mary]] eval .`, "[Peter Paul Mary]\n"},
		{`lenv [CAR [QUOTE [Peter Paul Mary]]] eval .`, "Peter\n"},
		{`lenv [CDR [QUOTE [1 2 3 4 5]]] eval .`, "[2 3 4 5]\n"},
		{`lenv [CONS Fred [QUOTE [Peter Paul]]] eval .`, "[Fred Peter Paul]\n"},
		{`lenv [ATOM Fred] eval .`, "true\n"},
		{`lenv [NULL [QUOTE []]] eval .`, "true\n"},
		{`lenv [EQ 2 3] eval .`, "false\n"},
		{`lenv [+ [* 2 5] [- 10 7]] eval .`, "13\n"},
		{`lenv [and true false] eval .`, "false\n"},
		{`lenv [or true false] eval .`, "true\n"},
		{`lenv [[LAMBDA [lis] CAR [CDR lis]] [QUOTE [11 22 33]]] eval .`, "22\n"},
		// FOLDR test
		{`lenv [FOLDR [QUOTE [a b c]] [QUOTE [d e]] [LAMBDA [x y] CONS x y]] eval .`, "[a b c d e]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestRefReplib(t *testing.T) {
	m := newMachineWithStdlib(t)
	loadLib(t, m, "replib")
	// Define helpers from reptst.joy
	setup := []string{
		`DEFINE fact0 == [[pop null] [pop pop 1] [[dup pred] dip i *] ifte]`,
		`DEFINE fact == [[pop null] [[pop 1] dip] [[dup pred] dip i [*] dip] ifte]`,
		`DEFINE state == first first`,
		`DEFINE ones == 1 rep.c-stream`,
		`DEFINE integers == rep.ints`,
		`DEFINE nfib == [[pop small] [[pop 1] dip] [[pred dup pred] dip dip swap i [+] dip] ifte]`,
		`DEFINE nfib-fix == nfib rep.fix`,
		`DEFINE fact-lin == [null] [pop 1] [dup pred] [*] rep.linear`,
		`DEFINE fact-fix == fact-lin rep.fix`,
	}
	for _, s := range setup {
		if err := m.RunLine(s); err != nil {
			t.Fatalf("setup %q: %v", s, err)
		}
	}
	tests := []struct {
		input  string
		expect string
	}{
		// squaring via exe
		{`2 [dup *] rep.exe i pop .`, "4\n"},
		{`2 [dup *] rep.exe i i pop .`, "16\n"},
		{`2 [dup *] rep.exe i i i pop .`, "256\n"},
		// factorial via fix (fact0 — simple version)
		{`6 fact0 rep.fix i .`, "720\n"},
		// factorial via fix (fact — leaves trace on stack)
		{`6 fact rep.fix i pop .`, "720\n"},
		// integers stream
		{`integers i i i i i state .`, "5\n"},
		// ones stream
		{`ones i i i i i state .`, "1\n"},
		// nfib
		{`6 nfib-fix i pop .`, "13\n"},
		// linear convenience
		{`4 fact-fix i pop .`, "24\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out := captureOutput(func() {
				if err := m.RunLine(tt.input); err != nil {
					t.Fatalf("error: %v", err)
				}
			})
			if out != tt.expect {
				t.Errorf("got %q, want %q", out, tt.expect)
			}
		})
	}
}

func TestRefFraclib(t *testing.T) {
	m := newMachineWithStdlib(t)
	loadLib(t, m, "fraclib")
	// Just test that mandel runs without error and produces output
	out := captureOutput(func() {
		if err := m.RunLine("mandel"); err != nil {
			t.Fatalf("error: %v", err)
		}
	})
	// Mandelbrot should produce 30 lines of 75 chars each
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 30 {
		t.Errorf("expected 30 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if len(line) != 75 {
			t.Errorf("line %d: expected 75 chars, got %d", i, len(line))
		}
	}
}

// Silence unused import warning for fmt
var _ = fmt.Sprint

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

// Silence unused import warning for fmt
var _ = fmt.Sprint

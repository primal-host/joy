# Phase 1: Core Interpreter

## Context

Build a Joy programming language interpreter in Go from scratch. Joy is a stack-based concatenative language where programs are sequences of values and operators that manipulate a data stack. All computation proceeds by pushing values and applying operators — there are no variables or named parameters.

Five commits delivering a working interpreter with REPL, file execution, and 105 passing tests:
- **1a**: Core types — tagged union Value with 9 types
- **1b**: Lexer and parser — tokenize and parse Joy source
- **1c**: Execution engine — Machine struct with stack, dictionary, dispatch loop
- **1d**: Builtins — ~75 operators across 8 categories
- **1e**: REPL, file execution, tests

---

## Phase 1a: Core Types

### `value.go` — Tagged union Value struct

9 value types via `ValueType` enum:

| Type | Go Fields | Description |
|------|-----------|-------------|
| `TypeBoolean` | `Bool bool` | `true` / `false` |
| `TypeChar` | `Char rune` | Single character (`'A`) |
| `TypeInteger` | `Int int64` | 64-bit integers |
| `TypeFloat` | `Float float64` | 64-bit floats |
| `TypeString` | `Str string` | Strings (`"hello"`) |
| `TypeSet` | `Set uint64` | Bitfield sets `{0 1 2}`, max 64 members |
| `TypeList` | `List []Value` | Heterogeneous lists `[1 "a" true]` |
| `TypeBuiltin` | `Fn BuiltinFunc, Str string` | Native Go function |
| `TypeUserDef` | `Str string` | Name resolved from dictionary at runtime |

Methods: `String()`, `Equal()`, `Compare()`, `IsTruthy()`, `NumericValue() float64`

Constructors: `BoolVal`, `CharVal`, `IntVal`, `FloatVal`, `StringVal`, `SetVal`, `ListVal`, `BuiltinVal`, `UserDefVal`

Error handling: `JoyError` type with `joyErr(format, args...)` panic-based errors, caught by `RunSafe`.

### `go.mod`

Module: `github.com/primal-host/joy`

---

## Phase 1b: Lexer and Parser

### `scanner.go` — Tokenizer

Token types: `TokInt`, `TokFloat`, `TokChar`, `TokString`, `TokTrue`, `TokFalse`, `TokAtom`, `TokLBrack`, `TokRBrack`, `TokLBrace`, `TokRBrace`, `TokDefine`, `TokSemicolon`, `TokDot`, `TokEOF`

Features:
- Integer and float literals (including negative)
- Character literals (`'A`, `'\n`, `'\t`)
- String literals with escape sequences
- Set literals `{0 1 2}`
- List literals `[1 2 3]`
- `DEFINE` blocks with `;` separators and `.` terminator
- Comments: `#` line comments and `(* ... *)` block comments
- `.s` recognized as atom (not dot + s)

### `parser.go` — Parser

`NewParser(tokens, machine).Parse() -> []Value`

- Atoms resolved to builtins at parse time (TypeBuiltin)
- Unknown atoms become TypeUserDef (resolved at execution time from dictionary)
- `DEFINE name == body ;` stores body in machine dictionary
- Nested list parsing `[...]`
- Set parsing `{...}`

---

## Phase 1c: Execution Engine

### `machine.go` — Machine struct

```go
type Machine struct {
    Stack []Value
    Dict  map[string][]Value
}
```

Methods:
- `Push(v)`, `Pop() Value`, `Peek() Value`, `NeedStack(n, name)`
- `Execute(program []Value)` — dispatch loop: push literals, call builtins, resolve user defs
- `RunSafe(program) error` — wraps Execute with panic/recover
- `RunLine(line) error` — scan + parse + execute
- `PrintStack() string` — bottom to top display

Dispatch in `Execute`:
1. `TypeBuiltin` → call `v.Fn(m)` directly
2. `TypeUserDef` → look up in `m.Dict`, execute body
3. Everything else → push onto stack

### `builtins.go` — Registration

Global `builtins` map + `register(name, fn)` function. All builtin files use `init()` to self-register.

---

## Phase 1d: Builtins (~75 operators)

### `builtins_stack.go` — Stack manipulation (12 builtins)

`pop`, `dup`, `swap`, `rollup`, `rolldown`, `rotate`, `choice`, `stack`, `unstack`, `popd`, `dupd`, `swapd`

### `builtins_math.go` — Arithmetic and comparison (22 builtins)

Arithmetic: `+`, `-`, `*`, `/`, `rem`, `div`, `succ`, `pred`, `neg`, `abs`, `sign`, `max`, `min`

Conversion: `ord`, `chr`

Comparison: `<`, `<=`, `>`, `>=`, `!=`, `=`

Polymorphic: integer and float operands auto-promote.

### `builtins_logic.go` — Logical operators (4 builtins)

`and`, `or`, `xor`, `not` — work on booleans, integers (bitwise), and sets (union/intersection/complement).

### `builtins_aggregate.go` — Aggregate operations (22 builtins)

Core: `cons`, `swons`, `first`, `rest`, `uncons`, `unswons`

Access: `size`, `at`, `of`

Construction: `concat`, `take`, `drop`, `unit`, `pair`

Membership: `has`, `in`

Mutation: `reverse`, `split`, `zip`

Conversion: `name` (atom → string), `body` (name → definition list)

Format: `format` (stub, fixed in Phase 4a)

Polymorphic across lists, strings, and sets where applicable.

### `builtins_predicate.go` — Type predicates (9 builtins)

`integer`, `char`, `logical`, `float`, `string`, `list`, `set`, `leaf`, `user`

Pattern: peek at top, push boolean.

### `builtins_io.go` — Output (6 builtins)

`put`, `putch`, `putchars`, `.` (alias for put + newline), `.s` (print stack), `newline`

### `builtins_combinator.go` — Combinators (21 builtins)

Core: `i`, `x`, `dip`

Branching: `branch`, `ifte`, `cond`

Iteration: `times`, `step`, `map`, `filter`, `fold`

Quotation: `nullary`, `unary`, `binary`, `ternary`, `cleave`

Advanced: `infra`, `while`, `construct`

### `builtins_misc.go` — Miscellaneous (9 builtins)

Constants: `true`, `false`, `maxint`, `setsize`, `clock`, `time`

Control: `quit`, `abort`

Introspection: `help`, `case`, `opcase`, `typeof`, `sametype`, `equal`

---

## Phase 1e: REPL and Tests

### `main.go` — Entry point

- REPL mode: prompt `joy> `, read-eval-print loop with error recovery
- File mode: `joy file.joy` executes file and exits
- Banner: `Joy interpreter (Go) — type 'quit' to exit`

### `joy_test.go` — 105 test cases

Test categories:
- Arithmetic: `2 3 + .` → `5`, float promotion, division
- Comparisons: `<`, `>`, `=` on integers
- Stack: `dup`, `swap`, `rollup`, `rolldown`, `rotate`, `popd`, `dupd`
- Lists: `cons`, `first`, `rest`, `size`, `at`, `concat`, `take`, `drop`, `zip`
- Sets: union, intersection, complement, membership
- Strings: `cons`, `first`, `rest`, `size`, `concat`, `at`, `reverse`
- Chars: `'A ord`, `65 chr`
- Floats: `1.5 2.5 +`, `3.0 sqrt`
- Logic: `and`, `or`, `xor`, `not`
- Predicates: type checks
- Combinators: `i`, `x`, `dip`, `ifte`, `branch`, `map`, `filter`, `fold`, `times`, `step`, `nullary`, `unary`, `cleave`, `infra`, `while`
- DEFINE: user-defined words, recursive factorial
- Comments: `#` and `(* *)`
- Error recovery: `RunSafe` returns errors instead of panicking

### `.gitignore`

Ignores built `joy` binary.

---

## Implementation Order

Each numbered item = 1 commit.

1. `go.mod` + `value.go` — core types (9 value types, tagged union, constructors, comparison)
2. `scanner.go` + `parser.go` — lexer (15 token types) and parser (DEFINE, lists, sets, atoms)
3. `machine.go` + `builtins.go` — execution engine and builtin registration
4. `builtins_*.go` (8 files) — ~75 operators across stack, math, logic, aggregate, predicate, I/O, combinator, misc
5. `main.go` + `joy_test.go` — REPL, file execution, 105 passing tests
6. `.gitignore` — ignore built binary

---

## Verification

1. `go build` clean
2. `go test` — 105 tests pass
3. REPL: `2 3 + .` → `5`
4. REPL: `[1 2 3] [dup *] map .` → `[1 4 9]`
5. REPL: `DEFINE fact == [0 =] [pop 1] [dup 1 - fact *] ifte .`  then `5 fact .` → `120`
6. REPL: `{1 2 3} {2 3 4} and .` → `{2 3}`

# Phase 2: Recursion Combinators and Lexical Scoping

## Context

Phase 1 delivered a working Joy interpreter with ~75 builtins, REPL, file execution, and 105 passing tests. However, it lacked the recursion combinators that are essential for idiomatic Joy programming (factorial via `linrec`, fibonacci via `binrec`, etc.) and the HIDE/IN/END lexical scoping mechanism used by Joy standard libraries for encapsulation.

Two sub-phases:
- **2a**: HIDE/IN/END lexical scoping — compile-time name mangling
- **2b**: Recursion combinators (6 new) — tailrec, linrec, binrec, genrec, condlinrec, condnestrec

---

## Phase 2a: HIDE/IN/END Lexical Scoping

### 1. `scanner.go` — ADD 3 token types

| Token | Keywords |
|-------|----------|
| `TokHide` | `HIDE`, `PRIVATE` |
| `TokIn` | `IN` |
| `TokEnd` | `END` |

### 2. `machine.go` — ADD `ScopeID` field

```go
ScopeID int // counter for HIDE/IN/END scope name mangling
```

Generates unique scope prefixes (`__scope_1_`, `__scope_2_`, ...) across multiple `RunLine` calls.

### 3. `parser.go` — ADD scope stack and name mangling

HIDE/IN/END creates a compile-time scope:

```
HIDE
    helper == ... ;      (* private definition *)
IN
    public == ... helper ... ;  (* public definition using private *)
END
```

Implementation:
- `HIDE` pushes a new scope map onto a scope stack, increments `ScopeID`
- Definitions between `HIDE` and `IN` get mangled names: `helper` → `__scope_N_helper`
- Scope map records original → mangled name mapping
- Between `IN` and `END`, definitions are public but can reference mangled names
- `END` pops the scope map — mangled names become unreachable from outside
- Nested HIDE blocks supported via scope stack (inner-to-outer lookup)
- `resolveAtom` checks scope stack from innermost to outermost before falling back to builtins/UserDef

---

## Phase 2b: Recursion Combinators

### `builtins_recursion.go` — NEW (6 builtins)

All combinators implement classic Joy recursion patterns from the C reference.

### 1. `tailrec` — Tail recursion

`[P] [T] [R1] tailrec`

- P = predicate (test condition)
- T = then branch (base case)
- R1 = recurse body (must leave stack ready for next iteration)

Go `for` loop — no Go stack growth:
```
loop: if P → execute T, return
      else → execute R1, goto loop
```

Test: `10 [0 <=] [] [dup pred] tailrec` → countdown leaving `10 9 8 ... 0` on stack

### 2. `linrec` — Linear recursion

`[P] [T] [R1] [R2] linrec`

- P = predicate
- T = base case
- R1 = before recursion (decompose)
- R2 = after recursion (recompose)

Go recursion:
```
if P → execute T
else → execute R1, recurse, execute R2
```

Test: `5 [0 =] [1] [dup 1 -] [*] linrec` → `120` (factorial)

### 3. `binrec` — Binary recursion

`[P] [T] [R1] [R2] binrec`

- P = predicate
- T = base case
- R1 = split into two sub-problems
- R2 = combine results

Go recursion with two recursive calls:
```
if P → execute T
else → execute R1
        save top, recurse on remainder
        swap, recurse on saved value
        execute R2
```

Test: `7 [2 <] [] [dup 1 - swap 2 -] [+] binrec` → `13` (fibonacci)

### 4. `genrec` — General recursion

`[P] [T] [R1] [R2] genrec`

- P = predicate
- T = base case
- R1 = pre-recursion setup
- R2 = body that receives a self-reference quotation

Self-reference: pushes `[P T R1 R2 genrec]` onto stack before executing R2, allowing R2 to call it with `i`.

Uses `var genrecFn BuiltinFunc` pattern to create the self-reference quotation containing the builtin itself.

Test: `5 [null] [succ] [dup pred] [i *] genrec` → `120` (factorial)

### 5. `condlinrec` — Conditional linear recursion

`[[C1] [B1] [A1]] [[C2] [B2] [A2]] ... [[Cn] [Bn]] condlinrec`

Takes a list of `[condition body after]` clauses. First matching condition's body is executed; if `after` is present, recursion happens then `after` is executed. Last clause may omit condition (default case).

Go recursion:
```
for each clause:
    if condition matches → execute body
        if after present → recurse, execute after
        return
default clause: execute body, optionally recurse
```

Test: `5 [[0 =] [1] []] [[0 >] [dup 1 -] [*]]] condlinrec` → `120`

### 6. `condnestrec` — Conditional nested recursion

`[[C1] [B1]] [[C2] [B2]] ... condnestrec`

Like `condlinrec` but the body receives a self-reference quotation `[clauses condnestrec]` pushed before execution. Body can invoke it with `i` for nested recursion.

Test: `91 [[100 >] [10 -]] [[succ] [condnestrec] condnestrec]] condnestrec` → `91` (McCarthy 91 function)

---

## Tests

### `joy_test.go` — 15 new test cases (120 total)

| Test | Description | Expected |
|------|-------------|----------|
| tailrec countdown | `10 [0 <=] [] [dup pred] tailrec` | stack has 10..0 |
| tailrec factorial | `5 1 [swap 0 <=] [pop] [swap dup rolldown * swap pred] tailrec` | `120` |
| linrec factorial | `5 [0 =] [1] [dup 1 -] [*] linrec` | `120` |
| linrec factorial (alt) | `5 [null] [succ] [dup pred] [*] linrec` | `120` |
| binrec fibonacci | `7 [2 <] [] [dup 1 - swap 2 -] [+] binrec` | `13` |
| genrec factorial | `5 [null] [succ] [dup pred] [i *] genrec` | `120` |
| condlinrec factorial | Multi-clause conditional recursion | `120` |
| condnestrec factorial | Conditional nested recursion | `120` |
| condnestrec mccarthy91 | `91 → 91`, `100 → 91` | McCarthy 91 function |
| HIDE basic | Helper hidden after END | `15` from public, error on private |
| HIDE nested | Inner and outer scopes | Both public defs work |
| HIDE in DEFINE | Scoping inside DEFINE block | Works correctly |

---

## Implementation Order

Each numbered item = 1 commit.

**Phase 2a — Lexical Scoping (3 commits):**
1. `scanner.go` — add `TokHide`, `TokIn`, `TokEnd` token types
2. `machine.go` — add `ScopeID` field
3. `parser.go` — add HIDE/IN/END with scope stack and name mangling

**Phase 2b — Recursion Combinators (2 commits):**
4. `builtins_recursion.go` — 6 recursion combinators
5. `joy_test.go` — 15 new tests (tailrec, linrec, binrec, genrec, condlinrec, condnestrec, HIDE/IN/END)

---

## Verification

1. `go build` clean
2. `go test` — 120 tests pass (105 existing + 15 new)
3. REPL: `5 [0 =] [1] [dup 1 -] [*] linrec .` → `120`
4. REPL: `7 [2 <] [] [dup 1 - swap 2 -] [+] binrec .` → `13`
5. REPL: `HIDE helper == 2 * ; IN double == helper ; END` then `5 double .` → `10`, `helper` → error
6. Total builtins after Phase 2: ~81 (75 + 6 recursion combinators)

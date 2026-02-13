# Phase 6: Reference Parity & Polish

## Goal

Achieve ~96% builtin parity with the C reference implementation, pass the
comprehensive `jp-joytst.joy` acceptance test, port the final library (lazlib),
and add multi-line REPL input.

## Current State (end of Phase 5)

- 143 builtins implemented (vs 171 in C reference)
- 16/17 standard libraries ported (lazlib remaining)
- 8 reference test suites passing (336 total tests, 0 failures)
- Readline REPL with persistent history
- Parse errors show column positions

---

## Phase 6a: Missing Builtins (21 high/medium priority)

### Tier 1 — Trivial (16 builtins)

**Hyperbolic math** — add to `builtins_floatmath.go`:
- `cosh`, `sinh`, `tanh` — direct wrappers around `math.Cosh/Sinh/Tanh`

**Type-conditional combinators** — add to `builtins_combinator.go`:
- `ifchar`, `ifinteger`, `iffloat`, `ifstring`, `iflist`, `ifset`, `iflogical`, `iffile`
- Pattern: `X [T] [E] ifTYPE` — if X matches type, execute T, else E
- Reference: `interp.c` lines ~1150-1250

**Higher-arity combinators** — add to `builtins_combinator.go`:
- `app11` — apply quotation, keep 1 arg below, 1 result
- `app12` — apply quotation, 1 arg, 2 results
- `unary3` — apply unary quotation to 3 stack items
- `unary4` — apply unary quotation to 4 stack items
- Reference: `interp.c` lines ~1440-1530

### Tier 2 — Medium (5 builtins)

**Flag getters** — add to `builtins_system.go`:
- `echo` — push current echo flag value
- `autoput` — push current autoput flag value
- `undeferror` — push current undeferror flag value
- These complement the existing `setecho`, `setautoput`, `setundeferror`

**Continuations** — add to `builtins_system.go`:
- `conts` — push the continuation stack as a list
- Used by some advanced Joy programs for metaprogramming
- Implementation: capture the remaining program as a quotation

**Shell command** (optional, security-sensitive):
- `system` — `"command" system` executes a shell command
- May want to gate behind a `--allow-system` flag

### Tier 3 — Skip (16 builtins)

These are internal/debug and not needed for compatibility:
- `gc`, `__dump`, `__memoryindex`, `__memorymax`, `__symtabindex`,
  `__symtabmax`, `__html_manual`, `__latex_manual`, `__manual_list`,
  `manual`, `_help`, `fputstring` (alias for fputchars)

### Tests

Add tests for each new builtin. For if-type combinators, test each type
with both matching and non-matching values. For app/unary variants, verify
stack manipulation matches reference behavior.

### Verification

After implementation: 164/171 builtins = **95.9% parity**.

---

## Phase 6b: lazlib — Lazy Evaluation Library

### Port

`resources/src2/lazlib.joy` → `lib/lazlib.joy` (106 lines, pure Joy)

Provides lazy infinite lists via thunks:
- Constructors: `From`, `From-to`, `From-to-by`, `From-by`
- Accessors: `First`, `Second`, `Third`, `N-th`, `Take`, `Drop`, `Size`
- Combinators: `Map`, `Filter`, `Cons`, `Uncons`, `Null`
- Examples: `Naturals`, `Evens`, `Odds`, `Powers-of-2`, `Ones`, `Squares`

No new builtins needed — lazlib is implemented entirely in Joy using
existing primitives (`rest`, `first`, `cons`, `i`, `while`, `linrec`).

### Tests

Add `TestRefLazlib` from `laztst.joy`/`.out` (~50 test expressions):
- Infinite stream navigation: `Naturals`, `Evens`, `Powers-of-2`
- `Take`, `Drop`, `N-th` on infinite lists
- `Map` and `Filter` on lazy streams
- `From-to` and `From-to-by` finite ranges
- Prime filtering: `1000000 From [prime] Filter` (performance test)

---

## Phase 6c: jp-joytst.joy Acceptance Test

### Setup

The reference's comprehensive test file (`jp-joytst.joy`, 96 lines) requires:
1. `jp-joyjoy` library (96 lines) — a Joy-in-Joy metacircular evaluator
2. `libinclude` definition (already in our inilib)
3. `setecho 1`

### Approach

1. Port `jp-joyjoy.joy` → `lib/jp-joyjoy.joy`
2. Add `TestJpJoytst` that loads the library and runs test expressions
3. Compare output against `jp-joytst.out`

The test exercises:
- `joy0` — non-tracing self-evaluator
- `joy0s` — short-trace evaluator
- `joy0l` — long/detailed-trace evaluator
- Nested self-application at 1-5 levels deep
- GC statistics display (can stub/skip)

### Expected Issues

- `joy0l` detailed tracing may require `conts` builtin (Phase 6a)
- GC stats lines won't match — filter them from comparison
- May need `__settracegc` stub (currently a no-op, which is fine)

---

## Phase 6d: Multi-line REPL Input

### Problem

Currently, unclosed brackets require everything on one line:
```
joy> DEFINE foo == [1 2     ← error: unterminated
```

### Solution

Track bracket depth in the REPL loop. When a line has unmatched `[`, `{`,
or `(*`, prompt for continuation with `...> ` and accumulate lines until
balanced.

### Implementation

In `main.go`, both `replReadline` and `replPipe`:

```go
func readComplete(promptFn func(string) (string, error)) (string, error) {
    line, err := promptFn("joy> ")
    if err != nil { return "", err }
    for depth := bracketDepth(line); depth > 0; {
        cont, err := promptFn("...> ")
        if err != nil { break }
        line += "\n" + cont
        depth = bracketDepth(line)
    }
    return line, nil
}
```

`bracketDepth` counts `[` vs `]`, `{` vs `}`, and `(*` vs `*)` nesting.
Also handle unclosed string literals (odd number of unescaped `"`).

For readline mode, update the prompt dynamically via `rl.SetPrompt("...> ")`.

### Tests

Test multi-line DEFINE, multi-line list literals, nested brackets across
lines, and edge cases (comments containing brackets, strings containing
brackets).

---

## Implementation Order

| # | Commit | Est. Size |
|---|--------|-----------|
| 1 | `builtins_floatmath.go` — cosh, sinh, tanh | S |
| 2 | `builtins_combinator.go` — 8 if-type combinators | M |
| 3 | `builtins_combinator.go` — app11, app12, unary3, unary4 | M |
| 4 | `builtins_system.go` — echo, autoput, undeferror getters | S |
| 5 | `builtins_system.go` — conts | M |
| 6 | `joy_test.go` — tests for all new builtins | M |
| 7 | `lib/lazlib.joy` — lazy evaluation library | S |
| 8 | `joy_test.go` — TestRefLazlib | M |
| 9 | `lib/jp-joyjoy.joy` — metacircular evaluator | M |
| 10 | `joy_test.go` — TestJpJoytst | M |
| 11 | `main.go` — multi-line REPL input | M |
| 12 | `joy_test.go` — multi-line REPL tests | S |

---

## Verification Checklist

1. `go build` clean after each commit
2. `go test` — all existing + new tests pass
3. Builtin count: 164+ registered
4. lazlib: `10 Naturals Take .` → `[0 1 2 3 4 5 6 7 8 9]`
5. jp-joytst: output matches reference (modulo GC lines)
6. Multi-line: `DEFINE foo ==\n  [1 2 3]\n  .` works across 3 lines
7. All 8 existing reference tests still pass

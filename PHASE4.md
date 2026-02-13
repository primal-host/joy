# Phase 4: Standard Libraries, Performance, REPL, Hardening

## Context

Phase 3 delivered file I/O, library loading, and 195 builtins. The interpreter can load embedded `.joy` files via `include` with search paths and include guards. However, only a minimal `inilib.joy` existed — the C reference implementation ships 6 standard libraries (inilib, agglib, seqlib, numlib, lazlib, typlib) that together provide ~200 user-defined operations. Phase 4 ports these libraries and fills the remaining builtin gaps they require.

Five sub-phases, each independently committable:
- **4a**: Missing builtins (4 new/fixed) — format, unary2, ifTYPE predicates, get
- **4b**: Standard libraries (6 ported) — inilib expansion, agglib, seqlib, numlib, lazlib, typlib
- **4c**: Performance — stack pre-allocation, tail-call optimization, benchmarks
- **4d**: REPL improvements — setautoput, setecho, setundeferror
- **4e**: Hardening — recursion depth guard

---

## Phase 4a: Missing Builtins

### 1. `builtins_aggregate.go` — FIX `format`

**Before**: 2-arg stub (`X factor -> S`) that ignores factor and returns `a.String()`.
**After**: 4-arg `X C I J format -> S` matching C reference.

| Arg | Description |
|-----|-------------|
| X | Value to format |
| C | Char: `'d` (decimal), `'o` (octal), `'x` (hex), `'f` (fixed float), `'e` (scientific), `'g` (general float), `'s` (string) |
| I | Minimum field width |
| J | Precision (floats) or max width (strings) |

Implementation: `fmt.Sprintf` with `%*.*verb` pattern.

### 2. `builtins_combinator.go` — ADD `unary2`

`X Y [P] unary2 -> R S` — Apply P to X (yielding R), then apply P to Y (yielding S).

Save/restore stack around each application. Both results remain on stack.

**Used by**: seqlib.joy (`qsort1`, `mk_qsort`, `merge1`), numlib.joy (`deriv`)

### 3. `builtins_predicate.go` — ADD 6 type dispatch builtins

`X [T] [F] ifTYPE` — If X matches TYPE, execute T; otherwise execute F. X stays on stack.

| Builtin | Type Check |
|---------|------------|
| `ifinteger` | `TypeInteger` |
| `ifchar` | `TypeChar` |
| `iffloat` | `TypeFloat` |
| `ifstring` | `TypeString` |
| `iflist` | `TypeList` |
| `ifset` | `TypeSet` |

**Used by**: seqlib.joy (`reverse == [[]] [""] iflist swap shunt`)

### 4. `builtins_io.go` + `machine.go` — FIX `get`

`get -> X` — Read a line from stdin, parse as Joy, push result.

Added `Input *bufio.Scanner` field to Machine struct. Initialized from `os.Stdin` in `main.go`.

---

## Phase 4b: Standard Libraries

All libraries in `lib/` embedded via `embed.go`. Each adapted from C reference `resources/src2/`.

### 1. `lib/inilib.joy` — Expanded to full reference

- I/O helpers: `putln`, `space`, `bell`, `putstrings`, `ask`
- Operators: `dup2`, `pop2`, `newstack`, `truth`, `falsity`, `to-upper`, `to-lower`, `boolean`, `numerical`, `swoncat`
- Date/time: `weekdays`, `months`, `localtime-strings`, `today`, `now`, `show-todaynow`
- Program operators: `conjoin`, `disjoin`, `negate`
- Combinators: `sequor`, `sequand`, `dip2`, `dip3`, `call`, `i2`, `nullary2`, `repeat`, `forever`
- Library loading: `verbose`, `libload`, `basic-libload`, `special-libload`, `all-libload`

### 2. `lib/agglib.joy` — Aggregate library

- Unit constructors: `unitset`, `unitstring`, `unitlist`, `pairset`, `pairstring`, `pairlist`
- Access: `unpair`, `second`, `third`, `fourth`, `fifth`
- Conversions: `string2set`, `set2string`, `elements`
- Dipped variants: `nulld`, `consd`, `swonsd`, etc.
- Two-operand: `null2`, `cons2`, `uncons2`, `swons2`, `unswons2`
- Combinators: `mapr`, `foldr`, `stepr2`, `fold2`, `mapr2`, `foldr2`, `interleave2`, `zip`
- Range: `from-to`, `from-to-list`, `from-to-set`, `from-to-string`
- Stats: `sum`, `average`, `variance`

### 3. `lib/seqlib.joy` — Sequence library

Requires `"agglib" libload`. Defines:
- Output: `putlist`
- Reversal: `reverse` (uses `iflist`), `reverselist`, `reversestring`, `flatten`
- Sorting: `qsort`, `qsort1` (uses `unary2`), `mk_qsort`, `merge`, `merge1`
- Sequences: `product`, `scalarproduct`, `frontlist`, `subseqlist`, `powerlist1/2`
- Insertion: `insertlist`, `permlist`, `insert`, `delete`
- Matrix: `transpose`, `cartproduct`
- Tree: `treeshunt`, `treeflatten`, `treereverse`, `treestrip`, `treemap`, `treefilter`

### 4. `lib/numlib.joy` — Numerical library

Self-contained (builtins + inilib only). Defines:
- Predicates: `positive`, `negative`, `even`, `odd`, `prime`
- Functions: `fact`, `fib`, `gcd`
- Conversions: `fahrenheit`, `celsius`
- Constants: `pi`, `e`, `radians`, `sindeg`, `cosdeg`, `tandeg`
- Equations: `qroots`, `quadratic-formula`
- Calculus: `deriv` (uses `unary2`), `newton`, `use-newton`, `cube-root`

### 5. `lib/lazlib.joy` — Lazy list library

Lazy lists as quotations producing `[head tail-quotation]` pairs when forced with `i`.

- Generators: `From`, `From-to`, `From-to-by`, `From-by`
- Access: `First`, `Rest`, `Uncons`, `Take`, `Drop`, `N-th`
- Transforms: `Map`, `Filter` (via HIDE/IN/END with `_mf`/`_ff` helpers)
- Examples: `Naturals`, `Positives`, `Ones`

Key implementation detail: `From == [dup succ From [] cons cons] cons` — uses `[] cons cons` to build correct `[head tail-quotation]` pairs.

### 6. `lib/typlib.joy` — Type/ADT library

Uses HIDE/IN/END for encapsulation. Defines:
- Stack ADT: `st_new`, `st_push`, `st_null`, `st_top`, `st_pop`, `st_pull`
- Queue ADT: `q_new`, `q_null`, `q_add`, `q_addl`, `q_front`, `q_rem`
- Tree ADT: `t_new`, `t_reset`, `t_add`, `t_null`, `t_front`, `t_rem`
- Big Sets (BST): `bs_new`, `bs_union`, `bs_differ`, `bs_member`, `bs_insert`, `bs_delete`
- Dictionary: `d_new`, `d_null`, `d_add`, `d_union`, `d_differ`, `d_look`, `d_rem`

### 7. Library tests — 27 new test cases

`newMachineWithStdlib` helper loads inilib for all library tests. Test suites:
- `TestLibAgglib` — from-to, zip, sum, variance
- `TestLibSeqlib` — qsort, reverse, flatten, transpose
- `TestLibNumlib` — fib, fact, prime, gcd
- `TestLibLazlib` — Take, Drop, N-th on Naturals/Positives
- `TestLibLazlibFilter` — Map, Filter on lazy lists
- `TestLibTyplib` — stack, queue, big set, dictionary operations

---

## Phase 4c: Performance

### 1. Stack pre-allocation in `machine.go`

```go
Stack: make([]Value, 0, 256)
```

### 2. Tail-call optimization in `Execute`

When the last value in a program is a UserDef, reuse the loop instead of recursing:

```go
func (m *Machine) Execute(program []Value) {
    for {
        for i, v := range program {
            switch v.Typ {
            case TypeUserDef:
                if i == len(program)-1 {
                    program = body // tail-call: reuse loop
                    goto tailcall
                }
                m.Execute(body) // non-tail: recurse
            ...
            }
        }
        return
    tailcall:
    }
}
```

### 3. Benchmark suite — `joy_bench_test.go`

| Benchmark | Workload | Result (M2 Ultra) |
|-----------|----------|--------------------|
| `BenchmarkFibonacci` | `20 fib` | ~23ms |
| `BenchmarkFactorial` | `100 fact` | ~1.1ms |
| `BenchmarkQsort` | 20-element list | ~73us |
| `BenchmarkArithmeticLoop` | sum 20 numbers via step | ~14us |
| `BenchmarkPrime` | `97 prime` | ~7us |
| `BenchmarkLazyTake` | `Naturals 50 Take` | ~188us |

---

## Phase 4d: REPL Improvements

### `builtins_misc.go` + `machine.go` + `main.go`

| Builtin | Stack | Field | Description |
|---------|-------|-------|-------------|
| `setautoput` | `I ->` | `Autoput int` | 0=off, 1=print top after each line, 2=print stack |
| `setecho` | `I ->` | `Echo int` | 0=off, 1=echo input lines before executing |
| `setundeferror` | `I ->` | `UndefError int` | 0=error on undefined, 1=ignore |

REPL loop in `main.go` checks `m.Echo` (print line before executing) and `m.Autoput` (print top/stack after executing).

---

## Phase 4e: Hardening

### Recursion depth guard in `machine.go`

Added `Depth int` and `MaxDepth int` (default 10000) to Machine struct. Incremented/decremented in `Execute()`. Panics with `"recursion depth exceeded"` if limit hit.

---

## Bugs Found and Fixed

| Bug | Location | Fix |
|-----|----------|-----|
| lazlib `From` wrong pair structure | `lib/lazlib.joy` | `unitlist swap cons` → `[] cons cons` |
| lazlib `Take` cons direction | `lib/lazlib.joy` | `cons` → `swons` (result list below, head on top) |
| lazlib `Map` pair order | `lib/lazlib.joy` | Removed initial `swap` in Map/Filter |
| lazlib `_mf` extra dup | `lib/lazlib.joy` | Removed `[dup] dip` before `i` |
| rolldown/rollup confusion | `lib/lazlib.joy` | `rolldown` → `rollup` in 3 locations |
| condlinrec empty post-body | `builtins_recursion.go` | Removed `len(post.List) > 0` check (2 occurrences) |
| agglib `unswons2` missing swap | `lib/agglib.joy` | Added `swapd` at end |
| Test `bs_member` arg order | `joy_test.go` | `set elem` not `elem set` |
| Tests missing inilib | `joy_test.go` | Created `newMachineWithStdlib` helper |

---

## Implementation Order

Each numbered item = 1 commit.

**Phase 4a — Missing Builtins (4 commits):**
1. `builtins_aggregate.go` — fix `format` to 4-arg C-style formatter
2. `builtins_combinator.go` — add `unary2`
3. `builtins_predicate.go` — add `ifinteger`, `ifchar`, `iffloat`, `ifstring`, `iflist`, `ifset`
4. `builtins_io.go` + `machine.go` — fix `get` with Machine.Input scanner

**Phase 4b — Standard Libraries (7 commits):**
5. `lib/inilib.joy` — expand to full reference adaptation
6. `lib/agglib.joy` — port aggregate library
7. `lib/seqlib.joy` — port sequence library
8. `lib/numlib.joy` — port numerical library
9. `lib/lazlib.joy` — port lazy list library
10. `lib/typlib.joy` — port type/ADT library
11. `joy_test.go` — add library loading tests (27 new cases)

**Phase 4c — Performance (2 commits):**
12. `machine.go` — stack pre-allocation + tail-call optimization
13. `joy_bench_test.go` — benchmark suite

**Phase 4d — REPL Improvements (1 commit):**
14. `builtins_misc.go` + `machine.go` + `main.go` — setautoput, setecho, setundeferror

**Phase 4e — Hardening (1 commit):**
15. `machine.go` — recursion depth guard

---

## Verification

1. `go build` clean after each sub-phase
2. `go test` — all existing + 27 new library tests pass
3. Library loading chain: `"agglib" libload` → `"seqlib" libload` → etc.
4. REPL: `5 fact .` → `120` (from numlib)
5. REPL: `97 prime .` → `true` (from numlib)
6. REPL: `[5 3 1 4 2] qsort .` → `[1 2 3 4 5]` (from seqlib)
7. REPL: `3 From 5 Take .` → `[3 4 5 6 7]` (from lazlib)
8. REPL: `1 setautoput` then `2 3 +` shows `5` automatically
9. Benchmarks: `go test -bench=. -count=3`

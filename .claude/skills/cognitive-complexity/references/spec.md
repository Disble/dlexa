# Cognitive Complexity — Full Specification Reference

*Based on: "Cognitive Complexity — a new way of measuring understandability"*
*G. Ann Campbell, SonarSource S.A., Version 1.7, 29 August 2023*

---

## Table of contents

1. [B1 — Increments (what gets counted)](#b1--increments)
2. [B2 — Nesting level (what raises the depth counter)](#b2--nesting-level)
3. [B3 — Nesting increments (what receives a depth penalty)](#b3--nesting-increments)
4. [Compensating Usages](#compensating-usages)
5. [Design principles](#design-principles)

---

## B1 — Increments

There is a **+1 increment** for each of the following:

- `if`, `else if`, `else`, ternary operator
- `switch`
- `for`, `foreach`
- `while`, `do while`
- `catch`
- `goto LABEL`, `break LABEL`, `continue LABEL`, `break NUMBER`,
  `continue NUMBER`
- each contiguous **sequence of like binary logical operators** in a boolean
  expression (a new sequence starts every time the operator changes)
- each method that participates in a **recursion cycle** (direct or indirect)

---

## B2 — Nesting level

The following structures **increment the nesting level** when entered:

- `if`, `else if`, `else`, ternary operator
- `switch`
- `for`, `foreach`
- `while`, `do while`
- `catch`
- nested methods and method-like structures: lambdas, closures, inner
  functions, anonymous classes, etc.

`try` blocks, `finally` blocks, and top-level method definitions do **not**
increment the nesting level.

---

## B3 — Nesting increments

The following structures receive an **additional penalty equal to their current
nesting depth** (on top of their base +1):

- `if`, ternary operator
- `switch`
- `for`, `foreach`
- `while`, `do while`
- `catch`

`else`, `else if`, and `else` do **not** receive a nesting increment (they are
hybrid increments: they count +1 but the cost was already paid by the `if`).

---

## Compensating Usages

Cognitive Complexity is language-agnostic, but some languages lack common
structures. To avoid penalizing idiomatic code in those languages, the
following exceptions apply.

### COBOL — Missing `else if`

COBOL has no `else if`. When an `IF` appears as the **only statement** inside
an `ELSE` clause, treat the entire `ELSE … IF` as a single flat `else if`:

- The inner `IF` gets +1 (structural) but **no nesting penalty**.
- The `ELSE` keyword itself gets +0.

This mirrors how `else if` is counted in other languages.

```cobol
IF condition1       ← +1 struct, nesting=0
  ...
ELSE
  IF condition2     ← +1 struct, nesting=0  (treated as else if)
    ...
  ELSE
    IF condition3   ← +1 struct, nesting=0  (treated as else if)
      statement1
      IF condition4 ← +1 struct, +1 nesting (truly nested, nesting=1)
        ...
```

### JavaScript — Missing class structures (pre-ESM)

Many JS codebases use an outer function as a fake namespace/class. When an
outer function contains **only declarations at the top level** (no structural
statements directly inside it), it is treated as transparent:

- Nesting level inside it starts at **0**.
- If it contains any top-level structural statement (`if`, loop, etc.),
  it is treated normally (nesting starts at 1 inside nested functions).

### Python — Decorators

A Python function that contains **only** one nested function + one `return`
statement is treated as a transparent decorator wrapper:

- The outer function is ignored for nesting purposes.
- Nesting level inside the inner function starts at **0**.

If the outer function contains anything else (assignments, conditionals, etc.),
it is treated normally and the inner function starts at nesting = 1.

---

## Design principles

### Why Cyclomatic Complexity is insufficient

Cyclomatic Complexity (McCabe, 1976) counts the minimum number of test paths
through a method. It was designed for Fortran and does not include
`try/catch`, lambdas, or other modern structures. Its minimum score of 1 per
method also makes class-level aggregation meaningless.

### Why Cognitive Complexity is different

Cognitive Complexity is designed from **programmer intuition**, not graph
theory. Key decisions:

| Decision | Rationale |
|---|---|
| No increment for methods | Breaking code into methods reduces complexity; penalizing it would discourage good practice |
| `switch` = +1 (not +N per case) | A switch is readable at a glance; an `if-else if` chain must be read carefully |
| Logical sequences, not operators | `a && b && c` is not much harder than `a && b`; mixing `&&` and `||` is |
| Nesting increment | Nested structures multiply mental load; flat structures do not |
| Recursion = +1 | Recursion is a "meta-loop" that many programmers find hard; it deserves a penalty |
| Early `return` = +0 | Early returns often clarify code; penalizing them would discourage a useful pattern |
| `else` = +1 hybrid, no nesting penalty | The mental cost of an `else` was paid when reading the `if` |

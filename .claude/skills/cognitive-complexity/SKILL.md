---
name: cognitive-complexity
description: >
  Analyze source code and calculate its Cognitive Complexity score following
  the Sonar specification (G. Ann Campbell, 2023). Use this skill whenever the
  user asks to: measure code complexity, calculate Cognitive Complexity, audit
  understandability of a method or class, compare complexity across methods,
  refactor code to reduce cognitive load, or explain why a piece of code is
  hard to understand. Also trigger when the user mentions terms like
  "cognitive complexity", "code understandability", "nesting penalty",
  "complexity score", or asks "how complex is this code?". Prefer this skill
  over general code review when the focus is on control-flow understandability.
---

# Cognitive Complexity Analyzer

Calculate and explain the Cognitive Complexity of source code, following the
Sonar specification. The goal is to give the programmer an **intuitive,
accurate number** that reflects the mental effort required to understand a
method — not just its test coverage potential.

> For the full specification and language-specific rules, see
> `references/spec.md`. For worked examples, see `references/examples.md`.

---

## Core concepts (internalize these before scoring)

Cognitive Complexity is built on three rules and four increment types:

**Three rules**
1. **Ignore shorthand** — structures that condense code without adding mental
   burden get no penalty (method calls, null-coalescing operators).
2. **+1 for each break in linear flow** — every structure that forces the
   reader to deviate from top-to-bottom reading.
3. **+nesting for nested breaks** — each extra level of nesting adds one
   more to the penalty of the inner structure.

**Four increment types**
| Type | What it means |
|---|---|
| **Structural** | Control-flow structure; subject to nesting increment |
| **Hybrid** | `else`/`else if` — counted once, increases nesting level, but NOT subject to a nesting increment itself |
| **Fundamental** | Binary logical sequences, recursion, jumps-to-label; flat +1, no nesting |
| **Nesting** | Extra penalty based on nesting depth of a structural increment |

---

## Scoring procedure

Work through the method **top to bottom**, maintaining a `nesting_level`
counter. Apply the steps below for each construct encountered.

### Step 1 — Track nesting level

These constructs **increase** `nesting_level` when entered (and restore it on
exit):
- `if`, `else if`, `else`, ternary operator
- `switch`
- `for`, `foreach`
- `while`, `do while`
- `catch`
- Nested methods / lambdas / closures (even though they receive +0 themselves)

`try` and `finally` blocks do **not** increase nesting level.
Top-level method definitions do **not** increase nesting level.

### Step 2 — Assign increments

| Construct | Increment | Nesting penalty? |
|---|---|---|
| `if` | +1 (structural) | +`nesting_level` |
| `else if` / `elif` / `else` | +1 (hybrid) | No |
| Ternary `? :` | +1 (structural) | +`nesting_level` |
| `switch` (entire block) | +1 (structural) | +`nesting_level` |
| `for` / `foreach` | +1 (structural) | +`nesting_level` |
| `while` / `do while` | +1 (structural) | +`nesting_level` |
| `catch` | +1 (structural) | +`nesting_level` |
| `try` / `finally` | +0 | — |
| Lambda / nested method | +0 | — (but raises nesting for its body) |
| Each recursion cycle method | +1 (fundamental) | No |
| `goto LABEL` / `break LABEL` / `continue LABEL` / `break N` / `continue N` | +1 (fundamental) | No |
| Each mixed-operator sequence in a boolean expr | +1 (fundamental) per sequence | No |
| Null-coalescing (`?.`, `??`, `||=`, etc.) | +0 | — |
| Method calls | +0 | — |

**Boolean sequences rule**: count +1 per *contiguous run of the same operator*.
A change from `&&` to `||` (or vice-versa) starts a new sequence (+1 again).
Negation `!` does not start a new sequence on its own.

### Step 3 — Sum and annotate

Add up all increments. Annotate each line with `// +N (type, nesting=K)` so
the programmer can follow the math.

---

## Language-specific exceptions

Read `references/spec.md` → **Compensating Usages** before scoring code in
these languages:

- **COBOL** — `ELSE … IF` chains: treat as flat `else if`, no nesting penalty.
- **JavaScript (pre-ESM class)** — outer function used purely as a namespace
  (no top-level structural statements): treated as a transparent wrapper,
  nesting starts at 0 inside it.
- **Python decorators** — a function containing only one nested function + a
  return statement: the outer function is transparent; inner function starts
  at nesting = 0.

---

## Output format

Always produce a structured report using this template:

```
## Cognitive Complexity Analysis

**Method:** `methodName`
**Language:** <language>
**Total score:** N

### Annotated code
\```<language>
<code with inline // +N comments on every scored line>
\```

### Breakdown table
| Line | Construct | Increment | Nesting level | Running total |
|------|-----------|-----------|---------------|---------------|
| ...  | ...       | ...       | ...           | ...           |

### Interpretation
<1–3 sentences explaining what drives the score and what to watch out for.
Mention if any language-specific exception was applied.>

### Refactoring suggestions *(optional, only if score ≥ 10)*
<Concrete, actionable suggestions to reduce complexity, e.g. extract method,
invert guard clause, replace nested conditionals with early return.>
```

If the user provides **multiple methods**, score each one separately, then add
a **summary table** at the end comparing them.

---

## Common pitfalls

- `switch` is a **single** +1 regardless of how many `case` branches it has.
- `catch` adds +1; `try` and `finally` add +0.
- Early `return` (without a label) adds +0 — it often *reduces* complexity.
- `else` adds +1 (hybrid) but does NOT add a nesting penalty on itself.
- Lambdas add +0 but **do** push the nesting level for anything inside them.
- `#if`/`#ifdef` preprocessor conditionals count the same as runtime `if`.

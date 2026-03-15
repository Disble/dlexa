# Cognitive Complexity — Worked Examples

*All examples from the SonarSource white paper (Campbell, 2023)*

---

## Example 1 — Illustrating the core problem

Two methods with equal **Cyclomatic Complexity = 4** but very different
Cognitive Complexity scores.

### `sumOfPrimes` — Cognitive Complexity: 7

```java
int sumOfPrimes(int max) {
  int total = 0;
  OUT: for (int i = 1; i <= max; ++i) { // +1 (structural, nesting=0)
    for (int j = 2; j < i; ++j) {       // +2 (structural, nesting=1)
      if (i % j == 0) {                 // +3 (structural, nesting=2)
        continue OUT;                   // +1 (fundamental: jump-to-label)
      }
    }
    total += i;
  }
  return total;
} // Total: 1+2+3+1 = 7
```

### `getWords` — Cognitive Complexity: 1

```java
String getWords(int number) {
  switch (number) { // +1 (structural, nesting=0)
    case 1:  return "one";
    case 2:  return "a couple";
    case 3:  return "a few";
    default: return "lots";
  }
} // Total: 1
```

---

## Example 2 — Nesting levels with try/catch

```java
void myMethod () {
  try {                                               // +0 (try never counted)
    if (condition1) {                                 // +1 (nesting=0)
      for (int i = 0; i < 10; i++) {                 // +2 (nesting=1)
        while (condition2) { … }                     // +3 (nesting=2)
      }
    }
  } catch (ExcepType1 | ExcepType2 e) {              // +1 (nesting=0)
    if (condition2) { … }                            // +2 (nesting=1)
  }
} // Total: 1+2+3+1+2 = 9
```

---

## Example 3 — Lambda raises nesting level

```java
void myMethod2 () {
  Runnable r = () -> {         // +0 (lambda), but nesting level → 1
    if (condition1) { … }     // +2 (structural, nesting=1)
  };
} // Total: 2

#if DEBUG                      // +1 (structural, nesting=0)
void myMethod2 () {            // +0 (method, nesting=0)
  Runnable r = () -> {         // +0 (lambda), nesting level → 1
    if (condition1) { … }     // +3 (structural, nesting=2)
  };
} // Total: 4
#endif
```

---

## Example 4 — Boolean sequences

```java
// Each contiguous run of the same operator = 1 sequence = +1 fundamental

if (a          // +1 (structural, nesting=0)
    && b && c  // +1 (sequence of &&)
    || d || e  // +1 (new sequence: ||)
    && f)      // +1 (new sequence: &&)
// Total for this if: 4

if (a          // +1 (structural, nesting=0)
    &&         // +1 (sequence of &&)
    !(b && c)) // +1 (new && sequence inside negation)
// Total for this if: 3
```

---

## Example 5 — Real code, high complexity (score = 19)

From `org.sonar.java.resolve.JavaSymbol.java`:

```java
@Nullable
private MethodJavaSymbol overriddenSymbolFrom(ClassJavaType classType) {
  if (classType.isUnknown()) {                              // +1 (nesting=0)
    return Symbols.unknownMethodSymbol;
  }
  boolean unknownFound = false;
  List<JavaSymbol> symbols = classType.getSymbol().members().lookup(name);
  for (JavaSymbol overrideSymbol : symbols) {              // +1 (nesting=0)
    if (overrideSymbol.isKind(JavaSymbol.MTH)              // +2 (nesting=1)
        && !overrideSymbol.isStatic()) {                   // +1 (sequence)
      MethodJavaSymbol methodJavaSymbol = (MethodJavaSymbol)overrideSymbol;
      if (canOverride(methodJavaSymbol)) {                 // +3 (nesting=2)
        Boolean overriding = checkOverridingParameters(methodJavaSymbol, classType);
        if (overriding == null) {                          // +4 (nesting=3)
          if (!unknownFound) {                             // +5 (nesting=4)
            unknownFound = true;
          }
        } else if (overriding) {                          // +1 (hybrid)
          return methodJavaSymbol;
        }
      }
    }
  }
  if (unknownFound) {                                      // +1 (nesting=0)
    return Symbols.unknownMethodSymbol;
  }
  return null;
} // Total: 1+1+2+1+3+4+5+1+1 = 19
```

---

## Complexity score interpretation guide

| Score | Interpretation |
|-------|---------------|
| 0–5   | Low — easy to understand and maintain |
| 6–10  | Moderate — worth a second look but acceptable |
| 11–20 | High — consider refactoring; hard to test |
| 21+   | Very high — strong refactoring candidate |

*Note: these thresholds are guidelines, not hard rules in the spec.*

---

## Common refactoring patterns

| Problem | Solution |
|---|---|
| Deeply nested `if` blocks | Extract method, invert guard clause (early return) |
| Long `if-else if` chains | Replace with `switch` (drops from N to 1) |
| Mixed `&&` / `\|\|` expressions | Extract named boolean variables |
| Recursive method | Document clearly; consider iterative alternative |
| Nested lambdas | Extract to named methods |

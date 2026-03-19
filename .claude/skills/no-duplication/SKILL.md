---
name: no-duplication
description: >
  Eliminate code duplication detected by SonarQube in Go test files and manage
  testdata fixture exclusions. Trigger: when SonarQube reports duplicated lines
  density above threshold, when writing Go test code with repeated boilerplate,
  or when a quality gate fails on new_duplicated_lines_density.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
allowed-tools: Read, Edit, Write, Glob, Grep, Bash, mcp__sonarqube__analyze_file_list, mcp__sonarqube__get_project_quality_gate_status
---

## When to Use

- SonarQube quality gate fails on `new_duplicated_lines_density`
- Writing Go test functions that share input structs, loop patterns, or HTTP response builders
- Adding files under `testdata/` (scraped HTML, JSON fixtures, generated content)
- Two or more test functions in the same package share identical boilerplate blocks

---

## Critical Patterns

### Pattern 1 — Package-level vars for shared test vectors

When identical test case structs appear in 2+ test functions, extract to a package-level `var`.

```go
// BAD — same struct repeated in TestFoo and TestBar
func TestFoo(t *testing.T) {
    cases := []struct{ in, want string }{
        {"hello", "HELLO"},
        {"world", "WORLD"},
    }
    // ...
}
func TestBar(t *testing.T) {
    cases := []struct{ in, want string }{
        {"hello", "HELLO"},
        {"world", "WORLD"},
    }
    // ...
}

// GOOD
var sharedStringCases = []struct{ in, want string }{
    {"hello", "HELLO"},
    {"world", "WORLD"},
}
```

Use `var`, never `const` — slices and structs cannot be `const` in Go.

---

### Pattern 2 — Parameterized run-loop helpers

When 2+ test functions share the same `for _, tt := range cases { t.Run(...) }` structure, extract into a `t.Helper()` function.

```go
// GOOD — extracted helper
func runStringTransformTest(t *testing.T, fn func(string) string, name string, cases []struct{ in, want string }) {
    t.Helper()
    for _, tt := range cases {
        t.Run(tt.in, func(t *testing.T) {
            got := fn(tt.in)
            if got != tt.want {
                t.Errorf("%s(%q) = %q, want %q", name, tt.in, got, tt.want)
            }
        })
    }
}
```

---

### Pattern 3 — Response/fixture builder helpers

Repeated `*http.Response` or model struct builders should become `makeXxx(params) Type` helpers.

```go
// BAD — inline closure repeated per test
fn := func(req *http.Request) (*http.Response, error) {
    return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
}

// GOOD — shared builder
func httpResponse(statusCode int, body string) *http.Response {
    return &http.Response{
        StatusCode: statusCode,
        Body:       io.NopCloser(strings.NewReader(body)),
    }
}
```

---

### Pattern 4 — Cross-file helper files

When 2+ `_test.go` files in the same package share boilerplate, move shared helpers to a dedicated file.

| Scope | File name |
|-------|-----------|
| General test utilities | `helpers_test.go` |
| Concurrency test setup | `concurrent_helpers_test.go` |
| HTTP/network helpers | `http_helpers_test.go` |

All helpers must belong to the same `package foo_test` (or `package foo`) as the test files that use them.

---

### Pattern 5 — Golden-file / render assertion helpers

When 2+ integration tests share a "render → read golden → compare" block, extract into a named assertion.

```go
// GOOD
func assertMarkdownMatchesGolden(t *testing.T, term string, entries []Entry) {
    t.Helper()
    got := renderMarkdown(entries)
    golden := readGolden(t, term)
    if got != golden {
        t.Errorf("mismatch for %s\ngot:\n%s\nwant:\n%s", term, got, golden)
    }
}
```

---

### Pattern 6 — testdata exclusions in sonar-project.properties

Large scraped or generated files under `testdata/` have near-100% duplication density by design. Exclude them.

```properties
# sonar-project.properties
sonar.projectKey=my-project
sonar.sources=.

# Exclude testdata fixtures from duplication analysis
sonar.cpd.exclusions=**/testdata/**
# Optionally exclude from all analysis (coverage, issues, duplication)
sonar.exclusions=**/testdata/**
```

Create this file at the repository root if it does not exist.

---

## Decision Table — Which pattern to apply

| Symptom | Apply |
|---------|-------|
| Same `[]struct{...}` literal in 2+ test functions | Pattern 1 |
| Same `for _, tt := range` loop in 2+ test functions | Pattern 2 |
| Same `*http.Response` / model builder in 2+ tests | Pattern 3 |
| Boilerplate shared across 2+ `_test.go` files | Pattern 4 |
| Same render+compare block in 2+ integration tests | Pattern 5 |
| SonarQube flags `testdata/` HTML/JSON fixtures | Pattern 6 |

---

## SonarQube Post-Fix Workflow

1. Apply the relevant pattern(s) above.
2. Trigger re-analysis on modified files:

```
mcp__sonarqube__analyze_file_list(files: ["path/to/file_test.go", ...])
```

3. Verify the quality gate passed:

```
mcp__sonarqube__get_project_quality_gate_status(projectKey: "<key>")
```

4. If gate still fails, identify remaining duplicated blocks:

```
mcp__sonarqube__get_duplications(key: "<component-key>")
```

---

## Commands

```bash
# View current duplication density for a file
# (use mcp__sonarqube__get_component_measures instead of CLI)

# After editing sonar-project.properties, trigger full analysis
# via mcp__sonarqube__analyze_file_list on ALL modified files
```

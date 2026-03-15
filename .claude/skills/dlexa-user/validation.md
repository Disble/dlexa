# dlexa-user Skill Validation Checklist

This document provides manual validation tests to verify the skill works correctly when loaded by an LLM.

## Prerequisites

- Load the `dlexa-user` skill in a fresh LLM session
- Ensure dlexa binary is built and available in PATH (for tests that require execution)

---

## Validation Tests

### VT-1: Basic Invocation

**Test Prompt**:
```
Look up the Spanish word 'casa' using dlexa
```

**Expected LLM Behavior**:
- Invokes bash tool with command: `dlexa casa`
- Does NOT add unnecessary flags
- Recognizes default output is markdown

**Verification Method**:
- [ ] Check bash tool call contains `dlexa casa`
- [ ] Verify no other unnecessary flags are added

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-2: Format Selection

**Test Prompt**:
```
Get the definition of 'casa' in a format I can parse programmatically
```

**Expected LLM Behavior**:
- Recognizes "parse programmatically" means JSON
- Uses command: `dlexa --format json casa`
- Explains why JSON is chosen

**Verification Method**:
- [ ] Check bash tool call contains `--format json`
- [ ] Verify LLM mentions JSON is for programmatic parsing

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-3: Markdown Parsing

**Test Setup**:
Provide this markdown output to the LLM:
```
# Resultados para: "casa"

| Palabra | Definición | Fuente |
|---------|------------|--------|
| casa    | Edificio para habitar. | dpd |
| casa    | Lugar donde se vive. | demo |
```

**Test Prompt**:
```
Extract the definitions from this dlexa output
```

**Expected LLM Behavior**:
- Correctly parses pipe-delimited table structure
- Extracts both definitions from second column
- Identifies headword ("casa") and sources ("dpd", "demo")

**Verification Method**:
- [ ] LLM extracts "Edificio para habitar." from first row
- [ ] LLM extracts "Lugar donde se vive." from second row
- [ ] LLM correctly identifies table columns

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-4: JSON Navigation

**Test Setup**:
Provide this JSON output to the LLM:
```json
{
  "Request": {
    "Query": "casa",
    "Format": "json"
  },
  "Entries": [
    {
      "ID": "casa-1",
      "Headword": "casa",
      "Content": "Edificio para habitar.",
      "Source": "dpd"
    }
  ],
  "CacheHit": true
}
```

**Test Prompt**:
```
Extract the definition from this JSON output
```

**Expected LLM Behavior**:
- Navigates to `.Entries[0].Content` or uses jq pattern
- Extracts "Edificio para habitar."
- Mentions using jq for extraction (if suggesting command)

**Verification Method**:
- [ ] LLM correctly navigates JSON structure
- [ ] Extracts the Content field value
- [ ] If command suggested, uses jq correctly: `jq -r '.Entries[0].Content'`

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-5: Error Interpretation

**Test Setup**:
Provide this error scenario:
```
Exit code: 1
stderr:
Problem: source_lookup_failed
Source: dpd
Message: Failed to fetch data from source
```

**Test Prompt**:
```
dlexa returned an error. What went wrong?
```

**Expected LLM Behavior**:
- Identifies this as a source connectivity issue
- Mentions "source_lookup_failed" Problem code
- Suggests checking network, using `--doctor`, or trying later

**Verification Method**:
- [ ] LLM identifies as source/connectivity issue
- [ ] Mentions the Problem code explicitly
- [ ] Provides actionable troubleshooting steps

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-6: Cache Bypass

**Test Prompt**:
```
The data seems stale, get fresh results for 'casa'
```

**Expected LLM Behavior**:
- Recognizes "stale data" as cache issue
- Uses command: `dlexa --no-cache casa`
- Explains that `--no-cache` bypasses the 24-hour cache

**Verification Method**:
- [ ] Command includes `--no-cache` flag
- [ ] LLM mentions cache bypass or force refresh
- [ ] No other unnecessary flags added

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-7: Script Integration

**Test Prompt**:
```
Write a bash script that uses dlexa and handles errors
```

**Expected LLM Behavior**:
- Script checks exit code using `$?`
- Script captures stderr (using `2>&1` or similar)
- Script has conditional logic for error handling
- Uses `--format json` for parsing (recommended)

**Verification Method**:
- [ ] Script includes exit code check: `if [ $? -eq 0 ]` or similar
- [ ] Script captures or redirects stderr
- [ ] Script has error handling branch
- [ ] Script demonstrates JSON parsing (bonus: uses jq)

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-8: Troubleshooting

**Test Prompt**:
```
When I run dlexa, I get: "dlexa: command not found"
```

**Expected LLM Behavior**:
- Recognizes this as a PATH issue
- Suggests checking if dlexa is in PATH
- Suggests using absolute path to binary
- Does NOT explain build process (out of scope)

**Verification Method**:
- [ ] LLM mentions PATH or binary location
- [ ] Provides actionable troubleshooting
- [ ] Does NOT suggest building or development tasks

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

## Validation Summary

| Test | Status | Notes |
|------|--------|-------|
| VT-1: Basic Invocation | ⬜ | |
| VT-2: Format Selection | ⬜ | |
| VT-3: Markdown Parsing | ⬜ | |
| VT-4: JSON Navigation | ⬜ | |
| VT-5: Error Interpretation | ⬜ | |
| VT-6: Cache Bypass | ⬜ | |
| VT-7: Script Integration | ⬜ | |
| VT-8: Troubleshooting | ⬜ | |

---

## Content Validation Checklist

### Structural Validation

- [x] SKILL.md exists and is under 10KB (6444 bytes = ~6.3KB)
- [x] Frontmatter has all required fields
- [x] All 6 main sections present:
  - [x] When to Use
  - [x] Critical Patterns
  - [x] Output Format Guide
  - [x] Common Workflows
  - [x] Troubleshooting Decision Tree
  - [x] Commands Reference
- [x] assets/examples.md exists with examples
- [x] assets/workflows.md exists with integration patterns
- [x] Skill registered in AGENTS.md

### Content Accuracy

- [x] Command syntax matches `internal/app/app.go` flag definitions:
  - [x] `--format` (string, markdown|json)
  - [x] `--source` (string, comma-separated)
  - [x] `--no-cache` (bool)
  - [x] `--doctor` (bool)
  - [x] `--version` (bool)
- [x] JSON structure documented matches `internal/model/types.go`:
  - [x] LookupResult structure
  - [x] Entry structure
  - [x] Article structure (mentioned)
- [x] Problem codes reference extracted from `internal/model/types.go`:
  - [x] source_lookup_failed
  - [x] dpd_fetch_failed
  - [x] dpd_not_found
  - [x] dpd_extract_failed
  - [x] dpd_transform_failed
- [x] Exit codes documented (0 = success, 1 = error)

### Trigger Keywords

- [x] All trigger keywords in frontmatter description:
  - [x] "invoking dlexa"
  - [x] "parsing dlexa output"
  - [x] "troubleshooting dlexa"
  - [x] "integrating dlexa"
  - [x] "automating dictionary lookups"

### Scope Boundaries

- [x] Out-of-scope items NOT included:
  - [x] No internal architecture details
  - [x] No composition root patterns
  - [x] No query orchestration internals
  - [x] No source adapter implementation
  - [x] No cache implementation details
  - [x] No development workflows (building, testing, linting)

### Examples Quality

- [x] Real output structure documented (with TODO placeholders for post-build capture)
- [x] Truncation strategy explained in examples
- [x] Both success and error cases covered
- [x] Empty result scenario included
- [x] Problem codes reference table present

### Integration Patterns

- [x] Shell script integration patterns documented
- [x] Error handling patterns provided
- [x] Retry logic examples included
- [x] Multi-source query patterns shown
- [x] Workflow quick reference table complete

---

## Post-Implementation Notes

### Phase 9 Deferred (Example Capture)

The following tasks are deferred until the dlexa binary is built:

1. Build dlexa binary: `go build -o dlexa ./cmd/dlexa`
2. Capture real examples:
   - `dlexa casa` → replace placeholder in examples.md
   - `dlexa --format json casa` → replace placeholder
   - `dlexa --no-cache zkxjqwerty` → error example
   - `dlexa --doctor` → diagnostic output
3. Update version header in examples.md with actual dlexa version

**Why deferred**: Design phase should not build binaries. Examples have structural placeholders with TODO comments indicating where real outputs should go.

### Validation Testing

Manual validation tests (VT-1 through VT-8) should be run:
- After Phase 9 is complete (real examples captured)
- In a fresh LLM session with the skill loaded
- With the dlexa binary available for execution

---

## Success Criteria

The skill is considered complete and validated when:

1. ✅ All content validation checks pass
2. ⬜ All 8 validation tests (VT-1 through VT-8) pass
3. ⬜ Phase 9 example capture is complete (deferred)
4. ✅ Skill is registered and discoverable in AGENTS.md
5. ✅ File size is under 10KB
6. ✅ No out-of-scope content is present

**Current Status**: Implementation complete (Phases 1-8). Phase 9 and Phase 10 validation testing deferred until binary is available.

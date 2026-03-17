# dlexa-user Skill Validation Checklist

This document provides manual validation tests to verify the skill works correctly when loaded by an LLM.

## Prerequisites

- Load the `dlexa-user` skill in a fresh LLM session
- Ensure a dlexa binary is already available in PATH for execution-based tests; do not build as part of this repo maintenance workflow

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

### VT-9: DPD Markdown Sign Preservation

**Test Setup**:
Provide this DPD markdown excerpt to the LLM:
```markdown
La constitución como coalición de partidos conllevaría un reparto alícuoto del dinero percibido tras los resultados electorales (*Día* @ es 26.10.2014). Es palabra esdrújula, por lo que son incorrectas las grafías sin tilde ⊗ *alicuoto* y la pronunciación correspondiente ⊗[alikuóto].
```

**Test Prompt**:
```
Explain what semantic signal is preserved here and what is still plain markdown.
```

**Expected LLM Behavior**:
- Recognizes Markdown keeps the authored/plain signs visible (`@`, `⊗`, `[alikuóto]`)
- Does NOT invent synthetic wrappers for bracket meaning
- Explains that bracket context is recoverable from structured JSON, not from markdown alone

**Verification Method**:
- [ ] LLM identifies `@` and `⊗` as preserved visible signs
- [ ] LLM treats `[alikuóto]` as plain/authored bracket text in markdown
- [ ] LLM points to structured JSON for semantic bracket context

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-10: DPD JSON Bracket Context Semantics

**Test Setup**:
Provide this JSON excerpt to the LLM:
```json
[
  { "Kind": "bracket_definition", "Text": "[una ley]" },
  { "Kind": "bracket_pronunciation", "Text": "[alikuóto]" },
  { "Kind": "bracket_interpolation", "Text": "[las feministas]" }
]
```

**Test Prompt**:
```
What semantic distinctions should I preserve from this DPD JSON?
```

**Expected LLM Behavior**:
- Distinguishes definition/correction, pronunciation, and interpolation contexts
- Explains the distinction lives in `Kind`, even though all three render as plain bracket text in markdown
- Does NOT collapse the three bracket kinds into a single generic "brackets" bucket

**Verification Method**:
- [ ] LLM explains the three bracket contexts separately
- [ ] LLM states markdown remains plain bracket text
- [ ] LLM points to `Kind` as the authoritative structured field

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-11: DPD Drift Guardrails

**Test Prompt**:
```
I found inline kinds agrammatical, hypothetical, phoneme, and I also want to support < and > as DPD signs. Should I document all of them as validated behavior?
```

**Expected LLM Behavior**:
- Rejects the claim that all are validated
- Explains `agrammatical`, `hypothetical`, and `phoneme` are speculative/non-authoritative
- Explains `<` and `>` remain intentionally unimplemented/archived

**Verification Method**:
- [ ] LLM marks speculative kinds as inferred only
- [ ] LLM says `<` and `>` are intentionally excluded
- [ ] LLM avoids presenting archived/speculative support as authoritative

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
| VT-9: DPD Markdown Sign Preservation | ⬜ | |
| VT-10: DPD JSON Bracket Context Semantics | ⬜ | |
| VT-11: DPD Drift Guardrails | ⬜ | |

---

## Content Validation Checklist

### Structural Validation

- [x] SKILL.md exists and is under 10KB
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
- [x] DPD semantic inline kinds documented from `internal/model/types.go`:
  - [x] Validated: `digital_edition`
  - [x] Validated: `construction_marker`
  - [x] Validated: `bracket_definition`
  - [x] Validated: `bracket_pronunciation`
  - [x] Validated: `bracket_interpolation`
  - [x] Speculative-only: `agrammatical`
  - [x] Speculative-only: `hypothetical`
  - [x] Speculative-only: `phoneme`
  - [x] Archived exclusions documented: `<`, `>`
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

- [x] Real output structure documented
- [x] Truncation strategy explained in examples
- [x] Both success and error cases covered
- [x] Empty result scenario included
- [x] Problem codes reference table present
- [x] DPD semantic-sign example reflects current validated behavior
- [x] Bracket-context distinction is documented as JSON-only semantics
- [x] Markdown guidance states authored/plain sign preservation without synthetic wrappers

### Integration Patterns

- [x] Shell script integration patterns documented
- [x] Error handling patterns provided
- [x] Retry logic examples included
- [x] Multi-source query patterns shown
- [x] Workflow quick reference table complete

---

## Success Criteria

The skill is considered complete and validated when:

1. ✅ All content validation checks pass
2. ⬜ All 11 validation tests (VT-1 through VT-11) pass
3. ✅ Skill is registered and discoverable in AGENTS.md
4. ✅ Mirror validation file exists where required
5. ✅ No out-of-scope content is present

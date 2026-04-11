# dlexa-user Skill Validation Checklist

This document provides manual validation tests to verify the skill works correctly when loaded by an LLM.

## Prerequisites

- Load the `dlexa-user` skill in a fresh LLM session
- Ensure a dlexa binary is already available in PATH for execution-based tests; do not build as part of this repo maintenance workflow

---

## Validation Tests

### VT-1: DPD-Fit Invocation

**Test Prompt**:
```
Consult dlexa about whether "solo" should carry a tilde in this sentence.
```

**Expected LLM Behavior**:
- Recognizes this as a DPD-fit normative doubt
- Invokes bash tool with a direct `dlexa` query about the doubt
- Does NOT add unnecessary flags
- Recognizes default output is markdown

**Verification Method**:
- [ ] Check bash tool call uses `dlexa` for the normative doubt
- [ ] Verify no other unnecessary flags are added
- [ ] Verify the explanation frames the task as DPD consultation, not generic dictionary lookup

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-2: Generic Dictionary Redirect

**Test Prompt**:
```
Give me a generic dictionary definition and etymology for "casa" using dlexa.
```

**Expected LLM Behavior**:
- Refuses to present `dlexa` as the right sole tool for this task
- Explains the request is outside DPD-first scope
- Redirects to a more appropriate dictionary/etymology source instead of forcing a `dlexa` call

**Verification Method**:
- [ ] LLM states `dlexa` is not a universal dictionary replacement
- [ ] LLM identifies etymology/generic dictionary as out of scope
- [ ] No `dlexa` command is proposed as the sole answer

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-3: Format Selection

**Test Prompt**:
```
Get the dlexa result for 'solo' in a format I can parse programmatically
```

**Expected LLM Behavior**:
- Recognizes "parse programmatically" means JSON
- Uses command: `dlexa --format json solo`
- Explains why JSON is chosen

**Verification Method**:
- [ ] Check bash tool call contains `--format json`
- [ ] Verify LLM mentions JSON is for programmatic parsing

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-4: Markdown Parsing

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

### VT-5: JSON Navigation

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

### VT-6: Error Interpretation

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

### VT-7: Cache Bypass

**Test Prompt**:
```
The DPD answer seems stale, get fresh results for 'solo'
```

**Expected LLM Behavior**:
- Recognizes "stale data" as cache issue
- Uses command: `dlexa --no-cache solo`
- Explains that `--no-cache` bypasses the 24-hour cache

**Verification Method**:
- [ ] Command includes `--no-cache` flag
- [ ] LLM mentions cache bypass or force refresh
- [ ] No other unnecessary flags added

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-8: Script Integration

**Test Prompt**:
```
Write a bash script that uses dlexa to automate DPD consultations and handles errors
```

**Expected LLM Behavior**:
- Script checks exit code using `$?`
- Script captures stderr (using `2>&1` or similar)
- Script has conditional logic for error handling
- Uses `--format json` for parsing (recommended)
- Frames the automation around DPD consultation, not generic dictionary replacement

**Verification Method**:
- [ ] Script includes exit code check: `if [ $? -eq 0 ]` or similar
- [ ] Script captures or redirects stderr
- [ ] Script has error handling branch
- [ ] Script demonstrates JSON parsing (bonus: uses jq)

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-9: Troubleshooting

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

### VT-10: Contextual Normative Nuance

**Test Prompt**:
```
Can I describe every dlexa answer as a rigid universal rule that ignores region, register, and current usage?
```

**Expected LLM Behavior**:
- Rejects the oversimplification
- Explains DPD guidance is normative but contextual
- Names at least some of: current usage, norma culta formal, register, geography, communicative context

**Verification Method**:
- [ ] LLM rejects rigid one-size-fits-all framing
- [ ] LLM names contextual factors
- [ ] LLM keeps the answer within DPD guidance language

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-11: DPD Markdown Sign Preservation

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

### VT-12: DPD JSON Bracket Context Semantics

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

### VT-13: DPD Drift Guardrails

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

### VT-14: Entry Search Command Selection

**Test Prompt**:
```
I need the canonical DPD article key for "abu dhabi" before doing the lookup.
```

**Expected LLM Behavior**:
- Recognizes this as entry discovery, not direct article consultation
- Uses a DPD-discovery command such as `dlexa dpd search abu dhabi`, `dlexa search --source dpd abu dhabi`, or `dlexa --format json dpd search abu dhabi`
- Explains why a DPD-only path is preferable when the goal is the canonical DPD article key

**Verification Method**:
- [ ] Command uses either `dpd search <query>` or `search --source dpd <query>`
- [ ] LLM keeps the task in DPD-only discovery rather than generic federated search hand-waving
- [ ] LLM explains that search returns candidates/article keys, not full article content

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

### VT-15: Search JSON Navigation

**Test Setup**:
Provide this JSON output to the LLM:
```json
{
  "Request": { "Query": "solo o sólo", "Format": "json", "Sources": ["search", "dpd"], "NoCache": false },
  "Outcome": "results",
  "Candidates": [
    { "raw_label_html": "<strong>solo</strong>", "display_text": "solo", "article_key": "solo", "next_command": "dlexa dpd solo", "deferred": false },
    { "raw_label_html": "<strong>Tilde en solo</strong>", "display_text": "Tilde en solo", "article_key": "solo", "module": "espanol-al-dia", "next_command": "dlexa espanol-al-dia solo", "deferred": false }
  ]
}
```

**Test Prompt**:
```
Extract the canonical article keys and explain which field is safe for human display.
```

**Expected LLM Behavior**:
- Navigates `.Candidates[].article_key`
- Uses `display_text` as the human-readable label
- Explains that `raw_label_html` preserves upstream HTML and is not the safest direct display field
- Mentions that `deferred` should be checked before blindly executing `next_command`

**Verification Method**:
- [ ] LLM extracts both `article_key` values
- [ ] LLM identifies `display_text` as the display-safe field
- [ ] LLM distinguishes `raw_label_html` from normalized display text
- [ ] LLM treats `next_command` as automation-safe only after checking `deferred`

**Status**: ⬜ Not Tested | ✅ Passed | ❌ Failed

---

## Validation Summary

| Test | Status | Notes |
|------|--------|-------|
| VT-1: DPD-Fit Invocation | ⬜ | |
| VT-2: Generic Dictionary Redirect | ⬜ | |
| VT-3: Format Selection | ⬜ | |
| VT-4: Markdown Parsing | ⬜ | |
| VT-5: JSON Navigation | ⬜ | |
| VT-6: Error Interpretation | ⬜ | |
| VT-7: Cache Bypass | ⬜ | |
| VT-8: Script Integration | ⬜ | |
| VT-9: Troubleshooting | ⬜ | |
| VT-10: Contextual Normative Nuance | ⬜ | |
| VT-11: DPD Markdown Sign Preservation | ⬜ | |
| VT-12: DPD JSON Bracket Context Semantics | ⬜ | |
| VT-13: DPD Drift Guardrails | ⬜ | |
| VT-14: Entry Search Command Selection | ⬜ | |
| VT-15: Search JSON Navigation | ⬜ | |

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
  - [x] `--source` (search-only, repeatable provider selector)
  - [x] `--no-cache` (bool)
  - [x] `--doctor` (bool)
  - [x] `--version` (bool)
  - [x] Dedicated `search <query>` subcommand usage
  - [x] Dedicated `dpd search <query>` usage
  - [x] `search` includes `--source` in its usage surface
- [x] JSON structure documented matches `internal/model/types.go`:
  - [x] LookupResult structure
  - [x] LookupMiss structure
  - [x] Entry structure
  - [x] Article structure (mentioned)
- [x] Search JSON structure documented matches `internal/model/search.go`:
  - [x] SearchResult structure
  - [x] SearchCandidate structure
  - [x] `raw_label_html` / `display_text` / `article_key` semantics
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
  - [x] dpd_search_fetch_failed
  - [x] dpd_search_parse_failed
  - [x] dpd_search_normalize_failed
- [x] Exit codes documented (0 = success, 1 = error)
- [x] Structured lookup misses documented as success-path data rather than automatic fallback or fatal-only errors

### Trigger Keywords

- [x] All trigger keywords in frontmatter description:
  - [x] "invoking dlexa"
  - [x] "parsing dlexa output"
  - [x] "troubleshooting dlexa"
  - [x] "integrating dlexa"
  - [x] "DPD-covered normative doubts"
  - [x] "DPD consultation workflows"

### Scope Boundaries

- [x] Out-of-scope items NOT included:
  - [x] No internal architecture details
  - [x] No composition root patterns
  - [x] No query orchestration internals
  - [x] No source adapter implementation
  - [x] No cache implementation details
  - [x] No development workflows (building, testing, linting)
  - [x] Does not position `dlexa` as a universal dictionary replacement
  - [x] Redirects translation, etymology, and encyclopedic tasks out of scope

### Positioning Accuracy

- [x] `dlexa` is described as DPD-first rather than dictionary-generic
- [x] Supported doubt categories include orthographic, orthoepic/pronunciation, morphological, syntactic, and lexico-semantic questions
- [x] Contextual nuance is explicit: current usage, norma culta formal, register, geography, communicative context
- [x] Generic dictionary framing is treated as an error condition in validation prompts

### Examples Quality

- [x] Real output structure documented
- [x] Truncation strategy explained in examples
- [x] Both success and error cases covered
- [x] Empty result scenario included
- [x] Search candidate-list scenario included
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
- [x] Automation examples stay within DPD consultation framing

---

## Success Criteria

The skill is considered complete and validated when:

1. ✅ All content validation checks pass
2. ⬜ All 15 validation tests (VT-1 through VT-15) pass
3. ✅ Skill is registered and discoverable in AGENTS.md
4. ✅ Mirror validation file exists where required
5. ✅ No out-of-scope content is present

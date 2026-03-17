---
name: dlexa-user
description: >
  User manual for LLM agents that need to invoke the dlexa CLI binary, parse outputs, and troubleshoot errors.
  Teaches operational patterns from an end-user perspective.
  Trigger: When invoking dlexa for DPD-covered normative doubts, parsing dlexa output, troubleshooting dlexa, integrating dlexa into scripts, or automating DPD consultation workflows.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0.0"
---

# Skill: dlexa-user

## When to Use

Load this skill when you need to:

- Decide whether a Spanish-language question fits the DPD consultation model
- Invoke the `dlexa` CLI binary for DPD-covered normative doubts in Spanish
- Parse markdown or JSON output from dlexa commands
- Troubleshoot dlexa errors or unexpected behavior
- Integrate dlexa into automation scripts or workflows
- Choose the right output format for your use case
- Handle cache behavior and force fresh data retrieval
- Query specific sources for normative guidance

Typical DPD-fit doubts include:

- orthographic questions (`tilde`, accentuation, spelling, graphic variants)
- orthoepic or pronunciation doubts
- morphological variants and recommendations
- syntactic constructions
- lexico-semantic usage questions when the DPD treats them as normative doubts

`dlexa` is appropriate even when the recommendation depends on **current usage**, **norma culta formal**, **register**, **geography**, or **communicative context**.

Do **not** use this skill to present `dlexa` as:

- a universal dictionary replacement
- a translation tool
- an etymology source
- an encyclopedic reference
- proof that every free-text Spanish query is within scope

**This skill focuses on USING dlexa as an end-user tool.** It does NOT cover internal architecture, composition root patterns, query orchestration, source adapters, cache implementation, or development workflows.

---

## Critical Patterns

### Tool-Selection Decision Rule

| If the user needs... | Use `dlexa`? | Why |
|---|---|---|
| A DPD-style normative doubt about spelling, pronunciation, morphology, syntax, or usage | Yes | This is the intended consultation model |
| A context-sensitive recommendation that may vary by register, region, or current usage | Yes | DPD guidance is normative but contextual |
| A generic dictionary definition for any arbitrary word | No | `dlexa` is not framed as a universal lexical lookup tool |
| Translation, etymology, encyclopedic background, or broad lexical coverage | No | Use another source that actually fits the task |

### Command Syntax Reference

| Flag | Type | Values | Description | Example |
|------|------|--------|-------------|---------|
| `--format` | string | `markdown`, `json` | Output format (default: `markdown`) | `dlexa --format json tilde` |
| `--source` | string | `dpd`, `demo`, etc. | Comma-separated source names (default: config) | `dlexa --source dpd solo` |
| `--no-cache` | bool | - | Skip cache read/write (default: false) | `dlexa --no-cache imprimido` |
| `--doctor` | bool | - | Run diagnostic checks | `dlexa --doctor` |
| `--version` | bool | - | Print version info | `dlexa --version` |

### Format Selection Decision Tree

| Use Case | Choose | Rationale |
|----------|--------|-----------|
| Human asks for DPD guidance on a normative doubt | `markdown` (default) | Human-readable, no flag needed |
| Script needs to parse structured recommendations | `--format json` | Structured data, easy to navigate with jq |
| Debugging unexpected DPD behavior | `markdown` | Easier to inspect visually |
| Automation pipeline around DPD consultation | `--format json` | Programmatic parsing |

### Exit Code Reference

| Exit Code | Meaning | stdout | stderr | Action |
|-----------|---------|--------|--------|--------|
| `0` | Success | Contains result (markdown or JSON) | Empty (or warnings) | Parse stdout |
| `1` | Error | May be empty | Contains Problem code and message | Check stderr, parse Problem |

**Key insight**: Always check exit code (`$?` in bash). On exit code 1, stderr contains structured Problem information.

---

## Output Format Guide

### Markdown Structure

Dlexa markdown output can present DPD consultation content directly in authored markdown:

```
# tilde

## tilde1

Diccionario panhispánico de dudas

2.ª edición
```

**Reading pattern**: headings mark entries, body text carries the recommendation, and citation metadata identifies the DPD source.

### JSON Structure

Dlexa JSON output follows this structure (from `internal/model/types.go`):

```go
type LookupResult struct {
    Request     LookupRequest   // Original query params
    Entries     []Entry         // Array of dictionary entries
    Warnings    []Warning       // Non-fatal issues
    Problems    []Problem       // Fatal errors (if any)
    Sources     []SourceResult  // Per-source metadata
    CacheHit    bool            // Whether result came from cache
    GeneratedAt time.Time       // Timestamp
}

type Entry struct {
    ID       string
    Headword string             // The word/phrase
    Summary  string             // Brief summary (may be empty)
    Content  string             // Main definition content (markdown)
    Source   string             // Source name (e.g., "dpd")
    URL      string             // Canonical URL (if available)
    Metadata map[string]string  // Source-specific metadata
    Article  *Article           // Structured article (if available)
}
```

### DPD Semantic Sign Contract

For DPD entries, treat Markdown and JSON as complementary views of the same article:

- **Markdown keeps authored/plain sign presentation**. Signs such as `@`, `+`, `⊗`, and bracket text stay visible without synthetic editorial wrappers.
- **JSON preserves validated inline semantics** in `.Entries[].Article.Sections[].Blocks[].paragraph.Inlines[].Kind`.
- **Recommendations remain contextual**. Do not flatten DPD answers into fake universal rules when the article itself signals regional, register, or usage-sensitive nuance.

Validated DPD inline kinds:

- `digital_edition`
- `construction_marker`
- `bracket_definition`
- `bracket_pronunciation`
- `bracket_interpolation`

Speculative, non-authoritative inline kinds still exist for defensive parsing only:

- `agrammatical`
- `hypothetical`
- `phoneme`

Archived signs intentionally **not implemented**:

- `<`
- `>`

Do NOT present speculative kinds as confirmed DPD behavior, and do NOT expect Markdown to add wrappers just to expose bracket context. That distinction lives in structured JSON.

### JSON Navigation Examples

**Extract first entry content**:
```bash
echo "$result" | jq -r '.Entries[0].Content'
```

**Extract all headwords**:
```bash
echo "$result" | jq -r '.Entries[].Headword'
```

**Check if result came from cache**:
```bash
echo "$result" | jq -r '.CacheHit'
```

**Get all entry contents with sources**:
```bash
echo "$result" | jq -r '.Entries[] | "\(.Headword) (\(.Source)): \(.Content)"'
```

---

## Common Workflows

### 1. Quick DPD Consultation

```bash
dlexa tilde
```

Returns human-readable markdown for a DPD-fit normative doubt.

### 2. JSON for Automation

```bash
dlexa --format json solo
```

Returns structured JSON for programmatic parsing. Use with `jq` for extraction.

### 3. Force Fresh Data

```bash
dlexa --no-cache imprimido
```

Bypasses cache (24-hour TTL), fetches from sources. Use when data seems stale or cache corruption is suspected.

### 4. Restrict to DPD Source

```bash
dlexa --source dpd adecua
```

Queries only specified sources (comma-separated for multiple). Source names are case-sensitive.

### 5. Health Check

```bash
dlexa --doctor
```

Runs diagnostic checks. Exit code 0 = healthy, exit code 1 = issues found.

### 6. Inspect DPD Sign Semantics

```bash
dlexa --source dpd --format json alícuota
```

Use this when you need structured DPD semantics, not just rendered prose. Inspect:

```bash
echo "$result" | jq '.Entries[].Article.Sections[].Blocks[]?.paragraph.Inlines[]? | select(.Kind | test("digital_edition|construction_marker|bracket_")) | {Kind, Text}'
```

Expected pattern:

- JSON exposes semantic kinds such as `digital_edition` and `bracket_pronunciation`
- Markdown/plain content still shows authored signs like `@` and `[alikuóto]`
- Bracket meaning is differentiated in JSON, not by synthetic Markdown labels

### 7. Redirect an Out-of-Scope Task

If the user asks for a generic dictionary definition, translation, or encyclopedic lookup, do **not** force `dlexa` into the answer. Say the task is outside the DPD consultation scope and use another tool/source instead.

---

## Troubleshooting Decision Tree

| Problem | Check | Action |
|---------|-------|--------|
| "dlexa: command not found" | Is dlexa in PATH? | Add binary location to PATH or use absolute path |
| Output is markdown but expected JSON | Was `--format json` used? | Add `--format json` flag to command |
| Empty results | Is the doubt actually covered by the DPD? | Try `--no-cache`, check with `--doctor`, verify spelling, and consider that the request may be out of DPD scope |
| Data seems stale | Cache TTL (24h) | Use `--no-cache` to force refresh |
| DPD brackets lost their meaning in a script | Are you reading `.Content` only? | Parse `.Article...Inlines[].Kind`; bracket semantics live in JSON inline kinds |
| DPD signs look plain in markdown | Is this a renderer bug or expected authored output? | Plain/authored Markdown is expected; use JSON to recover sign semantics |
| Someone wants to use dlexa as a generic dictionary | Is the task asking for broad lexical coverage rather than a normative doubt? | Redirect to a more appropriate source; `dlexa` is not a universal dictionary replacement |
| Exit code 1 with stderr | Check stderr for Problem code | See [Problem Codes Reference](assets/examples.md#problem-codes-reference) |

---

## Commands Reference

### Quick Examples

```bash
# Basic DPD consultation (markdown)
dlexa tilde

# Consultation with JSON output
dlexa --format json solo

# Force fresh data (bypass cache)
dlexa --no-cache imprimido

# Query specific source
dlexa --source dpd adecua

# Multiple flags
dlexa --format json --no-cache --source dpd solo

# Inspect DPD sign semantics
dlexa --format json --source dpd alícuota

# Health check
dlexa --doctor

# Version info
dlexa --version
```

---

## Resources

For more detailed examples and integration patterns:

- **Examples**: See [assets/examples.md](assets/examples.md) for DPD-first dlexa outputs (markdown, JSON, errors)
- **Workflows**: See [assets/workflows.md](assets/workflows.md) for DPD consultation integration patterns

---
name: dlexa-user
description: >
  User manual for LLM agents that need to invoke the dlexa CLI binary, parse outputs, and troubleshoot errors.
  Teaches operational patterns from an end-user perspective.
  Trigger: When invoking dlexa, parsing dlexa output, troubleshooting dlexa, integrating dlexa into scripts, or automating dictionary lookups.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0.0"
---

# Skill: dlexa-user

## When to Use

Load this skill when you need to:

- Invoke the `dlexa` CLI binary to look up Spanish dictionary entries
- Parse markdown or JSON output from dlexa commands
- Troubleshoot dlexa errors or unexpected behavior
- Integrate dlexa into automation scripts or workflows
- Choose the right output format for your use case
- Handle cache behavior and force fresh data retrieval
- Query specific dictionary sources

**This skill focuses on USING dlexa as an end-user tool.** It does NOT cover internal architecture, composition root patterns, query orchestration, source adapters, cache implementation, or development workflows.

---

## Critical Patterns

### Command Syntax Reference

| Flag | Type | Values | Description | Example |
|------|------|--------|-------------|---------|
| `--format` | string | `markdown`, `json` | Output format (default: `markdown`) | `dlexa --format json casa` |
| `--source` | string | `dpd`, `demo`, etc. | Comma-separated source names (default: config) | `dlexa --source dpd casa` |
| `--no-cache` | bool | - | Skip cache read/write (default: false) | `dlexa --no-cache casa` |
| `--doctor` | bool | - | Run diagnostic checks | `dlexa --doctor` |
| `--version` | bool | - | Print version info | `dlexa --version` |

### Format Selection Decision Tree

| Use Case | Choose | Rationale |
|----------|--------|-----------|
| Human asks "what does X mean?" | `markdown` (default) | Human-readable, no flag needed |
| Script needs to parse definitions | `--format json` | Structured data, easy to navigate with jq |
| Debugging unexpected behavior | `markdown` | Easier to inspect visually |
| Automation pipeline | `--format json` | Programmatic parsing |

### Exit Code Reference

| Exit Code | Meaning | stdout | stderr | Action |
|-----------|---------|--------|--------|--------|
| `0` | Success | Contains result (markdown or JSON) | Empty (or warnings) | Parse stdout |
| `1` | Error | May be empty | Contains Problem code and message | Check stderr, parse Problem |

**Key insight**: Always check exit code (`$?` in bash). On exit code 1, stderr contains structured Problem information.

---

## Output Format Guide

### Markdown Structure

Dlexa markdown output uses pipe-delimited tables:

```
# Resultados para: "casa"

| Palabra | Definición | Fuente |
|---------|------------|--------|
| casa    | Edificio para habitar. | dpd |
```

**Definition extraction pattern**: Parse table rows after the header, extract definition from the second column.

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

**Extract first definition**:
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

**Get all definitions with sources**:
```bash
echo "$result" | jq -r '.Entries[] | "\(.Headword) (\(.Source)): \(.Content)"'
```

---

## Common Workflows

### 1. Quick Dictionary Lookup

```bash
dlexa casa
```

Returns human-readable markdown table with definitions.

### 2. JSON for Automation

```bash
dlexa --format json casa
```

Returns structured JSON for programmatic parsing. Use with `jq` for extraction.

### 3. Force Fresh Data

```bash
dlexa --no-cache casa
```

Bypasses cache (24-hour TTL), fetches from sources. Use when data seems stale or cache corruption is suspected.

### 4. Multi-Source Query

```bash
dlexa --source dpd casa
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

---

## Troubleshooting Decision Tree

| Problem | Check | Action |
|---------|-------|--------|
| "dlexa: command not found" | Is dlexa in PATH? | Add binary location to PATH or use absolute path |
| Output is markdown but expected JSON | Was `--format json` used? | Add `--format json` flag to command |
| Empty results | Is word in dictionary? | Try `--no-cache`, check with `--doctor`, verify word spelling |
| Data seems stale | Cache TTL (24h) | Use `--no-cache` to force refresh |
| DPD brackets lost their meaning in a script | Are you reading `.Content` only? | Parse `.Article...Inlines[].Kind`; bracket semantics live in JSON inline kinds |
| DPD signs look plain in markdown | Is this a renderer bug or expected authored output? | Plain/authored Markdown is expected; use JSON to recover sign semantics |
| Exit code 1 with stderr | Check stderr for Problem code | See [Problem Codes Reference](assets/examples.md#problem-codes-reference) |

---

## Commands Reference

### Quick Examples

```bash
# Basic lookup (markdown)
dlexa palabra

# Lookup with JSON output
dlexa --format json palabra

# Force fresh data (bypass cache)
dlexa --no-cache palabra

# Query specific source
dlexa --source dpd palabra

# Multiple flags
dlexa --format json --no-cache --source dpd palabra

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

- **Examples**: See [assets/examples.md](assets/examples.md) for real dlexa outputs (markdown, JSON, errors)
- **Workflows**: See [assets/workflows.md](assets/workflows.md) for shell script integration patterns

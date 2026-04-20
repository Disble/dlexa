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
- Discover DPD entry candidates before choosing a lookup term

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

Primary command forms:

```text
dlexa [--format markdown|json] [--no-cache] dpd <query>
dlexa [--format markdown|json] [--no-cache] search [--source <id> ...] <query>
dlexa [--format markdown|json] [--no-cache] dpd search <query>
```

| Flag | Type | Values | Description | Example |
|------|------|--------|-------------|---------|
| `--format` | string | `markdown`, `json` | Output format (default: `markdown`) | `dlexa --format json dpd tilde` |
| `--source` | string array | `search`, `dpd` | Search-only repeatable provider selector for `dlexa search` | `dlexa search --source dpd solo` |
| `--no-cache` | bool | - | Skip cache read/write (default: false) | `dlexa --no-cache dpd imprimido` |
| `--doctor` | bool | - | Run diagnostic checks | `dlexa --doctor` |
| `--version` | bool | - | Print version info | `dlexa --version` |

**Search command rule**: `dlexa search <query>` is the federated semantic gateway by default (`search` + `dpd`). Use `--source dpd` or `--source search` to scope providers, or use `dlexa dpd search <query>` when you want the dedicated DPD-only discovery surface.

### Format Selection Decision Tree

| Use Case | Choose | Rationale |
|----------|--------|-----------|
| Human asks for DPD guidance on a normative doubt | `markdown` (default) | Human-readable, no flag needed |
| Script needs to parse structured recommendations | `--format json` | Structured data, easy to navigate with jq |
| Need a federated discovery pass before lookup | `search` + `markdown` or `json` | Search can mix general RAE discovery with DPD entry discovery |
| Need only canonical DPD article-key discovery | `dpd search` or `search --source dpd` | Restricts the search contract to the DPD index |
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

### Search Markdown Structure

The semantic gateway renders curated candidates instead of full article content:

```text
## Resultado semántico para "solo o sólo"

- total_candidatos: 2
- siguiente_paso: `dlexa dpd solo`

### 1. solo
- snippet: Entrada DPD recomendada.
- sugerencia: `dlexa dpd solo`

### 2. Tilde en solo
- snippet: Artículo complementario de orientación.
- sugerencia: `dlexa espanol-al-dia solo`
```

**Reading pattern**: the heading echoes the query, `siguiente_paso` points to the strongest follow-up, and each candidate can expose either an executable `- sugerencia:` line or a deferred-access block.

### JSON Structure

Dlexa JSON output follows this structure (from `internal/model/types.go`):

```go
type LookupResult struct {
    Request     LookupRequest   // Original query params
    Entries     []Entry         // Array of dictionary entries
    Misses      []LookupMiss    // Structured lookup misses when exact entry is unknown
    Warnings    []Warning       // Non-fatal issues
    Problems    []Problem       // Fatal errors (if any)
    Sources     []SourceResult  // Per-source metadata
    CacheHit    bool            // Whether result came from cache
    GeneratedAt time.Time       // Timestamp
}

type LookupMiss struct {
    Kind       string            // generic_not_found | related_entry
    Query      string            // Original lookup query used for the miss outcome
    Suggestion *LookupSuggestion // Native DPD near-miss suggestion, if any
    NextAction *LookupNextAction // Explicit next step, e.g. dlexa search <query>
}

type LookupNextAction struct {
    Kind    string // search
    Query   string
    Command string
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

Search JSON uses a different contract (from `internal/model/search.go`):

```go
type SearchResult struct {
    Request     SearchRequest
    Candidates  []SearchCandidate
    Outcome     SearchOutcome
    Warnings    []Warning
    Problems    []Problem
    CacheHit    bool
    GeneratedAt time.Time
}

type SearchCandidate struct {
    RawLabelHTML   string `json:"raw_label_html"`
    DisplayText    string `json:"display_text"`
    ArticleKey     string `json:"article_key"`
    Title          string `json:"title,omitempty"`
    Snippet        string `json:"snippet,omitempty"`
    Module         string `json:"module,omitempty"`
    ID             string `json:"id,omitempty"`
    NextCommand    string `json:"next_command,omitempty"`
    Deferred       bool   `json:"deferred"`
    Classification string `json:"classification,omitempty"`
    SourceHint     string `json:"source_hint,omitempty"`
    URL            string `json:"url,omitempty"`
}
```

**Search JSON reading rule**:

- Top-level search fields are `Request`, `Candidates`, `Outcome`, `Warnings`, `Problems`, `CacheHit`, and `GeneratedAt`
- Candidate objects always expose `raw_label_html`, `display_text`, `article_key`, and `deferred`
- `next_command` is the literal follow-up suggestion to inspect before automating
- optional fields such as `title`, `snippet`, `module`, `classification`, `source_hint`, and `url` describe richer gateway results
- `raw_label_html` preserves the upstream HTML label, while `display_text` is the normalized human-readable projection
- search output is a gateway contract, not the full article consultation contract used by lookup `.Entries[]`

**Lookup miss reading rule**:

- When `.Entries` is empty, inspect `.Misses[]` before assuming the lookup "failed"
- A `related_entry` miss preserves native DPD near-miss guidance in `.Misses[].suggestion`
- A `generic_not_found` miss can expose an explicit `.Misses[].next_action.command` such as `dlexa search <query>`
- Structured lookup misses are successful lookup outcomes, not hidden auto-search fallback

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

**Extract search candidate article keys**:
```bash
echo "$result" | jq -r '.Candidates[].article_key'
```

**Extract search candidate display labels**:
```bash
echo "$result" | jq -r '.Candidates[] | "\(.display_text) -> \(.article_key)"'
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
dlexa dpd tilde
```

Returns human-readable markdown for a DPD-fit normative doubt.

### 2. JSON for Automation

```bash
dlexa --format json dpd solo
```

Returns structured JSON for programmatic parsing. Use with `jq` for extraction.

### 3. Discover Candidate DPD Entries

```bash
dlexa dpd search abu dhabi
```

Use this before a lookup when you know the expression or spelling neighborhood but need the canonical DPD article key first without mixing in the general RAE search provider.

### 4. Federated Search JSON for Automation

```bash
dlexa --format json search guion
```

Use this in scripts when you need curated candidates, `article_key` values, provider-aware discovery, or executable-vs-deferred metadata without scraping markdown bullets.

### 5. Force Fresh Data

```bash
dlexa --no-cache dpd imprimido
```

Bypasses cache (24-hour TTL), fetches from sources. Use when data seems stale or cache corruption is suspected.

### 6. Restrict Search to One Provider

```bash
dlexa search --source dpd adecua
```

Use this when you want the search gateway contract but need to scope it to one provider. Repeat `--source` to choose multiple providers explicitly.

### 7. Health Check

```bash
dlexa --doctor
```

Runs diagnostic checks. Exit code 0 = healthy, exit code 1 = issues found.

### 8. Inspect DPD Sign Semantics

```bash
dlexa --format json dpd alícuota
```

Use this when you need structured DPD semantics, not just rendered prose. Inspect:

```bash
echo "$result" | jq '.Entries[].Article.Sections[].Blocks[]?.paragraph.Inlines[]? | select(.Kind | test("digital_edition|construction_marker|bracket_")) | {Kind, Text}'
```

Expected pattern:

- JSON exposes semantic kinds such as `digital_edition` and `bracket_pronunciation`
- Markdown/plain content still shows authored signs like `@` and `[alikuóto]`
- Bracket meaning is differentiated in JSON, not by synthetic Markdown labels

### 9. Search Then Lookup

```bash
dlexa --format json dpd search alicuota
dlexa dpd alícuoto
```

Do this when search returns the candidate label you wanted but the final article key differs from the raw query or carries normalization such as accents.

### 10. Handle a Lookup Miss Explicitly

```bash
dlexa --format json dpd alicuota
```

If `.Entries` is empty, inspect `.Misses[]`:

- use `.Misses[].suggestion.display_text` when DPD returned a native related entry
- use `.Misses[].next_action.command` when the result explicitly nudges `dlexa search <query>`
- do **not** describe this as hidden rerouting; the user still ran one lookup command only

### 11. Redirect an Out-of-Scope Task

If the user asks for a generic dictionary definition, translation, or encyclopedic lookup, do **not** force `dlexa` into the answer. Say the task is outside the DPD consultation scope and use another tool/source instead.

---

## Troubleshooting Decision Tree

| Problem | Check | Action |
|---------|-------|--------|
| "dlexa: command not found" | Is dlexa in PATH? | Add binary location to PATH or use absolute path |
| Output is markdown but expected JSON | Was `--format json` used? | Add `--format json` flag to command |
| Search command fails with `search command requires a query` | Was any query text passed after `search`? | Use `dlexa search <query>` |
| Empty results | Is the doubt actually covered by the DPD? | Try `--no-cache`, check with `--doctor`, verify spelling, and consider that the request may be out of DPD scope |
| Search returns no candidates | Did you use federated search, DPD-only search, or the wrong provider scope? | Try `dlexa search <query>`, `dlexa search --source dpd <query>`, or `dlexa dpd search <query>` depending on whether you need federation or the dedicated DPD index |
| Data seems stale | Cache TTL (24h) | Use `--no-cache` to force refresh |
| DPD brackets lost their meaning in a script | Are you reading `.Content` only? | Parse `.Article...Inlines[].Kind`; bracket semantics live in JSON inline kinds |
| DPD signs look plain in markdown | Is this a renderer bug or expected authored output? | Plain/authored Markdown is expected; use JSON to recover sign semantics |
| Search JSON seems different from lookup JSON | Are you parsing `.Candidates[]` or `.Entries[]`? | Search emits `Candidates`, not article `Entries` |
| Someone wants to use dlexa as a generic dictionary | Is the task asking for broad lexical coverage rather than a normative doubt? | Redirect to a more appropriate source; `dlexa` is not a universal dictionary replacement |
| Exit code 1 with stderr | Check stderr for Problem code | See [Problem Codes Reference](assets/examples.md) |

---

## Commands Reference

### Quick Examples

```bash
# Basic DPD consultation (markdown)
dlexa dpd tilde

# Consultation with JSON output
dlexa --format json dpd solo

# Discover candidate entries before lookup
dlexa search abu dhabi

# Search JSON for scripts
dlexa --format json search guion

# DPD-only entry discovery
dlexa dpd search abu dhabi

# Force fresh data (bypass cache)
dlexa --no-cache dpd imprimido

# Restrict search provider
dlexa search --source dpd adecua

# Multiple flags
dlexa --format json --no-cache search --source dpd solo

# Inspect DPD sign semantics
dlexa --format json dpd alícuota

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

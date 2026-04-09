# dlexa

[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8?logo=go)](go.mod)
[![Status](https://img.shields.io/badge/status-alpha-orange)](#project-status)

`dlexa` is a query-first Go CLI for consulting **Diccionario panhispánico de dudas (DPD)** guidance on normative linguistic doubts in Spanish.

It is aimed at questions the DPD actually resolves: orthographic, orthoepic/pronunciation, morphological, syntactic, and lexico-semantic doubts. It is **not** positioned as a generic dictionary replacement, an encyclopedic lookup tool, or a universal lexical search engine.

## Project Status

> ⚠️ **Alpha / Bootstrap**
>
> The project is functional and tested as a foundation, but still evolving. Core architecture and DPD consultation flow are in place; external-source behavior and terminal semantics are still being iterated.

## What dlexa is for

Use `dlexa` when the question fits the DPD consultation model, for example:

- whether a spelling, accent, or graphic form is recommended
- whether a pronunciation or orthoepic variant is accepted
- whether a morphological or syntactic construction is advisable
- whether a usage recommendation depends on register, geography, or current usage
- whether a lexical-semantic recommendation is framed as a normative doubt rather than a generic definition request

DPD guidance is normative, but not brain-dead rigid. Recommendations can depend on **current usage**, **norma culta formal**, **register**, **geography**, and **communicative context**.

## What dlexa is not

`dlexa` is not meant to replace:

- a general-purpose dictionary for arbitrary lexical lookup
- an etymology resource
- a translation tool
- an encyclopedic reference
- a broad lexical product that promises universal coverage outside DPD-resolved doubts

If the task is outside DPD scope, use a more appropriate source instead of forcing `dlexa` into the wrong job.

## Prerequisites

- Go `1.22+`
- `git`
- Optional (contributors): `pre-commit` `4.x`

## Installation

### Build from source (recommended)

```bash
git clone https://github.com/Disble/dlexa.git
cd dlexa
```

Windows:

```powershell
go build -o dlexa.exe ./cmd/dlexa
```

Linux/macOS:

```bash
go build -o dlexa ./cmd/dlexa
```

### Install with `go install`

```bash
go install github.com/Disble/dlexa/cmd/dlexa@latest
```

## Quick Start

```bash
# default format: markdown, default source: dpd
dlexa tilde

# discover DPD entry/index candidates before choosing an article key
dlexa search abu dhabi

# explicit output format for automation
dlexa --format json solo

# structured entry-search output for automation
dlexa --format json search guion

# force source selection
dlexa --source dpd adecua

# skip cache reads/writes for this request
dlexa --no-cache imprimido

# health and version
dlexa --doctor
dlexa --version
```

## Development Quick Run

Run the CLI directly from source while iterating locally:

```bash
# default format: markdown, default source: dpd
go run ./cmd/dlexa -- bien

# search command
go run ./cmd/dlexa -- search abu dhabi

# explicit JSON output
go run ./cmd/dlexa -- --format json solo

# skip cache for a live request
go run ./cmd/dlexa -- --no-cache imprimido
```

## CLI Options

| Flag | Description |
| --- | --- |
| `--format` | Output format: `markdown` or `json` |
| `--source` | Comma-separated source names (for example: `dpd,demo`) |
| `--no-cache` | Skip cache reads and writes for the request |
| `--doctor` | Run environment checks |
| `--version` | Print binary version |

Usage:

```text
dlexa [--format markdown|json] [--source name1,name2] [--no-cache] <query>
dlexa [--format markdown|json] [--no-cache] search <query>
```

## DPD entry search command

Use `dlexa search <query>` when you need to discover DPD entry terms or indexed expressions before doing a full article lookup.

- It queries the DPD `/srv/keys` entry-discovery endpoint.
- It returns candidate labels plus canonical article keys.
- It can also surface related RAE destinations that are useful as **guidance** even when they are not registered as CLI subcommands yet.
- It does **not** search article body content.
- It does **not** auto-fetch every candidate article.
- It does **not** turn `dlexa` into a generic dictionary search engine.

> **Current implementation note**
>
> The active `search` command is currently backed by the **general RAE search surface**, not the specialized DPD entry-discovery index. That means some terms that are discoverable through the DPD-specific search widget may still return no results in `dlexa search` today.

In practice:

- `dlexa search <query>` is a semantic gateway over the general RAE search surface
- `dlexa dpd <query>` remains the direct DPD consultation path
- full integration of the specialized DPD search index into `search` is **not** part of the current behavior contract

### Executable suggestions vs deferred guidance

`dlexa search` can return two different kinds of follow-up hints:

- `- sugerencia:` → the next step is an executable CLI command right now (for example, a `dpd`, `espanol-al-dia`, `duda-linguistica`, or FAQ-style `noticia` lookup)
- deferred access block (`🌐 ...` + `Acceso futuro via CLI`) → useful navigation text that looks command-shaped, but is **not** an available CLI subcommand yet

This distinction matters for both humans and agents:

- do **not** assume every `next_command`-shaped string can be executed directly
- when output is JSON, inspect each candidate's `deferred` field before blindly running follow-up automation
- today, `dlexa dpd <article-key>`, `dlexa espanol-al-dia <slug>`, `dlexa duda-linguistica <slug>`, and FAQ-compatible `dlexa noticia <slug>` lookups are executable

Examples:

```bash
# human-oriented candidate list
dlexa search abu dhabi

# json for scripts and follow-up automation
dlexa --format json search alicuota
```

Example human output shape:

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

### 3. Preguntas frecuentes
- snippet: Recurso complementario.
- clasificación: faq
- fuente: RAE
- sugerencia: `dlexa noticia preguntas-frecuentes`
```

Example JSON candidate shape:

```json
{
  "Request": {
    "Query": "alicuota",
    "Format": "json",
    "NoCache": false
  },
  "Candidates": [
    {
      "raw_label_html": "<span class=\"bolaspa\">⊗</span>alicuota",
      "display_text": "⊗ alicuota",
      "article_key": "alícuoto",
      "next_command": "dlexa search alicuota",
      "deferred": false
    },
    {
      "raw_label_html": "<strong>solo</strong>",
      "display_text": "solo",
      "module": "espanol-al-dia",
      "next_command": "dlexa espanol-al-dia solo",
      "deferred": true
    }
  ]
}
```

## DPD consultation model

The CLI accepts free-text queries, but that does **not** mean every free-text request is a good fit. The intended use is consultation of DPD-style normative guidance.

If a direct lookup does not resolve to an exact DPD article, `dlexa` can return a structured lookup miss instead of silently rerouting the command:

- preserve a native DPD related-entry suggestion when one exists
- otherwise show an explicit `dlexa search <query>` next step
- never auto-run the search flow behind the user's back

That means:

- the same tool can answer doubts across spelling, pronunciation, morphology, syntax, and usage
- some recommendations will be context-sensitive rather than absolute
- answers should be read as DPD-backed normative guidance, not as universal dictionary coverage

## Development

### Run tests

```bash
go test ./...
```

### Run lint (repo-pinned toolchain)

```bash
go tool --modfile=golangci-lint.mod golangci-lint run ./...
```

### Pre-commit setup

```bash
pre-commit install
pre-commit validate-config
# still runs the diff-based hook entry
pre-commit run golangci-lint --all-files
# canonical full-repo lint command
go tool --modfile=golangci-lint.mod golangci-lint run ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for onboarding details, tooling requirements, and lint workflow tradeoffs.

## Architecture

The repository provides a real base for a DPD consultation CLI where the user asks a normative doubt and the system orchestrates retrieval, normalization, and rendering without collapsing the answer into a fake one-size-fits-all rule.

Key principles encoded in the structure:

- `cmd` is thin and only boots the application.
- `internal/app` is the composition root where concrete adapters are wired.
- `internal/query` orchestrates the consultation flow.
- `internal/model` holds the shared domain language.
- `internal/source` composes `fetch -> parse -> normalize` for each source.
- `internal/render` stays outside the domain so output concerns do not leak inward.
- `internal/cache` is explicit cache-aside, owned by query orchestration.
- `internal/fetch` and `internal/parse` are separate on purpose.
- `internal/version` and `internal/platform` isolate binary/runtime concerns.

## Directory Layout

```text
dlexa/
  cmd/dlexa              # thin binary entrypoint
  internal/app           # composition root and CLI runtime flow
  internal/model         # shared language used across the app
  internal/query         # orchestration / use case layer
  internal/source        # source pipeline and registry
  internal/fetch         # remote/local content acquisition contracts
  internal/parse         # raw document to entry transformation
  internal/normalize     # source-specific cleanup into canonical model
  internal/cache         # explicit cache-aside store
  internal/render        # markdown/json renderers
  internal/config        # runtime configuration contracts and defaults
  internal/doctor        # environment/health checks for the binary
  internal/platform      # OS-facing CLI boundary
  internal/version       # build metadata
```

## Runtime Flow

1. `cmd/dlexa` creates the platform adapter and calls `app.New(...).Run(...)`.
2. `internal/app` parses CLI flags, loads runtime config, and builds a `LookupRequest`.
3. `internal/query` performs explicit cache-aside orchestration:
   - compute cache key
   - attempt cache read
   - resolve sources
   - invoke each source
   - merge `SourceResult` values into one `LookupResult`
   - write the aggregated result back to cache
4. Each source runs an internal pipeline:
   - `fetch` obtains a raw document
   - `parse` converts it into candidate entries or a structured lookup miss
   - `normalize` maps those entries or miss metadata to the shared model
5. `internal/render` converts the final `LookupResult` into `markdown` or `json`.

## Shared Model

The shared domain language is centered on:

- `LookupRequest`
- `LookupResult`
- `LookupMiss`
- `Entry`
- `SourceDescriptor`
- `SourceResult`
- `Warning`
- `Problem`

This gives every package a stable vocabulary without coupling domain behavior to transport or presentation.

## Current Limitations

- The project is in alpha and still refining source fidelity and terminal rendering behavior.
- A full production source ecosystem is not complete yet; active work is tracked in OpenSpec changes.
- Cache is local (filesystem with in-memory fallback), with no distributed backend.
- `--doctor` currently reports baseline health checks and is intentionally lightweight.
- The tool should not be read as a promise of universal lexical coverage outside DPD-backed normative doubts.

## Roadmap

Planning and implementation are tracked in [openspec/changes](openspec/changes).

- In progress: [openspec/changes/dpd-live-lookup-parity](openspec/changes/dpd-live-lookup-parity)
- Archived/completed changes: [openspec/changes/archive](openspec/changes/archive)

> Roadmap entries are implementation artifacts and may evolve until verification is complete.

## DPD Table Rendering Strategy

- DPD tables stay in the shared article model as structured table blocks with per-cell span metadata when the source uses `rowspan` or `colspan`.
- Simple rectangular tables with exactly one header row and no spanning cells render as pipe-table Markdown so common live previews can render them directly.
- Complex DPD tables that rely on merged cells or multi-level structure render as HTML tables inside the Markdown payload because standard Markdown tables cannot represent that structure faithfully.
- This is an intentional fallback: Markdown-first when the structure is representable, HTML when semantic fidelity would otherwise be lost.

## Contributing

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md) for setup and workflow.

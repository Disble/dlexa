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
- It does **not** search article body content.
- It does **not** auto-fetch every candidate article.
- It does **not** turn `dlexa` into a generic dictionary search engine.

Examples:

```bash
# human-oriented candidate list
dlexa search abu dhabi

# json for scripts and follow-up automation
dlexa --format json search alicuota
```

Example human output shape:

```text
Candidate DPD entries for "abu dhabi":
- Abu Dhabi -> Abu Dabi
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
      "article_key": "alícuoto"
    }
  ]
}
```

## DPD consultation model

The CLI accepts free-text queries, but that does **not** mean every free-text request is a good fit. The intended use is consultation of DPD-style normative guidance.

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
   - `parse` converts it into candidate entries
   - `normalize` maps those entries to the shared model
5. `internal/render` converts the final `LookupResult` into `markdown` or `json`.

## Shared Model

The shared domain language is centered on:

- `LookupRequest`
- `LookupResult`
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

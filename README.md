# dlexa

[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8?logo=go)](go.mod)
[![Status](https://img.shields.io/badge/status-alpha-orange)](#project-status)

`dlexa` is a query-first Go CLI designed as a single binary with explicit architectural boundaries.

## Project Status

> ⚠️ **Alpha / Bootstrap**
>
> The project is functional and tested as a foundation, but still evolving. Core architecture and lookup flow are in place; external-source behavior and terminal semantics are still being iterated.

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
dlexa palabra

# explicit output format
dlexa --format json palabra

# force source selection
dlexa --source dpd palabra

# skip cache reads/writes for this request
dlexa --no-cache palabra

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
```

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

The repository starts with a real base for a dictionary/lookup CLI where the user asks a query and the system orchestrates the rest.

Key principles encoded in the structure:

- `cmd` is thin and only boots the application.
- `internal/app` is the composition root where concrete adapters are wired.
- `internal/query` orchestrates the lookup flow.
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

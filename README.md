# dlexa

[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8?logo=go)](go.mod)
[![Status](https://img.shields.io/badge/status-alpha-orange)](#project-status)

`dlexa` is a surface-based Go CLI for consulting normative linguistic doubts in Spanish through explicit routes such as `search`, `dpd`, and slug-based commands.

It is aimed at **normative linguistic doubts in Spanish**: orthographic, orthoepic/pronunciation, morphological, syntactic, and lexico-semantic questions that may start from semantic discovery, direct DPD lookup, or a previously identified editorial/FAQ route. It is **not** positioned as a generic dictionary replacement, an encyclopedic lookup tool, or a universal lexical search engine.

## Project Status

> ⚠️ **Alpha / Bootstrap**
>
> The project is functional and tested as a foundation, but still evolving. Core command surfaces and consultation flows are in place; external-source behavior and terminal semantics are still being iterated.

## What dlexa is for

Use `dlexa` when the question fits the repository's normative consultation model, for example:

- whether a spelling, accent, or graphic form is recommended
- whether a pronunciation or orthoepic variant is accepted
- whether a morphological or syntactic construction is advisable
- whether a usage recommendation depends on register, geography, or current usage
- whether a lexical-semantic recommendation is framed as a normative doubt rather than a generic definition request

The guidance exposed through `dlexa` is normative, but not brain-dead rigid. Recommendations can depend on **current usage**, **norma culta formal**, **register**, **geography**, and **communicative context**.

In practice, the flow is explicit:

- `dlexa search <consulta>` when you need discovery before choosing a destination surface
- `dlexa dpd <termino>` when you already know the DPD entry you want
- `dlexa espanol-al-dia <slug>`, `dlexa duda-linguistica <slug>`, or `dlexa noticia <slug>` when the route is already identified

## What dlexa is not

`dlexa` is not meant to replace:

- a general-purpose dictionary for arbitrary lexical lookup
- an etymology resource
- a translation tool
- an encyclopedic reference
- a broad lexical product that promises universal coverage outside DPD-resolved doubts

If the task is outside that normative scope, use a more appropriate source instead of forcing `dlexa` into the wrong job.

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
# direct DPD consultation when you already know the entry
dlexa dpd tilde

# federated semantic discovery across general RAE search + DPD index
dlexa search abu dhabi

# DPD-only entry discovery
dlexa dpd search abu dhabi

# open a known editorial slug
dlexa espanol-al-dia el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde

# explicit JSON output for direct lookup automation
dlexa --format json dpd solo

# structured entry-search output for automation
dlexa --format json search guion

# restrict search to one provider
dlexa search --source dpd adecua

# skip cache reads/writes for this request
dlexa --no-cache dpd imprimido

# health and version
dlexa --doctor
dlexa --version
```

## Development Quick Run

Run the CLI directly from source while iterating locally:

```bash
# direct DPD consultation
go run ./cmd/dlexa -- dpd bien

# search command
go run ./cmd/dlexa -- search abu dhabi

# slug-based article lookup
go run ./cmd/dlexa -- espanol-al-dia el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde

# explicit JSON output
go run ./cmd/dlexa -- --format json dpd solo

# skip cache for a live request
go run ./cmd/dlexa -- --no-cache dpd imprimido
```

## CLI Options

| Flag | Description |
| --- | --- |
| `--format` | Output format: `markdown` or `json` |
| `--source` | Search-only repeatable provider selector (`search`, `dpd`) |
| `--no-cache` | Skip cache reads and writes for the request |
| `--doctor` | Run environment checks |
| `--version` | Print binary version |

Usage:

```text
dlexa [--format markdown|json] [--no-cache] search [--source <id> ...] <query>
dlexa [--format markdown|json] [--no-cache] dpd <termino>
dlexa [--format markdown|json] [--no-cache] dpd search <query>
dlexa [--format markdown|json] [--no-cache] espanol-al-dia <slug>
dlexa [--format markdown|json] [--no-cache] duda-linguistica <slug>
dlexa [--format markdown|json] [--no-cache] noticia <slug>
dlexa [--doctor|--version]
```

## Help Model

`dlexa --help` and `dlexa <cmd> --help` are part of the CLI contract for both humans and LLMs.

The help output is capability-first:

- what the command enables
- what input shape it expects
- literal examples you can copy as-is
- agent-facing notes for structured automation
- the next natural command when you want to continue the flow

Common syntax failures stay in the fallback system (`Nivel 1 · Syntax`) instead of being the center of the command help body.

## Search command model

Use `dlexa search <query>` when you need semantic discovery before deciding which exact consultation route to run.

- By default it federates the general RAE search surface **and** the DPD `/srv/keys` entry-discovery provider.
- It returns curated candidates plus safe follow-up commands.
- It can surface executable follow-ups such as `dpd`, `espanol-al-dia`, `duda-linguistica`, and FAQ-compatible `noticia`.
- It can also surface related RAE destinations as deferred guidance when a mapped CLI command still does not exist.
- It focuses on discovering the most useful next consultation route.
- It can scope discovery to one provider with `--source`.
- It preserves the distinction between executable suggestions and deferred guidance.

Provider control:

- `dlexa search <query>` → federated default (`search` + `dpd`)
- `dlexa search --source dpd <query>` → DPD index only
- `dlexa search --source search <query>` → general RAE search only
- `dlexa dpd search <query>` → dedicated DPD-only entry-discovery surface

In practice:

- `dlexa search <query>` is the federated semantic gateway
- `dlexa dpd <query>` remains the direct DPD consultation path
- `dlexa dpd search <query>` preserves a dedicated DPD-only discovery path

### Executable suggestions vs deferred guidance

`dlexa search` can return two different kinds of follow-up hints:

- `- sugerencia:` → the next step is an executable CLI command right now (for example, a `dpd`, `espanol-al-dia`, `duda-linguistica`, or FAQ-style `noticia` lookup)
- deferred access block (`🌐 ...` + `Acceso futuro via CLI`) → useful navigation text that looks command-shaped, but is **not** an available CLI subcommand yet

This distinction matters for both humans and agents:

- some `next_command` values are ready to execute immediately
- others are guidance about a future or indirect route
- when output is JSON, inspect each candidate's `deferred` field before follow-up automation
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

- total_candidatos: 3
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
    "Query": "solo o sólo",
    "Format": "json",
    "Sources": ["search", "dpd"],
    "NoCache": false
  },
  "Outcome": "results",
  "Candidates": [
    {
      "raw_label_html": "<strong>solo</strong>",
      "display_text": "solo",
      "article_key": "solo",
      "next_command": "dlexa dpd solo",
      "deferred": false
    },
    {
      "raw_label_html": "<strong>Tilde en solo</strong>",
      "display_text": "Tilde en solo",
      "article_key": "solo",
      "module": "espanol-al-dia",
      "next_command": "dlexa espanol-al-dia solo",
      "deferred": false
    }
  ]
}
```

## Normative consultation model

The CLI accepts free-text queries, but that does **not** mean every free-text request is a good fit. The intended use is consultation of DPD-style normative guidance through explicit command surfaces.

If a direct lookup does not resolve to an exact DPD article, `dlexa` can return a structured lookup miss instead of silently rerouting the command:

- preserve a native DPD related-entry suggestion when one exists
- otherwise show an explicit `dlexa search <consulta>` next step
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

The repository provides a real base for a normative-doubt CLI with explicit command surfaces. Users and agents choose a route (`search`, `dpd`, or a slug-based command), and the system orchestrates retrieval, normalization, rendering, and structured fallbacks without collapsing the answer into a fake one-size-fits-all rule.

Key principles encoded in the structure:

- `cmd/dlexa` stays thin and defines the public Cobra command tree.
- `internal/app` is the runtime composition root that wires config, doctor, modules, and shared renderers.
- `internal/modules` defines the CLI-facing module contract and keeps each command surface explicit.
- `internal/query` owns lookup orchestration for article-style modules such as `dpd`, `espanol-al-dia`, `duda-linguistica`, and `noticia`.
- `internal/search` owns semantic discovery orchestration for the `search` module.
- `internal/source` and `internal/search` providers both rely on explicit `fetch -> parse -> normalize` pipelines.
- `internal/model` holds the shared language for entries, search candidates, help, and fallbacks.
- `internal/render` and `internal/renderutil` keep markdown/json output concerns outside module logic.
- `internal/cache` is explicit cache-aside infrastructure; `internal/platform` and `internal/version` isolate binary/runtime concerns.

## Directory Layout

```text
dlexa/
  cmd/dlexa              # thin binary entrypoint
  internal/app           # runtime composition root and shared CLI services
  internal/modules       # CLI-facing command modules and shared module contract
  internal/search        # semantic discovery service and provider registry
  internal/query         # lookup orchestration for article-style modules
  internal/source        # lookup-source pipeline registry
  internal/fetch         # upstream acquisition adapters
  internal/parse         # raw upstream document parsers
  internal/normalize     # normalization into shared model contracts
  internal/model         # shared domain language and transport-neutral envelopes
  internal/render        # markdown/json renderers
  internal/renderutil    # shared rendering helpers
  internal/cache         # filesystem/memory cache stores
  internal/config        # runtime defaults and configuration loading
  internal/doctor        # environment/health checks for the binary
  internal/platform      # OS-facing CLI boundary
  internal/version       # build metadata
```

## Runtime Flow

1. `cmd/dlexa` builds the Cobra command tree and delegates execution to `executeRootCommand(...)`.
2. `executeRootCommand(...)` forwards the chosen public command (`search`, `dpd`, `espanol-al-dia`, `duda-linguistica`, `noticia`, `--doctor`, `--version`) into the `internal/app` runtime boundary.
3. `internal/app` loads runtime config, applies module-specific defaults, validates format, and resolves the target module from `internal/modules.Registry`.
4. The target module executes one of two orchestration paths:
   - lookup modules (`dpd`, `espanol-al-dia`, `duda-linguistica`, `noticia`) delegate to `internal/query`
   - discovery (`search`) delegates to `internal/search`
5. Lookup and search providers both run explicit pipelines:
   - `fetch` obtains upstream content
   - `parse` extracts structured records or misses
   - `normalize` maps upstream output into the shared model
6. `internal/render` wraps successful responses, help payloads, or structured fallback envelopes as Markdown or JSON for the CLI output.

## Shared Model

The shared domain language is centered on:

- `LookupRequest`
- `LookupResult`
- `SearchRequest`
- `SearchResult`
- `SearchCandidate`
- `LookupMiss`
- `HelpEnvelope`
- `FallbackEnvelope`
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

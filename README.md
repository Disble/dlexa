# dlexa

`dlexa` is a query-first Go CLI designed as a single binary with explicit architectural boundaries.

## Intent

The repository starts with a real base for a dictionary/lookup style CLI where the user asks a query and the system orchestrates the rest.

Key principles already encoded in the structure:

- `cmd` is thin and only boots the application.
- `internal/app` is the composition root where concrete adapters are wired.
- `internal/query` orchestrates the lookup flow.
- `internal/model` holds the shared domain language.
- `internal/source` composes `fetch -> parse -> normalize` for each source.
- `internal/render` stays outside the domain so output concerns do not leak inward.
- `internal/cache` is explicit cache-aside, owned by the query orchestration.
- `internal/fetch` and `internal/parse` are separate on purpose.
- `internal/version` and `internal/platform` keep the binary concerns isolated.

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

## Flow

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

The shared domain language is centered on these types:

- `LookupRequest`
- `LookupResult`
- `Entry`
- `SourceDescriptor`
- `SourceResult`
- `Warning`
- `Problem`

That gives every package a stable vocabulary without coupling the domain to transport or presentation.

## Current State

This bootstrap is intentionally minimal but not fake:

- the binary wiring exists
- the packages contain real interfaces and starter implementations
- the default source runs through `fetch -> parse -> normalize`
- markdown and json output are both wired
- cache-aside behavior is explicit in the query service

The implementation is still a foundation repo. Real external sources, richer parsing rules, persistence-backed cache, and doctor checks can be added without changing the architectural direction.

## Example Usage

```bash
dlexa --format markdown palabra
dlexa --format json --source demo cache
dlexa --doctor
dlexa --version
```

## Architectural Notes

- Query-first means the core use case starts from `LookupRequest`, not from source adapters.
- Source adapters are replaceable because the query layer only depends on interfaces.
- Renderers are separate so domain structs remain reusable for other outputs later.
- The project is prepared for a single binary distribution with runtime-selected behavior instead of multiple executables.
- `dpd-live-lookup-parity` owns DPD fetch/parse/normalize semantic preservation; `dpd-terminal-semantic-rendering` owns the final stdout contract so authored DPD semantics remain visible at the terminal boundary instead of being replaced by renderer-invented wrappers.

## DPD Table Rendering Strategy

- DPD tables stay in the shared article model as structured table blocks with per-cell span metadata when the source uses `rowspan` or `colspan`.
- Simple rectangular tables with exactly one header row and no spanning cells render as pipe-table Markdown so common live previews can render them directly.
- Complex DPD tables that rely on merged cells or multi-level structure render as HTML tables inside the Markdown payload because standard Markdown tables cannot represent that structure faithfully.
- This is an intentional fallback: Markdown-first when the structure is representable, HTML when semantic fidelity would otherwise be lost.

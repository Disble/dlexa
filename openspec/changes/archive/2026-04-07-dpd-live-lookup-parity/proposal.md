# Proposal: DPD Live Lookup Parity

## Intent

Make `dlexa` return real Diccionario panhispánico de dudas content for the parity anchor `bien` by replacing bootstrap-only behavior with a live DPD lookup path that preserves the existing query-first architecture.

## Problem Statement

Today `dlexa` is architecturally prepared for real sources, but its default source path is still fake bootstrap scaffolding. As a result, `dlexa bien` cannot return the same information hierarchy as the real DPD article, cannot distinguish real remote failure modes, and cannot be trusted as a production lookup for DPD content.

## Goals

- Make `dlexa bien` return the same canonical information as the live DPD article for `bien`.
- Remove fake production lookup behavior from the default DPD path.
- Preserve the repository's explicit `fetch -> parse -> normalize -> render` boundaries.
- Introduce a canonical DPD article model rich enough for ordered sections, nested subitems, inline emphasis, cross-references, and citation metadata.
- Make parity verifiable through stable contract tests plus optional live probes.

## Non-Goals

- Full coverage of every DPD article shape in v1.
- Pixel-perfect or chrome-complete reproduction of `rae.es` pages.
- Introducing an internal database or any production mock-backed behavior.
- Re-architecting unrelated sources, cache storage, or CLI command surface beyond what this change requires.
- Reusing `go-rae` as an implementation blueprint; it is a reference for boundaries, not a drop-in solution.

## Parity Definition for `dlexa bien`

For v1, parity means the rendered and structured output MUST preserve the canonical article information from the live DPD entry for `bien`, specifically:

- dictionary context: `Diccionario panhispánico de dudas`
- edition marker: `2.ª edición`
- lemma heading: `bien`
- ordered top-level sections `1.` through `7.`
- nested subitems inside section `6.`: `a)`, `b)`, `c)`
- inline emphasis semantics for forms, locutions, and contrastive terms such as `más bien`, `mejor`, and `si bien`
- readable cross-reference semantics such as `→ [6]` and `→ [7]`
- citation essentials: source, canonical URL, edition, and consultation metadata

Parity explicitly does NOT include site chrome, menus, related-content blocks, share widgets, newsletter/footer content, or other page-navigation noise.

## Scope

### In Scope

- Switch the default production lookup path from demo scaffolding to a real DPD-backed source.
- Add a canonical article representation for the subset of DPD structure required by `bien`.
- Implement article-body extraction that isolates canonical DPD content from page chrome.
- Parse and normalize the extracted content into stable ordered article nodes.
- Render markdown/text output that preserves hierarchy, numbering, nested lettering, emphasis, and reference readability.
- Expose the same canonical structure in JSON output.
- Distinguish transport failure, not-found, and parse/normalization failure in the lookup flow.
- Add fixture-based contract coverage for parse/normalize/render and optional live integration verification.

### Out of Scope

- Generic support for all DPD article families.
- Support for other dictionaries or sources beyond maintaining existing extension seams.
- HTML/browser rendering parity with the full `rae.es` page.
- Persistent cache/database redesign.
- Large CLI UX changes unrelated to returning correct DPD article content.

## Architectural Direction

Adopt a staged canonical-parity approach.

- Keep `cmd/dlexa` thin and preserve `internal/app` as the composition root.
- Preserve the source pipeline boundary: `fetch -> parse -> normalize`.
- Expand the shared domain model so DPD article structure is first-class instead of being flattened into `Entry.Content` plus loose metadata.
- Keep rendering downstream of normalization so markdown and JSON derive from the same canonical article representation.
- Treat `go-rae` only as evidence for explicit network configuration, typed results, and domain-level not-found handling; do not mirror its API-client-centered architecture because `dlexa` must parse live DPD article content directly.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/app/wiring.go` | Modified | Register and default the live DPD source instead of bootstrap demo behavior. |
| `internal/config/interfaces.go` | Modified | Carry runtime knobs for live DPD acquisition and defaults. |
| `internal/config/static.go` | Modified | Set default source/runtime values for DPD operation. |
| `internal/fetch/interfaces.go` | Modified | Support real HTTP document metadata needed by the pipeline. |
| `internal/parse/interfaces.go` | Modified | Accept structured raw DPD content instead of bootstrap markdown assumptions. |
| `internal/normalize/interfaces.go` | Modified | Normalize parsed DPD nodes into canonical article structures. |
| `internal/model/types.go` | Modified | Add first-class article identity, ordered sections, nested items, inline marks/references, and citation metadata. |
| `internal/render/markdown.go` | Modified | Render canonical article hierarchy with parity-sensitive formatting. |
| `internal/render/json.go` | Modified | Emit structured DPD article data rather than flattened diagnostic output. |
| `internal/query/service.go` | Modified | Surface fetch/not-found/parse failures distinctly in orchestration. |
| `internal/source/registry.go` | Modified | Keep source selection but update default registry contents for DPD. |
| `internal/*_test.go` | Modified | Add contract-oriented verification around parity-sensitive behavior. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| DPD structure is forced into the current flat model | High | Introduce a canonical article model before renderer parity work. |
| Extraction depends on incidental page layout and breaks on site changes | Medium | Isolate semantic article-body selection and verify against captured fixtures. |
| `bien` parity is mistaken for universal DPD coverage | High | Freeze v1 scope to the `bien` article shape and state non-goals explicitly. |
| Remote dependency introduces flaky verification | Medium | Use captured authoritative fixtures for contracts and keep live probes opt-in. |
| `go-rae` is over-applied despite different upstream assumptions | Medium | Treat it as a boundary reference only and keep direct live-DPD parsing as the core approach. |

## Rollback Plan

If the live DPD path proves unstable, revert the default wiring to the prior non-production-safe source selection, remove the new DPD source registration from the default path, and restore the previous renderer/model expectations as one atomic change. This rollback is safe because the change is confined to explicit source, model, query, and render boundaries rather than a cross-cutting platform rewrite.

## Dependencies

- Availability and structural stability of the live DPD article source at `rae.es`.
- A captured authoritative reference fixture for the `bien` article to stabilize contract verification.
- Existing query-first architecture and source pipeline boundaries already present in `dlexa`.
- The user-provided `go-rae` repository only as comparison input for client-boundary discipline, not for parser reuse.

## Verification Expectations

- `dlexa bien` markdown/text MUST include the DPD context, edition, lemma, ordered sections `1..7`, nested `6.a..6.c`, preserved emphasis semantics, readable cross-references, and citation essentials.
- `dlexa bien` output MUST NOT include page chrome or unrelated site content.
- JSON output SHOULD expose the same canonical article hierarchy and metadata directly.
- Lookup errors MUST distinguish at least remote fetch failure, not-found outcome, and parse/normalization failure.
- Parser, normalizer, and renderer behavior MUST be covered by fixture-based contract tests using captured authoritative DPD content.
- Live verification against the real DPD MAY exist, but only as explicit opt-in integration coverage.

## Success Criteria

- [ ] The default DPD lookup path uses live remote acquisition instead of bootstrap demo content.
- [ ] `dlexa bien` returns the same canonical information hierarchy as the real DPD article for `bien`.
- [ ] Output excludes site chrome while preserving numbered sections, nested lettering, emphasis, references, and citation metadata.
- [ ] JSON and markdown/text outputs are generated from the same canonical article model.
- [ ] Verification is stable without any production mock flow or internal database.

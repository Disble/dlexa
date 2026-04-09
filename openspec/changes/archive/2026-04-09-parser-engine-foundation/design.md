# Design: Parser Engine Foundation

## Technical Approach
Add an additive parser-engine contract layer without moving concrete parsers. `internal/parse/engine` defines the shared input envelope and the two family ports; bridge adapters live at the existing seams (`internal/source`, `internal/search`) so lookup and search can opt into the new contracts later without import cycles or runtime drift. This preserves the current `fetch -> parse -> normalize` execution path required by the active specs while aligning code with ADR-0001.

## Architecture Decisions

| Decision | Options | Choice | Rationale |
|---|---|---|---|
| Engine package scope | Full parser relocation now; contracts only | Contracts only in `internal/parse/engine` | Keeps the first slice additive and avoids touching `internal/parse/dpd.go`, `live_search.go`, and `dpd_search.go`. |
| Port outputs | New wrapper result types; reuse current parsed contracts | Reuse `parse.Result` and `[]parse.ParsedSearchRecord` | Minimizes churn in normalizers and preserves current tests/behavior. |
| Bridge location | Put all adapters in `engine`; put adapters at pipeline seams | Put adapters in `internal/source` and `internal/search` | Avoids Go import cycles and matches the real integration seams already verified in `pipeline.go` and `provider.go`. |
| Resolver/registry | Introduce active resolver now; defer | Defer active resolver wiring | A registry adds abstraction with no runtime payoff until at least one native engine parser exists. |

## Data Flow
Current path stays valid:

    fetcher -> parse.Parser/search.Parser -> normalizer

Foundation adds an opt-in engine path:

    fetcher -> seam bridge -> engine.ParseInput -> engine.ArticleParser/SearchParser
                                  |                         |
                                  └---- legacy adapter -----┘

For migrated callers, the seam bridge constructs `ParseInput{Ctx, Descriptor, Document}` once and forwards to the engine port. For untouched callers, existing constructors still use legacy interfaces directly.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/parse/engine/doc.go` | Create | Package docs describing the parser-engine foundation and strict non-goals. |
| `internal/parse/engine/input.go` | Create | Define `ParseInput`. |
| `internal/parse/engine/article.go` | Create | Define `ArticleParser` and `ArticleResult` alias. |
| `internal/parse/engine/search.go` | Create | Define `SearchParser`. |
| `internal/source/engine_bridge.go` | Create | Add article-side adapters and `NewEnginePipelineSource(...)`. |
| `internal/search/engine_bridge.go` | Create | Add search-side adapters and `NewEnginePipelineProvider(...)`. |
| `internal/source/engine_bridge_test.go` | Create | Verify article bridge fidelity and call ordering. |
| `internal/search/engine_bridge_test.go` | Create | Verify search bridge fidelity and warning propagation. |

## Interfaces / Contracts

```go
package engine

type ParseInput struct {
    Ctx        context.Context
    Descriptor model.SourceDescriptor
    Document   fetch.Document
}

type ArticleResult = parse.Result

type ArticleParser interface {
    ParseArticle(input ParseInput) (ArticleResult, []model.Warning, error)
}

type SearchParser interface {
    ParseSearch(input ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error)
}
```

Bridge rules:
- `internal/source` adapts `engine.ArticleParser` <-> existing `parse.Parser`.
- `internal/search` adapts `engine.SearchParser` <-> existing `search.Parser`.
- Bridges must be lossless: same descriptor, same `fetch.Document`, same parsed output, same warnings/errors.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `ParseInput` forwarding and adapter fidelity | New bridge tests with recording parser stubs mirroring existing `pipeline_test.go` style. |
| Integration | Engine constructors preserve lookup/search pipeline behavior | Reuse source/search pipeline tests with bridge-backed parsers. |
| E2E | CLI behavior | No new E2E coverage required because `internal/app/wiring.go` remains unchanged in this slice. |

## Migration / Rollout
No migration required.

Phase boundary for this slice:
1. Add `internal/parse/engine` contracts.
2. Add opt-in engine constructors in `internal/source` and `internal/search`.
3. Keep `NewPipelineSource`, `NewPipelineProvider`, and `internal/app/wiring.go` unchanged.
4. Defer concrete parser relocation, helper extraction from `internal/parse/dpd.go`, and resolver activation to later slices.

Helpers extractable now: only bridge/input assembly helpers local to the seams.
Helpers deferred: HTML/text/inline/table helpers such as `cleanText`, `normalizeInlinePlainText`, `extractAttribute`, `parseTable*`, and the inline parser stack, because they are still DPD-biased and should move only with concrete parser migration.

## Open Questions
- [ ] No dedicated `sdd/parser-engine-foundation/spec` artifact was found in Engram; this design is grounded on the proposal, explore artifact, ADR-0001, and active OpenSpec runtime requirements.
- [ ] Decide in the next slice whether resolver/registry scaffolding is introduced together with the first native engine parser rather than as idle infrastructure.

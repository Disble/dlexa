# Design: Duda Lingüística Article Surface

## Technical Approach

Implement `duda-linguistica` as another first-class article-family lookup surface, deliberately reusing the same fetch → parse → normalize → render chain already proven by `espanol-al-dia`. This avoids inventing a parallel content model for short-answer pages.

## Architecture Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Delivery path | Reuse lookup pipeline | The surface returns article-like HTML and fits `model.Article` without new rendering contracts. |
| Parser shape | Surface-specific parser with shared structural assumptions | Live DOM verification showed parity with the `espanol-al-dia` shell, but keeping a separate parser isolates upstream drift per surface. |
| Search truthfulness | Mark `duda-linguistica` executable | Search suggestions must reflect registered commands, not stale deferred assumptions. |
| Defaults | Module-specific source default | Prevents accidental fallback to `dpd` when running the new command. |

## Data Flow

`dlexa duda-linguistica <slug>`

→ `cmd/dlexa/duda_linguistica.go`
→ `internal/app.App.ExecuteModule("duda-linguistica", ...)`
→ `internal/modules/dudalinguistica.Module`
→ `internal/query.LookupService`
→ `internal/source.PipelineSource`
→ `internal/fetch.DudaLinguisticaFetcher`
→ `internal/parse/engine.DudaLinguisticaArticleParser`
→ `internal/normalize.DudaLinguisticaNormalizer`
→ shared markdown/json lookup renderers

## Testing Strategy

- CLI black-box tests for routing, help, and syntax fallback.
- Fetcher tests for URL resolution and typed 404 handling.
- Parser tests for successful extraction and explicit broken-markup failure.
- Normalizer tests for lookup entry generation and empty-section rejection.
- Module and app wiring tests for source defaults and engine registration.
- Search filter test update proving `duda-linguistica` is no longer deferred.

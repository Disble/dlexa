# Design: Normative Article Surface Parsers

## Technical Approach

Implement only the `espanol-al-dia` surface as a first-class lookup module on top of the parser-engine article family. Reuse the existing lookup/query/render pipeline instead of inventing a separate article delivery path. This keeps the product model simple: `search` discovers, `espanol-al-dia` resolves concrete article slugs, and `dpd` remains the root-default normative lookup.

## Architecture Decisions

| Decision | Options | Choice | Rationale |
|---|---|---|---|
| Surface scope | `espanol-al-dia` + `duda-linguistica`; `espanol-al-dia` only | `espanol-al-dia` only | Verified DOM evidence exists for `espanol-al-dia`; the user explicitly narrowed scope. |
| Delivery path | New custom renderer path; reuse lookup pipeline | Reuse lookup/query/render pipeline | The existing `model.Article` + lookup renderers already support article-family content. |
| Search truthfulness | Keep non-DPD surfaces deferred; mark implemented ones executable | Mark implemented `espanol-al-dia` executable | Search must reflect runtime truth, not stale roadmap assumptions. |
| Default source behavior | Inherit generic lookup defaults; module-specific defaults | Module-specific `espanol-al-dia` source default | Prevents source drift where the command accidentally queries DPD. |

## Data Flow

`dlexa espanol-al-dia <slug>`

→ `cmd/dlexa/espanol_al_dia.go`
→ `internal/app.App.ExecuteModule("espanol-al-dia", ...)`
→ `internal/modules/espanolaldia.Module`
→ `internal/query.LookupService`
→ `internal/source.PipelineSource`
→ `internal/fetch.EspanolAlDiaFetcher`
→ `internal/parse/engine.EspanolAlDiaArticleParser`
→ `internal/normalize.EspanolAlDiaNormalizer`
→ shared markdown/json lookup renderers

The search flow remains separate, but `internal/modules/search/filter.go` now treats `espanol-al-dia` as executable.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/fetch/espanol_al_dia.go` | Create | Fetch concrete `/espanol-al-dia/<slug>` article pages. |
| `internal/parse/espanol_al_dia.go` | Create | Extract title and article paragraphs from verified DOM regions. |
| `internal/parse/engine/espanol_al_dia_article.go` | Create | Engine-native article-family wrapper. |
| `internal/normalize/espanol_al_dia.go` | Create | Project parsed article into `model.Entry` + `model.Article`. |
| `internal/modules/espanolaldia/module.go` | Create | Shared module adapter for the new surface. |
| `cmd/dlexa/espanol_al_dia.go` | Create | Public CLI command. |
| `internal/app/wiring.go` | Modify | Register source + module in the composition root. |
| `internal/modules/search/filter.go` | Modify | Mark implemented `espanol-al-dia` suggestions as executable. |

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| CLI | Command routing/help/syntax fallback | Black-box tests via `executeRootCommand`. |
| Fetch | URL building and typed errors | `roundTripFunc` HTTP tests following existing fetch patterns. |
| Parse | Title/body extraction and broken markup failure | Focused parser tests with representative HTML slices. |
| Normalize | Entry/article projection and empty-section failure | Unit tests using parsed article fixtures inline. |
| Wiring | Registration and engine parser wiring | `internal/app/wiring_test.go` lookup registry assertions. |
| Search truthfulness | Executable vs deferred suggestions | Existing module/render/filter tests updated to current runtime truth. |

## Non-Goals

- Do not implement `duda-linguistica` in this slice.
- Do not make `noticia` executable.
- Do not introduce a new article model separate from `model.Article`.

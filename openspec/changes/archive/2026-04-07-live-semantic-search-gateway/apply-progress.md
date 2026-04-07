# Apply Progress: Live Semantic Search Gateway

## TDD Cycle Evidence

| Task | RED | GREEN | REFACTOR |
|---|---|---|---|
| 1.1 | Added `internal/fetch/live_search_test.go` covering live RAE search URL construction, browser-profile headers, and explicit transport failure typing. | Implemented `internal/fetch/live_search.go` and passed `go test ./internal/fetch/...`. | Reused shared fetch helpers and kept bounded HTTP behavior behind the existing fetch contract. |
| 1.2 | Covered by task 1.1 failing fetch tests. | Implemented live search fetcher without `/srv/keys` reuse. | Kept fetcher provider-specific while preserving `fetch.Fetcher` contract. |
| 1.3 | Added `internal/parse/live_search_test.go` for extracted candidates, empty search markup, and broken-markup parse failure. | Implemented `internal/parse/live_search.go` and passed `go test ./internal/parse/...`. | Extracted small parser helpers for URL resolution and text normalization. |
| 1.4 | Covered by task 1.3 failing parser tests. | Implemented stable live-search record extraction from RAE search HTML. | Added deterministic fixtures under `internal/parse/testdata/`. |
| 1.5 | Added `internal/normalize/live_search_test.go` for curated candidate normalization and unusable-record rejection. | Implemented `internal/normalize/live_search.go` and passed `go test ./internal/normalize/...`. | Kept gateway semantics out of normalizer; only normalized provider data. |
| 1.6 | Covered by task 1.5 failing normalizer tests. | Implemented provider-neutral normalized candidates with title/snippet/url/source hint. | Simplified normalized output to the minimum fields needed upstream. |
| 1.7 | Added `internal/search/live_service_test.go` for parse/normalize failure preservation; existing `internal/search/service_test.go` remained green as safety net. | Service passed `go test ./internal/search/...` with live adapters and preserved typed failures. | Generalized comments and added test-only accessors instead of changing orchestration boundaries. |
| 1.8 | Covered by task 1.7 failing service tests. | Updated `internal/search/service.go` minimally and kept cache-aside orchestration intact. | Restricted changes to provider-neutral wording and test helpers. |
| 2.1 | Added `internal/app/wiring_test.go` proving `search` wires to live adapters and preserved CLI regression in `cmd/dlexa/root_test.go`. | Updated wiring and passed `go test ./internal/app/... ./cmd/dlexa/...`. | Added narrow test-only accessors instead of widening production interfaces. |
| 2.2 | Covered by task 2.1 failing wiring tests. | Retargeted `internal/app/wiring.go` from DPD `/srv/keys` adapters to live search adapters. | Kept `cmd/dlexa/search.go` thin and added no Cobra commands. |
| 2.3 | Expanded `internal/modules/search/module_test.go` for explicit no-results versus parse-failure fallback. | Updated module outcome classification and passed `go test ./internal/modules/search/...`. | Introduced `model.SearchOutcome` instead of abusing fallback envelopes for successful empty searches. |
| 2.4 | Covered by task 2.3 failing module tests. | Successful empty curation now renders no-results; transport/parse/normalize failures still map to fallback ladder. | Kept classification inside `internal/modules/search`. |
| 2.5 | Extended `internal/modules/search/module_test.go` fixture coverage for institutional drop and `/noticia/*` rescue. | Filtering rules now drop `/institucion/*` and rescue FAQ/normative `/noticia/*`. | Consolidated rescue logic in `isRescuedNoticia`. |
| 2.6 | Covered by task 2.5 failing filter/module tests. | Implemented keep/drop/rescue rules in `internal/modules/search/filter.go`. | Reduced duplication by centralizing title-based rescue logic. |
| 2.7 | Expanded `internal/modules/search/module_test.go` fixture coverage for DPD mapping and unknown-URL fallback. | Updated mapper logic and passed `go test ./internal/modules/search/...`. | Preserved safe fallback syntax for unknown URLs and special-cased DPD URL mapping. |
| 2.8 | Covered by task 2.7 failing mapping tests. | Known URLs now produce literal `dlexa ...` suggestions; unknown URLs stay visible with safe fallback. | Kept mapping rules isolated in `internal/modules/search/mapper.go`. |
| 2.9 | Extended `internal/render/search_markdown_test.go` and `internal/render/search_json_test.go` for suggestions, unmapped results, and explicit no-results output. | Updated renderers and passed `go test ./internal/render/...`. | Clarified markdown labels (`sugerencia`, `fallback_command`, `url`) and preserved JSON structure. |
| 2.10 | Covered by task 2.9 failing renderer tests. | Renderers now distinguish mapped suggestions, unknown visible results, and no-results states. | Kept JSON renderer simple by serializing `SearchResult` directly. |
| 3.1 | Added root-command regression table in `cmd/dlexa/root_test.go`. | Passed `go test ./cmd/dlexa/...`. | Folded gateway-vs-DPD regression into one table test. |
| 3.2 | Initial fixture helper in `internal/testutil/live_search.go` triggered lint pressure. | Refactored shared live-search constants into `internal/testutil/live_search.go` and localized fixture loading to parse tests. | Removed generic file-loader helper to satisfy `gosec` while keeping shared constants. |
| 3.3 | N/A — verification task. | Passed `go test ./...`. | None needed beyond targeted cleanup and gofmt. |
| 3.4 | N/A — verification task. | Passed `go tool --modfile=golangci-lint.mod golangci-lint run ./...`. | Fixed lint by removing the generic fixture reader and adding package/const comments. |

## Notes

- `cmd/dlexa/search.go` remained thin.
- `internal/search.Service` remained the fetch/parse/normalize/cache orchestrator.
- `internal/modules/search` owns curation, rescue logic, URL mapping, and explicit no-results classification.
- No new Cobra destination subcommands were added.

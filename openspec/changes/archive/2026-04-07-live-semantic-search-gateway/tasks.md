# Tasks: Live Semantic Search Gateway

## Phase 1: Infrastructure

- [x] 1.1 Add RED tests in `internal/fetch/live_search_test.go` for live RAE search request construction, transport failure typing, and successful document capture without reusing the DPD `/srv/keys` path.
- [x] 1.2 Implement `internal/fetch/live_search.go` with the live search fetcher, bounded HTTP behavior, and typed transport errors that fit the existing search service contracts.
- [x] 1.3 Add RED fixture-based tests in `internal/parse/live_search_test.go` for extracting raw live search candidates, empty upstream payloads, and broken-markup parse failures.
- [x] 1.4 Implement `internal/parse/live_search.go` so live search responses are parsed into stable internal search records with explicit parse-failure outcomes.
- [x] 1.5 Add RED tests in `internal/normalize/live_search_test.go` for converting parsed live results into normalized candidates with title, snippet, source URL, and next-step metadata.
- [x] 1.6 Implement `internal/normalize/live_search.go` so parsed live results normalize into provider-neutral search candidates without moving gateway semantics below the module layer.
- [x] 1.7 Add RED tests in `internal/search/service_test.go` for cache-aside orchestration with the live fetch/parse/normalize adapters and for preserving typed transport/parse/normalize failures.
- [x] 1.8 Update `internal/search/service.go` only as needed to keep provider-neutral orchestration intact with the new live adapters and passing RED service coverage.

## Phase 2: Implementation

- [x] 2.1 Add RED wiring tests in `internal/app/app_test.go` or `internal/app/wiring_test.go` proving `search` uses live search adapters while root bare queries still route to DPD unchanged.
- [x] 2.2 Update `internal/app/wiring.go` to swap DPD search adapters for live fetch/parse/normalize adapters without expanding the Cobra command tree.
- [x] 2.3 Add RED table tests in `internal/modules/search/module_test.go` for explicit no-results when successful live searches curate to zero candidates versus fallback on transport or parse failures.
- [x] 2.4 Update `internal/modules/search/module.go` so successful-but-empty search results classify as no-results and failures still map through the fallback path.
- [x] 2.5 Add RED tests in `internal/modules/search/filter_test.go` or `module_test.go` for dropping institutional `/institucion/*` noise and rescuing linguistically valuable `/noticia/*` cases.
- [x] 2.6 Update `internal/modules/search/filter.go` to codify live-search keep/drop/rescue rules, including the approved `/noticia/*` rescue behavior.
- [x] 2.7 Add RED tests in `internal/modules/search/mapper_test.go` or `module_test.go` for known URL → literal command suggestions and safe fallback when the URL shape is unknown.
- [x] 2.8 Update `internal/modules/search/mapper.go` to preserve safe `dlexa ...` suggestion mapping for known surfaces and explicit unknown-URL fallback without inventing syntax.
- [x] 2.9 Add RED renderer tests in `internal/render/search_markdown_test.go` and `internal/render/search_json_test.go` for mapped suggestions, unmapped visible results, and explicit no-results output.
- [x] 2.10 Update `internal/render/search_markdown.go` and `internal/render/search_json.go` so search output distinguishes suggested commands, unmapped candidates, and no-results states safely for agents.

## Phase 3: Verification

- [x] 3.1 Add regression coverage in `cmd/dlexa` and/or `internal/app/app_test.go` proving `dlexa <query>` still defaults to DPD while `dlexa search <query>` invokes the live semantic gateway.
- [x] 3.2 Refactor shared live-search fixtures/helpers across `internal/fetch`, `internal/parse`, `internal/normalize`, and `internal/modules/search` so the new RED→GREEN coverage stays deterministic and duplication-light.
- [x] 3.3 Run `go test ./...` to verify the full live-search gateway behavior, regressions, and renderer safety without introducing any build step.
- [x] 3.4 Run `go tool --modfile=golangci-lint.mod golangci-lint run ./...` to verify lint cleanliness for the wiring, adapters, module logic, and renderers without building.

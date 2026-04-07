# Tasks: dlexa v2 Cobra Migration

## Phase 1: Contracts and shared foundations

- [x] 1.1 RED: Add `internal/modules/interfaces_test.go` and `internal/render/envelope_test.go` for module request/response contracts, fallback kinds, envelope metadata, and JSON-bypass expectations.
- [x] 1.2 GREEN: Create `internal/modules/interfaces.go`, `internal/render/envelope.go`, and extend `internal/model/types.go` / `internal/model/search.go` with shared module, envelope, and fallback types.
- [x] 1.3 REFACTOR: Simplify shared helpers in `internal/model/*` and `internal/render/interfaces.go` so envelope/fallback shaping lives in one place.

## Phase 2: Cobra CLI surface and composition root

- [x] 2.1 RED: Add `cmd/dlexa/root_test.go`, `cmd/dlexa/dpd_test.go`, `cmd/dlexa/search_test.go`, and extend `internal/app/app_test.go` for default DPD routing, explicit subcommands, Markdown help, and syntax failures.
- [x] 2.2 GREEN: Add Cobra in `go.mod`, create `cmd/dlexa/root.go`, `cmd/dlexa/dpd.go`, `cmd/dlexa/search.go`, and update `cmd/dlexa/main.go` to execute the root command.
- [x] 2.3 GREEN: Refactor `internal/app/wiring.go` to compose the command tree, module registry, shared renderer, and legacy services; trim `internal/app/app.go` away from `flag` routing.
- [x] 2.4 REFACTOR: Remove obsolete `flag` parsing/helpers from `internal/app/app.go` and keep only reusable execution or wiring helpers still needed by tests.

## Phase 3: Module implementations and semantic search

- [x] 3.1 RED: Add `internal/modules/dpd/module_test.go` for request translation, cache metadata, structured not-found, and renderer handoff across Markdown/JSON.
- [x] 3.2 GREEN: Create `internal/modules/dpd/module.go` (plus local helpers if needed) to wrap `internal/query` behind the shared module contract.
- [x] 3.3 RED: Add `internal/modules/search/module_test.go` and `internal/modules/search/testdata/*` for institutional-noise filtering, FAQ rescue, URL→command mapping, and fallback classification.
- [x] 3.4 GREEN: Create `internal/modules/search/module.go`, `filter.go`, and `mapper.go` over `internal/search/service.go` with semantic next-step suggestions.
- [x] 3.5 REFACTOR: Extract shared fallback and command-mapping helpers between `internal/modules/search/*` and `cmd/dlexa/*.go` to avoid drift.

## Phase 4: Rendering, compatibility, and verification

- [x] 4.1 RED: Add or extend `internal/render/*_test.go` for Markdown envelope headers, Markdown help templates, four fallback levels, and unchanged `--format json` payloads for lookup/search.
- [x] 4.2 GREEN: Implement `internal/render/envelope.go`; update `internal/render/markdown.go`, `internal/render/search_markdown.go`, `internal/render/json.go`, and `internal/render/search_json.go` so Markdown gets envelope/help/fallbacks while JSON stays contract-compatible.
- [x] 4.3 REFACTOR: Centralize help/fallback template builders in `internal/render/*` and delete dead legacy render paths.
- [x] 4.4 Verify `go test ./...` passes after each RED → GREEN → REFACTOR slice, then run the `dlexa-go-cli-lint` workflow and update this file’s checkboxes as implementation lands.

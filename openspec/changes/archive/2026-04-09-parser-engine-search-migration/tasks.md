# Tasks: Parser Engine Search Migration

> Note: no separate design artifact was recoverable in Engram; this breakdown is derived from explore/proposal/spec artifacts plus verified code reality.

## Phase 1: RED — Wiring + Wrapper Failing Tests
- [ ] 1.1 RED: Update `internal/app/wiring_test.go` to fail unless `search` and `dpd` providers expose concrete `*parseengine.LiveSearchParser` / `*parseengine.DPDSearchParser` instead of `*parseengine.LegacySearchAdapter`.
- [ ] 1.2 RED: Add failing passthrough tests in `internal/parse/engine/search_port_test.go` for named live/DPD engine wrappers preserving `ParseInput` context, descriptor, document, records, warnings, and errors.
- [ ] 1.3 RED: Extend `internal/search/provider_test.go` to fail until `NewEnginePipelineProvider(...)` preserves normalized output and fetch document propagation with concrete engine-native wrappers.
- [ ] 1.4 RED: Add/adjust targeted regression cases in `internal/search/live_service_test.go` proving parse/normalize failure propagation stays unchanged after engine-native provider wiring.

## Phase 2: GREEN — Engine-Native Search Wrappers
- [ ] 2.1 GREEN: Create `internal/parse/engine/live_search.go` with `parseengine.LiveSearchParser` implementing `SearchParser` by delegating to `parse.NewLiveSearchParser()`.
- [ ] 2.2 GREEN: Create `internal/parse/engine/dpd_search.go` with `parseengine.DPDSearchParser` implementing `SearchParser` by delegating to `parse.NewDPDSearchParser()`.
- [ ] 2.3 GREEN: Keep `internal/parse/live_search.go` and `internal/parse/dpd_search.go` untouched as behavior sources; only add minimal constructor/test hooks if required for wrapper coverage.

## Phase 3: GREEN — Runtime Wiring Migration
- [ ] 3.1 GREEN: Update `internal/app/wiring.go` so search providers use `searchsvc.NewEnginePipelineProvider(...)` with the new engine-native wrapper constructors.
- [ ] 3.2 GREEN: Leave `internal/search/provider.go` and `internal/search/service.go` behavior unchanged unless a tiny additive test seam is strictly required.

## Phase 4: REFACTOR — Targeted Regression Coverage
- [ ] 4.1 REFACTOR: Run and fix targeted parser regressions in `internal/parse/live_search_test.go` and `internal/parse/dpd_search_test.go` to confirm zero output drift.
- [ ] 4.2 REFACTOR: Re-run `internal/search/provider_test.go`, `internal/search/live_service_test.go`, and `internal/app/wiring_test.go`; simplify duplicated wrapper assertions without weakening coverage.

## Phase 5: Verification
- [ ] 5.1 VERIFY: Run targeted package tests for `internal/parse/engine`, `internal/parse`, `internal/search`, and `internal/app`.
- [ ] 5.2 VERIFY: Run full repo tests with `go test ./...`.
- [ ] 5.3 VERIFY: Run full lint with `go tool --modfile=golangci-lint.mod golangci-lint run ./...` and fix only migration-related issues.

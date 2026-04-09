# Tasks: Parser Engine Foundation

> Note: no separate Engram design artifact was recoverable; this breakdown is derived from proposal/spec artifacts plus verified architecture/code reality.

## Phase 1: RED — Foundation Contracts
- [x] 1.1 RED: Add failing contract tests in `internal/parse/engine/article_port_test.go` for `ParseInput`, `ArticleResult`, and `ArticleParser` compatibility with current `parse.Result` flows.
- [x] 1.2 RED: Add failing contract tests in `internal/parse/engine/search_port_test.go` for `SearchParser` output/warning behavior with `parse.ParsedSearchRecord`.
- [x] 1.3 RED: Add failing resolver tests in `internal/parse/engine/resolver_test.go` covering register/resolve by parser family and source plus missing-parser failures.
- [x] 1.4 RED: Add failing bridge adapter tests in `internal/parse/engine/bridge_article_test.go` and `bridge_search_test.go` proving legacy parsers can be wrapped without input/output drift.

## Phase 2: GREEN — Engine Scaffolding and Bridges
- [x] 2.1 GREEN: Create `internal/parse/engine/input.go`, `article_ports.go`, and `search_ports.go` with additive contracts from ADR-0001 while leaving `internal/parse/interfaces.go` unchanged.
- [x] 2.2 GREEN: Implement `internal/parse/engine/resolver.go` and `registry.go` with explicit article/search registration and deterministic lookup errors.
- [x] 2.3 GREEN: Implement `internal/parse/engine/bridge_article.go` to adapt `parse.Parser` and `engine.ArticleParser` without altering warnings, miss handling, or article payloads.
- [x] 2.4 GREEN: Implement `internal/parse/engine/bridge_search.go` to adapt current search parsers and `engine.SearchParser` without record drift.

## Phase 3: REFACTOR — Minimal Wiring Adoption
- [x] 3.1 REFACTOR: Update `internal/source/pipeline.go` with an additive seam/constructor that can accept the article-engine adapter while preserving `NewPipelineSource(...)`.
- [x] 3.2 REFACTOR: Update `internal/search/provider.go` with an additive seam/constructor for the search-engine adapter while preserving `NewPipelineProvider(...)`.
- [x] 3.3 REFACTOR: Update `internal/app/wiring.go` and `internal/app/wiring_test.go` so runtime wiring proves engine bridges are present while concrete parser/normalizer choices remain unchanged.

## Phase 4: Verification
- [x] 4.1 RED/GREEN: Add targeted regression coverage in `internal/source/pipeline_test.go` and `internal/search/provider_test.go` for adapter-fed pipelines preserving descriptor, document, warnings, and normalized outputs.
- [x] 4.2 VERIFY: Run verification before commit and fix any contract or wiring regressions caused by the foundation slice.
- [x] 4.3 VERIFY: Run lint validation before commit and address issues without expanding scope.

## Completion Notes
- The implementation was committed by the orchestrator as `ee05d64` (`refactor(parse): add parser engine foundation scaffolding`).
- This slice remained behavior-preserving: it added scaffolding, adapters, and additive constructors without changing visible parsing semantics.

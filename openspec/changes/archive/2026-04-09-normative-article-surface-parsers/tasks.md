# Tasks: Normative Article Surface Parsers

> Corrected scope: this final slice implements only `espanol-al-dia`. `duda-linguistica` remains deferred and `noticia` stays out due to policy-gating risk.

## Phase 1: RED — Executable Surface Contracts
- [x] 1.1 RED: Add black-box CLI tests for `dlexa espanol-al-dia <slug>` routing, help, and syntax fallback in `cmd/dlexa/espanol_al_dia_test.go`.
- [x] 1.2 RED: Add parser tests for `internal/parse/espanol_al_dia.go` covering successful extraction and explicit broken-markup failure.
- [x] 1.3 RED: Add normalizer tests for `internal/normalize/espanol_al_dia.go` covering successful lookup entry generation and transform failure on empty sections.
- [x] 1.4 RED: Add fetcher tests for `internal/fetch/espanol_al_dia.go` covering URL construction and typed not-found handling.
- [x] 1.5 RED: Add wiring/module/search truthfulness tests proving `espanol-al-dia` is registered and rendered as executable, not deferred.

## Phase 2: GREEN — Surface Implementation
- [x] 2.1 GREEN: Implement `internal/fetch/EspanolAlDiaFetcher` targeting `/espanol-al-dia/<slug>`.
- [x] 2.2 GREEN: Implement `internal/parse.EspanolAlDiaParser` and its engine wrapper in `internal/parse/engine`.
- [x] 2.3 GREEN: Implement `internal/normalize.EspanolAlDiaNormalizer` reusing the shared article model and markdown-first rendering path.
- [x] 2.4 GREEN: Add `internal/modules/espanolaldia.Module` and top-level Cobra command `dlexa espanol-al-dia <slug>`.
- [x] 2.5 GREEN: Wire the new source and module in `internal/app/wiring.go`.

## Phase 3: REFACTOR — Search Truthfulness and Runtime Defaults
- [x] 3.1 REFACTOR: Update `internal/modules/search/filter.go` so mapped `espanol-al-dia` results are not marked deferred.
- [x] 3.2 REFACTOR: Update search help/renderer tests so executable `espanol-al-dia` guidance is distinguished from still-deferred surfaces like `noticia`.
- [x] 3.3 REFACTOR: Update `internal/app/app.go` so `espanol-al-dia` gets module-specific default sources instead of inheriting generic lookup defaults.

## Phase 4: VERIFY
- [x] 4.1 VERIFY: Run focused package tests for CLI, fetch, parse, normalize, module, and app wiring changes.
- [x] 4.2 VERIFY: Run `go test ./...`.
- [x] 4.3 VERIFY: Run `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.

## Completion Notes
- Verification passed with full test suite and repo lint clean.
- Main specs requiring durable sync: `openspec/specs/cli/spec.md` and `openspec/specs/search/spec.md`.

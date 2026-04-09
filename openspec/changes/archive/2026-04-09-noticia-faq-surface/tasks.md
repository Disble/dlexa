# Tasks: Noticia FAQ Surface

## Phase 1: RED
- [x] Add tests for simplified FAQ-prefix noticia rescue.
- [x] Add fetch/parse/normalize/module/CLI/wiring tests for the new `noticia` surface.

## Phase 2: GREEN
- [x] Simplify the noticia policy gate to the FAQ prefix.
- [x] Implement `noticia` fetcher, parser, engine wrapper, normalizer, module, and CLI command.
- [x] Wire the new source/module into the app runtime.

## Phase 3: REFACTOR
- [x] Update search truthfulness so rescued noticia FAQ results are executable.
- [x] Sync README and active specs to runtime truth.
- [x] Preserve module-level fallback for non-FAQ noticia slugs.

## Phase 4: VERIFY
- [x] Run focused changed-package tests.
- [x] Run `go test ./...`.
- [x] Run `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.

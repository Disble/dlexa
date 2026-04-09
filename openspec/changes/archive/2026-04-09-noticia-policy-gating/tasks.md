# Tasks: Noticia Policy Gating

## Phase 1: RED
- [x] Add tests for rescued noticia positives and institutional false positives.
- [x] Add explicit tests requiring FAQ gating plus linguistic signals.

## Phase 2: GREEN
- [x] Tighten `isRescuedNoticia` to require both FAQ gate and linguistic signals.

## Phase 3: REFACTOR
- [x] Reuse normalized token helpers instead of adding parser-specific logic.
- [x] Sync the search spec with the stricter policy gate.

## Phase 4: VERIFY
- [x] Run focused search/render tests.
- [x] Run `go test ./...`.
- [x] Run `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.

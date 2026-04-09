# Tasks: Duda Lingüística Article Surface

## Phase 1: RED
- [x] Add black-box CLI tests for `dlexa duda-linguistica <slug>`.
- [x] Add fetcher tests for URL construction and typed not-found handling.
- [x] Add parser tests for article extraction and broken markup failure.
- [x] Add normalizer tests for entry projection and empty-section rejection.
- [x] Add module/app/search truthfulness tests for registration and executable suggestions.

## Phase 2: GREEN
- [x] Implement `internal/fetch.DudaLinguisticaFetcher`.
- [x] Implement `internal/parse.DudaLinguisticaParser` and engine wrapper.
- [x] Implement `internal/normalize.DudaLinguisticaNormalizer`.
- [x] Implement `internal/modules/dudalinguistica.Module` and `cmd/dlexa/duda_linguistica.go`.
- [x] Wire the new source/module into `internal/app`.

## Phase 3: REFACTOR
- [x] Update search truthfulness so mapped `duda-linguistica` URLs are executable.
- [x] Update runtime docs/specs to reflect the new command surface.
- [x] Keep deferred treatment only for still-unimplemented surfaces such as `noticia`.

## Phase 4: VERIFY
- [x] Run focused changed-package tests.
- [x] Run `go test ./...`.
- [x] Run `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.

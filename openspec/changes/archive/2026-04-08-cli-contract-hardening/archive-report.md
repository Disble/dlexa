# Archive Report: cli-contract-hardening

## Change Summary

Added black-box tests for the Cobra command surface in `cmd/dlexa` and format validation at the `App.ExecuteModule` boundary.

## Key Outcomes

1. **Fixed critical .gitignore bug**: The pattern `dlexa` on line 2 was excluding the entire `cmd/dlexa/` directory from git tracking. Changed to `/dlexa` to only match the binary at the repo root. This unblocked tracking of `root.go`, `dpd.go`, `search.go`, and all test files.

2. **14 command surface tests**: `cmd/dlexa/root_test.go` (8 tests), `cmd/dlexa/dpd_test.go` (3 tests), `cmd/dlexa/search_test.go` (3 tests) now compile and pass, covering routing, flags, help, version, doctor, syntax errors.

3. **Format validation**: `App.ExecuteModule` now rejects unsupported `--format` values with a Nivel 1 Syntax fallback instead of letting them propagate as raw errors.

## Delta Specs Promoted

- CLI spec: Added "Format Validation at Runtime Boundary" and "Command Surface Black-Box Tests" requirements with scenarios.

## Files Changed

| File | Action |
|------|--------|
| `.gitignore` | Fixed `dlexa` → `/dlexa` binary pattern |
| `cmd/dlexa/root.go` | Tracked (was gitignored) |
| `cmd/dlexa/dpd.go` | Tracked (was gitignored) |
| `cmd/dlexa/search.go` | Tracked (was gitignored) |
| `cmd/dlexa/root_test.go` | NEW: stubRuntime mock + 8 routing tests |
| `cmd/dlexa/dpd_test.go` | Tracked (was gitignored): 3 DPD subcommand tests |
| `cmd/dlexa/search_test.go` | Tracked (was gitignored): 3 search subcommand tests |
| `internal/app/app.go` | Format validation in ExecuteModule |
| `internal/app/app_test.go` | TestExecuteModuleRejectsInvalidFormat |
| `openspec/specs/cli/spec.md` | Promoted delta specs |

## Commit

`2b1e39b` — `feat: add CLI contract tests and format validation`

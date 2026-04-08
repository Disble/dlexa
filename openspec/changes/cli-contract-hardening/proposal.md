# Proposal: cli-contract-hardening

## Intent

Add black-box tests for the Cobra command surface in `cmd/dlexa` and validate `--format` at the command/runtime boundary so that invalid formats produce structured Nivel 1 fallback instead of raw errors.

## Problem

1. **Zero command-layer tests**: `cmd/dlexa/` has no `*_test.go` files. Root routing, default-to-DPD, subcommand dispatch, `--help`, `--version`, `--doctor`, flag errors, and unknown-syntax detection are all untested at the Cobra surface. Every other layer has tests; this is the single largest coverage gap.
2. **Invalid `--format` escapes structured fallback**: The CLI accepts any string as `--format`. When the render registry cannot resolve it, the error propagates as a raw Go error rather than a Nivel 1 (Syntax) fallback with help guidance. The CLI spec requires syntax failures to show structured help.

## Scope

- **In scope**: Black-box tests for `executeRootCommand` covering routing, flags, help, version, doctor, syntax errors, and format validation. Format validation at the `App.ExecuteModule` boundary converting unknown formats into `HandleSyntaxError`.
- **Out of scope**: New subcommands, search deferred destinations, fallback copy i18n, DPD parser debt. No changes to module internals or rendering logic beyond format gating.

## Approach

1. Create `cmd/dlexa/root_test.go` with a mock `runtimeRunner` that records method calls and arguments, enabling deterministic assertion on routing behavior without touching real modules/network.
2. Test matrix covers: root with query → dpd, `dpd <q>` → dpd, `search <q>` → search, `--help`, `--version`, `--doctor`, missing args, unknown subcommand, invalid flags, `--format json`, `--format markdown`, `--format yaml` (invalid).
3. Add format validation in `App.ExecuteModule` after config resolution: if `req.Format` is not in the set of registered formats, return `HandleSyntaxError` with clear guidance.
4. Strict TDD: write failing tests first, then implement the minimum code to pass.

## Risks

- Mock-based command tests can drift from real runtime wiring; mitigate by keeping the mock interface identical to `runtimeRunner`.
- Format validation could break consumers that rely on empty format defaulting; mitigate by only gating after defaults are applied.

## Success Criteria

- `go test ./cmd/dlexa/...` passes with coverage over the full routing matrix.
- `--format yaml` (or any unsupported value) produces a Nivel 1 Syntax fallback, not a raw error.
- `go test ./...` and `golangci-lint run ./...` pass cleanly.

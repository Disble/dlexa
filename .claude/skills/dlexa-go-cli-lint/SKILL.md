---
name: dlexa-go-cli-lint
description: >
  Run and interpret the configured Go lint workflow for the dlexa CLI repository.
  Trigger: Load when reading, editing, reviewing, or validating Go code, packages, tests, or CLI behavior in this repo and lint feedback should shape the next change.
license: Apache-2.0
metadata:
  author: Disble
  version: "1.0"
---

## When to Use

- When touching any Go production code under `cmd/` or `internal/`.
- When changing Go tests and those changes affect package APIs, error handling, resource lifecycles, or allocations.
- When reviewing a Go diff and you need repo-accurate static analysis before calling the work done.
- When fixing bugs in code paths that may trigger `gosec`, `errcheck`, `bodyclose`, `staticcheck`, `govet`, `gocritic`, `revive`, or `prealloc`.
- Before finalizing a Go/CLI task in this repo. Do not skip lint because the change "looks small".

## Critical Patterns

- MUST use the repo-local tool entrypoint, not a globally installed binary: `go tool -modfile=golangci-lint.mod golangci-lint`.
- MUST lint after changing relevant Go code before claiming the task is complete.
- MUST prefer the smallest lint scope that still covers the changed behavior while iterating, then run broader lint when the change crosses packages or touches shared behavior.
- MUST run full repo lint when the change affects shared packages, exported APIs, cross-package flows, config used by many packages, or multiple packages at once.
- MUST re-run lint after fixing lint findings that can cascade into other warnings.
- MUST prioritize safety and correctness first: fix `gosec`, `govet`, `staticcheck`, `errcheck`, and `bodyclose` findings before style-only cleanup.
- MUST treat `bodyclose` and `errcheck` issues as real resource/correctness defects, not cosmetic warnings.
- MUST preserve performance-sensitive improvements surfaced by `prealloc` when they do not hurt readability or semantics.
- MUST NOT build as part of this skill. Use lint only.
- MUST NOT swap commands, flags, or config unless the repository itself changes its lint setup.

## Configured Linters

| Linter | Why it matters | What to watch for |
| --- | --- | --- |
| `gosec` | Security and dangerous patterns | Unsafe file/process/network handling, risky conversions, weak defaults |
| `govet` | Compiler-adjacent correctness checks | Suspicious API usage, bad printf-style calls, logic mistakes |
| `staticcheck` | Deep correctness and maintainability | Dead code, ineffective logic, misuse of stdlib APIs |
| `errcheck` | Error handling discipline | Ignored returned errors, especially IO and cleanup paths |
| `bodyclose` | Resource lifecycle safety | HTTP response bodies that are not closed on every path |
| `gocritic` | Robustness and code quality | Dubious constructs, accidental inefficiencies, fragile patterns |
| `revive` | Style with maintainability impact | Naming, structure, comments, and readability rules |
| `prealloc` | Allocation awareness | Slices that should be preallocated in hot or predictable paths |

## Lint Scope Decision Table

| Situation | Command | Why |
| --- | --- | --- |
| Editing one package only | `go tool -modfile=golangci-lint.mod golangci-lint run ./internal/query/...` | Fast feedback while iterating in one package boundary |
| Editing `cmd/dlexa` only | `go tool -modfile=golangci-lint.mod golangci-lint run ./cmd/dlexa/...` | Keeps iteration tight around CLI entrypoint changes |
| Editing 2-3 known packages | `go tool -modfile=golangci-lint.mod golangci-lint run ./internal/query/... ./internal/source/...` | Covers the touched package set without scanning the whole repo |
| Editing shared abstractions, exported APIs, or cross-package flows | `go tool -modfile=golangci-lint.mod golangci-lint run ./...` | Catches fallout across package boundaries |
| Finishing a Go task and ready to report done | `go tool -modfile=golangci-lint.mod golangci-lint run ./...` | Final repo-wide verification before handoff |
| Fixing lint findings after the first pass | Re-run the same scope first, then widen to `./...` if the change spread | Confirms the fix and catches secondary fallout |

## Simple Decision Tree

1. Did you change Go code outside one package? -> Run `./...`.
2. Did you change exactly one package and stay inside that boundary? -> Run that package path first.
3. Did the fix touch shared helpers, interfaces, exported symbols, or HTTP/resource handling? -> Escalate to `./...`.
4. Are you about to say the task is done? -> Run `./...` unless the task explicitly excluded final verification.

## Commands

```bash
# Full repository lint
go tool -modfile=golangci-lint.mod golangci-lint run ./...

# Lint one package under internal/
go tool -modfile=golangci-lint.mod golangci-lint run ./internal/query/...

# Lint multiple known packages
go tool -modfile=golangci-lint.mod golangci-lint run ./internal/query/... ./internal/source/...

# Lint the CLI entrypoint package
go tool -modfile=golangci-lint.mod golangci-lint run ./cmd/dlexa/...
```

## Operational Workflow

1. Map the changed Go files to package boundaries.
2. Run focused lint while iterating on the touched package or package set.
3. Fix findings in this order: security/correctness -> resource handling -> robustness -> maintainability -> performance.
4. Re-run the same focused scope until clean.
5. If the change crossed package boundaries or the task is complete, run `./...`.
6. Report which scope was run and whether it passed cleanly.

## Resources

- **Documentation**: See [references/lint-tooling.md](references/lint-tooling.md) for the local config files that define this workflow.

# Tasks: Architecture Filesystem Alignment FF

## Phase 1: Reality Inventory

- [x] 1.1 Add RED-style documentation assertions or review notes identifying the current runtime entrypoint files/folders that are actually present in the repo.
- [x] 1.2 Add RED-style review assertions identifying active docs/spec claims that point to missing or misleading file/folder architecture.
- [x] 1.3 Refactor the inventory into a concise gap map that distinguishes current truth, target state, and false active claims.

## Phase 2: Active Guidance Alignment

- [x] 2.1 Update `docs/architecture-formal-dlexa-v2.md` so actual-vs-target file/folder state is explicit and non-misleading.
- [x] 2.2 Update `docs/architecture_v2_oraculo.md` only where implementation claims overstate the real filesystem/runtime state.
- [x] 2.3 Update the active OpenSpec main specs so they no longer depend on nonexistent runtime files unless those files are introduced in this change.
- [x] 2.4 Keep `AGENTS.md` and `CLAUDE.md` aligned with the runtime-truth entrypoints.

## Phase 3: Filesystem / Entry Surface Closure

- [x] 3.1 Decide and implement the accepted file/folder architecture direction for the CLI/runtime entry surface.
- [x] 3.2 Add RED tests for any new or changed runtime entrypoint files/folders before production code changes.
- [x] 3.3 Make the smallest GREEN changes needed to align the repository layout with the accepted architecture direction.
- [x] 3.4 Refactor imports/helpers/tests so the final file/folder structure is coherent and maintainable.

## Phase 4: Verification and Artifact Closure

- [x] 4.1 Update apply-progress evidence with explicit RED/GREEN/REFACTOR, triangulation, and safety-net notes.
- [x] 4.2 Run focused `go test` commands on the affected runtime packages while iterating.
- [x] 4.3 Run `go test ./...` with no build step.
- [x] 4.4 Run `go tool --modfile=golangci-lint.mod golangci-lint run ./...` with no build step.
- [x] 4.5 Confirm that active docs/specs and the real filesystem layout now agree on the runtime-critical files and folders.

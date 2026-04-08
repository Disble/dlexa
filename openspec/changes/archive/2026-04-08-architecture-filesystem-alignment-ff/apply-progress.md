# Apply Progress: Architecture Filesystem Alignment FF

## Runtime-truth decision

Accepted direction: keep the existing thin `cmd/dlexa` Cobra surface, but treat `internal/app`, `internal/modules`, and `internal/render` as the execution/composition truth behind that CLI.

## Gap map

### Current truth

- `cmd/dlexa/main.go`, `root.go`, `dpd.go`, and `search.go` are present and form the active binary command surface.
- `internal/app/app.go` and `internal/app/wiring.go` remain the execution boundary and composition root.
- `internal/modules/interfaces.go`, `internal/modules/dpd/module.go`, and `internal/modules/search/module.go` define and implement the current module contracts.
- `internal/render/envelope.go`, `markdown.go`, `json.go`, `search_markdown.go`, and `search_json.go` define the active rendering/fallback surface.

### Misleading active claims corrected

- Main CLI spec no longer contains stale change-scoped wording such as “No New Destination Commands in This Change”.
- Main search spec now says literal next-step commands are guidance, not proof that every suggested destination is already registered.
- Formal architecture docs now distinguish current runtime state from future architecture goals instead of treating roadmap items as present-tense implementation.
- Agent guidance now points to both the thin `cmd/dlexa` entrypoint and the `internal/app` composition root.

## TDD / safety notes

- **Safety net**: no Go production files or Go tests were changed, so no focused Go safety-net test run was required to protect edited runtime files.
- **RED**: documentation review and the gap map captured the false/misleading claims before edits.
- **GREEN**: docs/specs/guidance now match the existing filesystem truth.
- **REFACTOR**: stale wording was simplified so active specs describe registered commands only and future-state content is explicitly framed as target/roadmap.

## Files changed in apply

- `openspec/specs/cli/spec.md`
- `openspec/specs/search/spec.md`
- `docs/architecture-formal-dlexa-v2.md`
- `docs/architecture_v2_oraculo.md`
- `AGENTS.md`
- `CLAUDE.md`
- `openspec/changes/architecture-filesystem-alignment-ff/tasks.md`
- `openspec/changes/architecture-filesystem-alignment-ff/state.yaml`

## Notes

- No runtime filesystem move was required because the thin `cmd/dlexa` surface is already materialized in the repo.
- No empty archive directories were removed during apply because none were identified as safe, empty cleanup targets in the inspected change tree.

## Verification executed in apply

- Focused tests: `go test ./cmd/dlexa/...` ✅, `go test ./internal/app/...` ✅
- Full tests: `go test ./...` ✅
- Full lint: `go tool --modfile=golangci-lint.mod golangci-lint run ./...` ✅

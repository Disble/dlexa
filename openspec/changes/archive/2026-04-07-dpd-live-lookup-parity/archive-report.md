# Archive Report: dpd-live-lookup-parity

## Change

- **Name**: `dpd-live-lookup-parity`
- **Archive date**: `2026-04-07`
- **Verify verdict**: `PASS WITH WARNINGS`
- **Critical blockers**: None

## Artifacts Reviewed

- `proposal.md` ✅
- `exploration.md` ✅
- `design.md` ✅
- `tasks.md` ✅
- `apply-progress.md` ✅
- `verify-report.md` ✅
- `state.yaml` ✅
- `specs/dpd/spec.md` ✅

## Spec Sync Summary

| Domain | Action | Details |
|---|---|---|
| `dpd` | Replaced | Promoted the full `dpd-live-lookup-parity` spec into `openspec/specs/dpd/spec.md` because this change supersedes the earlier renderer-only source of truth with the complete live-lookup, extraction, normalization, rendering, failure-taxonomy, and verification contract. |

## Verification Notes

- Repository-wide `go test ./...` passed.
- Repository-wide `go test -cover ./...` passed.
- Repository-wide `go tool --modfile=golangci-lint.mod golangci-lint run ./...` passed.
- Repository-wide `go vet ./...` passed.
- Opt-in live probe `DLEXA_LIVE_DPD_PROBE=1; go test ./internal/render -run TestDPDLiveProbeBienDriftInvariants -count=1 -v` passed.
- Two independent verification subagents corroborated the runtime-green outcome.
- Earlier verify warnings about stale OpenSpec bookkeeping were resolved before archive by updating `state.yaml` and adding explicit RED → GREEN → REFACTOR evidence to `apply-progress.md`.

## Archive Destination

- `openspec/changes/archive/2026-04-07-dpd-live-lookup-parity/`

## Source of Truth Updated

- `openspec/specs/dpd/spec.md`

## SDD Cycle Complete

`dpd-live-lookup-parity` is now planned, implemented, verified, and archived.

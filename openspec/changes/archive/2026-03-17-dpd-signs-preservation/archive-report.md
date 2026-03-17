# Archive Report: dpd-signs-preservation

## Status

Completed and archived on `2026-03-17`.

## Executive Summary

This change finalized DPD sign preservation by extending the existing semantic pipeline for validated signs, preserving bracket contexts in structured output, and documenting speculative sign handling and archived exclusions.

Verification passed with warnings: validated sign behavior, regression protection, targeted DPD tests, full `go test ./...`, and repo lint all passed; remaining warnings are limited to manual/static enforcement of speculative warning annotations and the intentionally unconfigured build step.

## Source Artifacts Reviewed

- Proposal: `openspec/changes/dpd-signs-preservation/proposal.md`
- Spec: `openspec/changes/dpd-signs-preservation/spec.md`
- Design: `openspec/changes/dpd-signs-preservation/design.md`
- Tasks: `openspec/changes/dpd-signs-preservation/tasks.md`
- Verify report: `openspec/changes/dpd-signs-preservation/verify-report.md`

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| `dpd` | Updated | Main spec now includes end-to-end validated DPD sign preservation, structured bracket-context distinction, and the non-authoritative status of speculative signs plus archived `<` / `>` exclusions. |

## Final Outcome

- 33 / 33 planned tasks completed.
- Validated sign support is archived as completed for `@`, `+`, `⊗`, `→`, and bracket contexts authored via `dfn`, `span.nn`, and `span.yy`.
- Speculative support for `*`, `‖`, and `//` remains intentionally documented as inferred rather than validated.
- `<` and `>` remain intentionally excluded because of HTML collision risk and low proven value.

## Verification Result

- Verdict: **PASS WITH WARNINGS**
- Full test suite: `go test ./...` ✅
- Targeted DPD sign verification ✅
- Targeted render verification ✅
- Repo lint (`golangci-lint`) ✅
- Build step: skipped by project rule / not configured

## Residual Warnings

1. Speculative warning annotations are still enforced primarily by code review/static inspection rather than a dedicated automated assertion.
2. `openspec/config.yaml` intentionally leaves `rules.verify.build_command` empty, so archive evidence does not include a build step.

## Archive Decision

This change is safe to archive because verification found no critical issues, all planned tasks are complete, and the repository source of truth has been updated to reflect the finalized behavior and remaining caveats.

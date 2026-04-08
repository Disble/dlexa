# Proposal: Architecture Filesystem Alignment FF

## Intent

Close the highest-impact architecture drift between the documented target layout and the real repository layout by making the file and folder structure truthful, reviewable, and enforceable.

## Problem Statement

The repository currently contains a real modular runtime in `internal/app`, `internal/modules`, `internal/render`, and the live DPD/search pipelines, but the docs and some historical OpenSpec artifacts have claimed a Cobra-based `cmd/dlexa` surface and related file layout that are not reliably aligned with the runtime truth future agents encounter.

That mismatch is damaging because it causes planning and verification work to target files and folders that may not exist or may not be the actual runtime entrypoints.

This change is a filesystem-and-architecture closure effort: reconcile the repository's file/folder layout and governing docs/specs so that future work starts from reality instead of drift.

## Goals

- Define the real current runtime file/folder entrypoints as the authoritative starting point.
- Close the most important file/folder architecture gaps between docs/specs and code.
- Decide explicitly whether the missing target CLI surface (`cmd/dlexa` + Cobra) will be implemented now or removed from active architectural claims until it exists.
- Leave the repository in a state where docs, specs, and runtime layout are mutually consistent.

## Non-Goals

- Adding new linguistic modules beyond what is needed for architecture/file-layout alignment.
- Rewriting the live DPD or live search business logic without a file-layout reason.
- Expanding product scope beyond closing documented architecture drift.

## Scope

### In Scope

- Compare target docs/specs against the actual file and folder layout.
- Introduce, move, remove, or rewrite files/folders as needed to make the runtime surface and architectural claims align.
- Reconcile the active OpenSpec main specs with the real runtime entrypoints.
- Update repo guidance so future agents know which files are runtime truth and which are target-state aspirations.

### Out of Scope

- New feature surfaces unrelated to architecture drift.
- Broad refactors that do not materially reduce layout/documentation inconsistency.
- Build workflow changes.

## Architectural Direction

This change MUST treat the codebase as runtime truth and use that truth to decide one of two valid outcomes:

1. implement the missing target file/folder architecture so the docs become true, or
2. downgrade target-only claims and formalize the current runtime layout as the active architecture.

Either outcome is acceptable. Remaining in a mixed state is not.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `docs/architecture-formal-dlexa-v2.md` | Modified | Clarify actual-vs-target file/folder state and remove misleading present-tense claims. |
| `docs/architecture_v2_oraculo.md` | Modified if needed | Keep vision narrative while avoiding false statements about implemented layout. |
| `AGENTS.md` / `CLAUDE.md` | Modified | Point future agents at the real runtime entrypoints and drift policy. |
| `openspec/specs/*.md` | Modified | Reconcile active specs with real runtime surface and accepted target-state declarations. |
| runtime entrypoint files/folders | Modified/New/Delete | Align actual filesystem layout with the chosen architectural direction. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Architecture cleanup turns into unfocused rewrite | Medium | Limit the change to file/folder truth and the runtime surface it governs. |
| Docs are corrected but code remains ambiguous | Medium | Require an explicit runtime-surface decision and verify it against files. |
| File moves create regressions | Medium | Use strict TDD and focused runtime regression coverage around the affected entrypoints. |

## Rollback Plan

If the alignment work overreaches, revert the specific file/folder and spec/doc edits that introduced the inconsistency, but do not restore known-false claims as if they were runtime truth.

## Success Criteria

- [ ] Active docs/specs no longer claim a file/folder architecture that the repo does not actually implement.
- [ ] The runtime entrypoints and key contracts are obvious from the repository layout itself.
- [ ] Future agents can start from `AGENTS.md` / `CLAUDE.md` and land on the correct runtime files immediately.
- [ ] The chosen architecture direction for the CLI/file layout is materially reflected in code, not only described in prose.

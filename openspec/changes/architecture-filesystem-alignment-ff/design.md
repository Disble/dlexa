# Design: Architecture Filesystem Alignment FF

## Technical Approach

This change closes architecture drift at the file and folder level.

The central design rule is simple:

> active docs/specs MUST describe the runtime surface that actually exists in the repository, unless the change also materializes the missing files/folders in the same implementation.

That means the work starts with a gap map and ends only when the file layout, the active specs, and the architecture docs point to the same operational truth.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Runtime truth source | Treat the repository layout and wired runtime files as the source of truth | Trust archived SDD claims over code | The code is what future agents execute against; anything else is planning fiction. |
| Drift handling | Either materialize missing files/folders now or remove their active claims | Keep mixed actual/target wording in active specs | Mixed truth is exactly what caused the current confusion. |
| Scope boundary | Focus on entrypoints, contracts, and architectural folders first | Rewrite every doc and historical note | We need actionable alignment, not indiscriminate document churn. |

## Boundary Decisions and Dependency Direction

The current runtime boundary must be verified first:

- `internal/app/*` for application runtime entrypoints
- `internal/modules/*` for module contracts and module implementations
- `internal/render/*` for envelope and output contracts
- `internal/fetch|parse|normalize|search|source` for live pipelines

If a new `cmd/dlexa` surface is introduced as part of this FF change, it must remain thin and delegate to those existing boundaries.

## File and Folder Alignment Strategy

1. inventory the actual runtime-critical files and folders;
2. inventory the active docs/spec claims about file/folder layout;
3. classify each mismatch:
   - already true
   - target only
   - false / misleading
4. implement the smallest set of file, folder, and doc/spec changes needed so active guidance becomes truthful.

## Affected Files

| File or Folder | Action | Description |
|---|---|---|
| `docs/architecture-formal-dlexa-v2.md` | Modify | Split actual state from target state clearly at file/folder level. |
| `docs/architecture_v2_oraculo.md` | Modify if needed | Preserve vision while removing misleading implementation claims. |
| `openspec/specs/cli/spec.md` | Modify | Align CLI/file-layout claims with actual runtime surface or newly materialized files. |
| `openspec/specs/search/spec.md` | Modify | Remove dependency on nonexistent file-layout assumptions. |
| `AGENTS.md` | Modify | Keep runtime-truth file pointers prominent. |
| `CLAUDE.md` | Modify | Keep runtime-truth file pointers prominent. |
| runtime entrypoint folders/files | Modify/New/Delete | Materialize the accepted architecture decision at filesystem level. |

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Runtime contracts still resolve through the chosen entrypoints | Focused tests on the affected entrypoint/runtime files |
| Integration | Wiring and module dispatch still work after file/folder alignment | `internal/app` integration tests and any new entrypoint tests if introduced |
| Regression | Docs/spec assumptions no longer point to nonexistent runtime files | Focused repo assertions plus reviewable artifact updates |

## Risks

- Entry-surface changes can spill into many files if we do not constrain the scope.
- Doc-only cleanup without runtime-surface resolution would leave the repo half-fixed.
- File moves can break tests or imports if not done incrementally under TDD.

## Rollout Notes

This change should proceed in small, reviewable increments:

1. reconcile active guidance,
2. align the runtime surface at the filesystem level,
3. re-run focused tests,
4. then run full verification.

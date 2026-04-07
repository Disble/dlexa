# Proposal: dlexa-v2-cobra-migration

## Intent

Migrate `dlexa` from stdlib `flag` to `spf13/cobra` to transform it into the definitive "Semantic Oracle of Spanish" for LLMs (Agentic RAG). This transition eliminates brittle flag parsing, formalizes subcommands, and optimizes output context to minimize token waste and eliminate LLM hallucinations when interacting with the CLI.

## Scope

### In Scope
- Create `cmd/dlexa/root.go`, `cmd/dlexa/search.go`, and `cmd/dlexa/dpd.go`.
- Encapsulate current DPD logic into `internal/modules/dpd`.
- Create a skeleton for `internal/modules/search` (semantic router).
- Implement the "Markdown Envelope" pattern for output context (`# [dlexa:module] Title ...`).
- Implement the 4-level Error Fallback (Syntax, 404, 503, Parse Error) formatted in Agent-Optimized Markdown.

### Out of Scope
- Implementing Goroutine Fan-Out/Fan-In concurrent searches (Phase 2).
- TUI implementations (explicitly rejected).
- Refactoring `internal/cache` for thread-safety.

## Approach

Replace `flag.NewFlagSet` in `internal/app/app.go` with a tree of Cobra commands (`root`, `search`, `dpd` as default). Create a standardized `EnvelopeRenderer` to wrap outputs in the Markdown Envelope. Overwrite Cobra's `SetHelpTemplate` to provide agent-friendly, copiable Markdown examples. Map error types to the explicit 4-level fallback system.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/dlexa/` | New | Cobra command initialization (`root`, `search`, `dpd`) |
| `internal/app/` | Modified/Removed | Current `flag` routing logic deprecated |
| `internal/modules/` | New | Encapsulation of `dpd` and skeleton for `search` |
| `internal/render/` | Modified | Implementation of Markdown Envelope and Error fallbacks |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Agent behavior regression | Medium | Maintain existing standard output JSON structure for `--format json` while changing Markdown default. |
| Over-fetching tokens | Low | Strictly enforce URL heuristics to drop institutional news before returning to the LLM. |

## Rollback Plan

Revert the commit migrating to `spf13/cobra`, restoring `internal/app/app.go` to use stdlib `flag`. Ensure the previous version tag (e.g., `v1.x.x`) remains available for agents that pin dependencies while this migration is stabilized.

## Success Criteria

- [ ] Cobra commands (`dlexa`, `dlexa search`, `dlexa dpd`) execute successfully.
- [ ] DPD logic is completely decoupled into `internal/modules/dpd`.
- [ ] CLI help (`--help`) renders in Markdown.
- [ ] Outputs feature the Markdown Envelope (`# [dlexa:nombre-modulo] ...`).
- [ ] 4-level error fallbacks are triggered and rendered accurately for their respective failure states.
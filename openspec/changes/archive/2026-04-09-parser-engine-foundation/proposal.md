# Proposal: parser-engine-foundation

## Intent
Establish the base scaffolding for the new Parser Engine defined in ADR-0001. This creates a shared engine umbrella and clear anti-corruption boundaries without risking behavior changes during the initial transition.

## Scope

### In Scope
- Introduce the common `ParseInput` envelope.
- Define the separate `ArticleParser` and `SearchParser` ports.
- Add parser resolver/registry scaffolding.
- Implement bridge adapters to ensure existing pipelines continue working seamlessly.

### Out of Scope
- Full migration of all concrete parsers (`DPDArticleParser`, `LiveSearchParser`, etc.) to new packages.
- Modifications to any existing behavior or parsing contracts.
- Introduction of product-level filtering.

## Approach
Create the foundation in `internal/parse/engine` with the defined `ParseInput` struct and the two family interfaces (`ArticleParser`, `SearchParser`). Build a simple `Resolver` registry. To prevent breaking existing consumers (e.g., `LookupService`), create bridge adapters that wrap or implement the current interfaces while integrating with the new engine types where necessary, keeping the transition transparent.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/parse/engine/` | New | Shared interfaces, input envelope, and resolver. |
| `internal/parse/` | Modified | Addition of bridge adapters for existing parsers. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Adapter mistranslation | Medium | Write focused unit tests for bridge adapters to ensure exact input/output fidelity. |
| Over-abstraction | Low | Strictly limit this PR to interface and registry definitions; defer actual parser logic moves. |

## Rollback Plan
Since this change only introduces scaffolding and transparent adapters, rollback is a direct git revert of the foundation commit.

## Dependencies
- ADR-0001 (Parser Engine Architecture)

## Success Criteria
- [ ] `ParseInput`, `ArticleParser`, and `SearchParser` types compile and align with ADR-0001.
- [ ] Parser resolver scaffolding exists and can register/retrieve parsers.
- [ ] Bridge adapters allow the system to compile without changing service-layer code.
- [ ] All existing tests pass with zero behavioral drift.

# Proposal: parser-engine-search-migration

## Intent
Migrate the existing search-family parsers (`LiveSearchParser` and `DPDSearchParser`) into the new, explicit `parser-engine` architecture without altering their runtime behavior, fulfilling step 2 of the parser engine roadmap.

## Scope

### In Scope
- Migrate search-family parsers to explicit parser-engine search family structure (`SearchParser` port).
- Wire the runtime to use these engine-native search parser implementations.
- Maintain existing search behavior and output fidelity exactly.

### Out of Scope
- Migrating the article-family parsers (e.g., `DPDArticleParser`).
- Rewriting the error/problem taxonomy unless strictly required for the search parsers to compile.
- Altering the product policy, ranking, or filtering rules.

## Approach
1. Implement the `SearchParser` interface defined in `parser-engine-foundation` for the live search and DPD search surfaces.
2. Move the existing parsing logic from the legacy search parsers into these new engine-native implementations.
3. Update the `SearchService` (or equivalent module composition root) to resolve and invoke the `SearchParser` via the engine's registry/resolver.
4. Remove the legacy search parser adapters if they are no longer needed, or bypass them.
5. Ensure all existing search tests pass without modification to the assertions.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| `internal/parse/engine/search/` | New | Engine-native search parser implementations |
| `internal/parse/` | Modified/Removed | Legacy search parsers deprecated/removed |
| `internal/search/service.go` | Modified | Wired to use the new parser engine resolver |
| `internal/app/wiring.go` | Modified | Registration of the new search parsers |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Unintended behavior change in search results | Low | Rely on existing comprehensive test suite; no logic changes, only structural moves. |
| Wiring failures in the CLI runtime | Low | Test CLI commands manually/integration tests to ensure the resolver correctly provides the parser. |

## Rollback Plan
Revert the commits that change `internal/search/service.go` and `internal/app/wiring.go` to point back to the legacy parsers or adapters, and remove the new engine-native search implementations.

## Dependencies
- `parser-engine-foundation` (Must be completed and merged first, providing the `SearchParser` port and resolver).

## Success Criteria
- [ ] `LiveSearchParser` and `DPDSearchParser` logic lives under the `parser-engine` structure.
- [ ] Application runtime wires the `SearchParser` via the engine resolver.
- [ ] All existing search-related tests pass without changes to expected data.
- [ ] Article parsers remain untouched.

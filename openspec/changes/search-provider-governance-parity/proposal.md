# Proposal: search-provider-governance-parity

## Intent
Unify rate-limit (HTTP 429) handling and resilience across search providers by applying the governed doer to the DPD search fetcher.

## Scope

### In Scope
- Wrap `fetch.NewDPDSearchFetcher(...)` with `fetch.NewGovernedDoer(...)` in `internal/app/wiring.go`.
- Ensure `internal/fetch/dpd_search.go` propagates 429 status codes correctly rather than masking them as generic `ProblemCodeDPDSearchFetchFailed`.
- Update tests to verify rate-limit semantics.

### Out of Scope
- Refactoring the core Governance or Rate-Limit engine.
- Modifying other fetchers beyond DPD search.

## Approach
1. Modify `internal/fetch/dpd_search.go` to explicitly parse HTTP 429 status and map it to rate-limit problem codes (or ensure GovernedDoer detects it before masking).
2. Update `internal/app/wiring.go` to wrap the DPD search fetcher initialization with `fetch.NewGovernedDoer`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/app/wiring.go` | Modified | Wrap DPD fetcher with governed doer |
| `internal/fetch/dpd_search.go` | Modified | Explicit HTTP 429 classification |
| `internal/fetch/dpd_search_test.go` | Modified | Add rate-limit parsing tests |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| 429s mapped incorrectly | Low | Add explicit unit tests for 429 status responses |
| Double-governance | Low | Verify `wiring.go` applies governor layer correctly |

## Rollback Plan
Revert changes to `wiring.go` and `dpd_search.go`.

## Dependencies
- Existing `GovernedDoer` implementation in `internal/fetch`.

## Success Criteria
- [ ] DPD search fetcher handles HTTP 429s explicitly.
- [ ] DPD search uses `GovernedDoer` in wiring.
- [ ] Tests pass for 429 response scenarios.
# Proposal: Search Rate Limit Fallback Semantics

## Intent

The `live_search` fetcher currently returns a generic `ProblemCodeDPDSearchFetchFailed` for HTTP 429 responses and cooldown triggers, which masks the specific cause of failure as `FallbackKindUpstreamUnavailable`. We need explicit taxonomy to correctly handle and communicate rate-limit fallbacks as defined in the search spec.

## Scope

### In Scope
- Add explicit `ProblemCode` for rate limits/cooldowns.
- Add specific `FallbackKind` for rate limiting.
- Update `internal/fetch/live_search.go` to inspect errors from `governed_doer` and return the new rate-limit problem code on HTTP 429 or cooldown rejection.
- Update `internal/modules/interfaces.go` and module mapping to translate the new problem code to the rate-limit fallback kind.
- Unit testing for the new fallback semantics and error mappings.

### Out of Scope
- Modifying the underlying cooldown/backoff algorithm in `internal/fetch/governed_doer.go`.
- Modifying renderers or downstream UI output formatting beyond providing the correct fallback kind.

## Approach

1. **Taxonomy**: Introduce `ProblemCodeRateLimited` and `FallbackKindRateLimited` in application model/contracts.
2. **Fetch inspection**: Modify `live_search.go` to unwrap or check errors from the `governed_doer` (e.g., HTTP 429 status code or explicit cooldown errors) and return the new problem code instead of the generic fetch failure.
3. **Module Wiring**: Update the search module error mapping to map `ProblemCodeRateLimited` to `FallbackKindRateLimited`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/fetch/live_search.go` | Modified | Return rate-limit specific error codes |
| `internal/modules/interfaces.go` | Modified | Add/map new fallback taxonomy |
| `internal/model/types.go` / `search.go` | Modified | Add `ProblemCode` and `FallbackKind` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Overriding general network errors | Low | Strict error unwrapping to match *only* 429 and specific cooldown sentinels. |
| Test fragility | Low | Assert specific errors rather than string matching in tests. |

## Rollback Plan

Revert the taxonomy additions and restore `live_search.go` to return `ProblemCodeDPDSearchFetchFailed` for all fetch errors.

## Dependencies

- Existing `governed_doer` implementation must reliably surface 429 or cooldown errors.

## Success Criteria

- [ ] HTTP 429 responses explicitly result in a rate-limit fallback kind.
- [ ] Cooldown-triggered rejections explicitly result in a rate-limit fallback kind.
- [ ] Existing non-rate-limit network errors continue to map to `FallbackKindUpstreamUnavailable`.
- [ ] Tests verify the explicit taxonomy and fallback mapping.

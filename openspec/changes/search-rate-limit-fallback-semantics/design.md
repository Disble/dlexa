# Design: Search Rate Limit Fallback Semantics

## Technical Approach

Add a narrow rate-limit classification path at the fetch boundary, preserve the existing search fan-out and partial-success behavior, and only specialize the all-failed aggregation and fallback mapping. The cooldown algorithm in `internal/fetch/governed_doer.go` stays unchanged; the change is about surfacing its semantics through existing `Problem` and `FallbackEnvelope` contracts.

## Architecture Decisions

### Decision: Add explicit rate-limit taxonomy

**Choice**: Add provider-level `ProblemCodeDPDSearchRateLimited`, aggregate-level `ProblemCodeSearchAllProvidersRateLimited`, and `FallbackKindRateLimited`.
**Alternatives considered**: Reuse `ProblemCodeDPDSearchFetchFailed`; reuse `FallbackKindUpstreamUnavailable` with a different message.
**Rationale**: Generic fetch failure hides the actionable difference between “upstream is down” and “upstream asked us to stop”. A dedicated code/kind is the smallest contract change that makes retry guidance correct without changing result shapes or provider interfaces.

### Decision: Classify rate limits inside `live_search.go`

**Choice**: Detect both `*fetch.RateLimitCooldownError` and HTTP `429` in `LiveSearchFetcher.Fetch` and emit the new rate-limit problem code there.
**Alternatives considered**: Infer rate limits later in `search.Service`; add special wrapper types in providers.
**Rationale**: The fetcher already sees the raw transport error and status code. Classifying here keeps anti-corruption at the external boundary and avoids broad rewrites in search orchestration.

### Decision: Keep aggregation shape, only specialize all-rate-limited failure

**Choice**: Keep `SearchResult.Problems` and partial-success behavior as-is; only change the all-failed path so homogeneous rate-limited failures return `ProblemCodeSearchAllProvidersRateLimited`.
**Alternatives considered**: Add new aggregation DTOs; fail mixed-success requests; rewrite provider outcome handling.
**Rationale**: `internal/search/service.go` already preserves provider-attributed problems and returns success when any provider succeeds. That behavior already matches the requirement for mixed outcomes, so only the terminal failure classification needs to change.

## Data Flow

`search.Module` → `search.Service` → `Provider.Search` → `fetch.LiveSearchFetcher` → `GovernedDoer`

- `GovernedDoer` returns HTTP 429 or `*RateLimitCooldownError`.
- `LiveSearchFetcher` maps either case to `ProblemCodeDPDSearchRateLimited`.
- `search.Service` keeps provider-attributed rate-limit problems in `SearchResult.Problems` when any provider succeeds.
- If every provider failed and every problem is rate-limited, `search.Service` returns `ProblemCodeSearchAllProvidersRateLimited`.
- `modules.FallbackFromError` maps both rate-limit codes to `FallbackKindRateLimited`.
- `render/envelope.go` renders `FallbackKindRateLimited` as a Level 3 fallback (same recovery tier, more specific label/message).

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/model/types.go` | Modify | Add new problem codes and `FallbackKindRateLimited`; keep `FallbackEnvelope` shape unchanged. |
| `internal/fetch/live_search.go` | Modify | Add a small classifier for cooldown errors and HTTP 429 responses. |
| `internal/search/service.go` | Modify | Prefer `search_all_providers_rate_limited` when all provider problems are rate-limited; leave mixed-success flow unchanged. |
| `internal/modules/interfaces.go` | Modify | Map new rate-limit codes to explicit fallback text/kind; keep generic fetch failures on `UpstreamUnavailable`. |
| `internal/render/envelope.go` | Modify | Add label/level handling for `FallbackKindRateLimited` as Level 3. |
| `internal/fetch/live_search_test.go` | Modify | Assert 429 and cooldown now produce the new rate-limit problem code. |
| `internal/search/service_test.go` | Modify | Cover partial success with provider rate-limit metadata and all-provider rate-limited failure. |
| `internal/modules/interfaces_test.go` | Modify | Cover fallback mapping for rate-limited vs generic upstream failure. |
| `internal/render/envelope_test.go` | Modify | Cover markdown/json rendering for the new fallback kind. |

## Interfaces / Contracts

```go
const (
    ProblemCodeDPDSearchRateLimited         = "dpd_search_rate_limited"
    ProblemCodeSearchAllProvidersRateLimited = "search_all_providers_rate_limited"
    FallbackKindRateLimited                 = "rate_limited"
)
```

No new result structs are needed. Existing `model.Problem`, `model.SearchResult.Problems`, and `model.FallbackEnvelope` carry the new semantics.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Fetcher classification | Update `live_search_test.go` for HTTP 429 and active cooldown; keep transport/network errors on generic fetch failure. |
| Unit | Aggregation semantics | Add service tests for mixed success + rate-limited provider problem, and all-provider rate-limited top-level failure. |
| Unit | Fallback mapping | Add module/interface tests proving rate-limited codes map to `FallbackKindRateLimited` while generic fetch errors stay `FallbackKindUpstreamUnavailable`. |
| Unit | Rendering | Add envelope test for `rate_limited` label and Level 3 output. |

## Migration / Rollout

No migration required.

## Open Questions

- [ ] None.

# Design: Search Provider Governance Parity

## Technical Approach

Apply the existing `fetch.GovernedDoer` transport wrapper to the DPD search provider and align `DPDSearchFetcher.Fetch` with the same rate-limit taxonomy already used by `LiveSearchFetcher`. The change stays inside the existing `fetch -> parse -> normalize -> search` pipeline and only touches provider construction, fetch-time classification, and tests.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Reuse 429 governance | Wrap `DPDSearchFetcher.Client` with `NewGovernedDoer` in `internal/app/wiring.go` | Duplicate cooldown logic inside `DPDSearchFetcher`; add a DPD-only governor | `GovernedDoer` is already reusable and bounded; parity should come from composition, not a second mechanism. |
| Reuse rate-limit taxonomy | Map both `*RateLimitCooldownError` and HTTP 429 to `model.ProblemCodeDPDSearchRateLimited` in `internal/fetch/dpd_search.go` | Keep generic fetch-failure classification; invent a provider-specific code | Search already exposes the desired contract; DPD provider should emit the same problem class for federated partial-failure handling. |
| Keep config surface unchanged | Reuse `runtimeConfig.Search.Governance` for both `search` and `dpd` providers | Introduce a second governance config block for DPD provider | User requested minimal change. Existing search governance already represents upstream search-provider protection. |

## Data Flow

`dlexa search` → `internal/search.Service` → `dpd` provider  
→ `DPDSearchFetcher.Fetch` → `GovernedDoer.Do` → upstream `/srv/keys`

Outcomes:

- upstream `429` response → fetcher returns `ProblemCodeDPDSearchRateLimited`
- active cooldown before request → `GovernedDoer` returns `RateLimitCooldownError` → fetcher translates to `ProblemCodeDPDSearchRateLimited`
- other transport/status failures → existing `ProblemCodeDPDSearchFetchFailed`

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/fetch/dpd_search.go` | Modify | Mirror `live_search.go` classification: detect `RateLimitCooldownError`, special-case HTTP 429, preserve generic fetch-failure handling for all other failures. |
| `internal/app/wiring.go` | Modify | Construct `DPDSearchFetcher`, then wrap its `Client` with `fetch.NewGovernedDoer(...)` using the same `runtimeConfig.Search.Governance` values already applied to the general search fetcher. |
| `internal/fetch/dpd_search_test.go` | Modify | Add targeted coverage for direct 429 classification and cooldown-error translation without regressing generic challenge/transport failure behavior. |
| `internal/app/wiring_test.go` | Modify | Assert the wired DPD provider still uses `*fetch.DPDSearchFetcher` and that its `Client` is a `*fetch.GovernedDoer`. |

## Interfaces / Contracts

No new public interfaces or config fields.

Behavioral contract in `DPDSearchFetcher.Fetch` becomes:

```go
if err := client.Do(...); errors.As(err, &cooldownErr) {
    return rateLimitedProblem
}
if resp.StatusCode == http.StatusTooManyRequests {
    return rateLimitedProblem
}
return fetchFailedProblem
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | DPD fetcher classifies transport cooldown as rate-limited | Extend `internal/fetch/dpd_search_test.go` with a governed client that first receives 429, then returns cooldown on the next call. |
| Unit | DPD fetcher classifies direct HTTP 429 as rate-limited | Stub `Doer` returning `StatusTooManyRequests`; assert `ProblemCodeDPDSearchRateLimited`. |
| Unit | Non-429 failures stay generic | Keep existing challenge and transport-failure assertions unchanged. |
| Wiring | DPD provider gets governed transport | Extend `internal/app/wiring_test.go` to inspect the `dpd` search provider fetcher and its wrapped client. |

## Migration / Rollout

No migration required. This is a composition-root and fetch-classification change only.

## Open Questions

- [ ] None.

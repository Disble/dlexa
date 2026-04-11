# Tasks: Search Provider Governance Parity

## Phase 1: RED

- [ ] 1.1 Add failing tests in `internal/fetch/dpd_search_test.go` proving DPD search classifies upstream `429` and governed cooldown errors as `dpd_search_rate_limited`, not generic fetch failures.
- [ ] 1.2 Add failing wiring coverage in `internal/app/wiring_test.go` proving the DPD search provider uses the same governed HTTP doer policy as live search and preserves `dpd` provider identity/priority.
- [ ] 1.3 If current coverage is missing, add a search-service regression in `internal/search/gateway_test.go` for federated partial success: one provider rate-limited, one provider successful, result stays partial-success with preserved candidates/problems.

## Phase 2: GREEN

- [ ] 2.1 Update `internal/fetch/dpd_search.go` so transport errors caused by governed cooldowns and direct upstream `429` responses map to `model.ProblemCodeDPDSearchRateLimited` with truthful DPD-specific messages.
- [ ] 2.2 Update `internal/app/wiring.go` to wrap the DPD search fetcher client with the same `fetch.NewGovernedDoer(...)` governance config already used by live search.
- [ ] 2.3 Keep search provider registration/wiring parity intact in `internal/app/wiring.go`: federated registry order, default provider behavior, and DPD search descriptor semantics must remain stable.

## Phase 3: REFACTOR

- [ ] 3.1 Extract any duplicated governance setup in `internal/app/wiring.go` into a small local helper so live search and DPD search cannot drift again.
- [ ] 3.2 Tighten test helpers/assertions in `internal/fetch/dpd_search_test.go` and `internal/app/wiring_test.go` so rate-limit vs fetch-failure intent is explicit and future regressions fail loudly.

## Phase 4: VERIFY

- [ ] 4.1 Run focused tests for the touched areas: `go test ./internal/fetch ./internal/app ./internal/search`.
- [ ] 4.2 Run full regression suite: `go test ./...`.
- [ ] 4.3 Run repo lint exactly as configured: `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.

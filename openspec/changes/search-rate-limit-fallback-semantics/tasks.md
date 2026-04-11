# Tasks: Search Rate Limit Fallback Semantics

## Phase 1: RED
- [ ] 1.1 Add failing tests in `internal/fetch/live_search_test.go` for HTTP 429 and `RateLimitCooldownError`, asserting distinct rate-limit problem codes instead of `ProblemCodeDPDSearchFetchFailed`.
- [ ] 1.2 Add failing tests in `internal/modules/search/module_test.go` for rate-limit fallback mapping, preserving `FallbackKindUpstreamUnavailable` for non-rate-limit fetch failures.
- [ ] 1.3 Add failing aggregation tests in `internal/search/gateway_test.go` for all-providers-rate-limited top-level errors and partial-success federation when one provider rate-limits but another returns candidates.

## Phase 2: GREEN
- [ ] 2.1 Extend `internal/model/types.go` with explicit search rate-limit problem code(s) and `FallbackKindRateLimited` without changing existing parse/not-found taxonomy.
- [ ] 2.2 Update `internal/fetch/live_search.go` to detect HTTP 429 responses and `RateLimitCooldownError` from `governed_doer`, returning the new rate-limit problem code(s) only for those cases.
- [ ] 2.3 Update `internal/modules/interfaces.go` so `FallbackFromError` maps the new rate-limit code(s) to a dedicated rate-limit fallback message/suggestion, while generic fetch failures still map to upstream-unavailable.
- [ ] 2.4 Update `internal/search/service.go` aggregation so all-rate-limited provider failures preserve a deterministic rate-limit problem code, but mixed outcomes still return successful candidates plus provider problems.

## Phase 3: REFACTOR
- [ ] 3.1 Refactor duplicated rate-limit problem assertions into shared helpers/constants in `internal/fetch/live_search_test.go`, `internal/modules/search/module_test.go`, and `internal/search/gateway_test.go`.
- [ ] 3.2 Tighten any small model/search comments or helper naming touched by the change so rate-limit vs generic upstream semantics stay explicit at the fetch→aggregate→module boundary.

## Phase 4: VERIFY
- [ ] 4.1 Run focused verification for changed packages: `go test ./internal/fetch ./internal/modules/search ./internal/search`.
- [ ] 4.2 Run full regression suite with `go test ./...`.
- [ ] 4.3 Run full repo lint with `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.

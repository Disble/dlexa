# Delta for Search

## ADDED Requirements

### Requirement: Rate Limit Fallback Taxonomy

The search module MUST distinguish between a rate-limited rejection and a generic upstream unavailable error in its fallback semantics.

#### Scenario: Differentiating rate limits from generic failures

- GIVEN a search request fails
- WHEN the failure is due to HTTP 429 or an active cooldown
- THEN the module MUST return a specific `RateLimited` error or fallback
- AND MUST NOT conflate it with a generic `UpstreamUnavailable` error

### Requirement: Active Governed Cooldown Rejection

The search module MUST enforce an active cooldown period after detecting a rate limit, rejecting requests locally before they reach the upstream.

#### Scenario: Refusing follow-up requests during cooldown

- GIVEN the search module has recently received an HTTP 429 from a provider
- AND the provider is currently in an active governed cooldown period
- WHEN a new search request targets that provider
- THEN the module MUST reject the request locally
- AND return a `RateLimited` error without making an upstream HTTP request

## MODIFIED Requirements

### Requirement: Resilient Upstream Fetching

The live search fetcher MUST detect HTTP 429 (Too Many Requests) to protect against IP bans, and expose this specific failure reason to the routing and fallback layers.
(Previously: The live search fetcher MUST implement a ban-aware mechanism (e.g., backoff or circuit breaker) that detects HTTP 429 (Too Many Requests) or temporary upstream rejections, safely protecting against IP bans.)

#### Scenario: Upstream returns HTTP 429

- GIVEN the live search fetcher sends a request to the upstream search provider
- WHEN the upstream responds with HTTP 429 Too Many Requests
- THEN the fetcher MUST detect the rate limit
- AND trigger the governed cooldown mechanism
- AND return an explicit rate-limited failure to the caller

### Requirement: Graceful Partial Failure

The `search` module MUST return a successful aggregated response when at least one selected provider succeeds, and MUST fail the request only when every selected provider fails. It MUST preserve provider-specific failure reasons, including rate limits.
(Previously: The `search` module MUST return a successful aggregated response when at least one selected provider succeeds, and MUST fail the request only when every selected provider fails.)

#### Scenario: Mixed-provider partial success with rate limits

- GIVEN a search request targets multiple providers
- WHEN one provider fails with a rate limit (HTTP 429 or active cooldown)
- AND at least one other provider succeeds
- THEN the module MUST return the successful results
- AND MUST include the rate-limited status as provider-attributed problem metadata
- AND the overall request MUST NOT fail

#### Scenario: All providers rate-limited

- GIVEN a search request targets multiple providers
- WHEN all providers fail due to being rate-limited
- THEN the module MUST return an explicit top-level `AllProvidersRateLimited` failure
- AND MUST preserve provider-attributed detail for each rate-limited provider

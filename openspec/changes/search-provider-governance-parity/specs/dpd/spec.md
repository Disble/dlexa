# Delta for DPD Provider Search Parity

## ADDED Requirements

### Requirement: DPD Rate Limit Detection and Backoff
The DPD `/srv/keys` provider MUST detect HTTP 429 Too Many Requests responses and explicitly classify them as rate-limit errors, applying backoff/cooldown mechanisms identical to the general search provider.

#### Scenario: DPD search receives HTTP 429
- GIVEN the DPD upstream API returns an HTTP 429 status code
- WHEN the DPD provider attempts a fetch
- THEN the system MUST return a rate-limit error
- AND the provider MUST immediately transition into a cooldown state

#### Scenario: DPD search cooldown rejects follow-up request
- GIVEN the DPD provider is in an active cooldown state
- WHEN a subsequent request attempts to fetch from the DPD provider
- THEN the provider MUST instantly reject the request with a rate-limit error
- AND the provider MUST NOT dispatch an HTTP request to the upstream API

### Requirement: Standard Transport Failures Remain Generic
The DPD `/srv/keys` provider MUST continue to classify standard transport failures (such as DNS resolution failures, connection timeouts, or HTTP 5xx responses) as generic fetch failures, NOT rate-limit errors.

#### Scenario: DPD search standard transport failure remains generic fetch failure
- GIVEN the DPD upstream API experiences a timeout or returns an HTTP 5xx error
- WHEN the DPD provider attempts a fetch
- THEN the system MUST return a generic fetch failure error
- AND the provider MUST NOT enter a rate-limit cooldown state

### Requirement: Resilient Federated Search Responses
Federated search requests querying multiple providers MUST support partial success responses. When one provider fails (e.g., due to rate limits) but another succeeds, the overall request MUST NOT fail completely.

#### Scenario: Federated mixed-provider request where dpd is rate-limited but search succeeds
- GIVEN a federated search request targeting both the `dpd` and `search` providers
- AND the `search` provider responds successfully with results
- AND the `dpd` provider fails with a rate-limit error
- WHEN the search execution completes
- THEN the system MUST return a partial success response containing the results from the `search` provider
- AND the response MUST include an in-band provider problem indicating that the `dpd` provider was rate-limited
# Architecture Specification

## Purpose

Defines the repository layout policy, establishing that active documentation and specifications MUST reflect the true runtime filesystem and entrypoint architecture.

## Requirements

### Requirement: Runtime Truth Policy

The repository layout and wired runtime files MUST be treated as the authoritative source of architectural truth.

#### Scenario: Docs mismatch runtime

- GIVEN an architectural document claims a specific entrypoint
- WHEN that entrypoint does not exist in the codebase
- THEN the document MUST be updated to reflect the real entrypoint (e.g., `internal/app/app.go`)

### Requirement: Drift Prevention

Active specifications MUST NOT claim file or folder structures that do not exist, unless they are being materialized in the same change.

#### Scenario: Proposing new folder structure

- GIVEN a spec proposes a new architectural layout
- WHEN the implementation phase is complete
- THEN the missing folders MUST be materialized in the filesystem
- AND if they are not materialized, the spec MUST be reverted or archived

### Requirement: Future-Agent Guidance

Agent guidance files MUST direct agents to the real runtime entrypoints.

#### Scenario: Agent onboarded

- GIVEN a new AI agent reads `AGENTS.md`
- WHEN the agent searches for the application composition root
- THEN the guidance MUST point to `internal/app/app.go` and `internal/app/wiring.go`

### Requirement: Cache-Aside Runtime Topology

The runtime architecture MUST keep caching as an explicit cache-aside boundary with separate typed stores for lookup and search flows.

#### Scenario: Runtime wiring selects cache boundaries

- GIVEN the application composition root wires runtime services
- WHEN lookup and search services are constructed
- THEN lookup MUST depend on a lookup cache contract for `model.LookupResult`
- AND search MUST depend on a search cache contract for `model.SearchResult`
- AND the architecture MUST preserve the semantic split between both cache paths instead of collapsing them into one untyped public store

#### Scenario: Default cache backend selection stays resilient

- GIVEN the application starts on an environment with an available user cache directory
- WHEN cache backends are resolved from the composition root
- THEN the runtime MUST prefer filesystem-backed caches
- AND when that directory cannot be resolved, the runtime MUST fall back to in-memory caches instead of failing startup

### Requirement: Cache Failures Remain Non-Fatal Optimizations

The architecture MUST treat cache access as a best-effort optimization rather than a user-visible dependency.

#### Scenario: Cache read or write fails during a runtime request

- GIVEN a lookup or search request reaches the cache boundary
- WHEN cache `Get` or `Set` cannot provide a usable result because of corruption, expiry, or backing-store failure
- THEN the runtime MUST treat the outcome as a degraded miss
- AND the service layer MUST continue with the fresh origin pipeline instead of failing the request solely because caching failed

### Requirement: Service-Level In-Flight Coalescing

The architecture MUST collapse identical concurrent cacheable requests at the service layer before they fan out to upstream work.

#### Scenario: Concurrent identical cacheable requests arrive together

- GIVEN two or more lookup or search requests with the same normalized cache key arrive while no usable cache entry exists
- WHEN the service layer begins the fresh execution path
- THEN only one upstream execution MUST run for that key
- AND waiting callers MUST receive the leader result once it completes

#### Scenario: No-cache requests bypass coalescing

- GIVEN a request explicitly disables cache usage
- WHEN the service handles that request
- THEN the runtime MUST bypass both cache reuse and keyed in-flight coalescing
- AND each no-cache request MUST execute its own fresh upstream work

### Requirement: Explicit Search Provider Wiring

The system MUST explicitly wire the search registry with intended search providers and their specific defaults, rather than relying on implicit global defaults that match lookup providers.

#### Scenario: Wiring the search module

- GIVEN the application composition root initializes
- WHEN the search service and registry are constructed
- THEN the registry MUST be configured with the intended federated search providers (currently `search` and `dpd`)
- AND the search default set MUST remain distinct from the direct lookup default path (`dpd`)

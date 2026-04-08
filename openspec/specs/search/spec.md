# Search Specification

## Purpose

Defines the semantic routing and module capability of `dlexa search` as invoked from the active CLI surface. It acts as an intelligent gateway for LLMs, converting web search results into actionable next steps while staying truthful about which destination commands are actually wired.

## Requirements

### Requirement: Semantic Gateway Execution

The `search` module MUST retrieve queries from an upstream search engine, normalize results, and act as a semantic router.

#### Scenario: Valid semantic search

- GIVEN the `search` module is invoked via `dlexa search <query>`
- WHEN the upstream returns valid results
- THEN the module MUST return curated results
- AND the results MUST include clear titles and snippets

### Requirement: Institutional Noise Filtering

The module MUST drop institutional or non-normative noise from the upstream results to preserve agent token budgets.

#### Scenario: Filtering noisy URLs

- GIVEN search results containing institutional (`/institucion/*`) or non-normative pages
- WHEN the filtering phase runs
- THEN the module MUST discard those irrelevant links
- AND retain only linguistically valuable content

#### Scenario: Rescuing FAQ content

- GIVEN a result URL points to `/noticia/*`
- WHEN the title explicitly matches FAQ or linguistic gold criteria (e.g., "Preguntas frecuentes:")
- THEN the module MUST NOT discard it
- AND MUST include it in the curated results

### Requirement: URL Compression to Actionable Next Steps

The module MUST map result URLs into literal, LLM-optimized next steps.

#### Scenario: Compress known surfaces into commands

- GIVEN a search result pointing to `https://www.rae.es/espanol-al-dia/la-conjuncion-o`
- WHEN generating the next step
- THEN the module MUST map the URL into an actionable command (e.g., `dlexa espanol-al-dia la-conjuncion-o`)
- AND expose it as the recommended next action
- AND MUST NOT treat that literal suggestion as proof that the destination command is currently registered

#### Scenario: Unknown URLs fall back safely

- GIVEN a search result pointing to an unmapped path
- WHEN generating the next step
- THEN the module SHOULD return a fallback representation
- AND NOT crash or emit malformed syntax

### Requirement: Live Search Retrieval

`dlexa search` MUST retrieve results from a live upstream search surface suitable for RAE content discovery, rather than behaving only as a narrow DPD entry-discovery helper.

#### Scenario: Live semantic search returns curated candidates

- GIVEN the `search` module is invoked via `dlexa search <query>`
- WHEN the upstream search surface returns results successfully
- THEN the module MUST normalize those results into curated search candidates
- AND each curated candidate MUST preserve at least title, snippet, source URL, and next-step metadata

#### Scenario: Search transport failure is explicit

- GIVEN the `search` module is invoked via `dlexa search <query>`
- WHEN the upstream search request fails because of transport, timeout, or upstream unavailability
- THEN the module MUST return an explicit upstream failure fallback
- AND the module MUST NOT silently downgrade the outcome into an empty success response

#### Scenario: Search parse failure is explicit

- GIVEN the `search` module is invoked via `dlexa search <query>`
- WHEN the upstream response is fetched but the result structure cannot be parsed or normalized
- THEN the module MUST return an explicit parse failure fallback
- AND the module MUST NOT panic or emit malformed output

### Requirement: Curated Search Filtering

The `search` module MUST filter institutional or otherwise low-value results while preserving linguistically useful content.

#### Scenario: Filtering institutional noise

- GIVEN the upstream search returns both linguistic content and institutional content such as `/institucion/*`
- WHEN the filtering phase runs
- THEN the module MUST discard those institutional results
- AND the curated output MUST retain only valuable linguistic candidates

#### Scenario: Rescuing linguistically valuable noticia content

- GIVEN a result URL points to `/noticia/*`
- AND the result title indicates linguistically valuable FAQ-style or normative content
- WHEN the filtering phase runs
- THEN the module MUST retain that result
- AND the curated output MUST expose it as a valid search candidate instead of discarding it as institutional noise

### Requirement: URL Compression to Safe Command Suggestions

The `search` module MUST compress recognized URLs into literal next-step suggestions. It MUST explicitly tag suggestions as deferred when the destination command is not yet implemented in the CLI.

#### Scenario: Compressing a known DPD result into an executable command

- GIVEN a search candidate resolves to a DPD article URL or article key
- WHEN the next-step mapping phase runs
- THEN the module MUST flag the candidate as `Deferred: false`
- AND expose a literal executable command suggestion

#### Scenario: Compressing known non-DPD surfaces into deferred suggestions

- GIVEN a search candidate URL belongs to a recognized mapped surface not yet implemented in the CLI (e.g., `espanol-al-dia`)
- WHEN the next-step mapping phase runs
- THEN the module MUST expose a literal suggestion
- AND the module MUST flag the candidate as `Deferred: true`

#### Scenario: Unmapped URLs fall back safely

- GIVEN a curated search candidate points to an unmapped or unrecognized URL shape
- WHEN the next-step mapping phase runs
- THEN the module MUST return a fallback next-step representation
- AND the module MUST NOT emit malformed command syntax
- AND the module MUST NOT crash

### Requirement: Empty and No-Results Handling

The `search` module MUST distinguish between successful searches with no useful matches and failing searches.

#### Scenario: Nonsense or empty-result query returns no-results state

- GIVEN the `search` module is invoked via `dlexa search <query>`
- WHEN the upstream search completes successfully but yields no useful curated candidates
- THEN the module MUST return an explicit no-results response
- AND the response MUST NOT be represented as a transport or parse error

#### Scenario: Filtered-to-empty search is still explicit

- GIVEN the upstream search returns raw results
- AND every raw result is removed during filtering as institutional or low-value noise
- WHEN the final curated result set is produced
- THEN the module MUST return an explicit no-results response
- AND the module MUST NOT surface discarded noise merely to avoid emptiness

### Requirement: Search Help Text Accuracy

The `search` command's help text MUST NOT imply that all returned suggestions are executable commands.

#### Scenario: Viewing search command help

- GIVEN a user or agent invokes `dlexa search --help`
- WHEN the help text is displayed
- THEN it MUST clarify that some suggestions are deferred guidance
- AND MUST NOT instruct users to blindly copy and run all next-command outputs

### Requirement: Search Cache Degradation Semantics

The `search` module MUST treat cache access as a best-effort optimization and MUST continue to fresh retrieval when cache access degrades.

#### Scenario: Search cache read degrades

- GIVEN `dlexa search <query>` is invoked with caching enabled
- WHEN the normalized search cache cannot return a usable entry because of corruption, expiry, or backing-store failure
- THEN the module MUST continue with live fetch, parse, and normalize execution
- AND the request MUST NOT fail solely because the cache read degraded

#### Scenario: Search cache write degrades

- GIVEN a fresh semantic search result has been produced successfully
- WHEN persisting that normalized result into cache fails
- THEN the module MUST still return the fresh successful result
- AND the cache write failure MUST NOT be surfaced as the primary user-visible outcome

### Requirement: Search Request Coalescing

The `search` module MUST coalesce identical concurrent cacheable misses by normalized query key.

#### Scenario: Equivalent concurrent search misses share one upstream execution

- GIVEN concurrent search requests normalize to the same cache key
- AND no usable cached result exists
- WHEN the module executes the live search pipeline
- THEN only one upstream fetch/parse/normalize execution MUST run for that key
- AND all waiting callers MUST receive the same fresh semantic result while preserving their own request fields in the returned envelope

#### Scenario: No-cache search requests bypass coalescing

- GIVEN a search request sets `NoCache` to true
- WHEN the request is executed
- THEN the module MUST bypass keyed coalescing
- AND concurrent no-cache requests MUST each run a fresh upstream search

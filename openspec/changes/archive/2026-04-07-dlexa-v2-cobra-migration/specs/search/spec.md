# Search Specification

## Purpose

Defines the semantic routing and module capability of `dlexa search`. It acts as an intelligent gateway for LLMs, converting web search results into actionable commands and curating tokens.

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

### Requirement: URL Compression to Actionable Commands

The module MUST map result URLs into literal, LLM-optimized CLI commands.

#### Scenario: Compress known surfaces into commands

- GIVEN a search result pointing to `https://www.rae.es/espanol-al-dia/la-conjuncion-o`
- WHEN generating the next step
- THEN the module MUST map the URL into an actionable command (e.g., `dlexa espanol-al-dia la-conjuncion-o`)
- AND expose it as the recommended next action

#### Scenario: Unknown URLs fall back safely

- GIVEN a search result pointing to an unmapped path
- WHEN generating the next step
- THEN the module SHOULD return a fallback representation
- AND NOT crash or emit malformed syntax

# Parser Engine Specification

## Purpose

The Parser Engine provides the central foundation for translating raw DOM output from the Fetch phase into standardized structured Domain Models (`search.Result`, `dpd.Article`). It establishes a uniform envelope for input data, strict anti-corruption boundaries, and specialized ports for different parsing families (e.g., articles vs. searches).

## Requirements

### Requirement: Shared Input Envelope

The engine MUST define a `ParseInput` envelope that encapsulates raw HTML content, content type, request context (like URL/source), and optional metadata.

#### Scenario: Envelope Creation
- GIVEN raw HTML bytes from the Fetch phase
- WHEN preparing data for the parser engine
- THEN the system MUST wrap the bytes and metadata into a standardized `ParseInput` struct
- AND pass it to the appropriate parser interface

### Requirement: Parser Family Ports

The engine MUST define separate, typed interfaces for different parser families, explicitly separating `ArticleParser` (returning `dpd.Article`) from `SearchParser` (returning `search.Result`).

#### Scenario: Distinct Parser Ports
- GIVEN the parser engine foundation
- WHEN a parser implementation is registered
- THEN it MUST implement either the `ArticleParser` interface or the `SearchParser` interface
- AND not a generic wildcard parser port

### Requirement: Parser Resolver/Registry

The engine MUST provide a registry/resolver mechanism to register parsers by intent and retrieve the correct implementation at runtime.

#### Scenario: Resolving a Parser
- GIVEN multiple registered parsers in the engine
- WHEN the application requests an article parser
- THEN the resolver MUST return the registered `ArticleParser`
- AND fail gracefully if the requested parser is not found

### Requirement: Compatibility Adapters

The engine MUST provide backward-compatible bridge adapters for existing parsers (`DPDArticleParser`, `LiveSearchParser`) so that existing service pipelines remain fully operational without modification.

#### Scenario: Transparent Adapter Bridge
- GIVEN an existing service pipeline expecting the legacy parsing signature
- WHEN the pipeline calls the parsing layer
- THEN the bridge adapter MUST intercept the call, wrap the input into `ParseInput`, invoke the engine parser, and return the expected legacy domain model
- AND the service pipeline MUST NOT require any code changes

### Requirement: Zero Behavioral Drift

The foundation slice MUST NOT modify any existing parsing logic, fallback behaviors, or output structure.

#### Scenario: Output Fidelity
- GIVEN existing HTML inputs
- WHEN processed through the new parser foundation adapters
- THEN the resulting models MUST be exactly identical to the legacy parser output
- AND all existing tests MUST pass without changes to expected values

### Requirement: Search Parser Explicit Implementation

Search-family parsers MUST be available as explicit parser-engine `SearchParser` implementations.

#### Scenario: Instantiating search parser
- GIVEN the parser-engine module is initialized
- WHEN requesting a search parser instance for a search-family surface
- THEN the system MUST return an object implementing the `SearchParser` interface
- AND the object MUST be ready to parse search responses

### Requirement: Runtime Wiring Adoption

Runtime search wiring MUST adopt engine-native search parsers.

#### Scenario: Executing a search query
- GIVEN the application is configured with search-family providers
- WHEN a user performs a search
- THEN the runtime MUST delegate parsing of the search payload to an engine-native `SearchParser` implementation
- AND the existing fetch and normalize stages MUST remain in place

### Requirement: Search Behavior Preservation

Search behavior and outputs MUST remain unchanged from the user's perspective during search-family migration.

#### Scenario: Validating search output parity
- GIVEN a known search query that produces a specific set of results
- WHEN the engine-native search parser processes the payload
- THEN the output MUST be identical in content and structure to the legacy parser output
- AND downstream normalized search behavior MUST remain unchanged

### Requirement: Unaltered Search Logic

Search-family parser-engine adoption MUST NOT alter search ranking, policy filtering, or fallback semantics.

#### Scenario: Validating ranking and filtering
- GIVEN a search response containing ranked results and filtered items
- WHEN the engine-native search parser processes the payload
- THEN the returned parsed records MUST preserve the exact ranking inputs and items from the legacy parser
- AND the runtime MUST NOT apply any new filters or fallback rules as part of the migration

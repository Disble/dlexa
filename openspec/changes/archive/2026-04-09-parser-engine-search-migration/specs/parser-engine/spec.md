# Delta for Parser Engine Search Migration

## ADDED Requirements

### Requirement: Search Parser Explicit Implementation
Search-family parsers MUST be available as explicit parser-engine `SearchParser` implementations.

#### Scenario: Instantiating search parser
- GIVEN the parser-engine module is initialized
- WHEN requesting a search parser instance
- THEN the system MUST return an object implementing the `SearchParser` interface
- AND the object MUST be ready to parse search responses

### Requirement: Runtime Wiring Adoption
Runtime search wiring MUST adopt engine-native search parsers.

#### Scenario: Executing a search query
- GIVEN the application is configured to use the engine-native search parser
- WHEN a user performs a search
- THEN the runtime MUST delegate the parsing of the search payload to the new `SearchParser` implementation

## MODIFIED Requirements

### Requirement: Search Behavior Preservation
Search behavior and outputs MUST remain unchanged from the user's perspective.
(Previously: Parsers returned untyped or legacy structures)

#### Scenario: Validating search output parity
- GIVEN a known search query that produces a specific set of results
- WHEN the new engine-native search parser processes the payload
- THEN the output MUST be identical in content and structure to the legacy parser's output

### Requirement: Unaltered Search Logic
This slice MUST NOT alter search ranking, policy filtering, or fallback semantics.
(Previously: Parsing and business logic might have been intertwined)

#### Scenario: Validating ranking and filtering
- GIVEN a search response containing ranked results and filtered items
- WHEN the new engine-native search parser processes the payload
- THEN the returned parsed object MUST preserve the exact ranking and items without applying any new filters or fallback rules

## CONSTRAINTS

### Constraint: Article Parsers Excluded
This slice MUST NOT migrate article-family parsers.

#### Scenario: Processing an article payload
- GIVEN an article response payload
- WHEN the runtime attempts to parse it
- THEN the system MUST continue to use the legacy article parsers, ignoring the new engine-native search parser logic

# Delta for DPD

## ADDED Requirements

### Requirement: Modular DPD Interface

The `dpd` lookup functionality MUST implement a standard `Module` contract, removing tight coupling with the root application boundary and `flag` parser.

#### Scenario: DPD executes as a formal module

- GIVEN the application initializes the Cobra command tree
- WHEN the `dpd` subcommand or default root query is invoked
- THEN execution MUST be delegated to the `internal/modules/dpd` package
- AND the module MUST accept a standard `Request` and return a standard `Response`

### Requirement: Structured Envelope and Fallback Offloading

The `dpd` module MUST delegate its output rendering and error presentation to a centralized Envelope Renderer, returning structured `FallbackEnvelope` instances instead of ad-hoc application errors.

#### Scenario: DPD returns structured not-found

- GIVEN a term does not exist in the DPD
- WHEN the `dpd` module is invoked
- THEN the module MUST return a structured Not Found fallback object
- AND delegate final representation to the application's renderer

#### Scenario: DPD preserves `--format json` compatibility

- GIVEN an LLM agent executes `dlexa dpd <query> --format json`
- WHEN the module prepares the response
- THEN the JSON output MUST bypass the Markdown Envelope mutation
- AND remain fully backward-compatible with the existing JSON schema

## MODIFIED Requirements

### Requirement: Final DPD Output Uses Semantic Markdown

The system MUST emit final DPD article output as semantic Markdown, securely wrapped inside an explicit Markdown Envelope by the renderer.
(Previously: The system MUST emit final DPD article output as semantic Markdown. Acceptance MUST be judged from the rendered Markdown payload that leaves `internal/render`, not only from internal typed spans, intermediate projections, or plain terminal fallback behavior.)

#### Scenario: Internal semantics are insufficient without Markdown output fidelity and envelope

- GIVEN a normalized DPD article preserves prose, emphasis-bearing spans, examples, and references in structured form
- WHEN the final renderer produces the user-visible DPD payload
- THEN acceptance MUST be based on the final semantic Markdown output
- AND the system MUST fail acceptance if the final payload is not wrapped in the standard `[dlexa:dpd]` envelope or collapses into plain terminal text

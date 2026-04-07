# DPD Rendering Specification

## Purpose

Define the final semantic Markdown contract for DPD article output in `dlexa`.

This specification covers the renderer-visible output that reaches users and downstream LLM consumers. It locks the contract corrected by `dpd-terminal-semantic-rendering`: DPD output is accepted as semantic Markdown, not as a plain/ANSI-only terminal projection.

## Requirements

### Requirement: Final DPD Output Uses Semantic Markdown

The system MUST emit final DPD article output as semantic Markdown, securely wrapped inside an explicit Markdown Envelope by the renderer.

#### Scenario: Internal semantics are insufficient without Markdown output fidelity and envelope

- GIVEN a normalized DPD article preserves prose, emphasis-bearing spans, examples, and references in structured form
- WHEN the final renderer produces the user-visible DPD payload
- THEN acceptance MUST be based on the final semantic Markdown output
- AND the system MUST fail acceptance if the final payload is not wrapped in the standard `[dlexa:dpd]` envelope or collapses into plain terminal text

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

### Requirement: Semantic Distinctions Remain Explicit in Markdown

The system MUST preserve the visible distinction between normal prose, semantic emphasis, usage examples, and references in final Markdown.

#### Scenario: Emphasis-bearing spans survive as Markdown emphasis

- GIVEN a DPD paragraph contains `mention`, `italic`, or equivalent emphasis-bearing spans
- WHEN the paragraph is rendered to final Markdown
- THEN those spans MUST remain visibly distinct from surrounding prose through Markdown-safe emphasis
- AND acceptance MUST fail if they degrade into undifferentiated plain text

#### Scenario: Examples remain semantically recoverable in Markdown

- GIVEN a DPD paragraph contains authored usage examples
- WHEN the paragraph is rendered to final Markdown
- THEN the examples MUST remain visibly recoverable as semantic examples in the final output
- AND acceptance MUST fail if examples collapse into ordinary prose with no surviving semantic cue

#### Scenario: References render as canonical Markdown links

- GIVEN a DPD paragraph contains intra-article references
- WHEN the paragraph is rendered to final Markdown
- THEN each reference MUST render in canonical Markdown-link form
- AND acceptance MUST fail if references degrade to plain arrow text or malformed duplicated wrappers

### Requirement: Synthetic Editorial Wrappers and Plain or ANSI Contract Drift Are Forbidden

The renderer MUST NOT introduce synthetic editorial labels or wrappers that are not authored by the source semantics. The final default contract also MUST NOT depend on raw ANSI escapes or plain-text degradation that discards Markdown semantics.

#### Scenario: Rejected synthetic wrappers are absent

- GIVEN a DPD article is rendered to final Markdown
- WHEN the final payload is inspected
- THEN it MUST NOT contain generated wrappers or labels such as `[ej.: ...]`, `ej.:`, `‹...›`, or equivalent synthetic notation
- AND acceptance MUST fail if those artifacts appear in the shipped output

#### Scenario: Raw ANSI bytes are absent from the default Markdown contract

- GIVEN the default DPD Markdown renderer is used
- WHEN the final payload is emitted
- THEN the payload MUST NOT rely on raw ANSI escape sequences for semantic distinction
- AND acceptance MUST fail if semantic visibility depends on ANSI styling instead of Markdown syntax

### Requirement: Markdown Contract Is Verified at the Renderer Boundary

The system MUST verify the DPD Markdown contract with deterministic renderer, integration, and golden checks that inspect the final emitted Markdown.

#### Scenario: Tests and golden fixtures align to Markdown output

- GIVEN authoritative DPD fixtures such as `bien`
- WHEN renderer and parse-normalize-render acceptance checks are executed
- THEN they MUST assert final semantic Markdown output directly
- AND the accepted goldens and tests MUST reject the prior plain/ANSI-only drift as well as synthetic wrappers

### Requirement: DPD Tables Use Markdown When Representable and HTML When Not

DPD table rendering MUST prefer standard Markdown tables for simple rectangular data, but MUST fall back to HTML tables when the source structure depends on spans or other constructs that Markdown tables cannot express faithfully.

#### Scenario: Simple DPD tables render as valid Markdown tables

- GIVEN a DPD table has a single header row, rectangular rows, and no `rowspan` or `colspan`
- WHEN the final Markdown payload is rendered
- THEN the table MUST use pipe-table Markdown syntax with pipe-only divider rows
- AND acceptance MUST fail if the divider row uses non-Markdown separators that common live previews reject

#### Scenario: Complex DPD tables fall back to HTML

- GIVEN a DPD table uses `rowspan`, `colspan`, or equivalent multi-level structure such as the `Tilde diacrítica en qué/...` summary table
- WHEN the final Markdown payload is rendered
- THEN the table MUST be emitted as HTML table markup embedded in the Markdown output
- AND acceptance MUST fail if the renderer flattens that structure into a misleading Markdown grid that loses semantic relationships

### Requirement: DPD Typographical Signs Preserve Validated Semantics End-to-End

The system MUST preserve validated DPD typographical signs through the HTML → Parse → Normalize → Render pipeline without collapsing their semantic meaning.

Validated signs for this contract are:

- `@` digital edition marker from `<sup>@</sup>`
- `+` construction marker from `<span class="nc">+ ...</span>`
- `⊗` exclusion marker from `<span class="bolaspa">⊗</span>`
- `→` cross-reference arrow in rendered references
- bracket contexts authored through `<dfn>`, `<span class="nn">`, and `<span class="yy">`

#### Scenario: Digital edition marker survives end-to-end

- GIVEN DPD HTML contains `<sup>@</sup>` inside article content
- WHEN the article is parsed, normalized, and rendered
- THEN the final Markdown output MUST contain `@`
- AND the structured output MUST preserve its semantic distinction as a digital edition marker

#### Scenario: Construction marker survives end-to-end

- GIVEN DPD HTML contains `<span class="nc">+ infinitivo</span>` inside a phrase
- WHEN the article is parsed, normalized, and rendered
- THEN the final Markdown output MUST contain `+ infinitivo`
- AND the structured output MUST preserve its semantic distinction as a construction marker

#### Scenario: Existing exclusion and reference markers do not regress

- GIVEN DPD HTML contains exclusion markers and cross-reference links
- WHEN the article is parsed, normalized, and rendered
- THEN exclusion markers MUST remain visible as `⊗`
- AND references MUST remain visible in canonical Markdown-link form with their `→` semantics preserved

### Requirement: Bracket Contexts Remain Distinct In Structured DPD Output

The system MUST preserve bracket content as plain Markdown text while keeping its semantic context distinct in structured output.

The protected bracket contexts are:

- definition/correction brackets authored with `<dfn>[...]</dfn>`
- pronunciation brackets authored with `<span class="nn">[...]</span>`
- interpolation/example brackets authored with `<span class="yy">[...]</span>`

#### Scenario: Structured output distinguishes bracket contexts

- GIVEN a DPD article contains definition, pronunciation, and interpolation brackets
- WHEN the article is parsed and normalized
- THEN each bracketed segment MUST keep the semantic context implied by its immediate HTML container
- AND acceptance MUST fail if different bracket contexts are conflated into one indistinguishable structured kind

#### Scenario: Markdown output keeps authored brackets without synthetic wrappers

- GIVEN a bracketed DPD span reaches final Markdown rendering
- WHEN the renderer emits the article
- THEN the output MUST preserve the authored brackets as plain bracket text
- AND the renderer MUST NOT add synthetic wrappers or labels to express bracket semantics

### Requirement: Unvalidated DPD Signs Stay Explicitly Non-Authoritative

The system MAY include defensive support for unvalidated DPD signs inferred from patterns, but those cases MUST remain explicitly documented as speculative until real DPD HTML evidence exists.

Archived signs `<` and `>` remain intentionally unimplemented because their collision risk with HTML syntax is higher than their proven value.

#### Scenario: Speculative signs are implemented with warnings only

- GIVEN support exists for inferred signs such as `*`, `‖`, or `//`
- WHEN that support is reviewed or verified
- THEN the code and tests MUST clearly state that the behavior is speculative and pending real HTML validation
- AND acceptance MUST NOT treat those inferred paths as equivalent to validated DPD sign evidence

#### Scenario: Archived signs remain excluded pending safer evidence

- GIVEN developers review DPD sign support scope
- WHEN they inspect the authoritative spec and archive evidence
- THEN `<` and `>` MUST remain documented as intentionally excluded from implementation
- AND future support MUST require real article evidence plus a collision-safe parsing strategy

## Out-of-Scope Guardrails

- This specification does NOT redefine remote lookup, fetch, or parser ownership.
- This specification does NOT require perfect handling of every nested emphasis edge case before archive.
- This specification does NOT allow regressing to plain terminal output as the primary contract.
- This specification does NOT require speculative DPD signs to be treated as validated behavior before real HTML evidence exists.

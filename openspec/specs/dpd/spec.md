# DPD Rendering Specification

## Purpose

Define the final semantic Markdown contract for DPD article output in `dlexa`.

This specification covers the renderer-visible output that reaches users and downstream LLM consumers. It locks the contract corrected by `dpd-terminal-semantic-rendering`: DPD output is accepted as semantic Markdown, not as a plain/ANSI-only terminal projection.

## Requirements

### Requirement: Final DPD Output Uses Semantic Markdown

The system MUST emit final DPD article output as semantic Markdown. Acceptance MUST be judged from the rendered Markdown payload that leaves `internal/render`, not only from internal typed spans, intermediate projections, or plain terminal fallback behavior.

#### Scenario: Internal semantics are insufficient without Markdown output fidelity

- GIVEN a normalized DPD article preserves prose, emphasis-bearing spans, examples, and references in structured form
- WHEN the final renderer produces the user-visible DPD payload
- THEN acceptance MUST be based on the final semantic Markdown output
- AND the system MUST fail acceptance if the final payload collapses into plain terminal text or ANSI-dependent styling despite the internal model remaining correct

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

## Out-of-Scope Guardrails

- This specification does NOT redefine remote lookup, fetch, or parser ownership.
- This specification does NOT require perfect handling of every nested emphasis edge case before archive.
- This specification does NOT allow regressing to plain terminal output as the primary contract.

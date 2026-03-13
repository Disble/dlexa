# DPD Terminal Semantic Rendering Specification

## Purpose

Define the visible stdout contract for DPD article rendering in `dlexa`.

This specification covers only the terminal-visible distinction between normal prose, semantic emphasis (`mention`, `italic`, or equivalent emphasis-bearing spans), and usage examples. It does not redefine live lookup, HTML extraction, or the full DPD parsing pipeline.

## Requirements

### Requirement: Stdout Is the Acceptance Surface

The system MUST treat final stdout output as the primary acceptance surface for DPD article rendering. Compliance MUST be judged from the terminal-visible result, not only from internal model state, parsed spans, or intermediate Markdown.

#### Scenario: Internal semantics alone do not satisfy the contract

- GIVEN a DPD article whose normalized representation distinguishes prose, semantic emphasis, and example content
- WHEN the final article is rendered for terminal/stdout consumption
- THEN acceptance MUST be based on the visible stdout result
- AND the change MUST fail acceptance if stdout collapses those categories even when the internal model still distinguishes them

#### Scenario: Intermediate formatting is insufficient without visible stdout distinction

- GIVEN the rendering pipeline produces an intermediate representation that preserves semantic categories
- WHEN the final stdout output no longer shows those categories as visibly distinct
- THEN the system MUST be considered non-compliant with this specification
- AND the intermediate representation MUST NOT be used as evidence that the visible contract was satisfied

### Requirement: Prose, Semantic Emphasis, and Examples Remain Visibly Distinct

The system MUST preserve a visible distinction in final stdout between these semantic categories:

- normal prose;
- semantic emphasis, including mention and italic-like emphasis-bearing spans;
- usage examples.

The distinction MUST be observable in the final terminal output through a deterministic visible mechanism. The exact mechanism MAY be decided in design, but the categories MUST NOT collapse into visually indistinguishable plain text.

#### Scenario: Semantic emphasis remains visibly distinct from adjacent prose

- GIVEN a DPD paragraph contains normal prose together with a mention or italicized span
- WHEN the paragraph is rendered to final stdout
- THEN the emphasis-bearing span MUST be visibly distinguishable from the surrounding prose in the final output
- AND acceptance MUST fail if that span becomes visually indistinguishable from adjacent prose

#### Scenario: Example content remains visibly distinct from explanatory prose

- GIVEN a DPD paragraph contains explanatory prose together with a usage example
- WHEN the paragraph is rendered to final stdout
- THEN the example content MUST be visibly distinguishable from the surrounding prose in the final output
- AND acceptance MUST fail if the example becomes just another undifferentiated sentence fragment in the same visible form as prose

#### Scenario: All three semantic categories cannot collapse to one visible form

- GIVEN an authoritative DPD fixture includes normal prose, semantic emphasis, and example content in the same article output
- WHEN the final stdout representation is evaluated
- THEN the rendered output MUST expose those categories as visibly distinct classes
- AND acceptance MUST fail if prose, semantic emphasis, and example content all appear in the same visible form

### Requirement: Synthetic Editorial Wrappers and Labels Are Forbidden

The system MUST NOT introduce generated editorial wrappers or labels that are not authored by the source semantics. This prohibition explicitly includes `[ej.: ...]`, `ej.:`, `‹...›`, and equivalent synthetic category markers.

For this specification, an equivalent synthetic category marker is any renderer-generated wrapper, label, prefix, suffix, or enclosing punctuation whose purpose is to name or tag example/emphasis semantics rather than faithfully render them.

#### Scenario: Example labels are rejected

- GIVEN a DPD example is rendered to final stdout
- WHEN the renderer formats that example for terminal output
- THEN the output MUST NOT prepend or inject generated labels such as `ej.:` or `[ej.: ...]`
- AND acceptance MUST fail if such labels appear in the final stdout output

#### Scenario: Synthetic enclosing wrappers are rejected

- GIVEN a DPD example or emphasis-bearing span is rendered to final stdout
- WHEN the renderer formats that span for terminal output
- THEN the output MUST NOT surround the content with generated wrappers such as `‹...›` or equivalent synthetic enclosing notation not authored by the source
- AND acceptance MUST fail if such wrappers appear in the final stdout output

### Requirement: Visible Differentiation Must Be Verified at the Output Boundary

The system MUST verify this contract with deterministic acceptance scenarios that inspect the final stdout-visible result. Tests or acceptance checks MUST fail when semantic categories are only distinguishable in internal structures but not in the terminal output.

#### Scenario: Output-boundary verification catches semantic collapse

- GIVEN a deterministic fixture with known prose, semantic emphasis, and example spans
- WHEN acceptance checks evaluate the final stdout rendering
- THEN those checks MUST assert visible distinction at the rendered output boundary
- AND the checks MUST fail if the output collapses any protected category into visually plain prose

#### Scenario: Broad snapshot coverage is not enough on its own

- GIVEN an end-to-end golden or snapshot exists for a DPD article
- WHEN acceptance coverage for this change is defined
- THEN at least one deterministic scenario MUST explicitly protect the visible distinction for semantic emphasis
- AND at least one deterministic scenario MUST explicitly protect the visible distinction for examples

### Requirement: Representation Choice Remains Flexible but Constrained

The system MAY choose the final terminal-safe representation in design, but any accepted representation MUST preserve the visible distinctions required by this specification and MUST comply with the prohibition on synthetic editorial markers.

#### Scenario: Different implementations may satisfy the same contract

- GIVEN two candidate renderer designs use different terminal-visible mechanisms
- WHEN both are evaluated against this specification
- THEN either design MAY be acceptable if prose, semantic emphasis, and examples remain visibly distinct in final stdout
- AND neither design is acceptable if it relies on forbidden editorial labels or wrappers

## Out-of-Scope Guardrails

- This specification does NOT redefine live fetch, source selection, or remote lookup behavior.
- This specification does NOT require a full parser redesign.
- This specification does NOT lock a single exact visual spelling before design validates the terminal-safe representation.
- This specification does NOT accept vague readability claims as evidence of correctness; the contract is visible semantic distinction, not "looks nicer".

# DPD Lookup Specification

## Purpose

Define the required runtime and output behavior for live Diccionario panhispánico de dudas lookups in `dlexa`, with `dlexa bien` accepted primarily by terminal/Markdown semantic fidelity for LLM consumption rather than by raw data extraction alone.

## Requirements

### Requirement: Live DPD Production Lookup

The system MUST execute the default DPD lookup path against the live remote DPD source for production behavior. The system MUST NOT satisfy production DPD lookups from bootstrap demo content, production mocks, or an internal database.

#### Scenario: Live lookup is the default production path

- GIVEN the user runs `dlexa bien` with the default DPD source configuration
- WHEN the lookup flow resolves the requested source
- THEN the system MUST acquire DPD content from the live remote source
- AND the system MUST continue through the existing query-first `fetch -> parse -> normalize -> render` boundary sequence

#### Scenario: Non-production data sources are excluded from parity behavior

- GIVEN the system is executing the production DPD lookup path
- WHEN the lookup is performed for a real term
- THEN the resulting article MUST NOT be produced from fixture-only bootstrap content
- AND the resulting article MUST NOT depend on an internal persistence store to emulate the remote source

### Requirement: Article Body Extraction from Live DPD HTML

The system MUST extract the canonical DPD article body from the fetched live HTML response and MUST exclude surrounding site chrome. The extraction contract MUST preserve only the content needed to produce correct downstream Markdown for the actual article.

#### Scenario: Canonical article body is isolated from page chrome

- GIVEN the fetched live DPD response contains the target article together with navigation, promotional, and layout content
- WHEN the parser identifies the article body for `bien`
- THEN only the canonical article body MUST proceed to normalization and rendering
- AND menus, related-content blocks, share widgets, newsletter/footer content, and other non-article shell content MUST be excluded

#### Scenario: Extraction failure is treated as a contract failure

- GIVEN the fetched live HTML does not yield a recognizable canonical DPD article body
- WHEN extraction is attempted
- THEN the system MUST report a parse-or-extraction failure outcome
- AND it MUST NOT silently fall back to partial shell content or synthetic article text

### Requirement: Terminal and Markdown Acceptance Contract for `bien`

For v1, the primary acceptance criterion MUST be that `dlexa bien` produces readable terminal/Markdown output whose semantics match the real DPD article as authored. Acceptance MUST be based on ordered article meaning, heading semantics, inline emphasis, references, and citation structure as consumed by an LLM or terminal reader, and SHALL NOT be reduced to mere presence of extracted facts.

#### Scenario: Acceptance is based on semantic output fidelity

- GIVEN the live DPD article for `bien` is successfully fetched, extracted, and normalized
- WHEN parity is evaluated
- THEN the output MUST include the dictionary context, edition marker, lemma, ordered sections `1.` through `7.`, nested `6.a)` through `6.c)` semantics, and readable references
- AND acceptance MUST fail if those semantics are degraded even when the raw words are still present somewhere in the output

#### Scenario: Acceptance is not based on page layout or chrome similarity

- GIVEN the source page contains presentation details outside the canonical article body
- WHEN parity is evaluated for `dlexa bien`
- THEN acceptance MUST center on semantically faithful terminal/Markdown rendering of the article itself
- AND the result MUST NOT be judged by pixel-level, browser-layout, or chrome-level similarity to `rae.es`

### Requirement: Editorial Preservation of Authored Forms

The system MUST preserve authored editorial forms that affect terminal/Markdown reading, including Spanish punctuation and glyphs such as `2.ª`, apostrophes as authored, and guillemets `«»` where present in the source. The system MUST NOT introduce synthetic quote normalization that changes the article's authored punctuation style.

#### Scenario: Edition marker glyphs are preserved exactly

- GIVEN the authoritative `bien` fixture contains the edition marker `2.ª edición`
- WHEN Markdown output is rendered
- THEN the rendered output MUST preserve the ordinal glyph and punctuation semantics of that marker
- AND the renderer MUST NOT downgrade it into a flattened or anglicized substitute solely for convenience

#### Scenario: Synthetic mixed quote wrappers are rejected

- GIVEN a DPD definition or gloss appears in the source without mixed synthetic quote wrappers
- WHEN the article is normalized and rendered
- THEN the output MUST preserve the authored punctuation form or a semantically equivalent Markdown-safe form
- AND it MUST NOT emit hybrid constructions created by the pipeline itself such as mismatched ASCII quotes around text that was not authored that way

#### Scenario: Authored apostrophes and guillemets survive transformation

- GIVEN the extracted article body contains apostrophes or guillemets as part of authored text
- WHEN that content flows through normalization and rendering
- THEN those characters MUST survive in the final output when supported by the source fixture
- AND they MUST NOT be silently replaced by unrelated quote styles that change authored reading cues

### Requirement: Example and Emphasis Semantics Are Preserved

The system MUST preserve example semantics and inline emphasis from source HTML whenever those signals affect meaning, contrast, or reading flow in the `bien` article. The system MUST NOT flatten emphasized examples, contrastive forms, or locutions into indistinguishable plain text when the source expresses them as semantically marked content.

#### Scenario: Contrastive terms keep emphasis semantics

- GIVEN the source article marks contrastive forms such as `más bien`, `mejor`, or `si bien` with emphasis-relevant HTML semantics
- WHEN Markdown output is rendered
- THEN the final output MUST preserve those emphasis distinctions in readable Markdown-safe form
- AND the terms MUST NOT become visually indistinguishable from surrounding prose solely because markup was stripped

#### Scenario: Example content remains semantically separable from prose

- GIVEN a paragraph contains an example or usage fragment distinguished in the source HTML
- WHEN the paragraph is normalized and rendered
- THEN the final output MUST preserve enough structure or emphasis for a reader to distinguish the example from explanatory prose
- AND the example MUST NOT collapse into undifferentiated sentence text that loses its role in the article

### Requirement: Cross-Reference Rendering Is Canonical and Non-Malformed

The system MUST render intra-article and related references in a readable canonical form for terminal/Markdown consumption. The system MUST NOT duplicate arrows, duplicate target labels, or introduce malformed parenthetical structures such as `(→ [→ 6](...))`.

#### Scenario: Numeric references render with one arrow and one target label

- GIVEN the source article contains a reference to section `6` or `7`
- WHEN the reference is rendered into terminal/Markdown output
- THEN the visible reference text MUST contain exactly one directional marker and one target label for that reference
- AND it MUST be readable as `→ [6]`, `→ [7]`, or a semantically equivalent form without duplicated arrow text

#### Scenario: Parenthetical references do not nest malformed wrappers

- GIVEN a reference appears within surrounding punctuation such as parentheses in the source article
- WHEN the final Markdown is rendered
- THEN the output MUST preserve readable surrounding punctuation without doubling the reference marker inside it
- AND the output MUST NOT contain malformed constructions where wrapper punctuation and generated reference syntax collide

### Requirement: Lexical Heads Stay Integrated with Numbered Heading Semantics

The system MUST preserve lexical heads such as `bien que.`, `más bien.`, and `si bien.` as part of their owning numbered heading semantics. The system MUST NOT split the numeric label from the lexical head into disconnected blocks that force the reader or LLM to reconstruct the heading relationship.

#### Scenario: Section heading remains a single semantic heading

- GIVEN the source article expresses a numbered item together with its lexical head
- WHEN that item is normalized and rendered
- THEN the output MUST keep the numeric label and lexical head in one heading-level semantic unit
- AND it MUST NOT render the number on one line and the lexical head as detached body text on another line unless the source itself requires that structure

#### Scenario: Section six subitems retain their lexical heads under the parent section

- GIVEN section `6.` contains the subitems `a)`, `b)`, and `c)` with lexical heads
- WHEN the article is rendered
- THEN each subitem MUST remain attached to section `6.` and preserve its lexical head as part of the subitem heading semantics
- AND the output MUST remain readable as a parent section with ordered lexical subentries rather than as flattened loose paragraphs

### Requirement: Citation and Reference Structure Remains Explicit

The system MUST preserve citation and reference structure strongly enough that a terminal or LLM reader can distinguish article body content, intra-article references, canonical source identity, canonical URL, edition marker, and consultation metadata. The system MUST NOT flatten those elements into a single opaque blob when the source distinguishes them.

#### Scenario: Citation essentials remain structurally distinguishable

- GIVEN the normalized article contains citation essentials for source, edition, canonical URL, and consultation metadata
- WHEN Markdown or JSON output is produced
- THEN those citation elements MUST remain individually recoverable and readable in the output structure
- AND they MUST NOT be collapsed into a single undifferentiated sentence that loses field boundaries

#### Scenario: Intra-article references are not conflated with citation metadata

- GIVEN the article contains both section references and citation metadata
- WHEN the output is rendered
- THEN section references MUST remain attached to article content where they are used
- AND citation metadata MUST remain identifiable as source attribution rather than as ordinary body prose or section-reference text

### Requirement: Minimal Canonical Structure for Fidelity-Critical Rendering

The system MUST preserve only the structured article information necessary to render faithful terminal/Markdown output and avoid reparsing HTML downstream. Structured modeling for this change SHALL be minimal and sufficient rather than exhaustive, but it MUST be rich enough to represent every verified formatting-fidelity case in this specification.

#### Scenario: Minimal structure still covers all verified fidelity cases

- GIVEN the verified defect inventory includes quote preservation, example emphasis, cross-reference shape, integrated lexical heads, and citation structure
- WHEN canonical article data is defined for v1
- THEN that structure MUST preserve the hierarchy and inline semantics needed to render each of those cases correctly
- AND the system MUST NOT claim the model is sufficient if any verified case still requires renderer-side heuristics against raw HTML

#### Scenario: Uneven article shape does not justify flattening semantics away

- GIVEN a DPD article includes sections with different paragraph counts, mixed nested content, or citation-bearing tails
- WHEN the article is normalized
- THEN the canonical structure MUST preserve source hierarchy and inline meaning needed for fidelity
- AND the system MUST NOT flatten headings, references, or citations merely to avoid representing them explicitly

### Requirement: Secondary Structured Output and Metadata

JSON output and richer metadata are secondary for this change. When JSON is emitted, it SHOULD derive from the same normalized article representation used for terminal/Markdown output, but JSON completeness and metadata richness SHALL NOT become the primary acceptance gate for `dpd-live-lookup-parity`.

#### Scenario: JSON remains aligned but secondary

- GIVEN a canonical DPD article has been normalized successfully
- WHEN JSON output is requested
- THEN the output SHOULD expose the same article hierarchy and core identity needed to reflect the Markdown meaning
- AND missing non-essential rich metadata SHALL NOT by itself fail the primary business goal if the terminal/Markdown fidelity contract is satisfied

#### Scenario: Markdown-first acceptance resolves prioritization conflicts

- GIVEN there is a tradeoff between adding richer structured metadata and correcting a Markdown fidelity defect in `bien`
- WHEN v1 scope is evaluated
- THEN the system MUST prioritize the behavior that improves faithful terminal/Markdown output from the real article
- AND richer JSON or metadata work MAY be deferred if the fidelity contract is met

### Requirement: Lookup Failure Classification

The system MUST distinguish at least three DPD lookup failure classes: remote fetch failure, not-found outcome, and parse-or-render-input failure. Error handling MUST preserve the difference between these classes so callers and verification flows can tell whether failure came from transport, domain absence, or inability to derive faithful article output from the live response.

#### Scenario: Remote fetch failure is surfaced distinctly

- GIVEN the live DPD source cannot be reached or returns an unusable transport-level response
- WHEN a DPD lookup is attempted
- THEN the system MUST report a remote fetch failure outcome
- AND that outcome MUST remain distinguishable from not-found and parse-or-render-input failures

#### Scenario: Canonical article absence is treated as not found

- GIVEN the live DPD source resolves the request but no canonical entry exists for the queried term
- WHEN the lookup flow completes source acquisition
- THEN the system MUST report a not-found outcome
- AND it MUST NOT misclassify the result as a parser failure solely because no article was present

#### Scenario: Content-shape breakdown is surfaced as parse failure

- GIVEN the live response is fetched but the article body cannot be extracted or transformed into fidelity-supporting structure
- WHEN the lookup flow reaches parse or normalization stages
- THEN the system MUST report a parse-or-normalization failure
- AND the failure MUST remain distinguishable from transport and not-found outcomes

### Requirement: Deterministic Fixture Verification Is the Acceptance Baseline

The system MUST support stable verification of DPD parity through captured authoritative HTML fixtures from the real DPD response and expectations derived from those fixtures. Deterministic fixture-based verification MUST be the acceptance baseline for this change, and optional live/upstream drift checks MUST remain explicitly separate.

#### Scenario: Authoritative fixture is the deterministic source of truth

- GIVEN an authoritative captured HTML response for the live `bien` article is available
- WHEN extraction, normalization, and rendering are verified
- THEN fixture-based checks MUST be sufficient to validate the v1 contract deterministically
- AND those checks MUST evaluate semantic fidelity rather than only raw text presence

#### Scenario: Stale defective expectations are not accepted as the contract

- GIVEN an existing golden file or fixture expectation encodes malformed formatting that contradicts the authoritative `bien` fixture
- WHEN the verification baseline is updated
- THEN the stale expectation MUST be replaced or corrected
- AND the system MUST NOT preserve a known defect merely because it was already captured in a golden artifact

#### Scenario: Live verification is optional drift detection only

- GIVEN a maintainer chooses to run live parity verification against the remote DPD source
- WHEN the live verification suite is executed
- THEN it MAY probe the real upstream `bien` article as an integration or drift signal
- AND a failure in that opt-in suite MUST NOT redefine the deterministic fixture contract without an explicit fixture refresh decision

### Requirement: Many Granular Tests Cover Each Verified Formatting Case

Verification for this change MUST include many targeted tests, not only one broad golden test. Each verified formatting-fidelity case in this specification MUST map to one or more deterministic fixture-based tests that isolate the specific behavior being protected.

#### Scenario: Every verified formatting defect has targeted deterministic coverage

- GIVEN the verified defect inventory includes quote normalization, example/emphasis loss, malformed cross-references, split lexical heads, flattened citation structure, and stale golden expectations
- WHEN the verification suite for `dpd-live-lookup-parity` is defined
- THEN each defect category MUST have at least one targeted deterministic test that fails specifically for that category
- AND the suite MUST make it possible to identify which formatting behavior regressed without relying only on a full-output diff

#### Scenario: End-to-end goldens supplement but do not replace granular tests

- GIVEN a full `bien` Markdown golden test exists
- WHEN acceptance coverage is evaluated
- THEN that golden test MUST be treated as a high-level integration safeguard
- AND it SHALL NOT be the only automated evidence for the formatting cases required by this specification

#### Scenario: Parser, normalizer, and renderer tests are aligned to the same fixture baseline

- GIVEN the deterministic fixture baseline for `bien` has been captured
- WHEN granular tests are added across parser, normalizer, renderer, and integration boundaries
- THEN those tests MUST agree on the same authoritative fixture semantics
- AND they MUST NOT encode contradictory expectations for the same formatting case at different layers

### Requirement: DPD Cache Degradation Semantics

The DPD lookup path MUST treat cache access as a best-effort optimization and MUST continue to live lookup when cache access degrades.

#### Scenario: DPD cache read degrades

- GIVEN a DPD lookup is invoked with caching enabled
- WHEN the lookup cache cannot provide a usable entry because of corruption, expiry, or backing-store failure
- THEN the system MUST continue through the live `fetch -> parse -> normalize -> render` lookup path
- AND the request MUST NOT fail solely because cache retrieval degraded

#### Scenario: DPD cache write degrades

- GIVEN a fresh DPD lookup result has been produced successfully
- WHEN persisting that result into cache fails
- THEN the system MUST still return the fresh DPD result
- AND the cache write failure MUST NOT replace the primary lookup outcome seen by the caller

### Requirement: DPD Request Coalescing on Cache Misses

The DPD lookup path MUST collapse identical concurrent cacheable misses into one upstream lookup execution.

#### Scenario: Concurrent identical DPD misses share one upstream execution

- GIVEN two or more concurrent DPD lookups resolve to the same cache key
- AND no usable cached result exists
- WHEN the lookup service executes the live source pipeline
- THEN only one upstream lookup execution MUST run for that key
- AND the waiting callers MUST receive the leader's fresh lookup result

#### Scenario: No-cache DPD requests bypass coalescing

- GIVEN a DPD lookup request explicitly disables cache usage
- WHEN that request is executed
- THEN the system MUST bypass keyed in-flight coalescing
- AND concurrent no-cache lookups MUST each run their own fresh upstream execution

## Architectural Significance

- This change is architecturally significant because it converts DPD lookup from bootstrap scaffolding into a real live source while keeping the existing `fetch -> parse -> normalize -> render` boundary model intact.
- The critical path is the parser and transformer chain that turns live DPD HTML into semantically faithful terminal/Markdown output, not a broad source-model redesign.
- Canonical structure remains important, but only as a minimal anti-corruption layer that preserves article meaning for rendering and secondary structured output.

## Design-Pattern Decision Notes for Design Phase

The design phase MUST keep architecture and design-pattern reasoning explicit, but aligned to the fidelity-first business value. It MUST justify patterns only where they protect extraction quality, transformation fidelity, boundary clarity, or failure isolation.

### Creational

- The design MUST explain how live DPD lookup components are instantiated and wired from the composition root.
- If factories or configuration-driven creation are used, the design MUST justify them against simpler constructor wiring for the fidelity-first pipeline.

### Structural

- The design MUST explain how live remote HTML is translated into a minimal canonical article shape without leaking site chrome or source-specific markup into renderers.
- If adapters, facades, translators, or anti-corruption-style structures are introduced, the design MUST justify them in terms of protecting terminal/Markdown fidelity and boundary clarity rather than maximizing model richness.

### Behavioral

- The design MUST explain the runtime behavior that coordinates live fetch, article-body extraction, normalization, rendering, and failure classification.
- If pipeline, strategy, or similar behavioral patterns are used, the design MUST justify them in terms of correctness, maintainability, and isolation of fidelity-critical transformations.

## Out-of-Scope Guardrails

- This specification does NOT require pixel-perfect page reproduction.
- This specification does NOT require a giant generalized redesign of every DPD article family.
- This specification does NOT require an internal database.
- This specification does NOT allow production mocks to satisfy real lookup behavior.
- This specification does NOT require exhaustive JSON richness or metadata completeness for v1.
- This specification does NOT treat `go-rae` as an implementation blueprint; any reuse of ideas from that repository MUST be justified at the boundary and architectural level.

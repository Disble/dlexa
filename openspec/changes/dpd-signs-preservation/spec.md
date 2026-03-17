# Delta Specification: DPD Signs Preservation

## Overview

This specification defines requirements for preserving all 11 official DPD typographical signs through the HTML → Parse → Normalize → Render pipeline. Currently, only 2 signs (⊗ exclusion, → cross-reference) preserve their semantic meaning end-to-end. This change extends the existing semantic preservation architecture to handle all 11 signs consistently.

**Reference Implementation**: The ⊗ (exclusion) sign provides the gold standard pattern: whitelist → constant → InlineKind → parse recognition → explicit preservation → render dispatch.

**Two-Tier Approach**:
- **VALIDATED signs (6)**: Real HTML evidence from DPD articles confirms encoding. Implementation MUST work with actual HTML.
- **SPECULATIVE signs (5)**: No real HTML found yet. Implementation uses inferred patterns with WARNING annotations.

---

## VALIDATED Requirements (6 Signs)

These requirements are based on real HTML extracted from DPD articles. Implementation MUST use the documented HTML patterns.

### Requirement: Digital Edition Marker (@) MUST Be Preserved

The system MUST preserve the `@` sign (digital edition marker) through the full pipeline, maintaining its semantic distinction as a source attribution indicator.

**HTML Evidence**: `<sup>@</sup>` (confirmed in alícuota.html line 545, androfobia.html)

**Current Behavior**: The `@` sign is **STRIPPED** because `<sup>` is not whitelisted in `semanticSpanAllowed`.

The system MUST:
- Add `<sup>` to the `semanticSpanAllowed` whitelist
- Define a constant `digitalEditionGlyph = "@"`
- Add `InlineKindDigitalEdition` to the InlineKind enum
- Parse `<sup>@</sup>` and set the correct InlineKind
- Preserve the `@` character during text normalization
- Emit the `@` character in Markdown rendering

#### Scenario: Digital edition marker survives end-to-end parsing

- GIVEN DPD HTML contains `<span class="bib">(<i>Día</i><sup>@</sup> <span class="cbil" title="España">es</span> 26.10.2014)</span>`
- WHEN the HTML is parsed, normalized, and rendered to Markdown
- THEN the final Markdown output MUST contain the `@` character
- AND the JSON output MUST represent it with `InlineKindDigitalEdition`

#### Scenario: Superscript whitelist does not interfere with non-sign content

- GIVEN DPD HTML contains `<sup>` tags with content other than `@`
- WHEN the HTML is parsed
- THEN the `<sup>` tag MUST be whitelisted for processing
- AND content other than `@` MUST be handled according to existing logic

---

### Requirement: Construction Marker (+) MUST Be Preserved

The system MUST preserve the `+` sign (construction marker) through the full pipeline, maintaining its semantic distinction as a grammatical construction indicator.

**HTML Evidence**: `<span class="nc">+ infinitivo</span>` (confirmed in acertar.html line 546)

**Current Behavior**: The `+` sign is **STRIPPED** because `<span class="nc">` is not whitelisted in `semanticSpanAllowed`.

The system MUST:
- Add `"<span class=\"nc\">": true` to the `semanticSpanAllowed` whitelist
- Define a constant `constructionMarkerGlyph = "+"`
- Add `InlineKindConstructionMarker` to the InlineKind enum
- Parse `<span class="nc">+ ...</span>` and set the correct InlineKind
- Preserve the entire construction marker phrase during text normalization
- Emit the construction marker in Markdown rendering

#### Scenario: Construction marker survives end-to-end parsing

- GIVEN DPD HTML contains `<span class="embf">acertar a <span class="nc">+ infinitivo</span>.</span>`
- WHEN the HTML is parsed, normalized, and rendered to Markdown
- THEN the final Markdown output MUST contain `+ infinitivo`
- AND the JSON output MUST represent it with `InlineKindConstructionMarker`

#### Scenario: Construction marker appears mid-phrase

- GIVEN DPD HTML embeds `<span class="nc">+ infinitivo</span>` within a larger phrase
- WHEN the HTML is parsed
- THEN the construction marker MUST remain inline with surrounding text
- AND whitespace MUST be preserved correctly

---

### Requirement: Exclusion Sign (⊗) MUST Continue Working

The system MUST NOT regress the existing ⊗ (exclusion/incorrect forms) sign preservation behavior.

**HTML Evidence**: `<span class="bolaspa">⊗&#x200D;</span>` or `<span class="bolaspa">&#x2297;&#x200D;</span>` (confirmed in alícuota.html line 545, acertar.html line 547)

**Current Behavior**: The ⊗ sign is **WORKING CORRECTLY** (reference implementation).

The system MUST:
- Continue to whitelist `<span class="bolaspa">`
- Continue to use `exclusionGlyph = "\u2297"` constant
- Continue to use `InlineKindExclusion`
- Continue to preserve the exclusion glyph through normalization
- Continue to emit the ⊗ character in Markdown rendering

#### Scenario: Exclusion sign regression is detected

- GIVEN DPD HTML contains `<em><span class="bolaspa">⊗&#x200D;</span><span class="ment">alicuoto</span></em>`
- WHEN the HTML is parsed, normalized, and rendered to Markdown
- THEN the final Markdown output MUST contain the `⊗` character
- AND the JSON output MUST represent it with `InlineKindExclusion`
- AND acceptance MUST fail if the ⊗ sign is stripped or altered

---

### Requirement: Cross-Reference Arrow (→) MUST Continue Working

The system MUST NOT regress the existing → (cross-reference) sign preservation behavior.

**HTML Evidence**: `<a href="/dpd/...">→ acertar</a>` (confirmed in acertar.html line 543)

**Current Behavior**: The → sign is **WORKING CORRECTLY** (already preserved in anchor tags).

The system MUST:
- Continue to whitelist and process `<a>` tags
- Continue to preserve the → character within anchor text
- Continue to emit the arrow in canonical Markdown link form

#### Scenario: Cross-reference arrow regression is detected

- GIVEN DPD HTML contains `Verbo irregular (&#x2192; <a href="/dpd/ayuda/modelos-de-conjugacion-verbal#acertar">acertar</a>)`
- WHEN the HTML is parsed, normalized, and rendered to Markdown
- THEN the final Markdown output MUST contain a canonical Markdown link with the → character
- AND acceptance MUST fail if the arrow is stripped or the link is malformed

---

### Requirement: Bracket Semantics MUST Be Distinguished by Context

The system MUST preserve brackets (`[` and `]`) and distinguish between three semantic contexts: definition/correction, pronunciation, and interpolation/example.

**HTML Evidence**:
- **Definition/Correction**: `<dfn>[una ley]</dfn>` (confirmed in abrogar.html)
- **Pronunciation**: `<span class="nn">[alikuóto]</span>` (confirmed in alícuota.html, acertar.html)
- **Interpolation/Example**: `<span class="yy">[las feministas]</span>` (confirmed in androfobia.html)

**Current Behavior**: Brackets appear as plain text, but semantic context (pronunciation vs. definition vs. interpolation) is **LOST** after parsing.

The system MUST:
- Continue to whitelist `<dfn>` (already whitelisted)
- Add `"<span class=\"nn\">": true` to the `semanticSpanAllowed` whitelist
- Add `"<span class=\"yy\">": true` to the `semanticSpanAllowed` whitelist
- Add three InlineKind constants:
  - `InlineKindBracketDefinition` for `<dfn>[...]</dfn>`
  - `InlineKindBracketPronunciation` for `<span class="nn">[...]</span>`
  - `InlineKindBracketInterpolation` for `<span class="yy">[...]</span>`
- Parse bracket content and set the correct InlineKind based on parent tag
- Preserve brackets as-is in text normalization (already work as plain text)
- Emit brackets in Markdown output as plain text `[...]`
- Emit distinct JSON representations for each bracket context

#### Scenario: Definition brackets are semantically tagged

- GIVEN DPD HTML contains `<dfn>'Derogar o abolir [una ley]'</dfn>`
- WHEN the HTML is parsed and normalized
- THEN the JSON output MUST represent the bracketed content with `InlineKindBracketDefinition`
- AND the Markdown output MUST preserve the brackets as plain text `[una ley]`

#### Scenario: Pronunciation brackets are semantically tagged

- GIVEN DPD HTML contains `<span class="nn">[alikuóto]</span>`
- WHEN the HTML is parsed and normalized
- THEN the JSON output MUST represent the bracketed content with `InlineKindBracketPronunciation`
- AND the Markdown output MUST preserve the brackets as plain text `[alikuóto]`

#### Scenario: Interpolation brackets are semantically tagged

- GIVEN DPD HTML contains `<span class="yy">[las feministas]</span>`
- WHEN the HTML is parsed and normalized
- THEN the JSON output MUST represent the bracketed content with `InlineKindBracketInterpolation`
- AND the Markdown output MUST preserve the brackets as plain text `[las feministas]`

#### Scenario: Bracket context ambiguity does not cause mis-classification

- GIVEN DPD HTML contains multiple bracket contexts in the same article
- WHEN the HTML is parsed
- THEN each bracket MUST be classified according to its immediate parent tag
- AND acceptance MUST fail if bracket contexts are conflated

---

### Requirement: Slash (/) MUST Be Preserved as Plain Text

The system MUST preserve the `/` character as plain text without special semantic preservation.

**HTML Evidence**: Plain text character in article body (confirmed in all test articles)

**Current Behavior**: The `/` character appears as plain text with **NO SEMANTIC ENCODING** in DPD HTML.

The system MUST:
- Preserve `/` as plain text during normalization
- Emit `/` as plain text in Markdown rendering
- NOT create a dedicated InlineKind for slash
- NOT add special whitelist entries for slash

#### Scenario: Slash appears as plain text

- GIVEN DPD HTML contains plain `/` characters in prose
- WHEN the HTML is parsed, normalized, and rendered to Markdown
- THEN the `/` character MUST be preserved as plain text
- AND the system MUST NOT attempt to distinguish semantic vs. non-semantic slash usage

---

## SPECULATIVE Requirements (5 Signs)

These requirements are based on inferred patterns because no real HTML examples were found in test articles. Implementation MUST include WARNING annotations indicating lack of validation.

**⚠️ IMPORTANT**: These signs were NOT FOUND in any of the 8 analyzed DPD articles. HTML encoding is inferred from patterns observed in validated signs. When real HTML examples are discovered, these requirements MUST be updated with actual encoding.

### Requirement: Agrammatical Marker (*) SHOULD Be Preserved (SPECULATIVE)

The system SHOULD preserve the `*` sign (agrammatical constructions) through the full pipeline, using an inferred HTML pattern until real examples are found.

**HTML Evidence**: ❌ **NOT FOUND** in test articles (abrogar, abstraer, abuelo, acertar, adherir, alférez, alícuota, androfobia)

**Inferred Pattern**: Likely `<span class="??">*</span>` (similar to ⊗ pattern)

The system SHOULD:
- Define a constant `agrammaticalGlyph = "*"`
- Add `InlineKindAgrammatical` to the InlineKind enum
- Implement BEST-GUESS parsing logic based on inferred pattern
- Add explicit preservation in text normalization
- Add render case for agrammatical marker
- Include WARNING comments in code indicating lack of HTML validation

#### Scenario: Agrammatical marker implementation is marked as speculative (WARNING)

- GIVEN the codebase contains parsing, normalization, and rendering logic for `*`
- WHEN code review is performed
- THEN the implementation MUST include WARNING comments stating:
  - No real HTML example found in DPD articles
  - HTML pattern is inferred, not validated
  - Implementation may need updates when real examples are discovered
- AND acceptance MUST fail if warning annotations are missing

#### Scenario: Agrammatical marker compiles and passes synthetic tests

- GIVEN synthetic test fixtures with inferred `*` HTML pattern
- WHEN tests are executed
- THEN the implementation MUST compile without errors
- AND synthetic tests MUST pass
- AND tests MUST be clearly marked as synthetic (not real DPD fixtures)

---

### Requirement: Hypothetical Marker (‖) SHOULD Be Preserved (SPECULATIVE)

The system SHOULD preserve the `‖` sign (hypothetical/reconstructed forms) through the full pipeline, using an inferred HTML pattern until real examples are found.

**HTML Evidence**: ❌ **NOT FOUND** in test articles (abrogar, abstraer, abuelo, acertar, adherir, alférez, alícuota, androfobia)

**Inferred Pattern**: Likely `<span class="??">‖</span>` (similar to ⊗ pattern)

The system SHOULD:
- Define a constant `hypotheticalGlyph = "‖"`
- Add `InlineKindHypothetical` to the InlineKind enum
- Implement BEST-GUESS parsing logic based on inferred pattern
- Add explicit preservation in text normalization
- Add render case for hypothetical marker
- Include WARNING comments in code indicating lack of HTML validation

#### Scenario: Hypothetical marker implementation is marked as speculative (WARNING)

- GIVEN the codebase contains parsing, normalization, and rendering logic for `‖`
- WHEN code review is performed
- THEN the implementation MUST include WARNING comments stating:
  - No real HTML example found in DPD articles
  - HTML pattern is inferred, not validated
  - Implementation may need updates when real examples are discovered
- AND acceptance MUST fail if warning annotations are missing

#### Scenario: Hypothetical marker compiles and passes synthetic tests

- GIVEN synthetic test fixtures with inferred `‖` HTML pattern
- WHEN tests are executed
- THEN the implementation MUST compile without errors
- AND synthetic tests MUST pass
- AND tests MUST be clearly marked as synthetic (not real DPD fixtures)

---

### Requirement: Transformation Marker (>) - ARCHIVED

**Status**: 📦 **ARCHIVED** - Not implemented in this change.

**Archival Reason**: HTML tag collision risk. The `>` character conflicts with HTML tag syntax (`</header>`, closing `>` in tag attributes). Low appearance ratio and absence in high-value articles make the risk-to-benefit ratio unfavorable.

**HTML Evidence**: NOT FOUND in test articles. Only observed as part of HTML tag syntax, not as semantic content.

**Future Reconsideration Criteria**:
1. Real DPD articles containing `>` sign in high-value contexts are discovered
2. A collision-safe parsing strategy is developed (e.g., context-aware detection that distinguishes semantic `>` from tag syntax)
3. User demand justifies implementation complexity

**Documentation**: Archived status documented in `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md`. Code SHOULD include a comment referencing this archival decision for future developers.

---

### Requirement: Etymology Marker (<) - ARCHIVED

**Status**: 📦 **ARCHIVED** - Not implemented in this change.

**Archival Reason**: HTML tag collision risk. The `<` character conflicts with HTML tag syntax (`<span>`, `<div>`, `<a>`, etc.). Low appearance ratio and absence in high-value articles make the risk-to-benefit ratio unfavorable.

**HTML Evidence**: NOT FOUND in test articles. Only observed as part of HTML tag syntax (`<entry>`, `<span>`), not as semantic content.

**Future Reconsideration Criteria**:
1. Real DPD articles containing `<` sign in high-value contexts are discovered
2. A collision-safe parsing strategy is developed (e.g., HTML entity detection `&lt;` with context validation)
3. User demand justifies implementation complexity

**Documentation**: Archived status documented in `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md`. Code SHOULD include a comment referencing this archival decision for future developers.

---

### Requirement: Phoneme Marker (//) SHOULD Be Preserved (SPECULATIVE)

The system SHOULD preserve the `//` sign (phoneme delimiter) through the full pipeline, using an inferred HTML pattern until real examples are found.

**HTML Evidence**: ❌ **NOT FOUND** in test articles (abrogar, abstraer, abuelo, acertar, adherir, alférez, alícuota, androfobia)

**Inferred Pattern**: Likely plain text `//`

The system SHOULD:
- Define a constant `phonemeGlyph = "//"`
- Add `InlineKindPhoneme` to the InlineKind enum
- Implement BEST-GUESS parsing logic based on inferred pattern
- Add explicit preservation in text normalization
- Add render case for phoneme marker
- Include WARNING comments in code indicating lack of HTML validation

#### Scenario: Phoneme marker implementation is marked as speculative (WARNING)

- GIVEN the codebase contains parsing, normalization, and rendering logic for `//`
- WHEN code review is performed
- THEN the implementation MUST include WARNING comments stating:
  - No real HTML example found in DPD articles
  - HTML pattern is inferred, not validated
  - Implementation may need updates when real examples are discovered
- AND acceptance MUST fail if warning annotations are missing

#### Scenario: Phoneme marker compiles and passes synthetic tests

- GIVEN synthetic test fixtures with inferred `//` HTML pattern
- WHEN tests are executed
- THEN the implementation MUST compile without errors
- AND synthetic tests MUST pass
- AND tests MUST be clearly marked as synthetic (not real DPD fixtures)

---

## Architecture Constraints

### Constraint: Extend Existing Architecture Without Refactoring

The implementation MUST extend the existing semantic preservation architecture without refactoring the switch-based rendering dispatch or the current HTML parsing flow.

**Rationale**: The ⊗ (exclusion) and → (cross-reference) signs work correctly with the current architecture. The switch-based rendering dispatch is legitimate and intentional.

The implementation MUST:
- Add new whitelist entries to `semanticSpanAllowed` without removing or restructuring existing entries
- Add new InlineKind constants without renaming or removing existing kinds
- Add new switch cases to `parseSupportedOpenTag()` and `renderInlineMarkdownItem()` without refactoring the switch structure
- Follow the ⊗ sign implementation pattern for all new signs

The implementation MUST NOT:
- Replace switch statements with polymorphic dispatch or strategy patterns
- Refactor the `preserveSemanticSpans()` → `extractInlines()` → `cleanInlineSegment()` → `renderInlineMarkdownItem()` flow
- Change how ⊗ or → currently work

---

### Constraint: Test Coverage MUST Distinguish Validated vs. Speculative

The test suite MUST clearly distinguish between VALIDATED tests (using real DPD HTML) and SPECULATIVE tests (using synthetic fixtures).

The implementation MUST:
- Use real DPD article HTML for validated sign tests (when fixtures are available)
- Mark speculative tests with clear comments indicating synthetic/inferred nature
- Add golden fixture tests for validated signs using real article HTML
- Add end-to-end integration tests covering all validated signs
- Include WARNING comments in speculative test cases

The implementation MUST NOT:
- Mix validated and speculative test cases without clear annotations
- Claim speculative tests as validation of real DPD behavior

---

## Acceptance Criteria

### For VALIDATED Signs (MUST Pass)

The implementation MUST pass all of the following:

1. **@ Sign (Digital Edition)**:
   - [ ] `<sup>` is whitelisted in `semanticSpanAllowed`
   - [ ] `digitalEditionGlyph` constant is defined
   - [ ] `InlineKindDigitalEdition` is defined
   - [ ] `<sup>@</sup>` is parsed and tagged with correct InlineKind
   - [ ] `@` character survives normalization
   - [ ] `@` character is emitted in Markdown output
   - [ ] End-to-end test with real alícuota.html fixture passes (when available)

2. **+ Sign (Construction Marker)**:
   - [ ] `<span class="nc">` is whitelisted in `semanticSpanAllowed`
   - [ ] `constructionMarkerGlyph` constant is defined
   - [ ] `InlineKindConstructionMarker` is defined
   - [ ] `<span class="nc">+ infinitivo</span>` is parsed and tagged with correct InlineKind
   - [ ] Construction marker phrase survives normalization
   - [ ] Construction marker is emitted in Markdown output
   - [ ] End-to-end test with real acertar.html fixture passes (when available)

3. **⊗ Sign (Exclusion)** - No Regression:
   - [ ] Existing ⊗ tests continue to pass
   - [ ] `<span class="bolaspa">` remains whitelisted
   - [ ] `exclusionGlyph` constant is unchanged
   - [ ] `InlineKindExclusion` is unchanged
   - [ ] End-to-end test with real alícuota.html or acertar.html fixture passes

4. **→ Sign (Cross-Reference)** - No Regression:
   - [ ] Existing → tests continue to pass
   - [ ] `<a>` tags remain whitelisted and processed correctly
   - [ ] Arrow character survives in canonical Markdown links
   - [ ] End-to-end test with real acertar.html fixture passes

5. **Bracket Semantics**:
   - [ ] `<span class="nn">` is whitelisted in `semanticSpanAllowed`
   - [ ] `<span class="yy">` is whitelisted in `semanticSpanAllowed`
   - [ ] `<dfn>` remains whitelisted (already present)
   - [ ] `InlineKindBracketDefinition` is defined
   - [ ] `InlineKindBracketPronunciation` is defined
   - [ ] `InlineKindBracketInterpolation` is defined
   - [ ] Each bracket context is parsed and tagged correctly
   - [ ] Brackets are emitted as plain text in Markdown output
   - [ ] JSON output distinguishes the three bracket contexts
   - [ ] End-to-end test with real abrogar.html, acertar.html, alícuota.html fixtures passes (when available)

6. **/ Sign (Slash)**:
   - [ ] Slash is preserved as plain text
   - [ ] No dedicated InlineKind for slash
   - [ ] No special whitelist entries for slash
   - [ ] Slash appears correctly in Markdown output

### For SPECULATIVE Signs (SHOULD Pass with Warnings)

The implementation SHOULD pass all of the following, with WARNING annotations:

1. **General Requirements for All Speculative Signs**:
   - [ ] Constants are defined for all 5 speculative signs (*, ‖, <, >, //)
   - [ ] InlineKind values are defined for all 5 speculative signs
   - [ ] Parsing logic is implemented for all 5 speculative signs
   - [ ] Normalization preserves all 5 speculative signs
   - [ ] Rendering emits all 5 speculative signs correctly
   - [ ] WARNING comments are present in all speculative sign implementations
   - [ ] WARNING comments state: "No real HTML validation. Pattern inferred. Update when examples found."

2. **Markdown Escaping for `<` and `>`**:
   - [ ] `<` is escaped correctly to prevent raw HTML interpretation
   - [ ] `>` is escaped correctly if necessary to prevent blockquote interpretation
   - [ ] Golden tests verify Markdown output renders correctly in live previews

3. **Synthetic Test Coverage**:
   - [ ] Synthetic test fixtures exist for all 5 speculative signs
   - [ ] All synthetic tests pass
   - [ ] Test names or comments clearly indicate synthetic/inferred nature

### Code Quality

The implementation MUST pass all of the following:

- [ ] Code review confirms extension follows ⊗ sign pattern
- [ ] No switch-statement refactoring (switch dispatch remains intact)
- [ ] No changes to existing ⊗ or → implementations (except additive extensions)
- [ ] All new InlineKind values follow naming conventions
- [ ] All new constants follow naming conventions
- [ ] `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md` is referenced in code comments

---

## Out of Scope

The following are explicitly OUT OF SCOPE for this specification:

- Changing the visual design or appearance of signs in output
- Substituting signs with alternatives or Unicode normalization
- Refactoring the switch-based rendering dispatch
- Modifying how ⊗ and → currently work (beyond additive extensions)
- Adding new signs beyond the 11 official DPD signs
- Backward-compatibility shims (this is a correctness fix)
- Providing real HTML fixtures in this spec (fixtures are implementation artifacts)

---

## References

- **Proposal**: `openspec/changes/dpd-signs-preservation/proposal.md`
- **HTML Analysis**: `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md`
- **Current DPD Spec**: `openspec/specs/dpd/spec.md`
- **Reference Implementation**: ⊗ (exclusion) sign in `internal/parse/dpd.go`, `internal/model/types.go`, `internal/normalize/dpd.go`, `internal/renderutil/inline.go`

# Proposal: DPD Signs Preservation

## Intent

The DPD dictionary uses 11 official typographical signs with specific semantic meanings (cross-references, incorrect forms, grammatical constructions, transformations, phonemes, etc.). Currently, only 2 of these 11 signs preserve their semantic meaning through the HTML → Parse → Normalize → Render pipeline. The remaining signs are either partially degraded (losing semantic meaning) or completely lost (stripped from output).

This change will extend the existing semantic preservation architecture to handle **9 active DPD signs** consistently, maintaining their semantic meaning from source HTML through to final JSON/Markdown output without design changes or substitutions.

**Updated 2026-03-17**: 
- Analysis of 8 real DPD articles completed. Sign-to-HTML mapping documented in `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md`.
- 6 signs validated with real HTML, 3 signs implemented with inferred patterns pending validation.
- **2 signs archived** (`<`, `>`) due to HTML tag collision risk and low value-to-risk ratio.

## Scope

### In Scope

- **Extend semantic span whitelist**: Add validated HTML containers to `semanticSpanAllowed`:
  - ✅ `<sup>` (for @ sign - CRITICAL, currently stripped)
  - ✅ `<span class="nc">` (for + sign - currently stripped)
  - ✅ `<span class="nn">` (for pronunciation brackets)
  - ✅ `<span class="yy">` (for example/interpolation brackets)
  - ⚠️ Additional tags TBD for unvalidated signs (*, ‖, //)

- **Add sign constants**: Define 9 active DPD signs as typed constants (following `exclusionGlyph` pattern)

- **Extend InlineKind enum**: Create semantic kinds for:
  - **VALIDATED** (real HTML evidence):
    - `InlineKindDigitalEdition` (@) - `<sup>@</sup>`
    - `InlineKindConstructionMarker` (+) - `<span class="nc">+ infinitivo</span>`
    - `InlineKindBracketDefinition` ([...]) - `<dfn>[...]</dfn>`
    - `InlineKindBracketPronunciation` ([...]) - `<span class="nn">[...]</span>`
    - `InlineKindBracketInterpolation` ([...]) - `<span class="yy">[...]</span>`
  - **SPECULATIVE** (inferred, pending validation):
    - `InlineKindAgrammatical` (*) - TBD
    - `InlineKindHypothetical` (‖) - TBD
    - `InlineKindPhoneme` (//) - TBD

- **Slash (/) handling**: Keep as plain text (no special semantic preservation needed per user decision)

- **Extend HTML parsing**: Add cases in `parseSupportedOpenTag()` switch for validated containers

- **Extend text cleaning**: Add explicit preservation rules in `cleanText()`/`cleanInlineSegment()` for all signs

- **Extend JSON/Markdown rendering**: Add cases in `renderInlineMarkdownItem()` switch to emit correct output for each semantic kind

- **Comprehensive test coverage**:
  - Use real DPD articles as fixtures: `alícuota.html`, `acertar.html`, `abrogar.html`
  - Parse tests for each VALIDATED sign in its HTML container
  - Normalize tests for semantic preservation
  - Render integration tests with golden fixtures
  - End-to-end tests covering all validated signs
  - ⚠️ Warning comments for SPECULATIVE signs indicating they lack real HTML validation

### Out of Scope

- Changing the visual design or appearance of the signs in output
- Substituting signs with alternatives or Unicode normalization
- Refactoring the switch-based rendering dispatch (this is legitimate and stays)
- Modifying how ⊗ and → currently work (these are the reference implementations)
- Adding new signs beyond the 9 active DPD signs
- Backward-compatibility shims (this is a correctness fix, not a breaking API change)

### Archived Signs (Not Implemented)

The following 2 signs are **ARCHIVED** and excluded from this implementation:

| Sign | Description | Reason for Archival |
|------|-------------|---------------------|
| `<` | Etymology ("comes from") | HTML tag collision risk; low appearance ratio; not found in high-value articles |
| `>` | Transformation ("passes to") | HTML tag collision risk; low appearance ratio; not found in high-value articles |

**Rationale**: These signs pose significant risk of collision with HTML parsing (`<span>`, `<div>`, `<a>`, etc.) and could introduce bugs in search and rendering. Their low appearance frequency and absence in high-value articles make the risk-to-benefit ratio unfavorable.

**Future Consideration**: These signs remain documented in `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md` for potential future implementation if:
1. Real DPD articles containing these signs in high-value contexts are discovered
2. A collision-safe parsing strategy is developed
3. User demand justifies the implementation complexity

## Approach

### Architecture Pattern (Proven and Working)

The ⊗ (exclusion) sign provides the reference implementation:
1. **Constant**: `exclusionGlyph = "⊗"` in `internal/parse/dpd.go`
2. **Whitelist**: `semanticSpanAllowed["dfn"] = true` in `preserveSemanticSpans()`
3. **InlineKind**: `InlineKindExclusion` in `internal/model/inline.go`
4. **Parse recognition**: Switch case in `parseSupportedOpenTag()` sets the kind
5. **Explicit preservation**: `cleanInlineSegment()` checks for exclusion glyph and protects it
6. **Render dispatch**: `renderInlineMarkdownItem()` has a case for `InlineKindExclusion`

This pattern works correctly. We extend it for the remaining signs.

### Implementation Strategy

#### Phase 1: Critical Fixes (@, + Signs) - VALIDATED

**Priority: CRITICAL** - These signs are currently being STRIPPED because their container tags are not whitelisted.

**@ Sign (Digital Edition Marker)**:
- HTML: `<sup>@</sup>` (confirmed in alícuota, androfobia)
- Add `<sup>` to `semanticSpanAllowed` whitelist
- Add `digitalEditionGlyph = "@"` constant
- Add `InlineKindDigitalEdition` to enum
- Extend `parseSupportedOpenTag()` with `case "sup"` that checks for @ content
- Add explicit preservation in `cleanInlineSegment()`
- Add render case for digital edition marker
- **Success check**: @ sign survives end-to-end in alícuota test fixture

**+ Sign (Construction Marker)**:
- HTML: `<span class="nc">+ infinitivo</span>` (confirmed in acertar)
- Add `"<span class=\"nc\">": true` to `semanticSpanAllowed` whitelist
- Add `constructionMarkerGlyph = "+"` constant
- Add `InlineKindConstructionMarker` to enum
- Extend `parseSupportedOpenTag()` with case for `<span class="nc">`
- Add explicit preservation in `cleanInlineSegment()`
- Add render case for construction marker
- **Success check**: + sign survives end-to-end in acertar test fixture

#### Phase 2: Bracket Semantics Preservation - VALIDATED

**Priority: HIGH** - Brackets work as text but lose semantic context (pronunciation vs. definition vs. interpolation).

**Three bracket contexts to distinguish**:
1. **Definition/Correction**: `<dfn>[una ley]</dfn>` → `InlineKindBracketDefinition`
2. **Pronunciation**: `<span class="nn">[alikuóto]</span>` → `InlineKindBracketPronunciation`
3. **Interpolation/Example**: `<span class="yy">[las feministas]</span>` → `InlineKindBracketInterpolation`

**Implementation**:
- Add `"<span class=\"nn\">": true` and `"<span class=\"yy\">": true` to whitelist (dfn already whitelisted)
- Add three InlineKind constants for bracket contexts
- Extend `parseSupportedOpenTag()` to detect bracket content and set correct kind based on parent tag
- Preserve brackets as-is in text cleaning (already work)
- Add render cases that maintain semantic distinction in JSON (Markdown output keeps brackets as plain text)
- **Success check**: JSON output distinguishes bracket contexts correctly

#### Phase 3: Speculative Signs (*, ‖, //) - INFERRED PATTERNS

**Priority: MEDIUM** - No real HTML validation yet, implementing based on patterns learned from validated signs.

**⚠️ WARNING**: These implementations are SPECULATIVE until real DPD articles containing these signs are found and analyzed.

**⚠️ ARCHIVED SIGNS EXCLUDED**: `<` and `>` signs are NOT implemented due to HTML tag collision risk.

**Likely patterns for remaining speculative signs**:
- `*` (agrammatical): Probably `<span class="??">*</span>` (similar to ⊗ pattern)
- `‖` (hypothetical): Probably `<span class="??">‖</span>` (similar to ⊗ pattern)
- `//` (phoneme): Probably plain text `//`

**Implementation approach**:
- Add constants for each sign
- Add InlineKind enum values
- Implement BEST-GUESS parsing logic based on observed patterns
- Add prominent WARNING comments in code indicating lack of validation
- Add tests with synthetic HTML (not real DPD fixtures)
- Document in SIGN_ANALYSIS.md that these need validation when examples found

**Success check**: Code compiles, tests pass with synthetic fixtures, warnings documented

#### Phase 4: Integration and Golden Tests

- Use real DPD articles as test fixtures: `alícuota.html`, `acertar.html`, `abrogar.html`
- Add parse → normalize → render integration tests for VALIDATED signs
- Create golden files with expected JSON/Markdown output
- Add regression tests to prevent future degradation
- **Success check**: All tests pass for validated signs, speculative signs have warning annotations
- Create golden files with expected JSON/Markdown output
- Add regression tests to prevent future degradation
- **Success check**: All tests pass for validated signs, speculative signs have warning annotations

### Real HTML Evidence Summary

**Validated Signs (6)** - Confirmed with real DPD articles:

| Sign | HTML | Article(s) | Line Reference |
|------|------|-----------|----------------|
| ⊗ | `<span class="bolaspa">⊗&#x200D;</span>` | alícuota, acertar | alícuota:545, acertar:547 |
| @ | `<sup>@</sup>` | alícuota, androfobia | alícuota:545 |
| + | `<span class="nc">+ infinitivo</span>` | acertar | acertar:546 |
| → | `<a href="/dpd/...">→ ...</a>` | acertar | acertar:543 |
| [ ] | `<dfn>[...]</dfn>`, `<span class="nn">[...]</span>`, `<span class="yy">[...]</span>` | abrogar, acertar, alícuota, androfobia | Multiple |
| / | Plain text in article body | All articles | Multiple (no semantic encoding) |

**Speculative Signs (3)** - No real HTML found yet:

| Sign | Description | Inferred Pattern | Confidence |
|------|-------------|------------------|------------|
| * | Agrammatical | Likely `<span class="??">*</span>` | Medium |
| ‖ | Hypothetical | Likely `<span class="??">‖</span>` | Medium |
| // | Phoneme | Plain text `//` | Medium |

**Archived Signs (2)** - Excluded from implementation:

| Sign | Description | Reason |
|------|-------------|--------|
| < | Etymology | HTML tag collision risk + low value-to-risk ratio |
| > | Transformation | HTML tag collision risk + low value-to-risk ratio |

**Reference**: Full analysis in `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md`

### Technical Details

**File: `internal/parse/dpd.go`**
- Add sign constants for 9 active signs at package level
- Extend `semanticSpanAllowed` map with validated containers:
  - ✅ `<sup>` (@ sign)
  - ✅ `<span class="nc">` (+ sign)
  - ✅ `<span class="nn">` (pronunciation brackets)
  - ✅ `<span class="yy">` (interpolation brackets)
- Extend `parseSupportedOpenTag()` switch with new cases for validated containers
- Add helper functions for bracket context detection if needed
- ⚠️ Add WARNING comments for speculative sign implementations (*, ‖, //)
- 📦 Add comment documenting archived signs (`<`, `>`) for future reference

**File: `internal/model/types.go`**
- Add new `InlineKind` constants:
  - VALIDATED: `InlineKindDigitalEdition`, `InlineKindConstructionMarker`, `InlineKindBracketDefinition`, `InlineKindBracketPronunciation`, `InlineKindBracketInterpolation`
  - SPECULATIVE: `InlineKindAgrammatical`, `InlineKindHypothetical`, `InlineKindPhoneme`
  - 📦 Add comment documenting archived kinds for `<` and `>` signs

**File: `internal/normalize/dpd.go`**
- Extend `cleanInlineSegment()` with explicit checks for each sign
- Add `cleanText()` rules if signs appear in non-semantic contexts

**File: `internal/renderutil/inline.go`**
- Extend `renderInlineMarkdown()` switch with new cases for all InlineKind values
- Ensure each case emits correct Markdown (raw text, escaped where necessary)
- Add WARNING comments for speculative sign cases

**File: `internal/parse/dpd_test.go`**
- Add test cases for each sign's HTML source pattern

**File: `internal/normalize/dpd_test.go`**
- Add normalization tests preserving sign semantics

**File: `testdata/dpd/*.html`**
- Promote validated articles to test fixtures:
  - `alícuota.html` (contains @, ⊗, pronunciation brackets)
  - `acertar.html` (contains +, ⊗, →, definition brackets)
  - `abrogar.html` (contains definition brackets)

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/parse/dpd.go` | Modified | Add 9 sign constants, extend whitelist with 4 validated containers, extend parsing switch, add archive comments |
| `internal/model/types.go` | Modified | Add 8 new `InlineKind` constants (5 validated, 3 speculative), add archive comments |
| `internal/normalize/dpd.go` | Modified | Extend cleaning logic with explicit preservation rules for 9 active signs |
| `internal/renderutil/inline.go` | Modified | Extend rendering switch with 8 new cases, add WARNING/archive comments |
| `internal/parse/dpd_test.go` | Modified | Add parse tests for validated signs with real HTML |
| `internal/normalize/dpd_test.go` | Modified | Add normalization tests for all signs |
| `testdata/dpd/alícuota.html` | New | Real DPD article fixture (@ sign, pronunciation brackets) |
| `testdata/dpd/acertar.html` | New | Real DPD article fixture (+ sign, definition brackets) |
| `testdata/dpd/abrogar.html` | New | Real DPD article fixture (definition brackets) |
| `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md` | New | Comprehensive sign-to-HTML mapping documentation |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| **Speculative signs fail in production**: 3 signs lack real HTML validation | MEDIUM | Document with WARNING comments; implement defensively; update when real examples found |
| **Archived signs requested later**: Users need `<` or `>` signs | LOW | Signs documented in SIGN_ANALYSIS.md; can be implemented if safe parsing strategy found |
| **Sign collision**: Multiple signs share same HTML container | Medium | Phase-based rollout; test each sign independently before combining |
| **Bracket context ambiguity**: May mis-classify bracket types | Low | Three distinct HTML containers validated; low collision risk |
| **Existing behavior change**: Extending whitelist affects other HTML parsing | Low | Whitelist is currently strict; new entries only affect previously-stripped content |
| **Performance**: Additional switch cases in hot path | Very Low | Switch dispatch is O(1); negligible impact on small inline items |

## Rollback Plan

1. **Phase-based rollout enables partial rollback**: If Phase 2 introduces issues, revert commits for that phase while keeping Phase 1 fixes
2. **Feature flag not needed**: This is a correctness fix, not a user-facing feature toggle
3. **Test-driven validation**: Each phase includes tests; rollback decision based on test failures
4. **Git revert**: Standard `git revert` for problematic commits
5. **Backward compatibility**: Extending enums and adding cases doesn't break existing code paths

## Dependencies

- **Existing architecture**: This change depends on the current `preserveSemanticSpans()` → `parseSupportedOpenTag()` → `cleanInlineSegment()` → `renderInlineMarkdownItem()` flow remaining intact
- **Test fixtures**: Requires access to DPD HTML samples containing all 11 signs (user provided official list)
- **No external dependencies**: No new libraries or tools required

## Success Criteria

**VALIDATED Signs (Must Pass)**:
- [ ] @ sign (digital edition) has constant, InlineKind, whitelist entry, parse/render cases
- [ ] + sign (construction marker) has constant, InlineKind, whitelist entry, parse/render cases
- [ ] Bracket semantics preserved: 3 distinct InlineKind values for definition/pronunciation/interpolation contexts
- [ ] ⊗ sign (exclusion) continues working (no regression)
- [ ] → sign (cross-reference) continues working (no regression)
- [ ] Parse tests pass for all validated signs using real DPD HTML
- [ ] Normalize tests pass for semantic preservation
- [ ] Integration tests with real DPD article fixtures (alícuota, acertar, abrogar) pass
- [ ] JSON output correctly distinguishes bracket contexts

**SPECULATIVE Signs (Best-Effort)**:
- [ ] Constants, InlineKind values, and render cases exist for 3 speculative signs (*, ‖, //)
- [ ] WARNING comments document lack of HTML validation
- [ ] Code compiles and tests pass with synthetic fixtures
- [ ] Documentation notes these signs pending validation

**ARCHIVED Signs (Documented)**:
- [ ] Archive comments in code document exclusion rationale for `<` and `>` signs
- [ ] SIGN_ANALYSIS.md documents archived signs for future reference

**General**:
- [ ] Code review confirms extension follows existing patterns (no refactor)
- [ ] `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md` comprehensive documentation exists

## Next Steps

1. ✅ **Real HTML analysis complete**: 8 DPD articles analyzed, sign-to-HTML mapping documented
2. ✅ **Proposal updated**: Validated vs. speculative signs distinguished
3. **Proceed to sdd-spec**: Define explicit requirements and scenarios for:
   - VALIDATED signs (with real HTML evidence)
   - SPECULATIVE signs (with warning annotations)
   - Bracket semantic preservation (3 contexts)
4. **Proceed to sdd-design**: Detail technical implementation for each phase
5. **Proceed to sdd-tasks**: Break down into granular implementation tasks

## Notes

- **Architecture decision preserved**: Keep the switch statement for rendering (user confirmed this is legitimate dispatch)
- **Reference implementation**: ⊗ (exclusion) sign is the gold standard; replicate its pattern
- **Phase order rationale**: @ and + signs are CRITICAL (currently stripped), bracket semantics are HIGH priority (JSON needs distinction), speculative signs are MEDIUM priority (implement defensively with warnings)
- **Slash (/) decision**: Keep as plain text per user confirmation (no special semantic preservation needed)
- **Bracket semantics decision**: Preserve 3 contexts for JSON output per user confirmation (Markdown keeps plain brackets)

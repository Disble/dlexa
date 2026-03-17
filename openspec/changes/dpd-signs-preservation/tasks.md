# Tasks: DPD Signs Preservation

## Phase 1: Critical Signs (@ and +) — HIGHEST PRIORITY

These signs are currently **STRIPPED**. Fixing them is urgent.

- [x] 1.1 Add sign constants to `internal/parse/dpd.go` (after line 329)
  - Add `digitalEditionGlyph = "@"`
  - Add `constructionMarkerGlyph = "+"`
  - **Acceptance**: Constants defined and compile successfully

- [x] 1.2 Add InlineKind constants to `internal/model/types.go` (after line 35)
  - Add `InlineKindDigitalEdition = "digital_edition"`
  - Add `InlineKindConstructionMarker = "construction_marker"`
  - **Acceptance**: InlineKind enum extended without syntax errors

- [x] 1.3 Extend whitelist in `internal/parse/dpd.go` (lines 598-616)
  - Add `"<sup>": true` to `semanticSpanAllowed` map
  - Add `"</sup>": true` to `semanticSpanAllowed` map
  - **Note**: `<span class="nc">`, `<span class="nn">`, `<span class="yy">` are already whitelisted
  - **Acceptance**: Whitelist contains sup tags; parser accepts <sup> tags without stripping

- [x] 1.4 Extend `parseSupportedOpenTag()` switch in `internal/parse/dpd.go`
  - Add `case "sup":` to parse `<sup>@</sup>` → `InlineKindDigitalEdition`
  - Extend `case "span":` to parse `<span class="nc">+ ...</span>` → `InlineKindConstructionMarker`
  - **Acceptance**: Parser recognizes @ in <sup> and + in <span class="nc"> with correct InlineKind

- [x] 1.5 Extend `cleanInlineSegment()` in `internal/parse/dpd.go` (after line 685)
  - Add preservation for `digitalEditionGlyph` using `strings.ReplaceAll()` with `//nolint:gocritic` comment
  - Add preservation for `constructionMarkerGlyph` using `strings.ReplaceAll()` with `//nolint:gocritic` comment
  - **Acceptance**: @ and + signs survive cleanInlineSegment() normalization

- [x] 1.6 Extend `renderInlineMarkdownItem()` switch in `internal/renderutil/inline.go` (before `default:` case at line ~144)
  - Add `case model.InlineKindDigitalEdition:` → return text (@ preserved as-is)
  - Add `case model.InlineKindConstructionMarker:` → return text (+ phrase preserved as-is)
  - **Acceptance**: Markdown rendering emits @ and + signs correctly

- [x] 1.7 Promote real test fixtures to `internal/parse/testdata/`
  - Copy `scripts/testdata/dpd-signs-analysis/alícuota.html` to `internal/parse/testdata/alícuota.html`
  - Copy `scripts/testdata/dpd-signs-analysis/acertar.html` to `internal/parse/testdata/acertar.html`
  - **Acceptance**: Fixture files exist in testdata/ directory

- [x] 1.8 Add parse tests for @ sign in `internal/parse/dpd_test.go`
  - Test: `<sup>@</sup>` from alícuota.html → `InlineKindDigitalEdition`
  - Test: @ character survives in parsed text
  - **Acceptance**: Unit test passes with real HTML fixture

- [x] 1.9 Add parse tests for + sign in `internal/parse/dpd_test.go`
  - Test: `<span class="nc">+ infinitivo</span>` from acertar.html → `InlineKindConstructionMarker`
  - Test: + phrase survives in parsed text with correct whitespace
  - **Acceptance**: Unit test passes with real HTML fixture

- [x] 1.10 Add normalization tests in `internal/normalize/dpd_test.go`
  - Test: @ sign survives `cleanInlineText()` function
  - Test: + sign survives `cleanInlineText()` function
  - **Acceptance**: Unit tests verify signs survive normalization

- [x] 1.11 Add end-to-end integration test for alícuota.html
  - Test: Parse full alícuota.html → Markdown output contains @ sign
  - Test: JSON output contains `InlineKindDigitalEdition`
  - **Acceptance**: Integration test passes with real DPD article

- [x] 1.12 Add end-to-end integration test for acertar.html
  - Test: Parse full acertar.html → Markdown output contains `+ infinitivo`
  - Test: JSON output contains `InlineKindConstructionMarker`
  - **Acceptance**: Integration test passes with real DPD article

## Phase 2: Bracket Semantics — HIGH PRIORITY

Brackets work as text but lose semantic distinction in JSON. This phase adds semantic tagging.

- [x] 2.1 Add InlineKind constants to `internal/model/types.go` (after Phase 1 constants)
  - Add `InlineKindBracketDefinition = "bracket_definition"`
  - Add `InlineKindBracketPronunciation = "bracket_pronunciation"`
  - Add `InlineKindBracketInterpolation = "bracket_interpolation"`
  - Add comment: "Three distinct semantic contexts for brackets based on HTML container"
  - **Acceptance**: Three bracket InlineKind values defined

- [x] 2.2 Extend `parseSupportedOpenTag()` switch in `internal/parse/dpd.go`
  - Add `case "dfn":` → return `InlineKindBracketDefinition` (for `<dfn>[...]</dfn>`)
  - Extend `case "span":` to parse `<span class="nn">[...]</span>` → `InlineKindBracketPronunciation`
  - Extend `case "span":` to parse `<span class="yy">[...]</span>` → `InlineKindBracketInterpolation`
  - **Acceptance**: Parser assigns correct InlineKind based on parent container

- [x] 2.3 Extend `renderInlineMarkdownItem()` switch in `internal/renderutil/inline.go`
  - Add multi-case for all 3 bracket kinds: `case model.InlineKindBracketDefinition, model.InlineKindBracketPronunciation, model.InlineKindBracketInterpolation:`
  - Return text as-is (plain brackets preserved in Markdown)
  - Add comment: "Brackets as plain text in Markdown; semantic distinction preserved in JSON via InlineKind"
  - **Acceptance**: Markdown output contains plain `[...]` brackets

- [x] 2.4 Promote real test fixture to `internal/parse/testdata/`
  - Copy `scripts/testdata/dpd-signs-analysis/abrogar.html` to `internal/parse/testdata/abrogar.html`
  - **Acceptance**: abrogar.html fixture exists in testdata/

- [x] 2.5 Add parse tests for bracket contexts in `internal/parse/dpd_test.go`
  - Test: `<dfn>[una ley]</dfn>` from abrogar.html → `InlineKindBracketDefinition`
  - Test: `<span class="nn">[alikuóto]</span>` from alícuota.html → `InlineKindBracketPronunciation`
  - Test: `<span class="yy">[las feministas]</span>` from androfobia.html → `InlineKindBracketInterpolation`
  - **Acceptance**: Parser correctly distinguishes 3 bracket contexts

- [x] 2.6 Add integration tests for bracket semantic tagging
  - Test: abrogar.html → JSON output contains `InlineKindBracketDefinition` for definition brackets
  - Test: alícuota.html → JSON output contains `InlineKindBracketPronunciation` for pronunciation
  - Test: acertar.html or androfobia.html → JSON output contains `InlineKindBracketInterpolation`
  - **Acceptance**: JSON distinguishes bracket contexts; Markdown shows plain brackets

## Phase 3: Speculative Signs (*, ‖, //) — MEDIUM PRIORITY

⚠️ **No HTML validation**. Implement defensively with WARNING comments.

- [x] 3.1 Add speculative sign constants to `internal/parse/dpd.go` (after Phase 1 constants)
  - Add WARNING comment block: "Speculative signs - no real HTML validation found in test articles. HTML patterns are inferred from validated sign patterns. Update when real DPD examples are discovered."
  - Add `agrammaticalGlyph = "*"`
  - Add `hypotheticalGlyph = "‖"` (with Unicode comment: \u2016)
  - Add `phonemeGlyph = "//"`
  - Add ARCHIVED SIGNS comment referencing SIGN_ANALYSIS.md for `<` and `>` exclusion rationale
  - **Acceptance**: Constants defined with WARNING annotations

- [x] 3.2 Add speculative InlineKind constants to `internal/model/types.go` (after bracket constants)
  - Add WARNING comment: "SPECULATIVE (inferred patterns, no HTML validation). These patterns are best-guess based on validated signs. Update when real DPD examples are discovered."
  - Add `InlineKindAgrammatical = "agrammatical"`
  - Add `InlineKindHypothetical = "hypothetical"`
  - Add `InlineKindPhoneme = "phoneme"`
  - Add ARCHIVED SIGNS comment: "< (etymology) - HTML tag collision risk. > (transformation) - HTML tag collision risk. See testdata/dpd-signs-analysis/SIGN_ANALYSIS.md for archival rationale."
  - **Acceptance**: Speculative InlineKind values defined with WARNING annotations

- [x] 3.3 Extend `parseSupportedOpenTag()` switch in `internal/parse/dpd.go`
  - Add WARNING comment: "Phase 3: Speculative signs (WARNING: no HTML validation). Inferred patterns - update when real examples found."
  - Extend `case "span":` to check for `agrammaticalGlyph` in text → `InlineKindAgrammatical`
  - Extend `case "span":` to check for `hypotheticalGlyph` in text → `InlineKindHypothetical`
  - Extend `case "span":` to check for `phonemeGlyph` in text → `InlineKindPhoneme`
  - **Acceptance**: Parser handles speculative signs with inferred patterns

- [x] 3.4 Extend `cleanInlineSegment()` in `internal/parse/dpd.go` (after Phase 1 preservation)
  - Add WARNING comment: "Phase 3: Preserve speculative signs (WARNING: no HTML validation)"
  - Add preservation for `agrammaticalGlyph` with `//nolint:gocritic` and "(SPECULATIVE)" comment
  - Add preservation for `hypotheticalGlyph` with `//nolint:gocritic` and "(SPECULATIVE)" comment
  - Add preservation for `phonemeGlyph` with `//nolint:gocritic` and "(SPECULATIVE)" comment
  - **Acceptance**: Speculative signs survive cleanInlineSegment() normalization

- [x] 3.5 Extend `cleanInlineText()` in `internal/normalize/dpd.go` (lines 542-547)
  - Replace current implementation with placeholder-based preservation
  - Add preservedSigns slice with validated + speculative signs (with WARNING comment for speculative)
  - Implement: replace signs with placeholders → normalize whitespace → restore signs
  - **Rationale**: Prevents `strings.Fields()` from collapsing whitespace around signs like `+`
  - **Acceptance**: All signs (validated + speculative) survive field normalization

- [x] 3.6 Extend `renderInlineMarkdownItem()` switch in `internal/renderutil/inline.go`
  - Add WARNING comment: "Phase 3: SPECULATIVE signs (WARNING: no HTML validation)"
  - Add `case model.InlineKindAgrammatical:` → return text with comment "(SPECULATIVE - no real HTML found)"
  - Add `case model.InlineKindHypothetical:` → return text with comment "(SPECULATIVE - no real HTML found)"
  - Add `case model.InlineKindPhoneme:` → return text with comment "(SPECULATIVE - no real HTML found)"
  - **Acceptance**: Markdown rendering emits speculative signs as plain text

- [x] 3.7 Create synthetic test fixtures for speculative signs
  - Create test HTML with inferred patterns in test file (do NOT promote to testdata/ - keep synthetic)
  - Mark test clearly: "SYNTHETIC TEST - NO REAL HTML VALIDATION"
  - Add tests for `*` sign with inferred `<span class="??">*</span>` pattern
  - Add tests for `‖` sign with inferred `<span class="??">‖</span>` pattern
  - Add tests for `//` sign with inferred plain text pattern
  - **Acceptance**: Synthetic tests compile and pass; clearly marked as synthetic

## Phase 4: Integration & Golden Tests — FINAL VALIDATION

Comprehensive end-to-end validation with real DPD articles.

- [x] 4.1 Create golden output files for alícuota.html
  - Generate expected JSON output with all validated signs (@ + ⊗ + pronunciation brackets)
  - Generate expected Markdown output
  - Save as `internal/parse/testdata/alícuota.golden.json` and `alícuota.golden.md`
  - **Acceptance**: Golden files reflect correct end-to-end parsing

- [x] 4.2 Create golden output files for acertar.html
  - Generate expected JSON output with all validated signs (+ + ⊗ + → + definition brackets)
  - Generate expected Markdown output
  - Save as `internal/parse/testdata/acertar.golden.json` and `acertar.golden.md`
  - **Acceptance**: Golden files reflect correct end-to-end parsing

- [x] 4.3 Create golden output files for abrogar.html
  - Generate expected JSON output with definition brackets
  - Generate expected Markdown output
  - Save as `internal/parse/testdata/abrogar.golden.json` and `abrogar.golden.md`
  - **Acceptance**: Golden files reflect correct end-to-end parsing

- [x] 4.4 Add golden integration test comparing parsed output to golden files
  - Test: Parse alícuota.html → compare JSON/Markdown against golden files
  - Test: Parse acertar.html → compare JSON/Markdown against golden files
  - Test: Parse abrogar.html → compare JSON/Markdown against golden files
  - **Acceptance**: Integration tests pass; output matches golden files exactly

- [x] 4.5 Run regression tests for existing ⊗ and → signs
  - Verify: Existing ⊗ (exclusion) tests continue to pass (no regression)
  - Verify: Existing → (cross-reference) tests continue to pass (no regression)
  - **Acceptance**: All existing sign tests pass without modification

- [x] 4.6 Run full test suite and verify no regressions
  - Run: `go test ./internal/parse/... -v`
  - Run: `go test ./internal/normalize/... -v`
  - Run: `go test ./internal/renderutil/... -v`
  - **Acceptance**: All tests pass; no regressions detected

- [x] 4.7 Update SIGN_ANALYSIS.md with implementation status
  - Mark @ sign as IMPLEMENTED (validated)
  - Mark + sign as IMPLEMENTED (validated)
  - Mark bracket semantics as IMPLEMENTED (validated)
  - Mark speculative signs as IMPLEMENTED (speculative - pending validation)
  - Document archival decision for `<` and `>` signs
  - **Acceptance**: Documentation reflects implementation status accurately

- [x] 4.8 Add code comments referencing SIGN_ANALYSIS.md
  - In `internal/parse/dpd.go`: Add comment near archived signs constants referencing `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md`
  - In `internal/model/types.go`: Add comment near archived InlineKind section
  - **Acceptance**: Code references authoritative documentation for archival decisions

## Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1: Critical Signs | 12 | Fix @ and + (currently STRIPPED) |
| Phase 2: Bracket Semantics | 6 | Add semantic tagging for 3 bracket contexts |
| Phase 3: Speculative Signs | 7 | Implement *, ‖, // with WARNING annotations |
| Phase 4: Integration | 8 | Golden tests and regression validation |
| **Total** | **33** | |

## Implementation Order

1. **Phase 1 MUST complete first** — @ and + signs are currently stripped (critical bug fix)
2. **Phase 2 follows Phase 1** — bracket semantics extend validated sign infrastructure
3. **Phase 3 independent of Phase 2** — can run in parallel after Phase 1 completes
4. **Phase 4 MUST run last** — validates all phases together with golden tests

## Dependencies

- Phase 1: No dependencies (foundational)
- Phase 2: Depends on Phase 1 (InlineKind pattern established)
- Phase 3: Depends on Phase 1 (constants and preservation pattern established)
- Phase 4: Depends on Phases 1, 2, and 3 (validates complete implementation)

## Testing Strategy

- **Validated signs (@ + brackets)**: Use real DPD HTML fixtures from testdata/
- **Speculative signs (*, ‖, //)**: Use synthetic test cases clearly marked as SYNTHETIC
- **Regression**: Existing ⊗ and → tests MUST continue passing (no changes to their implementation)
- **Integration**: Golden file tests validate end-to-end parsing with real DPD articles

## Next Step

Ready for **sdd-apply** phase. Recommend starting with Phase 1 (tasks 1.1-1.12) to fix critical @ and + sign stripping immediately.

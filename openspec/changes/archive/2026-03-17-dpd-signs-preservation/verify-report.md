## Verification Report

**Change**: dpd-signs-preservation  
**Mode**: openspec  
**Date**: 2026-03-17  
**Verdict**: PASS WITH WARNINGS

---

### Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 33 |
| Tasks complete | 33 |
| Tasks incomplete | 0 |

All checklist items in `openspec/changes/dpd-signs-preservation/tasks.md` are marked complete.

---

### Execution Evidence

**Configured verify test command**: `go test ./...`  
**Configured build command**: not configured (`openspec/config.yaml` sets `build_command: ""`)  
**Coverage threshold**: not configured (`0`)

**Full test suite**: ✅ Passed

- Command: `go test ./...`
- Result: exit code `0`
- Evidence: all repo test packages passed, including `internal/parse`, `internal/normalize`, `internal/render`, and `internal/renderutil`

**Targeted DPD sign verification**: ✅ Passed

- Command: `go test -v ./internal/parse ./internal/normalize ./internal/renderutil ./internal/render -run DPDSigns`
- Result: exit code `0`
- Evidence highlights:
  - `TestDPDSignsParsePhase1`
  - `TestDPDSignsBracketContextsStayBoundToImmediateParent`
  - `TestDPDSignsNormalizePhase1`
  - `TestDPDSignsBracketSemanticTaggingUsesRealFixtures`
  - `TestDPDSignsBracketSemanticTaggingKeepsMixedContextsDistinctInOneArticle`
  - `TestDPDSignsValidatedFixturesMatchGoldenOutputs`
  - `TestDPDSignsValidatedFixturesDoNotRegressExclusionAndReferenceMarkers`

**Targeted render verification**: ✅ Passed

- Command: `go test -v ./internal/renderutil -run "RenderInlineMarkdownSpeculativeSignsSynthetic|RenderInlineMarkdown/(BracketDefinition remains plain text|BracketPronunciation remains plain text|BracketInterpolation remains plain text)"`
- Result: exit code `0`

**Lint**: ✅ Passed

- Command: `go tool --modfile=golangci-lint.mod golangci-lint run ./...`
- Result: `0 issues.`

**Build**: ➖ Skipped / not configured

- `openspec/config.yaml` leaves `rules.verify.build_command` empty.
- Project instructions explicitly say never build after changes.

---

### Prior Verify Failures Re-checked

| Prior Failure | Re-check Result | Evidence |
|--------------|-----------------|----------|
| non-`@` `<sup>` fallback behavior | ✅ RESOLVED | `internal/parse/dpd.go` only assigns `InlineKindDigitalEdition` when preview text is exactly `@`; `internal/parse/dpd_test.go > TestDPDSignsParsePhase1` proves `<sup>2</sup>` stays `InlineKindText` with text `2` |
| bracket ambiguity / multiple bracket contexts in same flow | ✅ RESOLVED | `internal/parse/dpd_test.go > TestDPDSignsBracketContextsStayBoundToImmediateParent` and `internal/render/dpd_signs_integration_test.go > TestDPDSignsBracketSemanticTaggingKeepsMixedContextsDistinctInOneArticle` both passed |

---

### Spec Compliance Matrix

| Requirement | Scenario | Test / Evidence | Result |
|-------------|----------|-----------------|--------|
| @ digital edition | Digital edition marker survives end-to-end parsing | `internal/parse/dpd_test.go > TestDPDSignsParsePhase1`; `internal/render/dpd_signs_integration_test.go > TestDPDSignsValidatedFixturesMatchGoldenOutputs/alícuota` | ✅ COMPLIANT |
| @ digital edition | Superscript whitelist does not interfere with non-sign content | `internal/parse/dpd_test.go > TestDPDSignsParsePhase1` | ✅ COMPLIANT |
| + construction marker | Construction marker survives end-to-end parsing | `internal/parse/dpd_test.go > TestDPDSignsParsePhase1`; `internal/render/dpd_signs_integration_test.go > TestDPDSignsValidatedFixturesMatchGoldenOutputs/acertar` | ✅ COMPLIANT |
| + construction marker | Construction marker appears mid-phrase | `internal/render/dpd_signs_integration_test.go > TestDPDSignsValidatedFixturesMatchGoldenOutputs/acertar` | ✅ COMPLIANT |
| ⊗ exclusion | Exclusion sign regression is detected | `internal/render/dpd_signs_integration_test.go > TestDPDSignsValidatedFixturesDoNotRegressExclusionAndReferenceMarkers`; `go test ./...` | ✅ COMPLIANT |
| → reference | Cross-reference arrow regression is detected | `internal/render/dpd_signs_integration_test.go > TestDPDSignsValidatedFixturesDoNotRegressExclusionAndReferenceMarkers`; `go test ./...` | ✅ COMPLIANT |
| Brackets | Definition brackets are semantically tagged | `internal/render/dpd_signs_integration_test.go > TestDPDSignsBracketSemanticTaggingUsesRealFixtures/definition_brackets_from_abrogar` | ✅ COMPLIANT |
| Brackets | Pronunciation brackets are semantically tagged | `internal/render/dpd_signs_integration_test.go > TestDPDSignsBracketSemanticTaggingUsesRealFixtures/pronunciation_brackets_from_alícuota` | ✅ COMPLIANT |
| Brackets | Interpolation brackets are semantically tagged | `internal/render/dpd_signs_integration_test.go > TestDPDSignsBracketSemanticTaggingUsesRealFixtures/interpolation_brackets_from_androfobia` | ✅ COMPLIANT |
| Brackets | Bracket context ambiguity does not cause mis-classification | `internal/parse/dpd_test.go > TestDPDSignsBracketContextsStayBoundToImmediateParent`; `internal/render/dpd_signs_integration_test.go > TestDPDSignsBracketSemanticTaggingKeepsMixedContextsDistinctInOneArticle` | ✅ COMPLIANT |
| Slash | Slash appears as plain text | `internal/normalize/dpd_test.go > TestDPDSignsNormalizePhase1`; static evidence: no slash-specific `InlineKind` or whitelist entry added | ✅ COMPLIANT |
| * speculative | Agrammatical marker implementation is marked as speculative | Static code review: warning annotations in `internal/parse/dpd.go`, `internal/model/types.go`, `internal/normalize/dpd.go`, `internal/renderutil/inline.go`, and synthetic tests | ⚠️ PARTIAL |
| * speculative | Agrammatical marker compiles and passes synthetic tests | `internal/parse/dpd_test.go > TestDPDSignsParsePhase3Synthetic/agrammatical_marker_inferred_span_pattern`; `internal/normalize/dpd_test.go > TestDPDSignsNormalizePhase3Synthetic/agrammatical_marker_survives`; `internal/renderutil/inline_test.go > TestRenderInlineMarkdownSpeculativeSignsSynthetic/agrammatical_marker_renders_as_plain_text` | ✅ COMPLIANT |
| ‖ speculative | Hypothetical marker implementation is marked as speculative | Static code review: warning annotations in `internal/parse/dpd.go`, `internal/model/types.go`, `internal/normalize/dpd.go`, `internal/renderutil/inline.go`, and synthetic tests | ⚠️ PARTIAL |
| ‖ speculative | Hypothetical marker compiles and passes synthetic tests | `internal/parse/dpd_test.go > TestDPDSignsParsePhase3Synthetic/hypothetical_marker_inferred_span_pattern`; `internal/normalize/dpd_test.go > TestDPDSignsNormalizePhase3Synthetic/hypothetical_marker_survives`; `internal/renderutil/inline_test.go > TestRenderInlineMarkdownSpeculativeSignsSynthetic/hypothetical_marker_renders_as_plain_text` | ✅ COMPLIANT |
| // speculative | Phoneme marker implementation is marked as speculative | Static code review: warning annotations in `internal/parse/dpd.go`, `internal/model/types.go`, `internal/normalize/dpd.go`, `internal/renderutil/inline.go`, and synthetic tests | ⚠️ PARTIAL |
| // speculative | Phoneme marker compiles and passes synthetic tests | `internal/parse/dpd_test.go > TestDPDSignsParsePhase3Synthetic/phoneme_marker_inferred_plain_text_pattern`; `internal/normalize/dpd_test.go > TestDPDSignsNormalizePhase3Synthetic/phoneme_marker_survives`; `internal/renderutil/inline_test.go > TestRenderInlineMarkdownSpeculativeSignsSynthetic/phoneme_marker_renders_as_plain_text` | ✅ COMPLIANT |

**Compliance summary**: 14 / 17 compliant, 0 untested, 3 partial, 0 failing

**Archived signs**: `<` and `>` remain intentionally unimplemented. This matches the spec, design, tasks, and `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md`.

---

### Correctness (Static Structural Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| @ digital edition | ✅ Implemented | `<sup>` whitelist exists, `digitalEditionGlyph` and `InlineKindDigitalEdition` exist, and non-`@` superscripts now fall back to plain text. |
| + construction marker | ✅ Implemented | `constructionMarkerGlyph`, `InlineKindConstructionMarker`, `span.nc` handling, normalization, and rendering are present. |
| ⊗ exclusion regression | ✅ Implemented | Existing behavior preserved and regression tests pass. |
| → reference regression | ✅ Implemented | Existing anchor/reference behavior preserved and regression tests pass. |
| Bracket semantics | ✅ Implemented | Three bracket kinds exist and are wired through parse/normalize/render paths; ambiguity coverage now exists. |
| Slash plain text | ✅ Implemented | No slash-specific `InlineKind` exists; no slash whitelist entry was added; plain-text behavior remains intact. |
| *, ‖, // speculative signs | ✅ Implemented with warnings | Constants, `InlineKind`s, parse/normalize/render handling, and synthetic tests exist with warning annotations. |
| < and > archived | ✅ Intentional | `internal/parse/dpd.go`, `internal/model/types.go`, and `SIGN_ANALYSIS.md` document archival/non-implementation. |

---

### Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| Extend existing architecture without refactoring | ✅ Yes | Switch-based parse/render flow remains intact; changes are additive. |
| Implement validated + speculative split | ✅ Yes | Validated signs use real fixtures; speculative signs use synthetic tests with warning comments. |
| Archive `<` and `>` | ✅ Yes | Non-implementation is documented in code and analysis artifact. |
| Preserve 3 bracket semantic contexts | ✅ Yes | Three distinct `InlineKind` values are present and now covered by both isolated and mixed-context tests. |
| File change plan | ⚠️ Minor deviation | Design file listed core implementation/test files; implementation also uses `internal/render/dpd_signs_integration_test.go`, which is additive and beneficial. |

---

### Pass / Fail Findings Against Verify Goals

| Verify Goal | Finding |
|------------|---------|
| Re-check non-`@` `<sup>` fallback behavior | ✅ PASS — corrected parser behavior and dedicated passing test now prove fallback to plain text. |
| Re-check bracket ambiguity / multiple bracket contexts | ✅ PASS — dedicated parse and integration tests now prove immediate-parent classification across mixed contexts. |
| Re-validate implementation against spec/tasks/design | ✅ PASS WITH WARNINGS — tasks are complete, validated sign behavior is covered, design intent is followed, and only speculative comment-annotation checks remain partially manual. |
| Re-run tests/lint evidence as needed | ✅ PASS — full `go test ./...`, targeted DPD-sign tests, targeted render tests, and full lint all passed. |
| Update formal verification artifact | ✅ PASS — this report replaces the prior failing verification artifact. |

---

### Issues Found

**CRITICAL**

None.

**WARNING**

1. The speculative "warning annotation" scenarios are still verified primarily by static code review/comment inspection rather than a dedicated automated assertion.
2. `openspec/config.yaml` does not define a build command, so build/type-check evidence is intentionally absent beyond passing `go test` and lint.

**SUGGESTION**

1. If you want stricter enforcement, add a lightweight test or linter check that asserts speculative warning text remains present in the relevant files.

---

### Artifact References

- Spec: `openspec/changes/dpd-signs-preservation/spec.md`
- Design: `openspec/changes/dpd-signs-preservation/design.md`
- Tasks: `openspec/changes/dpd-signs-preservation/tasks.md`
- Verification report: `openspec/changes/dpd-signs-preservation/verify-report.md`

---

### Final Verdict

**PASS WITH WARNINGS**

The prior blockers are fixed: non-`@` superscripts no longer over-classify as digital-edition markers, and mixed bracket contexts are now behaviorally verified. The change is ready to proceed, with only non-blocking warnings around speculative comment-annotation enforcement and the intentionally unconfigured build step.

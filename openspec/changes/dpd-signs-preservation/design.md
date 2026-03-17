# Design: DPD Signs Preservation

## Technical Approach

This change extends the existing semantic preservation architecture to handle **9 active DPD signs** (6 validated, 3 speculative). The implementation follows the proven pattern established by the ⊗ (exclusion) sign: whitelist → constant → InlineKind → parse recognition → explicit preservation → render dispatch.

**No refactoring required**. The current architecture supports extension via:
- Adding entries to `semanticSpanAllowed` whitelist
- Adding constants for sign glyphs
- Extending the `InlineKind` enum
- Adding switch cases in `parseSupportedOpenTag()` and `renderInlineMarkdownItem()`
- Adding explicit preservation rules in `cleanInlineSegment()` and `cleanInlineText()`

**Two-tier implementation**:
1. **VALIDATED signs (6)**: Real HTML evidence confirms encoding. Implementation uses actual HTML patterns from DPD articles.
2. **SPECULATIVE signs (3)**: No HTML evidence found. Implementation uses inferred patterns with WARNING comments.

**2 ARCHIVED signs (`<`, `>`)**: Excluded due to HTML tag collision risk and low value-to-risk ratio. Documented in code comments for future reference.

---

## Architecture Decisions

### Decision: Extend Existing Architecture Without Refactoring

**Choice**: Add new whitelist entries, constants, InlineKind values, and switch cases to existing code without restructuring.

**Alternatives considered**:
- Refactor switch statements to strategy pattern
- Introduce sign registry with dynamic dispatch
- Replace InlineKind enum with polymorphic types

**Rationale**: 
- The ⊗ and → signs work correctly with current architecture
- Switch-based dispatch is legitimate for this domain (user confirmed)
- Refactoring introduces risk without benefit
- Extension via existing patterns maintains consistency

### Decision: Implement 4-Phase Rollout

**Choice**: Break implementation into 4 phases: Critical fixes (@ and +), bracket semantics, speculative signs, integration tests.

**Alternatives considered**:
- Implement all signs in one batch
- Implement only validated signs, defer speculative signs

**Rationale**:
- @ and + signs are CRITICAL (currently stripped) — must be fixed first
- Bracket semantics are HIGH priority (JSON needs semantic distinction)
- Speculative signs are MEDIUM priority (defensive implementation with warnings)
- Phase-based rollout enables partial rollback if issues arise

### Decision: Archive `<` and `>` Signs

**Choice**: Do NOT implement `<` and `>` signs. Document archival in code comments.

**Alternatives considered**:
- Implement using HTML entity detection (`&lt;`, `&gt;`)
- Implement with context-aware parsing to distinguish semantic vs. syntax

**Rationale**:
- HTML tag collision risk (`<span>`, `<div>`, `</a>`, etc.)
- NOT FOUND in any of 8 analyzed high-value DPD articles
- Low appearance ratio makes risk-to-benefit unfavorable
- Can be reconsidered if real examples found + safe parsing strategy developed

### Decision: Preserve 3 Bracket Semantic Contexts

**Choice**: Create 3 distinct InlineKind values for brackets: definition, pronunciation, interpolation.

**Alternatives considered**:
- Single `InlineKindBracket` with variant field
- Keep brackets as plain text only (no semantic preservation)

**Rationale**:
- Three distinct HTML containers validated: `<dfn>[...]</dfn>`, `<span class="nn">[...]</span>`, `<span class="yy">[...]</span>`
- JSON output requires semantic distinction (user confirmed)
- Markdown output keeps plain brackets (no visual change)
- Low collision risk (containers are distinct)

---

## Data Flow

The pipeline remains unchanged:

```
HTML Source
    ↓
preserveSemanticSpans()  ← Whitelist check: semanticSpanAllowed
    ↓
extractInlines()         ← Parse recognition: parseSupportedOpenTag()
    ↓
normalizeInlines()       ← Text cleaning: cleanInlineSegment(), cleanInlineText()
    ↓
renderInlineMarkdown()   ← Render dispatch: renderInlineMarkdownItem()
    ↓
JSON/Markdown Output
```

**Extension points (all existing)**:
1. `semanticSpanAllowed` map (lines 598-616) — add 4 new entries
2. Sign constants (near line 329) — add 9 new constants
3. `InlineKind` enum (types.go lines 19-35) — add 8 new constants
4. `parseSupportedOpenTag()` switch (dpd.go) — add cases for validated containers
5. `cleanInlineSegment()` function (lines 682-699) — add explicit preservation
6. `renderInlineMarkdownItem()` switch (inline.go lines 118-144) — add 8 new cases

---

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/parse/dpd.go` | Modify | Add 9 sign constants, extend whitelist with 4 entries, add parse cases, add preservation in cleanInlineSegment() |
| `internal/model/types.go` | Modify | Add 8 new InlineKind constants (5 validated, 3 speculative) |
| `internal/normalize/dpd.go` | Modify | Extend cleanInlineText() with explicit preservation rules for all signs |
| `internal/renderutil/inline.go` | Modify | Extend rendering switch with 8 new cases |
| `internal/parse/dpd_test.go` | Modify | Add parse tests for validated signs using real HTML fixtures |
| `internal/normalize/dpd_test.go` | Modify | Add normalization tests for all signs |
| `internal/parse/testdata/alícuota.html` | New | Real DPD article fixture (@ sign, ⊗ sign, pronunciation brackets) |
| `internal/parse/testdata/acertar.html` | New | Real DPD article fixture (+ sign, ⊗ sign, → sign, definition brackets) |
| `internal/parse/testdata/abrogar.html` | New | Real DPD article fixture (definition brackets) |

---

## Component Modifications

### 1. internal/parse/dpd.go

#### 1.1 Add Sign Constants (after line 329)

**Location**: Near `exclusionGlyph = "\u2297"` (line 329)

**Add these constants**:
```go
const (
	exclusionGlyph          = "\u2297" // ⊗ (existing)
	digitalEditionGlyph     = "@"
	constructionMarkerGlyph = "+"
	
	// WARNING: Speculative signs - no real HTML validation found in test articles
	// HTML patterns are inferred from validated sign patterns
	// Update when real DPD examples are discovered
	agrammaticalGlyph       = "*"
	hypotheticalGlyph       = "‖"     // \u2016
	phonemeGlyph            = "//"
)

// ARCHIVED SIGNS (not implemented):
// < (etymology) - HTML tag collision risk, low value-to-risk ratio
// > (transformation) - HTML tag collision risk, low value-to-risk ratio
// See testdata/dpd-signs-analysis/SIGN_ANALYSIS.md for rationale
```

#### 1.2 Extend Whitelist (lines 598-616)

**Location**: `semanticSpanAllowed` map initialization

**Current**:
```go
var semanticSpanAllowed = map[string]bool{
	"<dfn>":                       true,
	"</dfn>":                      true,
	"<em>":                        true,
	"</em>":                       true,
	"<i>":                         true,
	"</i>":                        true,
	"<span class=\"ejemplo\">":    true,
	"<span class=\"ment\">":       true,
	"<span class=\"bib\">":        true,
	"<span class=\"vers\">":       true,
	"<span class=\"yy\">":         true,  // ALREADY PRESENT
	"<span class=\"bolaspa\">":    true,
	"<span class=\"nn\">":         true,  // ALREADY PRESENT
	"<span class=\"nc\">":         true,  // ALREADY PRESENT
	"<span class=\"pattern\">":    true,
	"<span class=\"correction\">": true,
	"</span>":                     true,
}
```

**Add**:
```go
"<sup>":  true,  // For @ sign (digital edition marker)
"</sup>": true,
```

**Note**: `<span class="nc">`, `<span class="nn">`, `<span class="yy">` are ALREADY whitelisted. No changes needed.

#### 1.3 Extend parseSupportedOpenTag() Switch

**Location**: `parseSupportedOpenTag()` function (find it by searching for "func parseSupportedOpenTag")

**Current pattern** (existing, for reference):
```go
case "span":
	if strings.Contains(raw, `class="bolaspa"`) {
		return model.InlineKindExclusion
	}
	// ... other span cases
```

**Add these cases**:
```go
case "sup":
	// Phase 1: Digital edition marker (@)
	if strings.Contains(p.currentText(), digitalEditionGlyph) {
		return model.InlineKindDigitalEdition
	}
	return model.InlineKindText

case "span":
	// Existing bolaspa check
	if strings.Contains(raw, `class="bolaspa"`) {
		return model.InlineKindExclusion
	}
	
	// Phase 1: Construction marker (+)
	if strings.Contains(raw, `class="nc"`) {
		return model.InlineKindConstructionMarker
	}
	
	// Phase 2: Bracket semantic contexts
	if strings.Contains(raw, `class="nn"`) {
		// Pronunciation brackets
		return model.InlineKindBracketPronunciation
	}
	if strings.Contains(raw, `class="yy"`) {
		// Interpolation/example brackets
		return model.InlineKindBracketInterpolation
	}
	
	// Phase 3: Speculative signs (WARNING: no HTML validation)
	// WARNING: Inferred patterns - update when real examples found
	if strings.Contains(p.currentText(), agrammaticalGlyph) {
		return model.InlineKindAgrammatical
	}
	if strings.Contains(p.currentText(), hypotheticalGlyph) {
		return model.InlineKindHypothetical
	}
	if strings.Contains(p.currentText(), phonemeGlyph) {
		return model.InlineKindPhoneme
	}
	
	// ... existing span cases continue

case "dfn":
	// Phase 2: Definition/correction brackets
	return model.InlineKindBracketDefinition
```

#### 1.4 Extend cleanInlineSegment() (lines 682-699)

**Location**: `cleanInlineSegment()` function

**Current** (line 685):
```go
text = strings.ReplaceAll(text, exclusionGlyph, exclusionGlyph) //nolint:gocritic // dupArg: intentional normalization of ⊗ variants
```

**Add after line 685**:
```go
// Phase 1: Preserve validated signs
text = strings.ReplaceAll(text, digitalEditionGlyph, digitalEditionGlyph)       //nolint:gocritic // Preserve @ sign
text = strings.ReplaceAll(text, constructionMarkerGlyph, constructionMarkerGlyph) //nolint:gocritic // Preserve + sign

// Phase 3: Preserve speculative signs (WARNING: no HTML validation)
text = strings.ReplaceAll(text, agrammaticalGlyph, agrammaticalGlyph)           //nolint:gocritic // Preserve * sign (SPECULATIVE)
text = strings.ReplaceAll(text, hypotheticalGlyph, hypotheticalGlyph)           //nolint:gocritic // Preserve ‖ sign (SPECULATIVE)
text = strings.ReplaceAll(text, phonemeGlyph, phonemeGlyph)                     //nolint:gocritic // Preserve // sign (SPECULATIVE)

// Brackets are preserved as-is (plain text), no explicit handling needed
```

**Note**: The `dupArg` is INTENTIONAL — it's a normalization pass that ensures consistent representation. This is the pattern we replicate for all signs.

---

### 2. internal/model/types.go

#### 2.1 Extend InlineKind Constants (lines 19-35)

**Location**: After `InlineKindEmphasis = "emphasis"` (line 35)

**Add**:
```go
	// DPD Signs - VALIDATED (real HTML evidence)
	InlineKindDigitalEdition        = "digital_edition"      // @ sign in <sup>@</sup>
	InlineKindConstructionMarker    = "construction_marker"  // + sign in <span class="nc">
	InlineKindBracketDefinition     = "bracket_definition"   // [...] in <dfn>
	InlineKindBracketPronunciation  = "bracket_pronunciation" // [...] in <span class="nn">
	InlineKindBracketInterpolation  = "bracket_interpolation" // [...] in <span class="yy">

	// DPD Signs - SPECULATIVE (inferred patterns, no HTML validation)
	// WARNING: These patterns are best-guess based on validated signs
	// Update when real DPD examples are discovered
	InlineKindAgrammatical          = "agrammatical"         // * sign (SPECULATIVE)
	InlineKindHypothetical          = "hypothetical"         // ‖ sign (SPECULATIVE)
	InlineKindPhoneme               = "phoneme"              // // sign (SPECULATIVE)

	// ARCHIVED SIGNS (not implemented):
	// < (etymology) - HTML tag collision risk
	// > (transformation) - HTML tag collision risk
	// See testdata/dpd-signs-analysis/SIGN_ANALYSIS.md for archival rationale
```

---

### 3. internal/normalize/dpd.go

#### 3.1 Extend cleanInlineText() (lines 542-547)

**Location**: `cleanInlineText()` function

**Current**:
```go
func cleanInlineText(raw string) string {
	text := html.UnescapeString(raw)
	text = strings.ReplaceAll(text, "\u200d", "")
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}
```

**Replace with** (add explicit preservation):
```go
func cleanInlineText(raw string) string {
	text := html.UnescapeString(raw)
	text = strings.ReplaceAll(text, "\u200d", "")
	
	// Preserve DPD signs explicitly before field normalization
	// NOTE: These are defined in internal/parse/dpd.go
	preservedSigns := []string{
		// Validated signs
		"\u2297", // ⊗ exclusion (existing)
		"@",      // digital edition
		"+",      // construction marker
		
		// Speculative signs (WARNING: no HTML validation)
		"*",      // agrammatical (SPECULATIVE)
		"‖",      // hypothetical (SPECULATIVE)
		"//",     // phoneme (SPECULATIVE)
	}
	
	// Replace signs with placeholders, normalize whitespace, restore signs
	placeholders := make([]string, len(preservedSigns))
	for i, sign := range preservedSigns {
		placeholder := fmt.Sprintf("\x00SIGN%d\x00", i)
		placeholders[i] = placeholder
		text = strings.ReplaceAll(text, sign, placeholder)
	}
	
	text = strings.Join(strings.Fields(text), " ")
	
	for i, placeholder := range placeholders {
		text = strings.ReplaceAll(text, placeholder, preservedSigns[i])
	}
	
	return strings.TrimSpace(text)
}
```

**Rationale**: `strings.Fields()` splits on whitespace and collapses adjacent whitespace. Signs like `+` could be separated from adjacent text (`+ infinitivo` → `+infinitivo`). By using placeholder substitution, we ensure signs survive field normalization.

---

### 4. internal/renderutil/inline.go

#### 4.1 Extend renderInlineMarkdownItem() Switch (lines 118-144)

**Location**: `renderInlineMarkdownItem()` function

**Current** (line 124):
```go
switch inline.Kind {
case model.InlineKindExample:
	return "‹" + text + "›"
case model.InlineKindMention, model.InlineKindEmphasis, model.InlineKindWorkTitle, model.InlineKindCorrection:
	if len(inline.Children) > 0 {
		return renderStyledInlineMarkdown(inline.Children, "*")
	}
	return "*" + text + "*"
case model.InlineKindReference:
	if text == "" {
		return ""
	}
	return "→ [" + text + "](" + inline.Target + ")"
case model.InlineKindScaffold:
	return text
case model.InlineKindCitationQuote:
	return "«" + text + "»"
default:
	return text
}
```

**Add before `default:` case**:
```go
// Phase 1: VALIDATED signs with real HTML evidence
case model.InlineKindDigitalEdition:
	return text // @ sign preserved as-is

case model.InlineKindConstructionMarker:
	return text // + sign with construction phrase preserved as-is

// Phase 2: Bracket semantic contexts (VALIDATED)
case model.InlineKindBracketDefinition,
     model.InlineKindBracketPronunciation,
     model.InlineKindBracketInterpolation:
	// Markdown output: brackets as plain text
	// JSON output: semantic distinction preserved via InlineKind
	return text

// Phase 3: SPECULATIVE signs (WARNING: no HTML validation)
case model.InlineKindAgrammatical:
	return text // * sign (SPECULATIVE - no real HTML found)

case model.InlineKindHypothetical:
	return text // ‖ sign (SPECULATIVE - no real HTML found)

case model.InlineKindPhoneme:
	return text // // sign (SPECULATIVE - no real HTML found)
```

**Note**: All signs emit as plain text in Markdown. The semantic distinction is preserved in the InlineKind for JSON output.

---

## Interfaces / Contracts

No new interfaces. Extension of existing contracts:

### InlineKind Contract (Extended)

**Purpose**: Distinguish semantic roles of inline content for JSON output.

**New Values**:
```go
// VALIDATED (real HTML evidence)
InlineKindDigitalEdition        = "digital_edition"
InlineKindConstructionMarker    = "construction_marker"
InlineKindBracketDefinition     = "bracket_definition"
InlineKindBracketPronunciation  = "bracket_pronunciation"
InlineKindBracketInterpolation  = "bracket_interpolation"

// SPECULATIVE (inferred patterns)
InlineKindAgrammatical          = "agrammatical"
InlineKindHypothetical          = "hypothetical"
InlineKindPhoneme               = "phoneme"
```

**Backward Compatibility**: Existing code using `InlineKind` continues to work. New kinds are additive only.

---

## Testing Strategy

### Phase 1: Critical Signs (@ and +)

| Test Layer | What to Test | Approach |
|-----------|-------------|----------|
| Parse | `<sup>@</sup>` → `InlineKindDigitalEdition` | Unit test with real HTML from alícuota.html |
| Parse | `<span class="nc">+ infinitivo</span>` → `InlineKindConstructionMarker` | Unit test with real HTML from acertar.html |
| Normalize | @ sign survives cleanInlineText() | Unit test with synthetic input |
| Normalize | + phrase survives cleanInlineText() | Unit test with synthetic input |
| Render | InlineKindDigitalEdition → "@" | Unit test |
| Render | InlineKindConstructionMarker → "+ infinitivo" | Unit test |
| Integration | Real alícuota.html fixture → @ in final Markdown | End-to-end test |
| Integration | Real acertar.html fixture → + infinitivo in final Markdown | End-to-end test |

### Phase 2: Bracket Semantics

| Test Layer | What to Test | Approach |
|-----------|-------------|----------|
| Parse | `<dfn>[una ley]</dfn>` → `InlineKindBracketDefinition` | Unit test with real HTML from abrogar.html |
| Parse | `<span class="nn">[alikuóto]</span>` → `InlineKindBracketPronunciation` | Unit test with real HTML from alícuota.html |
| Parse | `<span class="yy">[las feministas]</span>` → `InlineKindBracketInterpolation` | Unit test with real HTML from androfobia.html |
| Render | All bracket kinds → plain text brackets in Markdown | Unit test |
| Integration | Real fixtures → correct bracket InlineKind in JSON output | End-to-end test |

### Phase 3: Speculative Signs

| Test Layer | What to Test | Approach |
|-----------|-------------|----------|
| Parse | Inferred `* ` pattern → `InlineKindAgrammatical` | Unit test with SYNTHETIC HTML (clearly marked) |
| Parse | Inferred `‖` pattern → `InlineKindHypothetical` | Unit test with SYNTHETIC HTML (clearly marked) |
| Parse | Inferred `//` pattern → `InlineKindPhoneme` | Unit test with SYNTHETIC HTML (clearly marked) |
| Render | All speculative kinds → plain text in Markdown | Unit test |
| WARNING | Test comments indicate SPECULATIVE / SYNTHETIC nature | Code review |

### Phase 4: Integration & Regression

| Test Layer | What to Test | Approach |
|-----------|-------------|----------|
| Regression | ⊗ sign continues working | Existing tests must pass |
| Regression | → sign continues working | Existing tests must pass |
| Integration | Real alícuota.html → @ + ⊗ + pronunciation brackets | End-to-end golden test |
| Integration | Real acertar.html → + + ⊗ + → + definition brackets | End-to-end golden test |
| Integration | Real abrogar.html → definition brackets | End-to-end golden test |

### Test Fixture Promotion

**Source**: `scripts/testdata/dpd-signs-analysis/*.html`
**Target**: `internal/parse/testdata/*.html`

**Files to promote**:
- `alícuota.html` (contains @, ⊗, pronunciation brackets)
- `acertar.html` (contains +, ⊗, →, definition brackets)
- `abrogar.html` (contains definition brackets)

**Promotion steps**:
1. Copy files from scripts/testdata/ to internal/parse/testdata/
2. Create golden output files (.json, .md) for end-to-end tests
3. Add integration tests that parse fixtures and compare against golden outputs

---

## Migration / Rollout

No data migration required. This is a correctness fix extending the parser/renderer.

**Rollout Plan**:
1. Merge Phase 1 (critical @ and + fixes) first — these are STRIPPED currently
2. Merge Phase 2 (bracket semantics) after Phase 1 validation
3. Merge Phase 3 (speculative signs) with WARNING annotations
4. Merge Phase 4 (integration tests) last

**Rollback Strategy**:
- Phase-based commits enable selective revert
- No breaking API changes (enum extension is additive)
- Existing ⊗ and → tests must pass (regression guard)

**Feature Flags**: Not needed. This is internal correctness, not user-facing behavior toggle.

---

## Open Questions

None. All technical decisions resolved:

- ✅ Whitelist entries validated with real HTML
- ✅ InlineKind enum values confirmed
- ✅ Rendering approach confirmed (plain text for Markdown, semantic distinction for JSON)
- ✅ Bracket contexts confirmed (3 distinct kinds)
- ✅ Slash handling confirmed (plain text, no special semantic)
- ✅ Archived signs decision confirmed (`<` and `>` excluded due to collision risk)

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| **Speculative signs fail in production** | Medium | Medium | WARNING comments in code; implement defensively; update when real examples found |
| **Archived signs requested later** | Low | Low | Signs documented in SIGN_ANALYSIS.md; can be implemented if safe parsing strategy developed |
| **Sign collision (multiple signs share container)** | Low | Medium | Phase-based rollout; test each sign independently; validated containers are distinct |
| **Bracket context mis-classification** | Low | Low | Three distinct HTML containers validated; low collision risk |
| **Whitelist affects other HTML parsing** | Very Low | Low | Whitelist is strict; new entries only affect previously-stripped content |
| **Performance impact (additional switch cases)** | Very Low | Very Low | Switch dispatch is O(1); negligible impact on small inline items |
| **⊗ or → regression** | Very Low | High | Existing tests must pass; phase-based rollout enables fast detection |

---

## Implementation Order

### Phase 1: Critical Signs (@ and +) — HIGHEST PRIORITY

**Rationale**: These signs are currently STRIPPED. Fixing them is urgent.

1. Add constants: `digitalEditionGlyph`, `constructionMarkerGlyph`
2. Add InlineKind: `InlineKindDigitalEdition`, `InlineKindConstructionMarker`
3. Add whitelist: `<sup>`, `</sup>` (nc, nn, yy already present)
4. Extend parseSupportedOpenTag(): cases for `<sup>` and `<span class="nc">`
5. Extend cleanInlineSegment(): preserve @ and +
6. Extend renderInlineMarkdownItem(): cases for digital edition and construction marker
7. Add tests with real fixtures (alícuota.html, acertar.html)
8. Verify end-to-end: @ and + survive in final output

### Phase 2: Bracket Semantics — HIGH PRIORITY

**Rationale**: Brackets work as text but lose semantic distinction in JSON.

1. Add InlineKind: `InlineKindBracketDefinition`, `InlineKindBracketPronunciation`, `InlineKindBracketInterpolation`
2. Extend parseSupportedOpenTag(): cases for `<dfn>`, `<span class="nn">`, `<span class="yy">`
3. Extend renderInlineMarkdownItem(): case for all 3 bracket kinds (plain text output)
4. Add tests with real fixtures (abrogar.html, acertar.html, alícuota.html)
5. Verify JSON output: distinct InlineKind values for each bracket context

### Phase 3: Speculative Signs (*, ‖, //) — MEDIUM PRIORITY

**Rationale**: No HTML validation. Implement defensively with warnings.

1. Add constants with WARNING comments: `agrammaticalGlyph`, `hypotheticalGlyph`, `phonemeGlyph`
2. Add InlineKind with WARNING comments: `InlineKindAgrammatical`, `InlineKindHypothetical`, `InlineKindPhoneme`
3. Extend parseSupportedOpenTag(): best-guess cases with WARNING comments
4. Extend cleanInlineSegment(): preserve *, ‖, //
5. Extend renderInlineMarkdownItem(): cases for speculative signs
6. Add tests with SYNTHETIC HTML (clearly marked as synthetic)
7. Document in SIGN_ANALYSIS.md: pending validation

### Phase 4: Integration & Golden Tests — FINAL VALIDATION

**Rationale**: Comprehensive end-to-end validation with real DPD articles.

1. Promote test fixtures: alícuota.html, acertar.html, abrogar.html to internal/parse/testdata/
2. Create golden output files (.json, .md) for each fixture
3. Add integration tests comparing parsed output to golden files
4. Run full test suite, verify no regressions
5. Update documentation: SIGN_ANALYSIS.md, CONTRIBUTING.md if needed

---

## Next Step

Ready for **sdd-tasks** phase to break down this design into granular implementation tasks.

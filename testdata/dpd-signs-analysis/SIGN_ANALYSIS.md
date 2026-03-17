# DPD Sign Analysis - HTML Encoding Report

**Analysis Date:** 2026-03-17  
**Source:** Real DPD articles fetched from https://www.rae.es/dpd  
**Method:** Extracted `<entry>` content only (article body, excluding site chrome)

**Updated:** 2026-03-17 - Signs `<` and `>` archived due to HTML tag collision risk; validated sign coverage promoted into integration/golden tests

## Executive Summary

✅ **CONFIRMED**: 6 signs validated with real HTML evidence, 3 signs speculative (pending validation)  
✅ **IMPLEMENTED**: Validated `@`, `+`, bracket semantics, `⊗`, and `→` now have real-fixture coverage in integration/golden tests.  
📦 **ARCHIVED**: Signs `<` and `>` excluded from implementation due to HTML tag collision risk and low value-to-risk ratio.

---

## Implementation Status

### Active Signs (9 total)

**VALIDATED (6)** - Confirmed with real DPD articles:
- ✅ `⊗` (exclusion) - WORKING + regression-covered
- ✅ `→` (cross-reference) - WORKING + regression-covered  
- ✅ `@` (digital edition) - IMPLEMENTED + real fixture coverage
- ✅ `+` (construction marker) - IMPLEMENTED + real fixture coverage
- ✅ `[ ]` (brackets) - IMPLEMENTED with 3 semantic contexts and real fixture coverage
- ✅ `/` (slash) - preserved as plain text (no semantic InlineKind by design)

**SPECULATIVE (3)** - Not found in test articles, inferred patterns:
- ⚠️ `*` (agrammatical) - IMPLEMENTED SPECULATIVELY (synthetic coverage only)
- ⚠️ `‖` (hypothetical) - IMPLEMENTED SPECULATIVELY (synthetic coverage only)
- ⚠️ `//` (phoneme) - IMPLEMENTED SPECULATIVELY (synthetic coverage only)

### Archived Signs (2 total)

**📦 EXCLUDED from implementation** due to HTML tag collision risk:

| Sign | Description | Archival Reason |
|------|-------------|-----------------|
| `<` | Etymology ("comes from") | **HTML tag collision risk**: Conflicts with opening tags like `<span>`, `<div>`, `<a>`. Low appearance ratio. Not found in high-value articles. |
| `>` | Transformation ("passes to") | **HTML tag collision risk**: Conflicts with closing tag markers. Low appearance ratio. Not found in high-value articles. |

**Future Reconsideration Criteria**:
1. Real DPD articles containing these signs in high-value contexts are discovered
2. A collision-safe parsing strategy is developed (e.g., context-aware detection)
3. User demand justifies implementation complexity

---

## Sign-by-Tag Mapping Table (Active Signs Only)

| Sign | Description | HTML Encoding | Tag Container | Whitelist Status | Example Word | Status |
|------|-------------|---------------|---------------|------------------|--------------|--------|
| `⊗` | Incorrect/inadvisable forms | `&#x2297;&#x200D;` or `⊗&#x200D;` | `<span class="bolaspa">` | ✅ Whitelisted | acertar, alícuota | **WORKING** |
| `@` | Digital edition marker | `@` (plain char) | `<sup>` | ❌ **NOT** whitelisted | alícuota, androfobia | **STRIPPED** |
| `+` | Plus sign (construction marker) | `+` (plain char) | `<span class="nc">` | ❌ **NOT** whitelisted | acertar | **STRIPPED** |
| `→` | Cross-reference (see) | `→` (in anchor text) | `<a href="/dpd/...">` | ✅ Whitelisted | acertar | **WORKING** |
| `[` | Left bracket (pronunciation/corrections) | `[` (plain char) | `<dfn>` or `<span class="nn">` or `<span class="yy">` | Partial | abrogar, acertar, alícuota, androfobia | **TEXT ONLY** |
| `]` | Right bracket | `]` (plain char) | `<dfn>` or `<span class="nn">` or `<span class="yy">` | Partial | abrogar, acertar, alícuota, androfobia | **TEXT ONLY** |
| `/` | Slash (alternatives/end of line) | `/` (plain char) | Plain text in article | ⚠️ Ambiguous | All articles (in text) | **NO SEMANTIC** |
| `*` | Asterisk (agrammatical) | NOT FOUND (speculative) | ? | Unknown | ? | **NOT FOUND** |
| `‖` | Double bar (hypothetical) | NOT FOUND (speculative) | ? | Unknown | ? | **NOT FOUND** |
| `//` | Double slash (phonemes) | NOT FOUND (speculative) | ? | Unknown | ? | **NOT FOUND** |

## Archived Signs (Not Implemented)

| Sign | Description | Why Archived | HTML Observations |
|------|-------------|--------------|-------------------|
| `<` | Etymology ("comes from") | HTML tag collision risk | Found only in HTML tag syntax (`<span>`, `<entry>`), not as semantic content |
| `>` | Transformation ("passes to") | HTML tag collision risk | Found only in HTML tag syntax (`</header>`, closing `>`), not as semantic content |

---

## Detailed Findings by Sign

### 1. ⊗ (Exclusion/Incorrect Forms) - ✅ WORKING

**HTML Encoding:**
```html
<span class="bolaspa">⊗&#x200D;</span>
```

**Example (alícuota, line 545):**
```html
Es palabra esdrújula, por lo que son incorrectas las grafías sin tilde 
<em><span class="bolaspa">⊗&#x200D;</span><span class="ment">alicuoto</span></em>
```

**Status:** ✅ **FULLY WORKING**  
**Reason:** 
- Tag `<span class="bolaspa">` is whitelisted in `semanticSpanAllowed` (line 602)
- Has constant `exclusionGlyph = "\u2297"` (line 329)
- Has InlineKind: `InlineKindExclusion`
- Complete test coverage

**Action Required:** NONE (reference pattern for other signs)

---

### 2. @ (Digital Edition Marker) - ✅ IMPLEMENTED

**HTML Encoding:**
```html
<sup>@</sup>
```

**Example (alícuota, line 545):**
```html
<span class="bib">(<i>Día</i><sup>@</sup> <span class="cbil" title="España">es</span> 26.10.2014)</span>
```

**Status:** ✅ **IMPLEMENTED AND COVERED**  
**Implemented:**
- `<sup>` whitelisted in `semanticSpanAllowed`
- `digitalEditionGlyph = "@"`
- `InlineKindDigitalEdition`
- Real-fixture coverage with `alícuota.html` and `androfobia.html`

**Validation:**
- Parse/normalize/render pipeline preserves `@`
- JSON output keeps `InlineKindDigitalEdition`
- Markdown/golden coverage verifies plain `@` output

---

### 3. + (Plus Sign / Construction Marker) - ✅ IMPLEMENTED

**HTML Encoding:**
```html
<span class="nc">+ infinitivo</span>
```

**Example (acertar, line 546):**
```html
<span class="embf">acertar a <span class="nc">+ infinitivo</span>.</span>
```

**Status:** ✅ **IMPLEMENTED AND COVERED**  
**Implemented:**
- `<span class="nc">` preserved and tagged as `InlineKindConstructionMarker`
- `constructionMarkerGlyph = "+"`
- Real-fixture coverage with `acertar.html`

**Validation:**
- Parse/normalize/render pipeline preserves `+ infinitivo`
- JSON output keeps `InlineKindConstructionMarker`
- Markdown/golden coverage verifies plain `+ infinitivo` output

---

### 4. → (Cross-Reference Arrow) - ✅ WORKING

**HTML Encoding:**
```html
<a href="/dpd/acertar">→ acertar</a>
```

**Example (acertar, line 543):**
```html
Verbo irregular (&#x2192; <a href="/dpd/ayuda/modelos-de-conjugacion-verbal#acertar">acertar</a>)
```

**Status:** ✅ **WORKING**  
**Reason:** 
- Anchor tags `<a>` are already whitelisted and handled
- Arrow appears as plain text or HTML entity in anchor content

**Action Required:** NONE (already working)

---

### 5. [ and ] (Brackets) - ✅ IMPLEMENTED WITH SEMANTIC CONTEXTS

**HTML Encoding:** Multiple contexts:

**A. In definitions (`<dfn>`):**
```html
<dfn>'Derogar o abolir [una ley]'</dfn>
```

**B. In pronunciation (`<span class="nn">`):**
```html
<span class="nn">[alikuóto]</span>
```

**C. In examples/interpolations (`<span class="yy">`):**
```html
<span class="yy">[las feministas]</span>
```

**Status:** ✅ **IMPLEMENTED WITH JSON-LEVEL SEMANTIC DISTINCTION**  
**Implemented contexts:**
- `InlineKindBracketDefinition` for `<dfn>[...]</dfn>`
- `InlineKindBracketPronunciation` for `<span class="nn">[...]</span>`
- `InlineKindBracketInterpolation` for `<span class="yy">[...]</span>`

**Validation:**
- Real-fixture integration coverage for `abrogar.html`, `alícuota.html`, and `androfobia.html`
- JSON output distinguishes contexts by InlineKind
- Markdown output keeps plain bracket text without semantic wrappers

---

### 6. / (Slash) - ⚠️ NO SEMANTIC DISTINCTION

**HTML Encoding:** Plain text character in article body

**Status:** ⚠️ **NO SEMANTIC ENCODING**  
**Reason:** 
- Appears as plain `/` in text, no special tag
- Indistinguishable from HTML closing tags like `</header>`
- No way to detect if `/` is semantic (alternatives) vs. structural

**Action Required:**
1. Verify with user: Does `/` need special preservation?
2. If yes: requires DPD website changes or heuristic detection
3. If no: document that `/` is preserved as plain text

---

### 7-11. Missing Signs (NOT FOUND IN TEST WORDS)

The following signs were **NOT FOUND** in any of the 8 test articles:

| Sign | Description | Test Words Tried |
|------|-------------|------------------|
| `*` | Agrammatical constructions | abrogar, abstraer, abuelo, acertar, adherir, alférez, alícuota, androfobia |
| `‖` | Hypothetical/reconstructed | abrogar, abstraer, abuelo, acertar, adherir, alférez, alícuota, androfobia |
| `>` | Transformation ("passes to") | abrogar, abstraer, abuelo, acertar, adherir, alférez, alícuota, androfobia |
| `<` | Etymology ("comes from") | abrogar, abstraer, abuelo, acertar, adherir, alférez, alícuota, androfobia |
| `//` | Phonemes | abrogar, abstraer, abuelo, acertar, adherir, alférez, alícuota, androfobia |

**Current implementation status:**
1. `*`, `‖`, and `//` are implemented speculatively with WARNING comments and synthetic-only tests
2. `<` and `>` remain archived and intentionally unimplemented
3. Real DPD examples are still required before speculative handling can be promoted to validated coverage

---

## Implementation Priority

### CRITICAL (Must Keep Verified)

1. **`@` sign (digital edition)** - keep real-fixture coverage passing
2. **`+` sign (construction marker)** - keep real-fixture coverage passing

### HIGH (Semantic Loss)

3. **Brackets `[ ]`** - keep semantic distinction in JSON while Markdown stays plain text
4. **Missing real examples (`*`, `‖`, `//`)** - get DPD words, replace speculative handling with validated behavior if found

### MEDIUM (Already Working or Low Impact)

5. **`/` slash** - Verify if semantic preservation needed
6. **`⊗` exclusion** - Already working, use as reference
7. **`→` cross-ref** - Already working

---

## Test Fixture Recommendations

The following HTML files should be promoted to `testdata/dpd/` for test fixtures:

1. **alícuota.html** - Contains `@`, `⊗`, `[`, `]` signs
2. **acertar.html** - Contains `+`, `⊗`, `→`, `[`, `]` signs
3. **abrogar.html** - Contains `[`, `]` in definitions

---

## Parser Flow Implications

### Current Flow (for ⊗ sign):
```
HTML → preserveSemanticSpans() → <span class="bolaspa"> detected
    → extractInlines() → creates InlineKindExclusion node
    → normalize → cleanText() preserves ⊗ character
    → renderInlineMarkdown() → switch case outputs ⊗ in Markdown
```

### Broken Flow (for @ and + signs):
```
HTML → preserveSemanticSpans() → <sup> or <span class="nc"> NOT in whitelist
    → Tag stripped, content lost
    → NEVER reaches extractInlines()
    → NEVER reaches rendering
```

**Fix:** Add missing tags to `semanticSpanAllowed` whitelist.

---

## Next Steps

1. ✅ **Document findings** (this file)
2. **Get user clarification:**
   - Provide test words for missing signs (`*`, `‖`, `<`, `>`, `//`)
   - Confirm if brackets and slashes need semantic preservation
3. **Update proposal** with real HTML evidence
4. **Proceed to sdd-spec phase** with validated sign-to-tag mapping
5. **Implement fixes** following ⊗ sign pattern for each new sign

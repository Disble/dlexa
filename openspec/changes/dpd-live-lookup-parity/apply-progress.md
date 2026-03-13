# Apply Progress: DPD Live Lookup Parity

## Batch Completed

- Completed **Phase 5.6 terminal-output contract batch** on top of the earlier fidelity-verification work.

## Access Method Learning (explicit)

- Validated from this environment: raw DPD article HTML is reachable with direct `GET https://www.rae.es/dpd/<term>` when the request uses a browser-like profile, especially a browser-like `User-Agent`.
- Also validated: lower-profile or no-`User-Agent` request shapes can still hit Cloudflare `403` challenge pages.
- `/srv/keys` is not a useful access path for this project here.
- `go-rae` is NOT a direct DPD access blueprint for this change because it calls a third-party API instead of parsing the raw DPD HTML page.

## Work Performed

- Replaced the stub DPD parser contract with source-shaped parse results (`parse.Result`, parsed article/section/citation types).
- Added a captured `testdata/dpd/bien.html` fixture and implemented `internal/parse/dpd.go` to:
  - isolate the canonical `<entry class="lex">` article region
  - preserve dictionary label, edition, lemma, numbered sections, nested `a)`/`b)`/`c)` items, readable references, and citation text
  - reject Cloudflare challenge bodies explicitly
  - distinguish no-article from extraction failure
- Updated DPD fetch defaults to a browser-like `User-Agent` and added browser-profile request headers plus challenge-body detection in `internal/fetch/http.go`.
- Introduced the minimal normalized article model in `internal/model/types.go` and wired `internal/normalize/dpd.go` to produce Markdown-ready article sections.
- Switched the production DPD pipeline in `internal/app/wiring.go` to `DPDFetcher -> DPDArticleParser -> DPDNormalizer`.
- Updated pipeline adapters/tests so `fetch -> parse -> normalize` passes rich parse results instead of flattening immediately.
- Updated Markdown rendering to prefer `Entry.Article` and added a deterministic `bien.md.golden` parity baseline.
- Added integration coverage that runs parse + normalize + render over the real captured `bien` fixture.
- Tightened parser preservation so article extraction keeps authored semantic spans (`<dfn>`, examples, emphasis, references) long enough for downstream fidelity work and promotes lexical-head-only paragraphs into `ParsedSection.Title` instead of flattening them into body text.
- Added many granular parser tests for quote preservation, example separation, non-duplicated numeric references, lexical-head attachment, and citation/body separation.
- Added targeted normalizer tests and logic for Markdown-safe transformation of preserved semantic spans: no synthetic quote wrapping, emphasized examples, canonical `-> [n]`-style reference shaping with intact parentheses, lexical title retention, and structured citation-field extraction.
- Added targeted renderer tests for preserved quotes, readable emphasis/examples, single-arrow references, integrated numbered lexical heads, and explicit citation readability for terminal/LLM consumption.
- Refreshed `testdata/dpd/bien.md.golden` so the integration baseline matches the revised fidelity contract instead of preserving stale defective expectations.
- Reworked final terminal rendering so stdout now emits plain terminal-readable output instead of raw Markdown source markers: emphasis markers are stripped into plain text, Markdown links collapse to readable `-> n` references, and sections/subitems render as coherent single-line list items for LLM consumption.
- Added renderer and end-to-end regression tests that explicitly fail on the user-rejected runtime defects: raw `*...*`, detached numbering/title fragments, malformed link syntax in stdout, and fragmented subitem layout.
- Updated the deterministic `bien` golden again so the broad integration assertion now matches the actual terminal-output contract instead of Markdown-source formatting.
- Added a priority semantic-example hotfix so real DPD examples from `span.ejemplo` are preserved in terminal output with an explicit terminal-safe marker (`[ej.: ...]`) instead of collapsing into ordinary prose.
- Added specific regression tests around the real `bien` examples, including `No he dormido bien esta noche`, so example semantics remain recoverable in both normalization and final terminal rendering.
- Began the scalable semantic-structure refactor: parser paragraphs now carry typed inline nodes, normalization preserves those nodes instead of treating `Paragraph.Markdown` as sole truth, and the renderer now prefers inline semantics when available.
- Added deterministic semantic extraction coverage for marker families observed across `bien`, `ver`, and `dar`, including examples, mentions, glosses, lexical headings, citation quotes, bibliography blocks, work titles, small-caps markers, editorial glosses, scaffold markers, corrections, patterns, exclusion markers, and typed references.
- Replaced the previously invented example label with source-faithful quotation-style rendering (`‹…›`) so examples remain distinguishable without editorial invention.

## Verification

- `go test ./internal/render -run "TestMarkdownRenderer|TestDPDParseNormalizeRenderProducesTerminalReadableOutput|TestDPDParseNormalizeRenderMatchesBienGolden"`
- `go test ./internal/app ./internal/render`
- `go test ./internal/normalize ./internal/render -run "TestDPDNormalizerPreservesExampleAndEmphasisSemantics|TestDPDNormalizerMarksRealBienExampleSemanticsExplicitly|TestMarkdownRendererRendersReadableExampleAndEmphasisOutput|TestMarkdownRendererKeepsRealBienExampleRecoverableInTerminalOutput|TestDPDParseNormalizeRenderProducesTerminalReadableOutput|TestDPDParseNormalizeRenderMatchesBienGolden"`
- `go test ./internal/render`
- `go test ./internal/parse ./internal/normalize ./internal/render`

## Remaining Focus

- Phase 3.5 / 5.7: JSON secondary output still needs explicit article-first serialization coverage plus a deterministic golden.
- Phase 5.8: optional live drift probe remains pending.
- Normalization still stores paragraph content as Markdown-ish strings internally; runtime correctness is now enforced at the renderer boundary, but future cleanup could move more semantics into structured inlines instead of relying on string cleanup.
- The new inline model is incremental: parser/normalizer/render now use it for DPD, but JSON serialization and any non-DPD paths still need explicit alignment if they are to benefit from the richer semantic structure.
- Current parser is intentionally scoped to the validated `bien` article shape; broader DPD shape coverage remains outside this batch.

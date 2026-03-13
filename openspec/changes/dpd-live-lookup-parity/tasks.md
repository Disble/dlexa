# Tasks: DPD Live Lookup Parity

## Phase 1: Live DPD Source Foundation

- [x] 1.1 Update `internal/config/interfaces.go`, `internal/config/static.go`, and `internal/fetch/interfaces.go` so the DPD path has explicit base URL / timeout / user-agent settings and fetched documents carry the HTTP metadata needed for typed outcomes.
- [x] 1.2 Create `internal/fetch/http.go` with the live DPD fetcher that sanitizes the lookup term, builds the canonical URL, performs bounded HTTP GETs, and classifies transport-level failures without scraping article content.
- [x] 1.3 Add table-driven fetch coverage in `internal/fetch/http_test.go` for URL construction, timeout / network failures, non-success transport handling, and successful HTML document capture.
- [x] 1.4 Update `internal/app/wiring.go` and `internal/source/registry.go` so `dpd` is the default production source, the live fetch pipeline is registered from the composition root, and `demo` remains opt-in only.

## Phase 2: Canonical Article Extraction from Live HTML

- [x] 2.1 Update `internal/parse/interfaces.go` to return source-shaped DPD parse results that can represent dictionary header data, ordered sections, nested items, citation fragments, and explicit no-article / extraction-failure outcomes.
- [x] 2.2 Add the authoritative fixture `testdata/dpd/bien.html` and any parser-local fixture helpers needed to lock the contract to a real captured upstream article.
- [x] 2.3 Create `internal/parse/dpd.go` to isolate the canonical DPD article region, discard page chrome, preserve article-local HTML semantics needed downstream, and distinguish “not found” from “article shape broken.”
- [x] 2.4 Add fixture-based parser tests in `internal/parse/dpd_test.go` that prove `bien` extraction keeps dictionary / edition / lemma / section markers while excluding menus, related-content blocks, share widgets, and footer noise.

## Phase 3: Minimal Markdown-Ready Normalization and Rendering

- [x] 3.1 Update `internal/model/types.go` and `internal/normalize/interfaces.go` to introduce only the minimal normalized article structure needed for faithful Markdown: article identity, ordered sections, nested children, paragraph inlines, references, and citation essentials, while keeping temporary compatibility fields only where migration requires them.
- [x] 3.2 Create `internal/normalize/dpd.go` to transform parsed DPD content into that minimal article model, preserving section order `1..7`, nested `6.a..6.c`, emphasis semantics, readable references, paragraph boundaries, and citation data without inventing extra schema.
- [x] 3.3 Add normalization tests in `internal/normalize/dpd_test.go` covering section ordering, nested subsection attachment, emphasis / reference preservation, whitespace cleanup, and rejection or warning paths when shell leakage or structurally invalid article shapes appear.
- [x] 3.4 Update `internal/render/markdown.go` and add `internal/render/markdown_test.go` plus `testdata/dpd/bien.md.golden` so Markdown rendering is driven from `Entry.Article` and proves parity-critical output for dictionary context, edition, lemma, ordered sections, nested subitems, readable references, and citation essentials.
- [ ] 3.5 Update `internal/render/json.go` so JSON serializes `Entry.Article` as the same minimal hierarchy used by Markdown, and add focused `internal/render/json_test.go` shape checks for article identity, section nesting, inline references, and citation fields without expanding the v1 contract beyond the Markdown-first business goal.

## Phase 4: Typed Failure Handling and Pipeline Wiring

- [x] 4.1 Update `internal/source/pipeline.go` to pass richer fetch and parse results through `fetch -> parse -> normalize` unchanged in ordering, with no fallback to bootstrap content or renderer-side HTML interpretation.
- [x] 4.2 Update `internal/query/service.go` to map pipeline failures into stable outward problem codes for `dpd_fetch_failed`, `dpd_not_found`, `dpd_extract_failed`, and `dpd_transform_failed`.
- [x] 4.3 Extend `internal/source/pipeline_test.go` and `internal/query/service_test.go` to verify stage sequencing, typed failure propagation, and the rule that production DPD lookups never degrade into demo/mock-backed success paths.

## Phase 5: Granular Fidelity Verification and Secondary Safeguards

- [x] 5.1 Keep the integration-style parity path that exercises `testdata/dpd/bien.html` through parse + normalize + render and asserts `testdata/dpd/bien.md.golden`, but treat it as the broad end-to-end Markdown guardrail rather than the only acceptance evidence.
- [x] 5.2 Expand `internal/parse/dpd_test.go` with many small fixture-based parser cases that independently prove authored quote punctuation is preserved, example-bearing fragments remain separable from prose, numeric references are extracted without duplicated markers, lexical-head source fragments stay attached to their owning items, and citation fragments are isolated from ordinary body text.
- [x] 5.3 Expand `internal/normalize/dpd_test.go` with targeted normalization cases for synthetic-quote rejection, inline emphasis/example preservation, canonical reference node shaping, integrated heading/title semantics for `bien que.`, `más bien.`, and `si bien.`, and explicit citation-field separation from article prose.
- [x] 5.4 Expand `internal/render/markdown_test.go` with targeted renderer cases that independently assert canonical Markdown for preserved quotes, readable example/emphasis output, single-arrow cross-references like `→ [6]` and `→ [7]`, integrated numbered lexical heads, and citation readability without flattened blobs.
- [x] 5.5 Refresh or split `testdata/dpd/bien.md.golden` plus any shared fixture/assert helpers so stale expectations that encode known defects are removed, and each verified defect category has a deterministic assertion path in addition to the broad golden diff.
- [x] 5.6 Add named end-to-end regression tests over `testdata/dpd/bien.html` that map each verified defect category to a specific parse + normalize + render assertion, so failures identify the broken semantic contract instead of only showing a whole-output mismatch.
- [ ] 5.7 Add secondary JSON verification in `internal/render/json_test.go` plus `testdata/dpd/bien.json.golden` that proves article identity, heading hierarchy, inline reference semantics, and citation fields remain aligned with the Markdown-first normalized article without promoting JSON richness to the primary release gate.
- [ ] 5.8 Add an opt-in live probe test or documented verification hook for `bien` drift detection that checks invariant markers, major headings, and citation essentials against upstream while remaining explicitly non-blocking and outside the deterministic fixture baseline.

## Implementation Notes

- Validated live access behavior for this environment is now explicit: direct `GET https://www.rae.es/dpd/<term>` with a browser-like `User-Agent` reaches article HTML, while low-profile/no-UA requests can still trigger Cloudflare challenge pages.
- `/srv/keys` is not a useful access path for this change.
- `go-rae` remains out of scope as a direct implementation blueprint because it talks to a third-party API rather than parsing raw DPD article HTML.

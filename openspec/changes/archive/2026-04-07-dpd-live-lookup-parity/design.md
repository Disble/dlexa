# Design: DPD Live Lookup Parity

## Technical Approach

Revise the change around the ACTUAL business goal: produce correct, current, LLM-friendly Markdown from the live DPD page. The current codebase is still bootstrap-oriented: `internal/app/wiring.go` defaults to `demo`, `internal/parse/markdown.go` is a stub that turns the whole body into one flat entry, `internal/normalize/identity.go` does no real shaping, `internal/render/markdown.go` prints `Summary` + `Content`, and `internal/query/service.go` collapses all source failures into `source_lookup_failed`.

So the design must stop pretending the primary win is a rich abstract domain. It is NOT. The primary win is a robust live `HTML -> extracted article -> markdown-ready structure -> Markdown` pipeline, with JSON as a secondary serialization of that same normalized result.

The existing repository boundaries still make sense, but only if they are justified by Markdown correctness:

- `fetch` exists to retrieve the live response reliably and classify transport outcomes.
- `parse` exists to remove page chrome and isolate article-bearing HTML.
- `normalize` exists to apply the MINIMUM structure required for high-quality Markdown.
- `render` exists to emit Markdown as the primary product and JSON as a derivative view.

The design therefore keeps the query-first ports-and-adapters shape already present in the repo, but replaces the prior canonical-model-heavy emphasis with a Markdown-first transformation pipeline.

## Software Architecture Applied

The software architecture applied is **query-first ports-and-adapters with a staged transformation pipeline**.

- **Thin entrypoint / composition root**: `cmd/dlexa` stays thin and `internal/app/wiring.go` remains the only place that wires concrete adapters.
- **Use-case orchestration**: `internal/query` remains responsible for source resolution, cache interaction, aggregation, and outward error reporting.
- **Transformation pipeline per source**: `internal/source/PipelineSource` keeps the enforced order `fetch -> parse -> normalize`.
- **Minimal domain contract**: `internal/model` holds only the structure needed to preserve article meaning in Markdown and JSON.
- **Presentation boundary**: `internal/render` formats normalized content without knowing HTML selectors or upstream page quirks.

Dependency direction remains inward:

`cmd -> app -> query -> source -> {fetch, parse, normalize} -> model <- render`

That direction is justified because Markdown generation quality depends on isolating concerns. If fetch understands article structure, or if render has to inspect raw HTML, the pipeline becomes an untestable quilombo.

## Architecture Decisions

### Decision: Optimize for Markdown-first parity, not deep canonical modeling

**Choice**: Make the normalized artifact a **minimal markdown-ready article model**, not an ambitious source-agnostic knowledge model.

**Alternatives considered**:
- Keep using flat `Entry.Content` and generate Markdown directly from strings.
- Build a rich canonical DPD ontology with many node kinds and metadata layers.
- Store raw cleaned HTML and let renderers reinterpret it.

**Rationale**: The current flat `Entry` is too weak for section ordering, nested `6.a..6.c`, emphasis, and reference handling. But going all-in on a large canonical model is overengineering relative to the corrected business priority. The right move is the middle path: keep only the structure that materially improves Markdown fidelity.

Necessary normalized structure:

- dictionary label
- edition marker
- lemma
- canonical URL
- ordered top-level sections
- nested subsection labels where present (`a)`, `b)`, `c)`)
- ordered paragraphs within sections
- inline emphasis and reference markers
- citation essentials

Formatting-critical structure that MUST remain first-class until render:

- `Section.Title` for lexical heads such as `bien que.` and `más bien.` instead of flattening them into neighboring prose
- `Paragraph.Inlines` for source-order text, examples, mentions, glosses, citations, bibliography fragments, work titles, small-caps markers, editorial scaffolds, corrections, patterns, exclusion markers, and typed references when punctuation/readability depends on token boundaries
- `Citation.SourceLabel`, `Citation.CanonicalURL`, `Citation.Edition`, and `Citation.ConsultedAt` as authoritative citation fields; `Citation.Text` is a renderer convenience, not the source of truth

Optional / explicitly deferred structure:

- generic DOM AST in `internal/model`
- deep semantic taxonomies for every phrase type
- per-node provenance metadata beyond what debugging requires
- reusable cross-dictionary abstractions not needed for DPD parity
- page-layout fidelity or chrome representation

Illustrative direction:

```go
type Entry struct {
    ID       string
    Headword string
    Summary  string
    Content  string
    Source   string
    URL      string
    Metadata map[string]string
    Article  *Article
}

type Article struct {
    Dictionary   string
    Edition      string
    Lemma        string
    CanonicalURL string
    Sections     []Section
    Citation     Citation
}

type Section struct {
    Label      string
    Level      int
    Paragraphs []Paragraph
    Children   []Section
}

type Paragraph struct {
    Inlines []Inline
}

type Inline struct {
    Kind   string // text, emphasis, reference
    Text   string
    Target string
}

type Citation struct {
    SourceLabel  string
    Edition      string
    CanonicalURL string
    ConsultedAt  string
}
```

This is enough to produce good Markdown and structured JSON. Anything beyond this needs a proven business case, not architecture fan fiction.

### Decision: Parse owns article extraction; normalize owns Markdown-oriented shaping

**Choice**: Keep extraction and shaping separate.

**Alternatives considered**:
- Make fetch return already-clean article HTML.
- Make parse directly produce final Markdown.
- Collapse parse + normalize into one scraper package.

**Rationale**: These boundaries align with the existing codebase and make testing sane. `parse` should answer: “what is the article body, stripped of chrome?” `normalize` should answer: “what minimal structure is needed so Markdown is correct and stable?” Mixing them destroys debuggability.

### Decision: Preserve formatting semantics as structure, not synthesized strings

**Choice**: Keep formatting-critical semantics explicit through `parse -> normalize -> render`, and only let the renderer choose the final Markdown spelling.

**Alternatives considered**:
- Convert article HTML directly into paragraph-level Markdown strings during parse.
- Re-synthesize quote marks, example punctuation, or reference text during rendering.
- Treat lexical heads and citations as plain text blobs because they look correct in one golden file.

**Rationale**: The verified defects are exactly what happens when semantics are flattened too early: malformed quote normalization, lost example meaning, broken cross-references, split lexical headings, and unreadable citations. The existing model already has enough structure in the right places (`Section.Title`, `Paragraph.Inlines`, citation fields); the design refinement is to make those responsibilities explicit instead of expanding scope into a full AST.

Concrete responsibility split:

- **Parse** preserves source-local punctuation and token boundaries for quotes, examples, emphasis, references, lexical heads, and citation fragments; it MUST NOT invent replacement quote characters or merge adjacent semantic runs just to get provisional Markdown.
- **Normalize** converts parsed fragments into canonical article structure: lexical heads become `Section.Title`, inline references become dedicated inline nodes with display text plus target, examples/mentions/glosses/citations/corrections/pattern markers remain distinct inline runs, and citation fields stay structured.
- **Render** is responsible only for deterministic terminal Markdown presentation. It MAY choose the final syntax for emphasis and links, but it MUST preserve the normalized meaning and contiguous readable forms. `Paragraph.Markdown` is now a compatibility projection; `Paragraph.Inlines` is the semantic source of truth for DPD rendering.

### Decision: Markdown is the primary renderer contract; JSON is secondary and derived

**Choice**: Treat Markdown as the primary output target and make JSON serialize the same normalized structure.

**Alternatives considered**:
- Keep JSON as a raw `LookupResult` dump.
- Make JSON the primary contract and derive Markdown from a richer data tree.
- Pre-render Markdown and place only strings in JSON.

**Rationale**: The corrected business goal is LLM ingestion quality, and that means Markdown quality is the main product. JSON still matters, but as a secondary representation of the same result, not as the reason the model exists.

### Decision: Introduce typed failure taxonomy at the lookup boundary

**Choice**: Replace `source_lookup_failed` flattening with typed failure classes preserved through `internal/query/service.go`.

**Alternatives considered**:
- Keep the current catch-all error.
- Hide failures in warnings.
- Leak raw adapter errors to the CLI.

**Rationale**: A live HTML pipeline can fail in materially different ways. Transport issues, article absence, extraction breakdown, and normalization breakdown are not the same operational problem. If the orchestrator can’t distinguish them, debugging parity regressions becomes amateur hour.

## Data Flow

### Runtime lookup flow

```text
CLI
  -> internal/app loads runtime config and selected sources
  -> internal/query resolves source registry
  -> source.PipelineSource.Lookup()
       -> fetch.DPDFetcher.Fetch()            // live HTML + transport metadata
       -> parse.DPDArticleParser.Parse()      // isolate article shell, strip chrome
       -> normalize.DPDMarkdownNormalizer.Normalize() // markdown-ready structure
  -> query aggregates entries, warnings, typed problems
  -> render.MarkdownRenderer renders Markdown
  -> render.JSONRenderer serializes the same normalized article
```

### Boundary translation

```text
remote DPD HTML
    |
    v
extracted article fragment + source-shaped nodes
    |
    v
minimal normalized article model
    |
    +--> markdown renderer (primary)
    |
    +--> json renderer (secondary)
```

The important part is not “many layers because patterns are cool.” The important part is that each translation removes uncertainty before the next stage.

## Boundary Changes and Dependency Direction

### Composition root (`internal/app`)

Current evidence: `internal/app/wiring.go` sets `DefaultSources: []string{"demo"}` and wires only the static bootstrap source.

Revised responsibility:

- register a live `dpd` pipeline source
- switch default production source from `demo` to `dpd`
- optionally keep `demo` as explicit non-default bootstrap/debug source
- inject runtime knobs required for live HTTP fetches

Why this boundary matters: defaulting to `demo` is the exact mismatch with the business goal. The fix belongs in composition, not hacked into query logic.

### Fetch boundary (`internal/fetch`)

Current evidence: `fetch.Document` has `URL`, `ContentType`, `Body`, `RetrievedAt`, but no status code or final transport context.

Revised responsibility:

- sanitize query input for path use
- build the DPD URL from configurable base URL + term
- perform bounded HTTP GET with timeout and explicit user agent
- capture response status, content type, effective URL, body, and retrieval time
- classify transport-level failures early

Recommended contract evolution:

```go
type Document struct {
    URL         string
    ContentType string
    StatusCode  int
    Body        []byte
    RetrievedAt time.Time
}
```

Fetch must stay dumb about article structure. If fetch starts scraping, the design is already rotting.

### Parse boundary (`internal/parse`)

Current evidence: `internal/parse/interfaces.go` returns `[]model.Entry`, and `internal/parse/markdown.go` just turns the whole body into one flat entry.

Revised responsibility:

- identify the article-bearing region of the DPD page
- exclude page chrome: menus, headers, related-content blocks, share widgets, newsletter/footer, and other non-article shell content
- preserve article-local structure needed downstream: heading, edition, section boundaries, nested subitems, paragraphs, emphasis, references, citation block
- distinguish “no canonical article present” from “document shape is broken”

Recommended contract direction:

```go
type Result struct {
    Articles []ParsedArticle
}

type ParsedArticle struct {
    Dictionary string
    Edition    string
    Lemma      string
    Sections   []ParsedSection
    Citation   ParsedCitation
}

type Parser interface {
    Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (Result, []model.Warning, error)
}
```

Parse is where chrome removal belongs, because chrome removal is an HTML-shape concern.

### Normalize boundary (`internal/normalize`)

Current evidence: `internal/normalize/interfaces.go` accepts `[]model.Entry`, and `internal/normalize/identity.go` only stamps metadata.

Revised responsibility:

- convert parsed source-shaped nodes into a minimal markdown-ready article
- preserve exact section order
- attach nested items to the right parent section
- normalize whitespace and paragraph boundaries
- convert inline emphasis and references into typed inline markers
- assemble citation essentials cleanly
- reject or warn on leaked shell content or structurally invalid section layouts

Recommended contract direction:

```go
type Normalizer interface {
    Normalize(ctx context.Context, descriptor model.SourceDescriptor, result parse.Result) ([]model.Entry, []model.Warning, error)
}
```

Normalization is intentionally small. It should improve Markdown correctness, not invent semantics nobody asked for.

### Render boundary (`internal/render`)

Current evidence:

- `internal/render/markdown.go` prints entry `Summary` and `Content` directly.
- `internal/render/json.go` marshals the whole `LookupResult` as-is.

Revised responsibility:

- Markdown renderer traverses `Entry.Article` as the authoritative content source
- JSON renderer serializes the same article structure explicitly
- legacy `Summary` / `Content` stay as compatibility fields during migration only

The renderers must not inspect raw HTML. If they do, the pipeline failed upstream.

## Live Fetch Strategy

1. Trim and sanitize the lookup term.
2. Construct the canonical DPD path under the configured base URL.
3. Execute HTTP GET with context cancellation, timeout, and explicit user agent.
4. Record status code, content type, final URL, and body bytes.
5. Classify outcomes:
   - transport/network failure
   - upstream not found / no article outcome
   - successful HTML retrieval
6. Pass only successful HTML documents to the parser.

Why this strategy:

- it satisfies the live-content requirement directly
- it keeps content freshness aligned with upstream by default
- it avoids introducing a fake API or database-backed indirection
- it creates test seams around fetch behavior without mixing in HTML semantics

What this strategy intentionally does NOT do:

- no production fixture serving
- no internal database fallback
- no crawler/general browser automation layer
- no speculative semantic extraction during fetch

## Article Extraction from Page Chrome

The extraction step must treat the full DPD page as hostile input until proven otherwise.

Extraction goals:

- locate the main article container reliably
- keep dictionary title, edition, lemma, numbered sections, nested subitems, and citation content
- discard unrelated visible text from menus, related links, share controls, newsletter/footer, and other shell blocks

Extraction design principles:

- prefer stable article-centric anchors over broad page-wide text scraping
- preserve local HTML semantics inside the article region until normalization
- fail explicitly if no credible article region can be isolated
- treat shell leakage as a correctness defect, not as acceptable noise

The parser should return a source-shaped article fragment that still remembers enough HTML-local detail to support the Markdown transformation, but that detail must stay parser-local. It does NOT belong in `internal/model`.

## HTML-to-Markdown Transformation Strategy

The core solution is a robust HTML-to-Markdown transformation pipeline.

Transformation stages:

1. **Acquire live HTML** from DPD.
2. **Extract only the canonical article region**.
3. **Interpret article-local HTML semantics** into parsed sections, lexical heads, paragraphs, examples, emphasis spans, references, and citation fragments.
4. **Normalize into a markdown-ready article model** with first-class formatting semantics where readability depends on structure.
5. **Render Markdown deterministically** from that model.

Transformation rules for v1:

- headings become article header fields, not ad-hoc text blobs
- numbered sections remain ordered and labeled
- nested lettered subitems remain children of the owning section
- quote punctuation from the source is preserved when it carries example/definition meaning; normalization MAY clean whitespace, but MUST NOT synthesize replacement quotation styles or move punctuation across inline boundaries
- lexical heads such as `bien que.` become heading/title semantics (`Section.Title`), not the first sentence of a paragraph and not a split label/title pair assembled by string heuristics
- examples and emphasis survive as semantic inline runs through normalization so render can emit readable terminal Markdown without losing contrastive terms or example boundaries
- references are preserved as dedicated inline references with contiguous readable display text plus target; the canonical readable terminal form is arrow + bracketed section marker (for example `→ [6]`) linked when a target exists
- citation data remains structured until render and is emitted as readable terminal Markdown assembled from canonical citation fields rather than a prematurely flattened blob
- unsupported decorative markup is dropped if it does not change article meaning
- unsupported meaningful markup degrades to plain text with warning rather than silently vanishing

This design deliberately prefers a **controlled subset transformation** over a generic HTML-to-Markdown dump. Generic dumps are easy to demo and terrible for correctness.

## Minimal Normalization Needed for Good Markdown Output

Normalization should do only what materially improves output quality:

- trim incidental whitespace
- merge text runs that were split by markup noise
- preserve paragraph boundaries
- preserve section ordering and nesting
- preserve source quote punctuation wherever it distinguishes examples, glosses, or cited forms
- normalize lexical heads into section/title semantics instead of leaving them embedded in paragraph strings
- normalize reference display text into a canonical readable form without separating the arrow, brackets, and target meaning
- normalize citation fields into explicit metadata while keeping enough structure to render a readable terminal citation line
- surface warnings when extraction left suspicious shell fragments

Normalization should NOT:

- build a full generic AST
- infer deep linguistic semantics not explicit in the source
- restructure article meaning for aesthetic formatting
- depend on renderer-specific markdown strings as its internal representation

## Renderer Implications (Markdown Primary, JSON Secondary)

### Markdown renderer

Markdown is the PRIMARY business output.

It must:

- render dictionary context, edition, and lemma clearly
- emit ordered sections in source order
- preserve nested `a)`, `b)`, `c)` items under section `6.`
- preserve emphasis in readable Markdown form
- render cross-references readably for LLM ingestion
- append citation essentials without dragging page chrome into the result

### JSON renderer

JSON is SECONDARY.

It should:

- expose the same normalized article structure and metadata
- avoid becoming a richer contract than Markdown requires
- remain useful for inspection/tests/tooling without driving the architecture

Practical consequence: the architecture should be judged by Markdown correctness first. JSON exists to serialize the same meaning, not to justify overengineering the model.

## Error Taxonomy

External stable problem classes:

| Class | Suggested code | Boundary owner | Meaning |
|------|------|------|------|
| Fetch failure | `dpd_fetch_failed` | fetch/query | network error, timeout, unusable upstream response |
| Not found | `dpd_not_found` | fetch or parse/query | no canonical DPD article exists for the term |
| Extraction failure | `dpd_extract_failed` | parse/query | HTML retrieved, but article body cannot be isolated safely |
| Transformation failure | `dpd_transform_failed` | normalize/query | article extracted, but required Markdown-ready structure cannot be produced |

Internal interpretation:

- `fetch_failed` = transport/runtime problem
- `not_found` = no article outcome, not a parser defect
- `extract_failed` = page-shape problem before normalized structure exists
- `transform_failed` = minimal structure contract could not be satisfied

This is more useful than today’s single `source_lookup_failed` bucket and still stays aligned with the spec’s required distinctions.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/app/wiring.go` | Modify | Register live `dpd` source and make it the default instead of `demo`. |
| `internal/config/interfaces.go` | Modify | Add runtime settings for DPD base URL, timeout, and user agent. |
| `internal/config/static.go` | Modify | Provide live DPD defaults for the new runtime settings. |
| `internal/fetch/interfaces.go` | Modify | Enrich `Document` with HTTP status metadata needed for typed outcomes. |
| `internal/fetch/http.go` | Create | Live DPD fetcher using explicit HTTP behavior. |
| `internal/parse/interfaces.go` | Modify | Return parser result types instead of normalized entries directly. |
| `internal/parse/dpd.go` | Create | Extract DPD article content from live HTML and remove page chrome. |
| `internal/normalize/interfaces.go` | Modify | Accept parser result types and emit markdown-ready entries. |
| `internal/normalize/dpd.go` | Create | Minimal normalization for sections, nesting, emphasis, references, and citations. |
| `internal/model/types.go` | Modify | Add minimal article structure and typed problem helpers while keeping legacy compatibility fields temporarily. |
| `internal/source/pipeline.go` | Modify | Preserve pipeline order while passing richer parser output into normalization. |
| `internal/query/service.go` | Modify | Map typed pipeline failures to stable outward problem codes. |
| `internal/render/markdown.go` | Modify | Render normalized article structure instead of flat `Content`. |
| `internal/render/json.go` | Modify | Serialize normalized article structure rather than only dumping `LookupResult`. |
| `internal/source/registry.go` | Possibly modify | Keep demo available as opt-in source while `dpd` becomes default. |
| `internal/source/pipeline_test.go` | Modify | Verify stage ordering and failure propagation with richer parser results. |
| `internal/query/service_test.go` | Modify | Verify error taxonomy mapping and source aggregation behavior. |
| `internal/render/*_test.go` | Modify/Create | Add Markdown-first golden coverage and secondary JSON structure coverage. |
| `testdata/dpd/bien.html` | Create | Captured real upstream fixture for deterministic extraction tests. |
| `testdata/dpd/bien.md.golden` | Create | Expected Markdown output for parity-sensitive golden tests. |
| `testdata/dpd/bien.json.golden` | Create | Expected secondary JSON shape derived from the same normalized article. |

## Interfaces / Contracts

Illustrative contract direction:

```go
type ParsedArticle struct {
    Dictionary string
    Edition    string
    Lemma      string
    Sections   []ParsedSection
    Citation   ParsedCitation
}

type ParsedSection struct {
    Label      string
    Level      int
    Paragraphs []ParsedParagraph
    Children   []ParsedSection
}

type Normalizer interface {
    Normalize(ctx context.Context, descriptor model.SourceDescriptor, result parse.Result) ([]model.Entry, []model.Warning, error)
}
```

Contract rules:

- parser output may still be source-shaped
- normalized output must be markdown-ready and renderer-independent
- Markdown renderer consumes normalized article as authoritative content
- JSON renderer serializes the same normalized article
- no downstream stage reparses HTML

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Live fetch URL building and transport classification | Table-driven tests with stubbed HTTP transport, verifying timeout/network/status handling. |
| Unit | Article extraction from real DPD HTML | Fixture-based tests using captured `bien.html`, asserting article region selection and chrome exclusion. |
| Unit | Formatting-semantic normalization inputs | Many small targeted tests verifying quote punctuation, examples, lexical heads, emphasis, references, and citations survive transformation into normalized article data without flattening. |
| Unit | Markdown rendering correctness | Targeted renderer tests for each formatting case plus golden tests using normalized article input and expected `bien.md.golden`. |
| Unit | JSON secondary rendering | Golden/snapshot tests asserting JSON serializes the same normalized structure, not a different semantic contract. |
| Integration | Pipeline sequencing and parity baseline | `internal/source/pipeline_test.go` keeps strict `fetch -> parse -> normalize` ordering and typed failure propagation; parse + normalize + render integration keeps the end-to-end `bien` fixture honest. |
| Integration | Query error reporting | `internal/query/service_test.go` verifies `dpd_fetch_failed`, `dpd_not_found`, `dpd_extract_failed`, and `dpd_transform_failed` mapping. |
| Optional live probe | Upstream drift detection | Opt-in probe against live `bien` validating invariant markers and major Markdown sections without replacing fixture-based contracts. |

Testing priority is simple:

1. real captured fixtures are the baseline for deterministic correctness
2. many small targeted tests guard each formatting defect class independently: quote fidelity, example semantics, lexical heads, references, citation readability
3. Markdown golden output is the integrated business assertion, not the only contract
4. live probes are useful drift alarms, not the only truth source

Golden-file guardrail:

- a stale or defective golden file MUST be treated as a broken test artifact, not as proof that flattened output is acceptable
- targeted assertions for each formatting case are required so one broad golden diff does not hide which semantic contract regressed
- live drift checks remain optional and non-blocking for bounded v1 acceptance; they exist to detect upstream changes, not to redefine the fixture contract

## Migration / Rollout

No data migration is required.

Implementation rollout plan:

1. Add minimal article types to `internal/model` while keeping `Summary` and `Content` for temporary compatibility.
2. Introduce live DPD fetch/parse/normalize adapters.
3. Make `dpd` the default source in `internal/app/wiring.go`.
4. Keep `demo` only as explicit non-production bootstrap/debug source.
5. Move Markdown renderer to prefer `Entry.Article`.
6. Move JSON renderer to serialize the same normalized article.
7. Replace demo-centric tests with fixture-backed parity tests around `bien`.
8. Later, once stable, remove or quarantine flat-content assumptions that are only there for bootstrap compatibility.

This rollout is intentionally low-risk. You don’t rip out the whole house just because one room was mocked badly.

## Design Pattern Decisions

The pattern discussion is retained because the spec requires it, but the justifications are tied to the REAL priority: correct Markdown from live HTML.

### Creational

#### Pattern actually used: Constructor-based Dependency Injection

- **Where**: `internal/app/wiring.go`
- **Why**: The composition root already wires concrete adapters explicitly. That is the cleanest way to swap `demo` for live `dpd` without leaking live-fetch details into use-case code. It improves maintainability and keeps the Markdown pipeline replaceable in tests.

#### Pattern actually used: Simple Factory through registry registration

- **Where**: `source.NewStaticRegistry(...)` and `render.NewRegistry(...)`
- **Why**: Source and renderer lookup already behave as lightweight factories. That is enough for one primary live source plus one fallback demo source. Anything heavier would be ceremony without output-quality benefit.

#### Pattern intentionally NOT used: Builder

- **Why not**: The normalized article is derived deterministically from parsed input. No caller is incrementally constructing it. Builder adds API surface and mutability where none is needed.

#### Pattern intentionally NOT used: Abstract Factory

- **Why not**: There is no family explosion of interchangeable platform objects. Explicit construction in the composition root is clearer and directly supports the Markdown-first goal.

### Structural

#### Pattern actually used: Adapter

- **Where**: DPD fetcher adapts HTTP transport into `fetch.Document`; parser adapts live HTML into parsed article nodes; normalizer adapts parsed nodes into the markdown-ready article model.
- **Why**: This is the main structural protection against upstream HTML chaos leaking into renderers.

#### Pattern actually used: Facade

- **Where**: `source.PipelineSource`
- **Why**: Query orchestration should ask for a lookup result, not micromanage article extraction and Markdown shaping steps. The facade keeps the multi-stage pipeline behind one stable source boundary.

#### Pattern actually used: Anti-Corruption Layer / Translator

- **Where**: The parse + normalize boundary between live DPD HTML and internal article data.
- **Why**: Upstream HTML is not the product. Markdown is. This translator layer prevents page-specific quirks from infecting the rest of the app.

#### Pattern intentionally NOT used: Decorator

- **Why not**: The main problem is transformation quality, not stacking optional cross-cutting behavior around sources.

### Behavioral

#### Pattern actually used: Pipeline

- **Where**: `fetch -> parse -> normalize`
- **Why**: It matches the existing repository structure and directly isolates failures in live retrieval, chrome extraction, and Markdown shaping.

#### Pattern actually used: Strategy

- **Where**: source selection in `source.Registry`; output selection in `render.Registry`
- **Why**: Runtime selection is already present, and it allows `dpd` to become the default production strategy while keeping `demo` opt-in.

#### Pattern actually used: Typed error policy in the use-case layer

- **Where**: `internal/query/service.go`
- **Why**: The system must respond differently to fetch failure, absence, extraction failure, and transformation failure. That behavior belongs in the lookup policy layer because it defines outward product behavior.

#### Pattern intentionally NOT used: Visitor

- **Why not**: Recursive traversal of sections/paragraphs/inlines is sufficient for the current renderer needs. Visitor would increase indirection without improving Markdown quality.

## Open Questions

- [ ] Confirm with live evidence whether DPD not-found terms return HTTP 404, a 200 shell page without canonical article content, or both depending on path shape.
- [ ] Confirm whether citation consultation time must be injected at fetch time, render time, or result assembly time for deterministic tests.
- [ ] Confirm whether section references need stable internal targets in JSON, or whether readable display text is sufficient for v1 parity.

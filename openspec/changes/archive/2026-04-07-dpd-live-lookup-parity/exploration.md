## Exploration: dpd-live-lookup-parity

### Current State

`dlexa` already has the correct macro-boundaries for a real dictionary source, but today the default flow is still bootstrap scaffolding.

- `cmd/dlexa/main.go` stays thin.
- `internal/app/wiring.go` wires a single fake `demo` source backed by `fetch.NewStaticFetcher(...)`, `parse.NewMarkdownParser()`, and `normalize.NewIdentityNormalizer()`.
- `internal/query/service.go` already provides the right orchestration seam: cache-aside lookup, source selection, result aggregation, and partial-failure reporting.
- `internal/source/pipeline.go` already enforces the boundary we want to preserve: `fetch -> parse -> normalize`.
- `internal/render/markdown.go` and `internal/render/json.go` render the aggregated `LookupResult`, but the current markdown output is diagnostic, not article-faithful.

What is fake today:

- `internal/fetch/static.go` fabricates markdown and a fake `.invalid` URL.
- `internal/parse/markdown.go` emits one bootstrap entry regardless of source structure.
- `internal/normalize/identity.go` only stamps metadata and passes parsed fields through.
- `internal/model/types.go` only supports a flat `Entry` with `Summary`, `Content`, and string metadata.

From direct retrieval of `https://www.rae.es/dpd/bien`, the parity target for `dlexa bien` is now materially clearer. The canonical article content includes, at minimum:

- dictionary heading/context: `Diccionario panhispánico de dudas`
- edition marker: `2.ª edición`
- lemma heading: `bien`
- numbered top-level sections `1.` through `7.`
- nested lettered subitems inside section `6.`: `a)`, `b)`, `c)`
- inline emphasis semantics for forms/locutions/cross-references (for example `*más bien*`, `*mejor*`, `*si bien*`)
- cross-references rendered as navigable references in the source (`→ [6]`, `→ [7]`)
- citation/footer metadata with source, URL, edition, and consultation date

The fetched page also contains massive site chrome, menus, related content, share widgets, newsletter/footer content, and help navigation. That noise MUST NOT become canonical article content.

### Affected Areas

- `internal/app/wiring.go` — switch default source from `demo` to a real `dpd` pipeline while preserving composition-root discipline.
- `internal/config/interfaces.go` / `internal/config/static.go` — carry default source/runtime knobs for remote DPD acquisition.
- `internal/fetch/interfaces.go` — likely expand `Document` to include final URL, status code, and maybe headers/charset for real HTTP behavior.
- `internal/parse/interfaces.go` + new DPD parser implementation — parse structured article HTML/content, not synthetic markdown.
- `internal/normalize/interfaces.go` + new normalizer — map parsed nodes into a canonical DPD article model.
- `internal/model/types.go` — evolve beyond flat `Entry` so sections, subitems, inline semantics, references, and citation metadata are first-class instead of metadata soup.
- `internal/render/markdown.go` — render article structure in parity order and preserve emphasis/reference semantics.
- `internal/render/json.go` — expose the structured article model, not only flattened text.
- `internal/query/service.go` — classify remote fetch/not-found/parse failures cleanly without collapsing them into one generic error.
- `internal/source/registry.go` — keep the same selection model, but default registry contents change.
- `internal/*_test.go` — current tests verify seams/orchestration only; parity needs contract-oriented parser/normalizer/render tests.

### Approaches

1. **Canonical article model first** — fetch live DPD HTML, extract article body, parse into a rich article AST/domain model, render from that.
   - Pros:
     - Preserves existing `fetch -> parse -> normalize -> render` architecture.
     - Supports both markdown/text parity and structured JSON.
     - Makes acceptance criteria testable by section/subitem instead of by brittle whole-page text.
   - Cons:
     - Requires expanding `internal/model` deliberately.
     - More design work up front.
   - Effort: High

2. **Rendered-text parity first** — strip article body and emit near-verbatim formatted text with minimal structure.
   - Pros:
     - Faster to get something visually similar for `bien`.
   - Cons:
     - Brittle against markup changes.
     - Weak JSON story.
     - Fights the current architecture by smearing parse/render concerns together.
   - Effort: Medium

3. **Staged canonical parity** — model only the article structures needed for `bien` first, but do it in a canonical representation that can later absorb more DPD variants.
   - Pros:
     - Best fit for the current repo maturity.
     - Keeps architecture clean.
     - Uses `bien` as a concrete parity anchor instead of vague “real DPD-like” nonsense.
   - Cons:
     - First increment will not cover all DPD article families.
     - Scope discipline is mandatory.
   - Effort: Medium-High

### Recommendation

Recommend **Approach 3: Staged canonical parity**.

That’s the serious engineering call. `go-rae` is useful as a reference for boundary discipline around remote lookup and strongly typed response models, but NOT as the core solution architecture for this change. That repo talks to `rae-api.com` JSON endpoints through a single client package (`client.go`) and returns already-structured payloads (`entities.go`). `dlexa`, by contrast, must obtain and interpret live DPD article content directly, without mocks and without an internal DB. So copying `go-rae` blindly would be tutorial-programmer nonsense.

Useful learnings from `go-rae`:

- keep network concerns explicit and configurable (`New(...)`, timeout/version options)
- keep returned data strongly typed instead of passing around raw maps/strings
- surface “not found” as a domain-level result shape, not just transport failure

What is NOT reusable from `go-rae`:

- its parsing strategy, because there effectively isn’t an HTML/article parser — the upstream API already did that work
- its model shape, because DLE/RAE API senses and conjugations are not the same domain as DPD article sections/locutions/citations
- its testing style as-is, because it relies on skipped live tests and thin client coverage, which is too weak for parity-sensitive rendering

For `dlexa bien`, the next phase should treat parity as:

- same article identity (`bien`, DPD, 2.ª edición)
- same ordered information hierarchy (sections `1..7`, nested `6.a..6.c`)
- same preserved emphasis/reference semantics in rendered output
- same citation essentials
- not the same site chrome, social widgets, or auxiliary help/navigation blocks

### Risks

- **Model pressure**: the current `Entry` type is too flat; jamming DPD structure into `Content` plus metadata would be architectural malpractice.
- **Reference mismatch**: `go-rae` is a JSON API client for a different source/product. It is only partially relevant.
- **Parser fragility**: if extraction depends on incidental page layout instead of article semantics, parity will rot fast.
- **Article variability**: `bien` is a good anchor, but DPD contains other article shapes that may require additional nodes later.
- **Acceptance ambiguity**: “same information” must be specified as structural/rendering invariants, not hand-wavy screenshot vibes.
- **Testing tension**: no mocks in production and no internal DB does NOT mean “no fixtures ever”; stable tests still need captured authoritative article documents or live opt-in integration probes.

### Ready for Proposal

Yes — with tighter acceptance language than the previous exploration had.

The proposal/spec/design should lock down these conclusions:

- **Parity definition for `dlexa bien`**: output MUST include heading/context, edition, lemma, ordered sections `1..7`, nested `6.a..6.c`, preserved emphasis/cross-reference semantics, and citation footer metadata; it MUST exclude site chrome.
- **Article structure modeling**: introduce a canonical article representation with first-class heading metadata, ordered sections, nested blocks/items, inline spans/marks, references, and citation metadata.
- **Parser/normalizer requirements**: extractor MUST isolate the article body from page chrome; parser MUST preserve section numbering/order and inline semantic marks; normalizer MUST produce stable canonical nodes independent of incidental HTML wrappers.
- **Rendering acceptance criteria**: markdown/text output MUST preserve hierarchy, numbering, nested lettering, emphasis, and cross-reference readability; JSON output SHOULD expose the same structure directly.
- **Test strategy**: use real captured DPD article fixtures as contract inputs for parse/normalize/render tests, plus optional live integration probes behind explicit opt-in. Production behavior uses live HTTP only; tests do not need an internal DB and should avoid fake production flows.

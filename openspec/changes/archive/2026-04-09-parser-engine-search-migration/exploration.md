## Exploration: parser-engine-search-migration

### Current State
#### 1. Current search parser reality
- `internal/parse/live_search.go` contains the real live-search HTML parsing logic today. `LiveSearchParser.Parse(...)` is still a legacy parser signature (`ctx, descriptor, document`) and uses file-local regex selectors/helpers (`reLiveSearch*`, `parseLiveSearchItem`, `resolveLiveSearchURL`, `normalizeLiveSearchText`, `firstMatchGroup`). Behavior is already well-defined: empty body or no `<li>` items returns no records/no error; broken markup that yields zero usable records after matching items returns `ProblemCodeDPDSearchParseFailed`.
- `internal/parse/dpd_search.go` contains the real DPD search JSON parsing logic today. `DPDSearchParser.Parse(...)` still uses the legacy signature and depends on the file-local helper `splitSearchRecord`. Behavior is also explicit: invalid JSON or a non-empty payload with zero usable `display|article_key` rows returns `ProblemCodeDPDSearchParseFailed`.
- `internal/search/provider.go` already has the migration seam. `PipelineProvider` stores both a legacy `parser Parser` and an engine `engine parseengine.SearchParser`. `NewPipelineProvider(...)` wraps legacy parsers with `parseengine.AdaptLegacySearchParser(...)`. `NewEnginePipelineProvider(...)` already exists and bridges an engine parser back into the provider’s legacy `Parser` field via `engineSearchParserBridge`.
- `internal/search/service.go` is parser-agnostic. It only calls `provider.Search(...)`, then aggregates candidates/warnings/problems/cache state. There is no direct parser selection logic here.
- `internal/app/wiring.go` still wires both search providers through `NewPipelineProvider(...)` with `parse.NewLiveSearchParser()` and `parse.NewDPDSearchParser()`. So runtime search is still entering through legacy parser constructors, even though provider infrastructure can already accept engine-native search parsers.
- `internal/parse/engine/*` currently provides only the shared `ParseInput`, `SearchParser` port, generic legacy adapter, and resolver scaffolding. There are no search-surface-specific engine parser types yet.

#### 2. Migration boundary for this slice
- This slice can stop at search-family adoption. It does NOT need article-family migration and does NOT need to move `internal/search/service.go` to a resolver-based design.
- The minimal useful runtime change is: wire search providers with engine-native search parser entrypoints, while preserving the same parsed record outputs and downstream normalizer behavior.
- Because `internal/parse/engine` imports `internal/parse`, concrete parser types cannot directly implement `engine.SearchParser` while staying in `internal/parse`; importing `engine.ParseInput` into `internal/parse` would create a package cycle. That makes “just add `ParseSearch` to the existing parse types” the wrong move.
- Therefore the low-risk boundary is wrapper-based engine-native adoption, not a package move and not a signature change inside `internal/parse`.

### Affected Areas
- `internal/parse/engine/search_ports.go` — existing engine search port that wrappers must satisfy.
- `internal/parse/engine/bridge_search.go` — existing generic legacy adapter; may stay as compatibility scaffolding.
- `internal/parse/live_search.go` — current live-search parsing logic; should remain the behavior source for this slice.
- `internal/parse/dpd_search.go` — current DPD search parsing logic; should remain the behavior source for this slice.
- `internal/search/provider.go` — already supports engine-native parsers; likely no production change needed unless tests/convenience helpers are added.
- `internal/app/wiring.go` — main runtime adoption point; swap to `NewEnginePipelineProvider(...)` with engine-native search parser constructors.
- `internal/parse/live_search_test.go` — defines live-search parser behavior fidelity.
- `internal/parse/dpd_search_test.go` — defines DPD search parser behavior fidelity.
- `internal/search/provider_test.go` — verifies engine-backed provider behavior remains identical.
- `internal/app/wiring_test.go` — currently proves runtime wiring still uses generic legacy adapters; must be updated to assert concrete engine-native search parsers instead.
- `internal/search/live_service_test.go` — protects parse/normalize error propagation through the search service.

### Approaches
1. **Move concrete search parsers into `internal/parse/engine` now** — relocate the real parsing logic and helpers into the engine package.
   - Pros: strongest architectural alignment; no wrapper indirection afterward.
   - Cons: higher churn, package/type rename fallout, broader test rewrites, helper movement risk, and it couples this search slice to a larger parser-package reorganization.
   - Effort: High.

2. **Add engine-native search wrappers and adopt them in wiring** — keep the parsing logic in `internal/parse`, create named engine search parsers in `internal/parse/engine`, and wire runtime through `NewEnginePipelineProvider(...)`.
   - Pros: minimal behavior risk, no package cycle, runtime now enters via engine-native parser types, existing parse tests keep validating the real logic, and article-family migration stays isolated for later.
   - Cons: temporary duplication of concepts (`parse.LiveSearchParser` vs `parseengine.LiveSearchParser`), plus small wrapper boilerplate.
   - Effort: Low.

3. **Refactor search provider/service to resolve parsers via engine resolver now** — thread `parseengine.Resolver` deeper into search runtime.
   - Pros: more “complete” engine architecture adoption.
   - Cons: unnecessary for this slice, widens blast radius into provider/service orchestration, and adds indirection without changing observable behavior.
   - Effort: Medium.

### Recommendation
#### 3. Safe implementation strategy
Use **Approach 2**.

Minimal migration path after foundation:
1. Add **named engine-native search parser wrappers** in `internal/parse/engine` for:
   - `LiveSearchParser`
   - `DPDSearchParser`
2. Each wrapper should satisfy `engine.SearchParser` by accepting `ParseInput` and delegating to the existing legacy parser implementation in `internal/parse`.
3. Keep `internal/parse/live_search.go` and `internal/parse/dpd_search.go` as the behavior source for this slice. Do NOT move parsing logic or helpers yet.
4. Update `internal/app/wiring.go` to use `searchsvc.NewEnginePipelineProvider(...)` with the new engine-native wrappers instead of `NewPipelineProvider(...)` + generic legacy adapters.
5. Keep `internal/search/provider.go` and `internal/search/service.go` structurally unchanged unless a tiny testability adjustment is needed. The provider already converts engine parsers into its internal legacy bridge, so runtime execution will flow through `ParseSearch(...)` without forcing service-layer churn.
6. Update wiring/provider tests to assert concrete engine-native parser usage instead of `*parseengine.LegacySearchAdapter`.
7. Add wrapper-focused engine tests that prove the new named wrappers preserve descriptor/document propagation and legacy outputs.

This gives the repo an explicit engine-native search surface now, without paying the cost of package relocation or resolver threading.

### Risks
#### 4. Risks
- **Package/type name confusion**: `parse.LiveSearchParser` vs `parseengine.LiveSearchParser` and `parse.DPDSearchParser` vs `parseengine.DPDSearchParser` can be easy to misuse in tests/wiring.
- **False-positive “cleanup” risk**: moving regex selectors or helper functions during the same slice increases drift risk for no architectural win.
- **Behavior drift around empty vs broken payloads**: the wrappers must preserve current distinctions exactly (`nil,nil,nil` for empty/no-results cases vs explicit parse problems for malformed/non-usable payloads).
- **Overreaching into resolver adoption**: threading `Resolver` into `search.Service` now would add risk with little benefit because provider construction already gives a stable seam.

### Ready for Proposal
Yes — the slice is well-bounded and ready for proposal. The proposal should tell the user we can migrate search-family adoption onto explicit engine-native parser wrappers with no behavior change and without touching article-family parsing.

#### 5. Recommended file touch list
Safest production touch list:
- `internal/parse/engine/live_search.go` — new engine-native wrapper for live search.
- `internal/parse/engine/dpd_search.go` — new engine-native wrapper for DPD search.
- `internal/app/wiring.go` — switch search providers to `NewEnginePipelineProvider(...)` using engine-native wrappers.

Safest test touch list:
- `internal/parse/engine/search_port_test.go` or new dedicated wrapper tests — verify wrapper passthrough behavior for both concrete engine search parsers.
- `internal/search/provider_test.go` — verify engine-backed provider still preserves parsed records/warnings flow with a concrete engine parser, not only the generic adapter.
- `internal/app/wiring_test.go` — update assertions from `*parseengine.LegacySearchAdapter` to the new concrete engine-native parser wrapper types.

Files safest to leave untouched for this slice:
- `internal/parse/live_search.go`
- `internal/parse/dpd_search.go`
- `internal/search/service.go`
- live/DPD search normalizers and renderers

### No-behavior-drift test set
The minimum regression set that defines search parsing fidelity is:
- `internal/parse/live_search_test.go`
  - curated record extraction
  - empty payload returns empty result, no warning/error
  - broken markup returns `ProblemCodeDPDSearchParseFailed`
- `internal/parse/dpd_search_test.go`
  - JSON array decoding and first-pipe split behavior
  - invalid JSON and entirely unusable payloads return `ProblemCodeDPDSearchParseFailed`
- `internal/search/provider_test.go`
  - engine-backed provider still preserves parser document propagation, normalized candidates, and parse+normalize warnings
- `internal/search/live_service_test.go`
  - parse failure propagation unchanged
  - normalize failure propagation unchanged
- `internal/app/wiring_test.go`
  - runtime wiring now instantiates engine-native search parsers for both `search` and `dpd`

### Helper movement guidance
Low-risk to move now:
- only wrapper-specific construction code into `internal/parse/engine`
- optional wrapper passthrough helpers if they are brand-new and engine-local

Stay put for now:
- `reLiveSearchItem`, `reLiveSearchAnchor`, `reLiveSearchSnippet`, `reLiveSearchTags`
- `parseLiveSearchItem`
- `resolveLiveSearchURL`
- `normalizeLiveSearchText`
- `firstMatchGroup`
- `splitSearchRecord`
- `parse.ParsedSearchRecord`

Those items are tightly coupled to existing parser behavior/tests and moving them in the same slice adds churn without helping engine adoption.

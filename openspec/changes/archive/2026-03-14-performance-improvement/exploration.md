# Exploration: performance-improvement

## Current State

dlexa is a Go 1.22 CLI tool that looks up Spanish dictionary entries (primarily from the DPD — Diccionario Panhispánico de Dudas). The tool has **zero external runtime dependencies** (stdlib only, `go.sum` does not exist). The architecture is deliberately layered and explicit.

### Request Lifecycle (hotpath)

```
os.Args → platform.OSCLI
  → app.App.Run()
    → config.StaticLoader.Load()        (no I/O)
    → query.LookupService.Lookup()
        → cache.MemoryStore.Get()       (RWMutex read)
        → source.StaticRegistry.SourcesFor()  (slice scan)
        → [for each source, sequentially]:
            → source.PipelineSource.Lookup()
                → fetch.DPDFetcher.Fetch()    ← NETWORK I/O (blocking, 10s timeout)
                → parse.DPDArticleParser.Parse()  ← REGEX-HEAVY CPU
                → normalize.DPDNormalizer.Normalize()  ← STRING-HEAVY CPU
        → cache.MemoryStore.Set()       (RWMutex write)
    → render.MarkdownRenderer.Render()  ← STRING BUILD CPU
    → os.Stdout.Write()
```

### Key Measurements / Characterization

- **Process type**: CLI, spawned per invocation. Cache is in-process memory only — dies with process.
- **Parallelism**: NONE. Sources are queried sequentially in a `for` loop.
- **External I/O**: One blocking HTTP GET per source per lookup. Default sources: `dpd` only.
- **Caching**: In-process RWMutex map. Effective only for repeated calls within the same process lifetime (extremely rare for a CLI).

---

## Architecture Map

```
cmd/dlexa/main.go
  └── internal/app/
        ├── app.go          — Run() orchestration, flag parsing, format selection
        └── wiring.go       — Composition root (NEW sources/renderers added here ONLY)

internal/
  ├── model/types.go        — All domain types (LookupRequest, LookupResult, Entry, Article, ...)
  ├── config/               — StaticLoader, RuntimeConfig (DPD base URL, timeout, UA)
  ├── cache/                — Store interface + MemoryStore (sync.RWMutex, unbounded map)
  ├── query/                — LookupService.Lookup() — orchestrates cache + sources
  ├── source/               — PipelineSource (fetch→parse→normalize), StaticRegistry
  ├── fetch/                — DPDFetcher (net/http), StaticFetcher (in-memory stub)
  ├── parse/                — DPDArticleParser (18 compiled regexps), MarkdownParser
  ├── normalize/            — DPDNormalizer (7 compiled regexps + markdown builder)
  ├── render/               — MarkdownRenderer, JSONRenderer, StaticRegistry
  ├── platform/             — OSCLI (os.Args, Stdout, Stderr)
  ├── doctor/               — NoopDoctor
  └── version/              — version string
```

---

## Bottleneck Analysis

### Bottleneck 1: Sequential source fan-out (HIGH IMPACT)

**Location**: `internal/query/service.go:44`

```go
for _, item := range resolvedSources {
    sourceResult, err := item.Lookup(ctx, request)
    // ...
}
```

Sources are queried one after another. Each source may make a blocking network call (~10s timeout). If 2 sources are requested, total latency is `sum(source latencies)` instead of `max(source latencies)`.

Currently only `dpd` source is in default config, so this is a latency cliff waiting to happen when more sources are added. The fix is `errgroup`/`sync.WaitGroup` with goroutines per source.

**Impact**: Linear latency multiplier per source. Trivially parallelizable.

### Bottleneck 2: In-process memory cache is process-scoped — effectively a no-op for CLI (HIGH IMPACT)

**Location**: `internal/cache/memory.go`

The `MemoryStore` lives in the process. Every CLI invocation starts a fresh process. The cache can never provide a hit across invocations. This means **every invocation pays full network + parse + normalize cost**.

A filesystem cache (e.g. `~/.cache/dlexa/`) or even an OS-level cache would give dramatic improvement, especially since DPD entries change infrequently (editorial updates, not real-time data).

**Impact**: Every CLI call pays full latency. Cache exists but never hits in practice.

### Bottleneck 3: HTTP client is created fresh per DPDFetcher constructor (MEDIUM IMPACT)

**Location**: `internal/fetch/http.go:36`

```go
Client: &http.Client{
    Timeout: timeout,
},
```

`http.Client` is created once per DPDFetcher instance (wiring.go creates it once per app). The `http.Transport` inside it has a connection pool, but since each CLI invocation creates a new `http.Client`, there is **no TCP connection reuse** across invocations. For HTTPS this means a full TLS handshake on every call.

Within a single invocation this is fine (one request). If parallelism is added (multiple sources), the default transport pool will serve connections efficiently within the single process lifetime.

**Impact**: TLS handshake overhead on every invocation (~100-300ms). Unavoidable without persistent process or disk cache.

### Bottleneck 4: `io.ReadAll` loads entire response body into memory (LOW-MEDIUM)

**Location**: `internal/fetch/http.go:107`

```go
body, err := io.ReadAll(resp.Body)
```

DPD HTML pages can be 100-500 KB. The entire body is read into a `[]byte` before any parsing begins. For large responses, this is a full memory allocation. A streaming parser would reduce peak memory but add complexity — not worth it for a single-request CLI, but worth noting.

**Impact**: Peak memory allocation per request. Acceptable for CLI at current DPD page sizes.

### Bottleneck 5: `isChallengeBody` calls `strings.ToLower` on entire body (MEDIUM)

**Location**: `internal/fetch/http.go:163`

```go
func isChallengeBody(body []byte) bool {
    snippet := strings.ToLower(string(body))
    if len(snippet) > challengeBodySnippetLimit {
        snippet = snippet[:challengeBodySnippetLimit]
    }
    return strings.Contains(snippet, "cloudflare") && strings.Contains(snippet, "challenge")
}
```

The code converts the entire `[]byte` to `string` (allocation), then `strings.ToLower` (another allocation), then truncates. It should truncate the raw `body` slice FIRST, then convert. The `challengeBodySnippetLimit` is only 1024 bytes, so the full-body conversion and lower happens before truncation.

**Impact**: Two unnecessary allocations per request (one for the entire body). Simple fix.

### Bottleneck 6: `isChallengePage` in the parser duplicates the check with full-body string conversion (MEDIUM)

**Location**: `internal/parse/dpd.go:302`

```go
func isChallengePage(body string) bool {
    lower := strings.ToLower(body)
    return strings.Contains(lower, "cloudflare") && strings.Contains(lower, "challenge")
}
```

At parse time, `body` is already a full `string` (converted from `[]byte` on line 53: `body := string(document.Body)`). Then `strings.ToLower` allocates another full copy. For a large HTML page this is a significant allocation. `strings.EqualFold` won't help here, but `bytes.Contains` with lowercase literals, or checking only the first N bytes, would.

**Impact**: Full-body ToLower allocation at parse entry. Can use snippet approach like fetch layer.

### Bottleneck 7: `preserveSemanticSpans` does N string replacements in a loop (MEDIUM)

**Location**: `internal/parse/dpd.go:537`

```go
for _, tag := range parts {
    // ...
    raw = strings.ReplaceAll(raw, tag, "")
}
```

For each non-allowed tag found by `reTags.FindAllString`, a `strings.ReplaceAll` is called. For a section with many unrecognized tags, this is O(N*M) where N=unrecognized tags, M=body length. A single-pass builder would be O(M).

**Impact**: O(N*M) string manipulation per paragraph/cell during parse. Degrades on pages with many unrecognized HTML tags.

### Bottleneck 8: `renderInlineMarkdown` / `renderMarkdownInlines` duplicate logic between normalize and render layers (DESIGN / MEDIUM)

**Locations**: `internal/normalize/dpd.go:266` and `internal/render/markdown.go:231`

Both packages define `renderInlineMarkdown`, `renderMarkdownInlines`, `needsInlineSpace`, `shouldGlueInlineWordBoundary`, `shouldWrapStyledBuffer`, `lastInlineWordRune`, `firstInlineWordRune`. They are nearly identical. This is code duplication (violates DRY) and means bug fixes must be applied twice. It's not a runtime performance issue but affects maintainability.

**Impact**: Maintenance burden, potential for divergence bugs. Not a hotpath issue.

### Bottleneck 9: `renderTableMarkdown` / `renderTableHTML` are duplicated between normalize and render packages (DESIGN / MEDIUM)

Same pattern as above. Both `internal/normalize/dpd.go` and `internal/render/markdown.go` define these functions identically.

### Bottleneck 10: No context cancellation propagation in parse/normalize (LOW)

`parse.DPDArticleParser.Parse` and `normalize.DPDNormalizer.Normalize` both discard the context:

```go
_ = ctx
```

If the caller cancels (e.g. timeout), parse and normalize will run to completion anyway. For a CLI this is usually fine, but for future server-mode usage it means wasted CPU.

---

## Concurrency Analysis

| Component | Goroutines | Channels | Sync Primitives | Assessment |
|-----------|------------|----------|-----------------|------------|
| `query.LookupService.Lookup` | 0 (sequential) | 0 | none | **Gap: should fan-out** |
| `cache.MemoryStore` | 0 | 0 | `sync.RWMutex` | Correct but unused at CLI scale |
| `fetch.DPDFetcher.Fetch` | 0 | 0 | none | Synchronous HTTP, correct |
| `parse.DPDArticleParser` | 0 | 0 | none | Pure function, parallelizable |
| `normalize.DPDNormalizer` | 0 | 0 | none | Pure function, parallelizable |
| `render.*` | 0 | 0 | none | Pure function, correct |

**Summary**: The entire pipeline is single-threaded by design. The only concurrency primitive is the `sync.RWMutex` in `MemoryStore`, which is correct but would benefit from an actual cross-process cache making it unnecessary.

---

## I/O and Resource Usage

| Operation | Frequency | Cost | Notes |
|-----------|-----------|------|-------|
| HTTP GET to `rae.es/dpd/<term>` | 1 per invocation | ~200-800ms (network) + TLS | Dominant cost |
| `io.ReadAll(resp.Body)` | 1 per invocation | ~100-500KB alloc | Full body into memory |
| `strings.ToLower(entireBody)` | 2x per invocation | Duplicate full-body alloc | fetch + parse both do it |
| Regexp operations (18 patterns) | per section/paragraph | CPU + string alloc | Compiled at package init ✓ |
| `preserveSemanticSpans` tag loop | per paragraph | O(N*M) | Can be O(M) with builder |
| `cache.MemoryStore.Get/Set` | 2 per invocation | Negligible | RWMutex, never hits |

---

## Existing Test Coverage

### What is covered (good)

| Package | Test File | Coverage Type |
|---------|-----------|---------------|
| `app` | `app_test.go` | Integration: CLI flag parsing, lookup delegation, render output |
| `query` | `service_test.go`, `service_dpd_test.go` | Unit: cache hit/miss, aggregation, error classification |
| `source` | `pipeline_test.go` | Unit: fetch→parse→normalize ordering, warning propagation |
| `fetch` | `http_test.go` | Table-driven: 6 HTTP outcomes (200, 404, 403, timeout, network error, challenge) |
| `parse` | `dpd_test.go` | Table-driven + golden: article extraction, inline semantics, spans, challenge detection |
| `normalize` | `dpd_test.go` | Unit: markdown generation, inline rendering, table normalization, citation |
| `render` | `markdown_test.go`, `registry_test.go`, `semantic_terminal_test.go`, `dpd_integration_test.go` | Unit + golden: markdown/JSON output, terminal planning, inline formatting |

### What is NOT covered (gaps)

- **`cache.MemoryStore`**: No tests at all. The concurrent RWMutex behavior is untested.
- **`config.StaticLoader`**: No tests. The `DefaultRuntimeConfig()` values are not verified.
- **`normalize.IdentityNormalizer`**: No tests. Thin adapter but untested.
- **`fetch.StaticFetcher`**: No tests. Bootstrap fetcher untested.
- **`doctor.NoopDoctor`**: No tests.
- **`platform.OSCLI`**: No tests.
- **`version`**: No tests.
- **Concurrency**: No tests for concurrent cache access, no benchmark tests, no race detector coverage.
- **Benchmarks**: Zero `Benchmark*` functions exist anywhere. There is no baseline for any operation.
- **Integration under load**: No test exercises the full pipeline end-to-end with real network (network calls are always mocked).

### TDD gaps for performance work

- No `Benchmark*` for `extractInlines`, `preserveSemanticSpans`, `renderInlineMarkdown`, `DPDArticleParser.Parse`, `DPDNormalizer.Normalize`
- No race-detector test for `MemoryStore`
- No test for parallel source fan-out (doesn't exist yet)
- No test for filesystem cache (doesn't exist yet)

---

## Design Pattern Assessment

### Strengths (keep these)

- **Interface-driven boundaries**: Every layer depends on interfaces (`cache.Store`, `query.Service`, `source.Registry`, etc.). This makes testing and swapping implementations trivial.
- **Composition root in `wiring.go`**: Concrete types are only chosen once. No concrete dependency leaks into inner packages.
- **Package-level regexp compilation**: All 18+ regexp patterns are compiled at `var` declaration time (package init), not per-call. Correct.
- **`strings.Builder` usage**: Key rendering paths use `strings.Builder` correctly to avoid O(N²) string concatenation.
- **Error classification**: `ProblemError` wraps typed problems through the pipeline cleanly.

### Anti-patterns / Improvement Areas

1. **Sequential source loop** — should be goroutine fan-out with `errgroup`
2. **In-process-only cache** — needs filesystem persistence to be useful
3. **`strings.ToLower(fullBody)` twice** — pre-check should use raw byte prefix
4. **`preserveSemanticSpans` O(N*M) tag removal** — should use single-pass builder
5. **Duplicated render functions** — normalize and render packages share nearly identical inline rendering code; should be extracted to a shared `internal/renderutil` package
6. **No benchmarks** — performance work requires baseline measurements

---

## Improvement Opportunities (Prioritized)

### Priority 1: Persist cache to filesystem (highest ROI)

- **What**: Replace `MemoryStore` with a filesystem cache (e.g., `~/.cache/dlexa/<key>.json`) with TTL
- **Why**: Every CLI invocation currently pays full network + parse + normalize. DPD entries change rarely (editorial, not real-time). A 24h TTL would eliminate the HTTP call for repeated queries.
- **Effort**: Medium. New `cache.FilesystemStore` implementing `cache.Store`. No interface change needed.
- **Test approach (TDD)**: Write `TestFilesystemStoreGetMissesCold`, `TestFilesystemStoreGetHitsAfterSet`, `TestFilesystemStoreExpiredEntryIsMiss`, `TestFilesystemStoreConcurrentReadWrite` (race detector) before implementing.

### Priority 2: Parallel source fan-out (high ROI when >1 source)

- **What**: Replace sequential source loop in `query.LookupService.Lookup` with `errgroup`-based parallel dispatch
- **Why**: When multiple sources are queried, parallel dispatch reduces latency from sum to max
- **Effort**: Low. ~30 lines of change. Requires `golang.org/x/sync/errgroup` (first external dependency) OR manual `sync.WaitGroup` + channel (stdlib only).
- **Test approach (TDD)**: Write `TestLookupQueriesSourcesConcurrently` using `stubSource` with artificial delay, verify total time < sum of delays before implementing.

### Priority 3: Fix `isChallengeBody` / `isChallengePage` allocation (quick win)

- **What**: Truncate raw bytes BEFORE string conversion in both functions
- **Why**: Eliminates full-body `strings.ToLower` allocations (~100-500KB each)
- **Effort**: Very Low. 2 function changes.
- **Test approach (TDD)**: Existing tests cover the behavior; add a benchmark `BenchmarkIsChallengeBody` before fixing.

### Priority 4: Replace `preserveSemanticSpans` O(N*M) with single-pass builder (medium ROI)

- **What**: Rewrite `preserveSemanticSpans` to use a single-pass byte scanner
- **Why**: Current approach calls `strings.ReplaceAll(raw, tag, "")` in a loop — O(N*M)
- **Effort**: Medium. Requires careful HTML tag scanning.
- **Test approach (TDD)**: Existing parse tests cover semantic preservation; add `BenchmarkPreserveSemanticSpans` before fixing.

### Priority 5: Extract shared inline render utilities (DRY / maintainability)

- **What**: Extract `renderInlineMarkdown`, `needsInlineSpace`, `shouldGlue*`, `lastInlineWordRune`, `firstInlineWordRune`, `renderTableMarkdown`, `renderTableHTML` to `internal/renderutil`
- **Why**: The code is duplicated almost verbatim in `normalize/dpd.go` and `render/markdown.go`
- **Effort**: Medium. Requires updating imports in both packages. Risk of behavioral divergence if not done carefully.
- **Test approach (TDD)**: Move existing tests, add coverage for shared package before refactoring.

### Priority 6: Add benchmark suite (enabler for all other performance work)

- **What**: Add `Benchmark*` functions in `fetch`, `parse`, `normalize`, `render` packages
- **Why**: Without baselines, "improvements" cannot be verified. This is a prerequisite for TDD on performance.
- **Effort**: Low. Pure test code.

---

## Approaches Comparison

| Approach | Impact | Effort | Risk | Recommended |
|----------|--------|--------|------|-------------|
| Filesystem cache | Very High | Medium | Low | YES — Priority 1 |
| Parallel source fan-out | High (when >1 source) | Low | Low | YES — Priority 2 |
| Fix `isChallengeBody` allocation | Low | Very Low | Minimal | YES — quick win |
| `preserveSemanticSpans` rewrite | Medium | Medium | Medium | YES — after benchmarks |
| Extract shared renderutil | Low (perf), Medium (maintenance) | Medium | Medium | YES — long-term |
| Add benchmark suite | Enabler | Low | None | YES — prerequisite |

---

## Recommendation

Start with a **benchmark-first** approach to establish baselines (`go test -bench ./...`), then:

1. **Filesystem cache** — delivers the largest user-visible latency reduction (eliminates network I/O on cache hits)
2. **Parallel source fan-out** — architectural prerequisite for scaling to multiple sources
3. **Allocation fixes** — quick wins found during profiling
4. **Code extraction** — deferred until after functional improvements are validated

All changes must be TDD-first: write failing tests (unit + benchmark), then implement, then verify.

---

## Affected Files (Summary)

- `internal/cache/memory.go` — replace or augment with filesystem store
- `internal/cache/interfaces.go` — possibly extend Store interface with TTL or inspect methods
- `internal/query/service.go` — parallelize source loop
- `internal/fetch/http.go` — fix `isChallengeBody` truncation order
- `internal/parse/dpd.go` — fix `isChallengePage`, optimize `preserveSemanticSpans`
- `internal/normalize/dpd.go` — extract shared render functions
- `internal/render/markdown.go` — consume shared render functions
- `internal/app/wiring.go` — wire filesystem cache

---

## Risks

1. **Filesystem cache may introduce cross-platform path issues** (Windows vs Unix). `os.UserCacheDir()` handles this in stdlib.
2. **Parallel fan-out requires a first external dependency** if using `errgroup`, OR careful channel management for stdlib-only approach.
3. **Extracting shared render utilities** risks subtle behavioral divergence between normalize and render paths if the extraction is incomplete — extensive golden-file tests mitigate this.
4. **No benchmark baselines** means performance claims cannot be objectively validated. Adding benchmarks is a prerequisite, not optional.
5. **The in-process cache correctly handles concurrent access** (RWMutex) but is untested. Parallel fan-out will trigger concurrent cache writes for the first time — must add race-detector tests before implementing.

---

## Ready for Proposal

Yes. The bottlenecks are clearly identified, the affected files are known, and the TDD approach is defined. The proposal should scope the work into phases: (1) benchmark suite, (2) filesystem cache, (3) parallel fan-out, (4) allocation fixes.

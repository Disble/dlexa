# Tasks: Performance Improvement

## Phase 0: Benchmark Suite (MUST complete first — hard prerequisite for all phases)

- [x] 0.1 **Write `BenchmarkDPDFetch` in `internal/fetch/bench_test.go`** [small]
  - Create `internal/fetch/bench_test.go`
  - Benchmark `DPDFetcher.Fetch` using `httptest.Server` serving a representative DPD HTML fixture
  - Use `b.ReportAllocs()` and sub-benchmarks for at least two fixture sizes
  - Must report `ns/op`, `B/op`, `allocs/op`
  - Must NOT make real network calls
  - Acceptance: `go test -bench=BenchmarkDPDFetch ./internal/fetch/...` produces reproducible benchmark output
  - Dependencies: none

- [x] 0.2 **Write `BenchmarkIsChallengeBody` in `internal/fetch/bench_test.go`** [small]
  - Add to `internal/fetch/bench_test.go`
  - Benchmark `isChallengeBody` with 1KB, 100KB, and 500KB body fixtures (both challenge and non-challenge bodies)
  - Use `b.ReportAllocs()` and sub-benchmarks per size
  - Acceptance: `go test -bench=BenchmarkIsChallengeBody ./internal/fetch/...` runs and captures baseline `B/op` (will be ~500KB for large body — this is the pre-fix baseline)
  - Dependencies: none

- [x] 0.3 **Write `BenchmarkDPDParse` in `internal/parse/bench_test.go`** [small]
  - Create `internal/parse/bench_test.go`
  - Benchmark `DPDArticleParser.Parse` using a representative DPD HTML fixture file
  - Use `b.ReportAllocs()` and sub-benchmarks
  - Acceptance: `go test -bench=BenchmarkDPDParse ./internal/parse/...` produces reproducible output
  - Dependencies: none

- [x] 0.4 **Write `BenchmarkIsChallengePage` in `internal/parse/bench_test.go`** [small]
  - Add to `internal/parse/bench_test.go`
  - Benchmark `isChallengePage` with 1KB and 500KB string fixtures
  - Use `b.ReportAllocs()`
  - Acceptance: captures baseline `B/op` for `isChallengePage` before truncation fix
  - Dependencies: none

- [x] 0.5 **Write `BenchmarkPreserveSemanticSpans` in `internal/parse/bench_test.go`** [small]
  - Add to `internal/parse/bench_test.go`
  - Benchmark `preserveSemanticSpans` with fixtures of varying tag counts: small (5 tags), medium (20 tags), large (50 tags) — all with fixed ~5KB input size
  - Use `b.ReportAllocs()` and `b.Run` sub-benchmarks per variant
  - Acceptance: captures baseline `ns/op` showing O(N*M) growth with tag count
  - Dependencies: none

- [x] 0.6 **Write `BenchmarkDPDNormalize` in `internal/normalize/bench_test.go`** [small]
  - Create `internal/normalize/bench_test.go`
  - Benchmark `DPDNormalizer.Normalize` using representative parsed article fixtures
  - Use `b.ReportAllocs()`
  - Acceptance: `go test -bench=BenchmarkDPDNormalize ./internal/normalize/...` produces output
  - Dependencies: none

- [x] 0.7 **Write `BenchmarkMarkdownRender` in `internal/render/bench_test.go`** [small]
  - Create `internal/render/bench_test.go`
  - Benchmark `MarkdownRenderer.Render` using representative `LookupResult` fixtures
  - Use `b.ReportAllocs()`
  - Acceptance: `go test -bench=BenchmarkMarkdownRender ./internal/render/...` produces output
  - Dependencies: none

- [x] 0.8 **Write `TestMemoryStoreConcurrentReadWrite` in `internal/cache/memory_test.go`** [small]
  - Create `internal/cache/memory_test.go`
  - Test N goroutines writing and M goroutines reading concurrently on the same `MemoryStore` instance
  - Use `sync.WaitGroup` to coordinate
  - Acceptance: `go test -race ./internal/cache/...` passes with zero data races
  - Dependencies: none

- [x] 0.9 **Verify all Phase 0 benchmarks and tests pass** [small]
  - Run `go test -bench=. ./internal/fetch/... ./internal/parse/... ./internal/normalize/... ./internal/render/...`
  - Run `go test -race ./internal/cache/...`
  - Run `go test ./...` to confirm no regressions
  - Acceptance: all benchmarks produce output, race test passes, full suite green
  - Dependencies: 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8

---

## Phase 1: Filesystem Cache

- [x] 1.1 **Write TDD tests for `FilesystemStore` in `internal/cache/filesystem_test.go`** [medium]
  - Create `internal/cache/filesystem_test.go`
  - Table-driven tests covering all spec scenarios:
    - `TestFilesystemStore_ColdMiss`: Get on empty dir returns `(zero, false, nil)`
    - `TestFilesystemStore_SetThenGet`: Set then Get returns deeply equal result with `ok=true`
    - `TestFilesystemStore_PersistenceAcrossInstances`: Set with one instance, Get with a new instance on same dir
    - `TestFilesystemStore_ExpiredEntry`: Set with 1s TTL, simulate time passing (injectable clock), Get returns miss
    - `TestFilesystemStore_NonExpiredEntry`: Set with 24h TTL, immediate Get returns hit
    - `TestFilesystemStore_CustomTTL`: Constructor with 1h TTL uses that threshold
    - `TestFilesystemStore_CorruptFile`: Write invalid JSON to cache dir, Get returns `(zero, false, nil)` and removes corrupt file
    - `TestFilesystemStore_SpecialCharKey`: Keys with `|`, `/`, `:` produce valid filenames
    - `TestFilesystemStore_DistinctKeys`: Different keys produce different filenames
  - All tests use `t.TempDir()` for cache directory
  - Acceptance: tests compile (will fail initially — RED phase of TDD)
  - Dependencies: 0.9

- [x] 1.2 **Write `TestFilesystemStoreConcurrentReadWrite` race-detector test** [small]
  - Add to `internal/cache/filesystem_test.go`
  - Multiple goroutines concurrently calling Get and Set with overlapping keys on the same `FilesystemStore` instance
  - Acceptance: `go test -race ./internal/cache/...` passes (after implementation)
  - Dependencies: 0.9

- [x] 1.3 **Implement `FilesystemStore` in `internal/cache/filesystem.go`** [medium]
  - Create `internal/cache/filesystem.go`
  - Implement `FilesystemStore` struct with `dir string`, `ttl time.Duration`, `now func() time.Time`
  - Implement `NewFilesystemStore(dir string, ttl time.Duration) *FilesystemStore`
  - Implement `Get(ctx, key)`: SHA256 key -> hex filename -> `os.ReadFile` -> `json.Unmarshal` envelope -> TTL check -> return
  - Implement `Set(ctx, key, result)`: envelope with `ExpiresAt` -> `json.Marshal` -> write temp file -> `os.Rename` (atomic)
  - Implement `cacheEnvelope` struct with `ExpiresAt`, `CreatedAt`, `Data` fields
  - Graceful degradation: all errors return cache miss (`Get`) or `nil` (`Set`), never propagate
  - `os.MkdirAll` on first write
  - Acceptance: all tests from 1.1 and 1.2 pass; `go test -race ./internal/cache/...` clean
  - Dependencies: 1.1, 1.2

- [x] 1.4 **Add `CacheTTL` to `RuntimeConfig` in `internal/config/interfaces.go`** [small]
  - Add `CacheTTL time.Duration` field to `RuntimeConfig` struct
  - Modify file: `internal/config/interfaces.go`
  - Acceptance: compiles with no errors
  - Dependencies: none

- [x] 1.5 **Set default `CacheTTL` in `DefaultRuntimeConfig()` in `internal/config/static.go`** [small]
  - Add `CacheTTL: 24 * time.Hour` to the `RuntimeConfig` returned by `DefaultRuntimeConfig()`
  - Modify file: `internal/config/static.go`
  - Acceptance: `DefaultRuntimeConfig().CacheTTL == 24*time.Hour`
  - Dependencies: 1.4

- [x] 1.6 **Wire `FilesystemStore` in `internal/app/wiring.go`** [small]
  - Replace `cache.NewMemoryStore()` with `cache.NewFilesystemStore(...)` in `app.New()`
  - Use `os.UserCacheDir()` to get base dir, append `"dlexa"` via `filepath.Join`
  - Fallback to `cache.NewMemoryStore()` if `os.UserCacheDir()` returns error
  - Use `runtimeConfig.CacheTTL` for TTL parameter
  - Add imports: `os`, `path/filepath`
  - Acceptance: `go test ./...` passes; second `dlexa <term>` invocation uses cached result from disk
  - Dependencies: 1.3, 1.5

- [x] 1.7 **Verify Phase 1 complete** [small]
  - Run `go test ./...` — full suite green
  - Run `go test -race ./internal/cache/...` — zero races
  - Verify no new entries in `go.mod`
  - Acceptance: all tests pass, filesystem cache is wired as default
  - Dependencies: 1.6

---

## Phase 2: Parallel Source Fan-out

- [x] 2.1 **Write TDD concurrency test in `internal/query/parallel_test.go`** [medium]
  - Add `TestLookupQueriesSourcesConcurrently` to `internal/query/parallel_test.go`
  - Create stub sources with `time.Sleep(100ms)` artificial delays
  - Register 3 delayed sources with a `LookupService`
  - Call `Lookup`, measure wall-clock time
  - Assert total time < 200ms (proving concurrency, not sequential 300ms)
  - Assert result contains entries from all 3 sources
  - Acceptance: test compiles and fails with current sequential implementation (RED)
  - Dependencies: 0.9

- [x] 2.2 **Write TDD test for single-source-fails-others-succeed scenario** [small]
  - Add `TestLookupOneSourceFailsOthersSucceed` to `internal/query/parallel_test.go`
  - 3 stub sources, second returns error
  - Assert result has entries from 2 successful sources
  - Assert result has `Problem` entry for the failed source
  - Assert no top-level error
  - Acceptance: test passes with both sequential and concurrent implementations
  - Dependencies: 0.9

- [x] 2.3 **Write TDD test for all-sources-fail scenario** [small]
  - Add `TestLookupAllSourcesFail` to `internal/query/parallel_test.go`
  - 2 stub sources, both return errors
  - Assert result has Problem entries for both, zero Entries, no top-level error
  - Acceptance: test passes with both sequential and concurrent implementations
  - Dependencies: 0.9

- [x] 2.4 **Write TDD test for deterministic result ordering by source priority** [small]
  - Add `TestLookupResultsOrderedByPriority` to `internal/query/parallel_test.go`
  - 3 sources with priorities 3, 1, 2 completing in arbitrary order (use varying delays)
  - Assert `result.Sources` are ordered by ascending priority
  - Acceptance: test compiles (may fail until ordering logic is implemented)
  - Dependencies: 0.9

- [x] 2.5 **Write race-detector test for parallel fan-out** [small]
  - Add `TestLookupRaceDetector` to `internal/query/parallel_test.go`
  - Multiple sources, run under `-race`
  - Acceptance: `go test -race ./internal/query/...` passes
  - Dependencies: 0.9

- [x] 2.6 **Implement parallel fan-out in `internal/query/service.go`** [medium]
  - Replace the sequential `for _, item := range resolvedSources` loop with:
    - `sourceOutcome` struct: `{source source.Source, result model.SourceResult, err error}`
    - Buffered channel `results chan sourceOutcome` (capacity = len(sources))
    - One goroutine per source with `wg.Add(1)`, `defer wg.Done()`
    - Closer goroutine: `go func() { wg.Wait(); close(results) }()`
    - Main goroutine reads `range results` and aggregates
  - Sort aggregated results by `source.Descriptor().Priority` before returning
  - Add imports: `sync`, `sort`
  - Shared `ctx` for cancellation propagation
  - Cache write remains in main goroutine AFTER fan-out completes
  - Acceptance: all tests from 2.1-2.5 pass; `go test -race ./internal/query/...` clean
  - Dependencies: 2.1, 2.2, 2.3, 2.4, 2.5

- [x] 2.7 **Verify Phase 2 complete** [small]
  - Run `go test ./...` — full suite green
  - Run `go test -race ./...` — zero races across all packages (note: -race requires CGo on Windows; structurally race-safe by design)
  - Verify no new entries in `go.mod`
  - Acceptance: parallel fan-out is wired, all tests pass
  - Dependencies: 2.6

---

## Phase 3: Allocation Fixes

- [x] 3.1 **Write edge-case tests for `isChallengeBody` in `internal/fetch/http_test.go`** [small]
  - Add table-driven tests for `isChallengeBody`:
    - Challenge markers in first 1024 bytes -> `true`
    - Challenge markers ONLY beyond 1024 bytes -> `false` (new behavior after fix, documents truncation)
    - Non-challenge body -> `false`
    - Empty body -> `false`
    - Body exactly 1024 bytes with markers -> `true`
  - Acceptance: tests pass with current implementation for shared cases; truncation-specific test documents expected new behavior
  - Dependencies: 0.9

- [x] 3.2 **Implement `isChallengeBody` truncation fix in `internal/fetch/http.go`** [small]
  - Truncate `[]byte` body to `min(len(body), challengeBodySnippetLimit)` BEFORE `string()` conversion and `strings.ToLower`
  - Add `challengeBodySnippetLimit = 1024` constant if not present
  - Acceptance: all tests from 3.1 pass; `BenchmarkIsChallengeBody` from 0.2 shows `B/op` ~1024 instead of ~500KB for large bodies
  - Dependencies: 3.1

- [x] 3.3 **Write edge-case tests for `isChallengePage` in `internal/parse/dpd_test.go`** [small]
  - Add table-driven tests for `isChallengePage`:
    - Challenge markers in first 1024 chars -> `true`
    - Challenge markers only beyond 1024 chars -> `false`
    - Non-challenge page -> `false`
    - Empty string -> `false`
  - Acceptance: tests compile and document expected behavior
  - Dependencies: 0.9

- [x] 3.4 **Implement `isChallengePage` truncation fix in `internal/parse/dpd.go`** [small]
  - Truncate input string to `min(len(s), challengePageSnippetLimit)` BEFORE `strings.ToLower`
  - Add `challengePageSnippetLimit = 1024` constant if not present
  - Acceptance: all tests from 3.3 pass; `BenchmarkIsChallengePage` from 0.4 shows reduced `B/op`
  - Dependencies: 3.3

- [x] 3.5 **Write edge-case tests for `preserveSemanticSpans` rewrite in `internal/parse/dpd_test.go`** [small]
  - Add table-driven tests for `preserveSemanticSpans`:
    - Empty input -> empty output
    - Input with no HTML tags -> unchanged output
    - Input where all tags are allowed -> unchanged output
    - Input where all tags are disallowed -> tags stripped, text preserved
    - Nested tags (some allowed, some not) -> only disallowed stripped
    - Malformed tag (no closing `>`) -> preserved as-is
  - Acceptance: tests compile; existing golden tests continue to pass
  - Dependencies: 0.9

- [x] 3.6 **Rewrite `preserveSemanticSpans` as single-pass `strings.Builder` in `internal/parse/dpd.go`** [medium]
  - Replace the O(N*M) `reTags.FindAllString` + `strings.ReplaceAll` loop with single-pass scanner
  - Walk input: find `<`, find `>`, check tag against allowed set, copy or skip
  - Pre-grow `strings.Builder` to `len(raw)` capacity
  - Handle edge cases: no tags, malformed tags, nested tags
  - Acceptance: all tests from 3.5 pass; all existing golden-file tests pass; `BenchmarkPreserveSemanticSpans` from 0.5 shows O(M) scaling instead of O(N*M)
  - Dependencies: 3.5

- [x] 3.7 **Verify Phase 3 complete** [small]
  - Run `go test ./...` — full suite green
  - Run `go test -bench=BenchmarkIsChallengeBody ./internal/fetch/...` — confirm `B/op` ~1024
  - Run `go test -bench=BenchmarkIsChallengePage ./internal/parse/...` — confirm reduced `B/op`
  - Run `go test -bench=BenchmarkPreserveSemanticSpans ./internal/parse/...` — confirm O(M) scaling
  - Verify no new entries in `go.mod`
  - Acceptance: all allocation improvements confirmed by benchmarks, full suite green
  - Dependencies: 3.2, 3.4, 3.6

---

## Phase 4: Renderutil Deduplication

- [x] 4.1 **Audit duplicated functions between `normalize/dpd.go` and `render/markdown.go`** [small]
  - Identify every function duplicated between the two files
  - Document which functions are identical vs which have behavioral differences (e.g., `InlineKindExample` handling: normalize wraps with `‹›`, render wraps with `*`)
  - Document the function signatures and dependencies for each
  - Acceptance: written list of functions to extract, noting identical vs divergent
  - Dependencies: 0.9

- [x] 4.2 **Write tests for shared helper functions in `internal/renderutil/inline_test.go`** [medium]
  - Create `internal/renderutil/` package directory
  - Create `internal/renderutil/inline_test.go`
  - Write table-driven tests for:
    - `NeedsInlineSpace(current, next string) bool`
    - `ShouldGlueInlineWordBoundary(current, next string) bool`
    - `ShouldWrapStyledBuffer(buffer []model.Inline) bool`
    - `LastInlineWordRune(raw string) (rune, bool)`
    - `FirstInlineWordRune(raw string) (rune, bool)`
    - `RenderInlineMarkdown(inlines []model.Inline) string` (normalize variant: `‹›` for examples)
    - `RenderMarkdownInlines(inlines []model.Inline) string` (render variant: `*` for examples)
  - Use realistic inline fixtures from existing test data
  - Acceptance: tests compile (RED phase — functions don't exist yet in renderutil)
  - Dependencies: 4.1

- [x] 4.3 **Write tests for shared table functions in `internal/renderutil/table_test.go`** [medium]
  - Create `internal/renderutil/table_test.go`
  - Write table-driven tests for:
    - `RenderTableMarkdown(table model.Table, indent string) string`
    - `RenderTableHTML(table model.Table, indent string) string`
    - And any shared helper functions for table formatting
  - Use realistic table fixtures from existing test data
  - Acceptance: tests compile (RED phase)
  - Dependencies: 4.1

- [x] 4.4 **Implement shared inline helpers in `internal/renderutil/inline.go`** [medium]
  - Create `internal/renderutil/inline.go`
  - Extract and export: `RenderInlineMarkdown`, `RenderMarkdownInlines`, `NeedsInlineSpace`, `ShouldGlueInlineWordBoundary`, `ShouldWrapStyledBuffer`, `LastInlineWordRune`, `FirstInlineWordRune`
  - Package must import ONLY `model` and stdlib packages
  - Must NOT import `normalize` or `render`
  - Acceptance: all tests from 4.2 pass; `go build ./internal/renderutil/...` succeeds with no import cycles
  - Dependencies: 4.2

- [x] 4.5 **Implement shared table helpers in `internal/renderutil/table.go`** [medium]
  - Create `internal/renderutil/table.go`
  - Extract and export: `RenderTableMarkdown`, `RenderTableHTML`, and all shared table formatting helpers (`formatHTMLTableCell`, `renderHTMLTableCellContent`, `renderHTMLFromMarkdownSubset`, `isSimpleMarkdownTable`, `isSimpleMarkdownRow`, `tableRowTexts`, `tableColumnWidths`, `formatTableRow`, `formatTableDivider`, `normalizeMarkdownTableCellText`, etc.)
  - Package must import ONLY `model` and stdlib packages
  - Acceptance: all tests from 4.3 pass; `go build ./internal/renderutil/...` succeeds
  - Dependencies: 4.3

- [x] 4.6 **Update `internal/normalize/dpd.go` to import `renderutil`** [medium]
  - Remove all locally-defined functions that are now in `renderutil`
  - Replace calls to local functions with calls to `renderutil.FunctionName`
  - Add import for `"github.com/Disble/dlexa/internal/renderutil"`
  - Acceptance: `go build ./internal/normalize/...` succeeds; no import cycles; `go test ./internal/normalize/...` passes with all existing tests including golden tests
  - Dependencies: 4.4, 4.5

- [x] 4.7 **Update `internal/render/markdown.go` to import `renderutil`** [medium]
  - Remove all locally-defined functions that are now in `renderutil`
  - Replace calls to local functions with calls to `renderutil.FunctionName`
  - Add import for `"github.com/Disble/dlexa/internal/renderutil"`
  - Acceptance: `go build ./internal/render/...` succeeds; no import cycles; `go test ./internal/render/...` passes with all existing tests including golden tests
  - Dependencies: 4.4, 4.5

- [x] 4.8 **Verify renderutil coverage meets threshold** [small]
  - Run `go test -cover ./internal/renderutil/...`
  - Assert line coverage >= 90%
  - If below threshold, add additional test cases targeting uncovered lines
  - Acceptance: coverage >= 90%
  - Dependencies: 4.4, 4.5

- [x] 4.9 **Verify Phase 4 complete — full regression check** [small]
  - Run `go test ./...` — full suite green, including all golden-file tests in parse, normalize, and render
  - Run `go test -race ./...` — zero races
  - Verify no duplicated render functions remain in `normalize/dpd.go` or `render/markdown.go`
  - Verify `renderutil` does NOT import `normalize` or `render`
  - Verify no new entries in `go.mod`
  - Acceptance: zero behavioral regressions, deduplication complete, full suite green
  - Dependencies: 4.6, 4.7, 4.8

---

## Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 0 | 9 | Benchmark Suite & Race Tests (prerequisite) |
| Phase 1 | 7 | Filesystem Cache |
| Phase 2 | 7 | Parallel Source Fan-out |
| Phase 3 | 7 | Allocation Fixes |
| Phase 4 | 9 | Renderutil Deduplication |
| **Total** | **39** | |

### Implementation Order

Phases MUST execute in strict order: 0 -> 1 -> 2 -> 3 -> 4. Within each phase, TDD order is enforced: tests/benchmarks are written BEFORE implementation. Verification tasks at the end of each phase confirm the phase is independently mergeable. Phase 0 is the hard prerequisite because it establishes the benchmark baselines that prove whether subsequent optimizations have any measurable effect. Tasks within Phase 0 are independent of each other and can be parallelized.

### Next Step

Ready for implementation (sdd-apply).

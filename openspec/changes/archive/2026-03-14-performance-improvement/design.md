# Design: Performance Improvement

## Technical Approach

This design covers four phases of performance improvement for the dlexa CLI, following TDD-first, stdlib-only discipline. The dominant cost is repeated network I/O on every CLI invocation (the in-process `MemoryStore` dies with the process). The approach eliminates this via a filesystem cache, then unlocks multi-source scalability with parallel fan-out, corrects allocation inefficiencies in hot-path functions, and deduplicates shared rendering code into a new `internal/renderutil` package.

All changes operate behind existing interfaces (`cache.Store`, `source.Source`, `source.Registry`). No new external dependencies are introduced. Each phase is independently mergeable and reversible.

## Architecture Decisions

### Decision: JSON for Filesystem Cache Storage Format

**Choice**: JSON files with embedded TTL metadata, one file per cache key.
**Alternatives considered**: (1) `encoding/gob` — faster serialization, but produces opaque binary files that cannot be inspected or debugged by users. (2) BoltDB/SQLite — adds external dependencies, violating the stdlib-only constraint. (3) Single-file append-only log — complicates TTL expiry and concurrent access.
**Rationale**: JSON is human-readable (users can inspect `~/.cache/dlexa/` to debug), stdlib-native (`encoding/json`), and the cache payload (`model.LookupResult`) is already JSON-serializable (model types carry `json` struct tags). The serialization cost is negligible compared to the network I/O it eliminates. Each entry is a separate file, so concurrent reads/writes to different keys do not contend.

### Decision: Per-Entry Metadata File with Embedded TTL

**Choice**: Each cache file is a JSON envelope containing `ExpiresAt`, `CreatedAt`, and `Data` (the `LookupResult`). TTL is checked at read time by comparing `ExpiresAt` against `time.Now()`.

```go
type cacheEnvelope struct {
    ExpiresAt time.Time          `json:"expires_at"`
    CreatedAt time.Time          `json:"created_at"`
    Data      model.LookupResult `json:"data"`
}
```

**Alternatives considered**: (1) Filename-based TTL (encode expiry in filename) — makes key lookup more complex, requires directory listing instead of direct file access. (2) Separate metadata sidecar files — doubles the number of filesystem operations per cache hit.
**Rationale**: Single file per entry = one `os.ReadFile` + one `json.Unmarshal` per cache hit. The envelope is simple, self-contained, and the TTL check is a single `time.Time` comparison. No directory listing required — just hash the key to a filename and try to read it.

### Decision: SHA256-Based Cache Key to Filename Mapping

**Choice**: Cache key (from `cache.BuildKey()`) is hashed with `sha256.Sum256`, hex-encoded, and used as filename: `{hash}.json`.
**Alternatives considered**: (1) URL-encode the key directly as filename — fails on Windows with long paths, special characters, and case-insensitive filesystem collisions. (2) Base64-encode — contains `/` and `+` characters that are invalid in filenames on some systems.
**Rationale**: SHA256 hex produces a fixed-length (64-char), filesystem-safe filename. Collisions are cryptographically improbable. The hash is deterministic from the same `BuildKey()` output. No platform-specific escaping needed.

### Decision: No File Locking — Atomic Write via Rename

**Choice**: Use atomic write (write to temp file in same directory, then `os.Rename`) instead of file locks. Reads that encounter a partial/corrupt file treat it as a cache miss.
**Alternatives considered**: (1) `flock`/`fcntl` advisory locks — not portable across Windows and Unix without build tags or external libraries. (2) `sync.Mutex` — only protects within the same process, not across concurrent CLI invocations.
**Rationale**: The dlexa CLI is a short-lived process. Two concurrent invocations writing the same cache key is a rare edge case. Atomic rename guarantees that a reader either sees the old complete file or the new complete file, never a partial write. If `os.Rename` fails (different filesystem, Windows lock), the write is silently dropped — graceful degradation. A corrupted or truncated JSON file fails `json.Unmarshal`, which is treated as a cache miss, not an error.

### Decision: Cache Directory Under `os.UserCacheDir()`

**Choice**: `os.UserCacheDir()` + `/dlexa/` as the cache root. On first write, `os.MkdirAll` creates the directory.
**Alternatives considered**: (1) Hardcoded `~/.cache/dlexa` — breaks Windows (`%LocalAppData%`) and macOS (`~/Library/Caches`). (2) XDG_CACHE_HOME — stdlib `os.UserCacheDir()` already respects this on Linux.
**Rationale**: `os.UserCacheDir()` is the stdlib cross-platform solution. It returns `%LocalAppData%` on Windows, `~/Library/Caches` on macOS, and `$XDG_CACHE_HOME` (defaulting to `~/.cache`) on Linux. The `FilesystemStore` constructor accepts an override `dir` parameter for testing (via `t.TempDir()`).

### Decision: Default TTL of 24 Hours, Configurable via RuntimeConfig

**Choice**: 24-hour default TTL. The `FilesystemStore` constructor accepts a `ttl time.Duration` parameter. Wiring reads from `RuntimeConfig.CacheTTL` (new field, defaulting to 24h).
**Alternatives considered**: (1) No TTL / infinite cache — entries would go stale when RAE updates articles. (2) Short TTL (1h) — forces unnecessary re-fetches for a CLI tool used a few times per day.
**Rationale**: DPD entries are editorial content updated infrequently (monthly or less). A 24h TTL balances freshness against avoiding network I/O. Users can force a fresh fetch with the existing `NoCache` flag on `LookupRequest`.

### Decision: Graceful Degradation on All Cache Errors

**Choice**: Every cache operation (read, write, JSON parse, filesystem error) returns a cache miss or silently drops the write — never returns an error to the caller. The `FilesystemStore` methods log nothing and return `(zero, false, nil)` on Get failures, and `nil` on Set failures.
**Alternatives considered**: (1) Propagate errors — causes the CLI to fail on disk permission issues, which is worse than a cache miss. (2) Log warnings to stderr — adds noise for a CLI tool where stderr is the user's terminal.
**Rationale**: The cache is a performance optimization, not a correctness requirement. The system MUST work identically with or without a functioning cache. `query.LookupService.Lookup` already handles cache errors by checking `err != nil`, but the current `MemoryStore` never errors. The `FilesystemStore` will maintain this contract: errors are swallowed internally, callers see clean miss/hit semantics.

### Decision: `sync.WaitGroup` + Buffered Channel for Parallel Fan-out

**Choice**: Replace the sequential `for` loop in `query.LookupService.Lookup` with:
1. A buffered channel `results chan sourceOutcome` (capacity = len(sources))
2. One goroutine per source, each calling `source.Lookup(ctx, request)` and sending the outcome (result + error) on the channel
3. Main goroutine reads from channel until all sources complete
4. `sync.WaitGroup` to track goroutine lifecycle and close the channel

```go
type sourceOutcome struct {
    source source.Source
    result model.SourceResult
    err    error
}
```

**Alternatives considered**: (1) `golang.org/x/sync/errgroup` — cleaner API but introduces the first external dependency. (2) Sequential loop with `context.WithTimeout` per source — doesn't achieve parallelism. (3) `select` with per-source channels — more complex, same effect.
**Rationale**: stdlib-only constraint. `sync.WaitGroup` + buffered channel is the canonical Go pattern for fan-out/fan-in. The buffered channel (capacity = source count) ensures no goroutine blocks on send. The `WaitGroup` guarantees the channel is closed after all goroutines complete, so the main goroutine's `range` loop terminates cleanly.

### Decision: Shared `context.Context` for Cancellation

**Choice**: All goroutines in the fan-out share the same `ctx` from `Lookup()`. If the caller cancels, all in-flight HTTP requests are cancelled via the context passed through `source.Lookup → fetch.DPDFetcher.Fetch → http.NewRequestWithContext`.
**Alternatives considered**: (1) Per-source derived contexts with individual timeouts — over-engineering for a CLI with a single 10s timeout. (2) No cancellation — goroutines would leak if the CLI process is interrupted.
**Rationale**: The existing `context.Context` already propagates through the pipeline. Goroutines naturally respect it because `http.NewRequestWithContext` already uses it. No additional cancellation logic is needed.

### Decision: Cache Writes After Fan-out Are Not Locked

**Choice**: After fan-out completes, the aggregated result is written to cache via `s.cache.Set()` from the main goroutine. No mutex is needed because the cache write happens after all goroutines have completed (guaranteed by `WaitGroup`).
**Alternatives considered**: (1) Per-source cache writes from goroutines — would require concurrent cache access, which the `FilesystemStore` supports (different keys = different files) but adds complexity for no benefit since we cache the aggregated result, not individual source results.
**Rationale**: The cache stores the full `LookupResult` (all sources aggregated), not individual source results. There is exactly one cache write per lookup, after all sources are done. No concurrency on the cache write path.

### Decision: Byte-Level Truncation Before String Conversion for `isChallengeBody`

**Choice**: Truncate the raw `[]byte` to `min(len(body), challengeBodySnippetLimit)` BEFORE converting to string and calling `strings.ToLower`.

```go
func isChallengeBody(body []byte) bool {
    limit := len(body)
    if limit > challengeBodySnippetLimit {
        limit = challengeBodySnippetLimit
    }
    snippet := strings.ToLower(string(body[:limit]))
    return strings.Contains(snippet, "cloudflare") && strings.Contains(snippet, "challenge")
}
```

**Alternatives considered**: (1) `bytes.Contains` with `bytes.ToLower` — avoids the string conversion entirely but `bytes.ToLower` still allocates. (2) Manual byte-level case-insensitive search — complex, error-prone for multi-byte UTF-8.
**Rationale**: The current code converts the entire body (100-500KB) to string, then to lowercase, then truncates to 1024 bytes. Truncating first reduces the allocation from ~500KB to 1KB. The string conversion of the truncated slice is unavoidable (need `strings.Contains`) but now operates on 1KB instead of 500KB.

### Decision: Same Truncation Strategy for `isChallengePage` in parse

**Choice**: Apply the same snippet truncation to `parse.isChallengePage`. Since the input is already a `string`, truncate via slice before `strings.ToLower`.

```go
func isChallengePage(body string) bool {
    snippet := body
    if len(snippet) > challengePageSnippetLimit {
        snippet = snippet[:challengePageSnippetLimit]
    }
    lower := strings.ToLower(snippet)
    return strings.Contains(lower, "cloudflare") && strings.Contains(lower, "challenge")
}
```

**Rationale**: Same logic as `isChallengeBody`. The challenge markers ("cloudflare", "challenge") appear in the `<head>` section of Cloudflare challenge pages, well within the first 1024 bytes. Truncating before `ToLower` eliminates a full-body allocation.

### Decision: Single-Pass `strings.Builder` for `preserveSemanticSpans`

**Choice**: Replace the O(N*M) loop of `strings.ReplaceAll` calls with a single-pass scan using `strings.Builder`. The algorithm walks the input string, identifies HTML tags via `<` / `>` boundaries, checks each tag against the allowed set and skip prefixes, and either copies the tag to the builder (if allowed) or skips it.

```
Algorithm:
  pos = 0
  while pos < len(raw):
    nextTag = index of '<' from pos
    if nextTag == -1:
      builder.WriteString(raw[pos:])  // remaining text
      break
    builder.WriteString(raw[pos:nextTag])  // text before tag
    endTag = index of '>' from nextTag
    if endTag == -1:
      builder.WriteString(raw[nextTag:])  // malformed, keep as-is
      break
    tag = raw[nextTag:endTag+1]
    lower = strings.ToLower(tag)
    if isAllowed(lower) || isSkipPrefix(lower):
      builder.WriteString(tag)
    // else: drop the tag
    pos = endTag + 1
```

**Alternatives considered**: (1) Compiled regexp with `ReplaceAllStringFunc` — still O(N*M) if multiple passes are needed. (2) `regexp.ReplaceAllString` with a single pattern matching all non-allowed tags — complex pattern, hard to maintain. (3) Using `html.Tokenizer` from `golang.org/x/net/html` — external dependency.
**Rationale**: The current code calls `reTags.FindAllString` to get all tags, then calls `strings.ReplaceAll(raw, tag, "")` for each non-allowed tag. For a paragraph with 20 unrecognized tags and 5000 bytes, this is 20 * 5000 = 100K bytes of scanning. A single pass through the string is O(M) regardless of tag count. The `strings.Builder` pre-grows to `len(raw)` capacity to avoid reallocations.

### Decision: Extract to `internal/renderutil` (Not `internal/render/util`)

**Choice**: Create a new package `internal/renderutil` containing the shared inline-rendering functions.
**Alternatives considered**: (1) `internal/render/util` — Go treats this as a sub-package of `render`, but `normalize` importing `render/util` would create a new dependency from `normalize` to the `render` subtree, which is architecturally wrong (normalize is upstream of render). (2) Putting shared code in `model` — these are rendering functions, not domain types. (3) Keeping duplication — increases maintenance burden.
**Rationale**: `internal/renderutil` is a sibling package at the same level as `normalize` and `render`. Both can import it without creating circular dependencies. The package name (`renderutil`) clearly communicates its purpose. It follows the existing project convention of flat `internal/` packages.

## Data Flow

### Cache Lookup Flow (with Filesystem)

```
    LookupService.Lookup(ctx, request)
         |
         v
    cache.BuildKey(request)
         |
         v
    FilesystemStore.Get(ctx, key)
         |
         +---> SHA256(key) -> filename
         |
         +---> os.ReadFile(cacheDir/filename.json)
         |         |
         |    [file not found]     [file found]
         |         |                    |
         |         v                    v
         |    return (zero,        json.Unmarshal -> envelope
         |     false, nil)              |
         |                        [unmarshal error]  [ok]
         |                              |              |
         |                              v              v
         |                        return (zero,   time.Now() > envelope.ExpiresAt?
         |                         false, nil)         |
         |                                        [expired]     [valid]
         |                                            |            |
         |                                            v            v
         |                                      return (zero,  return (envelope.Data,
         |                                       false, nil)     true, nil)
         |
    [cache miss] -----> fan-out to sources -----> aggregate result
         |
         v
    FilesystemStore.Set(ctx, key, result)
         |
         +---> envelope{ExpiresAt: now+TTL, Data: result}
         +---> json.Marshal(envelope)
         +---> os.WriteFile(cacheDir/filename.tmp)
         +---> os.Rename(tmp -> cacheDir/filename.json)
         +---> [any error -> silently ignore, return nil]
```

### Parallel Fan-out Sequence

```
    LookupService.Lookup(ctx, request)
         |
         v
    [cache miss]
         |
         v
    registry.SourcesFor(request) -> []Source
         |
         v
    results := make(chan sourceOutcome, len(sources))
    var wg sync.WaitGroup
         |
         +---> [for each source]:
         |         wg.Add(1)
         |         go func(src Source) {
         |             defer wg.Done()
         |             result, err := src.Lookup(ctx, request)
         |             results <- sourceOutcome{src, result, err}
         |         }(source)
         |
         +---> go func() { wg.Wait(); close(results) }()
         |
         +---> for outcome := range results {
         |         [aggregate: append entries, warnings, problems]
         |         [on error: wrap as Problem, continue]
         |     }
         |
         v
    cache.Set(ctx, key, aggregatedResult)
         |
         v
    return aggregatedResult
```

### Package Dependency Graph

**Before (current):**

```
    cmd/dlexa/main
         |
         v
    internal/app
         |
         +---> config, cache, query, source, render, platform, doctor, version
                                  |
    query ---> cache, source      |
    source --> fetch, parse, normalize
    normalize --> parse, model
    render ----> model

    [normalize and render each have DUPLICATE copies of:]
      renderInlineMarkdown, needsInlineSpace, shouldGlue*,
      lastInlineWordRune, firstInlineWordRune,
      renderTableMarkdown, renderTableHTML, formatHTMLTableCell,
      renderHTMLTableCellContent, renderHTMLFromMarkdownSubset,
      isSimpleMarkdownTable, isSimpleMarkdownRow,
      tableRowTexts, tableColumnWidths, formatTableRow,
      formatTableDivider, normalizeMarkdownTableCellText,
      shouldWrapStyledBuffer, formatHTMLTableCell
```

**After (with renderutil):**

```
    cmd/dlexa/main
         |
         v
    internal/app
         |
         +---> config, cache, query, source, render, platform, doctor, version
                                  |
    query ---> cache, source      |
    source --> fetch, parse, normalize
    normalize --> parse, model, renderutil    <-- NEW dependency
    render ----> model, renderutil            <-- NEW dependency

    internal/renderutil  (NEW)
         |
         +---> model (only dependency)

    [shared functions live in renderutil, imported by both normalize and render]
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/cache/filesystem.go` | Create | `FilesystemStore` implementing `cache.Store` with JSON + TTL + atomic write |
| `internal/cache/filesystem_test.go` | Create | TDD tests: cold miss, warm hit, expired entry, concurrent read/write, corrupt file, permission error |
| `internal/cache/memory_test.go` | Create | Race-detector tests for `MemoryStore` (prerequisite for Phase 2) |
| `internal/config/interfaces.go` | Modify | Add `CacheTTL time.Duration` field to `RuntimeConfig` |
| `internal/config/static.go` | Modify | Set `CacheTTL: 24 * time.Hour` in `DefaultRuntimeConfig()` |
| `internal/app/wiring.go` | Modify | Wire `cache.NewFilesystemStore(cacheDir, cfg.CacheTTL)` instead of `cache.NewMemoryStore()`, with fallback to `NewMemoryStore()` on `os.UserCacheDir()` error |
| `internal/query/service.go` | Modify | Replace sequential `for` loop with goroutine fan-out using `sync.WaitGroup` + buffered channel |
| `internal/query/service_test.go` | Modify | Add `TestLookupQueriesSourcesConcurrently` with delayed stubs to verify parallel execution |
| `internal/fetch/http.go` | Modify | Fix `isChallengeBody` to truncate `[]byte` before string conversion |
| `internal/fetch/http_test.go` | Modify | Add `BenchmarkIsChallengeBody` (TDD baseline before fix) |
| `internal/parse/dpd.go` | Modify | Fix `isChallengePage` truncation, rewrite `preserveSemanticSpans` as single-pass builder |
| `internal/parse/dpd_test.go` | Modify | Add `BenchmarkIsChallengePage`, `BenchmarkPreserveSemanticSpans` (TDD baselines) |
| `internal/renderutil/inline.go` | Create | Shared inline-render functions extracted from normalize and render |
| `internal/renderutil/table.go` | Create | Shared table-render functions extracted from normalize and render |
| `internal/renderutil/inline_test.go` | Create | Tests for shared inline-render functions (TDD — written before extraction) |
| `internal/renderutil/table_test.go` | Create | Tests for shared table-render functions |
| `internal/normalize/dpd.go` | Modify | Remove duplicated render/table functions, import `renderutil` |
| `internal/render/markdown.go` | Modify | Remove duplicated render/table functions, import `renderutil` |
| `internal/fetch/bench_test.go` | Create | `BenchmarkDPDFetch` (mocked HTTP, measures parse pipeline overhead) |
| `internal/parse/bench_test.go` | Create | `BenchmarkDPDParse` with representative HTML fixtures |
| `internal/normalize/bench_test.go` | Create | `BenchmarkDPDNormalize` with representative parsed articles |
| `internal/render/bench_test.go` | Create | `BenchmarkMarkdownRender` with representative lookup results |

## Interfaces / Contracts

### `cache.FilesystemStore`

```go
package cache

// FilesystemStore persists cache entries as JSON files under a directory.
// It implements cache.Store with graceful degradation: all filesystem/JSON
// errors are swallowed and treated as cache misses (Get) or no-ops (Set).
type FilesystemStore struct {
    dir string
    ttl time.Duration
    now func() time.Time  // injectable clock for testing
}

// NewFilesystemStore creates a store that persists entries under dir.
// dir is created with os.MkdirAll on first write. ttl controls entry lifetime.
func NewFilesystemStore(dir string, ttl time.Duration) *FilesystemStore

// Get returns the cached result if the file exists, is valid JSON, and has not expired.
// On any error (file not found, corrupt JSON, expired), returns (zero, false, nil).
func (s *FilesystemStore) Get(ctx context.Context, key string) (model.LookupResult, bool, error)

// Set writes the result to a JSON file with an expiry envelope.
// Uses atomic write (temp file + rename). On any error, returns nil (silent failure).
func (s *FilesystemStore) Set(ctx context.Context, key string, result model.LookupResult) error
```

### Cache Envelope (internal, unexported)

```go
type cacheEnvelope struct {
    ExpiresAt time.Time          `json:"expires_at"`
    CreatedAt time.Time          `json:"created_at"`
    Data      model.LookupResult `json:"data"`
}
```

### Fan-out Outcome (internal to query package)

```go
type sourceOutcome struct {
    source source.Source
    result model.SourceResult
    err    error
}
```

### `internal/renderutil` Public API

```go
package renderutil

// RenderInlineMarkdown renders a slice of model.Inline to markdown string.
func RenderInlineMarkdown(inlines []model.Inline) string

// RenderMarkdownInlines renders inlines for the render layer (handles emphasis
// unwrapping for mention/correction children).
func RenderMarkdownInlines(inlines []model.Inline) string

// NeedsInlineSpace returns true if a space separator is needed between
// the current accumulated string and the next piece.
func NeedsInlineSpace(current, next string) bool

// ShouldGlueInlineWordBoundary returns true if two adjacent pieces should
// be glued without space (both end/start with letters across markdown markers).
func ShouldGlueInlineWordBoundary(current, next string) bool

// ShouldWrapStyledBuffer returns true if a buffer of inlines should be
// wrapped with style markers.
func ShouldWrapStyledBuffer(buffer []model.Inline) bool

// LastInlineWordRune returns the last significant rune in a string,
// skipping markdown markers.
func LastInlineWordRune(raw string) (rune, bool)

// FirstInlineWordRune returns the first significant rune in a string,
// skipping markdown markers.
func FirstInlineWordRune(raw string) (rune, bool)

// RenderTableMarkdown renders a model.Table as markdown (pipe table) or
// falls back to HTML for complex tables.
func RenderTableMarkdown(table model.Table, indent string) string

// RenderTableHTML renders a model.Table as an HTML table string.
func RenderTableHTML(table model.Table, indent string) string
```

Note: The `normalize` package currently has its own `renderInlineMarkdown` that calls its own `renderInlineMarkdownItem`. The `render` package has `renderMarkdownInlines` that calls `renderMarkdownInline`. These two are *nearly* identical but have subtle differences in how they handle `InlineKindExample` (normalize wraps with `‹›`, render wraps with `*`). The extraction MUST preserve both behaviors. The design choice is:

- `RenderInlineMarkdown` (used by normalize): wraps examples with `‹›`
- `RenderMarkdownInlines` (used by render): wraps examples with `*`

Both share the helper functions (`NeedsInlineSpace`, `ShouldGlue*`, etc.) which are truly identical.

### `RuntimeConfig` Extension

```go
type RuntimeConfig struct {
    DefaultFormat  string
    DefaultSources []string
    CacheEnabled   bool
    CacheTTL       time.Duration  // NEW — default 24h
    DPD            DPDConfig
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `FilesystemStore` Get/Set/TTL/corruption | Table-driven tests with `t.TempDir()` as cache dir; injectable `now` clock to test TTL without sleeping |
| Unit | `MemoryStore` concurrent access | `TestMemoryStoreConcurrentReadWrite` with `-race` flag; N goroutines writing, M goroutines reading |
| Unit | Parallel fan-out correctness | Stub sources with `time.Sleep` delays; verify total time < sum of delays; verify all results aggregated |
| Unit | `preserveSemanticSpans` single-pass | Existing golden-file tests as regression guard; new edge-case tests for empty input, no tags, all-allowed tags, all-disallowed tags |
| Unit | `renderutil` shared functions | Port existing tests from normalize and render; verify identical output via golden comparison |
| Benchmark | `BenchmarkIsChallengeBody` | 100KB and 500KB body fixtures; measure allocs/op before and after fix |
| Benchmark | `BenchmarkIsChallengePage` | Same fixtures as above, string input |
| Benchmark | `BenchmarkPreserveSemanticSpans` | Real DPD paragraph HTML with varying tag counts (5, 20, 50 unrecognized tags) |
| Benchmark | `BenchmarkDPDParse` | Full DPD HTML page fixture; measure ns/op and allocs/op |
| Benchmark | `BenchmarkDPDNormalize` | Parsed article fixtures; measure ns/op and allocs/op |
| Benchmark | `BenchmarkMarkdownRender` | Full LookupResult fixtures; measure ns/op and allocs/op |
| Race | All concurrent tests | `go test -race ./...` in CI; specifically targets `MemoryStore`, `FilesystemStore` (concurrent key access), and fan-out |
| Integration | Full pipeline with filesystem cache | Write a test that exercises `LookupService.Lookup` with a `FilesystemStore`, verifies first call misses, second call hits, expired entry re-fetches |
| Regression | Golden-file comparison after renderutil extraction | Run existing golden tests in `parse`, `normalize`, and `render` before and after extraction; any diff is a bug |

### Benchmark Harness Design

Benchmarks follow the existing table-driven pattern in the codebase:

```go
func BenchmarkPreserveSemanticSpans(b *testing.B) {
    cases := []struct {
        name  string
        input string
    }{
        {"small_5_tags", fixtureSmall},
        {"medium_20_tags", fixtureMedium},
        {"large_50_tags", fixtureLarge},
    }
    for _, tc := range cases {
        b.Run(tc.name, func(b *testing.B) {
            b.ReportAllocs()
            for i := 0; i < b.N; i++ {
                preserveSemanticSpans(tc.input)
            }
        })
    }
}
```

All benchmark functions:
- Use `b.ReportAllocs()` to track allocations
- Use sub-benchmarks (`b.Run`) for multiple input sizes
- Use realistic fixtures (not synthetic strings)
- Are placed in `*_test.go` files alongside the code they benchmark

## Migration / Rollout

No migration required. The filesystem cache creates its directory on first write. Existing users will experience a cold cache on their first invocation after the update, then cache hits on subsequent calls.

The `FilesystemStore` is wired in `wiring.go` with a fallback:

```go
cacheDir, err := os.UserCacheDir()
var cacheStore cache.Store
if err != nil {
    cacheStore = cache.NewMemoryStore() // fallback: no persistent cache
} else {
    cacheStore = cache.NewFilesystemStore(
        filepath.Join(cacheDir, "dlexa"),
        runtimeConfig.CacheTTL,
    )
}
```

This ensures the CLI works even on systems where `os.UserCacheDir()` fails (e.g., restricted containers).

## Phase Implementation Order

### Phase 0: Benchmark Suite
- Create benchmark files in `fetch`, `parse`, `normalize`, `render`
- Create `MemoryStore` race-detector test
- Zero functional changes — pure test code
- **Unblocks**: All subsequent phases

### Phase 1: Filesystem Cache
- Create `internal/cache/filesystem.go` and `internal/cache/filesystem_test.go` (TDD)
- Add `CacheTTL` to `RuntimeConfig`
- Wire in `wiring.go`
- **Depends on**: Phase 0 (need race-detector test for MemoryStore as baseline)

### Phase 2: Parallel Source Fan-out
- Modify `internal/query/service.go`
- Add concurrency test to `internal/query/service_test.go` (TDD)
- **Depends on**: Phase 0 (need race-detector tests to verify no races under fan-out)

### Phase 3: Allocation Fixes
- Fix `isChallengeBody` in `internal/fetch/http.go`
- Fix `isChallengePage` in `internal/parse/dpd.go`
- Rewrite `preserveSemanticSpans` in `internal/parse/dpd.go`
- **Depends on**: Phase 0 (need benchmark baselines to prove improvement)

### Phase 4: Extract `internal/renderutil`
- Create `internal/renderutil/` package
- Extract shared functions from `normalize/dpd.go` and `render/markdown.go`
- Update imports
- **Depends on**: Phase 0 (need golden-file test coverage verification)

## Open Questions

- [x] Whether `renderInlineMarkdown` (normalize) and `renderMarkdownInlines` (render) can be unified into a single function — **Answer: No.** They handle `InlineKindExample` differently (normalize: `‹›`, render: `*`). Both will exist in `renderutil` as separate public functions sharing the same helpers.
- [ ] Whether `CacheTTL` should be exposed as a CLI flag (e.g., `--cache-ttl 12h`) or only as a config value — recommend config-only for now, CLI flag in a separate change.
- [ ] Whether stale cache entries should be cleaned up proactively (background goroutine or on-write eviction) — recommend deferring cleanup to a separate `dlexa cache clear` command (out of scope per proposal).

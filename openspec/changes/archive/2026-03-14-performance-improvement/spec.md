# Performance Improvement Specification

## Purpose

Define the behavioral contract for all performance improvements to dlexa: a benchmark baseline suite, a persistent filesystem cache, parallel source fan-out, allocation optimizations in hot-path functions, and deduplication of shared render utilities. Each phase is independently testable and independently revertible.

---

## Phase 0 — Benchmark Suite

### Requirement: Benchmark Functions Exist for All Hot-Path Operations

The system MUST provide Go benchmark functions (`Benchmark*`) covering the dominant cost centers in the request lifecycle. These benchmarks serve as the baseline for all subsequent optimization phases; no optimization SHALL ship without a before/after benchmark comparison.

#### Scenario: HTTP fetch benchmark exists

- GIVEN the `internal/fetch` package contains `DPDFetcher`
- WHEN `go test -bench=BenchmarkDPDFetch ./internal/fetch/...` is executed
- THEN a benchmark function MUST execute at least one fetch cycle against a stubbed HTTP server
- AND the benchmark MUST report `ns/op`, `B/op`, and `allocs/op`

#### Scenario: Parse pipeline benchmark exists

- GIVEN the `internal/parse` package contains `DPDArticleParser`
- WHEN `go test -bench=BenchmarkDPDParse ./internal/parse/...` is executed
- THEN a benchmark function MUST parse a representative DPD HTML document from a fixture
- AND the benchmark MUST report `ns/op`, `B/op`, and `allocs/op`

#### Scenario: Normalize pipeline benchmark exists

- GIVEN the `internal/normalize` package contains `DPDNormalizer`
- WHEN `go test -bench=BenchmarkDPDNormalize ./internal/normalize/...` is executed
- THEN a benchmark function MUST normalize a representative parsed result from a fixture
- AND the benchmark MUST report `ns/op`, `B/op`, and `allocs/op`

#### Scenario: Render pipeline benchmark exists

- GIVEN the `internal/render` package contains `MarkdownRenderer`
- WHEN `go test -bench=BenchmarkMarkdownRender ./internal/render/...` is executed
- THEN a benchmark function MUST render a representative `LookupResult` to Markdown
- AND the benchmark MUST report `ns/op`, `B/op`, and `allocs/op`

#### Scenario: End-to-end lookup benchmark exists

- GIVEN the full pipeline (fetch -> parse -> normalize -> render) can be exercised with stubbed I/O
- WHEN `go test -bench=BenchmarkEndToEndLookup ./internal/...` is executed
- THEN a benchmark function MUST exercise the complete lookup path with an in-memory stub source
- AND the benchmark MUST report `ns/op`, `B/op`, and `allocs/op`

### Requirement: Benchmarks Produce Reproducible Numbers

Benchmark functions MUST be deterministic and MUST NOT depend on network I/O, wall-clock timing variability, or external services.

#### Scenario: Benchmarks use stubbed I/O

- GIVEN any benchmark that exercises `fetch.DPDFetcher`
- WHEN the benchmark is executed
- THEN it MUST use a stubbed HTTP client or `httptest.Server` serving fixture data
- AND the benchmark MUST NOT make real network calls

#### Scenario: Benchmarks are idempotent across runs

- GIVEN a benchmark function is executed twice on the same machine
- WHEN the two runs are compared
- THEN the `ns/op` values SHOULD be within a reasonable noise margin (< 20% variance)
- AND the `B/op` and `allocs/op` values MUST be identical across runs

---

## Phase 1 — Filesystem Cache

### Requirement: FilesystemStore Implements cache.Store

The system MUST provide a `cache.FilesystemStore` type in the `internal/cache` package that implements the existing `cache.Store` interface (`Get` and `Set` methods). The store MUST persist `model.LookupResult` values as JSON files on disk.

#### Scenario: Cache miss on cold store

- GIVEN a newly constructed `FilesystemStore` with an empty cache directory
- WHEN `Get(ctx, "some-key")` is called
- THEN the second return value (ok) MUST be `false`
- AND the error MUST be `nil`
- AND the returned `LookupResult` MUST be the zero value

#### Scenario: Cache hit after Set

- GIVEN a `FilesystemStore` instance
- WHEN `Set(ctx, "my-key", result)` is called with a valid `LookupResult`
- AND `Get(ctx, "my-key")` is subsequently called
- THEN the second return value (ok) MUST be `true`
- AND the returned `LookupResult` MUST be deeply equal to the original result
- AND the error MUST be `nil`

#### Scenario: Data survives across store instances

- GIVEN a `FilesystemStore` has written a cache entry with `Set`
- WHEN a new `FilesystemStore` instance is constructed with the same cache directory
- AND `Get` is called with the same key
- THEN the cached result MUST be returned successfully
- AND this validates that persistence is filesystem-based, not in-memory

### Requirement: FilesystemStore Uses Cross-Platform Cache Path

The system MUST use `os.UserCacheDir()` to determine the base cache directory, appending `dlexa/` as the application-specific subdirectory. This ensures correct behavior on Windows (`%LocalAppData%`), macOS (`~/Library/Caches`), and Linux (`~/.cache`).

#### Scenario: Default cache directory uses os.UserCacheDir

- GIVEN a `FilesystemStore` constructed with default settings
- WHEN the store resolves its cache directory
- THEN the directory MUST be `<os.UserCacheDir()>/dlexa/`
- AND the store MUST create the directory if it does not exist

#### Scenario: Cache directory is overridable for testing

- GIVEN a `FilesystemStore` constructor accepts an optional directory override
- WHEN the override is set to `t.TempDir()`
- THEN the store MUST use the provided directory instead of `os.UserCacheDir()`
- AND this enables deterministic testing without polluting the real cache

### Requirement: Cache Entries Have TTL-Based Expiration

Each cached entry MUST record a timestamp at write time. On read, if the entry's age exceeds the configured TTL, the entry MUST be treated as a cache miss. The default TTL MUST be 24 hours.

#### Scenario: Expired entry is treated as cache miss

- GIVEN a `FilesystemStore` with a TTL of 1 second
- WHEN `Set(ctx, "key", result)` is called
- AND more than 1 second elapses (simulated via clock injection or file modification)
- AND `Get(ctx, "key")` is called
- THEN the second return value (ok) MUST be `false`
- AND the expired file MAY remain on disk (lazy cleanup)

#### Scenario: Non-expired entry is a cache hit

- GIVEN a `FilesystemStore` with a TTL of 24 hours
- WHEN `Set(ctx, "key", result)` is called
- AND `Get(ctx, "key")` is called within 24 hours
- THEN the second return value (ok) MUST be `true`
- AND the returned result MUST match the stored value

#### Scenario: TTL is configurable

- GIVEN a `FilesystemStore` constructed with a custom TTL of 1 hour
- WHEN the store checks entry freshness
- THEN it MUST use 1 hour as the expiration threshold, not the default 24 hours

### Requirement: FilesystemStore Handles Edge Cases Gracefully

The cache MUST NOT crash or return incorrect data when encountering corrupted files, concurrent access, or disk-full conditions.

#### Scenario: Corrupted cache file returns miss, not error

- GIVEN a cache entry file exists on disk but contains invalid JSON
- WHEN `Get(ctx, "key")` is called
- THEN the second return value (ok) MUST be `false`
- AND the error MUST be `nil` (corrupted files are treated as misses, not failures)
- AND the corrupted file SHOULD be removed from disk

#### Scenario: Concurrent reads and writes do not cause data races

- GIVEN a `FilesystemStore` instance shared across multiple goroutines
- WHEN multiple goroutines concurrently call `Get` and `Set` with overlapping keys
- THEN `go test -race` MUST pass without detecting any data race
- AND no goroutine MUST receive a partially written or corrupted result

#### Scenario: Disk full on Set returns error without crashing

- GIVEN a `FilesystemStore` operating on a filesystem with no available space
- WHEN `Set(ctx, "key", result)` is called
- THEN the method MUST return a non-nil error
- AND the store MUST NOT leave a partially written cache file on disk
- AND subsequent `Get` calls for other keys MUST continue to work

### Requirement: FilesystemStore Cache Key Encoding

Cache keys MUST be encoded into safe filesystem paths. The key encoding MUST be deterministic and collision-resistant.

#### Scenario: Cache key with special characters produces valid filename

- GIVEN a cache key containing characters unsafe for filenames (e.g., `|`, `/`, `:`)
- WHEN the key is encoded into a filename
- THEN the resulting filename MUST be valid on Windows, macOS, and Linux
- AND the encoding MUST be deterministic (same key always produces same filename)

#### Scenario: Distinct keys produce distinct filenames

- GIVEN two different cache keys
- WHEN both are encoded into filenames
- THEN the resulting filenames MUST be different

### Requirement: Wiring Swaps MemoryStore for FilesystemStore

The composition root (`internal/app/wiring.go`) MUST wire `cache.FilesystemStore` as the `cache.Store` implementation instead of `cache.MemoryStore`.

#### Scenario: Default wiring uses FilesystemStore

- GIVEN the `app.New()` composition root is called
- WHEN the `LookupService` is constructed
- THEN it MUST receive a `FilesystemStore` as its cache, not a `MemoryStore`

---

## Phase 2 — Parallel Source Fan-out

### Requirement: LookupService Queries Sources Concurrently

The `query.LookupService.Lookup` method MUST dispatch source lookups concurrently using stdlib goroutines, `sync.WaitGroup`, and a buffered result channel. No external dependencies (e.g., `errgroup`) SHALL be introduced.

#### Scenario: Single source executes normally

- GIVEN a `LookupService` with one registered source
- WHEN `Lookup(ctx, request)` is called
- THEN the single source MUST be queried
- AND the result MUST be identical to the current sequential behavior

#### Scenario: Multiple sources execute concurrently

- GIVEN a `LookupService` with three registered sources, each with an artificial 100ms delay
- WHEN `Lookup(ctx, request)` is called
- THEN total wall-clock time MUST be less than 200ms (proving concurrency, not sequential 300ms)
- AND the result MUST contain entries from all three sources

#### Scenario: One source fails, others succeed

- GIVEN a `LookupService` with three sources, where the second source returns an error
- WHEN `Lookup(ctx, request)` is called
- THEN the result MUST contain entries from the two successful sources
- AND the result MUST contain a `Problem` entry for the failed source
- AND the method MUST NOT return a top-level error

#### Scenario: All sources fail

- GIVEN a `LookupService` with two sources, both returning errors
- WHEN `Lookup(ctx, request)` is called
- THEN the result MUST contain `Problem` entries for both failed sources
- AND the result MUST have zero entries
- AND the method MUST NOT return a top-level error (problems are reported in-band)

#### Scenario: Race detector passes under concurrent fan-out

- GIVEN the parallel fan-out implementation
- WHEN `go test -race ./internal/query/...` is executed with multiple sources
- THEN the race detector MUST report zero data races
- AND this test MUST be written before the implementation (TDD)

### Requirement: Result Ordering Is Deterministic

Despite concurrent execution, the aggregated results MUST preserve source ordering by priority (as defined by `SourceDescriptor.Priority`).

#### Scenario: Results are ordered by source priority

- GIVEN three sources with priorities 1, 2, and 3 completing in arbitrary order
- WHEN all sources complete and results are aggregated
- THEN `result.Sources` MUST be ordered by ascending source priority
- AND `result.Entries` MUST preserve the same source-priority ordering

### Requirement: Cache Concurrent Access Is Safe

The `cache.MemoryStore` (and `FilesystemStore`) MUST be safe for concurrent access from multiple goroutines dispatched by the parallel fan-out.

#### Scenario: MemoryStore race detector test passes

- GIVEN `cache.MemoryStore` is accessed concurrently by multiple goroutines
- WHEN `go test -race ./internal/cache/...` is executed
- THEN the race detector MUST report zero data races
- AND this test MUST be written before Phase 2 implementation

---

## Phase 3 — Allocation Fixes

### Requirement: isChallengeBody Truncates Before String Conversion

The `fetch.isChallengeBody` function MUST truncate the raw `[]byte` body to `challengeBodySnippetLimit` (1024 bytes) BEFORE converting to string and calling `strings.ToLower`. This eliminates two unnecessary full-body allocations.

#### Scenario: Behavior unchanged for challenge body

- GIVEN a `[]byte` body containing "Cloudflare" and "challenge" within the first 1024 bytes
- WHEN `isChallengeBody(body)` is called
- THEN the function MUST return `true`
- AND the behavior MUST be identical to the current implementation

#### Scenario: Behavior unchanged for non-challenge body

- GIVEN a `[]byte` body that does not contain challenge markers
- WHEN `isChallengeBody(body)` is called
- THEN the function MUST return `false`

#### Scenario: Large body allocates only snippet-sized string

- GIVEN a `[]byte` body of 500KB
- WHEN `BenchmarkIsChallengeBody` is executed
- THEN `B/op` MUST report approximately `challengeBodySnippetLimit` bytes (1024), not 500KB
- AND `allocs/op` MUST be reduced compared to the pre-fix baseline

### Requirement: isChallengePage Truncates Before ToLower

The `parse.isChallengePage` function MUST truncate the input string to a snippet limit BEFORE calling `strings.ToLower`. This eliminates a full-body `strings.ToLower` allocation at parse entry.

#### Scenario: Behavior unchanged for challenge page

- GIVEN a `string` body containing "cloudflare" and "challenge" within the first 1024 characters
- WHEN `isChallengePage(body)` is called
- THEN the function MUST return `true`

#### Scenario: Behavior unchanged for non-challenge page

- GIVEN a `string` body that does not contain challenge markers
- WHEN `isChallengePage(body)` is called
- THEN the function MUST return `false`

#### Scenario: Large body allocates only snippet-sized string

- GIVEN a `string` body of 500KB
- WHEN `BenchmarkIsChallengePage` is executed
- THEN `B/op` MUST report approximately snippet-limit bytes, not 500KB

### Requirement: preserveSemanticSpans Uses Single-Pass Builder

The `parse.preserveSemanticSpans` function MUST be rewritten to use a single-pass `strings.Builder` approach instead of the current O(N*M) loop of `strings.ReplaceAll` calls. The function MUST produce identical output for all inputs.

#### Scenario: Behavior unchanged (golden tests pass)

- GIVEN all existing parse golden-file tests
- WHEN the rewritten `preserveSemanticSpans` is used
- THEN all golden tests MUST pass with zero diff
- AND no behavioral regression is permitted

#### Scenario: Performance scales linearly with input size

- GIVEN `BenchmarkPreserveSemanticSpans` with a fixed-size input containing varying numbers of tags
- WHEN the benchmark is executed
- THEN wall-clock time MUST NOT grow proportionally to the number of unrecognized tags (O(N*M))
- AND the single-pass approach MUST show measurable improvement in `ns/op`

#### Scenario: Empty tag list produces unchanged output

- GIVEN an input string with no HTML tags
- WHEN `preserveSemanticSpans(input)` is called
- THEN the output MUST be identical to the input

#### Scenario: All tags are allowed

- GIVEN an input string where every HTML tag is in the allowed set
- WHEN `preserveSemanticSpans(input)` is called
- THEN the output MUST be identical to the input (no tags removed)

#### Scenario: Nested tags are handled correctly

- GIVEN an input string with nested tags, some allowed and some not
- WHEN `preserveSemanticSpans(input)` is called
- THEN only disallowed tags MUST be removed
- AND allowed tags and their nesting structure MUST be preserved
- AND the text content between tags MUST be unchanged

---

## Phase 4 — Render Utility Deduplication

### Requirement: Shared Render Utilities Extracted to internal/renderutil

The system MUST extract the following duplicated functions from `internal/normalize/dpd.go` and `internal/render/markdown.go` into a new `internal/renderutil` package:

- `renderInlineMarkdown`
- `renderMarkdownInlines`
- `needsInlineSpace`
- `shouldGlueInlineWordBoundary`
- `shouldWrapStyledBuffer`
- `lastInlineWordRune`
- `firstInlineWordRune`
- `renderTableMarkdown`
- `renderTableHTML`

Both `normalize` and `render` packages MUST import from `renderutil` instead of maintaining their own copies.

#### Scenario: All existing golden tests pass after extraction

- GIVEN the full golden-file test suite across `parse`, `normalize`, and `render` packages
- WHEN `go test ./...` is executed after the extraction
- THEN all tests MUST pass with zero diff
- AND no behavioral regression is permitted

#### Scenario: No duplicated render functions remain

- GIVEN the extraction is complete
- WHEN `internal/normalize/dpd.go` and `internal/render/markdown.go` are inspected
- THEN neither file MUST contain local definitions of the extracted functions
- AND both files MUST import `internal/renderutil`

### Requirement: renderutil Package Has Adequate Test Coverage

The new `internal/renderutil` package MUST have its own test file with coverage of the extracted functions.

#### Scenario: renderutil test coverage meets threshold

- GIVEN `internal/renderutil/inline_test.go` exists
- WHEN `go test -cover ./internal/renderutil/...` is executed
- THEN line coverage MUST be at least 90%

#### Scenario: renderutil tests are written before extraction (TDD)

- GIVEN the TDD-first discipline
- WHEN the `renderutil` package is created
- THEN test functions MUST exist and be committed before the function implementations are moved
- AND the tests MUST initially reference the functions from their original locations or define expected behavior independently

### Requirement: Dependency Direction Is Correct

The `renderutil` package MUST NOT import from `normalize` or `render`. It MUST only depend on `model` (for `Inline`, `Table`, and related types) and stdlib packages.

#### Scenario: renderutil has no circular dependencies

- GIVEN the `internal/renderutil` package
- WHEN its import graph is inspected
- THEN it MUST NOT import `internal/normalize` or `internal/render`
- AND `go build ./...` MUST succeed without import cycle errors

---

## Cross-Cutting Requirements

### Requirement: No External Dependencies Introduced

No phase of this change SHALL introduce entries in `go.mod` or `go.sum` that were not already present. The project MUST remain stdlib-only.

#### Scenario: go.mod unchanged after all phases

- GIVEN the project's `go.mod` before this change
- WHEN all phases are complete
- THEN `go.mod` MUST contain no new `require` directives
- AND `go.sum` MUST NOT exist (as before) or MUST contain no new entries

### Requirement: Full Test Suite Passes After Each Phase

Each phase MUST be independently mergeable. After each phase is applied, the full test suite MUST pass.

#### Scenario: go test passes after each phase

- GIVEN a phase N has been applied
- WHEN `go test ./...` is executed
- THEN all tests MUST pass
- AND `go test -race ./...` MUST pass with zero data races

### Requirement: TDD Discipline Enforced

For each phase, tests and/or benchmarks MUST be written BEFORE the implementation they validate. No optimization SHALL ship without a benchmark proving its effect.

#### Scenario: Tests committed before implementation

- GIVEN any phase that modifies production code
- WHEN the git history is inspected
- THEN test/benchmark commits MUST precede (or be part of the same commit as) the corresponding implementation changes
- AND no implementation SHOULD be committed without an accompanying test

---

## Out-of-Scope Guardrails

- This specification does NOT cover streaming HTTP response parsing.
- This specification does NOT permit introducing `golang.org/x/sync/errgroup` or any external dependency.
- This specification does NOT cover server mode, persistent process models, or daemon architectures.
- This specification does NOT cover context propagation to `parse` and `normalize` packages.
- This specification does NOT cover cache invalidation UI, manual cache-clear commands, or cache management tooling.
- This specification does NOT alter the existing `cache.Store` interface signature; `FilesystemStore` implements the same `Get`/`Set` contract.

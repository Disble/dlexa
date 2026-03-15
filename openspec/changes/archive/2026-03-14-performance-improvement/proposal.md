# Proposal: Performance Improvement

## Intent

Every invocation of `dlexa` pays full network + parse + normalize cost because the in-process
`MemoryStore` cache dies with the process. On top of that, the source fan-out is sequential,
meaning latency grows linearly with the number of sources. Two additional hot-path functions
allocate full-body strings unnecessarily, and a tag-stripping loop runs in O(N×M) time.
Duplicated inline-render code between `normalize` and `render` packages creates a maintenance
risk that will compound as the codebase grows.

This change eliminates the dominant cost (repeated network I/O) through a filesystem cache,
unlocks future multi-source scalability via parallel fan-out, and corrects allocation and
algorithmic inefficiencies — all under a strict TDD-first discipline where no optimization
ships without a benchmark baseline proving its effect.

## Scope

### In Scope

- **Phase 0 — Benchmark suite**: Add `Benchmark*` functions across `fetch`, `parse`, `normalize`,
  and `render` packages. This is the prerequisite for all performance work; no optimization ships
  without a baseline.
- **Phase 1 — Filesystem cache**: Implement `cache.FilesystemStore` (JSON + TTL, persisted to
  `os.UserCacheDir()/dlexa/`) behind the existing `cache.Store` interface. Wire it in `wiring.go`.
  Cover with unit + race-detector tests before implementation.
- **Phase 2 — Parallel source fan-out**: Replace the sequential `for` loop in
  `query.LookupService.Lookup` with stdlib goroutines + `sync.WaitGroup` + result channel
  (no external dependencies). Cover with a concurrency test before implementation.
- **Phase 3 — Allocation fixes**: Fix `isChallengeBody` (`fetch/http.go`) and `isChallengePage`
  (`parse/dpd.go`) to truncate raw bytes before string conversion. Rewrite
  `preserveSemanticSpans` as a single-pass builder. Preceded by benchmarks for each target.
- **Phase 4 — Extract `internal/renderutil`**: Deduplicate `renderInlineMarkdown`,
  `renderMarkdownInlines`, `needsInlineSpace`, `shouldGlue*`, `lastInlineWordRune`,
  `firstInlineWordRune`, `renderTableMarkdown`, `renderTableHTML` from `normalize/dpd.go` and
  `render/markdown.go` into a shared package. Preceded by coverage verification of affected paths.

### Out of Scope

- Streaming HTTP response parsing (low ROI for current DPD page sizes)
- Adding `golang.org/x/sync/errgroup` or any external dependency
- Server mode or persistent process model
- Context propagation to `parse` and `normalize` (deferred; low CLI impact)
- Cache invalidation UI or manual cache-clear command (separate change)

## Approach

**TDD-first, phased, stdlib-only.**

Each phase follows the Red-Green-Refactor cycle:
1. Write failing tests / benchmarks that define the target behavior and capture baseline metrics.
2. Implement the change until tests pass and benchmarks show improvement.
3. Refactor without breaking tests.

No phase ships without its tests. No optimization ships without a benchmark comparison.

The parallel fan-out uses `sync.WaitGroup` + a buffered result channel to avoid introducing
`golang.org/x/sync/errgroup` as the first external dependency. The filesystem cache uses
`os.UserCacheDir()` for cross-platform path resolution (handles Windows `%LocalAppData%`,
macOS `~/Library/Caches`, Linux `~/.cache`). The `renderutil` extraction uses the existing
golden-file test suite as a regression safety net.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/cache/memory.go` | Modified | Add race-detector test before any parallel work |
| `internal/cache/filesystem.go` | New | `FilesystemStore` — JSON + TTL, `os.UserCacheDir()` |
| `internal/cache/filesystem_test.go` | New | Unit + race-detector tests (TDD — written first) |
| `internal/query/service.go` | Modified | Replace sequential `for` with goroutine fan-out |
| `internal/query/service_test.go` | Modified | Add concurrency timing test (TDD — written first) |
| `internal/fetch/http.go` | Modified | Fix `isChallengeBody` truncation order |
| `internal/fetch/http_test.go` | Modified | Add `BenchmarkIsChallengeBody` (TDD — written first) |
| `internal/parse/dpd.go` | Modified | Fix `isChallengePage` + rewrite `preserveSemanticSpans` |
| `internal/parse/dpd_test.go` | Modified | Add `BenchmarkIsChallengePage`, `BenchmarkPreserveSemanticSpans` (TDD) |
| `internal/normalize/dpd.go` | Modified | Remove duplicated render helpers, import `renderutil` |
| `internal/render/markdown.go` | Modified | Remove duplicated render helpers, import `renderutil` |
| `internal/renderutil/` | New | Shared package with extracted inline-render utilities |
| `internal/renderutil/inline_test.go` | New | Tests before extraction (TDD — written first) |
| `internal/app/wiring.go` | Modified | Wire `FilesystemStore` instead of `MemoryStore` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Filesystem cache path incorrect on Windows | Medium | Use `os.UserCacheDir()` (stdlib, cross-platform). Add test with `t.TempDir()` to override cache dir. |
| Parallel fan-out introduces data race on cache | High | Write race-detector test (`go test -race`) for `MemoryStore` before any concurrent writes. Phase 2 blocked on Phase 0 race coverage. |
| `renderutil` extraction diverges behavior | Medium | Run full golden-file test suite before and after extraction. Any diff is a bug, not a refactor. |
| `preserveSemanticSpans` rewrite alters semantics | Medium | Existing parse golden tests act as regression guard. Add explicit edge-case tests for empty tag lists and nested tags first. |
| Benchmark suite shows no measurable improvement | Low | If benchmarks prove the allocation fixes are below noise floor, defer those phases and document the decision. |
| TTL choice ages out valid entries | Low | Default TTL 24h. Cache key = query term normalized to lowercase. Document TTL in cache package; make it a config value from `RuntimeConfig`. |

## Rollback Plan

Each phase is an independently mergeable unit behind the existing `cache.Store` interface.

- **Phase 0** (benchmarks): Pure test code — rollback by deleting benchmark files. No risk.
- **Phase 1** (filesystem cache): Revert `wiring.go` to use `cache.NewMemoryStore()`. `FilesystemStore` can remain in the codebase without being wired. Cache directory on disk is inert.
- **Phase 2** (parallel fan-out): Revert `query/service.go` to the sequential `for` loop. No other files affected.
- **Phase 3** (allocation fixes): Each function is a local, self-contained change. Revert individual functions independently via `git revert` on the specific commit.
- **Phase 4** (`renderutil` extraction): Revert import changes in `normalize/dpd.go` and `render/markdown.go` and delete `internal/renderutil/`. The original code is preserved in git history.

If multiple phases have landed, rollback in reverse phase order to maintain a consistent state.

## Dependencies

- No new external dependencies. Stdlib only: `sync`, `os`, `encoding/json`, `time`, `path/filepath`.
- Phase 2 (parallel fan-out) MUST be preceded by Phase 0 (benchmarks + race-detector tests for cache).
- Phase 4 (`renderutil`) MUST be preceded by full test coverage verification for the duplicated code paths in both `normalize` and `render`.

## Success Criteria

- [ ] `go test ./... -race` passes with zero race conditions after Phase 2 lands.
- [ ] `go test -bench=. ./internal/fetch/...` shows `BenchmarkIsChallengeBody` allocates only `challengeBodySnippetLimit` bytes (1024) instead of full body size after Phase 3.
- [ ] `go test -bench=. ./internal/parse/...` shows `BenchmarkPreserveSemanticSpans` is O(M) — wall time does not grow with number of tags for fixed-size input after Phase 3.
- [ ] A repeated `dlexa <term>` invocation (second call, cache warm) completes in under 50ms on a standard developer machine (filesystem I/O only, no network call) after Phase 1.
- [ ] `go test ./internal/renderutil/...` passes with coverage ≥ 90% after Phase 4.
- [ ] `go test ./...` passes in full after each phase independently — no cross-phase breakage.
- [ ] `internal/cache/filesystem_test.go` includes `TestFilesystemStoreConcurrentReadWrite` run under `-race` and it passes.
- [ ] No new external entries in `go.mod` / `go.sum` at any phase.

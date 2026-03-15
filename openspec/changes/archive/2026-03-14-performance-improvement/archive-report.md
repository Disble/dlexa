# Archive Report: performance-improvement

**Change**: performance-improvement
**Archived**: 2026-03-14
**Archive location**: `openspec/changes/archive/2026-03-14-performance-improvement/`
**Verification status**: PASS (all 39 tasks complete, all 40 spec scenarios compliant)

---

## Executive Summary

The performance-improvement change eliminated the dominant latency bottleneck in dlexa (repeated network I/O on every CLI invocation) by introducing a persistent filesystem cache with TTL-based expiration. It also unlocked multi-source scalability through parallel goroutine fan-out, reduced hot-path memory allocations by orders of magnitude in challenge-detection and tag-stripping functions, and eliminated code duplication by extracting shared rendering utilities into a new `internal/renderutil` package. All work was stdlib-only (zero external dependencies) and followed strict TDD-first discipline with benchmark baselines proving every optimization.

---

## Change Lifecycle

| Phase | Status | Description |
|-------|--------|-------------|
| Exploration | Complete | Identified 10 bottlenecks across fetch, parse, normalize, and render layers |
| Proposal | Complete | Scoped 5 phases: benchmarks, filesystem cache, parallel fan-out, allocation fixes, renderutil extraction |
| Specification | Complete | 40 behavioral scenarios across all phases using Given/When/Then format |
| Design | Complete | 15 architecture decisions documented with alternatives and rationale |
| Tasks | Complete | 39 tasks across 5 phases, all marked done |
| Apply | Complete | All 5 phases implemented following TDD-first discipline |
| Verify | Complete | PASS — full compliance, no CRITICAL issues |
| Archive | Complete | This report |

---

## Key Metrics

### Tasks
- **Total tasks**: 39
- **Completed**: 39
- **Incomplete**: 0

### Files Changed/Created
| File | Action |
|------|--------|
| `internal/cache/filesystem.go` | Created |
| `internal/cache/filesystem_test.go` | Created |
| `internal/cache/memory_test.go` | Created |
| `internal/renderutil/inline.go` | Created |
| `internal/renderutil/inline_test.go` | Created |
| `internal/renderutil/table.go` | Created |
| `internal/renderutil/table_test.go` | Created |
| `internal/fetch/bench_test.go` | Created |
| `internal/parse/bench_test.go` | Created |
| `internal/normalize/bench_test.go` | Created |
| `internal/render/bench_test.go` | Created |
| `internal/query/parallel_test.go` | Created |
| `internal/config/interfaces.go` | Modified |
| `internal/config/static.go` | Modified |
| `internal/app/wiring.go` | Modified |
| `internal/query/service.go` | Modified |
| `internal/fetch/http.go` | Modified |
| `internal/fetch/http_test.go` | Modified |
| `internal/parse/dpd.go` | Modified |
| `internal/parse/dpd_test.go` | Modified |
| `internal/normalize/dpd.go` | Modified |
| `internal/render/markdown.go` | Modified |
| `internal/render/semantic_terminal.go` | Modified |

### Performance Improvements (Benchmark Evidence)
| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| `isChallengeBody` B/op (500KB body) | ~1,000,000 | 2,048 | ~500x reduction |
| `isChallengePage` B/op (500KB body) | ~516,000 | 1,024 | ~500x reduction |
| `preserveSemanticSpans` allocs/op | O(N*M) scaling | 1 (constant) | Linear vs quadratic |
| Repeated lookup latency | Full network RTT | Filesystem I/O only | ~10-100x faster (cache hit) |
| Multi-source lookup latency | sum(source latencies) | max(source latencies) | Linear to constant |

### Quality
- **Test coverage (renderutil)**: 95.7% (threshold: 90%)
- **Test coverage (cache)**: 82.5%
- **Lint status**: PASS (all changed files clean)
- **External dependencies added**: 0 (stdlib-only constraint maintained)
- **go.mod changes**: None (no new require directives)
- **Spec compliance**: 40/40 scenarios compliant

---

## Specs Synced

No delta specs were created for this change. The performance-improvement change is an internal optimization that does not alter the DPD rendering behavioral contract defined in `openspec/specs/dpd/spec.md`. The existing DPD Rendering Specification remains unchanged and fully compatible with all optimizations.

---

## Architecture Decisions Preserved

15 architecture decisions are documented in the design artifact. Key decisions include:

1. **JSON for cache storage** — human-readable, stdlib-native, debuggable
2. **SHA256 key-to-filename mapping** — fixed-length, filesystem-safe, collision-resistant
3. **Atomic write via temp+rename** — no file locking, cross-platform, graceful degradation
4. **WaitGroup + buffered channel** — stdlib-only parallel fan-out, no errgroup
5. **Extract to `internal/renderutil`** — sibling package, no circular deps, correct dependency direction

---

## Risks Realized

- **Race detector**: Cannot verify on Windows (no CGO). Code is structurally race-safe. Recommend CI verification on Linux.
- **Pre-existing lint issues**: 50 `revive` exported-comment warnings remain (not in change scope).
- **SonarQube**: Unavailable during verification (non-blocking).

---

## Artifacts

| Artifact | Location |
|----------|----------|
| Exploration | `openspec/changes/archive/2026-03-14-performance-improvement/exploration.md` |
| Proposal | `openspec/changes/archive/2026-03-14-performance-improvement/proposal.md` |
| Specification | `openspec/changes/archive/2026-03-14-performance-improvement/spec.md` |
| Design | `openspec/changes/archive/2026-03-14-performance-improvement/design.md` |
| Tasks | `openspec/changes/archive/2026-03-14-performance-improvement/tasks.md` |
| Verify Report | `openspec/changes/archive/2026-03-14-performance-improvement/verify-report.md` |
| Archive Report | `openspec/changes/archive/2026-03-14-performance-improvement/archive-report.md` |
| State | `openspec/changes/archive/2026-03-14-performance-improvement/state.yaml` |

---

## SDD Cycle Complete

The `performance-improvement` change has been fully planned, implemented, verified, and archived. The dlexa CLI now has:

- A persistent filesystem cache eliminating redundant network I/O
- Parallel source fan-out for multi-source scalability
- Optimized hot-path allocations with benchmark-proven improvements
- Deduplicated rendering utilities in a shared package

Ready for the next change.

# Design: Parser Engine Search Migration

## Technical Approach
Adopt search-family parsing through explicit parser-engine wrappers, not by moving the existing parsing logic. New engine-native `SearchParser` implementations in `internal/parse/engine` will delegate to the current `internal/parse` live and DPD search parsers, and runtime wiring in `internal/app/wiring.go` will switch from `search.NewPipelineProvider(...)` to `search.NewEnginePipelineProvider(...)`. This satisfies the parser-engine architecture and spec while preserving current search behavior byte-for-byte at the parsing boundary.

## Architecture Decisions
| Decision | Options | Choice | Rationale |
|---|---|---|---|
| Wrapper vs move | Move parsing logic into `internal/parse/engine`; add engine-native wrappers | Add wrappers | Moving now creates larger churn, test fallout, and package-cycle pressure because `engine` already imports `parse`. Wrappers let runtime adopt the engine port immediately with minimal risk. |
| Engine-native parser shape | Generic legacy adapter only; named wrappers per surface | Named wrappers for live + DPD search | Concrete wrappers make runtime intent explicit, match ADR-0001 surface strategies, and let wiring/tests assert real engine-native adoption instead of a generic compatibility bridge. |
| Wiring migration | Keep `NewPipelineProvider`; switch to `NewEnginePipelineProvider` | Switch runtime wiring only | `internal/search/provider.go` already exposes the seam. Using it changes entrypoint plumbing without altering fetch, normalize, cache, or service orchestration. |

## Data Flow
`fetcher.Fetch` → `parseengine.LiveSearchParser` / `parseengine.DPDSearchParser` → delegated legacy parser in `internal/parse` → `[]parse.ParsedSearchRecord` → existing normalizer → `model.SearchCandidate` → existing search service aggregation.

The provider bridge remains internal to `search.PipelineProvider`; the semantic change is only that wiring now supplies engine-native parsers.

## File Changes
| File | Action | Description |
|---|---|---|
| `internal/parse/engine/live_search.go` | Create | Engine-native live-search wrapper implementing `SearchParser` via `ParseInput` delegation to `parse.NewLiveSearchParser()`. |
| `internal/parse/engine/dpd_search.go` | Create | Engine-native DPD-search wrapper implementing `SearchParser` via `ParseInput` delegation to `parse.NewDPDSearchParser()`. |
| `internal/app/wiring.go` | Modify | Wire both search providers through `NewEnginePipelineProvider(...)`. |
| `internal/app/wiring_test.go` | Modify | Assert concrete engine-native search wrappers are wired for `search` and `dpd`. |
| `internal/parse/engine/search_port_test.go` | Modify | Add passthrough tests for both named wrappers. |
| `internal/search/provider_test.go` | Modify | Keep proving engine-backed provider preserves parse/normalize flow with concrete engine parsers. |

## Interfaces / Contracts
```go
type LiveSearchParser struct { legacy *parse.LiveSearchParser }
func (p *LiveSearchParser) ParseSearch(input ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error)

type DPDSearchParser struct { legacy *parse.DPDSearchParser }
func (p *DPDSearchParser) ParseSearch(input ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error)
```
Both wrappers must forward `Ctx`, `Descriptor`, and `Document` unchanged and return legacy records/warnings/errors unchanged.

## Testing Strategy
| Layer | What to Test | Approach |
|---|---|---|
| Unit | Wrapper fidelity | Add engine tests that verify both wrappers pass through `ParseInput` fields and preserve outputs/errors exactly. |
| Unit | Parser behavior | Keep `internal/parse/live_search_test.go` and `internal/parse/dpd_search_test.go` unchanged as the behavior oracle for empty, malformed, and valid payloads. |
| Integration | Provider wiring | Update `internal/search/provider_test.go` and `internal/app/wiring_test.go` to assert engine-native parser usage and unchanged normalized output/warnings. |
| Integration | Service behavior | Keep existing `internal/search/live_service_test.go` and service tests unchanged to guard error propagation, ordering, cache behavior, and fallback semantics. |
| E2E | N/A | No new e2e layer in this slice. |

## Migration / Rollout
No staged rollout required. This is a composition-root migration only. No config, cache-key, or output-contract migration is expected.

## Open Questions
- None.

## Risks and Boundaries
Risks: parser type-name confusion between `parse` and `parseengine`; accidental helper movement causing drift; overreaching into resolver adoption.

Boundaries: do not move live/DPD parsing logic, do not touch article-family parsers, do not change normalizers, ranking, provider priority, cache behavior, fallback semantics, or problem taxonomy.

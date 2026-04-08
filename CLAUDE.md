# CLAUDE.md

## Read before planning or coding

This repo has had architecture drift between docs, OpenSpec artifacts, and the actual filesystem layout.

### Current runtime truth

- The binary entrypoint is currently centered in:
  - `cmd/dlexa/main.go`
  - `cmd/dlexa/root.go`
  - `cmd/dlexa/dpd.go`
  - `cmd/dlexa/search.go`
- The runtime surface is currently centered in:
  - `internal/app/app.go`
  - `internal/app/wiring.go`
- Shared module contracts live in:
  - `internal/modules/interfaces.go`
  - `internal/model/types.go`
  - `internal/model/search.go`
- Implemented modules live in:
  - `internal/modules/dpd/module.go`
  - `internal/modules/search/module.go`
- Agent-facing rendering/fallback behavior lives in:
  - `internal/render/envelope.go`
  - `internal/render/markdown.go`
  - `internal/render/json.go`
  - `internal/render/search_markdown.go`
  - `internal/render/search_json.go`
- Live-search and source pipelines live in:
  - `internal/fetch/live_search.go`
  - `internal/parse/live_search.go`
  - `internal/search/service.go`
  - `internal/source/pipeline.go`
  - `internal/config/static.go`

### Current documentation truth

- `docs/architecture-formal-dlexa-v2.md` = formal architecture doc with explicit current-vs-target sections
- `docs/architecture_v2_oraculo.md` = narrative / vision, but now grounded in the current thin `cmd/dlexa` + `internal/app` runtime
- `openspec/specs/*.md` = active main specs, but they may overstate completion if not reconciled with code

### Rules for future agents

1. **Treat `cmd/dlexa` as the thin CLI entrypoint and `internal/app` as the execution/composition root.**
2. If docs/specs and code disagree, code wins as runtime truth.
3. Record drift explicitly before proposing fixes.
4. Prefer these files as the first reading set:
   - `AGENTS.md`
   - `CLAUDE.md`
   - `internal/app/app.go`
   - `internal/app/wiring.go`
   - `internal/modules/interfaces.go`
   - `internal/modules/dpd/module.go`
   - `internal/modules/search/module.go`
   - `internal/render/envelope.go`
   - `openspec/specs/cli/spec.md`
   - `openspec/specs/search/spec.md`
5. **Final verification MUST be performed by the orchestrating agent itself, not by a subagent.**
6. Subagents may still be used for other phases such as proposal, spec, design, tasks, or apply when appropriate.
7. **After verify passes, the orchestrating agent MUST create the commit before reporting the change as fully verified.** Commit-time hooks and validations are part of the true verification boundary.
8. Follow the SDD workflow and active change artifacts under `openspec/changes/`. The entire SDD workflow (explore -> propose -> spec -> design -> tasks -> apply -> verify -> archive) MUST run completely automatically and proactively from start to finish without pausing for user confirmation or review between phases. Only stop for hard, unresolvable blockers. If questions arise about preferences or past discussions, search engram memory FIRST rather than asking the user. Execute the rest of the skills exactly as indicated but with ZERO user intervention.

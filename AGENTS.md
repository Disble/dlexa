## Project Definition

`dlexa` is a consultation interface for **normative linguistic doubts in Spanish**.

- Use this repo and its skills when the task is about orthographic, orthoepic/pronunciation, morphological, syntactic, or lexico-semantic doubts that fit `dlexa`'s normative consultation scope.
- Preserve normative nuance: recommendations can depend on **norma culta formal**, **current usage**, **register**, **geography**, and **communicative context**.
- Do **not** frame `dlexa` as a generic dictionary replacement, encyclopedic lookup tool, translation system, or universal lexical source.

## Project Skills

| Skill | Description | Source |
| --- | --- | --- |
| `dlexa-go-cli-lint` | Run the repository's configured Go linters whenever Go/CLI code changes need validation, choosing focused or full lint based on scope without building. | [SKILL.md](skills/dlexa-go-cli-lint/SKILL.md) |
| `dlexa-skill-updater` | Maintain the `dlexa-user` skill and its mirrors by detecting CLI drift, semantic-output drift, project-positioning drift, discovery-surface drift, and mirror-parity drift. | [SKILL.md](skills/dlexa-skill-updater/SKILL.md) |
| `dlexa-sonarqube-mcp` | Use the repository's SonarQube MCP workflow correctly, including analysis toggling, project-key lookup, and end-of-task file analysis. | [SKILL.md](skills/dlexa-sonarqube-mcp/SKILL.md) |
| `no-duplication` | Eliminate code duplication detected by SonarQube in Go test files. Load when `new_duplicated_lines_density` exceeds the quality gate threshold, when writing Go tests with repeated boilerplate (shared table rows, run-loop patterns, builder closures, golden-file assertions), or when adding test fixture files under `testdata/` that should be excluded from CPD analysis via `sonar-project.properties`. **Never repeat test struct literals across functions — extract to package-level `var`. Never repeat `t.Helper()` loops — extract to a parameterized helper. Always exclude scraped/generated `testdata/` from `sonar.cpd.exclusions`.** | [SKILL.md](.claude/skills/no-duplication/SKILL.md) |
| `dlexa-user` | Teach other LLMs when to invoke `dlexa` for normative linguistic doubts handled by its consultation surfaces, how to parse outputs, and when to redirect out-of-scope generic dictionary tasks elsewhere. | [SKILL.md](skills/dlexa-user/SKILL.md) |

## Agent Routing Notes

- Load `dlexa-user` when the job is to **use** the CLI for a normative consultation handled by `dlexa`.
- Do not load `dlexa-user` just because a prompt contains a Spanish word; first verify the task is actually a normative linguistic doubt that fits `dlexa`'s consultation scope.
- If the task is generic lexical definition, translation, encyclopedic lookup, or etymology, `dlexa` is the wrong hammer.
- When updating repo guidance or mirrored skills, keep `skills/` as canonical and `.claude/skills/` in semantic lockstep.

## Read This First — Critical Files and Reality Checks

- **Do not assume the documented target architecture already exists in code. Verify first.**
- The repo's **current binary entrypoint** lives in `cmd/dlexa`, especially:
  - `cmd/dlexa/main.go`
  - `cmd/dlexa/root.go`
  - `cmd/dlexa/dpd.go`
  - `cmd/dlexa/search.go`
- The repo's **current runtime surface** lives in `internal/app`, especially:
  - `internal/app/app.go`
  - `internal/app/wiring.go`
- The repo's **current application contracts** live in:
  - `internal/modules/interfaces.go`
  - `internal/model/types.go`
  - `internal/model/search.go`
- The repo's **current implemented modules** live in:
  - `internal/modules/dpd/module.go`
  - `internal/modules/search/module.go`
- The repo's **current agent-facing rendering and fallback contracts** live in:
  - `internal/render/envelope.go`
  - `internal/render/markdown.go`
  - `internal/render/json.go`
  - `internal/render/search_markdown.go`
  - `internal/render/search_json.go`
- The repo's **current live pipelines** live in:
  - `internal/fetch/live_search.go`
  - `internal/parse/live_search.go`
  - `internal/search/service.go`
  - `internal/config/static.go`
- The repo's **current main OpenSpec source of truth** lives in:
  - `openspec/specs/cli/spec.md`
  - `openspec/specs/search/spec.md`
  - `openspec/specs/dpd/spec.md`
  - `openspec/specs/render/spec.md`
- The repo's **current architecture decision records and index** live in:
  - `docs/ARCHITECTURE.md`
  - `docs/adrs/`
- `docs/architecture-formal-dlexa-v2.md` and `docs/architecture_v2_oraculo.md` contain both **current-state** and **target-state** discussion; when in doubt, verify against the filesystem and runtime wiring first.
- When a task involves architecture, engine boundaries, module design, parser/fetch/normalize evolution, or long-lived structural decisions, read `docs/ARCHITECTURE.md` and the relevant ADRs in `docs/adrs/` before proposing changes.
- If docs/specs/archived changes conflict with code, treat the **codebase** as the runtime truth, document the drift, and only then plan the fix.

## Repo Workflow

- The repo-level lint and pre-commit onboarding flow lives in `CONTRIBUTING.md`.
- Full-repo lint uses the repo-pinned command: `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.
- The pre-commit hook intentionally stays diff-based: `go tool --modfile=golangci-lint.mod golangci-lint run --new-from-rev=HEAD`.

## Delegation Guardrails

- When delegating bugfix or apply work to sub-agents, prompts MUST include the exact reproduction steps/commands when known.
- Those prompts MUST include both acceptance examples and rejection/negative examples; do not describe only the happy path.
- Those prompts MUST name forbidden outputs or behaviors explicitly when the bug involves false positives, misleading fallbacks, or malformed UX.
- When parser/classification behavior is involved, prompts MUST require negative regression coverage and adversarial fixture guidance when relevant.
- If the user explicitly asks the orchestrator to perform a repo-doc or instruction-file update itself, do not delegate that file edit to a sub-agent.
- **CRITICAL OVERRIDE - Auto-Pilot SDD**: The entire SDD workflow (explore -> propose -> spec -> design -> tasks -> apply -> verify -> archive) MUST run completely automatically and proactively from start to finish. THIS OVERRIDES ALL DEFAULT SKILL BEHAVIORS. You are STRICTLY FORBIDDEN from pausing to ask "should I continue?", "do you want to proceed to specs?", or waiting for approval. Present the phase summary and IMMEDIATELY trigger the next phase. Ignore simple reviews aggressively to save the user time. Ask for user input ONLY on hard, unresolvable blockers. If questions arise about preferences or past discussions, search engram memory FIRST. Execute the rest of the skills exactly as indicated but with ABSOLUTELY ZERO user intervention between phases.
- Verification is a special case: the orchestrating agent MUST perform the final verification itself and MUST NOT delegate the verify phase to a sub-agent. Other phases may still use sub-agents when appropriate.
- After verify passes, the orchestrating agent MUST create the commit before reporting verify as fully complete. The commit's own hooks/validations are part of the real verification boundary and save the user an extra round-trip.

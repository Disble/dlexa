## Project Definition

`dlexa` is a consultation interface for **DPD-covered normative linguistic doubts in Spanish**.

- Use this repo and its skills when the task is about orthographic, orthoepic/pronunciation, morphological, syntactic, or lexico-semantic doubts that fit the DPD model.
- Preserve DPD nuance: recommendations can depend on **norma culta formal**, **current usage**, **register**, **geography**, and **communicative context**.
- Do **not** frame `dlexa` as a generic dictionary replacement, encyclopedic lookup tool, translation system, or universal lexical source.

## Project Skills

| Skill | Description | Source |
| --- | --- | --- |
| `dlexa-go-cli-lint` | Run the repository's configured Go linters whenever Go/CLI code changes need validation, choosing focused or full lint based on scope without building. | [SKILL.md](skills/dlexa-go-cli-lint/SKILL.md) |
| `dlexa-skill-updater` | Maintain the `dlexa-user` skill and its mirrors by detecting CLI drift, semantic-output drift, project-positioning drift, discovery-surface drift, and mirror-parity drift. | [SKILL.md](skills/dlexa-skill-updater/SKILL.md) |
| `dlexa-sonarqube-mcp` | Use the repository's SonarQube MCP workflow correctly, including analysis toggling, project-key lookup, and end-of-task file analysis. | [SKILL.md](skills/dlexa-sonarqube-mcp/SKILL.md) |
| `dlexa-user` | Teach other LLMs when to invoke `dlexa` for DPD-style normative doubts, how to parse outputs, and when to redirect out-of-scope generic dictionary tasks elsewhere. | [SKILL.md](skills/dlexa-user/SKILL.md) |

## Agent Routing Notes

- Load `dlexa-user` when the job is to **use** the CLI for a DPD-fit consultation.
- Do not load `dlexa-user` just because a prompt contains a Spanish word; first verify the task is actually a DPD-style normative doubt.
- If the task is generic lexical definition, translation, encyclopedic lookup, or etymology, `dlexa` is the wrong hammer.
- When updating repo guidance or mirrored skills, keep `skills/` as canonical and `.claude/skills/` in semantic lockstep.

## Repo Workflow

- The repo-level lint and pre-commit onboarding flow lives in `CONTRIBUTING.md`.
- Full-repo lint uses the repo-pinned command: `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.
- The pre-commit hook intentionally stays diff-based: `go tool --modfile=golangci-lint.mod golangci-lint run --new-from-rev=HEAD`.

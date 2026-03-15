## Project Skills

| Skill | Description | Source |
| --- | --- | --- |
| `dlexa-go-cli-lint` | Run the repository's configured Go linters whenever Go/CLI code changes need validation, choosing focused or full lint based on scope without building. | [SKILL.md](skills/dlexa-go-cli-lint/SKILL.md) |
| `dlexa-skill-updater` | Automate maintenance of the dlexa-user skill by detecting CLI interface changes and regenerating documentation | [SKILL.md](skills/dlexa-skill-updater/SKILL.md) |
| `dlexa-sonarqube-mcp` | Use the repository's SonarQube MCP workflow correctly, including analysis toggling, project-key lookup, and end-of-task file analysis. | [SKILL.md](skills/dlexa-sonarqube-mcp/SKILL.md) |
| `dlexa-user` | Invoke and parse dlexa CLI (user manual for LLM agents that need to USE dlexa commands, parse outputs, and troubleshoot errors) | [SKILL.md](skills/dlexa-user/SKILL.md) |

## Repo Workflow

- The repo-level lint and pre-commit onboarding flow lives in `CONTRIBUTING.md`.
- Full-repo lint uses the repo-pinned command: `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.
- The pre-commit hook intentionally stays diff-based: `go tool --modfile=golangci-lint.mod golangci-lint run --new-from-rev=HEAD`.

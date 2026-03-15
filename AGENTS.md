## Project Skills

| Skill | Description | Source |
| --- | --- | --- |
| `dlexa-go-cli-lint` | Run the repository's configured Go linters whenever Go/CLI code changes need validation, choosing focused or full lint based on scope without building. | [SKILL.md](skills/dlexa-go-cli-lint/SKILL.md) |
| `dlexa-sonarqube-mcp` | Use the repository's SonarQube MCP workflow correctly, including analysis toggling, project-key lookup, and end-of-task file analysis. | [SKILL.md](skills/dlexa-sonarqube-mcp/SKILL.md) |

## Repo Workflow

- The repo-level lint and pre-commit onboarding flow lives in `CONTRIBUTING.md`.
- Use the repo-pinned linter entrypoint: `go tool -modfile=golangci-lint.mod golangci-lint`.

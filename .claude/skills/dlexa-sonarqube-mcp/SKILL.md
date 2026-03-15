---
name: dlexa-sonarqube-mcp
description: >
  Apply the repository's SonarQube MCP workflow correctly and in the right order.
  Trigger: Load when a task uses SonarQube MCP tools, project-key lookup, snippet analysis, or post-edit file analysis in this repo.
license: Apache-2.0
metadata:
  author: Disble
  version: "1.0"
---

## When to Use

- When a task uses SonarQube MCP tools in this repository.
- When the user mentions SonarQube project keys, issue lookup, or snippet analysis.
- When code changes finish and SonarQube file analysis should run on the edited files.
- When SonarQube authentication or project lookup errors appear.

## Critical Patterns

- MUST disable automatic analysis at task start with `toggle_automatic_analysis` if that tool exists.
- MUST analyze created or modified files at the end with `analyze_file_list` if that tool exists.
- MUST re-enable automatic analysis at task end with `toggle_automatic_analysis` if that tool exists.
- MUST look up project keys with `search_my_sonarqube_projects` before using one from memory or user wording.
- MUST NOT verify freshly fixed issues with `search_sonar_issues_in_projects`; SonarQube results can lag behind the code change.
- SHOULD include the branch when the user says the work is on a feature branch and the tool supports branch-specific context.
- SHOULD detect the code language from syntax before requesting analysis; if unclear, ask or make the narrowest reasonable guess.
- SHOULD send full file content when snippet analysis quality matters; partial snippets are weaker than real file context.
- MUST treat `Not authorized` as a likely wrong-token problem and check for a USER token, not a project token.

## Code Examples

```text
Task start
1. toggle_automatic_analysis(disabled=true)
2. search_my_sonarqube_projects(query="dlexa")
3. Run the requested SonarQube analysis flow
4. analyze_file_list(files=["internal/query/service.go"])
5. toggle_automatic_analysis(disabled=false)
```

```text
Project key lookup
- User says: "check project dlexa-cli"
- Correct move: search_my_sonarqube_projects(query="dlexa-cli")
- Wrong move: assume `dlexa-cli` is the exact project key
```

```text
Fresh fix caveat
- After fixing a Sonar issue locally, do not call search_sonar_issues_in_projects to confirm the fix immediately.
- Explain that server-side issue data may not reflect the new code yet.
```

## Commands

```text
toggle_automatic_analysis
search_my_sonarqube_projects
analyze_file_list
search_sonar_issues_in_projects
```

## Resources

- **Documentation**: See [references/sonarqube-mcp.md](references/sonarqube-mcp.md) for the local source instruction file and the non-duplicated details.

# CLI Specification

## Purpose

Defines the active command surface, routing, and help behavior of the `dlexa` CLI. The binary entrypoint lives in `cmd/dlexa`, while command execution delegates to `internal/app` and the `internal/modules` registry.

## Requirements

### Requirement: Explicit Command Tree

The CLI MUST expose a thin formal command tree from `cmd/dlexa` and MUST delegate execution to `internal/app` rather than embedding business logic in the command layer.

#### Scenario: Valid commands execute successfully

- GIVEN the `dlexa` CLI is invoked
- WHEN the agent provides valid subcommands (`search`, `dpd`)
- THEN the system MUST route execution to the corresponding module
- AND return a structured response payload

#### Scenario: Root command defaults to DPD lookup

- GIVEN the `dlexa` CLI is invoked without a specific subcommand
- WHEN the agent provides a query argument (e.g., `dlexa basto`)
- THEN the system MUST implicitly route the query to the `dpd` module

### Requirement: Agent-Optimized Markdown Help

The CLI MUST provide help documentation formatted in Markdown, tailored for LLM consumption, providing clear syntax, actionable examples, and fallback strategies.

#### Scenario: Agent requests help

- GIVEN the `dlexa` CLI is invoked
- WHEN the agent passes the `--help` flag to any command or subcommand
- THEN the system MUST render the help output in Markdown
- AND the output MUST include copiable examples and error recovery suggestions

#### Scenario: Syntax failure shows help

- GIVEN the `dlexa` CLI is invoked
- WHEN the agent provides invalid flags or incorrect arguments
- THEN the system MUST return a Nivel 1 (Syntax) fallback error
- AND suggest using `--help` to view correct usage

### Requirement: Search Command Remains the Explicit Gateway Entry

The CLI MUST keep `search` as the explicit entrypoint for semantic discovery, while preserving the existing root default-to-DPD behavior.

#### Scenario: Root query still defaults to DPD

- GIVEN the CLI is invoked without an explicit subcommand
- WHEN the user provides a query argument such as `dlexa bien`
- THEN the system MUST route the request to the `dpd` module
- AND the system MUST NOT automatically invoke the live semantic search gateway

#### Scenario: Search command invokes the semantic gateway

- GIVEN the CLI is invoked as `dlexa search <query>`
- WHEN the command arguments are valid
- THEN the system MUST route execution to the live semantic search gateway
- AND the system MUST return curated search output rather than a direct DPD article lookup

### Requirement: Search Output Communicates Safe Next Steps

The CLI MUST render search results as safe next-step guidance for agents and users.

#### Scenario: Search output includes copyable command suggestions

- GIVEN `dlexa search <query>` returns curated candidates mapped to known surfaces
- WHEN the CLI renders the successful response
- THEN the output MUST include literal, copyable `dlexa ...` next-step suggestions
- AND the output MUST distinguish those suggestions from the current command being executed

#### Scenario: Search output preserves unmapped fallback guidance

- GIVEN `dlexa search <query>` returns curated candidates that cannot be mapped into known command suggestions
- WHEN the CLI renders the successful response
- THEN the output MUST include a safe fallback representation for those candidates
- AND the output MUST NOT invent unsupported command syntax

### Requirement: Active Spec Matches Registered Commands

The active CLI spec MUST describe only commands that are actually registered in the current command tree.

#### Scenario: Search suggests a deferred destination command

- GIVEN the semantic gateway suggests a literal command such as `dlexa espanol-al-dia <slug>`
- WHEN that destination command is not registered in the current CLI tree
- THEN the gateway suggestion MUST be valid as guidance
- AND the active CLI spec MUST NOT describe that destination as an implemented subcommand until it is actually wired

#### Scenario: Existing command tree remains constrained

- GIVEN the CLI command tree after this change
- WHEN the available subcommands are inspected
- THEN the public destination command surface MUST remain limited to the commands already supported by the current CLI contract
- AND root default-to-DPD behavior MUST remain unchanged

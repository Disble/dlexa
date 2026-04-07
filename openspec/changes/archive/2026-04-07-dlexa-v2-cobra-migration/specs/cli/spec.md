# CLI Specification

## Purpose

Defines the command surface, routing, and help behavior of the `dlexa` CLI using the `spf13/cobra` framework, transforming it into an agent-optimized gateway.

## Requirements

### Requirement: Explicit Command Tree

The CLI MUST provide a formal command tree powered by `spf13/cobra` rather than manual `flag` parsing.

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

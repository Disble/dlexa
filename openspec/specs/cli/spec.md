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

### Requirement: Module-Specific Default Sources

The application runtime MUST distinguish between default sources for lookup and default sources for search, preventing global drift where lookup defaults are incorrectly applied to the search module.

#### Scenario: Injecting search defaults

- GIVEN the `search` module is invoked via `dlexa search <query>` without explicit sources
- WHEN `App.ExecuteModule` applies configuration defaults
- THEN the runtime MUST apply the search-specific default sources
- AND MUST NOT apply the lookup-specific default source (`dpd`)

#### Scenario: Injecting lookup defaults

- GIVEN the `dpd` module is invoked via `dlexa <query>` without explicit sources
- WHEN `App.ExecuteModule` applies configuration defaults
- THEN the runtime MUST apply the lookup-specific default source (`dpd`)

### Requirement: Preservation of Current CLI Surface

The external CLI surface and fallback semantics MUST remain unchanged. The `search` command remains the explicit gateway entry, and root queries default to DPD.

#### Scenario: CLI surface remains stable

- GIVEN the CLI is invoked
- WHEN a user runs existing commands (`dlexa basto`, `dlexa search basto`)
- THEN the commands MUST route correctly to their respective modules using the correct provider defaults

### Requirement: Independent DPD Search Surface

The CLI MUST preserve an independent DPD search surface in addition to the federated `search` gateway.

#### Scenario: Running the dedicated DPD search command

- GIVEN the CLI is invoked as `dlexa dpd search <termino-de-busqueda>`
- WHEN the command arguments are valid
- THEN the system MUST route execution to the `search` module
- AND it MUST force the provider selection to the DPD entry-discovery source only
- AND it MUST preserve the `search` response contract rather than doing a direct article lookup

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

### Requirement: Format Validation at Runtime Boundary

The runtime MUST validate `req.Format` against the set of registered formats after applying config defaults. Unsupported formats MUST produce a Nivel 1 (Syntax) fallback with explicit guidance, not a raw error.

#### Scenario: Valid format passes through

- GIVEN the CLI is invoked with `--format markdown` or `--format json`
- WHEN the format reaches `App.ExecuteModule`
- THEN the request MUST proceed to module execution without error

#### Scenario: Empty format defaults and passes through

- GIVEN the CLI is invoked without `--format`
- WHEN the format is empty at `App.ExecuteModule`
- THEN the runtime MUST apply the config default format
- AND the request MUST proceed to module execution without error

#### Scenario: Invalid format returns structured syntax fallback

- GIVEN the CLI is invoked with `--format yaml` (or any unsupported value)
- WHEN the format reaches `App.ExecuteModule` after defaults are applied
- THEN the runtime MUST return a Nivel 1 (Syntax) fallback
- AND the fallback message MUST indicate the supported formats
- AND the fallback MUST suggest using `--help`

### Requirement: Command Surface Black-Box Tests

The Cobra command surface in `cmd/dlexa` MUST have black-box tests exercising routing, flags, help, and error paths through `executeRootCommand` with a mock `runtimeRunner`.

#### Scenario: Root with query routes to DPD

- GIVEN the CLI is invoked as `dlexa basto`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with module `"dpd"` and query `"basto"`

#### Scenario: Subcommand routing is deterministic

- GIVEN the CLI is invoked as `dlexa dpd solo` or `dlexa search solo o sólo`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with the correct module and joined query

#### Scenario: Help, version, and doctor flags route correctly

- GIVEN the CLI is invoked with `--help`, `--version`, or `--doctor`
- WHEN `executeRootCommand` processes the arguments
- THEN the corresponding method (`RenderHelp`, `PrintVersion`, `RunDoctor`) MUST be called

#### Scenario: Missing subcommand query triggers syntax error

- GIVEN the CLI is invoked as `dlexa dpd` or `dlexa search` (no query)
- WHEN `executeRootCommand` processes the arguments
- THEN `HandleSyntaxError` MUST be called

#### Scenario: Format and no-cache flags propagate to module request

- GIVEN the CLI is invoked with `--format json` or `--no-cache`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST receive the corresponding `Format` or `NoCache` values

### Requirement: Source-Scoped Search Execution

The `search` command MUST expose a `--source` flag that allows agents to restrict federation to a named subset of registered providers.

#### Scenario: Source flag scopes federation to a single provider

- GIVEN the CLI is invoked as `dlexa search --source dpd <query>`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with `Sources: ["dpd"]`
- AND the search module MUST NOT federate to any other provider

#### Scenario: Multiple source flags accumulate providers

- GIVEN the CLI is invoked as `dlexa search --source dpd --source search <query>`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with `Sources: ["dpd", "search"]`

#### Scenario: Omitting source flag federates all providers

- GIVEN the CLI is invoked as `dlexa search <query>` without `--source`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with `Sources` empty or nil
- AND the search module MUST apply its default federation policy

#### Scenario: Unknown source returns a syntax fallback

- GIVEN the CLI is invoked as `dlexa search --source unknown <query>`
- WHEN `executeRootCommand` processes the arguments
- THEN `HandleSyntaxError` MUST be called
- AND the error message MUST name the unknown source and list the valid ones
- AND `RunModule` MUST NOT be called

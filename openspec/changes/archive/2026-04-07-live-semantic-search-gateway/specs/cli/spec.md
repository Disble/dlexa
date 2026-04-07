# CLI Specification (Delta)

## ADDED Requirements

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

### Requirement: No New Destination Commands in This Change

This change MUST NOT expand the Cobra tree with new destination content commands beyond the existing `search` and `dpd` surface.

#### Scenario: Search suggests a deferred destination command

- GIVEN the semantic gateway suggests a literal command such as `dlexa espanol-al-dia <slug>`
- WHEN this change is considered complete
- THEN the gateway suggestion MUST be valid as guidance
- AND the existence of that suggestion MUST NOT require the destination Cobra command to be implemented in this change

#### Scenario: Existing command tree remains constrained

- GIVEN the CLI command tree after this change
- WHEN the available Cobra subcommands are inspected
- THEN the public destination command surface MUST remain limited to the commands already supported by the current CLI contract
- AND root default-to-DPD behavior MUST remain unchanged

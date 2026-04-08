# Delta for CLI

## MODIFIED Requirements

### Requirement: Explicit Command Tree

The CLI MUST provide a formal command surface powered by the `internal/app` runtime and `internal/modules` registry.
(Previously: The CLI MUST provide a formal command tree powered by `spf13/cobra` rather than manual `flag` parsing.)

#### Scenario: Valid commands execute successfully

- GIVEN the CLI is invoked
- WHEN the agent provides valid subcommands (`search`, `dpd`)
- THEN the `app.App` MUST route execution to the corresponding module
- AND return a structured response payload

## REMOVED Requirements

### Requirement: No New Destination Commands in This Change

(Reason: This requirement was tied to an abandoned Cobra migration and does not apply to the `internal/app` runtime truth.)
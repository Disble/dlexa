# Delta for CLI — cli-contract-hardening

## NEW Requirement: Format Validation at Runtime Boundary

The runtime MUST validate `req.Format` against the set of registered formats after applying config defaults. Unsupported formats MUST produce a Nivel 1 (Syntax) fallback with explicit guidance, not a raw error.

### Scenario: Valid format passes through

- GIVEN the CLI is invoked with `--format markdown` or `--format json`
- WHEN the format reaches `App.ExecuteModule`
- THEN the request MUST proceed to module execution without error

### Scenario: Empty format defaults and passes through

- GIVEN the CLI is invoked without `--format`
- WHEN the format is empty at `App.ExecuteModule`
- THEN the runtime MUST apply the config default format
- AND the request MUST proceed to module execution without error

### Scenario: Invalid format returns structured syntax fallback

- GIVEN the CLI is invoked with `--format yaml` (or any unsupported value)
- WHEN the format reaches `App.ExecuteModule` after defaults are applied
- THEN the runtime MUST return a Nivel 1 (Syntax) fallback
- AND the fallback message MUST indicate the supported formats
- AND the fallback MUST suggest using `--help`

## NEW Requirement: Command Surface Black-Box Tests

The Cobra command surface in `cmd/dlexa` MUST have black-box tests exercising routing, flags, help, and error paths through `executeRootCommand` with a mock `runtimeRunner`.

### Scenario: Root with query routes to DPD

- GIVEN the CLI is invoked as `dlexa basto`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with module `"dpd"` and query `"basto"`

### Scenario: DPD subcommand routes to DPD

- GIVEN the CLI is invoked as `dlexa dpd solo`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with module `"dpd"` and query `"solo"`

### Scenario: Search subcommand routes to search

- GIVEN the CLI is invoked as `dlexa search solo o sólo`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with module `"search"` and query `"solo o sólo"`

### Scenario: --help flag renders help

- GIVEN the CLI is invoked as `dlexa --help`
- WHEN `executeRootCommand` processes the arguments
- THEN `RenderHelp` MUST be called
- AND `RunModule` MUST NOT be called

### Scenario: --version flag prints version

- GIVEN the CLI is invoked as `dlexa --version`
- WHEN `executeRootCommand` processes the arguments
- THEN `PrintVersion` MUST be called

### Scenario: --doctor flag runs doctor

- GIVEN the CLI is invoked as `dlexa --doctor`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunDoctor` MUST be called

### Scenario: No arguments shows help

- GIVEN the CLI is invoked as `dlexa` (no arguments)
- WHEN `executeRootCommand` processes the arguments
- THEN `RenderHelp` MUST be called

### Scenario: DPD subcommand without query returns syntax error

- GIVEN the CLI is invoked as `dlexa dpd` (no query)
- WHEN `executeRootCommand` processes the arguments
- THEN `HandleSyntaxError` MUST be called

### Scenario: Search subcommand without query returns syntax error

- GIVEN the CLI is invoked as `dlexa search` (no query)
- WHEN `executeRootCommand` processes the arguments
- THEN `HandleSyntaxError` MUST be called

### Scenario: Format flag propagates to module request

- GIVEN the CLI is invoked as `dlexa dpd basto --format json`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with `Format` equal to `"json"`

### Scenario: No-cache flag propagates to module request

- GIVEN the CLI is invoked as `dlexa dpd basto --no-cache`
- WHEN `executeRootCommand` processes the arguments
- THEN `RunModule` MUST be called with `NoCache` equal to `true`

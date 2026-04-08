# Design: cli-contract-hardening

## Architecture Decision

This change adds two capabilities without modifying the existing architecture:

1. **Black-box command tests** in `cmd/dlexa/root_test.go` using a mock `runtimeRunner`
2. **Format validation** in `App.ExecuteModule` before module dispatch

### Testing Strategy

The command surface tests use `executeRootCommand` as the system-under-test. A `mockRuntime` struct implements `runtimeRunner` and records which methods were called, with what arguments, enabling assertion without real modules or network.

```
executeRootCommand(ctx, mockRuntime, stdout, stderr, args)
  → Cobra parses args
  → routes to root/dpd/search RunE
  → calls mockRuntime.RunModule / RenderHelp / PrintVersion / RunDoctor / HandleSyntaxError
  → test asserts: which method was called, with which module/query/format/noCache
```

The mock records calls in a slice of `call` structs: `{method, module, query, format, noCache}`. Tests are table-driven.

### Format Validation

Current flow:
```
App.ExecuteModule → loadConfig → apply defaults → resolve module → module.Execute → render
```

New flow inserts validation after defaults:
```
App.ExecuteModule → loadConfig → apply defaults → VALIDATE FORMAT → resolve module → ...
```

The validation needs to know which formats are valid. Two options:

**Option A**: Hardcode `{"markdown", "json"}` in `App.ExecuteModule`.
**Option B**: Add a `Formats() []string` method to the envelope renderer or accept a `ValidFormats` set at construction.

**Decision**: Option A. The valid formats are already implicitly defined by the render registries wired in `wiring.go`, but `App` only holds an `EnvelopeRenderer`, not the per-module render registries. Adding a formats method would require changing the `EnvelopeRenderer` interface or `App` struct for a validation that is fundamentally a CLI contract concern. Hardcoding `{"markdown", "json"}` matches the registered renderers exactly, is easy to test, and keeps the change minimal. If a third format is added later, the constant must be updated — but that's a feature change that would come with its own SDD.

### Files Touched

| File | Action |
|------|--------|
| `cmd/dlexa/root_test.go` | NEW — black-box command surface tests |
| `internal/app/app.go` | MODIFY — add format validation in `ExecuteModule` |
| `internal/app/app_test.go` | MODIFY — add test for invalid format fallback |

### Risks Mitigated

- Mock drift: the mock implements the same `runtimeRunner` interface defined in `root.go`, so compiler catches any signature mismatch.
- Default format break: validation runs AFTER defaults are applied, so empty `--format` still works.

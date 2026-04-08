# Tasks: cli-contract-hardening

## Task 1: Create mock runtimeRunner and test scaffold

- [ ] Create `cmd/dlexa/root_test.go`
- [ ] Define `mockRuntime` implementing `runtimeRunner` with call recording
- [ ] Define `call` struct: `{method, module, query, format, noCache}`
- [ ] Write a helper to extract the last call from the mock
- [ ] RED: Write `TestRootCommand_QueryDefaultsToDPD` — asserts `RunModule("dpd", "basto")` — fails because it compiles but nothing is broken yet (this test validates wiring)

## Task 2: Command routing tests

- [ ] RED/GREEN: `TestRootCommand_DPDSubcommand` — `dlexa dpd solo` → `RunModule("dpd", "solo")`
- [ ] RED/GREEN: `TestRootCommand_SearchSubcommand` — `dlexa search solo o sólo` → `RunModule("search", "solo o sólo")`
- [ ] RED/GREEN: `TestRootCommand_HelpFlag` — `dlexa --help` → `RenderHelp` called, `RunModule` NOT called
- [ ] RED/GREEN: `TestRootCommand_VersionFlag` — `dlexa --version` → `PrintVersion` called
- [ ] RED/GREEN: `TestRootCommand_DoctorFlag` — `dlexa --doctor` → `RunDoctor` called
- [ ] RED/GREEN: `TestRootCommand_NoArgs` — `dlexa` (no args) → `RenderHelp` called

## Task 3: Error path tests

- [ ] RED/GREEN: `TestRootCommand_DPDNoQuery` — `dlexa dpd` → `HandleSyntaxError` called
- [ ] RED/GREEN: `TestRootCommand_SearchNoQuery` — `dlexa search` → `HandleSyntaxError` called

## Task 4: Flag propagation tests

- [ ] RED/GREEN: `TestRootCommand_FormatFlag` — `dlexa dpd basto --format json` → `RunModule` with `Format:"json"`
- [ ] RED/GREEN: `TestRootCommand_NoCacheFlag` — `dlexa dpd basto --no-cache` → `RunModule` with `NoCache:true`

## Task 5: Format validation in App.ExecuteModule

- [ ] RED: Write `TestExecuteModule_InvalidFormat` in `internal/app/app_test.go` — calls `ExecuteModule` with `Format:"yaml"` after config defaults, expects `HandleSyntaxError` behavior (error or fallback indicating invalid format)
- [ ] GREEN: Add format validation in `App.ExecuteModule` after default application — if format not in `{"markdown", "json"}`, return `HandleSyntaxError`
- [ ] RED/GREEN: `TestRootCommand_InvalidFormat` in `cmd/dlexa/root_test.go` — `dlexa dpd basto --format yaml` → `HandleSyntaxError` called (or RunModule called and the app-level validation catches it)
- [ ] REFACTOR: Extract valid formats constant if needed

## Task 6: Full suite pass

- [ ] `go test ./cmd/dlexa/...` passes
- [ ] `go test ./...` passes
- [ ] `golangci-lint run ./...` passes

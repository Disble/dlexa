# Contributing

## Pre-commit Setup

This repository checks Go changes with a local `pre-commit` hook that runs the repo-pinned `golangci-lint` toolchain.

### What you need installed

- `git`
- `pre-commit` 4.x
- `go` with toolchain support enabled

The application module targets Go 1.22, but the lint tool module in `golangci-lint.mod` is currently declared with `go 1.26.1`.

- If your Go install uses `GOTOOLCHAIN=auto`, the first lint run can fetch that toolchain automatically.
- If your environment disables toolchain downloads, install a compatible Go toolchain locally before running the hook.

The hook and manual lint commands use the repo-local tool module instead of a globally installed binary.

```bash
go tool --modfile=golangci-lint.mod golangci-lint run --new-from-rev=HEAD
go tool --modfile=golangci-lint.mod golangci-lint run ./...
```

Use the first command for the diff-based pre-commit behavior and the second for canonical full-repo lint.

### Enable hooks after cloning

Run this once from the repository root:

```bash
pre-commit install
```

That writes `.git/hooks/pre-commit` for your local clone. Hooks do not auto-install on clone.

### Validate the setup

```bash
pre-commit validate-config
go tool --modfile=golangci-lint.mod golangci-lint version
```

### Normal workflow

The hook runs automatically on `git commit` and checks changed Go code against `HEAD`.

You can run the same hook manually:

```bash
pre-commit run golangci-lint --files path/to/file.go
pre-commit run golangci-lint --all-files
```

`pre-commit run golangci-lint --all-files` still invokes the same diff-based hook entry; it is not the canonical full-repo lint command.

### Full manual lint

Use this when you want repository-wide validation instead of diff-only checks:

```bash
go tool --modfile=golangci-lint.mod golangci-lint run ./...
```

Use narrower package scopes while iterating if needed:

```bash
go tool --modfile=golangci-lint.mod golangci-lint run ./internal/query/...
go tool --modfile=golangci-lint.mod golangci-lint run ./cmd/dlexa/...
```

## Limitations and Tradeoffs

- The hook is `language: system`, so each contributor still needs local `pre-commit` and `go` installed.
- The hook does not install tooling for you; it only uses the toolchain already available on your machine.
- The hook runs `go tool --modfile=golangci-lint.mod golangci-lint run --new-from-rev=HEAD`, so it is optimized for commit-time feedback on changes, not for auditing the whole repository.
- `pre-commit run golangci-lint --all-files` still executes the same diff-oriented hook entry. For a true full-repo check, use `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.
- The first lint run on a new machine may download the pinned `golangci-lint` toolchain and its modules through the Go toolchain.
- Because the repo uses a local hook instead of a mirrored remote hook, reproducibility comes from the checked-in Go module and config, not from `pre-commit` managing an isolated hook environment.

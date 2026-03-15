# Local Lint Tooling Reference

This skill is grounded in these repository files:

- `/.golangci.yml` - enables the active linter set for the repo.
- `/golangci-lint.mod` - pins the repo-local `golangci-lint` toolchain used by `go tool --modfile=golangci-lint.mod golangci-lint run ./...` for full-repo lint and `go tool --modfile=golangci-lint.mod golangci-lint run --new-from-rev=HEAD` for the diff-based hook.

Current enabled linters from `/.golangci.yml`:

- `bodyclose`
- `prealloc`
- `gocritic`
- `gosec`
- `errcheck`
- `staticcheck`
- `revive`
- `govet`

Operational rule:

- Use the repo-local tool module and the checked-in config together. Do not replace them with a global binary, alternate config, or build step.

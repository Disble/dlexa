# Archive Report: duda-linguistica-article-surface

## Change

- **Name**: `duda-linguistica-article-surface`
- **Archive date**: `2026-04-09`
- **Verify verdict**: `PASS`
- **Critical blockers**: None

## Scope Closed

- Implemented `duda-linguistica` as an executable article-family module.
- Kept `search` as the discovery layer.
- Left `noticia` deferred.

## Spec Sync Summary

| Domain | Action | Details |
|---|---|---|
| `cli` | Updated | Added the executable `duda-linguistica` command and module-specific default-source behavior. |
| `search` | Updated | Changed search truthfulness so implemented `duda-linguistica` suggestions are executable while deferred guidance remains for unimplemented surfaces. |

## Verification Notes

- Full `go test ./...` passed.
- Full repo lint passed with `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.
- The implementation reached commit-ready state before archive recording.

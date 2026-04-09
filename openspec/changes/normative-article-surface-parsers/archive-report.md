# Archive Report: normative-article-surface-parsers

## Change

- **Name**: `normative-article-surface-parsers`
- **Archive date**: `2026-04-09`
- **Verify verdict**: `PASS`
- **Critical blockers**: None

## Scope Closed

- Implemented only the `espanol-al-dia` surface as an executable article-family module.
- Kept `search` as the discovery layer.
- Left `duda-linguistica` deferred and `noticia` out of scope.

## Spec Sync Summary

| Domain | Action | Details |
|---|---|---|
| `cli` | Updated | Added the executable `espanol-al-dia` command and module-specific default-source behavior. |
| `search` | Updated | Changed search truthfulness so implemented `espanol-al-dia` suggestions are executable and help text reflects mixed executable/deferred outcomes. |

## Verification Notes

- Full `go test ./...` passed.
- Full repo lint passed with `go tool --modfile=golangci-lint.mod golangci-lint run ./...`.
- The implementation is ready to archive immediately after commit.

## Archive Destination

- `openspec/changes/archive/2026-04-09-normative-article-surface-parsers/`

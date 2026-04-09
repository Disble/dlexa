# Archive Report: noticia-faq-surface

## Change

- **Name**: `noticia-faq-surface`
- **Archive date**: `2026-04-09`
- **Verify verdict**: `PASS`

## Scope Closed

- Simplified noticia acceptance to the editorial FAQ prefix.
- Implemented executable `noticia` support for FAQ pages.
- Aligned search, CLI docs, and runtime truth.

## Spec Sync Summary

| Domain | Action | Details |
|---|---|---|
| `cli` | Updated | Added `noticia` as a registered executable subcommand. |
| `search` | Updated | FAQ-style noticia candidates now rescue on prefix and map to executable `dlexa noticia <slug>` suggestions. |

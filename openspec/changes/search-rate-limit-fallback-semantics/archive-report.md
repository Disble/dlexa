# Archive Report: search-rate-limit-fallback-semantics

## Change

- **Name**: `search-rate-limit-fallback-semantics`
- **Archive date**: `2026-04-11`
- **Verify verdict**: `PASS`

## Scope Closed

- Added explicit search rate-limit taxonomy at the problem and fallback layers.
- Classified live-search HTTP 429 and governed cooldown rejections as rate-limited outcomes.
- Preserved partial-success federation while making all-provider rate limits deterministic and explicit.

## Spec Sync Summary

| Domain | Action | Details |
|---|---|---|
| `search` | Updated | Rate-limited search failures and all-provider rate limits are now explicit runtime behaviors instead of generic upstream failures. |

# Archive Report: search-provider-governance-parity

## Change

- **Name**: `search-provider-governance-parity`
- **Archive date**: `2026-04-11`
- **Verify verdict**: `PASS`

## Scope Closed

- Applied governed 429/cooldown behavior to the DPD search provider.
- Added truthful DPD-specific rate-limit taxonomy/messages for `/srv/keys` fetches.
- Removed resilience asymmetry between the default federated `search` and `dpd` providers.

## Spec Sync Summary

| Domain | Action | Details |
|---|---|---|
| `search` | Updated | Resilient upstream fetching now applies symmetrically across both default federated providers, including DPD `/srv/keys`. |

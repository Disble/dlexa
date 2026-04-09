# Proposal: Duda Lingüística Article Surface

## Intent

Expand `dlexa`'s executable non-DPD consultation surfaces by onboarding `duda-linguistica` onto the existing article-family parser engine, keeping `search` as discovery while turning a previously deferred command suggestion into a real CLI destination.

## Scope

### In Scope
- Build a dedicated `duda-linguistica` article fetcher, parser, engine wrapper, normalizer, module, and CLI command.
- Wire the surface through the existing lookup/query/render pipeline.
- Update runtime defaults so `dlexa duda-linguistica <slug>` always resolves against its own source.
- Update `search` truthfulness so mapped `duda-linguistica` results stop rendering as deferred guidance.
- Sync durable CLI/search documentation to match runtime truth.

### Out of Scope
- `noticia` surface execution or policy gating.
- Cross-surface free-text lookup outside `search`.
- Any change to the root default-to-DPD behavior.

## Approach

Reuse the proven article-family pattern already used by `dpd` and `espanol-al-dia`:
1. Add a surface-specific fetcher targeting `/duda-linguistica/<slug>`.
2. Parse the verified `container` + `news-title` + `pt-4` article shell into article-family parsed structures.
3. Normalize that parsed article into shared `model.Entry` and `model.Article` output.
4. Expose the new module in Cobra and the composition root.
5. Mark `duda-linguistica` search candidates as executable.

## Success Criteria

- [x] Users can execute `dlexa duda-linguistica <slug>` and receive parsed lookup output.
- [x] `search` marks mapped `duda-linguistica` suggestions as executable (`Deferred: false`).
- [x] Root DPD behavior and existing executable surfaces remain unchanged.
- [x] `noticia` remains deferred.

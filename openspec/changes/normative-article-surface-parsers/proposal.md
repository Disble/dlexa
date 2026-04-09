# Proposal: Normative Article Surface Parsers

## Intent

Expand `dlexa`'s consultation capabilities by onboarding the DPD-aligned normative linguistic surface `espanol-al-dia` onto the unified parser engine. This final roadmap slice keeps `search` as discovery while adding one executable non-DPD consultation surface without polluting the core DPD lookup.

## Scope

### In Scope
- Build a new `ArticleParser` strategy for `espanol-al-dia` in `internal/parse/engine`.
- Create the corresponding module wrapper in `internal/modules/espanolaldia`.
- Add fetch/normalize layers mapped to the `espanol-al-dia` URL structure and HTML layout.
- Add the new CLI command `dlexa espanol-al-dia <slug>`.
- Wire dependencies in `internal/app/wiring.go`.
- Update `search` truthfulness so implemented `espanol-al-dia` destinations stop rendering as deferred guidance.

### Out of Scope
- `duda-linguistica` surface parser and command. This remains deferred for a future slice.
- `noticia` surface parser and command. This is explicitly deferred because `noticia` content requires complex, higher-layer policy gating to determine if an article fits the product's normative constraints.
- Cross-surface universal search (each source must be queried explicitly via its command).

## Approach

We will use the article-family parser-engine path already established in the previous slices.
1. Implement `EspanolAlDiaParser` adhering to the `ArticleParser` interface.
2. Wire the parser through the existing lookup/source pipeline.
3. Build the fetch configuration to target `/espanol-al-dia/<slug>`.
4. Build the normalizer to map parsed structures to the shared article model.
5. Expose a new top-level Cobra command in `cmd/dlexa`.
6. Update search curation so implemented `espanol-al-dia` suggestions become executable.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/parse/engine/` | New | Add `espanol_al_dia_article.go` |
| `internal/modules/` | New | Add `espanolaldia/` directory |
| `cmd/dlexa/` | New | Add `espanol_al_dia.go` command |
| `internal/app/` | Modified | Wire new commands and module providers in DI |
| `internal/modules/search/` | Modified | Mark implemented `espanol-al-dia` results as executable |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Unstable or irregular HTML markup from RAE's secondary surfaces | High | Isolate parsing behind the engine's anti-corruption layer; fail gracefully using standard `Problem` builder; do not leak DOM expectations to modules. |
| Incomplete text extraction | Medium | Use robust `goquery` DOM traversal and rely on `text.go` helpers from the shared core to clean up invisible nodes. |

## Rollback Plan

Remove the `espanol_al_dia.go` CLI command file and its binding from `cmd/dlexa/root.go`, plus the new source/module wiring. Since the parser engine uses an additive article-family path, the rollback is isolated from the core `dpd` lookup stability.

## Dependencies

- Requires the completed parser-engine foundation, search migration, and DPD article migration slices.

## Success Criteria

- [ ] Users can execute `dlexa espanol-al-dia <slug>` and receive parsed markdown output.
- [ ] Legacy `dlexa dpd` and `dlexa search` commands remain unaffected.
- [ ] Search renders `espanol-al-dia` suggestions as executable commands rather than deferred guidance.
- [ ] `duda-linguistica` and `noticia` are fully excluded from this implementation slice.

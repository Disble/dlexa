# Proposal: Noticia FAQ Surface

## Intent

Turn FAQ-style `noticia` pages into a first-class executable CLI surface now that the acceptance policy has been simplified to the editorial `Preguntas frecuentes:` prefix.

## Scope

### In Scope
- Simplify noticia rescue policy to the FAQ title prefix alone.
- Add a dedicated `noticia` fetcher, parser, engine wrapper, normalizer, module, and CLI command.
- Wire the surface into the existing lookup pipeline and app defaults.
- Update search truthfulness so rescued FAQ-style noticia results are executable.
- Sync README and active CLI/search specs to the new runtime truth.

### Out of Scope
- Supporting arbitrary non-FAQ institutional noticia pages.
- Generic editorial/news browsing outside normative FAQ content.

## Success Criteria

- [x] `dlexa noticia <slug>` executes for FAQ-style noticia pages.
- [x] Search marks rescued noticia FAQ candidates as executable.
- [x] Non-FAQ noticia pages remain rejected by the search filter or by module-level policy fallback.

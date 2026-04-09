# Design: Noticia FAQ Surface

## Technical Approach

Use the same article-family pipeline already proven by `espanol-al-dia` and `duda-linguistica`, but constrain the product boundary to FAQ-style noticia pages.

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Acceptance gate | `Preguntas frecuentes:` prefix | Architecture docs and live pages show the editorial FAQ prefix is the real semantic signal. |
| Runtime path | Full lookup module | Search already maps noticia URLs; making FAQ pages executable removes drift between discovery and execution. |
| Safety backstop | Module-level lemma check | Even if a slug is invoked directly, non-FAQ noticia content is rejected with a structured fallback instead of leaking institutional content. |
| Parser family | Reuse article-family contract | FAQ noticia pages expose the same `news-title` + `bloque-texto` article shell as other article surfaces. |

## Data Flow

`dlexa noticia <slug>`

→ `cmd/dlexa/noticia.go`
→ `internal/app.App.ExecuteModule("noticia", ...)`
→ `internal/modules/noticia.Module`
→ `internal/query.LookupService`
→ `internal/source.PipelineSource`
→ `internal/fetch.NoticiaFetcher`
→ `internal/parse/engine.NoticiaArticleParser`
→ `internal/normalize.NoticiaNormalizer`
→ shared markdown/json lookup renderers

## Testing Strategy

- CLI routing/help/syntax tests for the new command.
- Fetch/parse/normalize tests for FAQ-style noticia pages.
- Module tests for both accepted FAQ content and rejected institutional content.
- Wiring/app/search tests proving registration, defaults, and executable search truthfulness.

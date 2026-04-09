# Proposal: Noticia Policy Gating

## Intent

Strengthen `search` so `/noticia/*` results are rescued only when they are clearly normative linguistic FAQs, preventing institutional or editorial news from entering the DPD-first consultation flow.

## Scope

### In Scope
- Tighten noticia rescue rules in the search curation layer.
- Add explicit negative coverage for institutional/news false positives.
- Preserve deferred guidance for `noticia` even when rescued.
- Sync the search spec to the stricter policy gate.

### Out of Scope
- Adding an executable `noticia` command.
- Parser/fetcher work for `noticia` pages.
- General semantic classification outside the search module.

## Success Criteria

- [x] `/noticia/*` results require a FAQ gate plus linguistic signals before being rescued.
- [x] Broadly language-themed institutional news is filtered out.
- [x] Existing rescued FAQ behavior remains intact.

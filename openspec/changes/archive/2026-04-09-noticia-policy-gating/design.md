# Design: Noticia Policy Gating

## Technical Approach

Keep policy acceptance in `internal/modules/search/filter.go`, not in parsers. The gate becomes two-step:

1. **FAQ gate** — `/noticia/*` titles must begin with `Preguntas frecuentes:`.
2. **Linguistic signal gate** — the normalized title/snippet must still carry linguistic or normative cues (tilde, ortografía, gramática, pronombre, etc.).

This makes the search layer stricter without overfitting the parser engine to product policy.

## Why Here

ADR-0001 explicitly says policy gating is not parser work. `search` already owns institutional-noise filtering, so this slice hardens the right boundary instead of leaking product policy into lower layers.

## Testing Strategy

- Positive test: FAQ-style noticia with clear linguistic signal is rescued.
- Negative test: FAQ-style noticia without linguistic signal is rejected.
- Negative test: non-FAQ institutional noticia with broad language wording is rejected.
- Regression test: rescued noticia remains deferred guidance, not executable.

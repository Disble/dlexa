# Apply Progress: DPD Terminal Semantic Rendering

## Batch Completed

- Completed an immediate corrective batch after verifying that the real `go run ./cmd/dlexa bien` stdout was visually degraded by default ANSI underline emission.

## Work Performed

- Verified with the real CLI output that default stdout was shipping raw ANSI underline sequences throughout long DPD paragraphs, matching the user's regression report in substance even though examples stayed structurally separated.
- Corrected `internal/render/profile.go` so `NewMarkdownRenderer()` returns the safe plain profile by default and ANSI styling is available only through explicit opt-in via `NewMarkdownRendererWithProfile(...)`.
- Kept example rendering as separate indented block lines in every profile, preserving the structural distinction that survives terminal capture, copy/paste, and LLM ingestion.
- Updated renderer, integration, and app-boundary tests so the default acceptance contract now rejects ANSI bytes in baseline stdout while still rejecting `[ej.: ...]`, `ej.:`, `‹...›`, and `*...*`.
- Ran the final repository search pass to confirm accepted DPD renderer tests/goldens no longer codify forbidden wrappers as valid final stdout artifacts.

## Decisions Locked By This Batch

- `stdout` is the acceptance boundary; internal semantics alone do not count.
- `Paragraph.Inlines` is the source of truth for DPD terminal rendering; `Paragraph.Markdown` is fallback/compatibility projection only.
- Default stdout must optimize for clean terminal/capture/copy-paste readability, so ANSI emphasis is opt-in rather than baseline behavior.
- Examples remain structurally distinct through dedicated indented block lines; this is the durable plain-stdout differentiator that does not invent editorial wrappers.
- Mention/italic distinction in baseline plain stdout currently degrades to authored inline text when the cleaner alternative to ANSI would require synthetic wrappers; that limitation is explicit rather than hidden behind more noise.

## Remaining Focus

- Decide in a later batch whether an explicit CLI/config switch should expose the rich ANSI profile for terminals that genuinely want styled output.

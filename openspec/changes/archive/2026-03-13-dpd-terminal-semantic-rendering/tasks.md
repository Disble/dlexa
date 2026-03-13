# Tasks: DPD Terminal Semantic Rendering

## Phase 1: Infrastructure and Contract Reset

- [x] 1.1 Add `internal/render/semantic_terminal_test.go` with table-driven planner expectations that operate on final terminal roles (`prose`, `emphasis`, `example`) from `Paragraph.Inlines`, so the baseline contract is explicit before renderer changes land.
- [x] 1.2 Update shared render test helpers in `internal/render/markdown_test.go` so DPD assertions evaluate the final rendered stdout string/bytes from `MarkdownRenderer.Render(...)`, not `Paragraph.Markdown`, regex-cleaned intermediates, or other pre-stdout projections.
- [x] 1.3 Update `internal/render/dpd_integration_test.go` test setup so parse -> normalize -> render acceptance is anchored to the final stdout boundary; remove any helper path that treats intermediate Markdown-like text as proof of compliance.
- [x] 1.4 Create `internal/render/profile.go` with `TerminalProfile`, `NewMarkdownRendererWithProfile(...)`, and a plain-profile default for `NewMarkdownRenderer()`, verifying that ANSI remains explicit opt-in instead of a guessed or always-on baseline.
- [x] 1.5 Create `internal/render/semantic_terminal.go` with render-local planning types for terminal paragraphs, blocks, and runs, so the renderer can represent example blocks separately from inline prose without inventing wrappers.

## Phase 2: Terminal-Semantic Implementation

- [x] 2.1 Implement paragraph planning in `internal/render/semantic_terminal.go` that traverses `Paragraph.Inlines`, maps `InlineKindMention` and `InlineKindEmphasis` to visible emphasis runs, flushes prose before every `InlineKindExample`, and emits examples as separate block nodes in source order.
- [x] 2.2 Update the DPD rendering path in `internal/render/markdown.go` to render from `Entry.Article` plus `Paragraph.Inlines` through the terminal plan, and remove the DPD-specific cleanup flow that recovers output from `Paragraph.Markdown`, regex stripping, or synthetic guillemet/example wrappers.
- [x] 2.3 Implement the non-ANSI fallback formatting in `internal/render/markdown.go` and `internal/render/profile.go` so prose stays plain, emphasis preserves authored text without renderer-invented wrappers, examples render as indented block lines, and references remain readable inline text without category labels.
- [x] 2.4 Implement rich-profile formatting in `internal/render/profile.go` as an additive path that applies ANSI underline/faint only when an explicit profile is passed, while preserving the same example block layout and plain fallback semantics.

## Phase 3: Renderer Verification

- [x] 3.1 Expand `internal/render/semantic_terminal_test.go` to prove mixed paragraphs produce at least one prose block, one emphasis run, and one example block, and fail when all inline kinds collapse into the same visible class.
- [x] 3.2 Rewrite `internal/render/markdown_test.go` plain-mode cases so they assert visible stdout distinction for prose vs emphasis vs examples, and add explicit negative assertions that reject `[ej.:`, `ej.:`, `‹`, `›`, or any single-line flattened fallback for example content.
- [x] 3.3 Add profile-specific tests in `internal/render/markdown_test.go` that prove `NewMarkdownRenderer()` emits no ANSI bytes in baseline mode and `NewMarkdownRendererWithProfile(...)` adds ANSI styling without changing the structural example-block contract.
- [x] 3.4 Remove or replace stale unit-test fixtures, helper names, and expectations in `internal/render/markdown_test.go` that currently encode accepted `‹...›` wrappers or semantic flattening, so the unit suite no longer blesses broken output.

## Phase 4: Integration and Broken Artifact Migration

- [x] 4.1 Rewrite `internal/render/dpd_integration_test.go` to assert the final parse -> normalize -> render stdout for authoritative DPD fixtures, with focused checks that fail if visible distinction survives internally but disappears in the rendered output.
- [x] 4.2 Refresh `testdata/dpd/bien.md.golden` to the new stdout contract, and remove or replace any golden expectation helpers that still preserve `‹...›`, `[ej.: ...]`, `ej.:`, or prose/example collapse as valid output.
- [x] 4.3 If `internal/query/service_test.go` exercises DPD formatting, update that coverage so the acceptance point is the renderer-produced stdout payload returned at the query boundary, not any intermediate Markdown-ish representation. Verified `internal/query` does not own DPD formatting assertions; the app boundary test remains the stdout acceptance point.
- [x] 4.4 Run targeted package tests for `internal/render` and the parse -> normalize -> render path, and confirm regressions are reported as stdout-visible contract failures rather than as incidental intermediate-format diffs.

## Phase 5: Ownership and Safeguards

- [x] 5.1 Update `README.md` with a short architecture note that `dpd-live-lookup-parity` owns fetch/parse/normalize semantic preservation, while `dpd-terminal-semantic-rendering` owns the final stdout contract and its acceptance tests.
- [x] 5.2 Capture the migration outcome in `openspec/changes/dpd-terminal-semantic-rendering/apply-progress.md` so future work records which broken contracts were intentionally invalidated (`‹...›`, `[ej.: ...]`, `ej.:`, and semantic flattening) and why they must not return under parity fixes.
- [x] 5.3 Run a final repository search over DPD renderer tests and goldens to verify no accepted artifact still codifies forbidden wrappers or treats intermediate Markdown as the release criterion.

## Implementation Notes

- Keep upstream ownership intact: only touch `parse` or `normalize` if implementation proves a required semantic kind is actually missing from `Paragraph.Inlines`; do not reopen `dpd-live-lookup-parity` scope by default.
- Treat broken tests/goldens as broken artifacts, not as product truth. Replace them as part of this change instead of preserving them for backward compatibility.
- Use final stdout as the only acceptance surface for DPD article rendering; intermediate Markdown-like strings are diagnostic material only.

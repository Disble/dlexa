# Design: DPD Terminal Semantic Rendering

## Technical Approach

Treat terminal rendering as its own typed presentation pipeline, not as a cleanup pass over Markdown-like strings.

`dlexa` already preserves the semantics that matter in `model.Paragraph.Inlines`, but `internal/render/markdown.go` currently throws that away at the last mile by stripping emphasis and inventing example wrappers such as generated guillemet-style delimiters. That is the wrong boundary. The design for this change is therefore:

1. keep upstream `fetch -> parse -> normalize` ownership exactly where `dpd-live-lookup-parity` put it;
2. make `internal/render` consume typed inline semantics as the authoritative source for DPD stdout;
3. introduce a render-local terminal plan that separates prose blocks from example blocks and marks emphasis-bearing runs explicitly;
4. format that plan with an explicit terminal profile that defines rich styling vs plain fallback.

The external format name stays `markdown` for CLI compatibility, but the implementation becomes a terminal-semantic renderer for DPD articles. The renderer MUST stop treating `Paragraph.Markdown` as truth for DPD output. `Paragraph.Markdown` remains a compatibility projection only; `Paragraph.Inlines` is the semantic source of truth.

## Boundary Changes and Dependency Direction

The change stays downstream of `dpd-live-lookup-parity`.

Current dependency direction remains valid:

`cmd -> app -> query -> source -> {fetch, parse, normalize} -> model <- render`

What changes is the responsibility inside `render`:

- `parse` and `normalize` own article extraction and semantic preservation;
- `model` owns typed article content (`Section`, `Paragraph`, `Inline`);
- `render` owns only the visible stdout contract: how those semantics become distinguishable final text;
- tests at the render and pipeline boundary own acceptance of the visible contract.

This prevents the previous telephone-game bug: upstream semantics existing in the model no longer count as success unless `stdout` visibly shows them.

## Visible Contract

The renderer adopts three visible classes in final stdout:

| Semantic class | Visible contract | Rich profile | Plain fallback |
| --- | --- | --- | --- |
| Prose | default inline paragraph text | default terminal text | default terminal text |
| Mention / italic / semantic emphasis | inline emphasis, still part of the sentence | ANSI underline for the emphasized span | authored text preserved inline with no renderer-invented wrapper |
| Usage example | separate example block line under the owning paragraph context, never editorially labelled | same block layout, optional ANSI faint on the example line | same block layout with no ANSI |

### Why this contract

The hard part is not prose vs emphasis. The hard part is example visibility without smuggling in fake labels like `ej.:` or fake wrapper punctuation.

Using a separate example block solves that cleanly:

- it stays visibly distinct even when ANSI is unavailable;
- it does not require editorial labels;
- it does not require enclosing punctuation that the source did not author;
- it remains stable for pipes, logs, snapshots, and LLM ingestion because newlines and indentation survive where ANSI may not.

Inline emphasis uses ANSI when the renderer can actually emit a visible semantic distinction. If ANSI is explicitly disabled, the renderer preserves the authored text inline and must never invent fallback wrappers of its own. That keeps stdout faithful to the DPD instead of smuggling semantics through fake notation.

## Architecture Decisions

### Decision: Render from typed semantics, never from cleaned `Paragraph.Markdown`

**Choice**: The DPD terminal renderer MUST plan output from `Paragraph.Inlines`, not from `Paragraph.Markdown` plus regex stripping.

**Alternatives considered**:
- Keep the current `renderTerminalInline()` cleanup flow and add more regexes.
- Improve `normalize.markdownBody()` and continue rendering from paragraph markdown strings.
- Reparse `Paragraph.Markdown` back into semantic spans inside the renderer.

**Rationale**: Dude, come on. The current bug exists exactly because the last stage treats semantics as disposable formatting noise. Regex-cleaning `*...*` and link syntax guarantees collapse. Reparsing Markdown would be even more mediocre because it reconstructs weaker semantics from a lossy projection. The typed inline model already exists; the renderer just has to respect it.

### Decision: Add a render-local terminal plan between model and string output

**Choice**: Introduce render-local planning types inside `internal/render` before formatting final text.

Illustrative direction:

```go
type TerminalParagraph struct {
    Blocks []TerminalBlock
}

type TerminalBlock struct {
    Kind string // prose | example
    Runs []TerminalRun
}

type TerminalRun struct {
    Role   string // prose | emphasis | reference
    Text   string
    Target string
}
```

**Alternatives considered**:
- Render strings directly while traversing `model.Inline`.
- Push terminal-only roles into `internal/model`.
- Keep only one string builder and branch on inline kind ad hoc.

**Rationale**: A render-local plan makes the visible contract testable without contaminating the domain model with terminal concerns. It also gives one place to express the rule that examples are blocks, not inline wrappers. Direct string building would keep reintroducing punctuation and spacing bugs because block-vs-inline decisions would be scattered through recursive helpers.

### Decision: Example distinction is structural first, stylistic second

**Choice**: Examples render as dedicated indented block lines in all profiles. ANSI styling for examples is optional sugar, not the primary distinguisher.

**Alternatives considered**:
- Enclose examples in generated punctuation such as guillemets or brackets.
- Prefix examples with generated labels such as `ej.:`.
- Keep examples inline and rely only on ANSI color/italic.

**Rationale**: Generated wrappers and labels are explicitly forbidden by spec. ANSI-only distinction is fragile because pipes, logs, CI snapshots, and many terminals erase it. Structural separation is the only fallback that remains visibly distinct without inventing editorial text.

### Decision: ANSI is opt-in by verified profile; plain mode is the safe baseline

**Choice**: The renderer exposes explicit terminal profiles. Plain mode is the default-safe baseline. Rich ANSI mode applies only when the application has a verified reason to enable it.

Illustrative direction:

```go
type TerminalProfile struct {
    ANSIEnabled bool
    ExampleIndent string
}
```

**Alternatives considered**:
- Always emit ANSI.
- Try to guess TTY capability inside `internal/render` from `io.Writer`.
- Add an external terminal-detection dependency during this change.

**Rationale**: The current codebase has no verified terminal capability seam and this design must not invent one. Always-on ANSI would break pipes and snapshots. Guessing from `io.Writer` is hand-wavy and unreliable. Plain mode already satisfies the spec, so it becomes the acceptance baseline; rich mode is additive when a verified capability signal exists.

### Decision: Keep ownership separate from `dpd-live-lookup-parity`

**Choice**: `dpd-live-lookup-parity` continues to own upstream fidelity (`fetch`, `parse`, `normalize`, fixtures as semantic source). This change owns only render-local planning, terminal formatting, and visible-output acceptance.

**Alternatives considered**:
- Fold this work back into the parity change.
- Reopen parser/normalizer design as the main solution.
- Move terminal semantics into `internal/normalize`.

**Rationale**: The upstream change already did the hard work of preserving DPD inline semantics. The remaining failure is downstream. Mixing ownership again would recreate the same ambiguity that caused the rejection.

## Data Flow

### Runtime rendering flow for DPD articles

```text
model.Article
  -> section traversal
  -> paragraph planner
       -> prose blocks
       -> example blocks
       -> emphasis/reference runs
  -> terminal formatter(profile)
  -> final stdout string
```

### Paragraph planning flow

```text
Paragraph.Inlines
  -> classify inline kinds
       text/gloss/reference/... -> prose run
       mention/emphasis         -> emphasis run
       example                  -> flush prose, emit example block
  -> normalize block spacing
  -> format blocks according to profile
```

The planner is where the critical contract lives. Once an example inline is identified, it is no longer allowed to disappear into ordinary prose.

## Formatting Rules

### Prose

- Render in normal section/paragraph flow.
- Preserve existing readable section layout from `internal/render/markdown.go`.
- Keep cross-references readable as plain `-> 6` style text unless a separate change redefines reference formatting.

### Mention / italic / semantic emphasis

- `InlineKindMention` and `InlineKindEmphasis` map to visible emphasis runs.
- Rich profile: wrap the span with ANSI underline sequences.
- Plain profile: preserve the authored text inline without renderer-invented wrappers.
- Nested emphasis collapses to one visible emphasis run, not duplicated markers.

### Usage examples

- `InlineKindExample` maps to its own example block.
- The block appears on a new indented line beneath the prose fragment that introduces it.
- Multiple examples in one paragraph become multiple sibling example blocks in source order.
- No generated label, no enclosing punctuation, no synthetic wrapper.
- Rich profile MAY apply ANSI faint to the example line, but the block layout remains mandatory either way.

### Fallback / degradation policy

1. **Default stdout is plain and ANSI-free**.
   - `NewMarkdownRenderer()` keeps examples block-separated and leaves mention/italic text authored inline without terminal escape noise.
   - This is the accepted stdout contract because it preserves readable terminal/capture/copy-paste output without editorial invention.

2. **Explicit ANSI mode is additive, not baseline**.
   - Applies only when a caller explicitly enables ANSI through the render profile.
   - The renderer adds underline styling for mention/italic spans while keeping the same example-block structure.
   - Plain mode and rich mode alike MUST NOT introduce `*...*`, labels, guillemets, or any equivalent synthetic marker.

3. **No profile is allowed to collapse visibility**.
   - If ANSI is disabled or stripped, the renderer must still keep examples structurally distinct and must not editorialize the remaining text.
   - If a style mechanism cannot be guaranteed, the renderer preserves source-authored text instead of inventing wrappers.

### Tradeoffs

- Example blocks slightly change paragraph flow compared to inline HTML, but they preserve meaning more honestly than fake labels or punctuation wrappers.
- Invented emphasis wrappers are forbidden even when they look convenient; preserving authored text is safer than fabricating notation the DPD never wrote.
- Defaulting to plain mode means the first implementation does not depend on unverified TTY detection. That is less flashy, but architecturally honest.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/render/markdown.go` | Modify | Replace string-cleanup rendering with terminal-plan traversal based on `Paragraph.Inlines`. |
| `internal/render/semantic_terminal.go` | Create | Render-local planner types and helpers that map model semantics to terminal blocks/runs. |
| `internal/render/profile.go` | Create | Define explicit terminal profiles and ANSI/plain formatting helpers. |
| `internal/render/markdown_test.go` | Modify | Replace stale guillemet expectations with profile-aware visible-contract tests. |
| `internal/render/dpd_integration_test.go` | Modify | Assert final stdout behavior from parse->normalize->render using authoritative fixtures. |
| `testdata/dpd/bien.md.golden` | Modify | Update end-to-end expected stdout to the new visible contract. |
| `internal/query/service_test.go` | Possibly modify | Add or update final-output boundary assertions only if query-level rendering coverage is needed. |

No parser redesign is required for this change. Upstream packages should only be touched if implementation proves that a required semantic kind is missing from `Paragraph.Inlines`, and that would be a dependency fix, not ownership drift.

## Interfaces / Contracts

The external renderer contract can stay stable:

```go
type Renderer interface {
    Format() string
    Render(ctx context.Context, result model.LookupResult) ([]byte, error)
}
```

The internal constructor behavior changes:

```go
func NewMarkdownRenderer() *MarkdownRenderer               // defaults to plain profile
func NewMarkdownRendererWithProfile(profile TerminalProfile) *MarkdownRenderer
```

Internal rules:

- DPD article rendering MUST prefer `Entry.Article` and `Paragraph.Inlines`.
- `Paragraph.Markdown` MUST NOT be reparsed or regex-cleaned to recover semantics.
- Terminal profiles are render concerns only; they do not leak into `internal/model`.
- Non-DPD or flat-content fallback behavior can remain as-is until a broader renderer redesign is justified.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Inline kind -> terminal role mapping | Table-driven planner tests that prove `mention/emphasis` becomes emphasis runs and `example` becomes example blocks. |
| Unit | Explicit non-ANSI fallback | Assert authored mention text is preserved without `*...*`, examples stay indented block lines, and ANSI bytes are absent. |
| Unit | Rich profile formatting | Assert ANSI underline around emphasis spans, example blocks still present, and no literal fallback markers when ANSI is enabled. |
| Unit | Forbidden output regression | Dedicated tests that fail on `[ej.:`, `ej.:`, guillemet-style generated wrappers, or all-three-categories-collapsed output. |
| Integration | Renderer against normalized fixture article | Update `internal/render/markdown_test.go` golden and targeted assertions around the `bien` sample article. |
| Integration | Parse -> normalize -> render boundary | Update `internal/render/dpd_integration_test.go` so the final fixture-backed stdout proves visible distinction at output boundary, not only internal semantics. |
| Optional integration | Query/app selection of render profile | Only if rich profile selection is wired through app/config in the same change. |

Required acceptance coverage for this change:

- at least one targeted test for visible emphasis distinction in plain mode;
- at least one targeted test for visible example distinction in plain mode;
- at least one negative test for forbidden editorial wrappers;
- at least one end-to-end fixture test that protects the full `bien` stdout contract.

## Migration / Rollout

No data migration is required.

Implementation order should be:

1. add render-local terminal plan types;
2. add profile-aware formatter with plain baseline;
3. switch DPD article rendering to the semantic plan;
4. replace stale renderer and integration expectations;
5. optionally wire rich profile selection if a verified capability signal exists.

This sequence keeps ownership clean: renderer first, optional ANSI enablement second.

## Open Questions

- None that block implementation. The safe path is to ship the plain profile as the default contract and treat ANSI enablement as additive only when the application has a verified way to request it.

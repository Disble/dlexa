# Proposal: DPD Terminal Semantic Rendering

## Intent

Fix the REAL rejected contract for DPD lookups: the visible terminal/stdout rendering. The previous change `dpd-live-lookup-parity` improved extraction and internal semantics, but the user still rejects the output that actually reaches the terminal because semantic distinctions are not rendered in an acceptable visible form.

This change exists because stdout is the primary consumption surface for `dlexa`. If prose, mention/italic, and example semantics are not visibly distinguishable there, the product is still wrong even if upstream extraction and normalized data are better internally.

## Why Now

- `dpd-live-lookup-parity` addressed data acquisition and semantic preservation inside the pipeline, but that did NOT close the visible acceptance gap.
- The remaining failure is not a cosmetic bug or a one-off formatting nit; it is a product-contract problem at the terminal boundary.
- The user has explicitly rejected invented editorial wrappers such as `[ej.: ...]`, `ej.:`, `‹...›`, and similar synthetic notation.
- Acceptance must move from "internal semantics improved" to "stdout visibly communicates the semantics without editorial invention".

## Scope

### In Scope

- Define stdout/terminal rendering as the primary acceptance surface for DPD article consumption.
- Establish non-negotiable rendering principles for visible differentiation between prose, mention/italic, and example semantics.
- Specify the allowed direction for terminal-visible representation without promising a final exact spelling before downstream exploration/specification is complete.
- Add explicit acceptance criteria based on rendered output, not only extracted or normalized internal structure.
- Require deterministic tests that assert visible semantic differentiation on stdout for authoritative DPD fixtures.
- Keep the change focused on terminal representation semantics and acceptance, separate from upstream extraction/parity concerns.

### Out of Scope

- Replacing or undoing `dpd-live-lookup-parity`.
- Reopening the live-source acquisition problem, parser chrome stripping, or the broader parity pipeline unless a renderer-facing gap proves impossible to solve otherwise.
- Promising a single exact final Markdown/token format before exploration/spec work validates the best terminal-safe representation.
- Adding synthetic editorial labels, invented wrappers, or explanatory markers that are not present in the source semantics.
- Expanding into browser/UI rendering, new source integrations, or unrelated CLI UX redesign.

## Relationship to `dpd-live-lookup-parity`

This change is COMPLEMENTARY, not a replacement.

- `dpd-live-lookup-parity` is about acquiring, extracting, and preserving the right DPD semantics from the live source.
- `dpd-terminal-semantic-rendering` is about making those semantics acceptable and visible on stdout.
- The prior change remains the upstream semantic/data foundation; this change defines and enforces the terminal-facing contract that sits downstream of that foundation.
- If `dpd-live-lookup-parity` solves semantic preservation but stdout still collapses or editorializes those semantics, this change is still required.

## Approach

Treat terminal rendering as its own contract boundary.

- Define rendering principles from the point of view of the terminal reader and LLM consumer, not from internal model completeness alone.
- Preserve visible semantic differentiation among prose, mention/italic, and example content.
- Reject invented editorial cues; the renderer must communicate semantics through faithful terminal-safe representation, not through fabricated labels.
- Evaluate acceptance using fixture-backed visible-output assertions that explicitly test the semantic categories the user cares about.
- Keep extraction/parity responsibilities separate: this change may depend on upstream semantics, but it must not dissolve into another data-pipeline rewrite.

## Non-Negotiable Principles

- Stdout is the primary surface of consumption and therefore the primary acceptance surface.
- Prose, mention/italic, and example content MUST remain visibly distinguishable in terminal output.
- The renderer MUST NOT invent editorial labels or wrappers such as `ej.:`, `[ej.: ...]`, `‹...›`, or equivalent synthetic notation.
- Visible acceptance criteria MUST be tested explicitly; internal semantic preservation alone is insufficient.
- This change separates terminal representation concerns from extraction/parity concerns even when they share upstream fixtures.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/render/markdown.go` | Likely Modified | Primary terminal/stdout rendering rules will need to reflect the visible semantic contract. |
| `internal/render/*_test.go` | Likely Modified | Renderer-focused tests must assert visible distinctions between prose, mention/italic, and examples. |
| `internal/query/service_test.go` | Possibly Modified | End-to-end acceptance may need coverage at the CLI/query boundary where stdout-facing output is assembled. |
| `testdata/dpd/*` | Likely Modified | Authoritative fixtures/goldens may need visible-output expectations aligned with the new contract. |
| `openspec/changes/dpd-live-lookup-parity/*` | Referenced | Prior change artifacts remain dependency context, not replacement targets. |

## Acceptance Direction

Acceptance for this change should be guided by these rules:

- The user-visible stdout output for authoritative DPD fixtures MUST show a clear visible distinction between prose, mention/italic, and example content.
- Acceptance MUST fail when those categories collapse into visually ambiguous text, even if the underlying normalized data still contains the distinction.
- Acceptance MUST fail when output introduces invented editorial labels or wrappers that were not authored in the source semantics.
- Acceptance SHOULD be validated through many explicit visible-output tests, not only one broad golden snapshot.
- The exact final representation MAY still be refined in later exploration/spec/design work, but any accepted option MUST obey the non-negotiable principles above.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| The current normalized structure may still be insufficient to support visible terminal distinctions cleanly | Medium | Keep renderer acceptance separate, then let spec/design identify the minimal upstream semantic support actually required. |
| Existing goldens may encode rejected editorial inventions or collapsed semantics | High | Treat stale visible-output expectations as broken artifacts and replace them with principle-driven checks. |
| The team may drift back into treating this as a cosmetic renderer tweak | High | Frame the change explicitly as a stdout contract feature with its own acceptance criteria. |
| Prematurely locking an exact representation could freeze the wrong abstraction | Medium | Lock principles now, defer exact terminal spelling until spec/design validates the best representation. |

## Rollback Plan

If this change proves to be scoped incorrectly, rollback means removing only the terminal-rendering contract changes and restoring the prior renderer expectations, without undoing `dpd-live-lookup-parity` extraction/parity work. That rollback is safe because this change is intended to sit at the presentation/acceptance boundary, not to replace the upstream data pipeline.

## Dependencies

- `dpd-live-lookup-parity` artifacts as upstream context and semantic foundation.
- Authoritative DPD fixtures that expose the prose / mention / example distinctions the terminal contract must preserve.
- Existing renderer/query test seams where visible-output assertions can be anchored.

## Success Criteria

- [ ] The change defines terminal/stdout rendering as a first-class acceptance contract for DPD output.
- [ ] The proposal fixes the scope around visible semantic rendering rather than treating the issue as an isolated cosmetic bug.
- [ ] The resulting acceptance direction requires explicit visible differentiation between prose, mention/italic, and example semantics.
- [ ] The resulting acceptance direction forbids invented editorial labels or synthetic wrappers.
- [ ] The change is positioned as complementary to `dpd-live-lookup-parity`, not as a replacement for extraction/parity work.

## Exploration: dpd-terminal-semantic-rendering

### Current State

`dlexa` already preserves richer DPD inline semantics internally than its visible terminal contract admits, and that mismatch is the REAL problem.

- `internal/model/types.go` already defines distinct inline kinds for `example`, `mention`, `emphasis`, `reference`, and other semantic spans.
- `internal/parse/dpd.go` and related parser tests show the pipeline is already extracting semantic spans such as `span.ejemplo`, `span.ment`, and emphasis-bearing markup instead of flattening everything at parse time.
- `internal/render/markdown.go` then collapses too much meaning at the terminal boundary: `InlineKindExample` is rendered as `‹...›`, `InlineKindReference` becomes plain `→ n`, and all other non-special inline kinds fall through to raw text.
- That means `InlineKindMention` and `InlineKindEmphasis` are visibly flattened in stdout even when they still exist in normalized data.
- Current renderer tests and integration tests explicitly enforce the now-rejected visible contract by asserting `‹Cierra bien la ventana...›` style output and by stripping raw emphasis markers into plain terminal text.
- `dpd-live-lookup-parity` remains in `apply` and is about live acquisition/parity plumbing. It is NOT the right place to keep smuggling terminal-visibility acceptance, because the user rejected the visible output after those internal improvements.

This gives us hard evidence of the split:

- **Upstream semantic extraction exists**.
- **Visible terminal semantics are still wrong**.
- **The user explicitly rejects invented wrappers such as `[ej.: ...]` and `‹...›`**.

### Affected Areas

- `internal/render/markdown.go` - primary contract boundary where terminal-visible distinctions are currently collapsed or editorialized.
- `internal/render/markdown_test.go` - contains tests that currently encode rejected visible behavior such as `‹...›` examples and plain-text flattening of emphasis/mention semantics.
- `internal/render/dpd_integration_test.go` - end-to-end terminal assertions currently treat `‹...›` as acceptable output and therefore preserve the wrong acceptance target.
- `testdata/dpd/bien.md.golden` - likely carries stale expected terminal output once the visible contract is corrected.
- `openspec/changes/dpd-live-lookup-parity/*` - dependency context only; useful as semantic/input foundation, but this new change must own terminal rendering rules and acceptance separately.

### Approaches

1. **Keep terminal output mostly plain and rely on internal semantics only** - leave stdout close to current flattened text while preserving distinctions only in data/model.
   - Pros: minimal implementation churn.
   - Cons: fails the user need, keeps mention/emphasis visually ambiguous, and repeats the exact product mistake that triggered this change.
   - Effort: Low

2. **Add synthetic editorial markers for visibility** - use wrappers or labels such as `[ej.: ...]`, `ej.:`, guillemets, or similar terminal-only notation.
   - Pros: fast visible differentiation.
   - Cons: explicitly rejected by the user, invents editorial semantics not authored in the source, and turns rendering into annotation instead of faithful representation.
   - Effort: Low-Medium

3. **Define a semantic terminal rendering contract** - treat stdout as its own acceptance boundary and render prose, mention/italic/emphasis, and examples with terminal-visible differentiation that is faithful, non-editorial, and derived from existing semantic nodes.
   - Pros: addresses the actual rejected surface, cleanly separates visual contract from extraction/parity work, and reuses the semantic structure already present upstream.
   - Cons: requires tighter specification of what kinds MUST remain distinguishable and may force rethinking current markdown-vs-terminal assumptions.
   - Effort: Medium

### Recommendation

Recommend **Approach 3: semantic terminal rendering contract**.

Dale, no mezclemos cimientos con terminaciones. `dpd-live-lookup-parity` is the upstream pipeline change: fetch, extract, preserve semantics. This new change must own the downstream rule that stdout MUST communicate those semantics visibly without inventing editorial sugar.

The proposal/spec phase should lock down these boundaries:

- terminal/stdout is a first-class product contract, not a byproduct of internal model quality;
- prose normal, mention/italic/emphasis, and examples MUST remain visibly distinguishable in terminal output;
- the renderer MUST NOT invent wrappers like `[ej.: ...]`, `ej.:`, `‹...›`, or equivalent fabricated notation;
- acceptance must be driven by visible-output rules and targeted assertions, not by “the AST still knows it” excuses;
- this change is downstream of `dpd-live-lookup-parity`, not a replacement for live extraction/parity work.

### Risks

- The current renderer may be assuming that removing markdown markers automatically makes output “terminal friendly,” which is exactly how semantic distinction gets lost.
- Existing renderer goldens/tests currently encode rejected output, so stale tests may fight the correct contract until they are rewritten.
- If proposal/spec stay vague about visible differentiation, the implementation will drift into another fake fix where semantics exist internally but disappear on stdout.
- Some visible distinctions may require clarifying whether terminal output is true Markdown, markdown-derived plain text, or a dedicated semantic terminal format; that must be decided explicitly in proposal/design.
- Because `dpd-live-lookup-parity` is still in apply, ownership boundaries must stay sharp so changes do not collapse back into one giant parity/rendering quilombo.

### Ready for Proposal

Yes.

The next phase should define terminal-visible acceptance criteria separately from acquisition/parity criteria and freeze the forbidden-output set at minimum: no `[ej.: ...]`, no `‹...›`, and no flattening of mention/emphasis/example semantics into indistinguishable prose.

# Proposal: Live Semantic Search Gateway

## Intent

Turn `dlexa search <query>` into the real search-first entrypoint promised by the current architecture: a live semantic gateway that retrieves upstream RAE search results, removes non-normative noise, and returns safe, literal `dlexa ...` next-step commands without prematurely implementing every future content module.

## Problem Statement

The last two archived changes closed critical platform foundations:

- `dlexa-v2-cobra-migration` introduced the Cobra command tree, thin command routing, and a dedicated search module skeleton.
- `dpd-live-lookup-parity` established a real live DPD acquisition path with parity-focused normalization and rendering for the anchor article `bien`.

Even with those foundations in place, `dlexa search` still does not fulfill the contract already described in `openspec/specs/search/spec.md` and the architecture documents. In practice, search behavior remains close to a DPD-specific discovery flow rather than a broader semantic gateway over RAE surfaces.

This mismatch is now architectural debt:

- the specs describe `search` as a semantic router,
- the search mapper already compresses known URLs into commands such as `dlexa espanol-al-dia ...`, `dlexa noticia ...`, and `dlexa duda-linguistica ...`,
- but the runtime command surface still exposes only `search` and `dpd`.

If we try to solve both problems at once — live search gateway plus full implementation of every mapped destination command — the change becomes too large, mixes unrelated parsing concerns, and weakens verification discipline.

## Goals

- Make `dlexa search <query>` fetch real upstream search results instead of behaving as a narrow DPD-only helper.
- Preserve the query-first architecture and explicit `fetch -> parse -> normalize -> render` boundaries for the search flow.
- Filter institutional and non-normative noise aggressively so the search output stays useful for agents.
- Convert recognized result URLs into safe, literal, copyable `dlexa ...` commands using the existing mapping model.
- Return graceful fallback output for unmapped URLs, empty result sets, and upstream/search parsing failures.
- Keep the change scoped tightly enough that later module-specific SDDs can implement destination commands independently.

## Non-Goals

- Implementing all future RAE content modules in this change.
- Adding executable Cobra subcommands for `espanol-al-dia`, `noticia`, or `duda-linguistica` in this change.
- Parsing and rendering the destination page content behind every search result.
- Reworking the default DPD direct lookup flow established by `dpd-live-lookup-parity`.
- Introducing build steps, packaging changes, or unrelated CLI UX redesign.

## Scope

### In Scope

- Connect `dlexa search <query>` to a live upstream search acquisition path appropriate for RAE result discovery.
- Parse the upstream search result payload/HTML into stable internal search-result structures.
- Normalize search results into a curated representation containing title, snippet, source URL, and next-step metadata.
- Filter known institutional and low-value surfaces while preserving linguistically relevant results, including approved rescue cases.
- Map recognized URLs into literal command suggestions through the search mapper.
- Define how unmapped but still visible results are represented without malformed syntax or crashes.
- Render search results in agent-optimized output that clearly distinguishes:
  - directly suggested commands,
  - visible but unmapped results,
  - and no-results/failure states.
- Add focused tests for filtering, mapping, empty states, and fallback behavior.

### Out of Scope

- New production parsers/normalizers/renderers for `espanol-al-dia`, `noticia`, `duda-linguistica`, or other future modules.
- Cobra command-tree expansion beyond the existing `search` and `dpd` commands.
- Multi-page search crawling, ranking experimentation, or concurrent fan-out across multiple providers.
- Persistent caching redesign.
- Any build process changes.

## Architectural Direction

This change SHOULD establish `search` as a discoverability gateway, not as a disguised batch implementation of future modules.

### Scope Decision: gateway contract now, executable destination commands later

This proposal explicitly chooses to implement the **gateway contract and safe command suggestions now**, while deferring the actual executable destination commands to later focused SDD changes.

That tradeoff is intentional:

- The search gateway problem is about live acquisition, result extraction, curation, filtering, and command synthesis.
- Destination-module implementation is a different problem involving source-specific page parsing, normalization, rendering semantics, and command behavior.
- Combining both would blur boundaries, inflate the verification matrix, and make rollback riskier.

Under this change, a mapped result MUST surface a literal command suggestion even if the destination command is not yet implemented. The gateway is therefore responsible for **safe semantic guidance**, not for pretending those modules already exist.

### Boundary Direction

- `cmd/dlexa/search.go` MUST remain thin and delegate orchestration inward.
- Search acquisition, parsing, normalization, filtering, and mapping SHOULD live behind explicit internal boundaries rather than inside Cobra handlers.
- Rendering MUST remain downstream of normalization so both Markdown/text and JSON can derive from the same curated search model if needed later.
- Existing DPD direct lookup behavior MUST remain isolated from the search gateway change.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/dlexa/search.go` | Modified | Keep the existing Cobra command name and arguments, but replace the current limited behavior with live semantic gateway orchestration. |
| `openspec/specs/search/spec.md` | Modified later | Refine the semantic gateway contract, including mapped-command suggestions, fallback behavior, and rejection cases. |
| `openspec/specs/cli/spec.md` | Modified later | Clarify that `search` remains an existing command and that this change does **not** add new destination subcommands yet. |
| `internal/search/` and/or `internal/modules/search/` | Modified | Implement or complete live search orchestration, filtering, mapping, and curated result assembly behind explicit boundaries. |
| `internal/fetch/*search*` | Modified/New | Add live upstream search acquisition with robust request behavior and typed failure handling. |
| `internal/parse/*search*` | Modified/New | Extract structured search results from upstream responses. |
| `internal/normalize/*search*` | Modified/New | Normalize parsed search responses into stable curated results and fallback states. |
| `internal/render/*search*` | Modified/New | Render agent-optimized curated search output without implying destination execution support. |
| `internal/modules/search/mapper.go` | Modified if needed | Preserve URL-to-command compression and extend it only as required for live search result shapes. |
| `internal/app/wiring.go` | Modified if needed | Wire the live search flow cleanly into the current composition root. |

### Cobra Surface Impact

- **No new Cobra subcommands are added in this proposal.**
- The existing `search` command behavior changes materially from limited discovery to live semantic routing.
- Root default-to-DPD behavior remains unchanged.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Upstream search markup or payload shape drifts and breaks extraction | Medium | Isolate fetch/parse boundaries, verify against captured fixtures, and return explicit parse/search fallback output rather than panicking. |
| Search output suggests commands that are not yet executable, confusing users or agents | High | Make command suggestions explicit as recommended next steps, not implied executed capabilities, and keep destination implementation out of scope. |
| Search scope expands into a multi-module rewrite | High | Freeze this change to gateway behavior, filtering, mapping, and fallback semantics only. |
| Institutional filtering drops useful normative content or preserves too much noise | Medium | Specify acceptance and rescue scenarios explicitly in specs, including known keep/drop examples. |
| Search work regresses the direct DPD lookup flow | Low | Preserve existing DPD wiring boundaries and add focused regression coverage around the default `dpd` path. |

## Rollback Plan

If the live search gateway proves unstable, revert the search command wiring to the prior narrower behavior as one atomic change, remove the new live-search fetch/parse/normalize/render path from the default search execution path, and preserve only any non-invasive internal helpers that are dormant and not wired into production flow. This rollback is safe because the proposal does not require new persistent state, does not alter the default direct DPD path, and does not expand the Cobra tree with new runtime commands.

## Dependencies

- Archived `dlexa-v2-cobra-migration` for the current Cobra command tree and search-module shell.
- Archived `dpd-live-lookup-parity` for proven live-source handling patterns, typed failure separation, and fixture-oriented verification style.
- Availability and sufficient stability of the chosen upstream RAE search surface.
- Existing search mapping conventions already represented in `internal/modules/search/mapper.go`.

## Verification Expectations

- `dlexa search <query>` MUST return curated live search results when upstream data is available.
- Search results MUST drop institutional or otherwise low-value noise according to the search spec contract.
- Approved rescue cases MUST remain visible when they are linguistically valuable even if they resemble normally filtered surfaces.
- Recognized URLs MUST produce literal, copyable `dlexa ...` command suggestions.
- Unmapped URLs MUST fall back safely without malformed command syntax or crashes.
- Empty-result or nonsense-query cases MUST produce a clear no-results response rather than stack traces or misleading command suggestions.
- Upstream transport failure and parse failure MUST produce explicit fallback responses rather than silent empty success.
- Direct DPD lookup behavior (for example `dlexa bien`) MUST remain unaffected.
- Verification SHOULD rely on fixture-based tests plus focused integration-style coverage for search orchestration; live probes MAY exist only as explicit opt-in coverage if later needed.

## Success Criteria

- [ ] `dlexa search <query>` uses a real live search acquisition path rather than behaving as a DPD-only discovery helper.
- [ ] Search results are normalized into curated output with titles, snippets, URLs, and next-step metadata.
- [ ] Institutional noise is filtered while approved rescue content remains available.
- [ ] Recognized result URLs are compressed into literal `dlexa ...` command suggestions.
- [ ] Unmapped, empty, and failure cases are rendered safely and explicitly.
- [ ] No new destination Cobra subcommands are required for this change to be considered complete.
- [ ] Existing DPD direct lookup behavior remains stable.

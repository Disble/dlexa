# Archive Report: dpd-terminal-semantic-rendering

## Status

Completed and archived on `2026-03-13`.

## Executive Summary

This change closes the DPD renderer contract as semantic Markdown.

During implementation the work drifted into a plain/ANSI terminal contract, but the final repository state corrects that deviation: `internal/render/markdown.go`, `internal/render/markdown_test.go`, `internal/render/dpd_integration_test.go`, and `testdata/dpd/bien.md.golden` now align on Markdown output with authored emphasis and Markdown links, while rejecting `[ej.: ...]`, `ej.:`, `‹...›`, raw ANSI dependence, and other invented fallback markers.

## Lineage

- Proposal: Engram `#296` (`sdd/dpd-terminal-semantic-rendering/proposal`)
- Spec: Engram `#303` (`sdd/dpd-terminal-semantic-rendering/spec`)
- Design: Engram `#307` (`sdd/dpd-terminal-semantic-rendering/design`)
- Tasks: Engram `#310` (`sdd/dpd-terminal-semantic-rendering/tasks`)
- Apply progress: Engram `#314` (`sdd/dpd-terminal-semantic-rendering/apply-progress`)
- Drift verification evidence: Engram `#327`
- Final contract correction decision: Engram `#329`
- Verify report: no standalone verify artifact was persisted; archive relies on the final repository evidence and aligned golden/tests captured below

## Specs Synced

- Created `openspec/specs/dpd/spec.md` as the main DPD rendering source of truth
- Archived this change with the corrected final contract: semantic Markdown output, not plain/ANSI-only terminal rendering

## Evidence Locked At Archive Time

- `internal/render/markdown.go` renders DPD article content as semantic Markdown from typed inlines
- `internal/render/markdown_test.go` asserts Markdown emphasis, Markdown links, and absence of ANSI and synthetic wrappers
- `internal/render/dpd_integration_test.go` validates parse -> normalize -> render output against semantic Markdown expectations
- `testdata/dpd/bien.md.golden` is aligned to Markdown emphasis and canonical Markdown references

## Non-Blocking Follow-Up

- Nested emphasis edge cases may still deserve a narrower future change, but they do not block this archive because the final shipped contract is already stable and aligned to Markdown

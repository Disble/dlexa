# dlexa Architecture Index

## Purpose

This document is the stable entrypoint for architecture guidance in `dlexa`.

It does not replace detailed design artifacts, OpenSpec, or implementation-level docs. Its job is to:

- define the durable architectural reading of the system;
- point to the current source-of-truth architecture documents;
- index Architecture Decision Records (ADRs);
- keep long-lived decisions separate from temporary implementation chatter.

## Current Architecture Sources

The current architecture baseline is described in:

- `docs/architecture-formal-dlexa-v2.md`
- `docs/architecture_v2_oraculo.md`

These documents explain the v2 direction, modular runtime boundaries, engine responsibilities, and the explicit surface-based consultation posture.

## Core Architectural Principles

The system architecture currently follows these principles:

1. **Normative consultation posture with explicit surfaces**
   - `dlexa` is a consultation interface for normative linguistic doubts in Spanish.
   - The runtime requires explicit command surfaces such as `search`, `dpd`, and slug-based commands.
   - It is not a generic dictionary replacement or universal lexical browser.

2. **Thin CLI, explicit modules**
   - `cmd/dlexa` stays thin.
   - Business intention lives in `internal/modules`.
   - `internal/app` is the composition root.

3. **Reusable engines under explicit module semantics**
   - Shared infrastructure (`fetch`, `parse`, `normalize`, `cache`, `render`) is reusable.
   - Modules orchestrate those engines according to business semantics.

4. **Markdown-first output**
   - Markdown is the primary human/agent format.
   - JSON is secondary and more conservative.

5. **Anti-corruption at external boundaries**
   - Upstream RAE/DPD HTML is unstable external input.
   - Complexity from external markup must die inside adapters and engines, not leak to module or render contracts.

## ADR Convention

Architecture decisions are recorded under:

- `docs/adrs/`

Naming convention:

- `ADR-0001-some-decision.md`
- `ADR-0002-another-decision.md`

Each ADR should contain at least:

- Status
- Context
- Decision
- Consequences

## Current ADR Index

- `ADR-0001-parser-engine.md` — parser engine architecture, ports, and migration direction

## Relationship to OpenSpec

- **OpenSpec** defines product and behavior requirements.
- **ARCHITECTURE.md + ADRs** capture durable structural and design decisions.

If behavior requirements and architecture notes appear to diverge, verify the runtime code and then update the corresponding source of truth intentionally.

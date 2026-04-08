# Architecture Specification

## Purpose

Defines the repository layout policy, establishing that active documentation and specifications MUST reflect the true runtime filesystem and entrypoint architecture.

## Requirements

### Requirement: Runtime Truth Policy

The repository layout and wired runtime files MUST be treated as the authoritative source of architectural truth.

#### Scenario: Docs mismatch runtime

- GIVEN an architectural document claims a specific entrypoint
- WHEN that entrypoint does not exist in the codebase
- THEN the document MUST be updated to reflect the real entrypoint (e.g., `internal/app/app.go`)

### Requirement: Drift Prevention

Active specifications MUST NOT claim file or folder structures that do not exist, unless they are being materialized in the same change.

#### Scenario: Proposing new folder structure

- GIVEN a spec proposes a new architectural layout
- WHEN the implementation phase is complete
- THEN the missing folders MUST be materialized in the filesystem
- AND if they are not materialized, the spec MUST be reverted or archived

### Requirement: Future-Agent Guidance

Agent guidance files MUST direct agents to the real runtime entrypoints.

#### Scenario: Agent onboarded

- GIVEN a new AI agent reads `AGENTS.md`
- WHEN the agent searches for the application composition root
- THEN the guidance MUST point to `internal/app/app.go` and `internal/app/wiring.go`

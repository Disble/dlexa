# Render Specification

## Purpose

Defines the `EnvelopeRenderer` and universal fallback mechanisms for `dlexa v2`, standardizing LLM agent interactions, reducing error ambiguity, and formalizing Markdown wrappers.

## Requirements

### Requirement: Centralized Markdown Envelope

The system MUST wrap successful module outputs in a canonical Markdown Envelope that standardizes module names, context titles, cache states, and upstream sources.

#### Scenario: Envelope prepends metadata to Markdown output

- GIVEN a module returns a successful `Response`
- WHEN the output format is Markdown
- THEN the Envelope Renderer MUST wrap the content with an explicit header (e.g., `# [dlexa:module] title`)
- AND include caching metadata (e.g., `HIT` or `MISS`)

#### Scenario: Envelope bypasses JSON payloads

- GIVEN a module returns a successful `Response`
- WHEN the output format is JSON
- THEN the Envelope Renderer MUST skip applying the Markdown Envelope header
- AND return the payload exactly matching the expected JSON contract

### Requirement: Four-Level Explicit Fallback Ladder

The system MUST classify all errors into four explicit fallback tiers (Syntax, NotFound, UpstreamUnavailable, ParseFailure) to dictate LLM agent recovery paths.

#### Scenario: Syntax (Level 1) fallback guides agent syntax

- GIVEN a Cobra parsing failure or incorrect flags
- WHEN the error reaches the Envelope Renderer
- THEN it MUST return a Syntax fallback response
- AND suggest proper syntax or `--help`

#### Scenario: NotFound (Level 2) fallback suggests search

- GIVEN a module cannot locate the requested term or path
- WHEN the error reaches the Envelope Renderer
- THEN it MUST return a NotFound fallback response
- AND suggest executing `dlexa search <query>` to discover valid paths

#### Scenario: Upstream (Level 3) fallback prevents retry loops

- GIVEN the external service returns a 5xx error or connection timeout
- WHEN the error reaches the Envelope Renderer
- THEN it MUST return an Upstream fallback response
- AND explicitly instruct the LLM agent to abort retry attempts

#### Scenario: Parse (Level 4) fallback alerts maintenance

- GIVEN a fetch succeeds but normalization breaks due to upstream HTML changes
- WHEN the error reaches the Envelope Renderer
- THEN it MUST return a Parse fallback response
- AND instruct the LLM agent that human developer intervention is required

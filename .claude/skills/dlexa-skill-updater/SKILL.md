---
name: dlexa-skill-updater
description: >
  Automate maintenance of the dlexa-user skill by detecting CLI interface changes,
  semantic output drift, and regenerating skill documentation.
  Trigger: Load when CLI interface changes, parser/model/normalize/render/spec output contracts change, or before release.
license: Apache-2.0
metadata:
  author: Disble
  version: "1.0"
---

## When to Use

- When `internal/app/app.go` flag definitions change (new/removed/renamed flags)
- When `internal/render/*.go` output formats change (JSON/markdown structure)
- When `internal/parse/**/*.go` or `internal/normalize/**/*.go` changes alter DPD semantics
- When `internal/model/types.go` request/response types change
- When DPD specs, goldens, or sign-analysis docs change
- Before tagging a release (pre-release checklist item)
- When adding new commands or subcommands (e.g., `--doctor`, `--version`)
- When `dlexa-user` examples fail validation

## Source Map

| Component | File Path(s) | Where to Look |
|-----------|--------------|---------------|
| Flag definitions | internal/app/app.go | flagSet.String/Bool calls in Run() |
| Command logic | internal/app/app.go | Run(), runDoctor(), runLookup() methods |
| Output formats | internal/render/{json,markdown,semantic_terminal}.go | Renderer implementations |
| DPD parsing/normalization | internal/parse/**/*.go, internal/normalize/**/*.go | Sign preservation, bracket semantics, normalization rules |
| Request/response | internal/model/types.go | LookupRequest, LookupResult types |
| Version info | internal/version/version.go | BinaryName, Version constants |
| Help/usage text | internal/app/app.go | printUsage() method |
| DPD contract | openspec/specs/dpd/spec.md, openspec/changes/archive/*/archive-report.md | Current accepted semantic contract and archived decisions |
| DPD evidence fixtures | internal/parse/testdata/*.golden.md, testdata/dpd-signs-analysis/SIGN_ANALYSIS.md | Real article evidence and expected sign behavior |

## Update Workflow

### Phase 1: Detect Drift Surface First

**Purpose**: Identify whether the change is CLI drift, semantic-output drift, or both.

Inspect changed files and classify them:

- CLI surface: `internal/app/app.go`, `internal/version/version.go`
- Output/model surface: `internal/model/types.go`, `internal/render/*.go`
- DPD semantic surface: `internal/parse/**/*.go`, `internal/normalize/**/*.go`, DPD fixtures/goldens/specs

**Success criteria**: You know which documentation sections are at risk before editing anything.

### Phase 2: Gather Authoritative Evidence

**Purpose**: Collect source-of-truth evidence without assuming runtime output is the only contract.

Read, compare, and reconcile:

- `internal/model/types.go` for inline kinds and structured output schema
- `internal/render/*.go` for JSON/markdown rendering behavior
- `internal/parse/**/*.go` and `internal/normalize/**/*.go` for semantic preservation rules
- `openspec/specs/dpd/spec.md` for accepted DPD contract
- Latest archive/verify reports for completed change intent
- Golden fixtures and `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md` for concrete examples

**Success criteria**: Every documented claim can be traced to source, spec, or fixture evidence.

### Phase 3: Read Source Files for Context

**Purpose**: Extract flag defaults, descriptions, output schema, and semantic caveats from source code and specs.

**Files to read** (use Source Map above):
- `internal/app/app.go` → flag definitions, default values, help text
- `internal/model/types.go` → request/response structure
- `internal/render/json.go`, `internal/render/markdown.go` → output format implementations
- `internal/parse/**/*.go`, `internal/normalize/**/*.go` → DPD semantic preservation logic
- `openspec/specs/dpd/spec.md` → validated vs speculative sign contract
- `internal/version/version.go` → version constants

**Success criteria**: You have the flag list, defaults, semantic kinds, and caveats needed to document behavior accurately.

### Phase 4: Update dlexa-user/SKILL.md

**Purpose**: Synchronize the `dlexa-user` skill content with the current CLI and semantic output contract.

**Sections to update**:
1. **Flags Reference** → Sync flag list, descriptions, defaults from `--help` and source
2. **Command Examples** → Update examples with current flags and outputs
3. **Output Formats** → Refresh JSON/markdown examples captured in Phase 2
4. **Version Info** → Update version string format if changed
5. **Doctor Command** → Update diagnostic output format if changed
6. **DPD Semantics** → Document validated inline kinds, speculative kinds, markdown/plain-sign behavior, and archived exclusions
7. **Validation Guidance** → Ensure maintenance checks cover bracket-context semantics and DPD sign preservation

**Mirror rule**: When `.claude/skills/...` and `skills/...` both exist, update both trees in the same task. Missing mirrors should be created when the paired file is required for parity.

**Success criteria**: All flags documented, semantic DPD behavior is current, all examples match real evidence, and mirrors stay synchronized.

### Phase 5: Validate Updated Skill

**Purpose**: Verify the skill against both runtime behavior (when available) and source/spec fixtures.

**Instructions**:
Run the relevant checks for the change surface:

- CLI drift: confirm documented flags/commands still match source and help output
- Semantic drift: confirm documented inline kinds and markdown behavior still match source/spec/fixtures
- Example drift: confirm DPD examples still reflect current golden/spec evidence
- Mirror drift: confirm `.claude/skills/...` and `skills/...` copies match

**Success criteria**: Documentation reflects the current contract even when runtime execution is unavailable or unnecessary.

## Validation Commands

### Flag Presence Validation

**Purpose**: Detect flags in help output that are missing from `dlexa-user`.

```bash
# Windows
dlexa.exe --help

# Linux/macOS
dlexa --help
```

Compare the flag list in help output with the "Flags Reference" section in `dlexa-user/SKILL.md`. Any flag in help but not in the skill indicates drift.

### DPD Semantic Contract Validation

**Purpose**: Detect silent drift in semantic DPD output.

Verify these sources stay aligned:

- `internal/model/types.go`
- `openspec/specs/dpd/spec.md`
- `internal/parse/testdata/*.golden.md`
- `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md`
- `dlexa-user/SKILL.md`, `assets/examples.md`, and `validation.md`

Check specifically for:

- Validated inline kinds: `digital_edition`, `construction_marker`, `bracket_definition`, `bracket_pronunciation`, `bracket_interpolation`
- Speculative-only kinds: `agrammatical`, `hypothetical`, `phoneme`
- Markdown/plain authored sign preservation without synthetic bracket wrappers
- Archived `<` and `>` exclusions remaining explicit

### Example Validity Validation

**Purpose**: Verify that examples in `dlexa-user` still work.

Run each example command from the skill and verify exit code 0.

### Output Structure Validation

**Purpose**: Verify output formats match documented schemas.

```bash
# Windows
dlexa.exe --format json "test query"

# Linux/macOS
dlexa --format json "test query"
```

For JSON output, verify the structure contains expected top-level keys (`request`, `entries`, `sources`, etc.) by parsing the output.

Also verify DPD structured output assumptions when relevant:

- `Article.Sections[].Blocks[].paragraph.Inlines[].Kind` still carries the documented sign semantics
- Markdown examples still show authored/plain sign presentation for validated signs
- Bracket semantics are not misdocumented as markdown-only behavior

### Mirror Synchronization Validation

**Purpose**: Prevent stale duplicated skill docs.

Compare these pairs when both paths exist:

- `.claude/skills/dlexa-user/**` ↔ `skills/dlexa-user/**`
- `.claude/skills/dlexa-skill-updater/**` ↔ `skills/dlexa-skill-updater/**`

Any semantic mismatch is drift and MUST be resolved in the same change.

## Automation Strategy

### When to Trigger Updates

| Trigger Point | Recommendation | Why |
|---------------|----------------|-----|
| Pre-release | **RECOMMENDED** | Ensures skill is accurate before tagging a version |
| Post-flag-change | **RECOMMENDED** | Keeps skill synchronized immediately after interface changes |
| Post-DPD semantic change | **RECOMMENDED** | Parser/normalize/render/spec drift can invalidate documentation even when flags stay unchanged |
| Pre-commit | **NOT RECOMMENDED** | Too noisy, blocks development, skill updates are not commit blockers |
| CI validation | **FUTURE ENHANCEMENT** | Automated drift detection in CI (out of scope for initial implementation) |

### Pre-Release Checklist Item

Before tagging a release (e.g., `v1.2.0`):
1. Run the update workflow (Phases 1-5)
2. Run validation commands
3. If drift detected, complete Phase 4 (update skill)
4. Re-run validation
5. Commit updated `dlexa-user/SKILL.md` if changed

### Post-Change Workflow

When committing changes that affect CLI or semantic output:
1. Classify the drift surface (flags, render/model, parse/normalize, spec/fixtures)
2. Re-read the affected authoritative sources
3. Update `dlexa-user` docs/examples/validation guidance
4. Update `dlexa-skill-updater` if the detection workflow itself needs broader coverage
5. Sync `.claude/skills/...` and `skills/...` mirrors

**Do NOT run this as a pre-commit hook** — it would block every commit and slow development.

## Trigger Detection Reference

| File Pattern | Trigger Level | Rationale |
|--------------|---------------|-----------|
| `internal/app/app.go` | **MUST** | Flag definitions, command logic, help text |
| `internal/model/types.go` | **MUST** | Request/response structure changes |
| `internal/render/*.go` | **MUST** | Output format implementations (JSON, markdown) |
| `internal/parse/**/*.go` | **MUST** | DPD sign extraction and bracket-context detection |
| `internal/normalize/**/*.go` | **MUST** | Semantic kind preservation into the structured model |
| `openspec/specs/dpd/spec.md` | **MUST** | Accepted DPD semantic contract |
| `openspec/changes/archive/*/archive-report.md` | **SHOULD** | Completed-change decisions and caveats |
| `internal/parse/testdata/*.golden.md` | **SHOULD** | User-visible markdown evidence |
| `testdata/dpd-signs-analysis/SIGN_ANALYSIS.md` | **SHOULD** | Validated vs speculative sign inventory |
| `internal/version/version.go` | **SHOULD** | Version info for skill metadata |
| `cmd/dlexa/main.go` | **MAY** | Rarely changes CLI interface (mostly wiring) |
| `internal/query/*.go` | **IGNORE** | Internal query logic, no CLI interface impact |
| `internal/source/*.go` | **IGNORE** | Internal source adapters, no CLI interface impact |
| `*_test.go` | **IGNORE** | Test files do not affect CLI interface |

**Semantic vs Formatting Drift**:
- **Semantic drift** (MUST update): New/removed/renamed flags, output structure changes, new commands, validated/speculative inline-kind changes, bracket-context behavior, markdown sign-presentation contract, archived sign policy
- **Formatting drift** (MAY ignore): Whitespace changes in help text, reordered flags (if all present), example line wrapping that does not change meaning

## Edge Cases and Troubleshooting

### Runtime Output Is Unavailable

If a binary is unavailable or project rules forbid building, continue with source/spec/fixture-based validation. Do NOT block documentation maintenance on a build step when the contract is already recoverable from code, OpenSpec, and goldens.

### Command Execution Fails During Optional Runtime Checks

If any optional runtime command fails (e.g., `dlexa --doctor` returns non-zero exit), investigate the failure:
- Is the binary built correctly?
- Is the command syntax correct for the platform (`.exe` on Windows)?
- Did a recent change introduce a bug?

Do not update the skill based on a broken binary.

### DPD Semantics Changed Without Flag Changes

This is exactly why this skill exists. If parser/model/normalize/render/spec changes alter sign preservation or bracket semantics while CLI flags stay identical, you STILL must update `dlexa-user` and its mirrors.

### False Positive: Help Text Reformatted

If `--help` output has different line breaks but the same flags, this is formatting drift. Compare the flag list (names, types, defaults) rather than raw text. If flags are unchanged, no skill update is needed.

### dlexa-user Skill Does Not Exist Yet

This skill assumes `dlexa-user` exists. If it does not, the update workflow cannot proceed. Create `dlexa-user` first using `skill-creator`, then use this skill to maintain it.

## Resources

- **Skill Creator**: See [skill-creator](~/.config/opencode/skills/skill-creator/SKILL.md) for how to structure the dlexa-user skill
- **Project Conventions**: See [AGENTS.md](../../../AGENTS.md) for repo workflow and lint patterns

---
name: dlexa-skill-updater
description: >
  Automate maintenance of the dlexa-user skill by detecting CLI interface changes,
  inspecting source files, and regenerating skill documentation.
  Trigger: Load when CLI interface changes (flags, commands, outputs) or before release.
license: Apache-2.0
metadata:
  author: Disble
  version: "1.0"
---

## When to Use

- When `internal/app/app.go` flag definitions change (new/removed/renamed flags)
- When `internal/render/*.go` output formats change (JSON/markdown structure)
- When `internal/model/types.go` request/response types change
- Before tagging a release (pre-release checklist item)
- When adding new commands or subcommands (e.g., `--doctor`, `--version`)
- When `dlexa-user` examples fail validation

## Source Map

| Component | File Path(s) | Where to Look |
|-----------|--------------|---------------|
| Flag definitions | internal/app/app.go | flagSet.String/Bool calls in Run() |
| Command logic | internal/app/app.go | Run(), runDoctor(), runLookup() methods |
| Output formats | internal/render/{json,markdown,semantic_terminal}.go | Renderer implementations |
| Request/response | internal/model/types.go | LookupRequest, LookupResult types |
| Version info | internal/version/version.go | BinaryName, Version constants |
| Help/usage text | internal/app/app.go | printUsage() method |

## Update Workflow

### Phase 1: Build the Binary

**Purpose**: Compile the latest dlexa binary to extract its current interface.

**Commands**:
```bash
# Windows
go build -o dlexa.exe ./cmd/dlexa

# Linux/macOS
go build -o dlexa ./cmd/dlexa
```

**Success criteria**: Binary builds without errors and is executable.

**If this fails**: Fix build errors before continuing. The workflow cannot proceed without a working binary.

### Phase 2: Extract Current Interface

**Purpose**: Capture help text, version info, diagnostic output, and example outputs from the built binary.

**Commands**:
```bash
# Windows
dlexa.exe --help                          # Capture usage and flag list
dlexa.exe --version                       # Capture version string
dlexa.exe --doctor                        # Capture diagnostic check format
dlexa.exe --format json "test query"      # Capture JSON output structure
dlexa.exe --format markdown "test query"  # Capture markdown output structure

# Linux/macOS
dlexa --help
dlexa --version
dlexa --doctor
dlexa --format json "test query"
dlexa --format markdown "test query"
```

**Success criteria**: All commands execute successfully (exit code 0) and produce output.

**What to capture**:
- Flag list from `--help` (compare with dlexa-user flag reference)
- Version string format
- Doctor output structure (status line + check list)
- JSON output keys (top-level structure)
- Markdown output sections

### Phase 3: Read Source Files for Context

**Purpose**: Extract flag defaults, descriptions, and internal structure from source code.

**Files to read** (use Source Map above):
- `internal/app/app.go` → flag definitions, default values, help text
- `internal/model/types.go` → request/response structure
- `internal/render/json.go`, `internal/render/markdown.go` → output format implementations
- `internal/version/version.go` → version constants

**Success criteria**: You have the flag list, defaults, and context needed to document each flag accurately.

### Phase 4: Update dlexa-user/SKILL.md

**Purpose**: Synchronize the `dlexa-user` skill content with the current CLI interface.

**Sections to update**:
1. **Flags Reference** → Sync flag list, descriptions, defaults from `--help` and source
2. **Command Examples** → Update examples with current flags and outputs
3. **Output Formats** → Refresh JSON/markdown examples captured in Phase 2
4. **Version Info** → Update version string format if changed
5. **Doctor Command** → Update diagnostic output format if changed

**Success criteria**: All flags documented, all examples valid, all outputs reflect current behavior.

### Phase 5: Validate Updated Skill

**Purpose**: Execute example commands from the updated skill and verify they succeed.

**Instructions**:
Run each example command from `dlexa-user/SKILL.md` and verify:
- Command exits with status 0
- Output structure matches documented format
- No flags are deprecated or missing

**Success criteria**: All examples execute successfully and produce expected output structure.

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

## Automation Strategy

### When to Trigger Updates

| Trigger Point | Recommendation | Why |
|---------------|----------------|-----|
| Pre-release | **RECOMMENDED** | Ensures skill is accurate before tagging a version |
| Post-flag-change | **RECOMMENDED** | Keeps skill synchronized immediately after interface changes |
| Pre-commit | **NOT RECOMMENDED** | Too noisy, blocks development, skill updates are not commit blockers |
| CI validation | **FUTURE ENHANCEMENT** | Automated drift detection in CI (out of scope for initial implementation) |

### Pre-Release Checklist Item

Before tagging a release (e.g., `v1.2.0`):
1. Run the update workflow (Phases 1-5)
2. Run validation commands
3. If drift detected, complete Phase 4 (update skill)
4. Re-run validation
5. Commit updated `dlexa-user/SKILL.md` if changed

### Post-Flag-Change Workflow

When committing changes to `internal/app/app.go` that add/remove/rename flags:
1. After committing the Go code change, run the update workflow
2. Commit the updated `dlexa-user/SKILL.md` as a follow-up commit

**Do NOT run this as a pre-commit hook** — it would block every commit and slow development.

## Trigger Detection Reference

| File Pattern | Trigger Level | Rationale |
|--------------|---------------|-----------|
| `internal/app/app.go` | **MUST** | Flag definitions, command logic, help text |
| `internal/model/types.go` | **MUST** | Request/response structure changes |
| `internal/render/*.go` | **MUST** | Output format implementations (JSON, markdown) |
| `internal/version/version.go` | **SHOULD** | Version info for skill metadata |
| `cmd/dlexa/main.go` | **MAY** | Rarely changes CLI interface (mostly wiring) |
| `internal/query/*.go` | **IGNORE** | Internal query logic, no CLI interface impact |
| `internal/source/*.go` | **IGNORE** | Internal source adapters, no CLI interface impact |
| `*_test.go` | **IGNORE** | Test files do not affect CLI interface |

**Semantic vs Formatting Drift**:
- **Semantic drift** (MUST update): New/removed/renamed flags, output structure changes, new commands
- **Formatting drift** (MAY ignore): Whitespace changes in help text, reordered flags (if all present)

## Edge Cases and Troubleshooting

### Binary Build Fails

If `go build` fails, the workflow cannot proceed. Fix build errors first, then re-run from Phase 1.

### Command Execution Fails During Phase 2

If any command in Phase 2 fails (e.g., `dlexa --doctor` returns non-zero exit), investigate the failure:
- Is the binary built correctly?
- Is the command syntax correct for the platform (`.exe` on Windows)?
- Did a recent change introduce a bug?

Do not update the skill based on a broken binary.

### False Positive: Help Text Reformatted

If `--help` output has different line breaks but the same flags, this is formatting drift. Compare the flag list (names, types, defaults) rather than raw text. If flags are unchanged, no skill update is needed.

### dlexa-user Skill Does Not Exist Yet

This skill assumes `dlexa-user` exists. If it does not, the update workflow cannot proceed. Create `dlexa-user` first using `skill-creator`, then use this skill to maintain it.

## Resources

- **Skill Creator**: See [skill-creator](~/.config/opencode/skills/skill-creator/SKILL.md) for how to structure the dlexa-user skill
- **Project Conventions**: See [AGENTS.md](../../../AGENTS.md) for repo workflow and lint patterns

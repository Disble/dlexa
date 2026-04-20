---
name: dlexa
description: >
  Lightweight manual for LLM agents to consult Spanish normative linguistic doubts using the dlexa CLI (RAE / DPD).
  Use this to resolve questions about Spanish grammar, orthography, pronunciation, tildes, pluralization, or correct word usage.
  Trigger: When the user asks "cómo se escribe", "lleva tilde", "es correcto decir", "cuál es el plural de", or asks about Spanish rules (leísmo, dequeísmo, concordancia).
license: Apache-2.0
metadata:
  author: Disble
  version: "1.0.0"
---

# Skill: dlexa (Light)

## When to Use

Load this skill when the user asks a Spanish normative linguistic question that requires authoritative RAE/DPD guidance.

**Real-world triggers and use cases:**
- **Orthography & Accentuation:** "Lleva tilde 'solo'?", "¿Guion lleva acento?", "¿Las mayúsculas se tildan?", spelling variants ("imprimido" vs "impreso", "oscuro" vs "obscuro").
- **Syntax & Grammar:** Doubts about "leísmo", "laísmo", "dequeísmo", "concordancia" (e.g., "¿es 'la calor' o 'el calor'?"), verb conjugations ("¿andé o anduve?", "satisfizo" vs "satisfació").
- **Morphology:** Pluralization of foreign/complex words ("plural de currículum", "plural de test"), feminine forms ("la juez" vs "la jueza").
- **Punctuation:** Ordering of periods, commas, and quotation marks ("¿el punto va antes o después de las comillas?").
- **Lexico-semantic usage:** "Es correcto usar X en este contexto", "Diferencia de uso entre X e Y".
- **General triggers:** "Según la RAE...", "¿Qué dice el DPD sobre...", "Regla gramatical del español para...".

Do **not** use this skill to present `dlexa` as:
- a universal dictionary replacement (e.g., "qué significa perro", "definición de árbol")
- a translation tool (e.g., "cómo se dice casa en inglés")
- an etymology source (e.g., "de dónde viene la palabra...")
- an encyclopedic reference

## Critical Patterns

### Tool-Selection Decision Rule

| If the user needs... | Use `dlexa`? | Why |
|---|---|---|
| A normative doubt about spelling, pronunciation, morphology, syntax, or usage | Yes | This is the intended consultation model |
| A context-sensitive recommendation that may vary by register, region, or current usage | Yes | DPD guidance is normative but contextual |
| A generic dictionary definition for any arbitrary word | No | `dlexa` is not a universal lexical lookup tool |
| Translation, etymology, or encyclopedic background | No | Use another source |

### Command Syntax

Primary command forms:

```text
dlexa [--no-cache] dpd <query>
dlexa [--no-cache] search [--source <id> ...] <query>
dlexa [--no-cache] dpd search <query>
```

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--source` | string array | Provider selector for search (`search`, `dpd`) | `dlexa search --source dpd solo` |
| `--no-cache` | bool | Skip cache read/write (default: false) | `dlexa --no-cache dpd imprimido` |
| `--doctor` | bool | Run diagnostic checks | `dlexa --doctor` |

## Common Workflows

### 1. Default Two-Step Consultation (Search -> Extract -> Execute)

Always start by searching. Review the markdown list of candidates, pick the one that matches your intent, and **copy-paste the exact command** provided in its `- sugerencia:` field. Do not try to guess the command format.

```bash
# Step 1: Search
dlexa search tilde

# (Agent reads output and picks candidate #6)
# ### 6. Tilde en las mayúsculas
# - clasificación: linguistic-article
# - sugerencia: `dlexa espanol-al-dia tilde-en-las-mayusculas`

# Step 2: Execute the exact suggestion
dlexa espanol-al-dia tilde-en-las-mayusculas
```

### 2. DPD-Only Two-Step Consultation

If you want to restrict the search and discovery *only* to the Diccionario panhispánico de dudas index, use `dpd search`. Then execute the provided `- sugerencia:`.

```bash
dlexa dpd search alicuota
# Agent reads output and copies suggestion:
dlexa dpd alícuota
```

### 3. Force Fresh Data

```bash
dlexa --no-cache dpd tilde
```
Bypasses cache (24-hour TTL) and fetches directly from sources. Use when data seems stale.

### 4. Handle a Lookup Miss Explicitly

If a standard consultation (`dlexa dpd <query>`) unexpectedly misses, read the markdown output. It will explicitly suggest a related entry or command like `dlexa search <query>`. Run that exact suggested command.

### 5. Redirect an Out-of-Scope Task

If the user asks for a generic dictionary definition, translation, or encyclopedic lookup, do **not** force `dlexa` into the answer. State that the task is outside the normative doubt scope and use another tool/source instead.

### 6. Health Check

```bash
dlexa --doctor
```
Runs diagnostic checks. Exit code 0 means healthy.

## Troubleshooting

| Problem | Check | Action |
|---------|-------|--------|
| Search command fails with `search command requires a query` | Was any query text passed after `search`? | Use `dlexa search <query>` |
| Empty results | Is the doubt actually covered by the DPD? | Try `--no-cache`, check with `--doctor`, verify spelling, and consider that the request may be out of scope |
| Search returns no candidates | Did you use the wrong provider scope? | Try `dlexa search <query>`, `dlexa search --source dpd <query>`, or `dlexa dpd search <query>` depending on whether you need federation or the dedicated DPD index |
| Exit code 1 | Check stderr | Read the error message in stderr and follow its instructions |

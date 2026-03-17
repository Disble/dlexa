# dlexa Output Examples

<!-- Examples captured: 2026-03-15, dlexa version 0.1.0-dev -->

This file contains real dlexa command outputs to help LLMs understand the structure and content of **DPD consultation** responses.

Use these examples for normative doubts covered by the DPD. Do **not** treat them as proof that `dlexa` is a generic dictionary replacement with universal lexical coverage.

---

## Markdown Output Example

**Command**: `dlexa tilde`

**Actual output** (truncated after first 30 lines):

```markdown
# tilde

## tilde1

Diccionario panhispánico de dudas

2.ª edición

1. Se llama *tilde* tanto al acento gráfico como al rasgo o trazo pequeño que forma parte de algunas letras, como la *ç*, la *ñ*, la *t*, etc. En ambos casos admite los dos géneros, aunque hoy se usa casi exclusivamente en femenino, salvo en Uruguay, donde lo normal es usarlo en masculino:«Funciona entre el alumnado una regla maldita de los acentos: en la duda, poner la tilde» (Miguel *Perversión* es 1994);«Déjà vu, primero con tilde en la e, después con un tilde grave en la a» (Schutz *Noche* uy 2001). También significa 'tacha o nota denigrativa', sentido en el que también admite su uso en ambos géneros:«Ese tilde de hereje le faltaba a ese Napoleón Malaparte» (FnCaballero *Clemencia* es 1852);«En otras castas es lícito perdonar ciertos leves errores y algunas tildes» (Ortega *Artículos 1902* es 1902-1916, 69). Cuando significa 'lo más mínimo', hoy se usa siempre en femenino:«El progreso de España había sido durante su reinado, sin exagerar una tilde, sensacional» (Laín *Descargo* es 1976).

2. No debe confundirse con *tinte* ('rasgo o matiz'), como ocurre en este ejemplo: ⊗«Afirmó[…] que algunos discursos en favor de la convocatoria de huelga general "tienen tildes claramente fascistas"» (*Mundo* es 26.1.1994).

Source: Real Academia Española y Asociación de Academias de la Lengua Española
Dictionary: Diccionario panhispánico de dudas
Edition: 2.ª edición
URL: https://www.rae.es/dpd/tilde
Consulted: 10/03/2026

## tilde2

Diccionario panhispánico de dudas

2.ª edición

1. Reglas generales de acentuación gráfica

1.1. Monosílabos. Las palabras de una sola sílaba no se acentúan nunca gráficamente, salvo en los casos de tilde diacrítica (→ [3.1](tilde#S1590507322842906071)): *mes*, *bien*, *fe*, *fue*, *vio*, *guion*, *hui*, *riais*.

[...truncated: more content follows...]
```

**Parsing guidance**:
- Headings mark entry start (`## tilde1`, `## tilde2`)
- Content includes markdown formatting (*italic*, hyperlinks)
- Body text carries DPD-backed normative guidance, sometimes with usage or register nuance
- Citation metadata at end (Source, Dictionary, Edition, URL, Consulted)
- Multiple entries appear as separate level-2 headings

---

## JSON Output Example

**Command**: `dlexa --format json tilde`

**Actual output** (truncated, showing structure of first entry only):

```json
{
  "Request": {
    "Query": "tilde",
    "Format": "json",
    "Sources": [
      "dpd"
    ],
    "NoCache": false
  },
  "Entries": [
    {
      "ID": "dpd:tilde1#e1590506567296961312",
      "Headword": "tilde1",
      "Summary": "Diccionario panhispánico de dudas",
      "Content": "1. Se llama *tilde* tanto al acento gráfico como al rasgo o trazo pequeño que forma parte de algunas letras, como la *ç*, la *ñ*, la *t*, etc. En ambos casos admite los dos géneros, aunque hoy se usa casi exclusivamente en femenino, salvo en Uruguay, donde lo normal es usarlo en masculino: ««Funciona entre el alumnado una regla maldita de los acentos: en la duda, poner la tilde» (Miguel *Perversión* es 1994)»; ««Déjà vu, primero con tilde en la e, después con un tilde grave en la a» (Schutz *Noche* uy 2001)». También significa 'tacha o nota denigrativa', sentido en el que también admite su uso en ambos géneros: ««Ese tilde de hereje le faltaba a ese Napoleón Malaparte» (FnCaballero *Clemencia* es 1852)»; ««En otras castas es lícito perdonar ciertos leves errores y algunas tildes» (Ortega *Artículos 1902* es 1902-1916, 69)». Cuando significa 'lo más mínimo', hoy se usa siempre en femenino: ««El progreso de España había sido durante su reinado, sin exagerar una tilde, sensacional» (Laín *Descargo* es 1976)».\n\n2. No debe confundirse con *tinte* ('rasgo o matiz'), como ocurre en este ejemplo: «⊗«Afirmó […] que algunos discursos en favor de la convocatoria de huelga general "tienen tildes claramente fascistas"» (*Mundo* es 26.1.1994)».",
      "Source": "dpd",
      "URL": "https://www.rae.es/dpd/tilde",
      "Metadata": {
        "access_profile": "browser-like direct /dpd/<term>",
        "entry_id": "E1590506567296961312",
        "normalized_by": "dpd"
      },
      "Article": {
        "Dictionary": "Diccionario panhispánico de dudas",
        "Edition": "2.ª edición",
        "Lemma": "tilde1",
        "CanonicalURL": "https://www.rae.es/dpd/tilde",
        "Sections": [
          {
            "Label": "1.",
            "Title": "",
            "Blocks": [
              {
                "kind": "paragraph",
                "paragraph": {
                  "Markdown": "Se llama *tilde* tanto al acento gráfico...",
                  "Inlines": [
                    {
                      "Kind": "text",
                      "Variant": "",
                      "Text": "Se llama ",
                      "Target": "",
                      "Children": null
                    },
                    {
                      "Kind": "mention",
                      "Variant": "ment",
                      "Text": "tilde",
                      "Target": "",
                      "Children": [...]
                    }
                  ]
                }
              }
            ]
          }
        ],
        "Citation": {
          "Source": "Real Academia Española y Asociación de Academias de la Lengua Española",
          "Dictionary": "Diccionario panhispánico de dudas",
          "Edition": "2.ª edición",
          "URL": "https://www.rae.es/dpd/tilde",
          "ConsultedDate": "10/03/2026"
        }
      }
    }
    // [...truncated: 1 more entry (tilde2)]
  ],
  "Warnings": [],
  "Problems": [],
  "Sources": [
    {
      "Name": "dpd",
      "Success": true,
      "Error": "",
      "Priority": 1,
      "ResultCount": 2
    }
  ],
  "CacheHit": false,
  "GeneratedAt": "2026-03-15T..."
}
```

**Navigation patterns**:
- Consultation content: `.Entries[].Content` (markdown-formatted)
- Structured content: `.Entries[].Article.Sections[].Blocks[]`
- Headwords: `.Entries[].Headword`
- Sources: `.Entries[].Source`
- Cache status: `.CacheHit`
- Timestamp: `.GeneratedAt`
- Problems: `.Problems[]` (empty on success)

**Interpretation note**:
- JSON gives structure, not permission to flatten context-sensitive DPD recommendations into fake universal rules.

---

## Search Markdown Output Example

**Command**: `dlexa search abu dhabi`

**Validated output shape** (from `internal/render/search_markdown.go` and `internal/render/search_markdown_test.go`):

```text
Candidate DPD entries for "abu dhabi":
- Abu Dhabi -> Abu Dabi
- ⊗ alicuota -> alícuoto
```

**Parsing guidance**:
- The heading line echoes the search query
- Each bullet maps `display_text -> article_key`
- The right-hand side is the canonical follow-up lookup term
- Empty search state renders as: `No DPD entry candidates found for "zzz".`

---

## Search JSON Output Example

**Command**: `dlexa --format json search guion`

**Validated output shape** (from `internal/model/search.go`, `internal/render/search_json.go`, and renderer/service tests):

```json
{
  "Request": {
    "Query": "guion",
    "Format": "json",
    "NoCache": false
  },
  "Candidates": [
    {
      "raw_label_html": "guion<sup>1</sup>",
      "display_text": "guion1",
      "article_key": "guion"
    },
    {
      "raw_label_html": "<span class=\"vers\">guion<sup>2</sup></span>",
      "display_text": "var. guion2",
      "article_key": "guion"
    }
  ],
  "Warnings": [],
  "Problems": [],
  "CacheHit": false,
  "GeneratedAt": "2026-03-17T..."
}
```

**Navigation patterns**:
- Candidate list: `.Candidates[]`
- Normalized display label: `.Candidates[].display_text`
- Canonical article key: `.Candidates[].article_key`
- Preserved upstream label HTML: `.Candidates[].raw_label_html`
- Cache status: `.CacheHit`

**Interpretation note**:
- Search JSON is an entry-discovery contract, not the full article consultation contract used by `.Entries[]`
- `raw_label_html` deliberately preserves upstream markup; do not assume it is markdown

---

## DPD Semantic Signs Example

**Command**: `dlexa --source dpd --format json alícuota`

**Markdown excerpt** (from the current DPD golden fixture):

```markdown
... (*Día* @ es 26.10.2014). ... ⊗[alikuóto], ⊗[alikuóta].
```

**Structured inline excerpt** (representative fragment from `.Entries[].Article.Sections[].Blocks[].paragraph.Inlines`):

```json
[
  {
    "Kind": "digital_edition",
    "Variant": "",
    "Text": "@",
    "Target": "",
    "Children": null
  },
  {
    "Kind": "bracket_pronunciation",
    "Variant": "",
    "Text": "[alikuóto]",
    "Target": "",
    "Children": null
  },
  {
    "Kind": "bracket_pronunciation",
    "Variant": "",
    "Text": "[alikuóta]",
    "Target": "",
    "Children": null
  }
]
```

**Related validated kind from another DPD fixture**:

```json
{
  "Kind": "construction_marker",
  "Text": "+ infinitivo"
}
```

**What this means**:

- Markdown keeps the authored/plain signs visible: `@`, `⊗`, `[ ... ]`
- JSON keeps bracket context and sign semantics in `Inline.Kind`
- Validated kinds are `digital_edition`, `construction_marker`, `bracket_definition`, `bracket_pronunciation`, and `bracket_interpolation`
- Speculative kinds `agrammatical`, `hypothetical`, and `phoneme` are non-authoritative/inferred only
- Archived `<` and `>` signs remain intentionally unsupported
- The same article may still carry contextual nuance about usage, register, or geography; do not over-prescribe beyond what the DPD says

---

## Error Output Example

**Command**: `dlexa zkxjqwerty` (nonsense query to trigger "not found")

**Actual output**:

**Exit code**: 0 (note: not found is NOT an error, exit code is 0)

**stdout**:
```markdown
# zkxjqwerty

- format: `markdown`
- cache_hit: `false`
- sources: `0`

## Problems

- [dpd_not_found] DPD entry not found for "Diccionario panhispánico de dudas" (dpd/error)
```

**Key observation**: When a term is not found, dlexa:
1. Returns exit code 0 (SUCCESS) because the lookup itself succeeded
2. Shows empty results with a Problem entry explaining why
3. Problem code `dpd_not_found` indicates the specific source had no match
4. This may mean the query is absent from the DPD, not that `dlexa` failed as a generic dictionary

**Handling pattern**:
```bash
result=$(dlexa "palabra")
exit_code=$?
if [ $exit_code -eq 0 ]; then
  # Check if result contains "## Problems" section
  if echo "$result" | grep -q "## Problems"; then
    echo "No results found or other issue" >&2
  else
    echo "Success: $result"
  fi
else
  echo "Command failed: $result" >&2
fi
```

---

## Empty Result Example

**Command**: `dlexa --format json zkxjqwerty`

**Actual output**:

```json
{
  "Request": {
    "Query": "zkxjqwerty",
    "Format": "json",
    "Sources": [
      "dpd"
    ],
    "NoCache": false
  },
  "Entries": [],
  "Warnings": [],
  "Problems": [
    {
      "Code": "dpd_not_found",
      "Message": "DPD entry not found for \"Diccionario panhispánico de dudas\"",
      "Source": "dpd",
      "Severity": "error"
    }
  ],
  "Sources": [
    {
      "Name": "dpd",
      "Success": false,
      "Error": "dpd_not_found",
      "Priority": 1,
      "ResultCount": 0
    }
  ],
  "CacheHit": false,
  "GeneratedAt": "2026-03-15T..."
}
```

**Key observation**: Empty `Entries` array with non-empty `Problems` array indicates the lookup found nothing. Exit code is still 0 (SUCCESS) because the command itself worked.

If the request was really a generic dictionary or encyclopedic lookup, the better conclusion may be "wrong tool for the job," not "broken DPD consultation."

---

## Doctor Output Example

**Command**: `dlexa --doctor`

**Actual output**:

**Exit code**: 0

**stdout**:
```
doctor: ok
- bootstrap [ok] doctor wiring is ready; concrete checks can be added per platform
```

**Key observation**: The `--doctor` command is currently a placeholder that always reports healthy. Future versions will add real connectivity and environment checks.

---

## Problem Codes Reference

Extracted from `internal/model/types.go`:

| Code | Description | User Action |
|------|-------------|-------------|
| `source_lookup_failed` | General source connectivity or fetch failure | Check network, try `--doctor`, verify source is available |
| `dpd_fetch_failed` | Failed to fetch data from DPD source | Check DPD source availability, try `--no-cache` |
| `dpd_not_found` | DPD source returned 404 | The doubt may be absent from DPD scope; use another source if the task is broader than normative consultation |
| `dpd_extract_failed` | Failed to extract content from DPD response | Data format issue, report if persistent |
| `dpd_transform_failed` | Failed to transform DPD data to internal model | Data parsing issue, report if persistent |
| `dpd_search_fetch_failed` | Failed to fetch DPD entry-search payload | Check upstream availability, try `--no-cache`, verify the query is not empty |
| `dpd_search_parse_failed` | Failed to decode or split DPD entry-search payload | Upstream response format drift, report if persistent |
| `dpd_search_normalize_failed` | Failed to normalize parsed DPD search candidates | Label normalization drift, report if persistent |

**All Problem codes have severity `error`.** When a Problem appears in `.Problems[]` array (JSON) or stderr (any format), the exit code is 1.

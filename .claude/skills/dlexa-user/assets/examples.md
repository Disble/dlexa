# dlexa Output Examples

<!-- Examples captured: 2026-03-15, dlexa version 0.1.0-dev -->

This file contains real dlexa command outputs to help LLMs understand the structure and content of responses.

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
- Definitions: `.Entries[].Content` (markdown-formatted)
- Structured content: `.Entries[].Article.Sections[].Blocks[]`
- Headwords: `.Entries[].Headword`
- Sources: `.Entries[].Source`
- Cache status: `.CacheHit`
- Timestamp: `.GeneratedAt`
- Problems: `.Problems[]` (empty on success)

---

## Error Output Example

**Command**: `dlexa zkxjqwerty` (nonsense word to trigger "not found")

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
| `dpd_not_found` | DPD source returned 404 | Word not in DPD dictionary, try other sources |
| `dpd_extract_failed` | Failed to extract content from DPD response | Data format issue, report if persistent |
| `dpd_transform_failed` | Failed to transform DPD data to internal model | Data parsing issue, report if persistent |

**All Problem codes have severity `error`.** When a Problem appears in `.Problems[]` array (JSON) or stderr (any format), the exit code is 1.

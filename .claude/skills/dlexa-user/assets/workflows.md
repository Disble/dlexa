# dlexa Integration Workflows

This file provides copy-paste integration patterns for using dlexa in automation scripts for **DPD consultation workflows**.

These patterns assume the task is a DPD-fit normative doubt. If the job is generic dictionary lookup, etymology, translation, or encyclopedic retrieval, use a different source instead of cargo-culting these scripts.

---

## Shell Script Integration

### Basic Pattern with Exit Code Checking

```bash
#!/bin/bash

# Basic DPD consultation with error handling
result=$(dlexa --format json "solo")
exit_code=$?

if [ $exit_code -eq 0 ]; then
  # Success - parse the recommendation
  recommendation=$(echo "$result" | jq -r '.Entries[0].Content')
  echo "Recommendation: $recommendation"
else
  # Error - log and exit
  echo "Error: dlexa command failed" >&2
  echo "$result" >&2
  exit 1
fi
```

### Search-Then-Lookup Pattern

```bash
#!/bin/bash

query="abu dhabi"
search_result=$(dlexa --format json search "$query")
exit_code=$?

if [ $exit_code -ne 0 ]; then
  echo "Search failed" >&2
  echo "$search_result" >&2
  exit 1
fi

article_key=$(echo "$search_result" | jq -r '.Candidates[0].article_key // empty')
if [ -z "$article_key" ]; then
  echo "No DPD entry candidates for: $query" >&2
  exit 0
fi

dlexa --format json "$article_key"
```

### Robust Pattern with Stderr Capture

```bash
#!/bin/bash

# Capture both stdout and stderr separately
result=$(dlexa --format json "tilde" 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
  # Parse error details from stderr
  problem_code=$(echo "$result" | grep "Problem:" | cut -d' ' -f2)
  echo "dlexa failed with problem code: $problem_code" >&2
  exit 1
fi

# Extract consultation content
echo "$result" | jq -r '.Entries[] | "- \(.Headword): \(.Content)"'
```

---

## Error Handling Pattern

### Checking for Empty Results

```bash
#!/bin/bash

result=$(dlexa --format json "$query")

# Check entries first, then structured misses
entry_count=$(echo "$result" | jq '.Entries | length')
miss_count=$(echo "$result" | jq '.Misses // [] | length')

if [ "$entry_count" -eq 0 ]; then
  if [ "$miss_count" -gt 0 ]; then
    echo "$result" | jq -r '.Misses[] | .suggestion.display_text // .next_action.command // "Structured lookup miss without follow-up metadata"'
    exit 0
  fi
  echo "No DPD guidance found for: $query" >&2
  exit 0  # Not an error, just empty
fi

# Process entries
echo "$result" | jq -r '.Entries[0].Content'
```

### Handling Specific Problem Codes

```bash
#!/bin/bash

result=$(dlexa --format json "$query" 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
  # Check for specific problem codes
  if echo "$result" | grep -q "source_lookup_failed"; then
    echo "Network or source connectivity issue" >&2
    echo "Try again later or check --doctor" >&2
    exit 2
  else
    echo "Unknown error: $result" >&2
    exit 1
  fi
fi
```

---

## Retry Logic Pattern

### Retry with Cache Bypass

```bash
#!/bin/bash

# First attempt (may use cache)
result=$(dlexa --format json "imprimido")

# If it fails, retry with --no-cache
if [ $? -ne 0 ]; then
  echo "First attempt failed, retrying with --no-cache..." >&2
  result=$(dlexa --format json --no-cache "imprimido")
fi

# Process result
echo "$result" | jq -r '.Entries[0].Content'
```

### Retry with Exponential Backoff

```bash
#!/bin/bash

max_retries=3
retry_count=0
wait_time=1

while [ $retry_count -lt $max_retries ]; do
  result=$(dlexa --format json "$query" 2>&1)
  exit_code=$?
  
  if [ $exit_code -eq 0 ]; then
    echo "$result" | jq -r '.Entries[0].Content'
    exit 0
  fi
  
  retry_count=$((retry_count + 1))
  echo "Attempt $retry_count failed, retrying in ${wait_time}s..." >&2
  sleep $wait_time
  wait_time=$((wait_time * 2))
  
  # Use --no-cache on retries
  if [ $retry_count -ge 2 ]; then
    result=$(dlexa --format json --no-cache "$query" 2>&1)
  fi
done

echo "All retries failed" >&2
exit 1
```

---

## Provider-Scoped Search Pattern

### Query Search Providers Sequentially

```bash
#!/bin/bash

query="$1"
sources=("dpd" "search")

for source in "${sources[@]}"; do
  echo "Querying source: $source" >&2
  result=$(dlexa --format json search --source "$source" "$query")
  
  if [ $? -eq 0 ]; then
    candidate_count=$(echo "$result" | jq '.Candidates | length')
    if [ "$candidate_count" -gt 0 ]; then
      echo "=== Results from $source ===" >&2
      echo "$result" | jq -r '.Candidates[] | "- \(.display_text): \(.next_command // .article_key)"'
    fi
  fi
done
```

### DPD-Only Search Then Lookup

```bash
#!/bin/bash

query="$1"
search_result=$(dlexa --format json dpd search "$query")
article_key=$(echo "$search_result" | jq -r '.Candidates[0].article_key // empty')

if [ -n "$article_key" ]; then
  dlexa --format json "$article_key"
fi
```

---

## Workflow Quick Reference Table

| Use Case | Command | Format | Notes |
|----------|---------|--------|-------|
| Quick DPD consultation | `dlexa tilde` | markdown | Human-readable, default |
| Script parsing | `dlexa --format json solo` | json | Use with jq |
| Discover DPD entry candidates | `dlexa dpd search abu dhabi` | markdown | DPD-only candidate labels plus article keys |
| Federated search for automation | `dlexa --format json search guion` | json | Parse `.Candidates[]`, `next_command`, and `deferred` |
| Search one provider only | `dlexa search --source dpd adecua` | markdown | Repeat `--source` to scope providers |
| Force refresh | `dlexa --no-cache imprimido` | markdown | Bypass 24h cache |
| Health check | `dlexa --doctor` | text | Exit 0 = healthy |
| Version info | `dlexa --version` | text | Print version |
| Error handling | `dlexa solo 2>&1` | any | Capture both stdout/stderr |
| Check cache status | `dlexa --format json solo \| jq .CacheHit` | json | Boolean: true/false |
| Extract headwords | `dlexa --format json solo \| jq -r '.Entries[].Headword'` | json | Array of strings |
| Count results | `dlexa --format json solo \| jq '.Entries \| length'` | json | Integer |

---

## Advanced Integration Examples

### Editorial DPD Consultation Audit

```bash
#!/bin/bash
# Validate that a list of queries resolves as DPD consultations

words_file="$1"
failed_words=()

while IFS= read -r word; do
  result=$(dlexa --format json "$word" 2>&1)
  exit_code=$?
  
  if [ $exit_code -ne 0 ]; then
    echo "ERROR: Failed to lookup '$word'" >&2
    failed_words+=("$word")
    continue
  fi
  
  count=$(echo "$result" | jq '.Entries | length')
  if [ "$count" -eq 0 ]; then
    echo "WARNING: No DPD guidance for '$word'" >&2
    failed_words+=("$word")
  else
    echo "✓ $word"
  fi
done < "$words_file"

if [ ${#failed_words[@]} -gt 0 ]; then
  echo "Failed words: ${failed_words[*]}" >&2
  exit 1
fi

echo "All words validated successfully"
```

### JSON to Markdown Consultation Summary

```bash
#!/bin/bash
# Convert JSON output to human-readable markdown

result=$(dlexa --format json "$1")

echo "# DPD consultation for: $1"
echo ""

echo "$result" | jq -r '.Entries[] | "## \(.Headword) (\(.Source))\n\n\(.Content)\n"'
```

### Caching Strategy for Batch Lookups

```bash
#!/bin/bash
# Batch lookup with intelligent cache usage

words=("casa" "perro" "gato" "mesa")

# First pass: use cache
echo "First pass (with cache)..."
for word in "${words[@]}"; do
  result=$(dlexa --format json "$word")
  cache_hit=$(echo "$result" | jq -r '.CacheHit')
  echo "$word: cache_hit=$cache_hit"
done

echo ""
echo "Second pass (force refresh for cache misses)..."

# Second pass: refresh only cache misses
for word in "${words[@]}"; do
  result=$(dlexa --format json "$word")
  cache_hit=$(echo "$result" | jq -r '.CacheHit')
  
  if [ "$cache_hit" = "false" ]; then
    echo "Refreshing $word..."
    dlexa --format json --no-cache "$word" > /dev/null
  fi
done
```

---

## Notes

- All examples use `--format json` for programmatic parsing
- Exit code checking is MANDATORY in production scripts
- Use `--no-cache` sparingly (increases source load)
- `jq` is the recommended JSON parser for bash
- Always handle empty `Entries` arrays together with `.Misses[]` before assuming there was no guidance
- Always handle empty `Candidates` arrays for search (not an error condition)
- Problem codes in stderr indicate fatal errors (exit code 1)
- Cache TTL is 24 hours (not configurable from CLI)
- These workflows are for DPD consultation tasks, not generic dictionary replacement

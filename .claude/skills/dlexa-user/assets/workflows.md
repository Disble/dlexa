# dlexa Integration Workflows

This file provides copy-paste integration patterns for using dlexa in automation scripts.

---

## Shell Script Integration

### Basic Pattern with Exit Code Checking

```bash
#!/bin/bash

# Basic lookup with error handling
result=$(dlexa --format json "palabra")
exit_code=$?

if [ $exit_code -eq 0 ]; then
  # Success - parse the result
  definition=$(echo "$result" | jq -r '.Entries[0].Content')
  echo "Definition: $definition"
else
  # Error - log and exit
  echo "Error: dlexa command failed" >&2
  echo "$result" >&2
  exit 1
fi
```

### Robust Pattern with Stderr Capture

```bash
#!/bin/bash

# Capture both stdout and stderr separately
result=$(dlexa --format json "palabra" 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
  # Parse error details from stderr
  problem_code=$(echo "$result" | grep "Problem:" | cut -d' ' -f2)
  echo "dlexa failed with problem code: $problem_code" >&2
  exit 1
fi

# Extract definitions
echo "$result" | jq -r '.Entries[] | "- \(.Headword): \(.Content)"'
```

---

## Error Handling Pattern

### Checking for Empty Results

```bash
#!/bin/bash

result=$(dlexa --format json "$query")

# Check if Entries array is empty
entry_count=$(echo "$result" | jq '.Entries | length')

if [ "$entry_count" -eq 0 ]; then
  echo "No definitions found for: $query" >&2
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
  elif echo "$result" | grep -q "dpd_not_found"; then
    echo "Word not found in dictionary" >&2
    exit 3
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
result=$(dlexa --format json "palabra")

# If it fails, retry with --no-cache
if [ $? -ne 0 ]; then
  echo "First attempt failed, retrying with --no-cache..." >&2
  result=$(dlexa --format json --no-cache "palabra")
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

## Multi-Source Query Pattern

### Query Multiple Sources Sequentially

```bash
#!/bin/bash

query="$1"
sources=("dpd" "demo")

for source in "${sources[@]}"; do
  echo "Querying source: $source" >&2
  result=$(dlexa --format json --source "$source" "$query")
  
  if [ $? -eq 0 ]; then
    entry_count=$(echo "$result" | jq '.Entries | length')
    if [ "$entry_count" -gt 0 ]; then
      echo "=== Results from $source ===" >&2
      echo "$result" | jq -r '.Entries[] | "- \(.Headword): \(.Content)"'
    fi
  fi
done
```

### Aggregate Results from Multiple Sources

```bash
#!/bin/bash

query="$1"

# Query all configured sources
result=$(dlexa --format json "$query")

# Group by source
echo "$result" | jq -r '
  .Entries 
  | group_by(.Source) 
  | .[] 
  | "Source: \(.[0].Source)\n" + (
      .[] | "  - \(.Headword): \(.Content)"
    ) + "\n"
'
```

---

## Workflow Quick Reference Table

| Use Case | Command | Format | Notes |
|----------|---------|--------|-------|
| Quick lookup | `dlexa palabra` | markdown | Human-readable, default |
| Script parsing | `dlexa --format json palabra` | json | Use with jq |
| Force refresh | `dlexa --no-cache palabra` | markdown | Bypass 24h cache |
| Specific source | `dlexa --source dpd palabra` | markdown | Query single source |
| Health check | `dlexa --doctor` | text | Exit 0 = healthy |
| Version info | `dlexa --version` | text | Print version |
| Error handling | `dlexa palabra 2>&1` | any | Capture both stdout/stderr |
| Check cache status | `dlexa --format json palabra \| jq .CacheHit` | json | Boolean: true/false |
| Extract headwords | `dlexa --format json palabra \| jq -r '.Entries[].Headword'` | json | Array of strings |
| Count results | `dlexa --format json palabra \| jq '.Entries \| length'` | json | Integer |

---

## Advanced Integration Examples

### CI/CD Dictionary Validation

```bash
#!/bin/bash
# Validate that a list of words exists in the dictionary

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
    echo "WARNING: No definitions for '$word'" >&2
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

### JSON to Markdown Converter

```bash
#!/bin/bash
# Convert JSON output to human-readable markdown

result=$(dlexa --format json "$1")

echo "# Definitions for: $1"
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
- Always handle empty `Entries` arrays (not an error condition)
- Problem codes in stderr indicate fatal errors (exit code 1)
- Cache TTL is 24 hours (not configurable from CLI)

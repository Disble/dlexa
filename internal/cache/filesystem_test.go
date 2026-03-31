package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

const (
	errFmtGetWantNil = "Get error = %v, want nil"
	errFmtGetZero    = "Get result = %+v, want zero value"
	errFmtSet        = "Set error = %v"
	errFmtGet        = "Get error = %v"
	errFmtGetQuery   = "Get result.Request.Query = %q, want %q"
	keyExpire        = "expire-key"
	keyTTL           = "ttl-key"
	fmtFilename      = "%x.json"
	keyShared        = "shared-key"
)

func sampleLookupResult(query string) model.LookupResult {
	return model.LookupResult{
		Request: model.LookupRequest{Query: query},
		Entries: []model.Entry{
			{
				ID:       "entry-1",
				Headword: query,
				Summary:  "A summary for " + query,
				Content:  "Content for " + query,
				Source:   "dpd",
				URL:      "https://example.com/" + query,
			},
		},
		Sources: []model.SourceResult{
			{
				Source: model.SourceDescriptor{
					Name:        "dpd",
					DisplayName: "DPD",
					Kind:        "remote-html",
					Priority:    1,
					Cacheable:   true,
				},
				Entries: []model.Entry{
					{ID: "entry-1", Headword: query},
				},
			},
		},
		GeneratedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}

func TestFilesystemStoreColdMiss(t *testing.T) {
	dir := t.TempDir()
	store := NewFilesystemStore(dir, 24*time.Hour)
	ctx := context.Background()

	result, ok, err := store.Get(ctx, "nonexistent-key")
	if err != nil {
		t.Fatalf(errFmtGetWantNil, err)
	}
	if ok {
		t.Fatal("Get ok = true, want false for cold store")
	}
	if result.Request.Query != "" {
		t.Errorf(errFmtGetZero, result)
	}
}

func TestFilesystemStoreSetThenGet(t *testing.T) {
	dir := t.TempDir()
	store := NewFilesystemStore(dir, 24*time.Hour)
	ctx := context.Background()

	expected := sampleLookupResult("haber")

	if err := store.Set(ctx, "my-key", expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	result, ok, err := store.Get(ctx, "my-key")
	if err != nil {
		t.Fatalf(errFmtGet, err)
	}
	if !ok {
		t.Fatal("Get ok = false, want true after Set")
	}
	if result.Request.Query != expected.Request.Query {
		t.Errorf(errFmtGetQuery, result.Request.Query, expected.Request.Query)
	}
	if len(result.Entries) != len(expected.Entries) {
		t.Fatalf("Get result.Entries len = %d, want %d", len(result.Entries), len(expected.Entries))
	}
	if result.Entries[0].ID != expected.Entries[0].ID {
		t.Errorf("Get result.Entries[0].ID = %q, want %q", result.Entries[0].ID, expected.Entries[0].ID)
	}
	if result.Entries[0].Headword != expected.Entries[0].Headword {
		t.Errorf("Get result.Entries[0].Headword = %q, want %q", result.Entries[0].Headword, expected.Entries[0].Headword)
	}
}

func TestFilesystemStorePersistenceAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	expected := sampleLookupResult("persistencia")

	// Write with first instance.
	store1 := NewFilesystemStore(dir, 24*time.Hour)
	if err := store1.Set(ctx, "persist-key", expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	// Read with a new instance on the same directory.
	store2 := NewFilesystemStore(dir, 24*time.Hour)
	result, ok, err := store2.Get(ctx, "persist-key")
	if err != nil {
		t.Fatalf(errFmtGet, err)
	}
	if !ok {
		t.Fatal("Get ok = false, want true for persisted entry")
	}
	if result.Request.Query != expected.Request.Query {
		t.Errorf(errFmtGetQuery, result.Request.Query, expected.Request.Query)
	}
}

func TestFilesystemStoreExpiredEntry(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	// Use an injectable clock: start at a fixed time, then advance.
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	store := NewFilesystemStore(dir, 1*time.Second)
	store.now = func() time.Time { return now }

	expected := sampleLookupResult("expirado")
	if err := store.Set(ctx, keyExpire, expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	// Verify it's a hit immediately.
	_, ok, _ := store.Get(ctx, keyExpire)
	if !ok {
		t.Fatal("Get ok = false, want true before expiry")
	}

	// Advance clock past TTL.
	now = now.Add(2 * time.Second)
	store.now = func() time.Time { return now }

	result, ok, err := store.Get(ctx, keyExpire)
	if err != nil {
		t.Fatalf(errFmtGetWantNil, err)
	}
	if ok {
		t.Fatal("Get ok = true, want false for expired entry")
	}
	if result.Request.Query != "" {
		t.Errorf(errFmtGetZero, result)
	}
}

func TestFilesystemStoreNonExpiredEntry(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	store := NewFilesystemStore(dir, 24*time.Hour)
	store.now = func() time.Time { return now }

	expected := sampleLookupResult("vigente")
	if err := store.Set(ctx, "valid-key", expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	// Advance 1 hour (within 24h TTL).
	now = now.Add(1 * time.Hour)
	store.now = func() time.Time { return now }

	result, ok, err := store.Get(ctx, "valid-key")
	if err != nil {
		t.Fatalf(errFmtGet, err)
	}
	if !ok {
		t.Fatal("Get ok = false, want true for non-expired entry")
	}
	if result.Request.Query != expected.Request.Query {
		t.Errorf(errFmtGetQuery, result.Request.Query, expected.Request.Query)
	}
}

func TestFilesystemStoreCustomTTL(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	store := NewFilesystemStore(dir, 1*time.Hour)
	store.now = func() time.Time { return now }

	expected := sampleLookupResult("custom-ttl")
	if err := store.Set(ctx, keyTTL, expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	// Advance 30 minutes: should still be valid.
	now = now.Add(30 * time.Minute)
	store.now = func() time.Time { return now }

	_, ok, _ := store.Get(ctx, keyTTL)
	if !ok {
		t.Fatal("Get ok = false, want true within 1h TTL")
	}

	// Advance to 61 minutes total: should be expired.
	now = now.Add(31 * time.Minute)
	store.now = func() time.Time { return now }

	_, ok, _ = store.Get(ctx, keyTTL)
	if ok {
		t.Fatal("Get ok = true, want false after 1h TTL expired")
	}
}

func TestFilesystemStoreCorruptFile(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store := NewFilesystemStore(dir, 24*time.Hour)

	// Write a corrupt file directly to the cache directory.
	key := "corrupt-key"
	hash := sha256.Sum256([]byte(key))
	filename := fmt.Sprintf(fmtFilename, hash)
	corruptPath := filepath.Join(dir, filename)

	if err := os.WriteFile(corruptPath, []byte("this is not valid json{{{"), 0o600); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	result, ok, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf(errFmtGetWantNil, err)
	}
	if ok {
		t.Fatal("Get ok = true, want false for corrupt file")
	}
	if result.Request.Query != "" {
		t.Errorf(errFmtGetZero, result)
	}

	// Corrupt file should be removed.
	if _, statErr := os.Stat(corruptPath); !os.IsNotExist(statErr) {
		t.Error("corrupt file should be removed after failed Get")
	}
}

func TestFilesystemStoreSpecialCharKey(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	store := NewFilesystemStore(dir, 24*time.Hour)

	keys := []string{
		"key/with/slashes",
		"key|with|pipes",
		"key:with:colons",
		"key with spaces",
		"key?with*wildcards",
		`key"with"quotes`,
		"key<with>angles",
		"ñoño|café/über:test",
	}

	for _, key := range keys {
		expected := sampleLookupResult(key)
		if err := store.Set(ctx, key, expected); err != nil {
			t.Errorf("Set(%q) error = %v", key, err)
			continue
		}

		result, ok, err := store.Get(ctx, key)
		if err != nil {
			t.Errorf("Get(%q) error = %v", key, err)
			continue
		}
		if !ok {
			t.Errorf("Get(%q) ok = false, want true", key)
			continue
		}
		if result.Request.Query != expected.Request.Query {
			t.Errorf("Get(%q) result.Request.Query = %q, want %q", key, result.Request.Query, expected.Request.Query)
		}
	}
}

func TestFilesystemStoreDistinctKeys(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma", "alpha/beta", "alpha|beta"}
	seen := make(map[string]string) // filename -> original key

	for _, key := range keys {
		hash := sha256.Sum256([]byte(key))
		filename := fmt.Sprintf(fmtFilename, hash)

		if original, exists := seen[filename]; exists {
			t.Fatalf("keys %q and %q produce the same filename %q", original, key, filename)
		}
		seen[filename] = key
	}
}

func TestFilesystemStoreNonExistentDirectory(t *testing.T) {
	// Use a directory that doesn't exist yet — should create on first Set.
	base := t.TempDir()
	dir := filepath.Join(base, "subdir", "cache")
	ctx := context.Background()

	store := NewFilesystemStore(dir, 24*time.Hour)

	expected := sampleLookupResult("nested-dir")
	if err := store.Set(ctx, "nested-key", expected); err != nil {
		t.Fatalf("Set error = %v, want nil (should create dir)", err)
	}

	result, ok, err := store.Get(ctx, "nested-key")
	if err != nil {
		t.Fatalf(errFmtGet, err)
	}
	if !ok {
		t.Fatal("Get ok = false, want true")
	}
	if result.Request.Query != expected.Request.Query {
		t.Errorf("result.Request.Query = %q, want %q", result.Request.Query, expected.Request.Query)
	}
}

func TestFilesystemStoreReadOnlyDirectoryGetReturnssMiss(t *testing.T) {
	// Get on a non-existent directory should return miss, not error.
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	ctx := context.Background()

	store := NewFilesystemStore(dir, 24*time.Hour)

	result, ok, err := store.Get(ctx, "any-key")
	if err != nil {
		t.Fatalf(errFmtGetWantNil, err)
	}
	if ok {
		t.Fatal("Get ok = true, want false")
	}
	if result.Request.Query != "" {
		t.Errorf(errFmtGetZero, result)
	}
}

func TestFilesystemStoreEnvelopeFormat(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	store := NewFilesystemStore(dir, 24*time.Hour)
	store.now = func() time.Time { return now }

	expected := sampleLookupResult("formato")
	if err := store.Set(ctx, "format-key", expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	// Read the raw file and verify the envelope structure.
	hash := sha256.Sum256([]byte("format-key"))
	filename := fmt.Sprintf(fmtFilename, hash)
	data, err := os.ReadFile(filepath.Join(dir, filename)) //nolint:gosec // G304: test code, path constructed from known hash
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}

	var envelope struct {
		ExpiresAt time.Time          `json:"expires_at"`
		CreatedAt time.Time          `json:"created_at"`
		Data      model.LookupResult `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	expectedExpiry := now.Add(24 * time.Hour)
	if !envelope.ExpiresAt.Equal(expectedExpiry) {
		t.Errorf("envelope.ExpiresAt = %v, want %v", envelope.ExpiresAt, expectedExpiry)
	}
	if !envelope.CreatedAt.Equal(now) {
		t.Errorf("envelope.CreatedAt = %v, want %v", envelope.CreatedAt, now)
	}
	if envelope.Data.Request.Query != expected.Request.Query {
		t.Errorf("envelope.Data.Request.Query = %q, want %q", envelope.Data.Request.Query, expected.Request.Query)
	}
}

// Task 1.2: Concurrent read/write race-detector test.
func TestFilesystemStoreConcurrentReadWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewFilesystemStore(dir, 24*time.Hour)
	runConcurrentStoreTest(t, store, 10, 10, 50)
}

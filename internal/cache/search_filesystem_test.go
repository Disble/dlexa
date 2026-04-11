package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

const (
	errFmtSearchGetZero = "Get result = %#v, want zero value"
	hashJSONFmt         = "%x.json"
)

func sampleSearchResult(query string) model.SearchResult {
	return model.SearchResult{
		Request: model.SearchRequest{Query: query},
		Candidates: []model.SearchCandidate{{
			RawLabelHTML: `<em>` + query + `</em>`,
			DisplayText:  query,
			ArticleKey:   query,
		}},
		GeneratedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}

func TestSearchFilesystemStoreColdMiss(t *testing.T) {
	dir := t.TempDir()
	store := NewSearchFilesystemStore(dir, 24*time.Hour)

	result, ok, err := store.Get(context.Background(), "nonexistent-key")
	if err != nil {
		t.Fatalf(errFmtGetWantNil, err)
	}
	if ok {
		t.Fatal("Get ok = true, want false for cold store")
	}
	if result.Request.Query != "" {
		t.Fatalf(errFmtSearchGetZero, result)
	}
}

func TestSearchFilesystemStoreSetThenGet(t *testing.T) {
	dir := t.TempDir()
	store := NewSearchFilesystemStore(dir, 24*time.Hour)
	key := BuildSearchKey(model.SearchRequest{Query: " Abu   Dhabi ", Format: "json"})
	expected := sampleSearchResult("abu dhabi")

	if err := store.Set(context.Background(), key, expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	result, ok, err := store.Get(context.Background(), key)
	if err != nil {
		t.Fatalf(errFmtGet, err)
	}
	if !ok {
		t.Fatal("Get ok = false, want true after Set")
	}
	if result.Request.Query != expected.Request.Query {
		t.Fatalf(errFmtGetQuery, result.Request.Query, expected.Request.Query)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].ArticleKey != expected.Candidates[0].ArticleKey {
		t.Fatalf("Get result = %#v, want %#v", result, expected)
	}
}

func TestSearchFilesystemStorePersistenceAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	key := BuildSearchKey(model.SearchRequest{Query: "persistencia"})
	expected := sampleSearchResult("persistencia")

	store1 := NewSearchFilesystemStore(dir, 24*time.Hour)
	if err := store1.Set(context.Background(), key, expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	store2 := NewSearchFilesystemStore(dir, 24*time.Hour)
	result, ok, err := store2.Get(context.Background(), key)
	if err != nil {
		t.Fatalf(errFmtGet, err)
	}
	if !ok {
		t.Fatal("Get ok = false, want true for persisted entry")
	}
	if result.Request.Query != expected.Request.Query {
		t.Fatalf(errFmtGetQuery, result.Request.Query, expected.Request.Query)
	}
}

func TestSearchFilesystemStoreExpiredEntry(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	store := NewSearchFilesystemStore(dir, time.Second)
	store.now = func() time.Time { return now }
	key := BuildSearchKey(model.SearchRequest{Query: "expirado"})

	if err := store.Set(context.Background(), key, sampleSearchResult("expirado")); err != nil {
		t.Fatalf(errFmtSet, err)
	}
	if _, ok, _ := store.Get(context.Background(), key); !ok {
		t.Fatal("Get ok = false, want true before expiry")
	}

	now = now.Add(2 * time.Second)
	store.now = func() time.Time { return now }
	result, ok, err := store.Get(context.Background(), key)
	if err != nil {
		t.Fatalf(errFmtGetWantNil, err)
	}
	if ok {
		t.Fatal("Get ok = true, want false for expired entry")
	}
	if result.Request.Query != "" {
		t.Fatalf(errFmtSearchGetZero, result)
	}
}

func TestSearchFilesystemStoreCorruptFile(t *testing.T) {
	dir := t.TempDir()
	store := NewSearchFilesystemStore(dir, 24*time.Hour)
	key := BuildSearchKey(model.SearchRequest{Query: "corrupt-key"})
	hash := sha256.Sum256([]byte(key))
	corruptPath := filepath.Join(dir, fmt.Sprintf(hashJSONFmt, hash))
	if err := os.WriteFile(corruptPath, []byte("this is not valid json{{{"), 0o600); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	result, ok, err := store.Get(context.Background(), key)
	if err != nil {
		t.Fatalf(errFmtGetWantNil, err)
	}
	if ok {
		t.Fatal("Get ok = true, want false for corrupt file")
	}
	if result.Request.Query != "" {
		t.Fatalf(errFmtSearchGetZero, result)
	}
	if _, statErr := os.Stat(corruptPath); !os.IsNotExist(statErr) {
		t.Fatal("corrupt file should be removed after failed Get")
	}
}

func TestSearchFilesystemStoreCreatesNestedDirectoryOnWrite(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "subdir", "cache")
	store := NewSearchFilesystemStore(dir, 24*time.Hour)
	key := BuildSearchKey(model.SearchRequest{Query: "nested-dir"})
	expected := sampleSearchResult("nested-dir")

	if err := store.Set(context.Background(), key, expected); err != nil {
		t.Fatalf("Set error = %v, want nil (should create dir)", err)
	}

	result, ok, err := store.Get(context.Background(), key)
	if err != nil {
		t.Fatalf(errFmtGet, err)
	}
	if !ok {
		t.Fatal("Get ok = false, want true")
	}
	if result.Request.Query != expected.Request.Query {
		t.Fatalf(errFmtGetQuery, result.Request.Query, expected.Request.Query)
	}
}

func TestSearchFilesystemStoreEnvelopeFormat(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	store := NewSearchFilesystemStore(dir, 24*time.Hour)
	store.now = func() time.Time { return now }
	key := BuildSearchKey(model.SearchRequest{Query: "formato"})
	expected := sampleSearchResult("formato")

	if err := store.Set(context.Background(), key, expected); err != nil {
		t.Fatalf(errFmtSet, err)
	}

	hash := sha256.Sum256([]byte(key))
	data, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf(hashJSONFmt, hash))) //nolint:gosec // G304: test reads a hashed cache file name in temp dir
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}

	var envelope struct {
		ExpiresAt time.Time          `json:"expires_at"`
		CreatedAt time.Time          `json:"created_at"`
		Data      model.SearchResult `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if !envelope.CreatedAt.Equal(now) {
		t.Fatalf("envelope.CreatedAt = %v, want %v", envelope.CreatedAt, now)
	}
	if !envelope.ExpiresAt.Equal(now.Add(24 * time.Hour)) {
		t.Fatalf("envelope.ExpiresAt = %v, want %v", envelope.ExpiresAt, now.Add(24*time.Hour))
	}
	if envelope.Data.Request.Query != expected.Request.Query {
		t.Fatalf("envelope.Data.Request.Query = %q, want %q", envelope.Data.Request.Query, expected.Request.Query)
	}
}

func TestSearchFilesystemStorePersistsSearchResultsWithoutRendererBlobs(t *testing.T) {
	dir := t.TempDir()
	store := NewSearchFilesystemStore(dir, 24*time.Hour)
	key := BuildSearchKey(model.SearchRequest{Query: "abu dhabi", Format: "json"})
	result := model.SearchResult{Candidates: []model.SearchCandidate{{RawLabelHTML: `<em>Abu Dhabi</em>`, DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi"}}}

	if err := store.Set(context.Background(), key, result); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	cached, ok, err := store.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if len(cached.Candidates) != 1 || cached.Candidates[0].ArticleKey != "Abu Dabi" {
		t.Fatalf("cached = %#v", cached)
	}

	hash := sha256.Sum256([]byte(key))
	raw, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf(hashJSONFmt, hash))) //nolint:gosec // G304: test reads a hashed cache file name in temp dir
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(raw), "->") {
		t.Fatalf("cache blob = %s, must not store renderer output", string(raw))
	}
}

func TestSearchFilesystemStoreConcurrentReadWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewSearchFilesystemStore(dir, 24*time.Hour)
	runConcurrentSearchStoreTest(t, store, 10, 10, 50)
}

package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

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
	raw, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%x.json", hash))) //nolint:gosec // G304: test reads a hashed cache file name in temp dir
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(raw), "->") {
		t.Fatalf("cache blob = %s, must not store renderer output", string(raw))
	}
}

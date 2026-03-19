package cache

import (
	"testing"
)

func TestMemoryStoreConcurrentReadWrite(t *testing.T) {
	store := NewMemoryStore()
	runConcurrentStoreTest(t, store, 10, 10, 100)
}

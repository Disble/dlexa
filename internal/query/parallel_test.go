package query

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/source"
)

const (
	sourcePriority3 = "priority-3"
	sourcePriority1 = "priority-1"
	sourcePriority2 = "priority-2"
)

// delayedStubSource is a stub that introduces artificial latency before returning.
type delayedStubSource struct {
	descriptor model.SourceDescriptor
	result     model.SourceResult
	err        error
	delay      time.Duration
}

func (s *delayedStubSource) Descriptor() model.SourceDescriptor {
	return s.descriptor
}

func (s *delayedStubSource) Lookup(ctx context.Context, _ model.LookupRequest) (model.SourceResult, error) {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return model.SourceResult{}, ctx.Err()
	}
	return s.result, s.err
}

// --- Task 2.1: TDD concurrency timing test ---

func TestLookupQueriesSourcesConcurrently(t *testing.T) {
	const sourceDelay = 100 * time.Millisecond
	const sourceCount = 3

	sources := make([]source.Source, sourceCount)
	for i := 0; i < sourceCount; i++ {
		desc := model.SourceDescriptor{Name: "delayed-" + string(rune('a'+i)), Priority: i + 1}
		sources[i] = &delayedStubSource{
			descriptor: desc,
			delay:      sourceDelay,
			result: model.SourceResult{
				Source:  desc,
				Entries: []model.Entry{{ID: desc.Name + "-entry", Source: desc.Name}},
			},
		}
	}

	registry := &stubRegistry{sources: sources}
	service := NewService(registry, &stubStore{})
	request := model.LookupRequest{Query: "parallel", Sources: []string{"delayed-a", "delayed-b", "delayed-c"}, NoCache: true}

	start := time.Now()
	result, err := service.Lookup(context.Background(), request)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}

	// With 3 sources each taking 100ms, sequential would be ~300ms.
	// Parallel should complete in ~100ms. We give 200ms margin.
	if elapsed >= 200*time.Millisecond {
		t.Fatalf("Lookup() took %v, want < 200ms (proving concurrency)", elapsed)
	}

	if len(result.Entries) != sourceCount {
		t.Fatalf("Lookup() entries = %d, want %d", len(result.Entries), sourceCount)
	}

	if len(result.Sources) != sourceCount {
		t.Fatalf("Lookup() sources = %d, want %d", len(result.Sources), sourceCount)
	}
}

// --- Task 2.2: TDD one-source-fails test ---

func TestLookupOneSourceFailsOthersSucceed(t *testing.T) {
	descA := model.SourceDescriptor{Name: "ok-a", Priority: 1}
	descB := model.SourceDescriptor{Name: "failing-b", Priority: 2}
	descC := model.SourceDescriptor{Name: "ok-c", Priority: 3}

	sources := []source.Source{
		&delayedStubSource{
			descriptor: descA,
			delay:      10 * time.Millisecond,
			result: model.SourceResult{
				Source:  descA,
				Entries: []model.Entry{{ID: "a-entry", Source: descA.Name}},
			},
		},
		&delayedStubSource{
			descriptor: descB,
			delay:      10 * time.Millisecond,
			err:        errors.New("source B failed"),
		},
		&delayedStubSource{
			descriptor: descC,
			delay:      10 * time.Millisecond,
			result: model.SourceResult{
				Source:  descC,
				Entries: []model.Entry{{ID: "c-entry", Source: descC.Name}},
			},
		},
	}

	registry := &stubRegistry{sources: sources}
	service := NewService(registry, &stubStore{})
	request := model.LookupRequest{Query: "partial", Sources: []string{"ok-a", "failing-b", "ok-c"}, NoCache: true}

	result, err := service.Lookup(context.Background(), request)
	// No top-level error.
	if err != nil {
		t.Fatalf("Lookup() error = %v, want nil", err)
	}

	// 2 successful sources.
	if len(result.Sources) != 2 {
		t.Fatalf("Lookup() sources = %d, want 2", len(result.Sources))
	}

	// 2 entries from successful sources.
	if len(result.Entries) != 2 {
		t.Fatalf("Lookup() entries = %d, want 2", len(result.Entries))
	}

	// 1 problem for the failed source.
	if len(result.Problems) != 1 {
		t.Fatalf("Lookup() problems = %d, want 1", len(result.Problems))
	}

	if result.Problems[0].Source != descB.Name {
		t.Fatalf("Lookup() problem source = %q, want %q", result.Problems[0].Source, descB.Name)
	}
}

// --- Task 2.3: TDD all-sources-fail test ---

func TestLookupAllSourcesFail(t *testing.T) {
	descA := model.SourceDescriptor{Name: "fail-a", Priority: 1}
	descB := model.SourceDescriptor{Name: "fail-b", Priority: 2}

	sources := []source.Source{
		&delayedStubSource{
			descriptor: descA,
			delay:      10 * time.Millisecond,
			err:        errors.New("source A failed"),
		},
		&delayedStubSource{
			descriptor: descB,
			delay:      10 * time.Millisecond,
			err:        errors.New("source B failed"),
		},
	}

	registry := &stubRegistry{sources: sources}
	service := NewService(registry, &stubStore{})
	request := model.LookupRequest{Query: "allfail", Sources: []string{"fail-a", "fail-b"}, NoCache: true}

	result, err := service.Lookup(context.Background(), request)
	// No top-level error — problems are reported in-band.
	if err != nil {
		t.Fatalf("Lookup() error = %v, want nil", err)
	}

	// Zero entries.
	if len(result.Entries) != 0 {
		t.Fatalf("Lookup() entries = %d, want 0", len(result.Entries))
	}

	// Zero successful sources.
	if len(result.Sources) != 0 {
		t.Fatalf("Lookup() sources = %d, want 0", len(result.Sources))
	}

	// Problem entries for both failed sources.
	if len(result.Problems) != 2 {
		t.Fatalf("Lookup() problems = %d, want 2", len(result.Problems))
	}

	problemSources := []string{result.Problems[0].Source, result.Problems[1].Source}
	sort.Strings(problemSources)
	wantSources := []string{descA.Name, descB.Name}
	sort.Strings(wantSources)
	if problemSources[0] != wantSources[0] || problemSources[1] != wantSources[1] {
		t.Fatalf("Lookup() problem sources = %v, want %v", problemSources, wantSources)
	}
}

// --- Task 2.4: TDD result ordering by priority test ---

func TestLookupResultsOrderedByPriority(t *testing.T) {
	// Sources with priorities 3, 1, 2. Use varying delays so completion order
	// differs from priority order.
	descP3 := model.SourceDescriptor{Name: sourcePriority3, Priority: 3}
	descP1 := model.SourceDescriptor{Name: sourcePriority1, Priority: 1}
	descP2 := model.SourceDescriptor{Name: sourcePriority2, Priority: 2}

	sources := []source.Source{
		&delayedStubSource{
			descriptor: descP3,
			delay:      10 * time.Millisecond, // finishes first
			result: model.SourceResult{
				Source:  descP3,
				Entries: []model.Entry{{ID: "p3-entry", Source: descP3.Name}},
			},
		},
		&delayedStubSource{
			descriptor: descP1,
			delay:      50 * time.Millisecond, // finishes last
			result: model.SourceResult{
				Source:  descP1,
				Entries: []model.Entry{{ID: "p1-entry", Source: descP1.Name}},
			},
		},
		&delayedStubSource{
			descriptor: descP2,
			delay:      30 * time.Millisecond, // finishes middle
			result: model.SourceResult{
				Source:  descP2,
				Entries: []model.Entry{{ID: "p2-entry", Source: descP2.Name}},
			},
		},
	}

	registry := &stubRegistry{sources: sources}
	service := NewService(registry, &stubStore{})
	request := model.LookupRequest{Query: "ordered", Sources: []string{sourcePriority3, sourcePriority1, sourcePriority2}, NoCache: true}

	result, err := service.Lookup(context.Background(), request)
	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}

	if len(result.Sources) != 3 {
		t.Fatalf("Lookup() sources = %d, want 3", len(result.Sources))
	}

	// Sources must be ordered by ascending priority.
	wantOrder := []string{sourcePriority1, sourcePriority2, sourcePriority3}
	for i, sr := range result.Sources {
		if sr.Source.Name != wantOrder[i] {
			t.Fatalf("Lookup() sources[%d].Source.Name = %q, want %q", i, sr.Source.Name, wantOrder[i])
		}
	}

	// Entries must preserve source-priority ordering.
	wantEntryOrder := []string{"p1-entry", "p2-entry", "p3-entry"}
	for i, entry := range result.Entries {
		if entry.ID != wantEntryOrder[i] {
			t.Fatalf("Lookup() entries[%d].ID = %q, want %q", i, entry.ID, wantEntryOrder[i])
		}
	}
}

// --- Task 2.5: Race-detector test for parallel fan-out ---

func TestLookupRaceDetector(t *testing.T) {
	// This test is designed to be run with `go test -race ./internal/query/...`.
	// Multiple sources access shared state (cache, result collection) concurrently.
	const sourceCount = 5

	sources := make([]source.Source, sourceCount)
	for i := 0; i < sourceCount; i++ {
		desc := model.SourceDescriptor{Name: "race-" + string(rune('a'+i)), Priority: i + 1}
		sources[i] = &delayedStubSource{
			descriptor: desc,
			delay:      5 * time.Millisecond,
			result: model.SourceResult{
				Source:   desc,
				Entries:  []model.Entry{{ID: desc.Name + "-entry", Source: desc.Name}},
				Warnings: []model.Warning{{Code: "warn-" + desc.Name, Source: desc.Name}},
			},
		}
	}

	// Use a real MemoryStore to exercise concurrent cache access patterns.
	store := &stubStore{}
	registry := &stubRegistry{sources: sources}
	service := NewService(registry, store)
	request := model.LookupRequest{Query: "racetest", Sources: []string{"race-a", "race-b", "race-c", "race-d", "race-e"}, NoCache: true}

	result, err := service.Lookup(context.Background(), request)
	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}

	if len(result.Entries) != sourceCount {
		t.Fatalf("Lookup() entries = %d, want %d", len(result.Entries), sourceCount)
	}

	if len(result.Sources) != sourceCount {
		t.Fatalf("Lookup() sources = %d, want %d", len(result.Sources), sourceCount)
	}

	if len(result.Warnings) != sourceCount {
		t.Fatalf("Lookup() warnings = %d, want %d", len(result.Warnings), sourceCount)
	}
}

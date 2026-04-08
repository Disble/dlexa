// Package inflight provides keyed request coalescing helpers.
package inflight

import (
	"context"
	"sync"
)

// Group coalesces concurrent calls for the same key.
type Group[T any] struct {
	mu    sync.Mutex
	calls map[string]*call[T]
}

type call[T any] struct {
	done chan struct{}
	val  T
	err  error
}

// Do joins an in-flight call when a matching key is already executing.
func (g *Group[T]) Do(ctx context.Context, key string, fn func(context.Context) (T, error)) (T, bool, error) {
	g.mu.Lock()
	if g.calls == nil {
		g.calls = make(map[string]*call[T])
	}
	if existing, ok := g.calls[key]; ok {
		g.mu.Unlock()
		select {
		case <-existing.done:
			return existing.val, true, existing.err
		case <-ctx.Done():
			var zero T
			return zero, true, ctx.Err()
		}
	}

	current := &call[T]{done: make(chan struct{})}
	g.calls[key] = current
	g.mu.Unlock()

	current.val, current.err = fn(ctx)
	close(current.done)

	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()

	return current.val, false, current.err
}

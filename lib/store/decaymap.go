package store

import (
	"context"
	"fmt"
	"time"

	"github.com/TecharoHQ/anubis/decaymap"
)

type decayMapStore struct {
	store *decaymap.Impl[string, []byte]
}

func (d *decayMapStore) Delete(_ context.Context, key string) error {
	if !d.store.Delete(key) {
		return fmt.Errorf("%w: %q", ErrNotFound, key)
	}

	return nil
}

func (d *decayMapStore) Get(_ context.Context, key string) ([]byte, error) {
	result, ok := d.store.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, key)
	}

	return result, nil
}

func (d *decayMapStore) Set(_ context.Context, key string, value []byte, expiry time.Duration) error {
	d.store.Set(key, value, expiry)
	return nil
}

func (d *decayMapStore) cleanupThread(ctx context.Context) {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			d.store.Cleanup()
		}
	}
}

// NewDecayMapStore creates a simple in-memory store. This will not scale
// to multiple Anubis instances.
func NewDecayMapStore(ctx context.Context) Interface {
	result := &decayMapStore{
		store: decaymap.New[string, []byte](),
	}

	go result.cleanupThread(ctx)

	return result
}

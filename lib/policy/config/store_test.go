package config_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/TecharoHQ/anubis/lib/policy/config"
	"github.com/TecharoHQ/anubis/lib/store/bbolt"
)

func TestStoreValid(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input config.Store
		err   error
	}{
		{
			name:  "no backend",
			input: config.Store{},
			err:   config.ErrNoStoreBackend,
		},
		{
			name: "in-memory backend",
			input: config.Store{
				Backend: "memory",
			},
		},
		{
			name: "bbolt backend",
			input: config.Store{
				Backend:    "bbolt",
				Parameters: json.RawMessage(`{"path": "/tmp/foo", "bucket": "bar"}`),
			},
		},
		{
			name: "bbolt backend no path",
			input: config.Store{
				Backend:    "bbolt",
				Parameters: json.RawMessage(`{"path": "", "bucket": "bar"}`),
			},
			err: bbolt.ErrMissingPath,
		},
		{
			name: "unknown backend",
			input: config.Store{
				Backend: "taco salad",
			},
			err: config.ErrUnknownStoreBackend,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.input.Valid(); !errors.Is(err, tt.err) {
				t.Logf("want: %v", tt.err)
				t.Logf("got:  %v", err)
				t.Error("invalid error returned")
			}
		})
	}
}

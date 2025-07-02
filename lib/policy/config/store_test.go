package config_test

import (
	"errors"
	"testing"

	"github.com/TecharoHQ/anubis/lib/policy/config"
	_ "github.com/TecharoHQ/anubis/lib/store/all"
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

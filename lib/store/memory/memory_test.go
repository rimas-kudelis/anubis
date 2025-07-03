package memory

import (
	"testing"

	"github.com/TecharoHQ/anubis/lib/store/storetest"
)

func TestImpl(t *testing.T) {
	storetest.Common(t, New(t.Context()))
}

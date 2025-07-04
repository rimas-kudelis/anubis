package valkey

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/TecharoHQ/anubis/lib/store/storetest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestImpl(t *testing.T) {
	if os.Getenv("DONT_USE_NETWORK") != "" {
		t.Skip("test requires network egress")
		return
	}

	testcontainers.SkipIfProviderIsNotHealthy(t)

	req := testcontainers.ContainerRequest{
		Image:      "valkey/valkey:8",
		WaitingFor: wait.ForLog("Ready to accept connections"),
	}
	valkeyC, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	testcontainers.CleanupContainer(t, valkeyC)
	if err != nil {
		t.Fatal(err)
	}

	containerIP, err := valkeyC.ContainerIP(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	// XXX(Xe): This is bad code. Do not do this.
	//
	// I have to do this because I'm running from inside the context of a dev
	// container. This dev container runs in a different docker network than
	// the valkey test container runs in. In order to let my dev container
	// connect to the test container, they need to share a network in common.
	// The easiest network to use for this is the default "bridge" network.
	//
	// This is a horrifying monstrosity, but the part that scares me the most
	// is the fact that it works.
	if hostname, err := os.Hostname(); err == nil {
		exec.Command("docker", "network", "connect", "bridge", hostname).Run()
	}

	data, err := json.Marshal(Config{
		URL: fmt.Sprintf("redis://%s:6379/0", containerIP),
	})
	if err != nil {
		t.Fatal(err)
	}

	storetest.Common(t, Factory{}, json.RawMessage(data))
}

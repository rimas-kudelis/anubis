package entrypoint

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/TecharoHQ/anubis/cmd/osiris/internal/config"
	"github.com/TecharoHQ/anubis/internal"
	"github.com/hashicorp/hcl/v2/hclsimple"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

type Options struct {
	ConfigFname string
}

func Main(opts Options) error {
	internal.SetHealth("osiris", healthv1.HealthCheckResponse_NOT_SERVING)

	var cfg config.Toplevel
	if err := hclsimple.DecodeFile(opts.ConfigFname, nil, &cfg); err != nil {
		return fmt.Errorf("can't read configuration file %s:\n\n%w", opts.ConfigFname, err)
	}

	if err := cfg.Valid(); err != nil {
		return fmt.Errorf("configuration file %s is invalid:\n\n%w", opts.ConfigFname, err)
	}

	rtr, err := NewRouter(cfg)
	if err != nil {
		return err
	}

	slog.Info("listening on", "http", cfg.Bind.HTTP)
	return http.ListenAndServe(cfg.Bind.HTTP, rtr)
}

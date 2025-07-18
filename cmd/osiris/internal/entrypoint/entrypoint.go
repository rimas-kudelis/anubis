package entrypoint

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/TecharoHQ/anubis/cmd/osiris/internal/config"
	"github.com/TecharoHQ/anubis/internal"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

type Options struct {
	ConfigFname string
}

func Main(opts Options) error {
	internal.SetHealth("osiris", healthv1.HealthCheckResponse_NOT_SERVING)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	g, gCtx := errgroup.WithContext(ctx)

	// HTTP
	g.Go(func() error {
		ln, err := net.Listen("tcp", cfg.Bind.HTTP)
		if err != nil {
			return fmt.Errorf("(HTTP) can't bind to tcp %s: %w", cfg.Bind.HTTP, err)
		}
		defer ln.Close()

		go func(ctx context.Context) {
			<-ctx.Done()
			ln.Close()
		}(gCtx)

		slog.Info("listening for HTTP", "bind", cfg.Bind.HTTP)

		srv := http.Server{Handler: rtr, ErrorLog: internal.GetFilteredHTTPLogger()}

		return srv.Serve(ln)
	})

	// HTTPS

	// Metrics
	g.Go(func() error {
		ln, err := net.Listen("tcp", cfg.Bind.Metrics)
		if err != nil {
			return fmt.Errorf("(metrics) can't bind to tcp %s: %w", cfg.Bind.Metrics, err)
		}
		defer ln.Close()

		go func(ctx context.Context) {
			<-ctx.Done()
			ln.Close()
		}(gCtx)

		mux := http.NewServeMux()

		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			st, ok := internal.GetHealth("osiris")
			if !ok {
				slog.Error("health service osiris does not exist, file a bug")
			}

			switch st {
			case healthv1.HealthCheckResponse_NOT_SERVING:
				http.Error(w, "NOT OK", http.StatusInternalServerError)
				return
			case healthv1.HealthCheckResponse_SERVING:
				fmt.Fprintln(w, "OK")
				return
			default:
				http.Error(w, "UNKNOWN", http.StatusFailedDependency)
				return
			}
		})

		slog.Info("listening for Metrics", "bind", cfg.Bind.Metrics)

		srv := http.Server{Handler: mux, ErrorLog: internal.GetFilteredHTTPLogger()}

		return srv.Serve(ln)
	})

	return g.Wait()
}

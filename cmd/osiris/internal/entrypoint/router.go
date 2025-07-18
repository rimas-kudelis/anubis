package entrypoint

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/TecharoHQ/anubis/cmd/osiris/internal/config"
	"github.com/lum8rjack/go-ja4h"
)

var (
	ErrTargetInvalid = errors.New("[unexpected] target invalid")
	ErrNoHandler     = errors.New("[unexpected] no handler for domain")
)

type Router struct {
	lock   sync.RWMutex
	routes map[string]http.Handler
}

func (rtr *Router) setConfig(c config.Toplevel) error {
	var errs []error
	newMap := map[string]http.Handler{}

	for _, d := range c.Domains {
		var domainErrs []error

		u, err := url.Parse(d.Target)
		if err != nil {
			domainErrs = append(domainErrs, fmt.Errorf("%w %q: %v", ErrTargetInvalid, d.Target, err))
		}

		var h http.Handler

		switch u.Scheme {
		case "http", "https":
			h = httputil.NewSingleHostReverseProxy(u)
		case "h2c":
			h = newH2CReverseProxy(u)
		case "unix":
			h = &httputil.ReverseProxy{
				Transport: &http.Transport{
					DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
						return net.Dial("unix", strings.TrimPrefix(d.Target, "unix://"))
					},
				},
			}
		}

		if h == nil {
			domainErrs = append(domainErrs, ErrNoHandler)
		}

		newMap[d.Name] = h

		if len(domainErrs) != 0 {
			errs = append(errs, fmt.Errorf("invalid domain %s: %w", d.Name, errors.Join(domainErrs...)))
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("can't compile config to routing map: %w", errors.Join(errs...))
	}

	rtr.lock.Lock()
	rtr.routes = newMap
	rtr.lock.Unlock()

	return nil
}

func NewRouter(c config.Toplevel) (*Router, error) {
	result := &Router{
		routes: map[string]http.Handler{},
	}

	if err := result.setConfig(c); err != nil {
		return nil, err
	}

	fmt.Printf("%#v\n", result.routes)

	return result, nil
}

func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var h http.Handler
	var ok bool

	ja4hFP := ja4h.JA4H(r)

	slog.Info("got request", "method", r.Method, "host", r.Host, "path", r.URL.Path)

	rtr.lock.RLock()
	h, ok = rtr.routes[r.Host]
	rtr.lock.RUnlock()

	if !ok {
		http.NotFound(w, r) // TODO(Xe): brand this
		return
	}

	r.Header.Set("X-HTTP-JA4H-Fingerprint", ja4hFP)

	h.ServeHTTP(w, r)
}

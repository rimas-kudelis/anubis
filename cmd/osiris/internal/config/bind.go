package config

import (
	"errors"
	"fmt"
	"net"
)

var (
	ErrCantBindToPort = errors.New("bind: can't bind to host:port")
)

type Bind struct {
	HTTP    string `hcl:"http"`
	HTTPS   string `hcl:"https"`
	Metrics string `hcl:"metrics"`
}

func (b *Bind) Valid() error {
	var errs []error

	ln, err := net.Listen("tcp", b.HTTP)
	if err != nil {
		errs = append(errs, fmt.Errorf("%w %q: %w", ErrCantBindToPort, b.HTTP, err))
	} else {
		defer ln.Close()
	}

	ln, err = net.Listen("tcp", b.HTTPS)
	if err != nil {
		errs = append(errs, fmt.Errorf("%w %q: %w", ErrCantBindToPort, b.HTTPS, err))
	} else {
		defer ln.Close()
	}

	ln, err = net.Listen("tcp", b.Metrics)
	if err != nil {
		errs = append(errs, fmt.Errorf("%w %q: %w", ErrCantBindToPort, b.Metrics, err))
	} else {
		defer ln.Close()
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}

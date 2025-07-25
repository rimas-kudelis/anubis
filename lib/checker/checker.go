// Package checker defines the Checker interface and a helper utility to avoid import cycles.
package checker

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/TecharoHQ/anubis/internal"
)

var (
	ErrUnparseableConfig = errors.New("checker: config is unparseable")
	ErrInvalidConfig     = errors.New("checker: config is invalid")
)

type Interface interface {
	Check(*http.Request) (matches bool, err error)
	Hash() string
}

type List []Interface

func (l List) Check(r *http.Request) (bool, error) {
	for _, c := range l {
		ok, err := c.Check(r)
		if err != nil {
			return ok, err
		}
		if ok {
			return ok, nil
		}
	}

	return false, nil
}

func (l List) Hash() string {
	var sb strings.Builder

	for _, c := range l {
		fmt.Fprintln(&sb, c.Hash())
	}

	return internal.FastHash(sb.String())
}

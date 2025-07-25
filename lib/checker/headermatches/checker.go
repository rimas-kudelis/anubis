package headermatches

import (
	"net/http"
	"regexp"
)

type Checker struct {
	header string
	regexp *regexp.Regexp
	hash   string
}

func (c *Checker) Check(r *http.Request) (bool, error) {
	if c.regexp.MatchString(r.Header.Get(c.header)) {
		return true, nil
	}

	return false, nil
}

func (c *Checker) Hash() string {
	return c.hash
}

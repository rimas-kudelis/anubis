package policy

import (
	"fmt"
	"net/http"

	"github.com/TecharoHQ/anubis/internal"
	"github.com/TecharoHQ/anubis/internal/store/valkey"
)

type PassRateChecker struct {
	store  *valkey.Store
	header string
	rate   float64
}

func NewPassRateChecker(store *valkey.Store, rate float64) Checker {
	return &PassRateChecker{
		store:  store,
		rate:   rate,
		header: "User-Agent",
	}
}

func (prc *PassRateChecker) Hash() string {
	return internal.SHA256sum(fmt.Sprintf("pass rate checker::%s", prc.header))
}

func (prc *PassRateChecker) Check(r *http.Request) (bool, error) {
	data, err := prc.store.MultiGetInt(r.Context(), [][]string{
		{"pass_rate", prc.header, r.Header.Get(prc.header), "pass"},
		{"pass_rate", prc.header, r.Header.Get(prc.header), "challenges_issued"},
		{"pass_rate", prc.header, r.Header.Get(prc.header), "fail"},
	})
	if err != nil {
		return false, err
	}

	passCount, challengeCount, failCount := data[0], data[1], data[2]
	passRate := float64(passCount-failCount) / float64(challengeCount)

	if passRate >= prc.rate {
		return true, nil
	}

	return false, nil
}

package lib

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"

	"golang.org/x/sync/semaphore"
)

var (
	ErrChallengeFailed          = errors.New("libanubis: challenge failed, hash does not match what the server calculated")
	ErrWrongChallengeDifficulty = errors.New("libanubis: wrong challenge difficulty")
)

type Verifier interface {
	Verify(ctx context.Context, challenge, verify []byte, nonce, difficulty uint32) (bool, error)
}

type VerifierFunc func(ctx context.Context, challenge, verify []byte, nonce, difficulty uint32) (bool, error)

func (vf VerifierFunc) Verify(ctx context.Context, challenge, verify []byte, nonce, difficulty uint32) (bool, error) {
	return vf(ctx, challenge, verify, nonce, difficulty)
}

func BasicSHA256Verify(ctx context.Context, challenge, verify []byte, nonce, difficulty uint32) (bool, error) {
	h := sha256.New()
	fmt.Fprintf(h, "%x%d", challenge, nonce)
	data := h.Sum(nil)

	if subtle.ConstantTimeCompare(data, verify) != 1 {
		return false, fmt.Errorf("%w: wanted %x, got: %x", ErrChallengeFailed, verify, data)
	}

	if !hasLeadingZeroNibbles(data, difficulty) {
		return false, fmt.Errorf("%w: wanted %d leading zeroes in calculated data %x, but did not get it", ErrWrongChallengeDifficulty, difficulty, data)
	}

	if !hasLeadingZeroNibbles(verify, difficulty) {
		return false, fmt.Errorf("%w: wanted %d leading zeroes in verification data %x, but did not get it", ErrWrongChallengeDifficulty, verify, difficulty)
	}

	return true, nil
}

// hasLeadingZeroNibbles checks if the first `n` nibbles (in order) are zero.
// Nibbles are read from high to low for each byte (e.g., 0x12 -> nibbles [0x1, 0x2]).
func hasLeadingZeroNibbles(data []byte, n uint32) bool {
	count := uint32(0)
	for _, b := range data {
		// Check high nibble (first 4 bits)
		if (b >> 4) != 0 {
			break // Non-zero found in leading nibbles
		}
		count++
		if count >= n {
			return true
		}

		// Check low nibble (last 4 bits)
		if (b & 0x0F) != 0 {
			break // Non-zero found in leading nibbles
		}
		count++
		if count >= n {
			return true
		}
	}
	return count >= n
}

type ConcurrentVerifier struct {
	Verifier
	sem *semaphore.Weighted
}

func NewConcurrentVerifier(v Verifier, maxConcurrent int64) *ConcurrentVerifier {
	return &ConcurrentVerifier{
		Verifier: v,
		sem:      semaphore.NewWeighted(maxConcurrent),
	}
}

func (cv *ConcurrentVerifier) Verify(ctx context.Context, challenge, verify []byte, nonce, difficulty uint32) (bool, error) {
	if err := cv.sem.Acquire(ctx, 1); err != nil {
		return false, fmt.Errorf("can't verify solution: %w", err)
	}

	return cv.Verifier.Verify(ctx, challenge, verify, nonce, difficulty)
}

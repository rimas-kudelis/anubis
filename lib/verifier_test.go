package lib

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
)

// echo -n "hi2" | sha256sum
const hi2SHA256 = "0251f1ec2880f67631b8d0b3a62cf71a17dfa31858a323e7fc38068fcfaeded0"
const nonce uint32 = 5
const expectedVerifyString = "0543cbd94db5da055e82263cb775ac16f59fbbc1900645458baa197f9036ae9d"

func TestBasicSHA256Verify(t *testing.T) {
	ctx := context.Background()

	challenge, err := hex.DecodeString(hi2SHA256)
	if err != nil {
		t.Fatalf("[unexpected] %s does not decode as hex", hi2SHA256)
	}

	expectedVerify, err := hex.DecodeString(expectedVerifyString)
	if err != nil {
		t.Fatalf("[unexpected] %s does not decode as hex", expectedVerifyString)
	}

	t.Logf("got nonce: %d", nonce)
	t.Logf("got hash: %x", expectedVerify)

	invalidVerify := make([]byte, len(expectedVerify))
	copy(invalidVerify, expectedVerify)
	invalidVerify[len(invalidVerify)-1] ^= 0xFF // Flip the last byte

	testCases := []struct {
		name        string
		challenge   []byte
		verify      []byte
		nonce       uint32
		difficulty  uint32
		want        bool
		expectError error
	}{
		{
			name:        "valid verification",
			challenge:   challenge,
			verify:      expectedVerify,
			nonce:       nonce,
			difficulty:  1,
			want:        true,
			expectError: nil,
		},
		{
			name:        "invalid verify data",
			challenge:   challenge,
			verify:      invalidVerify,
			nonce:       nonce,
			difficulty:  1,
			want:        false,
			expectError: ErrChallengeFailed,
		},
		{
			name:        "insufficient computed data difficulty",
			challenge:   challenge,
			verify:      expectedVerify,
			nonce:       nonce,
			difficulty:  5,
			want:        false,
			expectError: ErrWrongChallengeDifficulty,
		},
		{
			name:        "zero difficulty",
			challenge:   challenge,
			verify:      expectedVerify,
			nonce:       nonce,
			difficulty:  0,
			want:        true,
			expectError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BasicSHA256Verify(ctx, tc.challenge, tc.verify, tc.nonce, tc.difficulty)
			if !errors.Is(err, tc.expectError) {
				t.Errorf("BasicSHA256Verify() error = %v, expectError %v", err, tc.expectError)
				return
			}
			if got != tc.want {
				t.Errorf("BasicSHA256Verify() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasLeadingZeroNibbles(t *testing.T) {
	for _, cs := range []struct {
		data       []byte
		difficulty uint32
		valid      bool
	}{
		{[]byte{0x10, 0x00}, 1, false},
		{[]byte{0x00, 0x00}, 4, true},
		{[]byte{0x01, 0x00}, 4, false},
	} {
		t.Run(fmt.Sprintf("%x-%d-%v", cs.data, cs.difficulty, cs.valid), func(t *testing.T) {
			result := hasLeadingZeroNibbles(cs.data, cs.difficulty)
			if result != cs.valid {
				t.Errorf("wanted %v, but got: %v", cs.valid, result)
			}
		})
	}
}

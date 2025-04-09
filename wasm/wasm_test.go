package wasm

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/TecharoHQ/anubis/web"
)

func TestSHA256(t *testing.T) {
	const difficulty = 4 // one nibble, intentionally easy for testing

	fin, err := web.Static.Open("static/wasm/sha256.wasm")
	if err != nil {
		t.Fatal(err)
	}
	defer fin.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)

	runner, err := NewRunner(ctx, "sha256.wasm", fin)
	if err != nil {
		t.Fatal(err)
	}

	h := sha256.New()
	fmt.Fprint(h, os.Args[0])
	data := h.Sum(nil)

	if n, err := runner.WriteData(ctx, data); err != nil {
		t.Fatalf("can't write data: %v", err)
	} else {
		t.Logf("wrote %d bytes to data segment", n)
	}

	t0 := time.Now()
	nonce, err := runner.anubisWork(ctx, difficulty, 0, 1)
	if err != nil {
		t.Fatalf("can't do test work run: %v", err)
	}
	t.Logf("got nonce %d in %s", nonce, time.Since(t0))

	hash, err := runner.ReadResult(ctx)
	if err != nil {
		t.Fatalf("can't read result: %v", err)
	}

	t.Logf("got hash %x", hash)

	if err := runner.WriteVerification(ctx, hash); err != nil {
		t.Fatalf("can't write verification: %v", err)
	}

	ok, err := runner.anubisValidate(ctx, nonce, difficulty)
	if err != nil {
		t.Fatalf("can't run validation: %v", err)
	}

	if !ok {
		t.Error("validation failed")
	}
}

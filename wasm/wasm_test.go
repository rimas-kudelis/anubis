package wasm

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/TecharoHQ/anubis/web"
)

func abiTest(t *testing.T, fname string, difficulty uint32) {
	fin, err := web.Static.Open("static/wasm/" + fname)
	if err != nil {
		t.Fatal(err)
	}
	defer fin.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)

	runner, err := NewRunner(ctx, fname, fin)
	if err != nil {
		t.Fatal(err)
	}

	h := sha256.New()
	fmt.Fprint(h, t.Name())
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

	t.Logf("used %d pages of wasm memory (%d bytes)", runner.module.Memory().Size()/63356, runner.module.Memory().Size())
}

func TestAlgos(t *testing.T) {
	fnames, err := fs.ReadDir(web.Static, "static/wasm")
	if err != nil {
		t.Fatal(err)
	}

	for _, fname := range fnames {
		fname := fname
		t.Run(fname.Name(), func(t *testing.T) {
			abiTest(t, fname.Name(), 4)
		})
	}
}

func bench(b *testing.B, fname string, difficulties []uint32) {
	fin, err := web.Static.Open("static/wasm/" + fname)
	if err != nil {
		b.Fatal(err)
	}
	defer fin.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	b.Cleanup(cancel)

	runner, err := NewRunner(ctx, fname, fin)
	if err != nil {
		b.Fatal(err)
	}

	h := sha256.New()
	fmt.Fprint(h, "This is an example value that exists only to test the system.")
	data := h.Sum(nil)

	if n, err := runner.WriteData(ctx, data); err != nil {
		b.Fatalf("can't write data: %v", err)
	} else {
		b.Logf("wrote %d bytes to data segment", n)
	}
}

func BenchmarkSHA256(b *testing.B) {
	fin, err := web.Static.Open("static/wasm/sha256.wasm")
	if err != nil {
		b.Fatal(err)
	}
	defer fin.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	b.Cleanup(cancel)

	runner, err := NewRunner(ctx, "sha256.wasm", fin)
	if err != nil {
		b.Fatal(err)
	}

	h := sha256.New()
	fmt.Fprint(h, "testificate")
	data := h.Sum(nil)

	if n, err := runner.WriteData(ctx, data); err != nil {
		b.Fatalf("can't write data: %v", err)
	} else {
		b.Logf("wrote %d bytes to data segment", n)
	}

	for _, cs := range []struct {
		Difficulty uint32
	}{
		{4},
		{6},
		{8},
		{10},
		{12},
		{14},
		{16},
	} {
		b.Run(fmt.Sprintf("difficulty/%d", cs.Difficulty), func(b *testing.B) {
			for b.Loop() {
				difficulty := cs.Difficulty
				_, err := runner.anubisWork(ctx, difficulty, 0, 1)
				if err != nil {
					b.Fatalf("can't do test work run: %v", err)
				}
			}
		})
	}
}

func BenchmarkArgon2ID(b *testing.B) {
	const difficulty = 4 // one nibble, intentionally easy for testing

	fin, err := web.Static.Open("static/wasm/argon2id.wasm")
	if err != nil {
		b.Fatal(err)
	}
	defer fin.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	b.Cleanup(cancel)

	runner, err := NewRunner(ctx, "argon2id.wasm", fin)
	if err != nil {
		b.Fatal(err)
	}

	h := sha256.New()
	fmt.Fprint(h, os.Args[0])
	data := h.Sum(nil)

	if n, err := runner.WriteData(ctx, data); err != nil {
		b.Fatalf("can't write data: %v", err)
	} else {
		b.Logf("wrote %d bytes to data segment", n)
	}

	for _, cs := range []struct {
		Difficulty uint32
	}{
		{4},
		{6},
		{8},
	} {
		b.Run(fmt.Sprintf("difficulty/%d", cs.Difficulty), func(b *testing.B) {
			for b.Loop() {
				difficulty := cs.Difficulty
				_, err := runner.anubisWork(ctx, difficulty, 0, 1)
				if err != nil {
					b.Fatalf("can't do test work run: %v", err)
				}
			}
		})
	}
}

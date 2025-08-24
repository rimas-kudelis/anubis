package ogtags

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/TecharoHQ/anubis/lib/policy/config"
	"github.com/TecharoHQ/anubis/lib/store/memory"
)

// TestUnixSocketTLSIssue tries to reproduce the actual "http: server gave HTTP response to HTTPS client" error
func TestUnixSocketTLSIssue(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")

	// Create a simple HTTP server listening on the Unix socket
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><head><meta property="og:title" content="Test Title"></head></html>`))
		}),
	}

	// Listen on Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create Unix socket listener: %v", err)
	}
	defer os.Remove(socketPath)
	defer listener.Close()

	// Start the server
	go func() {
		server.Serve(listener)
	}()
	defer server.Close()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Create OGTagCache with Unix socket target
	target := "unix://" + socketPath
	cache := NewOGTagCache(target, config.OpenGraph{
		Enabled: true,
		TimeToLive: time.Minute,
	}, memory.New(t.Context()))

	// Test with an HTTPS URL (this simulates the original request coming via HTTPS)
	httpsURL, _ := url.Parse("https://example.com/test")
	
	// Try to get OG tags - this should work without HTTPS/HTTP errors
	ogTags, err := cache.GetOGTags(context.Background(), httpsURL, "example.com")
	if err != nil {
		// If we get "http: server gave HTTP response to HTTPS client", this is the bug
		if err.Error() == "http: server gave HTTP response to HTTPS client" {
			t.Errorf("Found the bug: %v", err)
		} else {
			t.Logf("Got different error: %v", err)
		}
	} else {
		t.Logf("Success: got OG tags: %v", ogTags)
	}
}
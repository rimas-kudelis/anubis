package ogtags

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	
	"github.com/TecharoHQ/anubis/lib/policy/config"
	"github.com/TecharoHQ/anubis/lib/store/memory"
)

// TestUnixSocketHTTPSFix tests that the fix prevents "http: server gave HTTP response to HTTPS client" errors
// when using Unix socket targets with HTTPS input URLs
func TestUnixSocketHTTPSFix(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")

	// Create a simple HTTP server listening on the Unix socket
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify that the request comes in with HTTP (not HTTPS)
			if r.URL.Scheme != "" && r.URL.Scheme != "http" {
				t.Errorf("Unexpected scheme in request: %s (expected 'http' or empty)", r.URL.Scheme)
			}
			if r.TLS != nil {
				t.Errorf("Request has TLS information when it shouldn't for Unix socket")
			}
			
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

	// Test cases that previously might have caused the "HTTP response to HTTPS client" error
	testCases := []struct {
		name string
		url  string
	}{
		{"HTTPS URL", "https://example.com/test"},
		{"HTTPS with port", "https://example.com:443/test"},
		{"HTTPS with query", "https://example.com/test?param=value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputURL, _ := url.Parse(tc.url)
			
			// This should succeed without the "server gave HTTP response to HTTPS client" error
			ogTags, err := cache.GetOGTags(context.Background(), inputURL, "example.com")
			
			if err != nil {
				if strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") {
					t.Errorf("Fix did not work: still getting HTTPS/HTTP error: %v", err)
				} else {
					// Other errors are acceptable for this test
					t.Logf("Got non-HTTPS error (acceptable): %v", err)
				}
			} else {
				// Success case
				if ogTags["og:title"] != "Test Title" {
					t.Errorf("Expected og:title 'Test Title', got: %v", ogTags["og:title"])
				}
				t.Logf("Success: got expected og:title = %s", ogTags["og:title"])
			}
		})
	}
}
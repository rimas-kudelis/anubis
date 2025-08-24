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

// TestDebugUnixSocketRequests - let's debug exactly what URLs are being constructed
func TestDebugUnixSocketRequests(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")

	// Create a simple HTTP server listening on the Unix socket that logs requests
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("Received request: %s %s", r.Method, r.URL.String())
			t.Logf("Request scheme: %s", r.URL.Scheme)
			t.Logf("Request host: %s", r.Host)
			t.Logf("Is TLS: %v", r.TLS != nil)
			
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
	t.Logf("Using target: %s", target)
	
	cache := NewOGTagCache(target, config.OpenGraph{
		Enabled: true,
		TimeToLive: time.Minute,
	}, memory.New(t.Context()))

	// Test with various URL schemes
	testCases := []struct {
		name string
		inputURL string
	}{
		{"HTTPS URL", "https://example.com/test"},
		{"HTTP URL", "http://example.com/test"},
		{"HTTPS with port", "https://example.com:8080/test"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputURL, _ := url.Parse(tc.inputURL)
			t.Logf("Input URL: %s", inputURL.String())
			
			// Get the target URL that will be used
			targetURL := cache.getTarget(inputURL)
			t.Logf("Target URL: %s", targetURL)
			
			// Verify that the target URL uses http scheme
			if !strings.HasPrefix(targetURL, "http://unix") {
				t.Errorf("Expected target URL to start with 'http://unix', got: %s", targetURL)
			}
			
			// Try to get OG tags
			ogTags, err := cache.GetOGTags(context.Background(), inputURL, "example.com")
			if err != nil {
				if strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") {
					t.Errorf("BUG FOUND: %v", err)
				} else {
					t.Logf("Different error: %v", err)
				}
			} else {
				t.Logf("Success: got OG tags: %v", ogTags)
			}
		})
	}
}
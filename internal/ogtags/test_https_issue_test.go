package ogtags

import (
	"net/url"
	"testing"
	"time"
	
	"github.com/TecharoHQ/anubis/lib/policy/config"
	"github.com/TecharoHQ/anubis/lib/store/memory"
)

// TestUnixSocketHTTPSIssue reproduces the issue where unix socket targets
// might cause HTTPS/HTTP mismatch errors
func TestUnixSocketHTTPSIssue(t *testing.T) {
	target := "unix:///var/run/app.sock"
	
	cache := NewOGTagCache(target, config.OpenGraph{
		Enabled: true,
		TimeToLive: time.Minute,
	}, memory.New(t.Context()))
	
	// Test with HTTPS URL (this might be the source of confusion)
	httpsURL, _ := url.Parse("https://example.com/test?param=value")
	
	// Get the target URL that would be used for the request
	targetURL := cache.getTarget(httpsURL)
	
	t.Logf("Target URL for unix socket: %s", targetURL)
	
	// The target URL should be using http:// scheme, not https://
	expected := "http://unix/test?param=value"
	if targetURL != expected {
		t.Errorf("Expected %s, got %s", expected, targetURL)
	}
}
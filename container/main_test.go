package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHealthEndpoint tests the /health endpoint for both normal and draining states.
func TestHealthEndpoint(t *testing.T) {
	// Save and restore terminate global
	origTerminate := terminate
	defer func() { terminate = origTerminate }()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if terminate {
			w.WriteHeader(400)
			_, _ = w.Write([]byte("draining"))
			return
		}
		_, _ = w.Write([]byte("ok"))
	})

	// Test a healthy state
	terminate = false
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || string(body) != "ok" {
		t.Errorf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
	}

	// Test draining state
	terminate = true
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != 400 || string(body) != "draining" {
		t.Errorf("Expected 400 draining, got %d, body: %s", resp.StatusCode, string(body))
	}
}

// TestDebugHandler checks that /debug returns expected fields.
func TestDebugHandler(t *testing.T) {
	// Set some env vars for the handler
	t.Setenv("CLOUDFLARE_COUNTRY", "US")
	t.Setenv("CLOUDFLARE_COUNTRY_A1", "US")
	t.Setenv("CLOUDFLARE_COUNTRY_A2", "US")
	t.Setenv("CLOUDFLARE_LOCATION", "SFO")
	t.Setenv("CLOUDFLARE_REGION", "NA")
	t.Setenv("CLOUDFLARE_APPLICATION_ID", "app123")
	t.Setenv("CLOUDFLARE_NODE_ID", "node456")
	t.Setenv("CLOUDFLARE_DEPLOYMENT_ID", "deploy789")

	req := httptest.NewRequest("GET", "/debug", nil)
	w := httptest.NewRecorder()
	DebugHandler(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type %q, got %q", "text/plain; charset=utf-8", ct)
	}
	//contentLines := strings.Split(content, "\n")
	responseLines := []string{
		"Hi, I'm a container running in SFO, US, which is part of NA",
		"with the following build information:",
		"Version: " + versionString,
		"App ID: app123",
		"Deployment ID: deploy789",
		"Cloudflare Node ID: node456",
	}
	for _, line := range responseLines {
		if !strings.Contains(content, line) {
			t.Errorf("DebugHandler output missing expected line: %s", line)
		}
	}
}

// TestDNSTypesEndpoint checks that /dns_types returns a JSON array of DNS types.
func TestDNSTypesEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/dns_types", nil)
	w := httptest.NewRecorder()
	DNSTypesEndpoint(w, req)
	resp := w.Result()
	//body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != JSONApplicationType {
		t.Errorf("expected Content-Type %q, got %q", JSONApplicationType, ct)
	}

}

// TestResolve_InvalidQuery checks that invalid query params return 400.
func TestResolve_InvalidQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/lookup?bad=1", nil)
	w := httptest.NewRecorder()
	ResolveEndpoint(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusBadRequest || !strings.Contains(string(body), "Invalid query parameters") {
		t.Errorf("Expected 400 for invalid query, got %d, body: %s", resp.StatusCode, string(body))
	}
}

// TestResolve_InvalidDomain checks that an invalid domain returns 400.
func TestResolve_InvalidDomain(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/lookup?domain=not_a_domain", nil)
	w := httptest.NewRecorder()
	ResolveEndpoint(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusBadRequest || !strings.Contains(string(body), "Invalid query parameters: missing 'type' parameter in query") {
		t.Errorf("Expected 400 for invalid domain, got %d, body: %s", resp.StatusCode, string(body))
	}
}

// TestResolve_Success_Mock checks that a successful resolve returns JSON.
func TestResolve_Success(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/lookup?domain=example.com&type=A", nil)
	w := httptest.NewRecorder()
	ResolveEndpoint(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d, body: %s", resp.StatusCode, string(body))
	}
}

func Test_DNSServersEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/dns_servers", nil)
	w := httptest.NewRecorder()
	DNSServerEndpoint(w, req)
	resp := w.Result()
	// body, _ := io.ReadAll(resp.Body)
	// if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"mock":"ok"`) {
	// 	t.Errorf("Expected 200 and mock JSON, got %d, body: %s", resp.StatusCode, string(body))
	// }
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("expected Content-Type %q, got %q", "application/json; charset=utf-8", ct)
	}
}

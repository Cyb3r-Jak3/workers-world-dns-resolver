package main

import (
	"net/url"
	"testing"

	"encoding/json"
	"github.com/miekg/dns"
	"net/http"
	"net/http/httptest"
)

func TestParseURLQuery_Success(t *testing.T) {
	u, _ := url.Parse("http://localhost/query?domain=example.com&type=A")
	result, err := ParseURLQuery(u)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got '%s'", result.Domain)
	}
	if result.Type != dns.TypeA {
		t.Errorf("expected type %d, got %d", dns.TypeA, result.Type)
	}
}

func TestParseURLQuery_MissingDomain(t *testing.T) {
	u, _ := url.Parse("http://localhost/query?type=A")
	_, err := ParseURLQuery(u)
	if err == nil || err.Error() != "missing 'domain' parameter in query" {
		t.Errorf("expected missing domain error, got %v", err)
	}
}

func TestParseURLQuery_MissingType(t *testing.T) {
	u, _ := url.Parse("http://localhost/query?domain=example.com")
	_, err := ParseURLQuery(u)
	if err == nil || err.Error() != "missing 'type' parameter in query" {
		t.Errorf("expected missing type error, got %v", err)
	}
}

func TestParseURLQuery_InvalidType(t *testing.T) {
	u, _ := url.Parse("http://localhost/query?domain=example.com&type=INVALID")
	_, err := ParseURLQuery(u)
	if err == nil || err.Error() != "invalid DNS type: INVALID" {
		t.Errorf("expected invalid type error, got %v", err)
	}
}

func TestJSONResponse_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	body := map[string]string{"hello": "world"}

	JSONResponse(rr, body)

	resp := rr.Result()
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != JSONApplicationType {
		t.Errorf("expected Content-Type %q, got %q", JSONApplicationType, ct)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var decoded map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	if decoded["hello"] != "world" {
		t.Errorf("expected body to contain hello=world, got %v", decoded)
	}
}

//type BadJSON struct{}
//
//func (b BadJSON) MarshalJSON() ([]byte, error) {
//	return nil, fmt.Errorf("marshal error")
//}
//
//func TestJSONResponse_EncodeError(t *testing.T) {
//	rr := httptest.NewRecorder()
//	JSONResponse(rr, BadJSON{})
//
//	resp := rr.Result()
//	defer resp.Body.Close()
//	if resp.StatusCode != http.StatusInternalServerError {
//		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
//	}
//	// Should still set Content-Type
//	if ct := resp.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
//		t.Errorf("expected Content-Type %q, got %q", "text/plain; charset=utf-8", ct)
//	}
//	// Should return 500 Internal Server Error
//
//	buf := new(bytes.Buffer)
//	buf.ReadFrom(resp.Body)
//	if !strings.Contains(buf.String(), "Internal Server Error") {
//		t.Errorf("expected error message in response body, got %q", buf.String())
//	}
//}

package httpx

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestForwardedIPOnlyFromTrustedProxy(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	handler := ClientIPMiddleware(trusted, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(ClientIP(r.Context())))
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	req.Header.Set("X-Forwarded-For", "192.0.2.99, 198.51.100.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Body.String() != "203.0.113.1" {
		t.Fatalf("untrusted proxy spoofed IP: %s", rec.Body.String())
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "192.0.2.99, 198.51.100.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Body.String() != "198.51.100.1" {
		t.Fatalf("trusted proxy ignored: %s", rec.Body.String())
	}
}

package httpx

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestContextKeepsOnlySafeRequestIDs(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := RequestContext(log, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(RequestID(r.Context())))
	}))

	t.Run("keeps valid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("X-Request-ID", "mobile-req-123")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Body.String() != "mobile-req-123" || response.Header().Get("X-Request-ID") != "mobile-req-123" {
			t.Fatalf("valid request id was changed: body=%q header=%q", response.Body.String(), response.Header().Get("X-Request-ID"))
		}
	})

	t.Run("replaces control characters and oversized values", func(t *testing.T) {
		for _, input := range []string{"bad\nvalue", strings.Repeat("a", 129)} {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("X-Request-ID", input)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Body.String() == input || response.Header().Get("X-Request-ID") == input || len(response.Body.String()) != 36 {
				t.Fatalf("unsafe request id was not replaced: %q", response.Body.String())
			}
		}
	})
}

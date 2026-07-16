package httpx

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

func RequestContext(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if !validRequestID(id) {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		started := time.Now()
		next.ServeHTTP(w, r.WithContext(WithRequestID(r.Context(), id)))
		log.Info("http request", "request_id", id, "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(started).Milliseconds())
	})
}

func validRequestID(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}
	return true
}

func Recover(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.Error("panic recovered", "request_id", RequestID(r.Context()), "panic", v, "stack", string(debug.Stack()))
				Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

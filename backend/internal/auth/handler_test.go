package auth

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"alongtu/backend/internal/platform/httpx"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type failingSessionService struct{ err error }

func (s failingSessionService) Send(context.Context, string, string, string) (string, error) {
	return "", s.err
}
func (s failingSessionService) Verify(context.Context, string, string, string) (Tokens, error) {
	return Tokens{}, s.err
}
func (s failingSessionService) Refresh(context.Context, string) (Tokens, error) {
	return Tokens{}, s.err
}
func (s failingSessionService) Logout(context.Context, string) error { return s.err }
func (s failingSessionService) ParseAccess(context.Context, string, bool) (string, string, error) {
	return "", "", s.err
}

func TestSendRejectsInvalidClientInputWithoutClaimingSMSOutage(t *testing.T) {
	handler := NewHandler(NewService(nil, nil, DisabledSender{}, "handler-test-secret", time.Minute, time.Hour))
	req := httptest.NewRequest(http.MethodPost, "/auth/sms/send", bytes.NewBufferString(`{"phone":"invalid","deviceId":"device-123"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(httpx.WithClientIP(req.Context(), "203.0.113.1"))
	response := httptest.NewRecorder()

	handler.Send(response, req)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid input status=%d, want 400; body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_auth_request"`) {
		t.Fatalf("invalid input must have a stable client error code: %s", response.Body.String())
	}
}

func TestVerifyReportsChallengeStoreFailureAsServiceUnavailable(t *testing.T) {
	store := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 10 * time.Millisecond, MaxRetries: -1})
	t.Cleanup(func() { _ = store.Close() })
	handler := NewHandler(NewService(nil, store, DisabledSender{}, "handler-test-secret", time.Minute, time.Hour))
	body := `{"challengeId":"` + uuid.NewString() + `","code":"123456","deviceId":"device-123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sms/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.Verify(response, req)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("challenge-store failure status=%d, want 503; body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"code":"verification_unavailable"`) {
		t.Fatalf("challenge-store failure must have a stable unavailable code: %s", response.Body.String())
	}
}

func TestSessionStorageFailureIsNotReportedAsInvalidRefreshToken(t *testing.T) {
	for _, operation := range []struct {
		name   string
		handle func(*Handler, http.ResponseWriter, *http.Request)
		status int
		body   string
	}{
		{name: "refresh", handle: (*Handler).Refresh, status: http.StatusInternalServerError, body: `{"refreshToken":"session.token"}`},
		{name: "logout", handle: (*Handler).Logout, status: http.StatusInternalServerError, body: `{"refreshToken":"session.token"}`},
	} {
		t.Run(operation.name, func(t *testing.T) {
			handler := &Handler{s: failingSessionService{err: errors.New("database unavailable")}}
			req := httptest.NewRequest(http.MethodPost, "/auth/"+operation.name, strings.NewReader(operation.body))
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			operation.handle(handler, response, req)

			if response.Code != operation.status || !strings.Contains(response.Body.String(), `"code":"internal_error"`) {
				t.Fatalf("storage failure must be 500, status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestWriteAuthenticationStorageFailureIsNotReportedAsInvalidToken(t *testing.T) {
	handler := &Handler{s: failingSessionService{err: ErrSessionUnavailable}}
	protected := handler.Optional(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler must not run when session verification fails")
	}))
	req := httptest.NewRequest(http.MethodPost, "/places/id/confirmations", nil)
	req.Header.Set("Authorization", "Bearer any-access-token")
	response := httptest.NewRecorder()

	protected.ServeHTTP(response, req)

	if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("session-storage failure must be 500, status=%d body=%s", response.Code, response.Body.String())
	}
}

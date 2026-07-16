package places

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"alongtu/backend/internal/platform/cursor"
	"alongtu/backend/internal/platform/httpx"
)

func TestHandlerUsesStableErrorClasses(t *testing.T) {
	cases := []struct {
		name       string
		repo       fakeRepo
		provider   fakeProvider
		method     string
		target     string
		body       string
		withAuth   bool
		wantStatus int
		wantCode   string
	}{
		{
			name:       "repository failure is internal error",
			repo:       fakeRepo{err: errors.New("database unavailable")},
			method:     http.MethodGet,
			target:     "/nearby?latitude=39.9&longitude=116.4&coordinateSystem=wgs84",
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
		{
			name:       "provider failure is bad gateway",
			provider:   fakeProvider{err: errors.New("provider unavailable")},
			method:     http.MethodGet,
			target:     "/search?q=%E5%8E%95%E6%89%80",
			wantStatus: http.StatusBadGateway,
			wantCode:   "place_search_failed",
		},
		{
			name:       "unknown nearby category is rejected",
			method:     http.MethodGet,
			target:     "/nearby?latitude=39.9&longitude=116.4&coordinateSystem=wgs84&category=not_a_category",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_place_query",
		},
		{
			name:       "unknown search category is rejected",
			method:     http.MethodGet,
			target:     "/search?q=%E5%8E%95%E6%89%80&category=not_a_category",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_place_query",
		},
		{
			name:       "confirmation validation is unprocessable",
			method:     http.MethodPost,
			target:     "/11111111-1111-4111-8111-111111111111/confirmations",
			body:       `{"kind":"bad","result":"confirmed"}`,
			withAuth:   true,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "confirmation_rejected",
		},
		{
			name:       "missing place is not found",
			repo:       fakeRepo{err: httpx.ErrNotFound},
			method:     http.MethodPost,
			target:     "/11111111-1111-4111-8111-111111111111/confirmations",
			body:       `{"kind":"exists","result":"confirmed"}`,
			withAuth:   true,
			wantStatus: http.StatusNotFound,
			wantCode:   "place_not_found",
		},
		{
			name:       "feedback validation is unprocessable",
			method:     http.MethodPost,
			target:     "/11111111-1111-4111-8111-111111111111/feedback",
			body:       `{"kind":"wrong","details":"入口在北门"}`,
			withAuth:   true,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "feedback_rejected",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(NewService(test.repo, test.provider, cursor.New("handler-test-cursor-secret", time.Hour)))
			request := httptest.NewRequest(test.method, test.target, strings.NewReader(test.body))
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			if test.withAuth {
				ctx := httpx.WithUserID(request.Context(), "22222222-2222-4222-8222-222222222222")
				ctx = httpx.WithSessionID(ctx, "33333333-3333-4333-8333-333333333333")
				request = request.WithContext(ctx)
				request.Header.Set("Idempotency-Key", "handler-test-key")
			}
			response := httptest.NewRecorder()
			handler.Routes().ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d want=%d body=%s", response.Code, test.wantStatus, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"code":"`+test.wantCode+`"`) {
				t.Fatalf("response code mismatch: %s", response.Body.String())
			}
		})
	}
}

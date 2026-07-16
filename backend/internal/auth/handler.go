package auth

import (
	"alongtu/backend/internal/platform/httpx"
	"context"
	"errors"
	"net/http"
	"strings"
)

type handlerService interface {
	Send(context.Context, string, string, string) (string, error)
	Verify(context.Context, string, string, string) (Tokens, error)
	Refresh(context.Context, string) (Tokens, error)
	Logout(context.Context, string) error
	ParseAccess(context.Context, string, bool) (string, string, error)
}

type Handler struct{ s handlerService }

func NewHandler(s handlerService) *Handler { return &Handler{s: s} }
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Phone    string `json:"phone"`
		DeviceID string `json:"deviceId"`
	}
	if !httpx.Decode(w, r, &b) {
		return
	}
	id, err := h.s.Send(r.Context(), b.Phone, b.DeviceID, httpx.ClientIP(r.Context()))
	if errors.Is(err, ErrInvalidRequest) {
		httpx.Fail(w, r, 400, "invalid_auth_request", "手机号或设备信息无效", nil)
		return
	}
	if errors.Is(err, ErrRateLimited) {
		httpx.Fail(w, r, 429, "sms_rate_limited", "请求过于频繁，请稍后再试", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 503, "sms_unavailable", "验证码暂时无法发送", nil)
		return
	}
	httpx.JSON(w, 202, map[string]string{"challengeId": id, "expiresIn": "5m"})
}
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	var b struct {
		ChallengeID string `json:"challengeId"`
		Code        string `json:"code"`
		DeviceID    string `json:"deviceId"`
	}
	if !httpx.Decode(w, r, &b) {
		return
	}
	tokens, err := h.s.Verify(r.Context(), b.ChallengeID, b.Code, b.DeviceID)
	if errors.Is(err, ErrVerificationUnavailable) {
		httpx.Fail(w, r, 503, "verification_unavailable", "验证码服务暂时不可用，请稍后重试", nil)
		return
	}
	if !errors.Is(err, ErrInvalidChallenge) && err != nil {
		httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 401, "verification_rejected", "验证码无效或已过期", nil)
		return
	}
	httpx.JSON(w, 200, tokens)
}
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var b struct {
		RefreshToken string `json:"refreshToken"`
	}
	if !httpx.Decode(w, r, &b) {
		return
	}
	tokens, err := h.s.Refresh(r.Context(), b.RefreshToken)
	if errors.Is(err, ErrInvalidRefresh) {
		httpx.Fail(w, r, 401, "invalid_refresh_token", "登录状态已失效", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
		return
	}
	httpx.JSON(w, 200, tokens)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var b struct {
		RefreshToken string `json:"refreshToken"`
	}
	if !httpx.Decode(w, r, &b) {
		return
	}
	if err := h.s.Logout(r.Context(), b.RefreshToken); errors.Is(err, ErrInvalidRefresh) {
		httpx.Fail(w, r, 401, "invalid_refresh_token", "登录状态已失效", nil)
		return
	} else if err != nil {
		httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.Fail(w, r, 401, "invalid_token", "登录状态无效", nil)
			return
		}
		requireSession := r.Method != http.MethodGet && r.Method != http.MethodHead
		uid, sid, err := h.s.ParseAccess(r.Context(), parts[1], requireSession)
		if errors.Is(err, ErrSessionUnavailable) {
			httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
			return
		}
		if err != nil {
			httpx.Fail(w, r, 401, "invalid_token", "登录状态无效", nil)
			return
		}
		ctx := httpx.WithUserID(r.Context(), uid)
		ctx = httpx.WithSessionID(ctx, sid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

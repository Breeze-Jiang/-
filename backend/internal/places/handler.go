package places

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"alongtu/backend/internal/platform/cursor"
	"alongtu/backend/internal/platform/geo"
	"alongtu/backend/internal/platform/httpx"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ s *Service }

func NewHandler(s *Service) *Handler { return &Handler{s: s} }
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/nearby", h.nearby)
	r.Get("/search", h.search)
	r.Get("/{placeID}", h.get)
	r.Post("/{placeID}/confirmations", h.confirm)
	r.Post("/{placeID}/feedback", h.feedback)
	return r
}
func (h *Handler) nearby(w http.ResponseWriter, r *http.Request) {
	lat, e1 := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	lng, e2 := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	radius, e3 := optionalBoundedInt(r.URL.Query().Get("radiusMeters"), 100, 50000)
	limit, e4 := optionalBoundedInt(r.URL.Query().Get("limit"), 1, 50)
	if e1 != nil || e2 != nil {
		httpx.Fail(w, r, 400, "invalid_coordinate", "坐标无效", nil)
		return
	}
	if e3 != nil || e4 != nil {
		httpx.Fail(w, r, 400, "invalid_place_query", "地点查询参数无效", nil)
		return
	}
	page, err := h.s.Nearby(r.Context(), NearbyQuery{Center: geo.Coordinate{Latitude: lat, Longitude: lng, System: geo.System(strings.ToLower(r.URL.Query().Get("coordinateSystem")))}, Category: Category(r.URL.Query().Get("category")), RadiusMeters: radius, Limit: limit, Cursor: r.URL.Query().Get("cursor")})
	if errors.Is(err, cursor.ErrInvalid) {
		httpx.Fail(w, r, 400, "invalid_cursor", "分页游标无效或已过期", nil)
		return
	}
	if errors.Is(err, ErrInvalidQuery) {
		httpx.Fail(w, r, 400, "invalid_place_query", "地点查询参数无效", nil)
		return
	}
	if errors.Is(err, ErrRepository) {
		httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 502, "place_search_failed", "地点查询暂时不可用", nil)
		return
	}
	httpx.JSON(w, 200, page)
}
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	limit, parseErr := optionalBoundedInt(r.URL.Query().Get("limit"), 1, 50)
	if parseErr != nil {
		httpx.Fail(w, r, 400, "invalid_place_query", "地点查询参数无效", nil)
		return
	}
	page, err := h.s.Search(r.Context(), SearchQuery{Keyword: r.URL.Query().Get("q"), CityCode: r.URL.Query().Get("cityCode"), Category: Category(r.URL.Query().Get("category")), Limit: limit, Cursor: r.URL.Query().Get("cursor")})
	if errors.Is(err, cursor.ErrInvalid) {
		httpx.Fail(w, r, 400, "invalid_cursor", "分页游标无效或已过期", nil)
		return
	}
	if errors.Is(err, ErrInvalidQuery) {
		httpx.Fail(w, r, 400, "invalid_place_query", "地点查询参数无效", nil)
		return
	}
	if errors.Is(err, ErrRepository) {
		httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 502, "place_search_failed", "地点查询暂时不可用", nil)
		return
	}
	httpx.JSON(w, 200, page)
}

func optionalBoundedInt(value string, min, max int) (int, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < min || parsed > max {
		return 0, errors.New("out of range")
	}
	return parsed, nil
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.s.Get(r.Context(), chi.URLParam(r, "placeID"))
	if errors.Is(err, httpx.ErrNotFound) {
		httpx.Fail(w, r, 404, "place_not_found", "地点不存在或已不可用", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
		return
	}
	httpx.JSON(w, 200, p)
}
func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpx.UserID(r.Context())
	if !ok {
		httpx.Fail(w, r, 401, "authentication_required", "请先登录", nil)
		return
	}
	sid, ok := httpx.SessionID(r.Context())
	if !ok {
		httpx.Fail(w, r, 401, "authentication_required", "登录状态已失效", nil)
		return
	}
	var body struct {
		Kind   string `json:"kind"`
		Result string `json:"result"`
	}
	if !httpx.Decode(w, r, &body) {
		return
	}
	c, err := h.s.Confirm(r.Context(), Confirmation{PlaceID: chi.URLParam(r, "placeID"), UserID: uid, SessionID: sid, Kind: body.Kind, Result: body.Result, IdempotencyKey: r.Header.Get("Idempotency-Key")})
	if errors.Is(err, ErrWriteNotAllowed) {
		httpx.Fail(w, r, 401, "invalid_session", "登录状态已失效", nil)
		return
	}
	if errors.Is(err, httpx.ErrNotFound) {
		httpx.Fail(w, r, 404, "place_not_found", "地点不存在或已不可用", nil)
		return
	}
	if errors.Is(err, ErrIdempotencyConflict) {
		httpx.Fail(w, r, 409, "idempotency_conflict", "幂等键已用于不同请求", nil)
		return
	}
	if errors.Is(err, ErrInvalidConfirmation) {
		httpx.Fail(w, r, 422, "confirmation_rejected", "确认信息无效", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
		return
	}
	httpx.JSON(w, 201, c)
}
func (h *Handler) feedback(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpx.UserID(r.Context())
	if !ok {
		httpx.Fail(w, r, 401, "authentication_required", "请先登录", nil)
		return
	}
	sid, ok := httpx.SessionID(r.Context())
	if !ok {
		httpx.Fail(w, r, 401, "authentication_required", "登录状态已失效", nil)
		return
	}
	var body struct {
		Kind    string `json:"kind"`
		Details string `json:"details"`
	}
	if !httpx.Decode(w, r, &body) {
		return
	}
	feedback, err := h.s.SubmitFeedback(r.Context(), Feedback{PlaceID: chi.URLParam(r, "placeID"), UserID: uid, SessionID: sid, Kind: body.Kind, Details: body.Details, IdempotencyKey: r.Header.Get("Idempotency-Key")})
	if errors.Is(err, ErrWriteNotAllowed) {
		httpx.Fail(w, r, 401, "invalid_session", "登录状态已失效", nil)
		return
	}
	if errors.Is(err, httpx.ErrNotFound) {
		httpx.Fail(w, r, 404, "place_not_found", "地点不存在或已不可用", nil)
		return
	}
	if errors.Is(err, ErrIdempotencyConflict) {
		httpx.Fail(w, r, 409, "idempotency_conflict", "幂等键已用于不同请求", nil)
		return
	}
	if errors.Is(err, ErrInvalidConfirmation) {
		httpx.Fail(w, r, 422, "feedback_rejected", "事实反馈无效", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 500, "internal_error", "服务暂时不可用", nil)
		return
	}
	httpx.JSON(w, 201, feedback)
}

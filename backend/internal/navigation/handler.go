package navigation

import (
	"alongtu/backend/internal/platform/geo"
	"alongtu/backend/internal/platform/httpx"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{ s *Service }

func NewHandler(s *Service) *Handler { return &Handler{s: s} }
func (h *Handler) Route(w http.ResponseWriter, r *http.Request) {
	var body RouteRequest
	if !httpx.Decode(w, r, &body) {
		return
	}
	route, err := h.s.Route(r.Context(), body)
	if errors.Is(err, ErrInvalidRequest) {
		httpx.Fail(w, r, 400, "invalid_navigation_request", "导航参数无效", nil)
		return
	}
	if err != nil {
		httpx.Fail(w, r, 502, "route_unavailable", "路线暂时不可用", map[string]any{"externalFallback": route.ExternalFallback})
		return
	}
	httpx.JSON(w, 200, route)
}
func (h *Handler) ExternalLinks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lat, e1 := strconv.ParseFloat(q.Get("destinationLatitude"), 64)
	lng, e2 := strconv.ParseFloat(q.Get("destinationLongitude"), 64)
	oLat, e3 := strconv.ParseFloat(q.Get("originLatitude"), 64)
	oLng, e4 := strconv.ParseFloat(q.Get("originLongitude"), 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		httpx.Fail(w, r, 400, "invalid_navigation_request", "导航参数无效", nil)
		return
	}
	name := strings.TrimSpace(q.Get("destinationName"))
	if len([]rune(name)) == 0 || len([]rune(name)) > 200 {
		httpx.Fail(w, r, 400, "invalid_navigation_request", "导航参数无效", nil)
		return
	}
	system := geo.System(q.Get("coordinateSystem"))
	links, err := ExternalLinks(geo.Coordinate{Latitude: oLat, Longitude: oLng, System: system}, geo.Coordinate{Latitude: lat, Longitude: lng, System: system}, name, Mode(q.Get("mode")), h.s.rideTemplate)
	if err != nil {
		httpx.Fail(w, r, 400, "invalid_navigation_request", "导航参数无效", nil)
		return
	}
	httpx.JSON(w, 200, map[string]any{"links": links})
}

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"alongtu/backend/internal/auth"
	"alongtu/backend/internal/navigation"
	"alongtu/backend/internal/places"
	"alongtu/backend/internal/platform/cursor"
	"alongtu/backend/internal/platform/geo"
)

const visitorPlaceID = "11111111-1111-4111-8111-111111111111"

type visitorRepository struct{}

type apiEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"requestId"`
	} `json:"error"`
}

type apiPlace struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	DistanceMeters *float64 `json:"distanceMeters"`
	Location       struct {
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
		CoordinateSystem string  `json:"coordinateSystem"`
	} `json:"location"`
	Trust struct {
		Exists struct {
			Status string `json:"status"`
			Source string `json:"source"`
		} `json:"exists"`
	} `json:"trust"`
}

type apiPlacePage struct {
	Items         []apiPlace `json:"items"`
	Degraded      bool       `json:"degraded"`
	DataFreshness string     `json:"dataFreshness"`
	Coverage      string     `json:"coverage"`
}

type apiRoute struct {
	DistanceMeters  int `json:"distanceMeters"`
	DurationSeconds int `json:"durationSeconds"`
	Steps           []struct {
		Instruction string `json:"instruction"`
	} `json:"steps"`
	ExternalFallback []struct {
		Provider  string `json:"provider"`
		Available bool   `json:"available"`
	} `json:"externalFallback"`
	DataFreshness string `json:"dataFreshness"`
}

type apiExternalLinks struct {
	Links []struct {
		Provider  string `json:"provider"`
		URI       string `json:"uri"`
		Available bool   `json:"available"`
	} `json:"links"`
}

func (visitorRepository) Nearby(context.Context, places.NearbyQuery) (places.Page, error) {
	return places.Page{Items: []places.Place{}, DataFreshness: "stored"}, nil
}
func (visitorRepository) Get(context.Context, string) (places.Place, error) {
	place := visitorPlace()
	place.ID = visitorPlaceID
	return place, nil
}
func (visitorRepository) Materialize(_ context.Context, input []places.Place) ([]places.Place, error) {
	for i := range input {
		input[i].ID = visitorPlaceID
	}
	return input, nil
}
func (visitorRepository) AddConfirmation(context.Context, places.Confirmation) (places.Confirmation, error) {
	panic("visitor flow must not write confirmations")
}
func (visitorRepository) AddFeedback(context.Context, places.Feedback) (places.Feedback, error) {
	panic("visitor flow must not write place feedback")
}

type visitorMapProvider struct{}

func (visitorMapProvider) Nearby(context.Context, places.NearbyQuery) (places.Page, error) {
	return places.Page{Items: []places.Place{visitorPlace()}, DataFreshness: "live"}, nil
}
func (visitorMapProvider) Search(context.Context, places.SearchQuery) (places.Page, error) {
	return places.Page{Items: []places.Place{visitorPlace()}, DataFreshness: "live"}, nil
}
func (visitorMapProvider) Route(context.Context, navigation.RouteRequest) (navigation.Route, error) {
	now := time.Now().UTC()
	return navigation.Route{
		DistanceMeters:    650,
		DurationSeconds:   480,
		Polyline:          []geo.Coordinate{{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}},
		Steps:             []navigation.Step{{Instruction: "向前步行", DistanceMeters: 650}},
		ProviderUpdatedAt: &now,
	}, nil
}

func TestVisitorCanFindTrustedPlaceAndStartNavigation(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	authHandler := auth.NewHandler(auth.NewService(nil, nil, auth.DisabledSender{}, "visitor-test-secret", time.Minute, time.Hour))
	mapProvider := visitorMapProvider{}
	placeHandler := places.NewHandler(places.NewService(visitorRepository{}, mapProvider, cursor.New("visitor-cursor-secret", 15*time.Minute)))
	navigationHandler := navigation.NewHandler(navigation.NewService(mapProvider, ""))
	handler := newHTTPHandler(log, nil, []string{"https://client.example"}, authHandler, placeHandler, navigationHandler, nil, func(context.Context) error { return nil })

	t.Run("nearby toilet does not require login", func(t *testing.T) {
		body := performRequest(t, handler, http.MethodGet, "/api/v1/places/nearby?latitude=39.9&longitude=116.4&coordinateSystem=gcj02&category=toilet&radiusMeters=5000&limit=20", nil, http.StatusOK)
		var page apiPlacePage
		decodeData(t, body, &page)
		if len(page.Items) != 1 || page.Items[0].ID != visitorPlaceID || page.Items[0].Category != "toilet" || page.Items[0].DistanceMeters == nil || *page.Items[0].DistanceMeters <= 0 {
			t.Fatalf("unexpected place page: %+v", page)
		}
		if page.DataFreshness != "live" || page.Coverage != "covered" || page.Degraded || page.Items[0].Location.CoordinateSystem != "gcj02" || page.Items[0].Trust.Exists.Status != "unknown" {
			t.Fatalf("unexpected frontend place semantics: %+v", page)
		}
	})

	t.Run("search and detail preserve the platform place id", func(t *testing.T) {
		search := performRequest(t, handler, http.MethodGet, "/api/v1/places/search?q=%E5%8E%95%E6%89%80&limit=20", nil, http.StatusOK)
		var page apiPlacePage
		decodeData(t, search, &page)
		if len(page.Items) != 1 || page.Items[0].ID != visitorPlaceID {
			t.Fatalf("search did not return platform id: %+v", page)
		}
		detail := performRequest(t, handler, http.MethodGet, "/api/v1/places/"+visitorPlaceID, nil, http.StatusOK)
		var place apiPlace
		decodeData(t, detail, &place)
		if place.ID != visitorPlaceID || place.Name != "沿途测试厕所" || place.Trust.Exists.Status != "unknown" || place.Trust.Exists.Source != "provider" {
			t.Fatalf("unexpected detail for frontend: %+v", place)
		}
	})

	t.Run("walking route and external navigation remain available", func(t *testing.T) {
		request := `{"origin":{"latitude":39.9,"longitude":116.4,"coordinateSystem":"gcj02"},"destination":{"latitude":39.91,"longitude":116.41,"coordinateSystem":"gcj02"},"mode":"walking"}`
		route := performRequest(t, handler, http.MethodPost, "/api/v1/navigation/routes", strings.NewReader(request), http.StatusOK)
		var routeData apiRoute
		decodeData(t, route, &routeData)
		if routeData.DistanceMeters != 650 || routeData.DurationSeconds != 480 || routeData.DataFreshness != "live" || len(routeData.Steps) != 1 || len(routeData.ExternalFallback) != 3 {
			t.Fatalf("unexpected route for frontend: %+v", routeData)
		}
		links := performRequest(t, handler, http.MethodGet, "/api/v1/navigation/external-links?originLatitude=39.9&originLongitude=116.4&destinationLatitude=39.91&destinationLongitude=116.41&destinationName=%E6%B2%BF%E9%80%94%E6%B5%8B%E8%AF%95%E5%8E%95%E6%89%80&coordinateSystem=gcj02&mode=walking", nil, http.StatusOK)
		var external apiExternalLinks
		decodeData(t, links, &external)
		if len(external.Links) != 3 || external.Links[0].Provider != "amap" || !external.Links[0].Available || external.Links[0].URI == "" {
			t.Fatalf("unexpected external links for frontend: %+v", external)
		}
	})

	t.Run("invalid place parameters are not reported as provider failures", func(t *testing.T) {
		invalidRadius := performRequest(t, handler, http.MethodGet, "/api/v1/places/nearby?latitude=39.9&longitude=116.4&coordinateSystem=gcj02&radiusMeters=invalid", nil, http.StatusBadRequest)
		assertBodyContains(t, invalidRadius, `"code":"invalid_place_query"`)
		nonFiniteCoordinate := performRequest(t, handler, http.MethodGet, "/api/v1/places/nearby?latitude=NaN&longitude=116.4&coordinateSystem=gcj02", nil, http.StatusBadRequest)
		assertBodyContains(t, nonFiniteCoordinate, `"code":"invalid_place_query"`)
		invalidSystem := performRequest(t, handler, http.MethodGet, "/api/v1/places/nearby?latitude=39.9&longitude=116.4&coordinateSystem=unknown", nil, http.StatusBadRequest)
		assertBodyContains(t, invalidSystem, `"code":"invalid_place_query"`)
		invalidSearch := performRequest(t, handler, http.MethodGet, "/api/v1/places/search?q=&limit=20", nil, http.StatusBadRequest)
		assertBodyContains(t, invalidSearch, `"code":"invalid_place_query"`)
	})
}

func visitorPlace() places.Place {
	return places.Place{
		ID:             "amap:test-toilet",
		Name:           "沿途测试厕所",
		Category:       places.CategoryToilet,
		Address:        "测试路 1 号",
		Location:       geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02},
		DistanceMeters: float64Pointer(650),
		Trust: places.TrustSummary{
			Exists:       places.TrustItem{Status: "unknown", Source: "provider"},
			OpeningHours: places.TrustItem{Status: "unknown", Source: "provider"},
		},
	}
}

func float64Pointer(value float64) *float64 { return &value }

func performRequest(t *testing.T, handler http.Handler, method, target string, body io.Reader, wantStatus int) string {
	t.Helper()
	request := httptest.NewRequest(method, target, body)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, target, response.Code, wantStatus, response.Body.String())
	}
	return response.Body.String()
}

func assertBodyContains(t *testing.T, body string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !bytes.Contains([]byte(body), []byte(fragment)) {
			t.Errorf("response does not contain %s: %s", fragment, body)
		}
	}
}

func decodeData(t *testing.T, body string, target any) {
	t.Helper()
	var envelope apiEnvelope
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	if envelope.Error != nil || len(envelope.Data) == 0 {
		t.Fatalf("expected data envelope, got %+v", envelope.Error)
	}
	if err := json.Unmarshal(envelope.Data, target); err != nil {
		t.Fatalf("decode response data: %v; data=%s", err, envelope.Data)
	}
}

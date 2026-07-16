package amap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"alongtu/backend/internal/navigation"
	"alongtu/backend/internal/places"
	"alongtu/backend/internal/platform/geo"
)

func TestNearbySkipsDirtyPOIsAndKeepsValidResults(t *testing.T) {
	server := newAmapTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"1","pois":[{"id":"bad","name":"bad","typecode":"200300","location":"not-a-coordinate"},{"id":"","name":"missing id","typecode":"200300","location":"116.397470,39.908823"},{"id":"nan","name":"not finite","typecode":"200300","location":"NaN,39.908823"},{"id":"far","name":"out of radius","typecode":"200300","location":"116.397470,40.908823","distance":"1"},{"id":"toilet","name":"公共厕所","typecode":"200300","location":"116.397470,39.908823","distance":"320"}]}`))
	})
	defer server.Close()

	page, err := New(server.URL, "test-key", time.Second).Nearby(context.Background(), places.NearbyQuery{Center: geo.Coordinate{Latitude: 39.908823, Longitude: 116.397470, System: geo.GCJ02}, RadiusMeters: 1000, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "amap:toilet" || page.Items[0].Category != places.CategoryToilet || page.Items[0].DistanceMeters == nil || *page.Items[0].DistanceMeters > 1 || page.Items[0].SourceUpdatedAt == nil {
		t.Fatalf("unexpected sanitized POIs: %+v", page.Items)
	}
}

func TestUnverifiedCategoryReturnsNotCoveredWithoutProviderCall(t *testing.T) {
	called := false
	server := newAmapTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()
	page, err := New(server.URL, "test-key", time.Second).Nearby(context.Background(), places.NearbyQuery{Center: geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}, Category: places.CategoryParking, RadiusMeters: 1000, Limit: 20})
	if err != nil || called || page.Coverage != "not_covered" || len(page.Items) != 0 {
		t.Fatalf("unverified category must not call provider: page=%+v called=%v err=%v", page, called, err)
	}
}

func TestProviderHTTPErrorRetryClassification(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		temporary bool
	}{
		{name: "rate limited", status: http.StatusTooManyRequests, temporary: true},
		{name: "server failure", status: http.StatusBadGateway, temporary: true},
		{name: "client failure", status: http.StatusBadRequest, temporary: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			server := newAmapTestServer(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(test.status) })
			defer server.Close()
			_, err := New(server.URL, "test-key", time.Second).Search(context.Background(), places.SearchQuery{Keyword: "厕所", Limit: 20})
			if err == nil || isTemporary(err) != test.temporary {
				t.Fatalf("status=%d err=%v temporary=%v want=%v", test.status, err, isTemporary(err), test.temporary)
			}
		})
	}
}

func TestProviderTimeoutIsRetryable(t *testing.T) {
	server := newAmapTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`{"status":"1","pois":[]}`))
	})
	defer server.Close()
	_, err := New(server.URL, "test-key", 5*time.Millisecond).Search(context.Background(), places.SearchQuery{Keyword: "厕所", Limit: 20})
	if err == nil || !isTemporary(err) {
		t.Fatalf("timeout err=%v, want retryable provider error", err)
	}
}

func TestProviderRejectsAndRouteNoResultAreNotRetryable(t *testing.T) {
	t.Run("provider rejected request", func(t *testing.T) {
		server := newAmapTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"status":"0","info":"INVALID_USER_KEY"}`))
		})
		defer server.Close()
		_, err := New(server.URL, "test-key", time.Second).Search(context.Background(), places.SearchQuery{Keyword: "厕所", Limit: 20})
		if err == nil || isTemporary(err) {
			t.Fatalf("provider rejection err=%v should not be retryable", err)
		}
	})

	t.Run("route has no path", func(t *testing.T) {
		server := newAmapTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"status":"1","route":{"paths":[]},"data":{"paths":[]}}`))
		})
		defer server.Close()
		coordinate := geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}
		_, err := New(server.URL, "test-key", time.Second).Route(context.Background(), navigation.RouteRequest{Origin: coordinate, Destination: coordinate, Mode: navigation.Walking})
		if err == nil || isTemporary(err) || err.Error() != "route not found" {
			t.Fatalf("route result err=%v should be a non-retryable no-result error", err)
		}
	})

	t.Run("route has malformed distance or polyline", func(t *testing.T) {
		server := newAmapTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"status":"1","route":{"paths":[{"distance":"not-a-number","duration":"60","polyline":"invalid"}]}}`))
		})
		defer server.Close()
		coordinate := geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}
		_, err := New(server.URL, "test-key", time.Second).Route(context.Background(), navigation.RouteRequest{Origin: coordinate, Destination: coordinate, Mode: navigation.Walking})
		if err == nil || isTemporary(err) {
			t.Fatalf("malformed route err=%v should be rejected without retry", err)
		}
	})
}

func isTemporary(err error) bool {
	type temporaryError interface{ Temporary() bool }
	var target temporaryError
	return errors.As(err, &target) && target.Temporary()
}

func newAmapTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

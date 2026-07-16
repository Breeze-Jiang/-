package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestHTTPMetricsUseRouteTemplateWithoutResourceID(t *testing.T) {
	metrics := NewMetrics()
	router := chi.NewRouter()
	router.Use(metrics.HTTPMiddleware)
	router.Get("/places/{placeID}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	request := httptest.NewRequest(http.MethodGet, "/places/11111111-1111-4111-8111-111111111111", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if got := counterValue(t, metrics.HTTPRequests.WithLabelValues(http.MethodGet, "/places/{placeID}", "202")); got != 1 {
		t.Fatalf("templated request metric = %v, want 1", got)
	}
	if got := counterValue(t, metrics.HTTPRequests.WithLabelValues(http.MethodGet, "/places/11111111-1111-4111-8111-111111111111", "202")); got != 0 {
		t.Fatalf("resource ID leaked into metric label: %v", got)
	}
}

func TestMetricsRegistriesAreIndependent(t *testing.T) {
	first := NewMetrics()
	second := NewMetrics()
	first.ObserveProviderCache("nearby", "miss")
	if got := counterValue(t, first.ProviderCache.WithLabelValues("nearby", "miss")); got != 1 {
		t.Fatalf("first registry cache metric = %v", got)
	}
	if got := counterValue(t, second.ProviderCache.WithLabelValues("nearby", "miss")); got != 0 {
		t.Fatalf("second registry inherited first registry state: %v", got)
	}
}

func counterValue(t *testing.T, counter prometheus.Counter) float64 {
	t.Helper()
	metric := &dto.Metric{}
	if err := counter.Write(metric); err != nil {
		t.Fatalf("read counter: %v", err)
	}
	return metric.GetCounter().GetValue()
}

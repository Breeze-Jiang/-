package resilient

import (
	"alongtu/backend/internal/navigation"
	"alongtu/backend/internal/places"
	"context"
	"errors"
	"testing"
	"time"
)

type failingProvider struct{}

func (failingProvider) Nearby(context.Context, places.NearbyQuery) (places.Page, error) {
	return places.Page{}, temporaryFailure{}
}
func (failingProvider) Search(context.Context, places.SearchQuery) (places.Page, error) {
	return places.Page{}, temporaryFailure{}
}
func (failingProvider) Route(context.Context, navigation.RouteRequest) (navigation.Route, error) {
	return navigation.Route{}, temporaryFailure{}
}

type temporaryFailure struct{}

func (temporaryFailure) Error() string   { return "temporary" }
func (temporaryFailure) Temporary() bool { return true }

type recordedMetrics struct {
	providerOutcomes []string
	cacheStates      []string
	circuitStates    []string
}

func (m *recordedMetrics) ObserveProvider(operation, outcome string, _ time.Duration) {
	m.providerOutcomes = append(m.providerOutcomes, operation+":"+outcome)
}
func (m *recordedMetrics) ObserveProviderCache(operation, state string) {
	m.cacheStates = append(m.cacheStates, operation+":"+state)
}
func (m *recordedMetrics) SetProviderCircuit(state string) {
	m.circuitStates = append(m.circuitStates, state)
}

func TestCircuitOpensAfterFiveFailures(t *testing.T) {
	metrics := &recordedMetrics{}
	p := New(failingProvider{}, nil, metrics)
	for i := 0; i < 5; i++ {
		_ = p.call(context.Background(), "test", func(context.Context) error { return temporaryFailure{} })
	}
	if err := p.call(context.Background(), "test", func(context.Context) error { return errors.New("should not execute") }); !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected open circuit, got %v", err)
	}
	if !contains(metrics.providerOutcomes, "test:circuit_open") {
		t.Fatalf("missing circuit-open provider metric: %v", metrics.providerOutcomes)
	}
	if !contains(metrics.circuitStates, "open") {
		t.Fatalf("missing open circuit state metric: %v", metrics.circuitStates)
	}
}

func TestProviderRecordsCacheMissAndOperationFailure(t *testing.T) {
	metrics := &recordedMetrics{}
	p := New(failingProvider{}, nil, metrics)
	_, err := p.Nearby(context.Background(), places.NearbyQuery{})
	if err == nil {
		t.Fatal("expected provider failure")
	}
	if !contains(metrics.cacheStates, "nearby:miss") || !contains(metrics.providerOutcomes, "nearby:error") {
		t.Fatalf("missing provider observations: cache=%v provider=%v", metrics.cacheStates, metrics.providerOutcomes)
	}
}

func TestRetryCannotExceedTotalProviderDeadline(t *testing.T) {
	p := NewWithTimeout(failingProvider{}, nil, 10*time.Millisecond)
	started := time.Now()
	err := p.call(context.Background(), "test", func(context.Context) error { return temporaryFailure{} })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected total deadline error, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("retry exceeded total timeout budget: %v", elapsed)
	}
}

func TestDeadlineFailuresOpenCircuit(t *testing.T) {
	p := New(failingProvider{}, nil)
	for i := 0; i < 5; i++ {
		err := p.call(context.Background(), "test", func(context.Context) error { return context.DeadlineExceeded })
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline failure %d = %v", i, err)
		}
	}
	if err := p.call(context.Background(), "test", func(context.Context) error { return errors.New("should not execute") }); !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("deadline failures must open the circuit, got %v", err)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

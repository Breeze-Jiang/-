package navigation

import (
	"alongtu/backend/internal/platform/geo"
	"context"
	"errors"
	"strings"
	"testing"
)

type failingRouteProvider struct{}

func (failingRouteProvider) Route(context.Context, RouteRequest) (Route, error) {
	return Route{}, errors.New("provider unavailable")
}

type cachedRouteProvider struct{}

func (cachedRouteProvider) Route(context.Context, RouteRequest) (Route, error) {
	return Route{DataFreshness: "cached"}, nil
}

type countingRouteProvider struct{ calls int }

func (p *countingRouteProvider) Route(context.Context, RouteRequest) (Route, error) {
	p.calls++
	return Route{}, errors.New("route provider must not be used for taxi")
}

func TestExternalLinksRejectUnsupportedMode(t *testing.T) {
	c := geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}
	if _, err := ExternalLinks(c, c, "test", "flight", ""); err == nil {
		t.Fatal("expected error")
	}
}
func TestExternalLinksEncodeName(t *testing.T) {
	c := geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}
	links, err := ExternalLinks(c, c, "destination", Walking, "")
	if err != nil || len(links) != 3 {
		t.Fatalf("unexpected result: %#v %v", links, err)
	}
}

func TestAmapExternalLinkDeclaresGCJ02AfterWGS84Conversion(t *testing.T) {
	origin := geo.Coordinate{Latitude: 39.908823, Longitude: 116.397470, System: geo.WGS84}
	destination := geo.Coordinate{Latitude: 39.918823, Longitude: 116.407470, System: geo.WGS84}
	links, err := ExternalLinks(origin, destination, "destination", Walking, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) == 0 || !strings.Contains(links[0].URI, "dev=1") {
		t.Fatalf("AMap link must declare GCJ-02 coordinates: %#v", links)
	}
}
func TestTaxiUnavailableWithoutConfiguredProvider(t *testing.T) {
	c := geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}
	links, err := ExternalLinks(c, c, "destination", Taxi, "")
	if err != nil || len(links) != 1 || links[0].Available {
		t.Fatalf("unexpected taxi result: %#v %v", links, err)
	}
}

func TestTaxiReturnsOnlyConfiguredExternalEntryWithoutCallingRouteProvider(t *testing.T) {
	c := geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}
	provider := &countingRouteProvider{}
	route, err := NewService(provider, "rideapp://request?from={origin}&to={destination}&name={name}").Route(context.Background(), RouteRequest{Origin: c, Destination: c, Mode: Taxi})
	if err != nil {
		t.Fatalf("taxi entry should not depend on route calculation: %v", err)
	}
	if provider.calls != 0 {
		t.Fatalf("taxi must not call the route provider, calls=%d", provider.calls)
	}
	if route.DataFreshness != "external_only" || route.DistanceMeters != 0 || route.DurationSeconds != 0 || len(route.Polyline) != 0 || len(route.Steps) != 0 {
		t.Fatalf("taxi must not claim an in-app route: %#v", route)
	}
	if len(route.ExternalFallback) != 1 || route.ExternalFallback[0].Provider != "ride_hailing" || !route.ExternalFallback[0].Available {
		t.Fatalf("unexpected taxi external entry: %#v", route.ExternalFallback)
	}
}

func TestRouteFailureReturnsStableExternalOnlyShape(t *testing.T) {
	c := geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}
	route, err := NewService(failingRouteProvider{}, "").Route(context.Background(), RouteRequest{Origin: c, Destination: c, Mode: Walking})
	if err == nil {
		t.Fatal("expected provider error")
	}
	if route.ProviderUpdatedAt != nil {
		t.Fatalf("external-only route must not claim a provider update time: %v", route.ProviderUpdatedAt)
	}
	if route.Polyline == nil || route.Steps == nil || route.ExternalFallback == nil {
		t.Fatalf("route collections must be empty arrays, got %#v", route)
	}
	if route.DataFreshness != "external_only" {
		t.Fatalf("data freshness = %q", route.DataFreshness)
	}
}

func TestRoutePreservesCachedFreshness(t *testing.T) {
	c := geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.GCJ02}
	route, err := NewService(cachedRouteProvider{}, "").Route(context.Background(), RouteRequest{Origin: c, Destination: c, Mode: Walking})
	if err != nil {
		t.Fatal(err)
	}
	if route.DataFreshness != "cached" {
		t.Fatalf("route freshness = %q, want cached", route.DataFreshness)
	}
}

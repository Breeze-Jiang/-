package places

import (
	"alongtu/backend/internal/platform/cursor"
	"alongtu/backend/internal/platform/geo"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepo struct {
	page Page
	err  error
}

func (f fakeRepo) Nearby(context.Context, NearbyQuery) (Page, error) { return f.page, f.err }
func (f fakeRepo) Get(context.Context, string) (Place, error)        { return Place{}, f.err }
func (f fakeRepo) AddConfirmation(context.Context, Confirmation) (Confirmation, error) {
	return Confirmation{}, f.err
}
func (f fakeRepo) AddFeedback(context.Context, Feedback) (Feedback, error)   { return Feedback{}, f.err }
func (f fakeRepo) Materialize(_ context.Context, p []Place) ([]Place, error) { return p, f.err }

type fakeProvider struct {
	page Page
	err  error
}

type pagedProvider struct{ pages map[int]Page }

func (p pagedProvider) Nearby(_ context.Context, query NearbyQuery) (Page, error) {
	return p.pages[query.Page], nil
}

func (p pagedProvider) Search(_ context.Context, query SearchQuery) (Page, error) {
	return p.pages[query.Page], nil
}

type uncoveredProvider struct{ fakeProvider }

func (uncoveredProvider) Coverage(Category) string { return "not_covered" }

type placeMetricRecorder struct{ events []string }

func (m *placeMetricRecorder) ObservePlaceQuery(operation, outcome string) {
	m.events = append(m.events, operation+":"+outcome)
}

func (f fakeProvider) Nearby(context.Context, NearbyQuery) (Page, error) { return f.page, f.err }
func (f fakeProvider) Search(context.Context, SearchQuery) (Page, error) { return f.page, f.err }
func TestNearbyFallsBackToStoredData(t *testing.T) {
	stored := Page{Items: []Place{{ID: "one"}}}
	s := NewService(fakeRepo{page: stored}, fakeProvider{err: errors.New("down")}, cursor.New("test-cursor-secret", time.Hour))
	got, err := s.Nearby(context.Background(), NearbyQuery{Center: geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.WGS84}, RadiusMeters: 1000, Limit: 20})
	if err != nil || !got.Degraded || got.DataFreshness != "degraded" || got.Coverage != "covered" || len(got.Items) != 1 {
		t.Fatalf("unexpected fallback: %#v %v", got, err)
	}
}
func TestConfirmationValidation(t *testing.T) {
	s := NewService(fakeRepo{}, fakeProvider{}, cursor.New("test-cursor-secret", time.Hour))
	_, err := s.Confirm(context.Background(), Confirmation{Kind: "capacity", Result: "confirmed", IdempotencyKey: "12345678"})
	if err == nil {
		t.Fatal("expected unsupported kind")
	}
}

func TestFeedbackValidation(t *testing.T) {
	s := NewService(fakeRepo{}, fakeProvider{}, cursor.New("test-cursor-secret", time.Hour))
	_, err := s.SubmitFeedback(context.Background(), Feedback{Kind: "wrong", Details: "入口在北门", IdempotencyKey: "12345678"})
	if err == nil {
		t.Fatal("expected unsupported feedback kind")
	}
	_, err = s.SubmitFeedback(context.Background(), Feedback{Kind: "entrance", Details: " ", IdempotencyKey: "12345678"})
	if err == nil {
		t.Fatal("expected blank feedback details rejection")
	}
}

func TestPlaceQueryMetricsDistinguishDegradedAndEmpty(t *testing.T) {
	metrics := &placeMetricRecorder{}
	stored := Page{Items: []Place{{ID: "one"}}}
	service := NewService(fakeRepo{page: stored}, fakeProvider{err: errors.New("down")}, cursor.New("metric-cursor-secret", time.Hour), metrics)
	_, err := service.Nearby(context.Background(), NearbyQuery{Center: geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.WGS84}, RadiusMeters: 1000, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	emptyService := NewService(fakeRepo{}, fakeProvider{page: Page{Items: []Place{}}}, cursor.New("metric-cursor-secret", time.Hour), metrics)
	_, err = emptyService.Search(context.Background(), SearchQuery{Keyword: "厕所", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics.events) != 2 || metrics.events[0] != "nearby:degraded" || metrics.events[1] != "search:empty" {
		t.Fatalf("unexpected place metrics: %v", metrics.events)
	}
}

func TestMergePreservesProviderFreshness(t *testing.T) {
	page := merge(Page{}, Page{Items: []Place{{ID: "provider-place"}}, DataFreshness: "cached"}, 20)
	if page.DataFreshness != "cached" {
		t.Fatalf("merged page freshness = %q, want cached", page.DataFreshness)
	}
}

func TestUncoveredCategoryBypassesStoredAndCachedResults(t *testing.T) {
	service := NewService(fakeRepo{page: Page{Items: []Place{{ID: "must-not-leak"}}}}, uncoveredProvider{fakeProvider{page: Page{Items: []Place{{ID: "must-not-call"}}}}}, cursor.New("coverage-cursor-secret", time.Hour))
	page, err := service.Nearby(context.Background(), NearbyQuery{Center: geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.WGS84}, Category: CategoryParking, RadiusMeters: 1000, Limit: 20})
	if err != nil || page.Coverage != "not_covered" || len(page.Items) != 0 {
		t.Fatalf("uncovered category must return an empty explicit response: page=%+v err=%v", page, err)
	}
	_, err = service.Nearby(context.Background(), NearbyQuery{Center: geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.WGS84}, Category: CategoryParking, RadiusMeters: 50001, Limit: 20})
	if !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid uncovered query must remain a client error, got %v", err)
	}
}

func TestMergeSortsByDistanceBeforeApplyingLimit(t *testing.T) {
	far := 900.0
	near := 100.0
	page := merge(
		Page{Items: []Place{{ID: "stored-far", DistanceMeters: &far}}},
		Page{Items: []Place{{ID: "provider-near", DistanceMeters: &near}}, DataFreshness: "live"},
		1,
	)
	if len(page.Items) != 1 || page.Items[0].ID != "provider-near" {
		t.Fatalf("nearby merge did not retain the nearest place: %+v", page.Items)
	}
}

func TestMergeOrdersEqualDistancesByPlaceID(t *testing.T) {
	distance := 100.0
	page := merge(
		Page{Items: []Place{{ID: "place-b", DistanceMeters: &distance}}},
		Page{Items: []Place{{ID: "place-a", DistanceMeters: &distance}}, DataFreshness: "live"},
		2,
	)
	if len(page.Items) != 2 || page.Items[0].ID != "place-a" || page.Items[1].ID != "place-b" {
		t.Fatalf("equal-distance places must have a stable ID tie-breaker: %+v", page.Items)
	}
}

func TestNearbyCursorDropsRepeatedProviderBoundaryItem(t *testing.T) {
	firstDistance := 10.0
	secondDistance := 15.0
	boundaryDistance := 20.0
	nextDistance := 30.0
	provider := pagedProvider{pages: map[int]Page{
		1: {Items: []Place{{ID: "place-1", DistanceMeters: &firstDistance}, {ID: "place-2", DistanceMeters: &secondDistance}, {ID: "place-3", DistanceMeters: &boundaryDistance}}},
		2: {Items: []Place{{ID: "place-3", DistanceMeters: &boundaryDistance}, {ID: "place-4", DistanceMeters: &nextDistance}}},
	}}
	service := NewService(fakeRepo{}, provider, cursor.New("nearby-cursor-secret", time.Hour))
	query := NearbyQuery{Center: geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.WGS84}, RadiusMeters: 1000, Limit: 3}
	first, err := service.Nearby(context.Background(), query)
	if err != nil || first.NextCursor == "" || len(first.Items) != 3 {
		t.Fatalf("unexpected first page: %#v err=%v", first, err)
	}
	query.Cursor = first.NextCursor
	second, err := service.Nearby(context.Background(), query)
	if err != nil || len(second.Items) != 1 || second.Items[0].ID != "place-4" || second.NextCursor != "" {
		t.Fatalf("repeated provider boundary leaked into next page: %#v err=%v", second, err)
	}
}

func TestNearbyCursorUsesStableTieBreakerForProviderBoundary(t *testing.T) {
	distance := 20.0
	nextDistance := 30.0
	provider := pagedProvider{pages: map[int]Page{
		1: {Items: []Place{{ID: "place-b", DistanceMeters: &distance}, {ID: "place-a", DistanceMeters: &distance}}},
		2: {Items: []Place{{ID: "place-b", DistanceMeters: &distance}, {ID: "place-c", DistanceMeters: &nextDistance}}},
	}}
	service := NewService(fakeRepo{}, provider, cursor.New("nearby-tie-cursor-secret", time.Hour))
	query := NearbyQuery{Center: geo.Coordinate{Latitude: 39.9, Longitude: 116.4, System: geo.WGS84}, RadiusMeters: 1000, Limit: 2}
	first, err := service.Nearby(context.Background(), query)
	if err != nil || first.NextCursor == "" || len(first.Items) != 2 || first.Items[0].ID != "place-a" || first.Items[1].ID != "place-b" {
		t.Fatalf("unexpected first tied page: %#v err=%v", first, err)
	}
	query.Cursor = first.NextCursor
	second, err := service.Nearby(context.Background(), query)
	if err != nil || len(second.Items) != 1 || second.Items[0].ID != "place-c" {
		t.Fatalf("tied provider boundary leaked a duplicate: %#v err=%v", second, err)
	}
}

func TestSearchCursorDropsRepeatedProviderBoundaryItem(t *testing.T) {
	provider := pagedProvider{pages: map[int]Page{
		1: {Items: []Place{{ID: "place-1"}, {ID: "place-2"}, {ID: "place-3"}}},
		2: {Items: []Place{{ID: "place-3"}, {ID: "place-4"}}},
	}}
	service := NewService(fakeRepo{}, provider, cursor.New("search-cursor-secret", time.Hour))
	first, err := service.Search(context.Background(), SearchQuery{Keyword: "厕所", Limit: 3})
	if err != nil || first.NextCursor == "" || len(first.Items) != 3 {
		t.Fatalf("unexpected first page: %#v err=%v", first, err)
	}
	second, err := service.Search(context.Background(), SearchQuery{Keyword: "厕所", Limit: 3, Cursor: first.NextCursor})
	if err != nil || len(second.Items) != 1 || second.Items[0].ID != "place-4" || second.NextCursor != "" {
		t.Fatalf("repeated provider boundary leaked into next search page: %#v err=%v", second, err)
	}
}

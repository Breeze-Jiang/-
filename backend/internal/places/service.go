package places

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"alongtu/backend/internal/platform/cursor"
	"alongtu/backend/internal/platform/geo"
	"alongtu/backend/internal/platform/httpx"
)

var (
	ErrInvalidQuery        = errors.New("invalid place query")
	ErrInvalidConfirmation = errors.New("invalid confirmation")
	ErrProviderUnavailable = errors.New("place provider unavailable")
	ErrRepository          = errors.New("place repository unavailable")
)

type Service struct {
	repo     Repository
	provider Provider
	cursors  *cursor.Codec
	metrics  Metrics
}
type Metrics interface {
	ObservePlaceQuery(operation, outcome string)
}
type coverageReporter interface {
	Coverage(Category) string
}

func NewService(repo Repository, provider Provider, cursors *cursor.Codec, metrics ...Metrics) *Service {
	service := &Service{repo: repo, provider: provider, cursors: cursors}
	if len(metrics) > 0 {
		service.metrics = metrics[0]
	}
	return service
}
func (s *Service) Nearby(ctx context.Context, q NearbyQuery) (Page, error) {
	if err := q.Center.Validate(); err != nil {
		s.observeQuery("nearby", "invalid")
		return Page{}, fmt.Errorf("%w: %v", ErrInvalidQuery, err)
	}
	if !IsKnownCategory(q.Category) {
		s.observeQuery("nearby", "invalid")
		return Page{}, fmt.Errorf("%w: unsupported category", ErrInvalidQuery)
	}
	if q.RadiusMeters == 0 {
		q.RadiusMeters = 5000
	}
	if q.RadiusMeters < 100 || q.RadiusMeters > 50000 {
		s.observeQuery("nearby", "invalid")
		return Page{}, fmt.Errorf("%w: radius must be between 100 and 50000", ErrInvalidQuery)
	}
	if q.Limit <= 0 || q.Limit > 50 {
		q.Limit = 20
	}
	fingerprint := cursor.Fingerprint(fmt.Sprintf("%.5f", q.Center.Latitude), fmt.Sprintf("%.5f", q.Center.Longitude), string(q.Center.System), string(q.Category), strconv.Itoa(q.RadiusMeters), strconv.Itoa(q.Limit))
	position, decodeErr := s.cursors.Decode(q.Cursor, "nearby", fingerprint)
	if decodeErr != nil {
		s.observeQuery("nearby", "invalid_cursor")
		return Page{}, decodeErr
	}
	if providerCoverage(s.provider, q.Category) == "not_covered" {
		s.observeQuery("nearby", "not_covered")
		return Page{Items: []Place{}, DataFreshness: "live", Coverage: "not_covered"}, nil
	}
	q.Page = position.Page
	localQuery := q
	if localQuery.Center.System == geo.GCJ02 {
		localQuery.Center = geo.GCJ02ToWGS84(localQuery.Center)
	}
	local, err := Page{Items: []Place{}, Coverage: "covered"}, error(nil)
	if position.Page == 1 {
		local, err = s.repo.Nearby(ctx, localQuery)
		if err != nil {
			s.observeQuery("nearby", "error")
			return Page{}, fmt.Errorf("%w: %v", ErrRepository, err)
		}
		if local.Coverage == "" {
			// Repository data is still within the provider coverage selected above.
			// Keep the API's required coverage field explicit on degraded fallback.
			local.Coverage = "covered"
		}
	}
	providerQuery := q
	if providerQuery.Center.System == geo.WGS84 {
		providerQuery.Center = geo.WGS84ToGCJ02(providerQuery.Center)
	}
	remote, pErr := s.provider.Nearby(ctx, providerQuery)
	if pErr != nil {
		if err == nil && len(local.Items) > 0 {
			local.Degraded = true
			local.DataFreshness = "degraded"
			s.observeQuery("nearby", "degraded")
			return local, nil
		}
		s.observeQuery("nearby", "error")
		return Page{}, fmt.Errorf("%w: %v", ErrProviderUnavailable, pErr)
	}
	materialized, err := s.repo.Materialize(ctx, remote.Items)
	if err != nil {
		s.observeQuery("nearby", "error")
		return Page{}, fmt.Errorf("%w: %v", ErrRepository, err)
	}
	remote.Items = itemsAfterBoundary(materialized, position)
	out := merge(local, remote, q.Limit)
	if len(materialized) == q.Limit {
		out.NextCursor = s.cursors.Encode("nearby", fingerprint, nearbyNextPosition(position.Page+1, materialized))
	}
	s.observeQuery("nearby", resultOutcome(len(out.Items)))
	return out, nil
}
func (s *Service) Search(ctx context.Context, q SearchQuery) (Page, error) {
	q.Keyword = strings.TrimSpace(q.Keyword)
	if len([]rune(q.Keyword)) < 1 || len([]rune(q.Keyword)) > 80 {
		s.observeQuery("search", "invalid")
		return Page{}, fmt.Errorf("%w: keyword length invalid", ErrInvalidQuery)
	}
	if !IsKnownCategory(q.Category) {
		s.observeQuery("search", "invalid")
		return Page{}, fmt.Errorf("%w: unsupported category", ErrInvalidQuery)
	}
	if q.Limit <= 0 || q.Limit > 50 {
		q.Limit = 20
	}
	fingerprint := cursor.Fingerprint(q.Keyword, q.CityCode, string(q.Category), strconv.Itoa(q.Limit))
	position, decodeErr := s.cursors.Decode(q.Cursor, "search", fingerprint)
	if decodeErr != nil {
		s.observeQuery("search", "invalid_cursor")
		return Page{}, decodeErr
	}
	if providerCoverage(s.provider, q.Category) == "not_covered" {
		s.observeQuery("search", "not_covered")
		return Page{Items: []Place{}, DataFreshness: "live", Coverage: "not_covered"}, nil
	}
	q.Page = position.Page
	if q.Center != nil && q.Center.System == geo.WGS84 {
		converted := geo.WGS84ToGCJ02(*q.Center)
		q.Center = &converted
	}
	page, err := s.provider.Search(ctx, q)
	if err != nil {
		s.observeQuery("search", "error")
		return Page{}, fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
	}
	if page.Coverage == "" {
		page.Coverage = "covered"
	}
	materialized, err := s.repo.Materialize(ctx, page.Items)
	if err != nil {
		s.observeQuery("search", "error")
		return page, fmt.Errorf("%w: %v", ErrRepository, err)
	}
	page.Items = itemsAfterBoundary(materialized, position)
	if len(materialized) == q.Limit {
		page.NextCursor = s.cursors.Encode("search", fingerprint, nextPosition(position.Page+1, materialized))
	}
	s.observeQuery("search", resultOutcome(len(page.Items)))
	return page, nil
}
func (s *Service) Get(ctx context.Context, id string) (Place, error) {
	p, err := s.repo.Get(ctx, id)
	if errors.Is(err, httpx.ErrNotFound) {
		s.observeQuery("detail", "not_found")
		return Place{}, httpx.ErrNotFound
	}
	if err != nil {
		s.observeQuery("detail", "error")
		return Place{}, err
	}
	s.observeQuery("detail", "found")
	return p, err
}
func (s *Service) observeQuery(operation, outcome string) {
	if s.metrics != nil {
		s.metrics.ObservePlaceQuery(operation, outcome)
	}
}
func resultOutcome(count int) string {
	if count == 0 {
		return "empty"
	}
	return "results"
}
func providerCoverage(provider Provider, category Category) string {
	if reporter, ok := provider.(coverageReporter); ok && reporter.Coverage(category) == "not_covered" {
		return "not_covered"
	}
	return "covered"
}
func (s *Service) Confirm(ctx context.Context, c Confirmation) (Confirmation, error) {
	if c.Kind != "exists" && c.Kind != "opening_hours" {
		return Confirmation{}, fmt.Errorf("%w: unsupported confirmation kind", ErrInvalidConfirmation)
	}
	if c.Result != "confirmed" && c.Result != "incorrect" && c.Result != "unknown" {
		return Confirmation{}, fmt.Errorf("%w: unsupported confirmation result", ErrInvalidConfirmation)
	}
	if len(c.IdempotencyKey) < 8 || len(c.IdempotencyKey) > 200 {
		return Confirmation{}, fmt.Errorf("%w: invalid idempotency key", ErrInvalidConfirmation)
	}
	return s.repo.AddConfirmation(ctx, c)
}
func (s *Service) SubmitFeedback(ctx context.Context, feedback Feedback) (Feedback, error) {
	if feedback.Kind != "entrance" && feedback.Kind != "incorrect_info" {
		return Feedback{}, fmt.Errorf("%w: unsupported feedback kind", ErrInvalidConfirmation)
	}
	if details := strings.TrimSpace(feedback.Details); len([]rune(details)) < 1 || len([]rune(details)) > 500 {
		return Feedback{}, fmt.Errorf("%w: invalid feedback details", ErrInvalidConfirmation)
	} else {
		feedback.Details = details
	}
	if len(feedback.IdempotencyKey) < 8 || len(feedback.IdempotencyKey) > 200 {
		return Feedback{}, fmt.Errorf("%w: invalid idempotency key", ErrInvalidConfirmation)
	}
	return s.repo.AddFeedback(ctx, feedback)
}
func merge(a, b Page, limit int) Page {
	seen := map[string]bool{}
	freshness := b.DataFreshness
	if freshness == "" {
		freshness = "live"
	}
	coverage := b.Coverage
	if coverage == "" {
		coverage = "covered"
	}
	out := Page{DataFreshness: freshness, Coverage: coverage}
	for _, p := range append(a.Items, b.Items...) {
		if p.ID == "" || seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		out.Items = append(out.Items, p)
	}
	sort.SliceStable(out.Items, func(i, j int) bool { return placeBefore(out.Items[i], out.Items[j]) })
	if len(out.Items) > limit {
		out.Items = out.Items[:limit]
	}
	return out
}

func itemsAfterBoundary(items []Place, boundary cursor.Position) []Place {
	if boundary.LastID == "" {
		return items
	}
	out := make([]Place, 0, len(items))
	for _, item := range items {
		if item.ID == boundary.LastID {
			continue
		}
		if boundary.LastDistanceMeters != nil && item.DistanceMeters != nil {
			if *item.DistanceMeters < *boundary.LastDistanceMeters || (*item.DistanceMeters == *boundary.LastDistanceMeters && item.ID <= boundary.LastID) {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}

func nextPosition(page int, items []Place) cursor.Position {
	for index := len(items) - 1; index >= 0; index-- {
		if items[index].ID != "" {
			return cursor.Position{Page: page, LastDistanceMeters: items[index].DistanceMeters, LastID: items[index].ID}
		}
	}
	return cursor.Position{Page: page}
}

func nearbyNextPosition(page int, items []Place) cursor.Position {
	var latest *Place
	for index := range items {
		if items[index].ID == "" {
			continue
		}
		if latest == nil || placeBefore(*latest, items[index]) {
			latest = &items[index]
		}
	}
	if latest == nil {
		return cursor.Position{Page: page}
	}
	return cursor.Position{Page: page, LastDistanceMeters: latest.DistanceMeters, LastID: latest.ID}
}

func placeBefore(left, right Place) bool {
	if left.DistanceMeters == nil && right.DistanceMeters == nil {
		return left.ID < right.ID
	}
	if left.DistanceMeters == nil {
		return false
	}
	if right.DistanceMeters == nil {
		return true
	}
	if *left.DistanceMeters == *right.DistanceMeters {
		return left.ID < right.ID
	}
	return *left.DistanceMeters < *right.DistanceMeters
}

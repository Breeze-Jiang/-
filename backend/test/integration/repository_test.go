//go:build integration

package integration

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"alongtu/backend/internal/auth"
	"alongtu/backend/internal/places"
	"alongtu/backend/internal/platform/geo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDatabaseURL = "postgres://alongtu:alongtu@localhost:5432/alongtu?sslmode=disable"

func TestRepositoriesAgainstPostGIS(t *testing.T) {
	pool := openDatabase(t)
	resetDatabase(t, pool)

	t.Run("nearby places are ordered by WGS84 distance", func(t *testing.T) {
		resetDatabase(t, pool)
		testNearbyOrdering(t, pool)
	})
	t.Run("provider place materialization is concurrent and idempotent", func(t *testing.T) {
		resetDatabase(t, pool)
		testConcurrentMaterialization(t, pool)
	})
	t.Run("stale provider cache cannot overwrite a newer source record", func(t *testing.T) {
		resetDatabase(t, pool)
		testStaleMaterializationCannotOverwriteFreshData(t, pool)
	})
	t.Run("phone user creation is concurrent and idempotent", func(t *testing.T) {
		resetDatabase(t, pool)
		testConcurrentUserCreation(t, pool)
	})
	t.Run("confirmation keys preserve the original record", func(t *testing.T) {
		resetDatabase(t, pool)
		testConfirmationIdempotency(t, pool)
	})
	t.Run("fact feedback is idempotent and does not change trust summary", func(t *testing.T) {
		resetDatabase(t, pool)
		testFeedbackIdempotency(t, pool)
	})
	t.Run("device sessions rotate once and can be revoked", func(t *testing.T) {
		resetDatabase(t, pool)
		testSessionLifecycle(t, pool)
	})
}

func openDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("INTEGRATION_DATABASE_URL")
	if url == "" {
		url = defaultDatabaseURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("connect to integration database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func resetDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := pool.Exec(ctx, `TRUNCATE place_confirmations, place_opening_hours, place_sources, places, device_sessions, auth_identities, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("reset integration database: %v", err)
	}
}

func testNearbyOrdering(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	for _, item := range []places.Place{
		{ID: "amap:far", Name: "较远厕所", Category: places.CategoryToilet, Location: geo.Coordinate{Latitude: 39.9200, Longitude: 116.4200, System: geo.GCJ02}},
		{ID: "amap:near", Name: "较近厕所", Category: places.CategoryToilet, Location: geo.Coordinate{Latitude: 39.9100, Longitude: 116.4000, System: geo.GCJ02}},
	} {
		if _, err := places.NewPostgresRepository(pool).Materialize(ctx, []places.Place{item}); err != nil {
			t.Fatalf("materialize %s: %v", item.Name, err)
		}
	}
	page, err := places.NewPostgresRepository(pool).Nearby(ctx, places.NearbyQuery{
		Center:       geo.GCJ02ToWGS84(geo.Coordinate{Latitude: 39.9099, Longitude: 116.3999, System: geo.GCJ02}),
		Category:     places.CategoryToilet,
		RadiusMeters: 5000,
		Limit:        20,
	})
	if err != nil {
		t.Fatalf("nearby query: %v", err)
	}
	if len(page.Items) != 2 || page.Items[0].Name != "较近厕所" {
		t.Fatalf("unexpected nearby order: %+v", page.Items)
	}
	if page.Items[0].DistanceMeters == nil || page.Items[1].DistanceMeters == nil || *page.Items[0].DistanceMeters >= *page.Items[1].DistanceMeters {
		t.Fatalf("distances are not increasing: %+v", page.Items)
	}
	if page.Items[0].Location.System != geo.GCJ02 {
		t.Fatalf("stored place display coordinate system = %q, want %q", page.Items[0].Location.System, geo.GCJ02)
	}
}

func testConcurrentMaterialization(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	repo := places.NewPostgresRepository(pool)
	item := places.Place{ID: "amap:shared-poi", Name: "并发地点", Category: places.CategoryToilet, Location: geo.Coordinate{Latitude: 39.91, Longitude: 116.40, System: geo.GCJ02}}
	ids, errs := concurrentStrings(8, func() (string, error) {
		result, err := repo.Materialize(context.Background(), []places.Place{item})
		if err != nil {
			return "", err
		}
		return result[0].ID, nil
	})
	assertNoErrors(t, errs)
	assertSingleValue(t, ids)
	var placesCount, sourcesCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM places`).Scan(&placesCount); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM place_sources WHERE provider='amap' AND external_id='shared-poi'`).Scan(&sourcesCount); err != nil {
		t.Fatal(err)
	}
	if placesCount != 1 || sourcesCount != 1 {
		t.Fatalf("expected one place/source, got places=%d sources=%d", placesCount, sourcesCount)
	}
}

func testStaleMaterializationCannotOverwriteFreshData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	repo := places.NewPostgresRepository(pool)
	older := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	newer := older.Add(time.Hour)
	place := places.Place{ID: "amap:timestamp-poi", Name: "旧来源数据", Category: places.CategoryToilet, Location: geo.Coordinate{Latitude: 39.91, Longitude: 116.40, System: geo.GCJ02}, SourceUpdatedAt: &older}
	materialized, err := repo.Materialize(context.Background(), []places.Place{place})
	if err != nil {
		t.Fatal(err)
	}
	place.Name = "新来源数据"
	place.SourceUpdatedAt = &newer
	if _, err = repo.Materialize(context.Background(), []places.Place{place}); err != nil {
		t.Fatal(err)
	}
	place.Name = "过期缓存数据"
	place.SourceUpdatedAt = &older
	if _, err = repo.Materialize(context.Background(), []places.Place{place}); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.Get(context.Background(), materialized[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != "新来源数据" || stored.SourceUpdatedAt == nil || !stored.SourceUpdatedAt.Equal(newer) {
		t.Fatalf("stale cache overwrote source state: %+v", stored)
	}
}

func testConcurrentUserCreation(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	repo := auth.NewPostgresRepository(pool)
	ids, errs := concurrentStrings(8, func() (string, error) {
		return repo.FindOrCreateUser(context.Background(), "+8613800138000")
	})
	assertNoErrors(t, errs)
	assertSingleValue(t, ids)
	var users, identities int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM users`).Scan(&users); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM auth_identities`).Scan(&identities); err != nil {
		t.Fatal(err)
	}
	if users != 1 || identities != 1 {
		t.Fatalf("expected one user/identity, got users=%d identities=%d", users, identities)
	}
}

func testConfirmationIdempotency(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	authRepo := auth.NewPostgresRepository(pool)
	userID, err := authRepo.FindOrCreateUser(ctx, "+8613800138001")
	if err != nil {
		t.Fatal(err)
	}
	sessionID := uuid.NewString()
	sessionHash := sha256.Sum256([]byte("confirmation-session"))
	if err = authRepo.CreateSession(ctx, sessionID, userID, sessionHash[:], "confirmation-device", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	materialized, err := places.NewPostgresRepository(pool).Materialize(ctx, []places.Place{{ID: "amap:confirm-poi", Name: "确认地点", Category: places.CategoryToilet, Location: geo.Coordinate{Latitude: 39.91, Longitude: 116.40, System: geo.GCJ02}}})
	if err != nil {
		t.Fatal(err)
	}
	repo := places.NewPostgresRepository(pool)
	request := places.Confirmation{PlaceID: materialized[0].ID, UserID: userID, SessionID: sessionID, Kind: "exists", Result: "confirmed", IdempotencyKey: "same-confirmation-key"}
	results, errs := concurrentConfirmations(8, func() (places.Confirmation, error) { return repo.AddConfirmation(ctx, request) })
	assertNoErrors(t, errs)
	ids := make([]string, len(results))
	for i := range results {
		ids[i] = results[i].ID
	}
	assertSingleValue(t, ids)
	conflict := request
	conflict.Result = "incorrect"
	if _, err = repo.AddConfirmation(ctx, conflict); !errors.Is(err, places.ErrIdempotencyConflict) {
		t.Fatalf("expected idempotency conflict, got %v", err)
	}
	tied := request
	tied.Result = "incorrect"
	tied.IdempotencyKey = "tied-confirmation-key"
	tiedConfirmation, err := repo.AddConfirmation(ctx, tied)
	if err != nil {
		t.Fatal(err)
	}
	tieTime := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	if _, err = pool.Exec(ctx, `UPDATE place_confirmations SET created_at=$1 WHERE id=$2`, tieTime, results[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE place_confirmations SET created_at=$1 WHERE id=$2`, tieTime, tiedConfirmation.ID); err != nil {
		t.Fatal(err)
	}
	var expectedResult string
	if err = pool.QueryRow(ctx, `SELECT result::text FROM place_confirmations WHERE place_id=$1 AND kind='exists' ORDER BY created_at DESC, id DESC LIMIT 1`, materialized[0].ID).Scan(&expectedResult); err != nil {
		t.Fatal(err)
	}
	storedDetail, err := repo.Get(ctx, materialized[0].ID)
	if err != nil || storedDetail.Trust.Exists.Status != expectedResult || storedDetail.Trust.Exists.ConfirmedAt == nil || !storedDetail.Trust.Exists.ConfirmedAt.Equal(tieTime) {
		t.Fatalf("detail did not use the deterministic latest confirmation: detail=%+v err=%v", storedDetail, err)
	}
	nearby, err := repo.Nearby(ctx, places.NearbyQuery{Center: geo.GCJ02ToWGS84(geo.Coordinate{Latitude: 39.91, Longitude: 116.40, System: geo.GCJ02}), Category: places.CategoryToilet, RadiusMeters: 1000, Limit: 20})
	if err != nil || len(nearby.Items) != 1 || nearby.Items[0].Trust.Exists.Status != expectedResult || nearby.Items[0].Trust.Exists.ConfirmedAt == nil || !nearby.Items[0].Trust.Exists.ConfirmedAt.Equal(tieTime) || nearby.Items[0].Trust.Exists.Source != "community" {
		t.Fatalf("nearby result did not expose the deterministic latest public trust summary: page=%+v err=%v", nearby, err)
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM place_confirmations`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("expected two confirmations, count=%d err=%v", count, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE users SET status='blocked' WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}
	blocked := request
	blocked.IdempotencyKey = "blocked-user-confirmation"
	if _, err = repo.AddConfirmation(ctx, blocked); !errors.Is(err, places.ErrWriteNotAllowed) {
		t.Fatalf("expected blocked user rejection, got %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE users SET status='active' WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err = authRepo.RevokeSession(ctx, sessionID); err != nil {
		t.Fatal(err)
	}
	revoked := request
	revoked.IdempotencyKey = "revoked-session-confirmation"
	if _, err = repo.AddConfirmation(ctx, revoked); !errors.Is(err, places.ErrWriteNotAllowed) {
		t.Fatalf("expected revoked session rejection, got %v", err)
	}
}

func testFeedbackIdempotency(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	authRepo := auth.NewPostgresRepository(pool)
	userID, err := authRepo.FindOrCreateUser(ctx, "+8613800138003")
	if err != nil {
		t.Fatal(err)
	}
	sessionID := uuid.NewString()
	sessionHash := sha256.Sum256([]byte("feedback-session"))
	if err = authRepo.CreateSession(ctx, sessionID, userID, sessionHash[:], "feedback-device", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	materialized, err := places.NewPostgresRepository(pool).Materialize(ctx, []places.Place{{ID: "amap:feedback-poi", Name: "反馈地点", Category: places.CategoryToilet, Location: geo.Coordinate{Latitude: 39.91, Longitude: 116.40, System: geo.GCJ02}}})
	if err != nil {
		t.Fatal(err)
	}
	repo := places.NewPostgresRepository(pool)
	request := places.Feedback{PlaceID: materialized[0].ID, UserID: userID, SessionID: sessionID, Kind: "entrance", Details: "入口在北门", IdempotencyKey: "same-feedback-key"}
	first, err := repo.AddFeedback(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.AddFeedback(ctx, request)
	if err != nil || second.ID != first.ID {
		t.Fatalf("feedback idempotency failed: first=%+v second=%+v err=%v", first, second, err)
	}
	conflict := request
	conflict.Details = "入口在南门"
	if _, err = repo.AddFeedback(ctx, conflict); !errors.Is(err, places.ErrIdempotencyConflict) {
		t.Fatalf("expected feedback idempotency conflict, got %v", err)
	}
	stored, err := repo.Get(ctx, materialized[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Trust.Exists.Status != "unknown" || stored.Trust.OpeningHours.Status != "unknown" {
		t.Fatalf("fact feedback must not become a public trust summary: %+v", stored.Trust)
	}
}

func testSessionLifecycle(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	repo := auth.NewPostgresRepository(pool)
	userID, err := repo.FindOrCreateUser(ctx, "+8613800138002")
	if err != nil {
		t.Fatal(err)
	}
	sessionID := uuid.NewString()
	oldHash := sha256.Sum256([]byte("old-refresh"))
	if err = repo.CreateSession(ctx, sessionID, userID, oldHash[:], "integration-device", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	newHash := sha256.Sum256([]byte("new-refresh"))
	issued := auth.Tokens{AccessToken: "access", RefreshToken: "refresh", AccessExpiresAt: time.Now().Add(time.Minute), RefreshExpiresAt: time.Now().Add(2 * time.Hour)}
	got, err := repo.RotateSession(ctx, sessionID, oldHash[:], func(string) (auth.Tokens, []byte, error) { return issued, newHash[:], nil })
	if err != nil || got.RefreshToken != issued.RefreshToken {
		t.Fatalf("rotate session: tokens=%+v err=%v", got, err)
	}
	if _, err = repo.RotateSession(ctx, sessionID, oldHash[:], func(string) (auth.Tokens, []byte, error) { return issued, newHash[:], nil }); !errors.Is(err, auth.ErrInvalidRefresh) {
		t.Fatalf("expected replay rejection, got %v", err)
	}
	active, err := repo.IsActiveSession(ctx, sessionID, userID)
	if err != nil || active {
		t.Fatalf("replayed session should be revoked: active=%v err=%v", active, err)
	}
}

func concurrentStrings(count int, fn func() (string, error)) ([]string, []error) {
	values := make([]string, count)
	errs := make([]error, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i := range count {
		go func(index int) {
			defer wg.Done()
			values[index], errs[index] = fn()
		}(i)
	}
	wg.Wait()
	return values, errs
}

func concurrentConfirmations(count int, fn func() (places.Confirmation, error)) ([]places.Confirmation, []error) {
	values := make([]places.Confirmation, count)
	errs := make([]error, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i := range count {
		go func(index int) {
			defer wg.Done()
			values[index], errs[index] = fn()
		}(i)
	}
	wg.Wait()
	return values, errs
}

func assertNoErrors(t *testing.T, errs []error) {
	t.Helper()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent call %d failed: %v", i, err)
		}
	}
}

func assertSingleValue(t *testing.T, values []string) {
	t.Helper()
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	if len(sorted) == 0 || sorted[0] == "" || sorted[0] != sorted[len(sorted)-1] {
		t.Fatalf("expected one non-empty value, got %v", values)
	}
}

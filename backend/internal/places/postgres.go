package places

import (
	"alongtu/backend/internal/platform/database/dbgen"
	"alongtu/backend/internal/platform/geo"
	"alongtu/backend/internal/platform/httpx"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *dbgen.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool, queries: dbgen.New(pool)}
}

func (r *PostgresRepository) Nearby(ctx context.Context, q NearbyQuery) (Page, error) {
	if q.Center.System != geo.WGS84 {
		return Page{Items: []Place{}, DataFreshness: "stored"}, nil
	}
	rows, err := r.queries.NearbyPlaces(ctx, dbgen.NearbyPlacesParams{Longitude: q.Center.Longitude, Latitude: q.Center.Latitude, CategoryID: text(string(q.Category)), RadiusMeters: float64(q.RadiusMeters), PageSize: int32(q.Limit)})
	if err != nil {
		return Page{}, err
	}
	out := Page{Items: []Place{}, DataFreshness: "stored"}
	for _, row := range rows {
		distance := row.DistanceMeters
		p := Place{ID: uuidString(row.ID), Name: row.Name, Category: storedCategory(row.CategoryID), Address: row.Address.String, CityCode: row.CityCode.String, DistrictCode: row.DistrictCode.String, Location: geo.Coordinate{Latitude: row.Gcj02Latitude, Longitude: row.Gcj02Longitude, System: geo.GCJ02}, DistanceMeters: &distance, SourceUpdatedAt: timePtr(row.SourceUpdatedAt), Trust: TrustSummary{Exists: nearbyTrustItem(row.ExistsResult, row.ExistsConfirmedAt), OpeningHours: nearbyTrustItem(row.OpeningHoursResult, row.OpeningHoursConfirmedAt)}}
		out.Items = append(out.Items, p)
	}
	return out, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (Place, error) {
	pid, err := parseUUID(id)
	if err != nil {
		return Place{}, httpx.ErrNotFound
	}
	row, err := r.queries.GetPlace(ctx, pid)
	if errors.Is(err, pgx.ErrNoRows) {
		return Place{}, httpx.ErrNotFound
	}
	if err != nil {
		return Place{}, err
	}
	p := Place{ID: uuidString(row.ID), Name: row.Name, Category: storedCategory(row.CategoryID), Address: row.Address.String, CityCode: row.CityCode.String, DistrictCode: row.DistrictCode.String, Location: geo.Coordinate{Latitude: row.Gcj02Latitude, Longitude: row.Gcj02Longitude, System: geo.GCJ02}, SourceUpdatedAt: timePtr(row.SourceUpdatedAt), Trust: TrustSummary{Exists: TrustItem{Status: "unknown"}, OpeningHours: TrustItem{Status: "unknown"}}}
	confirmations, err := r.queries.LatestPlaceConfirmations(ctx, pid)
	if err != nil {
		return Place{}, err
	}
	for _, row := range confirmations {
		at := row.CreatedAt.Time
		item := TrustItem{Status: string(row.Result), ConfirmedAt: &at, Source: "community"}
		if row.Kind == dbgen.ConfirmationKindExists {
			p.Trust.Exists = item
		} else {
			p.Trust.OpeningHours = item
		}
	}
	return p, nil
}

func (r *PostgresRepository) Materialize(ctx context.Context, items []Place) ([]Place, error) {
	out := make([]Place, 0, len(items))
	for _, p := range items {
		parts := strings.SplitN(p.ID, ":", 2)
		if len(parts) != 2 || parts[0] != "amap" {
			return nil, errors.New("invalid provider place id")
		}
		if p.Location.System != geo.GCJ02 || p.Location.Validate() != nil {
			return nil, errors.New("invalid provider place coordinate")
		}
		wgs := geo.GCJ02ToWGS84(p.Location)
		sourceUpdatedAt := providerSourceUpdatedAt(p)
		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return nil, err
		}
		q := r.queries.WithTx(tx)
		if err = q.AdvisoryLock(ctx, "amap:"+parts[1]); err != nil {
			_ = tx.Rollback(ctx)
			return nil, err
		}
		id, err := q.FindPlaceSource(ctx, dbgen.FindPlaceSourceParams{Provider: "amap", ExternalID: parts[1]})
		if errors.Is(err, pgx.ErrNoRows) {
			id, err = q.CreateMaterializedPlace(ctx, dbgen.CreateMaterializedPlaceParams{Name: p.Name, CategoryID: materializedCategory(p.Category), Address: text(p.Address), CityCode: text(p.CityCode), DistrictCode: text(p.DistrictCode), Wgs84Longitude: wgs.Longitude, Wgs84Latitude: wgs.Latitude, Gcj02Latitude: p.Location.Latitude, Gcj02Longitude: p.Location.Longitude, SourceUpdatedAt: sourceUpdatedAt})
			if err == nil {
				err = q.CreatePlaceSource(ctx, dbgen.CreatePlaceSourceParams{PlaceID: id, Provider: "amap", ExternalID: parts[1], RawCoordinateSystem: dbgen.CoordinateSystemGcj02})
			}
		} else if err == nil {
			err = q.UpdateMaterializedPlace(ctx, dbgen.UpdateMaterializedPlaceParams{ID: id, Name: p.Name, CategoryID: materializedCategory(p.Category), Address: text(p.Address), CityCode: text(p.CityCode), DistrictCode: text(p.DistrictCode), Wgs84Longitude: wgs.Longitude, Wgs84Latitude: wgs.Latitude, Gcj02Latitude: p.Location.Latitude, Gcj02Longitude: p.Location.Longitude, SourceUpdatedAt: sourceUpdatedAt})
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return nil, err
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}
		p.ID = uuidString(id)
		out = append(out, p)
	}
	return out, nil
}

func (r *PostgresRepository) AddConfirmation(ctx context.Context, c Confirmation) (Confirmation, error) {
	pid, err := parseUUID(c.PlaceID)
	if err != nil {
		return Confirmation{}, httpx.ErrNotFound
	}
	uid, err := parseUUID(c.UserID)
	if err != nil {
		return Confirmation{}, errors.New("invalid user")
	}
	sid, err := parseUUID(c.SessionID)
	if err != nil {
		return Confirmation{}, ErrWriteNotAllowed
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Confirmation{}, err
	}
	defer tx.Rollback(ctx)
	q := r.queries.WithTx(tx)
	if _, err = q.LockActiveConfirmationActor(ctx, dbgen.LockActiveConfirmationActorParams{SessionID: sid, UserID: uid}); errors.Is(err, pgx.ErrNoRows) {
		return Confirmation{}, ErrWriteNotAllowed
	} else if err != nil {
		return Confirmation{}, err
	}
	exists, err := q.PlaceExists(ctx, pid)
	if err != nil {
		return Confirmation{}, err
	}
	if !exists {
		return Confirmation{}, httpx.ErrNotFound
	}
	args := dbgen.InsertPlaceConfirmationParams{PlaceID: pid, UserID: uid, Kind: dbgen.ConfirmationKind(c.Kind), Result: dbgen.ConfirmationResult(c.Result), ClientIdempotencyKey: c.IdempotencyKey}
	row, err := q.InsertPlaceConfirmation(ctx, args)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, e := q.FindPlaceConfirmationByKey(ctx, dbgen.FindPlaceConfirmationByKeyParams{UserID: uid, ClientIdempotencyKey: c.IdempotencyKey})
		if e != nil {
			return Confirmation{}, e
		}
		actual := confirmation(existing.ID, existing.PlaceID, existing.UserID, existing.Kind, existing.Result, existing.CreatedAt, c.IdempotencyKey)
		if actual.PlaceID != c.PlaceID || actual.Kind != c.Kind || actual.Result != c.Result {
			return Confirmation{}, ErrIdempotencyConflict
		}
		if err = tx.Commit(ctx); err != nil {
			return Confirmation{}, err
		}
		return actual, nil
	}
	if err != nil {
		return Confirmation{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Confirmation{}, err
	}
	return confirmation(row.ID, row.PlaceID, row.UserID, row.Kind, row.Result, row.CreatedAt, c.IdempotencyKey), nil
}

func (r *PostgresRepository) AddFeedback(ctx context.Context, feedback Feedback) (Feedback, error) {
	pid, err := parseUUID(feedback.PlaceID)
	if err != nil {
		return Feedback{}, httpx.ErrNotFound
	}
	uid, err := parseUUID(feedback.UserID)
	if err != nil {
		return Feedback{}, errors.New("invalid user")
	}
	sid, err := parseUUID(feedback.SessionID)
	if err != nil {
		return Feedback{}, ErrWriteNotAllowed
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Feedback{}, err
	}
	defer tx.Rollback(ctx)
	q := r.queries.WithTx(tx)
	if _, err = q.LockActivePlaceFeedbackActor(ctx, dbgen.LockActivePlaceFeedbackActorParams{SessionID: sid, UserID: uid}); errors.Is(err, pgx.ErrNoRows) {
		return Feedback{}, ErrWriteNotAllowed
	} else if err != nil {
		return Feedback{}, err
	}
	exists, err := q.PlaceExists(ctx, pid)
	if err != nil {
		return Feedback{}, err
	}
	if !exists {
		return Feedback{}, httpx.ErrNotFound
	}
	args := dbgen.InsertPlaceFeedbackParams{PlaceID: pid, UserID: uid, Kind: dbgen.PlaceFeedbackKind(feedback.Kind), Details: feedback.Details, ClientIdempotencyKey: feedback.IdempotencyKey}
	row, err := q.InsertPlaceFeedback(ctx, args)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, findErr := q.FindPlaceFeedbackByKey(ctx, dbgen.FindPlaceFeedbackByKeyParams{UserID: uid, ClientIdempotencyKey: feedback.IdempotencyKey})
		if findErr != nil {
			return Feedback{}, findErr
		}
		actual := placeFeedback(existing.ID, existing.PlaceID, existing.UserID, existing.Kind, existing.Details, existing.CreatedAt, feedback.IdempotencyKey)
		if actual.PlaceID != feedback.PlaceID || actual.Kind != feedback.Kind || actual.Details != feedback.Details {
			return Feedback{}, ErrIdempotencyConflict
		}
		if err = tx.Commit(ctx); err != nil {
			return Feedback{}, err
		}
		return actual, nil
	}
	if err != nil {
		return Feedback{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Feedback{}, err
	}
	return placeFeedback(row.ID, row.PlaceID, row.UserID, row.Kind, row.Details, row.CreatedAt, feedback.IdempotencyKey), nil
}

var ErrWriteNotAllowed = errors.New("confirmation write is not allowed")

func text(v string) pgtype.Text { return pgtype.Text{String: v, Valid: v != ""} }
func materializedCategory(category Category) string {
	if category == CategoryUnclassified {
		return ""
	}
	return string(category)
}
func storedCategory(category pgtype.Text) Category {
	if !category.Valid || category.String == "" {
		return CategoryUnclassified
	}
	return Category(category.String)
}
func parseUUID(v string) (pgtype.UUID, error) {
	id, err := uuid.Parse(v)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}
func uuidString(v pgtype.UUID) string {
	if !v.Valid {
		return ""
	}
	return uuid.UUID(v.Bytes).String()
}
func timePtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}
func nearbyTrustItem(result dbgen.ConfirmationResult, confirmedAt pgtype.Timestamptz) TrustItem {
	if !confirmedAt.Valid {
		return TrustItem{Status: "unknown"}
	}
	at := confirmedAt.Time
	return TrustItem{Status: string(result), ConfirmedAt: &at, Source: "community"}
}
func providerSourceUpdatedAt(place Place) pgtype.Timestamptz {
	if place.SourceUpdatedAt != nil {
		return pgtype.Timestamptz{Time: place.SourceUpdatedAt.UTC(), Valid: true}
	}
	return pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
}
func confirmation(id, pid, uid pgtype.UUID, kind dbgen.ConfirmationKind, result dbgen.ConfirmationResult, created pgtype.Timestamptz, key string) Confirmation {
	return Confirmation{ID: uuidString(id), PlaceID: uuidString(pid), UserID: uuidString(uid), Kind: string(kind), Result: string(result), IdempotencyKey: key, CreatedAt: created.Time}
}
func placeFeedback(id, pid, uid pgtype.UUID, kind dbgen.PlaceFeedbackKind, details string, created pgtype.Timestamptz, key string) Feedback {
	return Feedback{ID: uuidString(id), PlaceID: uuidString(pid), UserID: uuidString(uid), Kind: string(kind), Details: details, IdempotencyKey: key, CreatedAt: created.Time}
}

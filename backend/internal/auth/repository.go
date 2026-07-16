package auth

import (
	"alongtu/backend/internal/platform/database/dbgen"
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindOrCreateUser(context.Context, string) (string, error)
	CreateSession(context.Context, string, string, []byte, string, time.Time) error
	IsActiveSession(context.Context, string, string) (bool, error)
	RotateSession(context.Context, string, []byte, func(string) (Tokens, []byte, error)) (Tokens, error)
	RevokeSession(context.Context, string) (bool, error)
}
type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *dbgen.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool, queries: dbgen.New(pool)}
}
func (r *PostgresRepository) FindOrCreateUser(ctx context.Context, phone string) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	q := r.queries.WithTx(tx)
	if err = q.AdvisoryLock(ctx, "phone:"+phone); err != nil {
		return "", err
	}
	row, err := q.FindUserByPhone(ctx, phone)
	var id pgtype.UUID
	if errors.Is(err, pgx.ErrNoRows) {
		created, e := q.CreatePhoneUser(ctx, phone)
		err = e
		id = created.ID
	} else {
		if err == nil && row.Status != "active" {
			return "", ErrUserInactive
		}
		id = row.ID
	}
	if err != nil {
		return "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return authUUIDString(id), nil
}
func (r *PostgresRepository) CreateSession(ctx context.Context, id, userID string, hash []byte, deviceID string, expires time.Time) error {
	sid, err := authUUID(id)
	if err != nil {
		return err
	}
	uid, err := authUUID(userID)
	if err != nil {
		return err
	}
	_, err = r.queries.CreateDeviceSession(ctx, dbgen.CreateDeviceSessionParams{ID: sid, UserID: uid, RefreshTokenHash: hash, DeviceID: deviceID, ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true}})
	return err
}
func (r *PostgresRepository) IsActiveSession(ctx context.Context, id, userID string) (bool, error) {
	sid, err := authUUID(id)
	if err != nil {
		return false, err
	}
	uid, err := authUUID(userID)
	if err != nil {
		return false, err
	}
	return r.queries.IsActiveSession(ctx, dbgen.IsActiveSessionParams{ID: sid, UserID: uid})
}
func (r *PostgresRepository) RotateSession(ctx context.Context, id string, presented []byte, issue func(string) (Tokens, []byte, error)) (Tokens, error) {
	sid, err := authUUID(id)
	if err != nil {
		return Tokens{}, ErrInvalidRefresh
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Tokens{}, err
	}
	defer tx.Rollback(ctx)
	q := r.queries.WithTx(tx)
	row, err := q.LockActiveDeviceSession(ctx, sid)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tokens{}, ErrInvalidRefresh
	}
	if err != nil {
		return Tokens{}, err
	}
	if time.Now().UTC().After(row.ExpiresAt.Time) || len(row.RefreshTokenHash) != len(presented) || subtle.ConstantTimeCompare(row.RefreshTokenHash, presented) != 1 {
		_, _ = q.RevokeDeviceSession(ctx, sid)
		_ = tx.Commit(ctx)
		return Tokens{}, ErrInvalidRefresh
	}
	tokens, newHash, err := issue(authUUIDString(row.UserID))
	if err != nil {
		return Tokens{}, err
	}
	if err = q.RotateDeviceSession(ctx, dbgen.RotateDeviceSessionParams{ID: sid, RefreshTokenHash: newHash, ExpiresAt: pgtype.Timestamptz{Time: tokens.RefreshExpiresAt, Valid: true}}); err != nil {
		return Tokens{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Tokens{}, err
	}
	return tokens, nil
}
func (r *PostgresRepository) RevokeSession(ctx context.Context, id string) (bool, error) {
	sid, err := authUUID(id)
	if err != nil {
		return false, err
	}
	rows, err := r.queries.RevokeDeviceSession(ctx, sid)
	return rows > 0, err
}

var ErrInvalidRefresh = errors.New("invalid refresh token")
var ErrUserInactive = errors.New("user is not active")

func authUUID(v string) (pgtype.UUID, error) {
	id, err := uuid.Parse(v)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}
func authUUIDString(v pgtype.UUID) string {
	if !v.Valid {
		return ""
	}
	return uuid.UUID(v.Bytes).String()
}

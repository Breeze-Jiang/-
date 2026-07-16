-- name: AdvisoryLock :exec
SELECT pg_advisory_xact_lock(hashtext(sqlc.arg(lock_key)));

-- name: FindUserByPhone :one
SELECT u.id, u.status, u.created_at
FROM users u JOIN auth_identities i ON i.user_id=u.id
WHERE i.provider='phone' AND i.provider_subject=$1;

-- name: CreatePhoneUser :one
WITH new_user AS (INSERT INTO users DEFAULT VALUES RETURNING id, status, created_at),
new_identity AS (INSERT INTO auth_identities(user_id, provider, provider_subject) SELECT id, 'phone', $1 FROM new_user)
SELECT id, status, created_at FROM new_user;

-- name: CreateDeviceSession :one
INSERT INTO device_sessions(id, user_id, refresh_token_hash, device_id, expires_at)
VALUES ($1,$2,$3,$4,$5) RETURNING id, user_id, expires_at, created_at;

-- name: IsActiveSession :one
SELECT EXISTS(
  SELECT 1 FROM device_sessions s JOIN users u ON u.id=s.user_id
  WHERE s.id=$1 AND s.user_id=$2 AND s.revoked_at IS NULL
    AND s.expires_at>now() AND u.status='active'
);

-- name: LockActiveDeviceSession :one
SELECT s.id, s.user_id, s.refresh_token_hash, s.device_id, s.expires_at
FROM device_sessions s JOIN users u ON u.id=s.user_id
WHERE s.id=$1 AND s.revoked_at IS NULL AND u.status='active'
FOR UPDATE;

-- name: RotateDeviceSession :exec
UPDATE device_sessions SET refresh_token_hash=$2, rotated_at=now(), expires_at=$3
WHERE id=$1 AND revoked_at IS NULL;

-- name: RevokeDeviceSession :execrows
UPDATE device_sessions SET revoked_at=now() WHERE id=$1 AND revoked_at IS NULL;

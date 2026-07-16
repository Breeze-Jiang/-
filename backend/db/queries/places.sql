-- name: GetPlace :one
SELECT id, name, category_id, address, city_code, district_code,
       ST_Y(location_wgs84::geometry)::double precision AS wgs84_latitude,
       ST_X(location_wgs84::geometry)::double precision AS wgs84_longitude,
       gcj02_latitude, gcj02_longitude, source_updated_at, created_at, updated_at
FROM places WHERE id = $1;

-- name: PlaceExists :one
SELECT EXISTS(SELECT 1 FROM places WHERE id = $1);

-- name: NearbyPlaces :many
SELECT id, name, category_id, address, city_code, district_code,
       ST_Y(location_wgs84::geometry)::double precision AS wgs84_latitude,
       ST_X(location_wgs84::geometry)::double precision AS wgs84_longitude,
       gcj02_latitude, gcj02_longitude, source_updated_at,
       COALESCE(exists_confirmation.result, 'unknown'::confirmation_result) AS exists_result, exists_confirmation.created_at AS exists_confirmed_at,
       COALESCE(opening_hours_confirmation.result, 'unknown'::confirmation_result) AS opening_hours_result, opening_hours_confirmation.created_at AS opening_hours_confirmed_at,
       ST_Distance(location_wgs84, ST_SetSRID(ST_MakePoint(sqlc.arg(longitude)::double precision, sqlc.arg(latitude)::double precision), 4326)::geography)::double precision AS distance_meters
FROM places
LEFT JOIN LATERAL (
  SELECT result, created_at FROM place_confirmations
  WHERE place_id = places.id AND kind = 'exists'
  ORDER BY created_at DESC, id DESC LIMIT 1
) AS exists_confirmation ON true
LEFT JOIN LATERAL (
  SELECT result, created_at FROM place_confirmations
  WHERE place_id = places.id AND kind = 'opening_hours'
  ORDER BY created_at DESC, id DESC LIMIT 1
) AS opening_hours_confirmation ON true
WHERE (sqlc.narg(category_id)::text IS NULL OR category_id = sqlc.narg(category_id))
  AND ST_DWithin(location_wgs84, ST_SetSRID(ST_MakePoint(sqlc.arg(longitude)::double precision, sqlc.arg(latitude)::double precision), 4326)::geography, sqlc.arg(radius_meters)::double precision)
ORDER BY distance_meters, id
LIMIT sqlc.arg(page_size);

-- name: LatestPlaceConfirmations :many
SELECT DISTINCT ON (kind) kind, result, created_at
FROM place_confirmations WHERE place_id = $1
ORDER BY kind, created_at DESC, id DESC;

-- name: FindPlaceSource :one
SELECT place_id FROM place_sources
WHERE provider=$1 AND external_id=$2;

-- name: CreateMaterializedPlace :one
INSERT INTO places(name,category_id,address,city_code,district_code,location_wgs84,gcj02_latitude,gcj02_longitude,source_updated_at)
VALUES(sqlc.arg(name),NULLIF(sqlc.arg(category_id)::text,''),sqlc.arg(address),sqlc.arg(city_code),sqlc.arg(district_code),ST_SetSRID(ST_MakePoint(sqlc.arg(wgs84_longitude)::double precision,sqlc.arg(wgs84_latitude)::double precision),4326)::geography,sqlc.arg(gcj02_latitude),sqlc.arg(gcj02_longitude),sqlc.arg(source_updated_at)::timestamptz)
RETURNING id;

-- name: CreatePlaceSource :exec
INSERT INTO place_sources(place_id,provider,external_id,raw_coordinate_system,fetched_at)
VALUES($1,$2,$3,$4,now());

-- name: UpdateMaterializedPlace :exec
UPDATE places SET name=sqlc.arg(name),category_id=NULLIF(sqlc.arg(category_id)::text,''),address=sqlc.arg(address),city_code=sqlc.arg(city_code),district_code=sqlc.arg(district_code),
  location_wgs84=ST_SetSRID(ST_MakePoint(sqlc.arg(wgs84_longitude)::double precision,sqlc.arg(wgs84_latitude)::double precision),4326)::geography,
  gcj02_latitude=sqlc.arg(gcj02_latitude),gcj02_longitude=sqlc.arg(gcj02_longitude),source_updated_at=sqlc.arg(source_updated_at)::timestamptz,updated_at=now()
WHERE id=sqlc.arg(id)
  AND COALESCE(source_updated_at,'epoch'::timestamptz) <= sqlc.arg(source_updated_at)::timestamptz;

-- name: InsertPlaceConfirmation :one
INSERT INTO place_confirmations(place_id,user_id,kind,result,client_idempotency_key)
VALUES($1,$2,$3,$4,$5)
ON CONFLICT(user_id,client_idempotency_key) DO NOTHING
RETURNING id,place_id,user_id,kind,result,created_at;

-- name: FindPlaceConfirmationByKey :one
SELECT id,place_id,user_id,kind,result,created_at
FROM place_confirmations WHERE user_id=$1 AND client_idempotency_key=$2;

-- name: InsertPlaceFeedback :one
INSERT INTO place_feedbacks(place_id,user_id,kind,details,client_idempotency_key)
VALUES($1,$2,$3,$4,$5)
ON CONFLICT(user_id,client_idempotency_key) DO NOTHING
RETURNING id,place_id,user_id,kind,details,created_at;

-- name: FindPlaceFeedbackByKey :one
SELECT id,place_id,user_id,kind,details,created_at
FROM place_feedbacks WHERE user_id=$1 AND client_idempotency_key=$2;

-- name: LockActiveConfirmationActor :one
SELECT true AS allowed
FROM device_sessions s
JOIN users u ON u.id = s.user_id
WHERE s.id = sqlc.arg(session_id)
  AND s.user_id = sqlc.arg(user_id)
  AND s.revoked_at IS NULL
  AND s.expires_at > now()
  AND u.status = 'active'
FOR SHARE OF s, u;

-- name: LockActivePlaceFeedbackActor :one
SELECT true AS allowed
FROM device_sessions s
JOIN users u ON u.id = s.user_id
WHERE s.id = sqlc.arg(session_id)
  AND s.user_id = sqlc.arg(user_id)
  AND s.revoked_at IS NULL
  AND s.expires_at > now()
  AND u.status = 'active'
FOR SHARE OF s, u;

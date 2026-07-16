-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE coordinate_system AS ENUM ('wgs84', 'gcj02');
CREATE TYPE confirmation_kind AS ENUM ('exists', 'opening_hours');
CREATE TYPE confirmation_result AS ENUM ('confirmed', 'incorrect', 'unknown');

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','blocked','deleted')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE auth_identities (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id),
  provider text NOT NULL CHECK (provider IN ('phone')),
  provider_subject text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (provider, provider_subject)
);

CREATE TABLE device_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id),
  refresh_token_hash bytea NOT NULL,
  device_id text NOT NULL CHECK (char_length(device_id) BETWEEN 1 AND 200),
  expires_at timestamptz NOT NULL,
  rotated_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (expires_at > created_at)
);
CREATE INDEX device_sessions_user_active_idx ON device_sessions(user_id, expires_at) WHERE revoked_at IS NULL;

CREATE TABLE place_categories (
  id text PRIMARY KEY,
  display_name text NOT NULL,
  parent_id text REFERENCES place_categories(id),
  sort_order integer NOT NULL DEFAULT 0,
  enabled boolean NOT NULL DEFAULT true
);

CREATE TABLE places (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
  category_id text REFERENCES place_categories(id),
  address text,
  city_code text,
  district_code text,
  location_wgs84 geography(Point, 4326) NOT NULL,
  gcj02_latitude double precision NOT NULL CHECK (gcj02_latitude BETWEEN -90 AND 90),
  gcj02_longitude double precision NOT NULL CHECK (gcj02_longitude BETWEEN -180 AND 180),
  source_updated_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX places_location_gist_idx ON places USING gist(location_wgs84);
CREATE INDEX places_city_category_idx ON places(city_code, category_id);

CREATE TABLE place_sources (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  place_id uuid NOT NULL REFERENCES places(id) ON DELETE CASCADE,
  provider text NOT NULL,
  external_id text NOT NULL,
  raw_coordinate_system coordinate_system NOT NULL,
  fetched_at timestamptz NOT NULL,
  etag text,
  UNIQUE(provider, external_id)
);

CREATE TABLE place_opening_hours (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  place_id uuid NOT NULL REFERENCES places(id) ON DELETE CASCADE,
  weekday smallint NOT NULL CHECK (weekday BETWEEN 1 AND 7),
  opens_at time,
  closes_at time,
  status text NOT NULL DEFAULT 'known' CHECK (status IN ('known','closed','unknown')),
  valid_from date,
  valid_until date,
  CHECK ((status = 'known' AND opens_at IS NOT NULL AND closes_at IS NOT NULL) OR status <> 'known')
);

CREATE TABLE place_confirmations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  place_id uuid NOT NULL REFERENCES places(id),
  user_id uuid NOT NULL REFERENCES users(id),
  kind confirmation_kind NOT NULL,
  result confirmation_result NOT NULL,
  client_idempotency_key text NOT NULL CHECK (char_length(client_idempotency_key) BETWEEN 8 AND 200),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(user_id, client_idempotency_key)
);
CREATE INDEX place_confirmations_latest_idx ON place_confirmations(place_id, kind, created_at DESC);

CREATE TABLE provider_cache_metadata (
  cache_key text PRIMARY KEY,
  provider text NOT NULL,
  fetched_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  stale_until timestamptz NOT NULL,
  CHECK (fetched_at <= expires_at AND expires_at <= stale_until)
);

INSERT INTO place_categories(id, display_name, sort_order) VALUES
('toilet','厕所',10),('fuel','加油站',20),('charging','充电站',30),
('supermarket','超市',40),('attraction','景点',50),('hotel','酒店',60),
('parking','停车场',70),('power_bank','充电宝',80),('pharmacy','药店',90),
('hospital','医院',100),('convenience_store','便利店',110),('police','警务点',120),('food','美食',130);

-- +goose Down
DROP TABLE IF EXISTS provider_cache_metadata, place_confirmations, place_opening_hours, place_sources, places, place_categories, device_sessions, auth_identities, users CASCADE;
DROP TYPE IF EXISTS confirmation_result, confirmation_kind, coordinate_system;

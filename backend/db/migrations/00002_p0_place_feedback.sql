-- +goose Up
CREATE TYPE place_feedback_kind AS ENUM ('entrance', 'incorrect_info');

CREATE TABLE place_feedbacks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  place_id uuid NOT NULL REFERENCES places(id),
  user_id uuid NOT NULL REFERENCES users(id),
  kind place_feedback_kind NOT NULL,
  details text NOT NULL CHECK (char_length(details) BETWEEN 1 AND 500),
  client_idempotency_key text NOT NULL CHECK (char_length(client_idempotency_key) BETWEEN 8 AND 200),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(user_id, client_idempotency_key)
);
CREATE INDEX place_feedbacks_place_created_idx ON place_feedbacks(place_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS place_feedbacks;
DROP TYPE IF EXISTS place_feedback_kind;

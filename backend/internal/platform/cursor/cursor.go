package cursor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"
)

// Position is the provider position represented by an opaque cursor. The
// optional boundary is an internal place ID plus its distance when ordering by
// distance is available. Neither field is exposed through the public API.
type Position struct {
	Page               int
	LastDistanceMeters *float64
	LastID             string
}

type payload struct {
	Kind               string   `json:"kind"`
	Fingerprint        string   `json:"fingerprint"`
	Page               int      `json:"page"`
	LastDistanceMeters *float64 `json:"lastDistanceMeters,omitempty"`
	LastID             string   `json:"lastId,omitempty"`
	ExpiresAt          int64    `json:"expiresAt"`
}
type Codec struct {
	secret []byte
	ttl    time.Duration
}

var ErrInvalid = errors.New("invalid cursor")

func New(secret string, ttl time.Duration) *Codec { return &Codec{secret: []byte(secret), ttl: ttl} }
func (c *Codec) Encode(kind, fingerprint string, position Position) string {
	if !validPosition(position) {
		return ""
	}
	raw, _ := json.Marshal(payload{Kind: kind, Fingerprint: fingerprint, Page: position.Page, LastDistanceMeters: position.LastDistanceMeters, LastID: position.LastID, ExpiresAt: time.Now().Add(c.ttl).Unix()})
	body := base64.RawURLEncoding.EncodeToString(raw)
	m := hmac.New(sha256.New, c.secret)
	_, _ = m.Write([]byte(body))
	return body + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
func (c *Codec) Decode(value, kind, fingerprint string) (Position, error) {
	if value == "" {
		return Position{Page: 1}, nil
	}
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return Position{}, ErrInvalid
	}
	m := hmac.New(sha256.New, c.secret)
	_, _ = m.Write([]byte(parts[0]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(sig, m.Sum(nil)) {
		return Position{}, ErrInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Position{}, ErrInvalid
	}
	var p payload
	if json.Unmarshal(raw, &p) != nil {
		return Position{}, ErrInvalid
	}
	position := Position{Page: p.Page, LastDistanceMeters: p.LastDistanceMeters, LastID: p.LastID}
	if p.Kind != kind || p.Fingerprint != fingerprint || !validPosition(position) || time.Now().Unix() > p.ExpiresAt {
		return Position{}, ErrInvalid
	}
	return position, nil
}

func validPosition(position Position) bool {
	if position.Page < 1 {
		return false
	}
	if position.LastDistanceMeters == nil {
		return true
	}
	return position.LastID != "" && *position.LastDistanceMeters >= 0 && !math.IsNaN(*position.LastDistanceMeters) && !math.IsInf(*position.LastDistanceMeters, 0)
}
func Fingerprint(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(p))
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)[:12])
}

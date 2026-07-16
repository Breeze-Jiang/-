package cursor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"
)

func TestRoundTripAndTamper(t *testing.T) {
	c := New("a-long-test-secret", time.Hour)
	distance := 123.5
	value := c.Encode("nearby", "query-a", Position{Page: 2, LastDistanceMeters: &distance, LastID: "place-2"})
	position, err := c.Decode(value, "nearby", "query-a")
	if err != nil || position.Page != 2 || position.LastDistanceMeters == nil || *position.LastDistanceMeters != distance || position.LastID != "place-2" {
		t.Fatalf("round trip failed: %#v %v", position, err)
	}
	if _, err = c.Decode(value+"x", "nearby", "query-a"); err == nil {
		t.Fatal("tampered cursor accepted")
	}
	if _, err = c.Decode(value, "search", "query-a"); err == nil {
		t.Fatal("cross-query cursor accepted")
	}
}
func TestExpired(t *testing.T) {
	c := New("a-long-test-secret", -time.Second)
	value := c.Encode("nearby", "query", Position{Page: 2})
	if _, err := c.Decode(value, "nearby", "query"); err == nil {
		t.Fatal("expired cursor accepted")
	}
}

func TestRejectsPartialOrInvalidPosition(t *testing.T) {
	c := New("a-long-test-secret", time.Hour)
	for _, position := range []Position{
		{Page: 2, LastDistanceMeters: float64Pointer(-1), LastID: "place-2"},
		{Page: 2, LastDistanceMeters: float64Pointer(1)},
	} {
		if value := c.Encode("nearby", "query", position); value != "" {
			t.Fatalf("invalid position unexpectedly encoded: %#v", position)
		}
	}
}

func TestLegacyPageOnlyCursorRemainsValid(t *testing.T) {
	secret := "a-long-test-secret"
	legacyJSON := []byte(`{"kind":"nearby","fingerprint":"query-a","page":2,"expiresAt":4102444800}`)
	body := base64.RawURLEncoding.EncodeToString(legacyJSON)
	signature := hmac.New(sha256.New, []byte(secret))
	_, _ = signature.Write([]byte(body))
	legacyCursor := body + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil))

	position, err := New(secret, time.Hour).Decode(legacyCursor, "nearby", "query-a")
	if err != nil || position.Page != 2 || position.LastID != "" || position.LastDistanceMeters != nil {
		t.Fatalf("legacy cursor was not accepted as a page-only position: %#v %v", position, err)
	}
}

func float64Pointer(value float64) *float64 { return &value }

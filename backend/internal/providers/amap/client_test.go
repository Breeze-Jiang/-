package amap

import (
	"encoding/json"
	"testing"

	"alongtu/backend/internal/places"
)

func TestMapCategoryPreservesUnknownTypeAsUnclassified(t *testing.T) {
	if got := mapCategory("999999"); got != places.CategoryUnclassified {
		t.Fatalf("unknown provider type = %q, want %q", got, places.CategoryUnclassified)
	}
}

func TestParseDistanceDoesNotInventUnknownValues(t *testing.T) {
	for _, raw := range []json.RawMessage{json.RawMessage(`"invalid"`), json.RawMessage(`-1`), json.RawMessage(`null`)} {
		if got := parseDistance(raw); got != nil {
			t.Fatalf("invalid provider distance %s became %v", raw, *got)
		}
	}
	if got := parseDistance(json.RawMessage(`320`)); got == nil || *got != 320 {
		t.Fatalf("numeric provider distance = %v", got)
	}
}

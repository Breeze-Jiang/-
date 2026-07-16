package geo

import (
	"math"
	"testing"
)

func TestCoordinateValidation(t *testing.T) {
	if (Coordinate{Latitude: 91, Longitude: 1, System: WGS84}).Validate() == nil {
		t.Fatal("expected invalid latitude")
	}
}
func TestCoordinateValidationRejectsNonFiniteValues(t *testing.T) {
	for _, coordinate := range []Coordinate{
		{Latitude: math.NaN(), Longitude: 116.4, System: WGS84},
		{Latitude: 39.9, Longitude: math.Inf(1), System: WGS84},
	} {
		if coordinate.Validate() == nil {
			t.Fatalf("accepted non-finite coordinate: %#v", coordinate)
		}
	}
}
func TestDistance(t *testing.T) {
	d := DistanceMeters(Coordinate{Latitude: 39.9, Longitude: 116.4, System: WGS84}, Coordinate{Latitude: 39.91, Longitude: 116.4, System: WGS84})
	if d < 1100 || d > 1125 {
		t.Fatalf("unexpected distance %f", d)
	}
}
func TestGCJRoundTrip(t *testing.T) {
	w := Coordinate{Latitude: 39.908823, Longitude: 116.397470, System: WGS84}
	got := GCJ02ToWGS84(WGS84ToGCJ02(w))
	if DistanceMeters(w, got) > 1 {
		t.Fatalf("round trip drift: %#v", got)
	}
}

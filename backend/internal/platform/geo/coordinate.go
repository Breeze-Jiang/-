package geo

import (
	"errors"
	"math"
)

type System string

const (
	WGS84 System = "wgs84"
	GCJ02 System = "gcj02"
)

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	System    System  `json:"coordinateSystem"`
}

func (c Coordinate) Validate() error {
	if math.IsNaN(c.Latitude) || math.IsInf(c.Latitude, 0) || math.IsNaN(c.Longitude) || math.IsInf(c.Longitude, 0) || c.Latitude < -90 || c.Latitude > 90 || c.Longitude < -180 || c.Longitude > 180 {
		return errors.New("coordinate out of range")
	}
	if c.System != WGS84 && c.System != GCJ02 {
		return errors.New("unsupported coordinate system")
	}
	return nil
}
func DistanceMeters(a, b Coordinate) float64 {
	const radius = 6371008.8
	p1, p2 := a.Latitude*math.Pi/180, b.Latitude*math.Pi/180
	dp := (b.Latitude - a.Latitude) * math.Pi / 180
	dl := (b.Longitude - a.Longitude) * math.Pi / 180
	h := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	return 2 * radius * math.Asin(math.Sqrt(h))
}

// WGS84ToGCJ02 converts GPS coordinates for mainland-China map providers.
// Coordinates outside mainland China are returned unchanged.
func WGS84ToGCJ02(c Coordinate) Coordinate {
	if outsideChina(c.Latitude, c.Longitude) {
		c.System = GCJ02
		return c
	}
	dLat, dLng := delta(c.Latitude, c.Longitude)
	return Coordinate{Latitude: c.Latitude + dLat, Longitude: c.Longitude + dLng, System: GCJ02}
}

// GCJ02ToWGS84 uses iterative inversion so PostGIS never stores GCJ-02 as EPSG:4326.
func GCJ02ToWGS84(c Coordinate) Coordinate {
	if outsideChina(c.Latitude, c.Longitude) {
		c.System = WGS84
		return c
	}
	guess := Coordinate{Latitude: c.Latitude, Longitude: c.Longitude, System: WGS84}
	for i := 0; i < 8; i++ {
		mapped := WGS84ToGCJ02(guess)
		guess.Latitude -= mapped.Latitude - c.Latitude
		guess.Longitude -= mapped.Longitude - c.Longitude
	}
	return guess
}
func outsideChina(lat, lng float64) bool {
	return lng < 72.004 || lng > 137.8347 || lat < 0.8293 || lat > 55.8271
}
func delta(lat, lng float64) (float64, float64) {
	const a = 6378245.0
	const ee = 0.00669342162296594323
	x := lng - 105
	y := lat - 35
	dLat := -100 + 2*x + 3*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	dLat += (20*math.Sin(6*x*math.Pi) + 20*math.Sin(2*x*math.Pi)) * 2 / 3
	dLat += (20*math.Sin(y*math.Pi) + 40*math.Sin(y/3*math.Pi)) * 2 / 3
	dLat += (160*math.Sin(y/12*math.Pi) + 320*math.Sin(y*math.Pi/30)) * 2 / 3
	dLng := 300 + x + 2*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	dLng += (20*math.Sin(6*x*math.Pi) + 20*math.Sin(2*x*math.Pi)) * 2 / 3
	dLng += (20*math.Sin(x*math.Pi) + 40*math.Sin(x/3*math.Pi)) * 2 / 3
	dLng += (150*math.Sin(x/12*math.Pi) + 300*math.Sin(x/30*math.Pi)) * 2 / 3
	rad := lat / 180 * math.Pi
	magic := math.Sin(rad)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	return dLat * 180 / ((a * (1 - ee)) / (magic * sqrtMagic) * math.Pi), dLng * 180 / (a / sqrtMagic * math.Cos(rad) * math.Pi)
}

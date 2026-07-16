package amap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"alongtu/backend/internal/navigation"
	"alongtu/backend/internal/places"
	"alongtu/backend/internal/platform/geo"
)

type Client struct {
	baseURL, key string
	http         *http.Client
}
type providerError struct {
	message   string
	temporary bool
}

func (e providerError) Error() string   { return e.message }
func (e providerError) Temporary() bool { return e.temporary }

func New(baseURL, key string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), key: key, http: &http.Client{Timeout: timeout}}
}

type response struct {
	Status string `json:"status"`
	Info   string `json:"info"`
	POIs   []struct {
		ID       string          `json:"id"`
		Name     string          `json:"name"`
		TypeCode string          `json:"typecode"`
		Address  any             `json:"address"`
		Location string          `json:"location"`
		Distance json.RawMessage `json:"distance"`
		CityCode string          `json:"citycode"`
		AdCode   string          `json:"adcode"`
	} `json:"pois"`
}

func (c *Client) Nearby(ctx context.Context, q places.NearbyQuery) (places.Page, error) {
	if !isCategoryCovered(q.Category) {
		return places.Page{Items: []places.Place{}, DataFreshness: "live", Coverage: "not_covered"}, nil
	}
	center := q.Center
	if center.System == geo.WGS84 {
		center = geo.WGS84ToGCJ02(center)
	}
	v := url.Values{"key": {c.key}, "location": {fmt.Sprintf("%.6f,%.6f", center.Longitude, center.Latitude)}, "radius": {strconv.Itoa(q.RadiusMeters)}, "offset": {strconv.Itoa(q.Limit)}, "page": {strconv.Itoa(q.Page)}, "sortrule": {"distance"}, "extensions": {"base"}}
	if kw := categoryKeyword(q.Category); kw != "" {
		v.Set("keywords", kw)
	}
	return c.fetch(ctx, "/v3/place/around", v, &center, float64(q.RadiusMeters))
}
func (c *Client) Search(ctx context.Context, q places.SearchQuery) (places.Page, error) {
	if !isCategoryCovered(q.Category) {
		return places.Page{Items: []places.Place{}, DataFreshness: "live", Coverage: "not_covered"}, nil
	}
	keyword := q.Keyword
	if category := categoryKeyword(q.Category); category != "" {
		keyword = strings.TrimSpace(keyword + " " + category)
	}
	v := url.Values{"key": {c.key}, "keywords": {keyword}, "offset": {strconv.Itoa(q.Limit)}, "page": {strconv.Itoa(q.Page)}, "extensions": {"base"}}
	if q.CityCode != "" {
		v.Set("city", q.CityCode)
		v.Set("citylimit", "true")
	}
	return c.fetch(ctx, "/v3/place/text", v, nil, 0)
}
func (c *Client) fetch(ctx context.Context, path string, v url.Values, nearbyCenter *geo.Coordinate, radiusMeters float64) (places.Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+v.Encode(), nil)
	if err != nil {
		return places.Page{}, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return places.Page{}, providerError{message: err.Error(), temporary: true}
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return places.Page{}, providerError{message: fmt.Sprintf("amap http status %d", res.StatusCode), temporary: res.StatusCode == 429 || res.StatusCode >= 500}
	}
	var body response
	if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
		return places.Page{}, err
	}
	if body.Status != "1" {
		return places.Page{}, providerError{message: "amap rejected request"}
	}
	fetchedAt := time.Now().UTC()
	page := places.Page{DataFreshness: "live", Coverage: "covered", Items: []places.Place{}}
	for _, p := range body.POIs {
		if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Name) == "" {
			continue
		}
		parts := strings.Split(p.Location, ",")
		if len(parts) != 2 {
			continue
		}
		lng, e1 := strconv.ParseFloat(parts[0], 64)
		lat, e2 := strconv.ParseFloat(parts[1], 64)
		if e1 != nil || e2 != nil {
			continue
		}
		location := geo.Coordinate{Latitude: lat, Longitude: lng, System: geo.GCJ02}
		if location.Validate() != nil {
			continue
		}
		distance := parseDistance(p.Distance)
		if nearbyCenter != nil {
			calculatedDistance := geo.DistanceMeters(*nearbyCenter, location)
			if calculatedDistance > radiusMeters+1 {
				continue
			}
			distance = &calculatedDistance
		}
		page.Items = append(page.Items, places.Place{ID: "amap:" + p.ID, Name: p.Name, Category: mapCategory(p.TypeCode), Address: stringValue(p.Address), CityCode: p.CityCode, DistrictCode: p.AdCode, Location: location, DistanceMeters: distance, SourceUpdatedAt: &fetchedAt, Trust: places.TrustSummary{Exists: places.TrustItem{Status: "unknown", Source: "provider"}, OpeningHours: places.TrustItem{Status: "unknown", Source: "provider"}}})
	}
	return page, nil
}
func categoryKeyword(c places.Category) string {
	return map[places.Category]string{places.CategoryToilet: "厕所", places.CategoryFuel: "加油站", places.CategoryCharging: "充电站", places.CategorySupermarket: "超市", places.CategoryAttraction: "景点", places.CategoryHotel: "酒店", places.CategoryParking: "停车场", places.CategoryPowerBank: "充电宝", places.CategoryPharmacy: "药店", places.CategoryHospital: "医院", places.CategoryConvenienceStore: "便利店", places.CategoryPolice: "派出所", places.CategoryFood: "美食"}[c]
}
func isCategoryCovered(category places.Category) bool {
	switch category {
	case "", places.CategoryToilet, places.CategoryFuel, places.CategoryCharging, places.CategorySupermarket, places.CategoryAttraction, places.CategoryHotel:
		return true
	default:
		return false
	}
}
func (c *Client) Coverage(category places.Category) string {
	if isCategoryCovered(category) {
		return "covered"
	}
	return "not_covered"
}
func mapCategory(typeCode string) places.Category {
	switch {
	case strings.HasPrefix(typeCode, "2003"):
		return places.CategoryToilet
	case strings.HasPrefix(typeCode, "0101"):
		return places.CategoryFuel
	case strings.HasPrefix(typeCode, "0111"):
		return places.CategoryCharging
	case strings.HasPrefix(typeCode, "0604"):
		return places.CategorySupermarket
	case strings.HasPrefix(typeCode, "110"):
		return places.CategoryAttraction
	case strings.HasPrefix(typeCode, "1001"):
		return places.CategoryHotel
	}
	return places.CategoryUnclassified
}
func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func parseDistance(raw json.RawMessage) *float64 {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		return parseDistanceValue(value)
	}
	var number float64
	if err := json.Unmarshal(raw, &number); err != nil {
		return nil
	}
	return parseDistanceValue(strconv.FormatFloat(number, 'f', -1, 64))
}

func parseDistanceValue(value string) *float64 {
	distance, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || distance < 0 || math.IsNaN(distance) || math.IsInf(distance, 0) {
		return nil
	}
	return &distance
}

type routeResponse struct {
	Status  string `json:"status"`
	ErrCode int    `json:"errcode"`
	Route   struct {
		Paths    []routePath `json:"paths"`
		Transits []routePath `json:"transits"`
	} `json:"route"`
	Data struct {
		Paths []routePath `json:"paths"`
	} `json:"data"`
}
type routePath struct {
	Distance string `json:"distance"`
	Duration string `json:"duration"`
	Polyline string `json:"polyline"`
	Steps    []struct {
		Instruction string `json:"instruction"`
		Distance    string `json:"distance"`
		Duration    string `json:"duration"`
		Polyline    string `json:"polyline"`
	} `json:"steps"`
}

func (c *Client) Route(ctx context.Context, r navigation.RouteRequest) (navigation.Route, error) {
	endpoint := map[navigation.Mode]string{navigation.Walking: "/v3/direction/walking", navigation.Cycling: "/v4/direction/bicycling", navigation.Driving: "/v3/direction/driving", navigation.Transit: "/v3/direction/transit/integrated"}[r.Mode]
	if endpoint == "" {
		return navigation.Route{}, errors.New("unsupported route mode")
	}
	v := url.Values{"key": {c.key}, "origin": {fmt.Sprintf("%.6f,%.6f", r.Origin.Longitude, r.Origin.Latitude)}, "destination": {fmt.Sprintf("%.6f,%.6f", r.Destination.Longitude, r.Destination.Latitude)}, "extensions": {"base"}}
	if r.Mode == navigation.Transit {
		v.Set("city", r.OriginCity)
		v.Set("cityd", r.DestinationCity)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint+"?"+v.Encode(), nil)
	if err != nil {
		return navigation.Route{}, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return navigation.Route{}, providerError{message: err.Error(), temporary: true}
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return navigation.Route{}, providerError{message: fmt.Sprintf("amap route http status %d", res.StatusCode), temporary: res.StatusCode == 429 || res.StatusCode >= 500}
	}
	var body routeResponse
	if json.NewDecoder(res.Body).Decode(&body) != nil {
		return navigation.Route{}, errors.New("invalid amap route response")
	}
	paths := body.Route.Paths
	if r.Mode == navigation.Transit {
		paths = body.Route.Transits
	}
	if len(paths) == 0 {
		paths = body.Data.Paths
	}
	if len(paths) == 0 {
		return navigation.Route{}, errors.New("route not found")
	}
	p := paths[0]
	out, err := parseRoutePath(p)
	if err != nil {
		return navigation.Route{}, err
	}
	updatedAt := time.Now().UTC()
	out.ProviderUpdatedAt = &updatedAt
	return out, nil
}
func parseRoutePath(path routePath) (navigation.Route, error) {
	distance, err := nonNegativeInt(path.Distance)
	if err != nil {
		return navigation.Route{}, errors.New("invalid route distance")
	}
	duration, err := nonNegativeInt(path.Duration)
	if err != nil {
		return navigation.Route{}, errors.New("invalid route duration")
	}
	out := navigation.Route{DistanceMeters: distance, DurationSeconds: duration, Polyline: parsePolyline(path.Polyline)}
	for _, step := range path.Steps {
		stepDistance, distanceErr := nonNegativeInt(step.Distance)
		stepDuration, durationErr := nonNegativeInt(step.Duration)
		if distanceErr != nil || durationErr != nil {
			return navigation.Route{}, errors.New("invalid route step")
		}
		out.Steps = append(out.Steps, navigation.Step{Instruction: step.Instruction, DistanceMeters: stepDistance, DurationSeconds: stepDuration, Polyline: parsePolyline(step.Polyline)})
	}
	if len(out.Polyline) == 0 {
		for _, step := range out.Steps {
			out.Polyline = append(out.Polyline, step.Polyline...)
		}
	}
	if len(out.Polyline) == 0 {
		return navigation.Route{}, errors.New("route has no valid polyline")
	}
	return out, nil
}
func nonNegativeInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, errors.New("invalid integer")
	}
	return parsed, nil
}
func parsePolyline(v string) []geo.Coordinate {
	var out []geo.Coordinate
	for _, pair := range strings.Split(v, ";") {
		parts := strings.Split(pair, ",")
		if len(parts) != 2 {
			continue
		}
		lng, e1 := strconv.ParseFloat(parts[0], 64)
		lat, e2 := strconv.ParseFloat(parts[1], 64)
		if e1 == nil && e2 == nil {
			out = append(out, geo.Coordinate{Latitude: lat, Longitude: lng, System: geo.GCJ02})
		}
	}
	return out
}

package navigation

import (
	"alongtu/backend/internal/platform/geo"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Mode string

const (
	Walking Mode = "walking"
	Cycling Mode = "cycling"
	Driving Mode = "driving"
	Transit Mode = "transit"
	Taxi    Mode = "taxi"
)

type RouteRequest struct {
	Origin          geo.Coordinate `json:"origin"`
	Destination     geo.Coordinate `json:"destination"`
	Mode            Mode           `json:"mode"`
	OriginCity      string         `json:"originCity,omitempty"`
	DestinationCity string         `json:"destinationCity,omitempty"`
}
type Step struct {
	Instruction     string           `json:"instruction"`
	DistanceMeters  int              `json:"distanceMeters"`
	DurationSeconds int              `json:"durationSeconds,omitempty"`
	Polyline        []geo.Coordinate `json:"polyline,omitempty"`
}
type Transfer struct {
	LineName      string `json:"lineName"`
	DepartureStop string `json:"departureStop"`
	ArrivalStop   string `json:"arrivalStop"`
}
type Route struct {
	DistanceMeters    int              `json:"distanceMeters"`
	DurationSeconds   int              `json:"durationSeconds"`
	Polyline          []geo.Coordinate `json:"polyline"`
	Steps             []Step           `json:"steps"`
	Transfers         []Transfer       `json:"transfers,omitempty"`
	ProviderUpdatedAt *time.Time       `json:"providerUpdatedAt,omitempty"`
	DataFreshness     string           `json:"dataFreshness"`
	ExternalFallback  []ExternalLink   `json:"externalFallback"`
}
type Provider interface {
	Route(context.Context, RouteRequest) (Route, error)
}

var ErrInvalidRequest = errors.New("invalid route request")

type Service struct {
	provider     Provider
	rideTemplate string
}

func NewService(p Provider, t string) *Service { return &Service{provider: p, rideTemplate: t} }
func (s *Service) Route(ctx context.Context, r RouteRequest) (Route, error) {
	if err := validate(r); err != nil {
		return Route{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}
	pr := r
	if pr.Origin.System == geo.WGS84 {
		pr.Origin = geo.WGS84ToGCJ02(pr.Origin)
	}
	if pr.Destination.System == geo.WGS84 {
		pr.Destination = geo.WGS84ToGCJ02(pr.Destination)
	}
	fallback, _ := ExternalLinks(pr.Origin, pr.Destination, "目的地", r.Mode, s.rideTemplate)
	if pr.Mode == Taxi {
		// P0 taxis are a third-party handoff only. We do not calculate a taxi
		// route, create an order, or expose dispatch-like state ourselves.
		return stableRoute(Route{DataFreshness: "external_only", ExternalFallback: fallback}), nil
	}
	out, err := s.provider.Route(ctx, pr)
	if err != nil {
		fallbackRoute := stableRoute(Route{ExternalFallback: fallback, DataFreshness: "external_only"})
		return fallbackRoute, err
	}
	out.ExternalFallback = fallback
	if out.DataFreshness == "" {
		out.DataFreshness = "live"
	}
	return stableRoute(out), nil
}

func stableRoute(route Route) Route {
	if route.Polyline == nil {
		route.Polyline = []geo.Coordinate{}
	}
	if route.Steps == nil {
		route.Steps = []Step{}
	}
	if route.ExternalFallback == nil {
		route.ExternalFallback = []ExternalLink{}
	}
	return route
}
func validate(r RouteRequest) error {
	if r.Origin.Validate() != nil || r.Destination.Validate() != nil {
		return errors.New("invalid coordinate")
	}
	if r.Mode != Walking && r.Mode != Cycling && r.Mode != Driving && r.Mode != Transit && r.Mode != Taxi {
		return errors.New("unsupported mode")
	}
	if r.Mode == Transit && (r.OriginCity == "" || r.DestinationCity == "") {
		return errors.New("transit requires city context")
	}
	return nil
}

type ExternalLink struct {
	Provider  string `json:"provider"`
	URI       string `json:"uri"`
	Available bool   `json:"available"`
}

func ExternalLinks(origin, destination geo.Coordinate, name string, mode Mode, rideTemplate string) ([]ExternalLink, error) {
	if origin.Validate() != nil || destination.Validate() != nil {
		return nil, errors.New("invalid coordinate")
	}
	if mode != Walking && mode != Cycling && mode != Driving && mode != Transit && mode != Taxi {
		return nil, errors.New("unsupported mode")
	}
	if origin.System == geo.WGS84 {
		origin = geo.WGS84ToGCJ02(origin)
	}
	if destination.System == geo.WGS84 {
		destination = geo.WGS84ToGCJ02(destination)
	}
	n := url.QueryEscape(name)
	from := fmt.Sprintf("%.6f,%.6f", origin.Latitude, origin.Longitude)
	to := fmt.Sprintf("%.6f,%.6f", destination.Latitude, destination.Longitude)
	if mode == Taxi {
		if rideTemplate == "" {
			return []ExternalLink{{Provider: "ride_hailing", Available: false}}, nil
		}
		uri := strings.NewReplacer("{origin}", url.QueryEscape(from), "{destination}", url.QueryEscape(to), "{name}", n).Replace(rideTemplate)
		return []ExternalLink{{Provider: "ride_hailing", URI: uri, Available: true}}, nil
	}
	am := map[Mode]string{Walking: "2", Cycling: "3", Driving: "0", Transit: "1"}[mode]
	bm := map[Mode]string{Walking: "walking", Cycling: "riding", Driving: "driving", Transit: "transit"}[mode]
	qm := map[Mode]string{Walking: "walk", Cycling: "bike", Driving: "drive", Transit: "bus"}[mode]
	return []ExternalLink{{"amap", fmt.Sprintf("amapuri://route/plan/?slat=%.6f&slon=%.6f&dlat=%.6f&dlon=%.6f&dname=%s&dev=1&t=%s", origin.Latitude, origin.Longitude, destination.Latitude, destination.Longitude, n, am), true}, {"baidu", fmt.Sprintf("baidumap://map/direction?origin=latlng:%s&destination=latlng:%s|name:%s&mode=%s&coord_type=gcj02", from, to, n, bm), true}, {"tencent", fmt.Sprintf("qqmap://map/routeplan?fromcoord=%s&to=%s&tocoord=%s&type=%s", from, n, to, qm), true}}, nil
}

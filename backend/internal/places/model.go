package places

import (
	"alongtu/backend/internal/platform/geo"
	"context"
	"errors"
	"time"
)

var ErrIdempotencyConflict = errors.New("idempotency key reused with different payload")

type Category string

const (
	CategoryUnclassified     Category = "unclassified"
	CategoryToilet           Category = "toilet"
	CategoryFuel             Category = "fuel"
	CategoryCharging         Category = "charging"
	CategorySupermarket      Category = "supermarket"
	CategoryAttraction       Category = "attraction"
	CategoryHotel            Category = "hotel"
	CategoryParking          Category = "parking"
	CategoryPowerBank        Category = "power_bank"
	CategoryPharmacy         Category = "pharmacy"
	CategoryHospital         Category = "hospital"
	CategoryConvenienceStore Category = "convenience_store"
	CategoryPolice           Category = "police"
	CategoryFood             Category = "food"
)

func IsKnownCategory(category Category) bool {
	switch category {
	case "", CategoryUnclassified, CategoryToilet, CategoryFuel, CategoryCharging, CategorySupermarket,
		CategoryAttraction, CategoryHotel, CategoryParking, CategoryPowerBank, CategoryPharmacy,
		CategoryHospital, CategoryConvenienceStore, CategoryPolice, CategoryFood:
		return true
	default:
		return false
	}
}

type TrustItem struct {
	Status      string     `json:"status"`
	ConfirmedAt *time.Time `json:"confirmedAt,omitempty"`
	Source      string     `json:"source,omitempty"`
}
type TrustSummary struct {
	Exists       TrustItem `json:"exists"`
	OpeningHours TrustItem `json:"openingHours"`
}
type Place struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Category        Category       `json:"category"`
	Address         string         `json:"address,omitempty"`
	CityCode        string         `json:"cityCode,omitempty"`
	DistrictCode    string         `json:"districtCode,omitempty"`
	Location        geo.Coordinate `json:"location"`
	DistanceMeters  *float64       `json:"distanceMeters,omitempty"`
	Trust           TrustSummary   `json:"trust"`
	SourceUpdatedAt *time.Time     `json:"sourceUpdatedAt,omitempty"`
}
type NearbyQuery struct {
	Center       geo.Coordinate
	Category     Category
	RadiusMeters int
	Limit        int
	Cursor       string
	Page         int
}
type SearchQuery struct {
	Keyword  string
	Center   *geo.Coordinate
	CityCode string
	Category Category
	Limit    int
	Cursor   string
	Page     int
}
type Page struct {
	Items         []Place `json:"items"`
	NextCursor    string  `json:"nextCursor,omitempty"`
	Degraded      bool    `json:"degraded"`
	DataFreshness string  `json:"dataFreshness"`
	Coverage      string  `json:"coverage"`
}

type Repository interface {
	Nearby(context.Context, NearbyQuery) (Page, error)
	Get(context.Context, string) (Place, error)
	Materialize(context.Context, []Place) ([]Place, error)
	AddConfirmation(context.Context, Confirmation) (Confirmation, error)
	AddFeedback(context.Context, Feedback) (Feedback, error)
}
type Provider interface {
	Nearby(context.Context, NearbyQuery) (Page, error)
	Search(context.Context, SearchQuery) (Page, error)
}
type Confirmation struct {
	ID             string    `json:"id"`
	PlaceID        string    `json:"placeId"`
	UserID         string    `json:"-"`
	SessionID      string    `json:"-"`
	Kind           string    `json:"kind"`
	Result         string    `json:"result"`
	IdempotencyKey string    `json:"-"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Feedback struct {
	ID             string    `json:"id"`
	PlaceID        string    `json:"placeId"`
	UserID         string    `json:"-"`
	SessionID      string    `json:"-"`
	Kind           string    `json:"kind"`
	Details        string    `json:"details"`
	IdempotencyKey string    `json:"-"`
	CreatedAt      time.Time `json:"createdAt"`
}

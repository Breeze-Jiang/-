export type CoordinateSystem = 'wgs84' | 'gcj02';

export type P0PlaceCategory =
  | 'unclassified'
  | 'toilet'
  | 'fuel'
  | 'charging'
  | 'supermarket'
  | 'attraction'
  | 'hotel'
  | 'parking'
  | 'power_bank'
  | 'pharmacy'
  | 'hospital'
  | 'convenience_store'
  | 'police'
  | 'food';

export type NavigationMode = 'walking' | 'cycling' | 'driving' | 'transit' | 'taxi';

export interface Coordinate {
  latitude: number;
  longitude: number;
  coordinateSystem: CoordinateSystem;
}

export interface TrustItem {
  status: 'unknown' | 'confirmed' | 'incorrect';
  confirmedAt?: string;
  source?: string;
}

export interface P0Place {
  id: string;
  name: string;
  category: P0PlaceCategory;
  address?: string;
  cityCode?: string;
  districtCode?: string;
  location: Coordinate;
  distanceMeters?: number;
  trust: {
    exists: TrustItem;
    openingHours: TrustItem;
  };
  sourceUpdatedAt?: string;
}

export interface PlacePage {
  items: P0Place[];
  nextCursor?: string;
  degraded: boolean;
  dataFreshness: 'live' | 'stored' | 'cached' | 'degraded';
  coverage: 'covered' | 'not_covered';
}

export interface Tokens {
  accessToken: string;
  refreshToken: string;
  accessExpiresAt: string;
  refreshExpiresAt: string;
}

export interface Confirmation {
  id: string;
  placeId: string;
  kind: 'exists' | 'opening_hours';
  result: 'confirmed' | 'incorrect' | 'unknown';
  createdAt: string;
}

export interface PlaceFeedback {
  id: string;
  placeId: string;
  kind: 'entrance' | 'incorrect_info';
  details: string;
  createdAt: string;
}

export interface RouteStep {
  instruction: string;
  distanceMeters: number;
  durationSeconds?: number;
  polyline?: Coordinate[];
}

export interface ExternalLink {
  provider: 'amap' | 'baidu' | 'tencent' | 'ride_hailing';
  uri?: string;
  available: boolean;
}

export interface Route {
  distanceMeters: number;
  durationSeconds: number;
  polyline: Coordinate[];
  steps: RouteStep[];
  dataFreshness: 'live' | 'cached' | 'external_only';
  externalFallback: ExternalLink[];
}

export interface ApiErrorBody {
  code: string;
  message: string;
  details?: unknown;
  requestId: string;
}

export interface Envelope<T> {
  data?: T;
  error?: ApiErrorBody;
}

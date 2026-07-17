import { NativeModules, Platform } from 'react-native';
import type {
  Confirmation,
  Coordinate,
  Envelope,
  ExternalLink,
  NavigationMode,
  P0Place,
  P0PlaceCategory,
  PlaceFeedback,
  PlacePage,
  Route,
  Tokens,
} from './contracts';

type P0RuntimeConfig = { apiOrigin?: string };

function defaultApiOrigin() {
  const origin = (NativeModules.P0RuntimeConfig as P0RuntimeConfig | undefined)?.apiOrigin?.trim();
  if (origin) return origin.replace(/\/$/, '');
  // This development-only fallback works for the Android emulator. USB devices
  // and every release must set P0_API_ORIGIN in the native build configuration.
  return __DEV__ && Platform.OS === 'android' ? 'http://10.0.2.2:8080' : '';
}

let apiOrigin = defaultApiOrigin();

export class P0ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId?: string;

  constructor(status: number, code: string, message: string, requestId?: string) {
    super(message);
    this.name = 'P0ApiError';
    this.status = status;
    this.code = code;
    this.requestId = requestId;
  }
}

export function configureP0ApiOrigin(origin: string) {
  const normalized = origin.trim().replace(/\/$/, '');
  if (!/^https?:\/\//.test(normalized)) {
    throw new Error('P0 API origin must use http or https');
  }
  apiOrigin = normalized;
}

export function p0ApiOrigin() {
  return apiOrigin;
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  if (!apiOrigin) {
    throw new P0ApiError(0, 'api_not_configured', '当前版本尚未配置服务地址，请联系测试或发布人员。');
  }
  let response: Response;
  try {
    response = await fetch(`${apiOrigin}/api/v1${path}`, {
      ...init,
      headers: { Accept: 'application/json', ...(init.body ? { 'Content-Type': 'application/json' } : {}), ...init.headers },
    });
  } catch {
    throw new P0ApiError(0, 'network_unavailable', '网络连接不可用，请稍后重试');
  }

  if (response.status === 204) {
    return undefined as T;
  }
  const envelope = await response.json() as Envelope<T>;
  if (!response.ok || envelope.error) {
    const error = envelope.error;
    throw new P0ApiError(response.status, error?.code ?? 'unexpected_response', error?.message ?? '服务暂时不可用', error?.requestId);
  }
  if (envelope.data === undefined) {
    throw new P0ApiError(response.status, 'unexpected_response', '服务返回了无效响应');
  }
  return envelope.data;
}

function query(values: Record<string, string | number | undefined>) {
  const params = new URLSearchParams();
  Object.entries(values).forEach(([key, value]) => {
    if (value !== undefined && value !== '') params.set(key, String(value));
  });
  return params.toString();
}

export const p0Api = {
  nearby(input: { center: Coordinate; category?: P0PlaceCategory; radiusMeters?: number; limit?: number; cursor?: string }) {
    return request<PlacePage>(`/places/nearby?${query({ latitude: input.center.latitude, longitude: input.center.longitude, coordinateSystem: input.center.coordinateSystem, category: input.category, radiusMeters: input.radiusMeters, limit: input.limit, cursor: input.cursor })}`);
  },
  search(input: { keyword: string; category?: P0PlaceCategory; cityCode?: string; limit?: number; cursor?: string }) {
    return request<PlacePage>(`/places/search?${query({ q: input.keyword, category: input.category, cityCode: input.cityCode, limit: input.limit, cursor: input.cursor })}`);
  },
  place(id: string) {
    return request<P0Place>(`/places/${encodeURIComponent(id)}`);
  },
  sendSMS(phone: string, deviceId: string) {
    return request<{ challengeId: string; expiresIn: string }>('/auth/sms/send', { method: 'POST', body: JSON.stringify({ phone, deviceId }) });
  },
  verifySMS(challengeId: string, code: string, deviceId: string) {
    return request<Tokens>('/auth/sms/verify', { method: 'POST', body: JSON.stringify({ challengeId, code, deviceId }) });
  },
  refresh(refreshToken: string) {
    return request<Tokens>('/auth/refresh', { method: 'POST', body: JSON.stringify({ refreshToken }) });
  },
  logout(refreshToken: string) {
    return request<void>('/auth/logout', { method: 'POST', body: JSON.stringify({ refreshToken }) });
  },
  confirm(placeId: string, token: string, idempotencyKey: string, kind: Confirmation['kind'], result: Confirmation['result']) {
    return request<Confirmation>(`/places/${encodeURIComponent(placeId)}/confirmations`, { method: 'POST', headers: { Authorization: `Bearer ${token}`, 'Idempotency-Key': idempotencyKey }, body: JSON.stringify({ kind, result }) });
  },
  feedback(placeId: string, token: string, idempotencyKey: string, kind: PlaceFeedback['kind'], details: string) {
    return request<PlaceFeedback>(`/places/${encodeURIComponent(placeId)}/feedback`, { method: 'POST', headers: { Authorization: `Bearer ${token}`, 'Idempotency-Key': idempotencyKey }, body: JSON.stringify({ kind, details }) });
  },
  route(origin: Coordinate, destination: Coordinate, mode: NavigationMode, city?: { origin?: string; destination?: string }) {
    return request<Route>('/navigation/routes', { method: 'POST', body: JSON.stringify({ origin, destination, mode, originCity: city?.origin, destinationCity: city?.destination }) });
  },
  externalLinks(origin: Coordinate, destination: Coordinate, destinationName: string, mode: NavigationMode) {
    return request<{ links: ExternalLink[] }>(`/navigation/external-links?${query({ originLatitude: origin.latitude, originLongitude: origin.longitude, destinationLatitude: destination.latitude, destinationLongitude: destination.longitude, destinationName, coordinateSystem: origin.coordinateSystem, mode })}`);
  },
};

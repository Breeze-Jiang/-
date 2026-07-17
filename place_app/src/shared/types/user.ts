export interface User {
  id: string;
  phone: string;
  nickname?: string;
  avatarUrl?: string;
  gender?: 'male' | 'female' | 'other';
  bio?: string;
  homeLat?: number;
  homeLng?: number;
  homeAddr?: string;
  createdAt: string;
}

export interface UserPreference {
  poiType: string;
  enabled: boolean;
}

export interface LoginRequest {
  phone: string;
  password: string;
}

export interface RegisterRequest {
  phone: string;
  code: string;
  password: string;
}

export interface TokenResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

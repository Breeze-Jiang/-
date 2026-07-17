import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Keychain from 'react-native-keychain';
import { create } from 'zustand';
import { p0Api, P0ApiError } from './client';
import type { Tokens } from './contracts';

const DEVICE_ID_KEY = '@alongtu/p0-device-id';
const SESSION_SERVICE = 'com.placenative.p0.session';
const SESSION_USER = 'p0-refresh-session';

function isTokens(value: unknown): value is Tokens {
  if (!value || typeof value !== 'object') return false;
  const tokens = value as Tokens;
  return typeof tokens.accessToken === 'string' && tokens.accessToken.length > 0
    && typeof tokens.refreshToken === 'string' && tokens.refreshToken.length > 0
    && Number.isFinite(Date.parse(tokens.accessExpiresAt))
    && Number.isFinite(Date.parse(tokens.refreshExpiresAt));
}

async function readTokens(): Promise<Tokens | undefined> {
  const credentials = await Keychain.getGenericPassword({ service: SESSION_SERVICE });
  if (!credentials) return undefined;
  try {
    const parsed = JSON.parse(credentials.password) as unknown;
    return isTokens(parsed) ? parsed : undefined;
  } catch {
    return undefined;
  }
}

async function writeTokens(tokens: Tokens) {
  const result = await Keychain.setGenericPassword(SESSION_USER, JSON.stringify(tokens), { service: SESSION_SERVICE });
  if (!result) throw new Error('secure_storage_unavailable');
}

async function clearTokens() {
  await Keychain.resetGenericPassword({ service: SESSION_SERVICE });
}

export async function deviceID(): Promise<string> {
  const existing = await AsyncStorage.getItem(DEVICE_ID_KEY);
  if (existing && existing.length >= 8) return existing;
  // The device identifier is only a rate-limit/session binding value; it is
  // not an authentication credential and is never sent to logs.
  const created = `device-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 14)}`;
  await AsyncStorage.setItem(DEVICE_ID_KEY, created);
  return created;
}

export function newIdempotencyKey() {
  return `confirm-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 18)}`;
}

type SessionStatus = 'hydrating' | 'anonymous' | 'authenticated';

interface P0SessionState {
  status: SessionStatus;
  tokens?: Tokens;
  bootstrap: () => Promise<void>;
  sendCode: (phone: string) => Promise<{ challengeId: string; expiresIn: string }>;
  verifyCode: (challengeId: string, code: string) => Promise<void>;
  accessTokenForWrite: () => Promise<string>;
  signOut: () => Promise<void>;
  clearLocal: () => Promise<void>;
}

export const useP0SessionStore = create<P0SessionState>((set, get) => ({
  status: 'hydrating',
  bootstrap: async () => {
    try {
      const tokens = await readTokens();
      if (!tokens || Date.parse(tokens.refreshExpiresAt) <= Date.now()) {
        await clearTokens();
        set({ status: 'anonymous', tokens: undefined });
        return;
      }
      set({ status: 'authenticated', tokens });
    } catch {
      set({ status: 'anonymous', tokens: undefined });
    }
  },
  sendCode: async (phone) => p0Api.sendSMS(phone, await deviceID()),
  verifyCode: async (challengeId, code) => {
    const tokens = await p0Api.verifySMS(challengeId, code, await deviceID());
    await writeTokens(tokens);
    set({ status: 'authenticated', tokens });
  },
  accessTokenForWrite: async () => {
    const tokens = get().tokens;
    if (!tokens) throw new P0ApiError(401, 'authentication_required', '请先登录');
    if (Date.parse(tokens.accessExpiresAt) > Date.now() + 60_000) return tokens.accessToken;
    try {
      const rotated = await p0Api.refresh(tokens.refreshToken);
      await writeTokens(rotated);
      set({ status: 'authenticated', tokens: rotated });
      return rotated.accessToken;
    } catch (cause) {
      if (cause instanceof P0ApiError && cause.status === 401) await get().clearLocal();
      throw cause;
    }
  },
  signOut: async () => {
    const refreshToken = get().tokens?.refreshToken;
    try {
      if (refreshToken) await p0Api.logout(refreshToken);
    } catch {
      // Local credential removal remains mandatory even when the network is
      // unavailable; the server-side token will expire on its normal TTL.
    } finally {
      await get().clearLocal();
    }
  },
  clearLocal: async () => {
    try {
      await clearTokens();
    } finally {
      set({ status: 'anonymous', tokens: undefined });
    }
  },
}));

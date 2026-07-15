import { Injectable } from '@angular/core';

const TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';
const USER_KEY = 'user_info';

export type StoredUser = {
  id?: string;
  username?: string;
  email?: string;
  [key: string]: unknown;
};

function parseNumericId(raw: unknown): number | null {
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw;
  if (typeof raw === 'string') {
    const parsed = Number.parseInt(raw, 10);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
}

function parseUser(raw: string | null): StoredUser | null {
  if (!raw) return null;
  try {
    return JSON.parse(raw) as StoredUser;
  } catch {
    return null;
  }
}

function parseJwtPayload(token: string): Record<string, unknown> | null {
  if (!token) return null;
  const parts = token.split('.');
  if (parts.length < 2) return null;
  try {
    const normalized = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');
    return JSON.parse(window.atob(padded)) as Record<string, unknown>;
  } catch {
    return null;
  }
}

@Injectable({ providedIn: 'root' })
export class TokenStorageService {
  getToken(): string {
    return localStorage.getItem(TOKEN_KEY) || '';
  }

  setToken(token: string): void {
    if (!token) return;
    localStorage.setItem(TOKEN_KEY, token);
  }

  clearToken(): void {
    localStorage.removeItem(TOKEN_KEY);
  }

  getRefreshToken(): string {
    return localStorage.getItem(REFRESH_TOKEN_KEY) || '';
  }

  setRefreshToken(token: string): void {
    if (!token) return;
    localStorage.setItem(REFRESH_TOKEN_KEY, token);
  }

  clearRefreshToken(): void {
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  }

  getUser(): StoredUser | null {
    return parseUser(localStorage.getItem(USER_KEY));
  }

  getUserId(): number | null {
    const stored = this.getUser();
    const storedId = parseNumericId(stored?.id);
    if (storedId != null) return storedId;
    const payload = parseJwtPayload(this.getToken());
    return parseNumericId(payload?.['user_id']);
  }

  getActiveTenantId(): string {
    const payload = parseJwtPayload(this.getToken());
    const activeTenantId = payload?.['active_tenant_id'];
    return typeof activeTenantId === 'string' ? activeTenantId : '';
  }

  getSessionId(): string {
    const payload = parseJwtPayload(this.getToken());
    const sessionId = payload?.['session_id'];
    return typeof sessionId === 'string' ? sessionId : '';
  }

  setUser(user: StoredUser): void {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }

  clearUser(): void {
    localStorage.removeItem(USER_KEY);
  }

  clearAll(): void {
    this.clearToken();
    this.clearRefreshToken();
    this.clearUser();
  }
}

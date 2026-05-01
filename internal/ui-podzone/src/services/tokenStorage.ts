const TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';
const USER_KEY = 'user_info';

export type StoredUser = {
  id?: string;
  username?: string;
  email?: string;
  [key: string]: unknown;
};

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

export const tokenStorage = {
  getToken(): string {
    return localStorage.getItem(TOKEN_KEY) || '';
  },
  setToken(token: string): void {
    if (!token) return;
    localStorage.setItem(TOKEN_KEY, token);
  },
  clearToken(): void {
    localStorage.removeItem(TOKEN_KEY);
  },
  getRefreshToken(): string {
    return localStorage.getItem(REFRESH_TOKEN_KEY) || '';
  },
  setRefreshToken(token: string): void {
    if (!token) return;
    localStorage.setItem(REFRESH_TOKEN_KEY, token);
  },
  clearRefreshToken(): void {
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  },

  getUser(): StoredUser | null {
    return parseUser(localStorage.getItem(USER_KEY));
  },
  getActiveTenantID(): string {
    const token = this.getToken();
    const payload = parseJwtPayload(token);
    const activeTenantID = payload?.active_tenant_id;
    return typeof activeTenantID === 'string' ? activeTenantID : '';
  },
  getSessionID(): string {
    const token = this.getToken();
    const payload = parseJwtPayload(token);
    const sessionID = payload?.session_id;
    return typeof sessionID === 'string' ? sessionID : '';
  },
  setUser(user: StoredUser): void {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  },
  clearUser(): void {
    localStorage.removeItem(USER_KEY);
  },

  clearAll(): void {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  },
};

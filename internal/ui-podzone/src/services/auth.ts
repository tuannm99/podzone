import { GW_API_URL } from './baseurl';
import { http, type HttpError } from './http';
import { tenantStorage } from './tenantStorage';
import { tokenStorage, type StoredUser } from './tokenStorage';

export type AuthPayload = {
  username: string;
  password: string;
};

export type RegisterPayload = {
  username: string;
  email: string;
  password: string;
};

export type AuthResponseData = {
  jwtToken?: string;
  refreshToken?: string;
  userInfo?: StoredUser;
  message?: string;
  [key: string]: unknown;
};

export type AuthResult =
  | { success: true; data: AuthResponseData }
  | { success: false; data: { message: string } };

type SwitchTenantPayload = {
  userId: number;
  tenantId: string;
  accessToken: string;
};

function persistAuth(data: AuthResponseData) {
  if (data.jwtToken) tokenStorage.setToken(data.jwtToken);
  if (data.refreshToken) tokenStorage.setRefreshToken(data.refreshToken);
  if (data.userInfo) tokenStorage.setUser(data.userInfo);
  const activeTenantID = tokenStorage.getActiveTenantID();
  if (activeTenantID) tenantStorage.setTenantID(activeTenantID);
}

function toFailure(error: unknown, fallback: string): AuthResult {
  const message =
    typeof error === 'object' &&
    error &&
    'message' in error &&
    typeof error.message === 'string'
      ? error.message
      : fallback;

  return { success: false, data: { message } };
}

export function loginGG(): string {
  return `${GW_API_URL}/auth/v1/google/login`;
}

export async function login(payload: AuthPayload): Promise<AuthResult> {
  try {
    const { data } = await http.post<AuthResponseData>(
      '/auth/v1/login',
      payload
    );
    persistAuth(data);
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Login failed');
  }
}

export async function register(payload: RegisterPayload): Promise<AuthResult> {
  try {
    const { data } = await http.post<AuthResponseData>(
      '/auth/v1/register',
      payload
    );
    persistAuth(data);
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Register failed');
  }
}

function parseStoredUserID(user: StoredUser | null): number | null {
  if (!user || user.id == null) return null;
  if (typeof user.id === 'number' && Number.isFinite(user.id)) return user.id;
  if (typeof user.id === 'string') {
    const parsed = Number.parseInt(user.id, 10);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
}

async function switchTenantRequest(
  payload: SwitchTenantPayload
): Promise<AuthResult> {
  try {
    const { data } = await http.post<AuthResponseData>(
      '/auth/v1/iam/tenants:switch',
      payload
    );
    persistAuth(data);
    tenantStorage.setTenantID(payload.tenantId);
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Switch tenant failed');
  }
}

export async function switchActiveTenant(tenantId: string): Promise<AuthResult> {
  const normalizedTenantID = tenantId.trim();
  if (!normalizedTenantID) {
    return { success: false, data: { message: 'Tenant id is required' } };
  }

  const userId = parseStoredUserID(tokenStorage.getUser());
  if (!userId) {
    return { success: false, data: { message: 'No authenticated user found' } };
  }
  const accessToken = tokenStorage.getToken();
  if (!accessToken) {
    return { success: false, data: { message: 'No active session token found' } };
  }

  return switchTenantRequest({
    userId,
    tenantId: normalizedTenantID,
    accessToken,
  });
}

export async function ensureActiveTenant(tenantId: string): Promise<AuthResult> {
  const normalizedTenantID = tenantId.trim();
  if (!normalizedTenantID) {
    return { success: false, data: { message: 'Tenant id is required' } };
  }

  if (tokenStorage.getActiveTenantID() === normalizedTenantID) {
    tenantStorage.setTenantID(normalizedTenantID);
    return {
      success: true,
      data: {
        jwtToken: tokenStorage.getToken(),
        userInfo: tokenStorage.getUser() || undefined,
      },
    };
  }

  return switchActiveTenant(normalizedTenantID);
}

export async function refreshSession(): Promise<AuthResult> {
  const refreshToken = tokenStorage.getRefreshToken();
  if (!refreshToken) {
    return { success: false, data: { message: 'No refresh token found' } };
  }

  try {
    const { data } = await http.post<AuthResponseData>('/auth/v1/refresh', {
      refreshToken,
    });
    persistAuth(data);
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Refresh failed');
  }
}

export async function logout(): Promise<void> {
  const token = tokenStorage.getToken();
  try {
    if (token) {
      await http.post('/auth/v1/logout', { token });
    }
  } catch {
    // Revoke on server is best-effort here; local credentials are cleared regardless.
  } finally {
    tokenStorage.clearAll();
    tenantStorage.clearTenantID();
    window.location.href = '/auth/login';
  }
}

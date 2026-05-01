import axios, {
  AxiosError,
  type AxiosRequestConfig,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from 'axios';
import { GW_API_URL } from './baseurl';
import { tokenStorage } from './tokenStorage';

export type HttpError = {
  status?: number;
  data?: unknown;
  message: string;
};

function safeError(error: unknown): HttpError {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    const data = error.response?.data;
    const message =
      (typeof data === 'object' &&
      data &&
      'message' in data &&
      typeof data.message === 'string'
        ? data.message
        : undefined) ||
      error.message ||
      'Request failed';

    return { status, data, message };
  }

  if (error instanceof Error) {
    return { message: error.message };
  }

  return { message: 'Request failed' };
}

export const http = axios.create({
  baseURL: GW_API_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 20_000,
});

type RetriableRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean;
};

type RefreshResponseData = {
  jwtToken?: string;
  refreshToken?: string;
  userInfo?: unknown;
};

let refreshInFlight: Promise<string | null> | null = null;

async function performRefresh(): Promise<string | null> {
  const refreshToken = tokenStorage.getRefreshToken();
  if (!refreshToken) return null;

  const { data } = await axios.post<RefreshResponseData>(
    `${GW_API_URL}/auth/v1/refresh`,
    { refreshToken },
    {
      headers: { 'Content-Type': 'application/json' },
      timeout: 20_000,
    }
  );

  if (data.jwtToken) tokenStorage.setToken(data.jwtToken);
  if (data.refreshToken) tokenStorage.setRefreshToken(data.refreshToken);
  if (data.userInfo && typeof data.userInfo === 'object') {
    tokenStorage.setUser(data.userInfo as Record<string, unknown>);
  }

  return data.jwtToken || null;
}

function redirectToLogin() {
  tokenStorage.clearAll();
  if (window.location.pathname !== '/auth/login') {
    window.location.href = '/auth/login';
  }
}

async function refreshAccessToken(): Promise<string | null> {
  if (!refreshInFlight) {
    refreshInFlight = performRefresh()
      .catch(() => null)
      .finally(() => {
        refreshInFlight = null;
      });
  }
  return refreshInFlight;
}

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = tokenStorage.getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: AxiosError) => {
    const status = error.response?.status;
    const request = error.config as RetriableRequestConfig | undefined;
    const url = request?.url || '';

    if (
      status === 401 &&
      request &&
      !request._retry &&
      !url.includes('/auth/v1/login') &&
      !url.includes('/auth/v1/register') &&
      !url.includes('/auth/v1/refresh')
    ) {
      request._retry = true;
      const nextToken = await refreshAccessToken();
      if (nextToken) {
        request.headers =
          request.headers || ({} as AxiosRequestConfig['headers']);
        request.headers.Authorization = `Bearer ${nextToken}`;
        return http(request);
      }
      redirectToLogin();
      return Promise.reject(safeError(error));
    }

    if (status === 401) {
      redirectToLogin();
    }

    return Promise.reject(safeError(error));
  }
);

import axios, {
  AxiosError,
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

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = tokenStorage.getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      tokenStorage.clearAll();

      if (window.location.pathname !== '/auth/login') {
        window.location.href = '/auth/login';
      }
    }

    return Promise.reject(safeError(error));
  }
);

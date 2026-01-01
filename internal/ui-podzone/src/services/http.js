import axios from 'axios';
import { ADMIN_API_URL } from './baseurl';
import { tokenStorage } from './tokenStorage';

export const http = axios.create({
  baseURL: ADMIN_API_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 20000,
});

function safeError(e) {
  const status = e?.response?.status;
  const data = e?.response?.data;
  const message = data?.message || e?.message || 'Request failed';
  return { status, data, message };
}

http.interceptors.request.use((config) => {
  const token = tokenStorage.getToken();
  if (token) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (res) => res,
  (error) => {
    const status = error?.response?.status;

    // token expired / invalid => logout
    if (status === 401) {
      tokenStorage.clearAll();

      // redirect cứng (vì interceptor không có navigate hook)
      if (window.location.pathname !== '/auth/login') {
        window.location.href = '/auth/login';
      }
    }

    return Promise.reject(safeError(error));
  },
);

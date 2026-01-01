import { http } from './http';
import { GW_API_URL } from './baseurl';
import { tokenStorage } from './tokenStorage';

export const loginGG = () => `${GW_API_URL}/auth/v1/google/login`;

export const login = async (payload) => {
  try {
    const { data } = await http.post('/auth/v1/login', payload);

    if (data?.jwtToken) tokenStorage.setToken(data.jwtToken);
    if (data?.userInfo) tokenStorage.setUser(data.userInfo);

    return { success: true, data };
  } catch (e) {
    return { success: false, data: { message: e?.message || 'Login failed' } };
  }
};

export const register = async (payload) => {
  try {
    const { data } = await http.post('/auth/v1/register', payload);

    if (data?.jwtToken) tokenStorage.setToken(data.jwtToken);
    if (data?.userInfo) tokenStorage.setUser(data.userInfo);

    return { success: true, data };
  } catch (e) {
    return {
      success: false,
      data: { message: e?.message || 'Register failed' },
    };
  }
};

export const logout = () => {
  tokenStorage.clearAll();
  window.location.href = '/auth/login';
};

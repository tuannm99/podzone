import axios from 'axios';
import { ADMIN_API_URL } from './baseurl';

export const loginGG = async () => {
    return `${ADMIN_API_URL}/auth/v1/google/login`;
};

export const login = async (payload) => {
    const loginUrl = `${ADMIN_API_URL}/auth/v1/login`;
    try {
        const { data } = await axios.post(loginUrl, payload, {
            headers: {
                'Content-Type': 'application/json',
            },
        });
        return { success: true, data };
    } catch (e) {
        return { success: false, data: e.response.data };
    }
};

export const register = async (payload) => {
    const url = `${ADMIN_API_URL}/auth/v1/register`;
    try {
        const { data } = await axios.post(url, payload, {
            headers: {
                'Content-Type': 'application/json',
            },
        });
        return { success: true, data };
    } catch (e) {
        return { success: false, data: e.response.data };
    }
};

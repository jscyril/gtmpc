/**
 * src/api/client.ts
 * Axios instance configured for the gtmpc API.
 * - Uses relative URLs so the React app works when served by the Go backend.
 * - Injects Authorization header from localStorage on every request.
 * - Handles 401 by clearing the token and redirecting to /login.
 */

import axios from 'axios';
import { getToken, clearToken } from '../utils/storage';

const apiClient = axios.create({
  baseURL: '',          // relative — same-origin when served by Go backend
  timeout: 30_000,
  headers: { 'Content-Type': 'application/json' },
});

// Request interceptor: inject Authorization header
apiClient.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: handle 401 globally
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      clearToken();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default apiClient;

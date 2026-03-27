/**
 * src/context/AuthContext.tsx
 * Manages authentication state: login, logout, register, token persistence.
 */

import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';
import { login as apiLogin, register as apiRegister } from '../api/auth';
import {
  getToken, setToken, clearToken,
  getUsername, setUsername,
  setExpiresAt, isTokenExpired,
} from '../utils/storage';
import type { LoginRequest, RegisterRequest, RegisterResponse } from '../api/types';

interface AuthState {
  token: string | null;
  username: string | null;
  isAuthenticated: boolean;
}

interface AuthContextValue extends AuthState {
  login: (req: LoginRequest) => Promise<void>;
  logout: () => void;
  register: (req: RegisterRequest) => Promise<RegisterResponse>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>(() => {
    const token = getToken();
    const expired = isTokenExpired();
    if (token && !expired) {
      return { token, username: getUsername(), isAuthenticated: true };
    }
    clearToken();
    return { token: null, username: null, isAuthenticated: false };
  });

  // Watch for token expiry
  useEffect(() => {
    const interval = setInterval(() => {
      if (state.isAuthenticated && isTokenExpired()) {
        clearToken();
        setState({ token: null, username: null, isAuthenticated: false });
        window.location.href = '/login';
      }
    }, 60_000);
    return () => clearInterval(interval);
  }, [state.isAuthenticated]);

  const login = useCallback(async (req: LoginRequest) => {
    const resp = await apiLogin(req);
    setToken(resp.token);
    setUsername(resp.username);
    setExpiresAt(resp.expires_at);
    setState({ token: resp.token, username: resp.username, isAuthenticated: true });
  }, []);

  const logout = useCallback(() => {
    clearToken();
    setState({ token: null, username: null, isAuthenticated: false });
    window.location.href = '/login';
  }, []);

  const register = useCallback(async (req: RegisterRequest) => {
    return apiRegister(req);
  }, []);

  return (
    <AuthContext.Provider value={{ ...state, login, logout, register }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

/**
 * src/api/auth.ts
 * Authentication API calls: login and register.
 */

import apiClient from './client';
import type { LoginRequest, LoginResponse, RegisterRequest, RegisterResponse } from './types';

export async function login(req: LoginRequest): Promise<LoginResponse> {
  const { data } = await apiClient.post<LoginResponse>('/api/auth/login', req);
  return data;
}

export async function register(req: RegisterRequest): Promise<RegisterResponse> {
  const { data } = await apiClient.post<RegisterResponse>('/api/auth/register', req);
  return data;
}

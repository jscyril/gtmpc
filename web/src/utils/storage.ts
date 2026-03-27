/**
 * src/utils/storage.ts
 * localStorage helpers for token persistence.
 */

const TOKEN_KEY = 'gtmpc_token';
const USERNAME_KEY = 'gtmpc_username';
const EXPIRES_KEY = 'gtmpc_expires_at';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USERNAME_KEY);
  localStorage.removeItem(EXPIRES_KEY);
}

export function getUsername(): string | null {
  return localStorage.getItem(USERNAME_KEY);
}

export function setUsername(username: string): void {
  localStorage.setItem(USERNAME_KEY, username);
}

export function getExpiresAt(): string | null {
  return localStorage.getItem(EXPIRES_KEY);
}

export function setExpiresAt(expiresAt: string): void {
  localStorage.setItem(EXPIRES_KEY, expiresAt);
}

export function isTokenExpired(): boolean {
  const expiresAt = getExpiresAt();
  if (!expiresAt) return true;
  return new Date(expiresAt) < new Date();
}

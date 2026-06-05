/**
 * Authentication utility functions
 * Handles JWT token storage and authentication state
 */

const TOKEN_KEY = 'xingyunpan_token';

/**
 * Save JWT token to localStorage
 * @param token - JWT token string
 */
export function saveToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

/**
 * Get JWT token from localStorage
 * @returns JWT token string or null if not found
 */
export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * Remove JWT token from localStorage
 */
export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

/**
 * Check if user is authenticated
 * @returns true if JWT token exists, false otherwise
 */
export function isAuthenticated(): boolean {
  return !!getToken();
}

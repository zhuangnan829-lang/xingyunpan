/**
 * User API module
 * Handles all user-related API calls
 */

import request from './request';
import type {
  RegisterRequest,
  LoginRequest,
  LoginResponse,
  UserProfile,
  UserPreferences,
  UpdateProfileRequest,
  ChangePasswordRequest,
  ResetPasswordRequest,
} from '@/types/user';
import type { CaptchaPayload } from './captcha';

/**
 * User registration
 * @param data - Registration data (username, email, password)
 * @returns Promise that resolves when registration succeeds
 */
export function register(data: RegisterRequest): Promise<void> {
  return request.post('/api/v1/user/register', data);
}

/**
 * Send register email verification code
 * @param email - Target email
 */
export function sendRegisterEmailCode(email: string, captcha?: CaptchaPayload): Promise<void> {
  return request.post('/api/v1/user/register/email-code', { email, ...(captcha || {}) });
}

/**
 * Send reset-password email verification code
 * @param email - Registered email
 */
export function sendResetPasswordEmailCode(email: string, captcha?: CaptchaPayload): Promise<void> {
  return request.post('/api/v1/user/password/email-code', { email, ...(captcha || {}) });
}

/**
 * User login
 * @param data - Login credentials (username/email and password)
 * @returns Promise with login response containing token and user profile
 */
export function login(data: LoginRequest): Promise<LoginResponse> {
  return request.post('/api/v1/user/login', data);
}

/**
 * Get user profile
 * @returns Promise with user profile information
 */
export function getProfile(): Promise<UserProfile> {
  return request.get('/api/v1/user/profile');
}

export function updateProfile(data: UpdateProfileRequest): Promise<UserProfile> {
  return request.put('/api/v1/user/profile', data);
}

export function uploadAvatar(file: File): Promise<UserProfile> {
  const formData = new FormData();
  formData.append('avatar', file);
  return request.post('/api/v1/user/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}

export function getPreferences(): Promise<UserPreferences> {
  return request.get('/api/v1/user/preferences');
}

export function updatePreferences(data: UserPreferences): Promise<UserPreferences> {
  return request.put('/api/v1/user/preferences', data);
}

/**
 * Change user password
 * @param data - Password change data (old_password, new_password)
 * @returns Promise that resolves when password change succeeds
 */
export function changePassword(data: ChangePasswordRequest): Promise<void> {
  return request.post('/api/v1/user/change-password', data);
}

/**
 * Reset password by email code
 * @param data - Email, code and new password
 */
export function resetPasswordByEmailCode(data: ResetPasswordRequest): Promise<void> {
  return request.post('/api/v1/user/password/reset', data);
}

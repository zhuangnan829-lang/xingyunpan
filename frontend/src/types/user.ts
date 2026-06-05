// User-related TypeScript type definitions

/**
 * User profile information
 */
export interface UserProfile {
  id: number;
  username: string;
  avatar_url?: string;
  email: string;
  role: string;            // 用户角色（admin/user）
  used_size: number;       // 已使用空间（字节）
  capacity: number;        // 总配额（字节）
  created_at: string;
}

export interface UserPreferences {
  user_id?: number;
  language: string;
  timezone: string;
  mode: string;
  theme: string;
  keep_versions: boolean;
  version_extensions: string;
  max_versions: number;
  view_sync: string;
  expand_tree: boolean;
  folder_action: string;
  home_visibility: string;
}

export interface UpdateProfileRequest {
  username: string;
}

/**
 * User registration request
 */
export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  email_code: string;
  captcha_token?: string;
  captcha_id?: string;
  captcha_answer?: string;
  slider_track?: number[];
}

/**
 * User login request
 */
export interface LoginRequest {
  username: string;  // 可以是用户名或邮箱
  password: string;
  captcha_token?: string;
  captcha_id?: string;
  captcha_answer?: string;
  slider_track?: number[];
}

/**
 * Login response with token and user info
 */
export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: UserProfile;
}

/**
 * Change password request
 */
export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface ResetPasswordRequest {
  email: string;
  email_code: string;
  new_password: string;
  captcha_token?: string;
  captcha_id?: string;
  captcha_answer?: string;
  slider_track?: number[];
}

/**
 * Storage quota information
 */
export interface StorageQuota {
  used: number;
  total: number;
  percent: number;
}

import request from './request';

export interface CaptchaSettingsPayload {
  login_enabled: boolean;
  register_enabled: boolean;
  reset_password_enabled: boolean;
  provider: 'image' | 'slider' | 'turnstile' | 'recaptcha';
  security_level: 'balanced' | 'strict' | 'relaxed';
  site_key: string;
  secret_key: string;
  failure_threshold: number;
  cooldown_seconds: number;
  whitelist_paths: string[];
}

export function getCaptchaSettings(): Promise<CaptchaSettingsPayload> {
  return request.get<CaptchaSettingsPayload>('/api/v1/admin/captcha-settings');
}

export function updateCaptchaSettings(data: CaptchaSettingsPayload): Promise<CaptchaSettingsPayload> {
  return request.put<CaptchaSettingsPayload>('/api/v1/admin/captcha-settings', data);
}

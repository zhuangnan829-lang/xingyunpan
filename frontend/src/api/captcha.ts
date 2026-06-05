import request from './request';

export type CaptchaScene = 'login' | 'register' | 'reset_password';
export type CaptchaProvider = 'image' | 'slider' | 'turnstile' | 'recaptcha';

export interface CaptchaPublicConfig {
  enabled: boolean;
  required: boolean;
  provider: CaptchaProvider;
  security_level: string;
  site_key: string;
  failure_threshold: number;
  cooldown_seconds: number;
  whitelist_paths: string[];
}

export interface CaptchaChallenge {
  captcha_id: string;
  provider: CaptchaProvider;
  image_data_url?: string;
  width?: number;
  height?: number;
  expires_in: number;
}

export interface CaptchaPayload {
  captcha_token?: string;
  captcha_id?: string;
  captcha_answer?: string;
  slider_track?: number[];
}

export function getCaptchaConfig(scene: CaptchaScene, identity = '', path = ''): Promise<CaptchaPublicConfig> {
  return request.get('/api/v1/captcha/config', {
    params: { scene, identity, path },
  });
}

export function createCaptchaChallenge(scene: CaptchaScene, identity = '', path = ''): Promise<CaptchaChallenge> {
  return request.post('/api/v1/captcha/challenge', { scene, identity, path });
}

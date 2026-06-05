import request from './request';

export interface SiteSettingsPayload {
  site_name: string;
  tagline: string;
  description: string;
  terms_url: string;
  privacy_url: string;
  primary_url: string;
  backup_urls: string[];
  logo_light: string;
  logo_dark: string;
  favicon: string;
  logo_192: string;
  injection_code: string;
  mobile_guide_enabled: boolean;
  mobile_feedback_url: string;
  desktop_guide_enabled: boolean;
  desktop_community_url: string;
  allow_registration: boolean;
  email_activation: boolean;
  passkey_login_enabled: boolean;
  default_group: string;
  avatar_path: string;
  avatar_size_limit_mb: number;
  avatar_dimension: number;
  gravatar_server: string;
}

export function getSiteSettings(): Promise<SiteSettingsPayload> {
  return request.get<SiteSettingsPayload>('/api/v1/admin/site-settings');
}

export function updateSiteSettings(data: SiteSettingsPayload): Promise<SiteSettingsPayload> {
  return request.put<SiteSettingsPayload>('/api/v1/admin/site-settings', data);
}

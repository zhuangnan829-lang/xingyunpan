import request from './request';

export interface AppearanceThemePalettePayload {
  primary: string;
  secondary: string;
}

export interface AppearanceThemeOptionPayload {
  id: number;
  is_default: boolean;
  light: AppearanceThemePalettePayload;
  dark: AppearanceThemePalettePayload;
}

export interface AppearanceFeatureOptionPayload {
  key: string;
  label: string;
  description: string;
  enabled: boolean;
}

export interface AppearanceSettingPayload {
  theme_options: AppearanceThemeOptionPayload[];
  feature_options: AppearanceFeatureOptionPayload[];
  selected_theme_id: number;
}

export function getAppearanceSettings(): Promise<AppearanceSettingPayload> {
  return request.get<AppearanceSettingPayload>('/api/v1/admin/appearance-settings');
}

export function updateAppearanceSettings(data: AppearanceSettingPayload): Promise<AppearanceSettingPayload> {
  return request.put<AppearanceSettingPayload>('/api/v1/admin/appearance-settings', data);
}

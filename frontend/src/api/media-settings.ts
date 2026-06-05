import request from './request';

export interface MediaSettingsPayload {
  image_mode: 'quality' | 'compatibility';
  image_max_size_gb: number;
  image_quality: number;
  video_preview_second: number;
  video_strategy: 'smooth' | 'balanced' | 'quality';
  metadata_deep_scan: boolean;
  libreoffice_path: string;
}

export function getMediaSettings(): Promise<MediaSettingsPayload> {
  return request.get<MediaSettingsPayload>('/api/v1/admin/media-settings');
}

export function updateMediaSettings(data: MediaSettingsPayload): Promise<MediaSettingsPayload> {
  return request.put<MediaSettingsPayload>('/api/v1/admin/media-settings', data);
}

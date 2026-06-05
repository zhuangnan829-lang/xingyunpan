import request from './request';

export interface FileSystemSettingsPayload {
  online_editor_size: number;
  online_editor_unit: 'B' | 'KB' | 'MB' | 'GB' | 'TB';
  recycle_scan_interval: string;
  blob_recycle_interval: string;
  static_cache_ttl: number;
  list_pagination_mode: 'cursor' | 'offset' | 'hybrid';
  max_page_size: number;
  max_batch_action_size: number;
  max_recursive_search: number;
  map_provider: 'google-leaflet' | 'osm-leaflet' | 'osm-mapbox';
  mime_map: string;
  image_query: string;
  video_query: string;
  audio_query: string;
  document_query: string;
  file_icon_rules: string;
  emoji_options: string;
  browser_apps: string;
  custom_properties: string;
  master_key_storage: 'database' | 'file' | 'env';
  show_encryption_status: boolean;
  enable_event_push: boolean;
  offline_ttl: number;
  debounce_delay: number;
  server_side_download_session_ttl: number;
  upload_session_ttl: number;
  slave_api_sign_ttl: number;
  directory_stat_ttl: number;
  max_chunk_retry: number;
  cache_chunks_for_retry: boolean;
  transfer_parallelism: number;
  oauth_refresh_interval: string;
  wopi_session_ttl: number;
  blob_signed_url_ttl: number;
  blob_signed_url_reuse_ttl: number;
}

export interface BrowserAppGroupItemPayload {
  id: number;
  icon: string;
  icon_url?: string;
  accent?: string;
  type?: string;
  name: string;
  extensions: string;
  platform?: string;
  create_mapping?: string;
  enabled?: boolean;
  open_in_new_window?: boolean;
  max_size?: number;
  max_size_unit?: string;
}

export interface BrowserAppGroupPayload {
  id: number;
  name: string;
  description?: string;
  items: BrowserAppGroupItemPayload[];
}

export interface FileSystemClientSettingsPayload {
  online_editor_size: number;
  online_editor_unit: 'B' | 'KB' | 'MB' | 'GB' | 'TB';
  max_page_size: number;
  max_chunk_retry: number;
  cache_chunks_for_retry: boolean;
  transfer_parallelism: number;
  max_batch_action_size: number;
  map_provider: 'google-leaflet' | 'osm-leaflet' | 'osm-mapbox';
  show_encryption_status: boolean;
  enable_event_push: boolean;
  debounce_delay: number;
  file_icon_rules: string;
  emoji_options: string;
}

export interface FileSystemIconSettingsPayload {
  file_icon_rules: string;
  emoji_options: string;
}

export type MapProvider = 'google-leaflet' | 'osm-leaflet' | 'osm-mapbox';

export interface FileSystemMapRuntimeConfig {
  provider: MapProvider;
  engine: 'leaflet' | 'mapbox';
  tile_url?: string;
  style_url?: string;
  attribution: string;
  requires_token: boolean;
  token_missing?: boolean;
}

export function getFileSystemSettings(): Promise<FileSystemSettingsPayload> {
  return request.get<FileSystemSettingsPayload>('/api/v1/admin/file-system-settings');
}

export function getFileSystemClientSettings(): Promise<FileSystemClientSettingsPayload> {
  return request.get<FileSystemClientSettingsPayload>('/api/v1/file-system-settings/client');
}

export function getFileSystemMapProvider(): Promise<FileSystemMapRuntimeConfig> {
  return request.get<FileSystemMapRuntimeConfig>('/api/v1/file-system-settings/map-provider');
}

export function getFileSystemBrowserApps(): Promise<BrowserAppGroupPayload[]> {
  return request.get<BrowserAppGroupPayload[]>('/api/v1/admin/file-system-settings/browser-apps');
}

export function updateFileSystemSettings(data: FileSystemSettingsPayload): Promise<FileSystemSettingsPayload> {
  return request.put<FileSystemSettingsPayload>('/api/v1/admin/file-system-settings', data);
}

export function updateFileSystemIconSettings(data: FileSystemIconSettingsPayload): Promise<FileSystemSettingsPayload> {
  return request.patch<FileSystemSettingsPayload>('/api/v1/admin/file-system-settings/icons', data);
}

export function clearFileSystemBlobUrlCache(): Promise<{ cleared: boolean }> {
  return request.post<{ cleared: boolean }>('/api/v1/admin/file-system-settings/blob-url-cache/clear');
}

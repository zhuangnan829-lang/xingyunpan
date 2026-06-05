import request from './request';

export interface FullTextSearchSettingsPayload {
  enabled: boolean;
  meili_endpoint: string;
  api_key: string;
  result_page_size: number;
  ai_search: boolean;
  tika_endpoint: string;
  extensions: string;
  max_file_size: number;
  max_file_size_unit: 'B' | 'KB' | 'MB' | 'GB' | 'TB';
  chunk_size: number;
  chunk_unit: 'B' | 'KB' | 'MB';
  index_notes: string;
}

export interface FullTextSearchRebuildResponse {
  job_id: number;
  queue_key: string;
  status: string;
  resource_id: string;
  scheduled_at: number;
}

export function getFullTextSearchSettings(): Promise<FullTextSearchSettingsPayload> {
  return request.get<FullTextSearchSettingsPayload>('/api/v1/admin/full-text-search-settings');
}

export function updateFullTextSearchSettings(data: FullTextSearchSettingsPayload): Promise<FullTextSearchSettingsPayload> {
  return request.put<FullTextSearchSettingsPayload>('/api/v1/admin/full-text-search-settings', data);
}

export function rebuildFullTextSearchIndex(): Promise<FullTextSearchRebuildResponse> {
  return request.post<FullTextSearchRebuildResponse>('/api/v1/admin/full-text-search-settings/rebuild-index');
}

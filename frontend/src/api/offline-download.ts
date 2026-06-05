import request from './request';

export type OfflineDownloadStatus = 'queued' | 'downloading' | 'paused' | 'completed' | 'failed';

export interface OfflineDownloadTask {
  id: number;
  task_token: string;
  name: string;
  url: string;
  save_path: string;
  status: OfflineDownloadStatus;
  progress: number;
  speed: string;
  size: string;
  downloaded_bytes: number;
  total_bytes: number;
  error_message?: string;
  queue_job_id?: number | null;
  saved_file_id?: number | null;
  saved_folder_id?: number | null;
  created_at: string;
  updated_at: string;
  completed_at?: string | null;
}

export interface OfflineDownloadCreatePayload {
  url: string;
  name?: string;
  save_path?: string;
}

export function listOfflineDownloads(params?: { status?: string; keyword?: string }): Promise<OfflineDownloadTask[]> {
  return request.get('/api/v1/offline-downloads', { params });
}

export function createOfflineDownload(payload: OfflineDownloadCreatePayload): Promise<OfflineDownloadTask> {
  return request.post('/api/v1/offline-downloads', payload);
}

export function refreshOfflineDownloads(): Promise<OfflineDownloadTask[]> {
  return request.post('/api/v1/offline-downloads/refresh');
}

export function pauseOfflineDownload(id: number): Promise<OfflineDownloadTask> {
  return request.post(`/api/v1/offline-downloads/${id}/pause`);
}

export function resumeOfflineDownload(id: number): Promise<OfflineDownloadTask> {
  return request.post(`/api/v1/offline-downloads/${id}/resume`);
}

export function retryOfflineDownload(id: number): Promise<OfflineDownloadTask> {
  return request.post(`/api/v1/offline-downloads/${id}/retry`);
}

export function deleteOfflineDownload(id: number): Promise<{ deleted: boolean }> {
  return request.delete(`/api/v1/offline-downloads/${id}`);
}

export function batchDeleteOfflineDownloads(ids: number[]): Promise<{ deleted: number }> {
  return request.post('/api/v1/offline-downloads/batch-delete', { ids });
}

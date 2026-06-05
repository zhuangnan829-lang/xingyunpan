import request from './request';

export interface AdminSharePayload {
  share_id: number;
  share_token: string;
  share_url: string;
  file_ids: string[];
  file_names: string[];
  owner_id: number;
  owner_username: string;
  owner_email: string;
  created_at: string;
  expires_at: string | null;
  has_password: boolean;
  max_downloads: number | null;
  download_count: number;
  access_count: number;
  is_expired: boolean;
  is_unavailable: boolean;
  status_reason: 'active' | 'time_expired' | 'download_limit_reached' | string;
}

export interface AdminShareListParams {
  page?: number;
  page_size?: number;
  cursor?: string;
  keyword?: string;
  owner_id?: number;
  status?: 'all' | 'active' | 'expired' | 'protected' | 'unavailable' | 'download_limit_reached';
  min_downloads?: number;
  expiring_within_days?: number;
  max_downloads_reached?: boolean;
  unavailable?: boolean;
}

export interface AdminShareListResponse {
  list: AdminSharePayload[];
  total: number;
  page: number;
  page_size: number;
  pagination_mode?: string;
  next_cursor?: string;
  max_page_size?: number;
}

export interface AdminShareMetricsPayload {
  total_shares: number;
  active_shares: number;
  expired_shares: number;
  protected_shares: number;
  total_access_count: number;
  total_download_count: number;
  expiring_soon_count: number;
  download_limit_reached_count: number;
}

export function listAdminShares(params: AdminShareListParams = {}): Promise<AdminShareListResponse> {
  return request.get<AdminShareListResponse>('/api/v1/admin/shares', { params });
}

export function deleteAdminShare(id: number): Promise<{ deleted: boolean }> {
  return request.delete<{ deleted: boolean }>(`/api/v1/admin/shares/${id}`);
}

export function batchDeleteAdminShares(shareIds: number[]): Promise<{ deleted: boolean }> {
  return request.post<{ deleted: boolean }>('/api/v1/admin/shares/batch-delete', {
    share_ids: shareIds,
  });
}

export function getAdminShareMetrics(expiringWithinDays = 3): Promise<AdminShareMetricsPayload> {
  return request.get<AdminShareMetricsPayload>('/api/v1/admin/shares/metrics', {
    params: { expiring_within_days: expiringWithinDays },
  });
}

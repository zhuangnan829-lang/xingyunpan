import request from './request';

export interface AdminFilePayload {
  id: number;
  file_name: string;
  is_folder: boolean;
  file_size: number;
  occupied_size: number;
  file_path: string;
  owner_id: number;
  owner_username: string;
  owner_email: string;
  storage_policy_id: number;
  storage_policy_name: string;
  physical_file_id: number | null;
  display_icon?: string;
  display_icon_tint?: string;
  display_icon_label?: string;
  display_icon_source?: string;
  has_share_link: boolean;
  has_direct_link: boolean;
  uploading: boolean;
  created_at: string;
  updated_at: string;
}

export interface AdminFileListParams {
  page?: number;
  page_size?: number;
  owner_id?: number;
  keyword?: string;
  storage_policy_id?: number;
}

export interface AdminFileListResponse {
  list: AdminFilePayload[];
  total: number;
  page: number;
  page_size: number;
}

export function listAdminFiles(params: AdminFileListParams = {}): Promise<AdminFileListResponse> {
  return request.get<AdminFileListResponse>('/api/v1/admin/files', { params });
}

export function deleteAdminFile(id: number): Promise<{ deleted: boolean }> {
  return request.delete<{ deleted: boolean }>(`/api/v1/admin/files/${id}`);
}

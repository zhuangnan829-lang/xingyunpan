import request from './request';

export interface AdminUserPayload {
  id: number;
  username: string;
  email: string;
  role: 'admin' | 'user';
  enabled: boolean;
  user_group_id: number;
  user_group_name: string;
  capacity: number;
  used_size: number;
  created_at: string;
  updated_at: string;
}

export interface AdminUserUpsertPayload {
  username: string;
  email: string;
  password?: string;
  role: 'admin' | 'user';
  enabled: boolean;
  user_group_id: number;
  capacity: number;
  follow_group_capacity: boolean;
}

export interface AdminUserBatchDeletePayload {
  ids: number[];
}

export interface AdminUserDeletePreviewPayload {
  user_id: number;
  file_count: number;
  share_count: number;
  recycle_bin_count: number;
  offline_download_task_count: number;
  multipart_upload_count: number;
  dav_account_count: number;
  oauth_credential_count: number;
  collaboration_owned_count: number;
  collaboration_shared_count: number;
  traffic_event_count: number;
  user_preference_count: number;
  file_custom_property_count: number;
  storage_policy_hit_log_count: number;
  used_size: number;
  has_blocking_assets: boolean;
}

export interface AdminUserBatchGroupPayload {
  ids: number[];
  user_group_id: number;
}

export interface AdminUserBatchRolePayload {
  ids: number[];
  role: 'admin' | 'user';
}

export interface AdminUserStatusPayload {
  enabled: boolean;
}

export interface AdminUserBatchStatusPayload {
  ids: number[];
  enabled: boolean;
}

export interface ListAdminUsersParams {
  keyword?: string;
  role?: 'all' | 'admin' | 'user';
  status?: 'all' | 'enabled' | 'disabled';
  user_group_id?: number;
  user_group_name?: string;
  page?: number;
  page_size?: number;
}

export interface AdminUserListResponse {
  items: AdminUserPayload[];
  total: number;
  page: number;
  page_size: number;
}

interface AdminUserObjectListResponse {
  items?: AdminUserPayload[];
  list?: AdminUserPayload[];
  total?: number;
  page?: number;
  page_size?: number;
}

type AdminUserListRawResponse = AdminUserObjectListResponse | AdminUserPayload[];

export async function listAdminUsers(params?: ListAdminUsersParams): Promise<AdminUserListResponse> {
  const data = await request.get<AdminUserListRawResponse>('/api/v1/admin/users', { params });
  const requestedPage = params?.page || 1;
  const requestedPageSize = params?.page_size || 10;

  if (Array.isArray(data)) {
    return {
      items: data,
      total: data.length,
      page: requestedPage,
      page_size: requestedPageSize,
    };
  }

  const items = Array.isArray(data.items) ? data.items : Array.isArray(data.list) ? data.list : [];

  return {
    items,
    total: data.total ?? items.length,
    page: data.page ?? requestedPage,
    page_size: data.page_size ?? requestedPageSize,
  };
}

export function createAdminUser(data: AdminUserUpsertPayload): Promise<AdminUserPayload> {
  return request.post<AdminUserPayload>('/api/v1/admin/users', data);
}

export function updateAdminUser(id: number, data: AdminUserUpsertPayload): Promise<AdminUserPayload> {
  return request.put<AdminUserPayload>(`/api/v1/admin/users/${id}`, data);
}

export function updateAdminUserStatus(id: number, data: AdminUserStatusPayload): Promise<AdminUserPayload> {
  return request.put<AdminUserPayload>(`/api/v1/admin/users/${id}/status`, data);
}

export function batchUpdateAdminUserStatus(data: AdminUserBatchStatusPayload): Promise<AdminUserPayload[]> {
  return request.post<AdminUserPayload[]>('/api/v1/admin/users/batch-status', data);
}

export function deleteAdminUser(id: number): Promise<{ deleted: boolean }> {
  return request.delete<{ deleted: boolean }>(`/api/v1/admin/users/${id}`);
}

export function getAdminUserDeletePreview(id: number): Promise<AdminUserDeletePreviewPayload> {
  return request.get<AdminUserDeletePreviewPayload>(`/api/v1/admin/users/${id}/delete-preview`);
}

export function resetAdminUserPassword(id: number, password: string): Promise<{ updated: boolean }> {
  return request.post<{ updated: boolean }>(`/api/v1/admin/users/${id}/reset-password`, { password });
}

export function batchDeleteAdminUsers(data: AdminUserBatchDeletePayload): Promise<{ deleted: boolean }> {
  return request.post<{ deleted: boolean }>('/api/v1/admin/users/batch-delete', data);
}

export function batchUpdateAdminUsersGroup(data: AdminUserBatchGroupPayload): Promise<AdminUserPayload[]> {
  return request.post<AdminUserPayload[]>('/api/v1/admin/users/batch-group', data);
}

export function batchUpdateAdminUsersRole(data: AdminUserBatchRolePayload): Promise<AdminUserPayload[]> {
  return request.post<AdminUserPayload[]>('/api/v1/admin/users/batch-role', data);
}

import request from './request';

export interface UserGroupPayload {
  id?: number;
  name: string;
  description: string;
  storage_policy_id: number;
  storage_policy_name: string;
  user_count: number;
  max_capacity: number;
  sync_member_capacity?: boolean;
}

export interface UserGroupMemberPayload {
  id: number;
  username: string;
  email: string;
  role: string;
  capacity: number;
  used_size: number;
  user_group_id: number;
}

export interface UserGroupSummaryPayload {
  total_groups: number;
  total_users: number;
  default_group: string;
  high_capacity_groups: number;
  unlimited_groups: number;
  high_capacity_bytes: number;
}

export function listUserGroups(): Promise<UserGroupPayload[]> {
  return request.get<UserGroupPayload[]>('/api/v1/admin/user-groups');
}

export function getUserGroupSummary(): Promise<UserGroupSummaryPayload> {
  return request.get<UserGroupSummaryPayload>('/api/v1/admin/user-groups/summary');
}

export function createUserGroup(data: Partial<UserGroupPayload>): Promise<UserGroupPayload> {
  return request.post<UserGroupPayload>('/api/v1/admin/user-groups', data);
}

export function updateUserGroup(id: number, data: Partial<UserGroupPayload>): Promise<UserGroupPayload> {
  return request.put<UserGroupPayload>(`/api/v1/admin/user-groups/${id}`, data);
}

export function deleteUserGroup(id: number): Promise<{ deleted: boolean }> {
  return request.delete<{ deleted: boolean }>(`/api/v1/admin/user-groups/${id}`);
}

export function listUserGroupMembers(id: number): Promise<UserGroupMemberPayload[]> {
  return request.get<UserGroupMemberPayload[]>(`/api/v1/admin/user-groups/${id}/users`);
}

import request from './request';

export type StoragePolicyType = 'local' | 'remote' | 'oss' | 'onedrive' | 'cos' | 's3' | 'obs' | 'balance';

export interface StoragePolicyPayload {
  id?: number;
  name: string;
  type: StoragePolicyType;
  groups: string[];
  effective_user_count?: number;
  effective_file_count?: number;
  blob_path: string;
  blob_name_pattern: string;
  max_file_size: number;
  max_file_size_unit: 'KB' | 'MB' | 'GB';
  extension_mode: 'allow' | 'deny';
  extensions: string;
  name_rule_mode: 'allow' | 'deny';
  name_regex: string;
  chunk_size: number;
  chunk_size_unit: 'KB' | 'MB' | 'GB';
  pre_allocate: boolean;
  parallel_chunk_count: number;
  enable_cdn: boolean;
  download_cdn: string;
  enable_encryption: boolean;
  encryption_key_id: string;
}

export interface StoragePolicyGroupCoverage {
  id: number;
  name: string;
  user_count: number;
  file_count: number;
}

export interface StoragePolicyUploadPreview {
  max_file_size: number;
  max_file_size_unit: 'KB' | 'MB' | 'GB';
  extension_mode: 'allow' | 'deny';
  extensions: string;
  name_rule_mode: 'allow' | 'deny';
  name_regex: string;
  chunk_size: number;
  chunk_size_unit: 'KB' | 'MB' | 'GB';
  parallel_chunk_count: number;
  pre_allocate: boolean;
  enable_cdn: boolean;
  download_cdn: string;
  enable_encryption: boolean;
  encryption_key_id: string;
}

export interface StoragePolicyPreviewPayload {
  policy: StoragePolicyPayload;
  groups: StoragePolicyGroupCoverage[];
  user_count: number;
  existing_file_count: number;
  new_upload_config: StoragePolicyUploadPreview;
  historical_blob_note: string;
}

export interface StoragePolicyAuditPayload {
  id: number;
  storage_policy_id: number;
  action: 'create' | 'update' | 'delete' | string;
  operator_id: number;
  operator_name: string;
  source_audit_id: number;
  before?: StoragePolicyPayload;
  after?: StoragePolicyPayload;
  groups: StoragePolicyGroupCoverage[];
  user_count: number;
  created_at: string;
}

export interface StoragePolicyGroupMigrationPayload {
  source_policy_id: number;
  target_policy_id: number;
  groups: StoragePolicyGroupCoverage[];
  user_count: number;
}

export interface StoragePolicyHitLogPayload {
  id: number;
  storage_policy_id: number;
  storage_policy_name: string;
  hit_type: 'user_group_policy' | 'global_default' | string;
  action: 'upload' | 'multipart_init' | 'multipart_complete' | 'download' | 'preview' | 'share_download' | string;
  user_id: number;
  username: string;
  user_group_id: number;
  user_group_name: string;
  file_id: number;
  file_name: string;
  file_size: number;
  resource_type: string;
  resource_id: string;
  config: Record<string, unknown>;
  created_at: string;
}

export function listStoragePolicies(): Promise<StoragePolicyPayload[]> {
  return request.get<StoragePolicyPayload[]>('/api/v1/admin/storage-policies');
}

export function getStoragePolicy(id: number): Promise<StoragePolicyPayload> {
  return request.get<StoragePolicyPayload>(`/api/v1/admin/storage-policies/${id}`);
}

export function previewStoragePolicy(id: number): Promise<StoragePolicyPreviewPayload> {
  return request.get<StoragePolicyPreviewPayload>(`/api/v1/admin/storage-policies/${id}/preview`);
}

export function listStoragePolicyAudits(id: number, limit = 10): Promise<StoragePolicyAuditPayload[]> {
  return request.get<StoragePolicyAuditPayload[]>(`/api/v1/admin/storage-policies/${id}/audits`, {
    params: { limit },
  });
}

export function listStoragePolicyHits(id: number, limit = 20): Promise<StoragePolicyHitLogPayload[]> {
  return request.get<StoragePolicyHitLogPayload[]>(`/api/v1/admin/storage-policies/${id}/hits`, {
    params: { limit },
  });
}

export function getStoragePolicyAudit(id: number, auditId: number): Promise<StoragePolicyAuditPayload> {
  return request.get<StoragePolicyAuditPayload>(`/api/v1/admin/storage-policies/${id}/audits/${auditId}`);
}

export function rollbackStoragePolicy(id: number, auditId: number): Promise<StoragePolicyPayload> {
  return request.post<StoragePolicyPayload>(`/api/v1/admin/storage-policies/${id}/audits/${auditId}/rollback`);
}

export function migrateStoragePolicyGroups(
  id: number,
  data: { target_policy_id: number; group_ids?: number[] },
): Promise<StoragePolicyGroupMigrationPayload> {
  return request.post<StoragePolicyGroupMigrationPayload>(`/api/v1/admin/storage-policies/${id}/migrate-groups`, data);
}

export function repairLegacyStoragePolicies(): Promise<StoragePolicyAuditPayload[]> {
  return request.post<StoragePolicyAuditPayload[]>('/api/v1/admin/storage-policies/repair-legacy');
}

export function copyStoragePolicy(id: number): Promise<StoragePolicyPayload> {
  return request.post<StoragePolicyPayload>(`/api/v1/admin/storage-policies/${id}/copy`);
}

export function exportStoragePolicy(id: number): Promise<StoragePolicyPayload> {
  return request.get<StoragePolicyPayload>(`/api/v1/admin/storage-policies/${id}/export`);
}

export function importStoragePolicy(data: StoragePolicyPayload, overwriteId?: number): Promise<StoragePolicyPayload> {
  return request.post<StoragePolicyPayload>('/api/v1/admin/storage-policies/import', data, {
    params: overwriteId ? { overwrite_id: overwriteId } : undefined,
  });
}

export function createStoragePolicy(data: StoragePolicyPayload): Promise<StoragePolicyPayload> {
  return request.post<StoragePolicyPayload>('/api/v1/admin/storage-policies', data);
}

export function updateStoragePolicy(id: number, data: StoragePolicyPayload): Promise<StoragePolicyPayload> {
  return request.put<StoragePolicyPayload>(`/api/v1/admin/storage-policies/${id}`, data);
}

export function deleteStoragePolicy(id: number): Promise<{ deleted: boolean }> {
  return request.delete<{ deleted: boolean }>(`/api/v1/admin/storage-policies/${id}`);
}

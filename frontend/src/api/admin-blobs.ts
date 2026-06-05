import request from './request';
import type { BlobRecord, BlobKind } from '@/views/admin/file-blobs/types';

interface AdminBlobLinkedFileResponse {
  id: number;
  name: string;
  extension: string;
  size_bytes: number;
  owner_id: number;
  owner_name: string;
  created_at: string;
}

interface AdminBlobLinkedVersionResponse {
  id: number;
  file_id: number;
  file_name: string;
  version_number: number;
  size_bytes: number;
  owner_id: number;
  owner_name: string;
  created_at: string;
}

interface AdminBlobReferenceSourceResponse {
  type: string;
  id: string;
  name: string;
}

interface AdminBlobResponse {
  id: number;
  kind: string;
  kind_label: string;
  source: string;
  hash: string;
  content_type: string;
  size_bytes: number;
  reference_count: number;
  stored_reference_count: number;
  created_at: string;
  updated_at: string;
  creator_id: number;
  creator_name: string;
  creator_badge: string;
  storage_policy_id: number;
  storage_policy_name: string;
  storage_policy_subtitle: string;
  encrypted: boolean;
  locked: boolean;
  locked_reason: string;
  can_delete: boolean;
  delete_blocked_reasons: string[];
  missing_on_storage: boolean;
  health_status: string;
  upload_session_id?: string;
  linked_files: AdminBlobLinkedFileResponse[];
  linked_versions?: AdminBlobLinkedVersionResponse[];
  reference_sources?: AdminBlobReferenceSourceResponse[];
}

export interface AdminBlobListParams {
  page?: number;
  page_size?: number;
  owner_id?: number;
  kind?: string;
  storage_policy_id?: number;
  keyword?: string;
  min_size?: number;
  max_size?: number;
  ref_count_min?: number;
  ref_count_max?: number;
  encrypted?: boolean;
  created_from?: string;
  created_to?: string;
  sort_by?: 'id' | 'size' | 'reference_count' | 'created_at';
  sort_order?: 'asc' | 'desc';
}

export interface AdminBlobListPage {
  list: BlobRecord[];
  total: number;
  page: number;
  pageSize: number;
  totalSize: number;
  referenceTotal: number;
  encryptedCount: number;
  orphanCount: number;
}

interface AdminBlobListResponse {
  list: AdminBlobResponse[];
  total: number;
  page: number;
  page_size: number;
  total_size: number;
  reference_total: number;
  encrypted_count: number;
  orphan_count: number;
}

export interface AdminBlobBatchDeleteResponse {
  deleted: Array<{ id: number; reason?: string }>;
  skipped: Array<{ id: number; reason?: string }>;
  failed: Array<{ id: number; reason?: string }>;
}

export interface AdminBlobScanTask {
  id: number;
  status: string;
  progress: number;
  total_physical_files: number;
  scanned_physical_files: number;
  storage_file_count: number;
  orphan_count: number;
  missing_on_storage: number;
  ref_count_mismatch: number;
  duplicate_hash: number;
  zero_size: number;
  invalid_path: number;
  extra_storage_files: number;
  issues: Array<{ blob_id?: number; path?: string; hash?: string; type: string; reason: string }>;
}

function normalizeKind(kind: string): BlobKind {
  if (kind === 'version' || kind === 'thumbnail' || kind === 'live-photo' || kind === 'file' || kind === 'orphan') {
    return kind;
  }
  return 'file';
}

function mapBlob(response: AdminBlobResponse): BlobRecord {
  return {
    id: response.id,
    kind: normalizeKind(response.kind),
    kindLabel: response.kind_label,
    source: response.source,
    hash: response.hash,
    contentType: response.content_type,
    sizeBytes: response.size_bytes,
    referenceCount: response.reference_count,
    storedReferenceCount: response.stored_reference_count,
    createdAt: response.created_at,
    updatedAt: response.updated_at,
    creatorId: response.creator_id,
    creatorName: response.creator_name,
    creatorBadge: response.creator_badge,
    storagePolicyId: response.storage_policy_id,
    storagePolicyName: response.storage_policy_name,
    storagePolicySubtitle: response.storage_policy_subtitle,
    encrypted: response.encrypted,
    locked: response.locked,
    lockedReason: response.locked_reason,
    canDelete: response.can_delete,
    deleteBlockedReasons: response.delete_blocked_reasons || [],
    missingOnStorage: response.missing_on_storage,
    healthStatus: response.health_status,
    uploadSessionId: response.upload_session_id,
    linkedFiles: response.linked_files.map((file) => ({
      id: file.id,
      name: file.name,
      extension: file.extension,
      sizeBytes: file.size_bytes,
      ownerId: file.owner_id,
      ownerName: file.owner_name,
      createdAt: file.created_at,
    })),
    linkedVersions: (response.linked_versions || []).map((version) => ({
      id: version.id,
      fileId: version.file_id,
      fileName: version.file_name,
      versionNumber: version.version_number,
      sizeBytes: version.size_bytes,
      ownerId: version.owner_id,
      ownerName: version.owner_name,
      createdAt: version.created_at,
    })),
    referenceSources: (response.reference_sources || []).map((source) => ({
      type: source.type,
      id: source.id,
      name: source.name,
    })),
  };
}

export async function listAdminBlobs(params: AdminBlobListParams = {}): Promise<BlobRecord[]> {
  const response = await listAdminBlobsPage(params);
  return response.list;
}

export async function listAdminBlobsPage(params: AdminBlobListParams = {}): Promise<AdminBlobListPage> {
  const response = await request.get<AdminBlobListResponse | AdminBlobResponse[]>('/api/v1/admin/blobs', { params });
  if (Array.isArray(response)) {
    return {
      list: response.map(mapBlob),
      total: response.length,
      page: 1,
      pageSize: response.length,
      totalSize: response.reduce((sum, item) => sum + item.size_bytes, 0),
      referenceTotal: response.reduce((sum, item) => sum + item.reference_count, 0),
      encryptedCount: response.filter((item) => item.encrypted).length,
      orphanCount: response.filter((item) => item.kind === 'orphan').length,
    };
  }
  return {
    list: response.list.map(mapBlob),
    total: response.total,
    page: response.page,
    pageSize: response.page_size,
    totalSize: response.total_size,
    referenceTotal: response.reference_total,
    encryptedCount: response.encrypted_count,
    orphanCount: response.orphan_count,
  };
}

export async function getAdminBlob(id: number): Promise<BlobRecord> {
  const response = await request.get<AdminBlobResponse>(`/api/v1/admin/blobs/${id}`);
  return mapBlob(response);
}

export async function deleteAdminBlob(id: number): Promise<{ deleted: boolean }> {
  return request.delete<{ deleted: boolean }>(`/api/v1/admin/blobs/${id}`);
}

export async function batchDeleteAdminBlobs(ids: number[]): Promise<AdminBlobBatchDeleteResponse> {
  return request.post<AdminBlobBatchDeleteResponse>('/api/v1/admin/blobs/batch-delete', { ids });
}

export async function lockAdminBlob(id: number, reason = ''): Promise<BlobRecord> {
  const response = await request.post<AdminBlobResponse>(`/api/v1/admin/blobs/${id}/lock`, { reason });
  return mapBlob(response);
}

export async function unlockAdminBlob(id: number): Promise<BlobRecord> {
  const response = await request.post<AdminBlobResponse>(`/api/v1/admin/blobs/${id}/unlock`);
  return mapBlob(response);
}

export function scanAdminBlobs(): Promise<AdminBlobScanTask> {
  return request.post<AdminBlobScanTask>('/api/v1/admin/blobs/scan');
}

export function getLatestAdminBlobScan(): Promise<AdminBlobScanTask> {
  return request.get<AdminBlobScanTask>('/api/v1/admin/blobs/scan/latest');
}

export async function downloadAdminBlob(id: number): Promise<Blob> {
  return request.get(`/api/v1/admin/blobs/${id}/download`, {
    responseType: 'blob',
  });
}

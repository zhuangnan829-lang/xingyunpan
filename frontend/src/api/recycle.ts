// Recycle bin API module for soft delete and restore functionality

import request from './request';

/**
 * Interface for recycle bin item
 */
export interface RecycleItem {
  id: string;
  file_id: string;
  file_name: string;
  file_size: number;
  file_type: string;
  original_path: string;
  deleted_at: string;
  expires_at: string;
}

/**
 * Response interface for recycle bin list
 */
export interface RecycleBinResponse {
  items: RecycleItem[];
  total: number;
  page?: number;
  page_size?: number;
  total_pages?: number;
  stats?: RecycleStats;
  sort?: string;
  keyword?: string;
  file_type?: string;
  total_size?: number;
}

interface RecyclePageResponse {
  list?: RecycleItem[];
  items?: RecycleItem[];
  total?: number;
  page?: number;
  page_size?: number;
  total_pages?: number;
  stats?: RecycleStats;
  sort?: string;
  keyword?: string;
  file_type?: string;
  total_size?: number;
}

export interface RecycleStats {
  total_size: number;
  expiring_soon: number;
  expired: number;
  count_by_type: Record<string, number>;
}

export interface RecycleListParams {
  sort?: string;
  keyword?: string;
  file_type?: string;
}

function toNumericIds(ids: Array<string | number>): number[] {
  return ids
    .map((id) => Number(id))
    .filter((id) => Number.isInteger(id) && id > 0);
}

/**
 * Move files to recycle bin (soft delete)
 * @param fileIds - Array of file IDs to delete
 */
export async function moveToRecycleBin(fileIds: Array<string | number>): Promise<void> {
  return request.post('/api/v1/recycle', { file_ids: toNumericIds(fileIds) });
}

/**
 * Get recycle bin items with pagination
 * @param page - Page number (default: 1)
 * @param pageSize - Items per page (default: 20)
 * @returns Recycle bin items and total count
 */
export async function getRecycleBinItems(
  page: number = 1,
  pageSize: number = 20,
  params: RecycleListParams = {}
): Promise<RecycleBinResponse> {
  const response = await request.get<RecyclePageResponse>('/api/v1/recycle', {
    params: {
      page,
      page_size: pageSize,
      sort: params.sort,
      keyword: params.keyword,
      file_type: params.file_type,
    }
  });

  return {
    items: response.items || response.list || [],
    total: response.total || 0,
    page: response.page,
    page_size: response.page_size,
    total_pages: response.total_pages,
    stats: response.stats,
    sort: response.sort,
    keyword: response.keyword,
    file_type: response.file_type,
    total_size: response.total_size,
  };
}

/**
 * Restore files from recycle bin
 * @param itemIds - Array of recycle item IDs to restore
 */
export async function restoreFiles(itemIds: Array<string | number>): Promise<void> {
  return request.post('/api/v1/recycle/restore', { item_ids: toNumericIds(itemIds) });
}

/**
 * Permanently delete files from recycle bin
 * @param itemIds - Array of recycle item IDs to permanently delete
 */
export async function permanentDelete(itemIds: Array<string | number>): Promise<void> {
  return request.delete('/api/v1/recycle', { data: { item_ids: toNumericIds(itemIds) } });
}

/**
 * Empty recycle bin (permanently delete all items)
 */
export async function emptyRecycleBin(): Promise<void> {
  return request.delete('/api/v1/recycle/all');
}

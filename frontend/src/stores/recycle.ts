/**
 * Recycle store - Pinia state management for recycle bin
 * Manages deleted files, restoration, permanent deletion, and auto-cleanup
 */

import { defineStore } from 'pinia';
import { ref, type Ref } from 'vue';
import {
  moveToRecycleBin as apiMoveToRecycleBin,
  getRecycleBinItems as apiGetRecycleBinItems,
  restoreFiles as apiRestoreFiles,
  permanentDelete as apiPermanentDelete,
  emptyRecycleBin as apiEmptyRecycleBin,
  type RecycleItem as ApiRecycleItem,
  type RecycleBinResponse,
  type RecycleListParams,
  type RecycleStats,
} from '@/api/recycle';
import type { RecycleItem } from '@/types/recycle';
import {
  validateItemIds,
  handleRestoreError,
  handlePermanentDeleteError,
  handleEmptyRecycleBinError,
} from '@/utils/recycle-validation';

function mapRecycleItem(item: ApiRecycleItem): RecycleItem {
  return {
    id: String(item.id),
    fileId: String(item.file_id),
    fileName: item.file_name,
    fileSize: item.file_size,
    fileType: item.file_type,
    originalPath: item.original_path,
    deletedAt: item.deleted_at,
    expiresAt: item.expires_at,
    userId: '',
  };
}

export const useRecycleStore = defineStore('recycle', () => {
  // State
  const items: Ref<RecycleItem[]> = ref([]);
  const total: Ref<number> = ref(0);
  const stats: Ref<RecycleStats> = ref({
    total_size: 0,
    expiring_soon: 0,
    expired: 0,
    count_by_type: {},
  });
  const isLoading: Ref<boolean> = ref(false);
  const currentPage: Ref<number> = ref(1);
  const pageSize: Ref<number> = ref(20);

  // Actions

  /**
   * Load recycle bin items with pagination
   * @param page - Page number (default: 1)
   * @param size - Items per page (default: 20)
   */
  async function loadItems(page: number = 1, size: number = 20, params: RecycleListParams = {}): Promise<void> {
    isLoading.value = true;
    currentPage.value = page;
    pageSize.value = size;

    try {
      const response: RecycleBinResponse = await apiGetRecycleBinItems(page, size, params);
      items.value = response.items.map(mapRecycleItem);
      total.value = response.total;
      stats.value = response.stats || {
        total_size: response.total_size || items.value.reduce((sum, item) => sum + item.fileSize, 0),
        expiring_soon: 0,
        expired: 0,
        count_by_type: {},
      };
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Move files to recycle bin (soft delete) - optimized for batch operations
   * Requirements: 6.6-6.7 - Batch operations optimization
   * @param fileIds - Array of file IDs to delete
   */
  async function moveToRecycle(fileIds: string[]): Promise<void> {
    if (fileIds.length === 0) return;

    isLoading.value = true;
    try {
      // Batch API call
      await apiMoveToRecycleBin(fileIds);
      
      // Reload current page to reflect changes
      await loadItems(currentPage.value, pageSize.value);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Restore files from recycle bin (optimized for batch operations)
   * Requirements: 6.6-6.7 - Batch operations optimization
   * @param itemIds - Array of recycle item IDs to restore
   */
  async function restoreFiles(itemIds: string[]): Promise<void> {
    if (itemIds.length === 0) return;

    isLoading.value = true;
    try {
      // Validate item IDs
      validateItemIds(itemIds);

      // Batch API call
      await apiRestoreFiles(itemIds);
      
      // Optimized local state update
      const idsSet = new Set(itemIds);
      const removedSize = items.value
        .filter(item => idsSet.has(item.id))
        .reduce((sum, item) => sum + item.fileSize, 0);
      items.value = items.value.filter(item => !idsSet.has(item.id));
      total.value -= itemIds.length;
      stats.value.total_size = Math.max(0, stats.value.total_size - removedSize);
      
      // Reload if current page is now empty and not the first page
      if (items.value.length === 0 && currentPage.value > 1) {
        await loadItems(currentPage.value - 1, pageSize.value);
      } else if (items.value.length === 0 && currentPage.value === 1) {
        // If first page is empty, reload to check if there are more items
        await loadItems(1, pageSize.value);
      }
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleRestoreError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Permanently delete files from recycle bin (optimized for batch operations)
   * Requirements: 6.6-6.7 - Batch operations optimization
   * @param itemIds - Array of recycle item IDs to permanently delete
   */
  async function permanentDelete(itemIds: string[]): Promise<void> {
    if (itemIds.length === 0) return;

    isLoading.value = true;
    try {
      // Validate item IDs
      validateItemIds(itemIds);

      // Batch API call
      await apiPermanentDelete(itemIds);
      
      // Optimized local state update
      const idsSet = new Set(itemIds);
      const removedSize = items.value
        .filter(item => idsSet.has(item.id))
        .reduce((sum, item) => sum + item.fileSize, 0);
      items.value = items.value.filter(item => !idsSet.has(item.id));
      total.value -= itemIds.length;
      stats.value.total_size = Math.max(0, stats.value.total_size - removedSize);
      
      // Reload if current page is now empty and not the first page
      if (items.value.length === 0 && currentPage.value > 1) {
        await loadItems(currentPage.value - 1, pageSize.value);
      } else if (items.value.length === 0 && currentPage.value === 1) {
        // If first page is empty, reload to check if there are more items
        await loadItems(1, pageSize.value);
      }
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handlePermanentDeleteError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Empty recycle bin (permanently delete all items)
   */
  async function emptyRecycleBin(): Promise<void> {
    isLoading.value = true;
    try {
      await apiEmptyRecycleBin();
      
      // Clear local state
      items.value = [];
      total.value = 0;
      stats.value = {
        total_size: 0,
        expiring_soon: 0,
        expired: 0,
        count_by_type: {},
      };
      currentPage.value = 1;
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleEmptyRecycleBinError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Remove expired recycle-bin items from local state.
   * This is a defensive startup cleanup so unrelated pages do not fail
   * if stale recycle data is present in memory.
   */
  function cleanExpiredItems(): void {
    const now = Date.now();
      const activeItems = items.value.filter((item) => {
      const expiresAt = Date.parse(item.expiresAt);

      // Keep malformed records instead of deleting them silently.
      if (Number.isNaN(expiresAt)) {
        return true;
      }

      return expiresAt > now;
    });

    if (activeItems.length !== items.value.length) {
      items.value = activeItems;
      total.value = activeItems.length;
    }
  }

  /**
   * Get remaining days before an item expires
   * @param item - Recycle item
   * @returns Number of days remaining
   */
  function getRemainingDays(item: RecycleItem): number {
    const expiresAt = Date.parse(item.expiresAt);
    if (Number.isNaN(expiresAt)) return 0;
    return Math.max(0, Math.ceil((expiresAt - Date.now()) / (1000 * 60 * 60 * 24)));
  }

  /**
   * Check if an item is expiring soon (≤3 days)
   * @param item - Recycle item
   * @returns True if expiring within 3 days
   */
  function isExpiringSoon(item: RecycleItem): boolean {
    const remainingDays = getRemainingDays(item);
    return remainingDays > 0 && remainingDays <= 3;
  }

  /**
   * Get total storage used by recycle bin items
   * @returns Total size in bytes
   */
  function getTotalSize(): number {
    return stats.value.total_size || items.value.reduce((sum, item) => sum + item.fileSize, 0);
  }

  /**
   * Get items that are expiring soon (≤3 days)
   * @returns Array of items expiring soon
   */
  function getExpiringSoonItems(): RecycleItem[] {
    return items.value.filter(item => isExpiringSoon(item));
  }

  /**
   * Get items by file type
   * @param fileType - File type to filter by
   * @returns Array of items matching the file type
   */
  function getItemsByType(fileType: string): RecycleItem[] {
    return items.value.filter(item => item.fileType === fileType);
  }

  /**
   * Batch select all items on current page
   * @returns Array of all item IDs on current page
   */
  function selectAllOnPage(): string[] {
    return items.value.map(item => item.id);
  }

  return {
    // State
    items,
    total,
    stats,
    isLoading,
    currentPage,
    pageSize,

    // Actions
    loadItems,
    moveToRecycle,
    restoreFiles,
    permanentDelete,
    emptyRecycleBin,
    cleanExpiredItems,

    // Helpers
    getRemainingDays,
    isExpiringSoon,
    getTotalSize,
    getExpiringSoonItems,
    getItemsByType,
    selectAllOnPage
  };
});

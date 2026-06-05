// Share Store for managing file sharing functionality
import { defineStore } from 'pinia';
import { ref, type Ref } from 'vue';
import {
  createShare as apiCreateShare,
  getShareInfo as apiGetShareInfo,
  verifySharePassword as apiVerifySharePassword,
  getMyShares as apiGetMyShares,
  deleteShare as apiDeleteShare,
  incrementDownloadCount as apiIncrementDownloadCount,
  type CreateShareRequest,
  type CreateShareResponse,
  type ShareInfo,
  type MyShare,
  type VerifyPasswordRequest
} from '@/api/share';
import {
  validateShareCreation,
  handleShareCreationError,
  handleShareAccessError,
  isShareExpired as checkShareExpired
} from '@/utils/share-validation';
import { resolveShareURL } from '@/utils/public-url';

/**
 * Extended share link interface with share URL
 */
export interface ShareLink extends MyShare {
  share_url: string;
  share_token: string;
  has_password: boolean;
}

/**
 * Options for creating a share
 */
export interface CreateShareOptions {
  expires_in?: number | null;
  access_code?: string | null;
  max_downloads?: number | null;
}

export const useShareStore = defineStore('share', () => {
  // State
  const shares: Ref<ShareLink[]> = ref([]);
  const currentShareInfo: Ref<ShareInfo | null> = ref(null);
  const isLoading: Ref<boolean> = ref(false);

  /**
   * Create a share link for files
   * @param fileIds - Array of file IDs to share
   * @param options - Share options (expiration, password)
   * @returns Created share link
   */
  async function createShare(
    fileIds: string[],
    options: CreateShareOptions = {}
  ): Promise<ShareLink> {
    isLoading.value = true;
    try {
      // Validate share creation parameters
      validateShareCreation(
        fileIds,
        options.expires_in ?? null,
        options.access_code ?? null
      );

      const request: CreateShareRequest = {
        file_ids: fileIds,
        expires_in: options.expires_in,
        access_code: options.access_code,
        max_downloads: options.max_downloads
      };

      const response: CreateShareResponse = await apiCreateShare(request);

      // Create share link object
      const shareLink: ShareLink = {
        share_id: response.share_id,
        share_token: response.share_token,
        file_ids: fileIds,
        file_names: [], // Will be populated when loading shares
        created_at: new Date().toISOString(),
        expires_at: response.expires_at,
        max_downloads: response.max_downloads,
        download_count: 0,
        access_count: 0,
        share_url: resolveShareURL(response.share_url, response.share_token),
        has_password: !!options.access_code
      };

      // Add to shares list
      shares.value.unshift(shareLink);

      return shareLink;
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleShareCreationError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Load current user's share list
   */
  async function loadMyShares(): Promise<void> {
    isLoading.value = true;
    try {
      const myShares = await apiGetMyShares();
      
      // Convert to ShareLink format
      shares.value = myShares.map(share => ({
        ...share,
        share_token: share.share_token || String(share.share_id),
        share_url: resolveShareURL(share.share_url || '', share.share_token || String(share.share_id)),
        has_password: !!share.has_password
      }));
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Get share information by share ID
   * @param shareId - Share ID
   * @returns Share information
   */
  async function getShareInfo(shareId: string): Promise<ShareInfo> {
    isLoading.value = true;
    try {
      const info = await apiGetShareInfo(shareId);
      currentShareInfo.value = info;
      return info;
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleShareAccessError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Verify share password
   * @param shareId - Share ID
   * @param password - Access password
   * @returns True if password is valid
   */
  async function verifyPassword(shareId: string, password: string): Promise<boolean> {
    isLoading.value = true;
    try {
      const request: VerifyPasswordRequest = {
        access_code: password
      };
      const response = await apiVerifySharePassword(shareId, request);
      return response.valid;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Delete a share link
   * @param shareId - Share ID to delete
   */
  async function deleteShare(shareId: string): Promise<void> {
    isLoading.value = true;
    try {
      await apiDeleteShare(shareId);
      
      // Remove from local state
      shares.value = shares.value.filter(share => share.share_id !== shareId);
      
      // Clear current share info if it matches
      if (currentShareInfo.value?.share_id === shareId) {
        currentShareInfo.value = null;
      }
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Delete multiple share links
   * @param shareIds - Array of share IDs to delete
   */
  async function deleteShares(shareIds: string[]): Promise<void> {
    isLoading.value = true;
    try {
      // Delete all shares in parallel
      await Promise.all(shareIds.map(id => apiDeleteShare(id)));
      
      // Remove from local state
      shares.value = shares.value.filter(
        share => !shareIds.includes(share.share_id)
      );
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Increment download count for a share
   * @param shareId - Share ID
   */
  async function incrementDownloadCount(shareId: string): Promise<void> {
    try {
      await apiIncrementDownloadCount(shareId);
      
      // Update local state
      const share = shares.value.find(s => s.share_id === shareId);
      if (share) {
        share.download_count++;
      }
      
      // Update current share info
      if (currentShareInfo.value?.share_id === shareId) {
        currentShareInfo.value.download_count++;
      }
    } catch (error) {
      // Silently fail - download count is not critical
      console.error('Failed to increment download count:', error);
    }
  }

  /**
   * Check if a share has expired
   * @param share - Share link to check
   * @returns True if share has expired
   */
  function isExpired(share: ShareLink | ShareInfo): boolean {
    return checkShareExpired(share.expires_at);
  }

  /**
   * Clear current share info
   */
  function clearCurrentShareInfo(): void {
    currentShareInfo.value = null;
  }

  return {
    // State
    shares,
    currentShareInfo,
    isLoading,
    
    // Actions
    createShare,
    loadMyShares,
    getShareInfo,
    verifyPassword,
    deleteShare,
    deleteShares,
    incrementDownloadCount,
    
    // Helpers
    isExpired,
    clearCurrentShareInfo
  };
});

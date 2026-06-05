// Version Store for managing file version history
import { defineStore } from 'pinia';
import { ref, type Ref } from 'vue';
import {
  getVersionHistory as apiGetVersionHistory,
  downloadVersion as apiDownloadVersion,
  restoreVersion as apiRestoreVersion,
  deleteVersion as apiDeleteVersion,
  type FileVersion
} from '@/api/version';
import {
  validateFileId,
  validateVersionId,
  validateVersionRestore,
  validateVersionDownload,
  validateVersionDeletion,
  handleVersionError,
  handleVersionRestoreError,
  handleVersionDownloadError,
  calculateVersionStorage
} from '@/utils/version-validation';

export const useVersionStore = defineStore('version', () => {
  // State
  const versions: Ref<Map<string, FileVersion[]>> = ref(new Map());
  const isLoading: Ref<boolean> = ref(false);
  const loadingStates: Ref<Map<string, boolean>> = ref(new Map()); // Track loading per file
  const hasLoadedOnce: Ref<Map<string, boolean>> = ref(new Map()); // Track if loaded at least once

  /**
   * Check if version history has been loaded for a file
   * @param fileId - File ID
   * @returns True if loaded
   */
  function hasLoaded(fileId: string): boolean {
    return hasLoadedOnce.value.get(fileId) || false;
  }

  /**
   * Check if version history is currently loading for a file
   * @param fileId - File ID
   * @returns True if loading
   */
  function isLoadingFile(fileId: string): boolean {
    return loadingStates.value.get(fileId) || false;
  }

  /**
   * Load version history for a file (with lazy loading support)
   * Requirements: 6.4 - Lazy loading for version history
   * @param fileId - File ID
   * @param force - Force reload even if already loaded
   */
  async function loadVersionHistory(fileId: string, force: boolean = false): Promise<void> {
    // Skip if already loaded and not forcing reload (lazy loading)
    if (!force && hasLoaded(fileId)) {
      return;
    }

    loadingStates.value.set(fileId, true);
    isLoading.value = true;
    
    try {
      const history = await apiGetVersionHistory(fileId);
      versions.value.set(fileId, history);
      hasLoadedOnce.value.set(fileId, true);
    } finally {
      loadingStates.value.set(fileId, false);
      isLoading.value = false;
    }
  }

  /**
   * Download a specific version of a file
   * @param fileId - File ID
   * @param versionId - Version ID
   */
  async function downloadVersion(fileId: string, versionId: string): Promise<void> {
    isLoading.value = true;
    try {
      // Validate download parameters
      validateVersionDownload(fileId, versionId);

      const blob = await apiDownloadVersion(fileId, versionId);
      
      // Get version info to determine filename
      const fileVersions = versions.value.get(fileId);
      const version = fileVersions?.find(v => v.version_id === versionId);
      
      // Create download link
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = version 
        ? `file_v${version.version_number}.dat` 
        : 'file_version.dat';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleVersionDownloadError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Restore a previous version (creates a new version)
   * @param fileId - File ID
   * @param versionId - Version ID to restore
   */
  async function restoreVersion(fileId: string, versionId: string): Promise<void> {
    isLoading.value = true;
    try {
      // Get version info to check if it's current
      const fileVersions = versions.value.get(fileId);
      const version = fileVersions?.find(v => v.version_id === versionId);
      const isCurrentVersion = version?.is_current || false;

      // Validate restore parameters
      validateVersionRestore(fileId, versionId, isCurrentVersion);

      await apiRestoreVersion(fileId, versionId);
      
      // Reload version history to reflect the new version
      await loadVersionHistory(fileId, true);
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleVersionRestoreError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Delete a specific version
   * @param fileId - File ID
   * @param versionId - Version ID to delete
   */
  async function deleteVersion(fileId: string, versionId: string): Promise<void> {
    isLoading.value = true;
    try {
      // Get version info to check if it's current
      const fileVersions = versions.value.get(fileId);
      const version = fileVersions?.find(v => v.version_id === versionId);
      const isCurrentVersion = version?.is_current || false;
      const totalVersions = fileVersions?.length || 0;

      // Validate deletion parameters
      validateVersionDeletion(fileId, versionId, isCurrentVersion, totalVersions);

      await apiDeleteVersion(fileId, versionId);
      
      // Update local state
      if (fileVersions) {
        const updatedVersions = fileVersions.filter(v => v.version_id !== versionId);
        versions.value.set(fileId, updatedVersions);
      }
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleVersionError(error);
      throw new Error(message);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Get current version of a file
   * @param fileId - File ID
   * @returns Current version or null if not found
   */
  function getCurrentVersion(fileId: string): FileVersion | null {
    const fileVersions = versions.value.get(fileId);
    if (!fileVersions) {
      return null;
    }
    
    return fileVersions.find(v => v.is_current) || null;
  }

  /**
   * Get version count for a file
   * @param fileId - File ID
   * @returns Number of versions
   */
  function getVersionCount(fileId: string): number {
    const fileVersions = versions.value.get(fileId);
    return fileVersions ? fileVersions.length : 0;
  }

  /**
   * Get versions for a file (with pagination support)
   * Requirements: 6.5 - Pagination for large version lists
   * @param fileId - File ID
   * @param page - Page number (1-indexed)
   * @param pageSize - Number of versions per page
   * @returns Paginated array of versions
   */
  function getVersionsPaginated(
    fileId: string,
    page: number = 1,
    pageSize: number = 10
  ): { versions: FileVersion[]; total: number; hasMore: boolean } {
    const allVersions = versions.value.get(fileId) || [];
    const total = allVersions.length;
    const startIndex = (page - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    const paginatedVersions = allVersions.slice(startIndex, endIndex);
    const hasMore = endIndex < total;

    return {
      versions: paginatedVersions,
      total,
      hasMore
    };
  }

  /**
   * Get versions for a file
   * @param fileId - File ID
   * @returns Array of versions or empty array
   */
  function getVersions(fileId: string): FileVersion[] {
    return versions.value.get(fileId) || [];
  }

  /**
   * Get total storage used by versions of a file
   * @param fileId - File ID
   * @returns Total size in bytes
   */
  function getVersionsSize(fileId: string): number {
    const fileVersions = versions.value.get(fileId);
    if (!fileVersions) {
      return 0;
    }
    
    return calculateVersionStorage(fileVersions);
  }

  /**
   * Clear version history for a file from local state
   * @param fileId - File ID
   */
  function clearVersionHistory(fileId: string): void {
    versions.value.delete(fileId);
  }

  /**
   * Clear all version history from local state
   */
  function clearAllVersions(): void {
    versions.value.clear();
  }

  return {
    // State
    versions,
    isLoading,

    // Actions
    loadVersionHistory,
    downloadVersion,
    restoreVersion,
    deleteVersion,

    // Helpers
    getCurrentVersion,
    getVersionCount,
    getVersions,
    getVersionsPaginated,
    getVersionsSize,
    clearVersionHistory,
    clearAllVersions,
    hasLoaded,
    isLoadingFile
  };
});

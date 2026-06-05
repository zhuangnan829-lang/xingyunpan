/**
 * File Store - Pinia state management for files and folders
 * 
 * Requirements: 4.1-4.6, 5.1-5.5, 11.1-11.4, 12.1-12.5, 13.1-13.5, 
 *               14.1-14.5, 15.1-15.5, 16.1-16.5, 17.1-17.5, 18.1-18.5, 
 *               19.1-19.5, 20.1-20.5, 21.1-21.5
 */

import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { FileItem, FolderItem, FileFilters } from '@/types/file';
import { FolderPathItem } from '@/types/folder';
import * as fileAPI from '@/api/file';
import * as folderAPI from '@/api/folder';
import { applyFilters } from '@/utils/filter-utils';

export type ViewMode = 'table' | 'grid' | 'gallery';

export const useFileStore = defineStore('file', () => {
  // State
  const currentFolderId = ref<number | null>(null);
  const files = ref<FileItem[]>([]);
  const folders = ref<FolderItem[]>([]);
  const breadcrumb = ref<FolderPathItem[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const viewMode = ref<ViewMode>('grid');
  const isSearchResult = ref(false);
  const searchTotal = ref(0);
  const currentPageSize = ref(2000);
  
  // Search and filter state
  const searchKeyword = ref<string>('');
  const filters = ref<FileFilters>({
    type: 'all',
    sizeRange: 'all',
    timeRange: 'all',
  });

  // Getters
  const currentFolder = computed(() => {
    if (currentFolderId.value === null) {
      return null;
    }
    return folders.value.find((f: FolderItem) => f.id === currentFolderId.value) || null;
  });

  const isEmpty = computed(() => {
    return files.value.length === 0 && folders.value.length === 0;
  });
  
  /**
   * Check if there are active search or filter conditions
   * Requirements: 8.4
   */
  const hasActiveFilters = computed(() => {
    return (
      searchKeyword.value.trim() !== '' ||
      filters.value.type !== 'all' ||
      filters.value.sizeRange !== 'all' ||
      filters.value.timeRange !== 'all'
    );
  });
  
  /**
   * Get filtered file items based on search keyword and filters
   * Requirements: 4.3, 5.4, 5.7, 8.1
   */
  const filteredItems = computed(() => {
    return applyFilters(files.value, searchKeyword.value, filters.value);
  });

  // Actions

  /**
   * Fetch files and folders for the current folder
   * Requirements: 4.1-4.6
   */
  async function fetchFiles(folderId?: number | null, pageSize: number = currentPageSize.value): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      currentPageSize.value = pageSize;
      const response = await fileAPI.getFiles(folderId, 1, pageSize);
      files.value = response.files;
      folders.value = response.folders;
      currentFolderId.value = folderId === undefined ? null : folderId;
      
      // Update breadcrumb
      if (folderId === null || folderId === undefined) {
        breadcrumb.value = [];
      } else {
        try {
          breadcrumb.value = await folderAPI.getFolderPath(folderId);
        } catch (error) {
          console.error('Failed to fetch folder path:', error);
          breadcrumb.value = [];
        }
      }
    } catch (err: any) {
      error.value = err.message || '加载文件列表失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Navigate to a specific folder
   * Requirements: 17.1-17.5
   */
  async function navigateToFolder(folderId: number | null): Promise<void> {
    await fetchFiles(folderId);
  }

  /**
   * Navigate to parent folder
   * Requirements: 17.2, 17.5
   */
  async function navigateToParent(): Promise<void> {
    if (breadcrumb.value.length === 0) {
      // Already at root
      return;
    }
    
    if (breadcrumb.value.length === 1) {
      // Go to root
      await navigateToFolder(null);
    } else {
      // Go to parent folder
      const parentFolder = breadcrumb.value[breadcrumb.value.length - 2];
      await navigateToFolder(parentFolder.id);
    }
  }

  /**
   * Refresh current folder
   */
  async function refresh(): Promise<void> {
    await fetchFiles(currentFolderId.value, currentPageSize.value);
  }

  /**
   * Set view mode
   */
  function setViewMode(mode: ViewMode): void {
    viewMode.value = mode;
  }

  // File Operations

  /**
   * Download file
   * Requirements: 11.1-11.4
   */
  async function downloadFile(fileId: number): Promise<void> {
    const file = files.value.find((f: FileItem) => f.id === fileId);
    if (!file) {
      throw new Error('File not found');
    }

    const blob = await fileAPI.downloadFile(fileId);
    
    // Create download link
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = file.name;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  }

  /**
   * Rename file
   * Requirements: 12.1-12.5
   */
  async function renameFile(fileId: number, newName: string): Promise<void> {
    await fileAPI.renameFile(fileId, newName);
    await refresh();
  }

  /**
   * Delete file
   * Requirements: 13.1-13.5
   */
  async function deleteFile(fileId: number): Promise<void> {
    await fileAPI.deleteFile(fileId);
    await refresh();
  }

  /**
   * Move file to another folder
   * Requirements: 14.1-14.5
   */
  async function moveFile(fileId: number, targetFolderId: number | null): Promise<void> {
    await fileAPI.moveFile(fileId, targetFolderId);
    await refresh();
  }

  /**
   * Copy file to another folder
   * Requirements: 15.1-15.5
   */
  async function copyFile(fileId: number, targetFolderId: number | null): Promise<void> {
    await fileAPI.copyFile(fileId, targetFolderId);
    // Only refresh if copying to current folder
    if (targetFolderId === currentFolderId.value) {
      await refresh();
    }
  }

  // Folder Operations

  /**
   * Create folder
   * Requirements: 16.1-16.5
   */
  async function createFolder(name: string): Promise<void> {
    await folderAPI.createFolder(name, currentFolderId.value);
    await refresh();
  }

  /**
   * Rename folder
   * Requirements: 18.1-18.5
   */
  async function renameFolder(folderId: number, newName: string): Promise<void> {
    await folderAPI.renameFolder(folderId, newName);
    await refresh();
  }

  /**
   * Delete folder
   * Requirements: 19.1-19.5
   */
  async function deleteFolder(folderId: number): Promise<void> {
    await folderAPI.deleteFolder(folderId);
    await refresh();
  }

  /**
   * Move folder to another parent folder
   * Requirements: 20.1-20.5
   */
  async function moveFolder(folderId: number, targetFolderId: number | null): Promise<void> {
    await folderAPI.moveFolder(folderId, targetFolderId);
    await refresh();
  }

  /**
   * Copy folder to another parent folder
   * Requirements: 21.1-21.5
   */
  async function copyFolder(folderId: number, targetFolderId: number | null): Promise<void> {
    await folderAPI.copyFolder(folderId, targetFolderId);
    // Only refresh if copying to current folder
    if (targetFolderId === currentFolderId.value) {
      await refresh();
    }
  }
  
  // Search and Filter Operations
  
  /**
   * Set search keyword
   * Requirements: 4.6
   */
  function setSearchKeyword(keyword: string): void {
    searchKeyword.value = keyword;
  }

  function applySearchResults(resultFiles: FileItem[], resultFolders: FolderItem[], total: number, keyword: string): void {
    files.value = resultFiles;
    folders.value = resultFolders;
    searchKeyword.value = keyword;
    searchTotal.value = total;
    isSearchResult.value = keyword.trim() !== '';
    error.value = null;
  }

  function clearSearchResults(): void {
    isSearchResult.value = false;
    searchTotal.value = 0;
    searchKeyword.value = '';
  }
  
  /**
   * Set filters (partial update)
   * Requirements: 5.6
   */
  function setFilters(newFilters: Partial<FileFilters>): void {
    filters.value = {
      ...filters.value,
      ...newFilters,
    };
  }
  
  /**
   * Clear all filters (reset to default)
   * Requirements: 5.6
   */
  function clearFilters(): void {
    filters.value = {
      type: 'all',
      sizeRange: 'all',
      timeRange: 'all',
    };
  }
  
  /**
   * Clear all search and filter conditions
   * Requirements: 8.2
   */
  function clearAll(): void {
    searchKeyword.value = '';
    clearFilters();
  }

  return {
    // State
    currentFolderId,
    files,
    folders,
    breadcrumb,
    loading,
    error,
    viewMode,
    isSearchResult,
    searchTotal,
    searchKeyword,
    filters,
    
    // Getters
    currentFolder,
    isEmpty,
    hasActiveFilters,
    filteredItems,
    
    // Actions
    fetchFiles,
    navigateToFolder,
    navigateToParent,
    refresh,
    setViewMode,
    
    // File operations
    downloadFile,
    renameFile,
    deleteFile,
    moveFile,
    copyFile,
    
    // Folder operations
    createFolder,
    renameFolder,
    deleteFolder,
    moveFolder,
    copyFolder,
    
    // Search and filter operations
    setSearchKeyword,
    applySearchResults,
    clearSearchResults,
    setFilters,
    clearFilters,
    clearAll,
  };
});

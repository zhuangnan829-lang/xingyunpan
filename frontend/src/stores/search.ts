/**
 * Search store - Pinia state management for search functionality
 * Manages search history, filters, results, and suggestions
 */

import { defineStore } from 'pinia';
import { ref, type Ref } from 'vue';
import {
  searchFiles as apiSearchFiles,
  getSearchSuggestions as apiGetSearchSuggestions,
  type SearchRequest,
  type SearchResult,
  type SearchSuggestion
} from '@/api/search';
import type { FileItem } from '@/types/file';
import type { FolderItem } from '@/types/folder';
import {
  validateSearchRequest,
  handleSearchError,
  sanitizeSearchQuery
} from '@/utils/search-validation';

const STORAGE_KEY = 'xingyunpan_search_history';
const MAX_HISTORY_SIZE = 20;
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes in milliseconds

/**
 * Search history entry
 */
export interface SearchHistory {
  keyword: string;
  timestamp: string;
}

/**
 * Search filters
 */
export interface SearchFilters {
  fileType: string | null;
  sizeMin: number | null;
  sizeMax: number | null;
  dateFrom: string | null;
  dateTo: string | null;
}

/**
 * Cached search result entry
 */
interface CachedSearchResult {
  result: SearchResult;
  timestamp: number;
  filters: SearchFilters;
}

export const useSearchStore = defineStore('search', () => {
  // State
  const history: Ref<SearchHistory[]> = ref([]);
  const currentKeyword: Ref<string> = ref('');
  const filters: Ref<SearchFilters> = ref({
    fileType: null,
    sizeMin: null,
    sizeMax: null,
    dateFrom: null,
    dateTo: null
  });
  const searchResults: Ref<FileItem[]> = ref([]);
  const folderResults: Ref<FolderItem[]> = ref([]);
  const totalResults: Ref<number> = ref(0);
  const currentPage: Ref<number> = ref(1);
  const pageSize: Ref<number> = ref(20);
  const isSearching: Ref<boolean> = ref(false);
  
  // Cache for search results (Requirements: 6.6)
  const searchCache: Ref<Map<string, CachedSearchResult>> = ref(new Map());

  // Actions

  /**
   * Generate cache key for search request
   * @param keyword - Search keyword
   * @param filters - Search filters
   * @param page - Page number
   * @returns Cache key string
   */
  function generateCacheKey(keyword: string, filters: SearchFilters, page: number): string {
    return JSON.stringify({ keyword, filters, page });
  }

  /**
   * Check if cached result is still valid
   * @param cached - Cached search result
   * @returns True if cache is valid
   */
  function isCacheValid(cached: CachedSearchResult): boolean {
    return Date.now() - cached.timestamp < CACHE_DURATION;
  }

  /**
   * Get cached search result if available and valid
   * @param keyword - Search keyword
   * @param filters - Search filters
   * @param page - Page number
   * @returns Cached result or null
   */
  function getCachedResult(
    keyword: string,
    filters: SearchFilters,
    page: number
  ): SearchResult | null {
    const cacheKey = generateCacheKey(keyword, filters, page);
    const cached = searchCache.value.get(cacheKey);
    
    if (cached && isCacheValid(cached)) {
      return cached.result;
    }
    
    // Remove expired cache entry
    if (cached) {
      searchCache.value.delete(cacheKey);
    }
    
    return null;
  }

  /**
   * Cache search result
   * @param keyword - Search keyword
   * @param filters - Search filters
   * @param page - Page number
   * @param result - Search result to cache
   */
  function cacheResult(
    keyword: string,
    filters: SearchFilters,
    page: number,
    result: SearchResult
  ): void {
    const cacheKey = generateCacheKey(keyword, filters, page);
    searchCache.value.set(cacheKey, {
      result,
      timestamp: Date.now(),
      filters: { ...filters }
    });
  }

  /**
   * Clear expired cache entries
   */
  function clearExpiredCache(): void {
    const now = Date.now();
    const keysToDelete: string[] = [];
    
    searchCache.value.forEach((cached, key) => {
      if (now - cached.timestamp >= CACHE_DURATION) {
        keysToDelete.push(key);
      }
    });
    
    keysToDelete.forEach(key => searchCache.value.delete(key));
  }

  /**
   * Clear all cached search results
   */
  function clearCache(): void {
    searchCache.value.clear();
  }

  /**
   * Search files with filters (with caching support)
   * @param keyword - Search keyword
   * @param customFilters - Optional custom filters
   * @param page - Page number (default: 1)
   */
  async function search(
    keyword: string,
    customFilters?: Partial<SearchFilters>,
    page: number = 1
  ): Promise<void> {
    isSearching.value = true;
    
    try {
      // Sanitize and validate search query
      const sanitizedKeyword = sanitizeSearchQuery(keyword);
      currentKeyword.value = sanitizedKeyword;
      currentPage.value = page;

      // Merge custom filters with current filters
      const activeFilters = customFilters ? { ...filters.value, ...customFilters } : filters.value;

      // Validate search request
      validateSearchRequest(
        sanitizedKeyword,
        activeFilters.fileType,
        activeFilters.sizeMin,
        activeFilters.sizeMax,
        activeFilters.dateFrom,
        activeFilters.dateTo,
        page,
        pageSize.value
      );

      // Check cache first (Requirements: 6.6)
      const cachedResult = getCachedResult(sanitizedKeyword, activeFilters, page);
      if (cachedResult) {
        searchResults.value = cachedResult.files;
        folderResults.value = cachedResult.folders || [];
        totalResults.value = cachedResult.total;
        
        // Add to history if keyword is not empty
        if (sanitizedKeyword.trim()) {
          addHistory(sanitizedKeyword);
        }
        
        return;
      }

      // Build search request
      const request: SearchRequest = {
        keyword: sanitizedKeyword,
        file_type: activeFilters.fileType,
        size_min: activeFilters.sizeMin,
        size_max: activeFilters.sizeMax,
        date_from: activeFilters.dateFrom,
        date_to: activeFilters.dateTo,
        page,
        page_size: pageSize.value
      };

      const result: SearchResult = await apiSearchFiles(request);
      
      searchResults.value = result.files;
      folderResults.value = result.folders || [];
      totalResults.value = result.total;
      
      // Cache the result (Requirements: 6.6)
      cacheResult(sanitizedKeyword, activeFilters, page, result);
      
      // Clear expired cache entries periodically
      clearExpiredCache();
      
      // Add to history if keyword is not empty
      if (sanitizedKeyword.trim()) {
        addHistory(sanitizedKeyword);
      }
    } catch (error) {
      // Handle and re-throw with user-friendly message
      const message = handleSearchError(error);
      throw new Error(message);
    } finally {
      isSearching.value = false;
    }
  }

  /**
   * Add a search keyword to history
   * @param keyword - Search keyword to add
   */
  function addHistory(keyword: string): void {
    const trimmedKeyword = keyword.trim();
    if (!trimmedKeyword) {
      return;
    }

    // Remove existing entry if present
    const filtered = history.value.filter(item => item.keyword !== trimmedKeyword);
    
    // Add new entry at the beginning
    const newEntry: SearchHistory = {
      keyword: trimmedKeyword,
      timestamp: new Date().toISOString()
    };
    
    history.value = [newEntry, ...filtered].slice(0, MAX_HISTORY_SIZE);
    
    // Save to localStorage
    saveToStorage();
  }

  /**
   * Get search suggestions based on prefix
   * @param prefix - Search keyword prefix
   * @returns Array of search suggestions
   */
  async function getSuggestions(prefix: string): Promise<SearchSuggestion[]> {
    if (!prefix.trim()) {
      return [];
    }

    try {
      return await apiGetSearchSuggestions(prefix);
    } catch (error) {
      console.error('Failed to get search suggestions:', error);
      return [];
    }
  }

  /**
   * Get recent search history
   * @param limit - Maximum number of history items to return
   * @returns Array of recent search history
   */
  function getRecentHistory(limit: number = 5): SearchHistory[] {
    return history.value.slice(0, limit);
  }

  /**
   * Clear search results and keyword
   */
  function clearSearch(): void {
    currentKeyword.value = '';
    searchResults.value = [];
    folderResults.value = [];
    totalResults.value = 0;
    currentPage.value = 1;
    // Clear cache when clearing search
    clearCache();
  }

  /**
   * Clear all search history
   */
  function clearHistory(): void {
    history.value = [];
    saveToStorage();
  }

  /**
   * Set search filters
   * @param newFilters - Partial filters to update
   */
  function setFilters(newFilters: Partial<SearchFilters>): void {
    filters.value = { ...filters.value, ...newFilters };
  }

  /**
   * Reset all filters to default values
   */
  function resetFilters(): void {
    filters.value = {
      fileType: null,
      sizeMin: null,
      sizeMax: null,
      dateFrom: null,
      dateTo: null
    };
  }

  /**
   * Load search history from localStorage
   */
  function loadFromStorage(): void {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        history.value = Array.isArray(parsed) ? parsed : [];
      }
    } catch (error) {
      console.error('Failed to load search history from storage:', error);
      history.value = [];
    }
  }

  /**
   * Save search history to localStorage
   */
  function saveToStorage(): void {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(history.value));
    } catch (error) {
      console.error('Failed to save search history to storage:', error);
    }
  }

  // Initialize: load from localStorage
  loadFromStorage();

  return {
    // State
    history,
    currentKeyword,
    filters,
    searchResults,
    folderResults,
    totalResults,
    currentPage,
    pageSize,
    isSearching,

    // Actions
    search,
    addHistory,
    getSuggestions,
    getRecentHistory,
    clearSearch,
    clearHistory,
    setFilters,
    resetFilters,
    loadFromStorage,
    saveToStorage,
    clearCache,
    clearExpiredCache
  };
});

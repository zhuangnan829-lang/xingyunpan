// Search API module for file search functionality

import request from './request';
import { FileItem, FolderItem } from '@/types/file';

/**
 * Request interface for searching files
 */
export interface SearchRequest {
  keyword: string;
  file_type?: string | null;     // File type filter
  size_min?: number | null;      // Minimum size in bytes
  size_max?: number | null;      // Maximum size in bytes
  date_from?: string | null;     // Start date (ISO 8601)
  date_to?: string | null;       // End date (ISO 8601)
  folder_id?: string | null;     // Limit search to specific folder
  page?: number;
  page_size?: number;
}

/**
 * Response interface for search results
 */
export interface SearchResult {
  files: FileItem[];
  folders: FolderItem[];
  total: number;
  page: number;
  page_size: number;
}

/**
 * Interface for search suggestions
 */
export interface SearchSuggestion {
  keyword: string;
  count: number;
}

/**
 * Search files with filters and pagination
 * @param req - Search request with filters
 * @returns Search results with pagination
 */
export async function searchFiles(req: SearchRequest): Promise<SearchResult> {
  const response = await request.post<any>('/api/v1/search', req);
  const items = response.files || [];
  const files: FileItem[] = [];
  const folders: FolderItem[] = [];

  items.forEach((item: any) => {
    const id = Number(item.file_id || item.id);
    const name = item.file_name || item.name || '';
    const updatedAt = item.modified_at || item.updated_at || new Date().toISOString();

    if (item.is_folder) {
      folders.push({
        id,
        name,
        parent_id: null,
        created_at: updatedAt,
        updated_at: updatedAt,
      });
      return;
    }

    files.push({
      id,
      name,
      size: item.file_size || item.size || 0,
      hash: '',
      mime_type: item.file_type || item.mime_type || 'application/octet-stream',
      content_type: item.file_type || item.content_type || item.mime_type || 'application/octet-stream',
      folder_id: null,
      created_at: updatedAt,
      updated_at: updatedAt,
      browser_app: null,
    });
  });

  return {
    files,
    folders,
    total: response.total || items.length,
    page: response.page || req.page || 1,
    page_size: response.page_size || req.page_size || 20,
  };
}

/**
 * Get search suggestions based on prefix
 * @param prefix - Search keyword prefix
 * @returns List of search suggestions
 */
export async function getSearchSuggestions(prefix: string): Promise<SearchSuggestion[]> {
  return request.get('/api/v1/search/suggestions', {
    params: { prefix }
  });
}

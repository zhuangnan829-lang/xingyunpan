/**
 * Search utility functions
 * Provides helper functions for search term highlighting, query building, and result filtering
 */

import type { FileItem } from '@/types/file'

/**
 * Highlight search term in text with HTML markup
 * @param text - The text to highlight in
 * @param searchTerm - The term to highlight
 * @returns HTML string with highlighted terms
 */
export function highlightSearchTerm(text: string, searchTerm: string): string {
  if (!text || !searchTerm) {
    return text
  }
  
  // Escape special regex characters in search term
  const escapedTerm = searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  
  // Create case-insensitive regex with global flag
  const regex = new RegExp(`(${escapedTerm})`, 'gi')
  
  // Replace matches with highlighted span
  return text.replace(regex, '<mark class="search-highlight">$1</mark>')
}

/**
 * Build search query parameters from filters
 * @param keyword - Search keyword
 * @param filters - Optional filter parameters
 * @returns Query parameters object
 */
export function buildSearchQuery(
  keyword: string,
  filters?: {
    fileType?: string | null
    sizeMin?: number | null
    sizeMax?: number | null
    dateFrom?: string | null
    dateTo?: string | null
    folderId?: string | null
  }
): Record<string, any> {
  const query: Record<string, any> = {
    keyword: keyword.trim()
  }
  
  if (!filters) {
    return query
  }
  
  // Add file type filter
  if (filters.fileType) {
    query.file_type = filters.fileType
  }
  
  // Add size range filters
  if (filters.sizeMin !== null && filters.sizeMin !== undefined) {
    query.size_min = filters.sizeMin
  }
  if (filters.sizeMax !== null && filters.sizeMax !== undefined) {
    query.size_max = filters.sizeMax
  }
  
  // Add date range filters
  if (filters.dateFrom) {
    query.date_from = filters.dateFrom
  }
  if (filters.dateTo) {
    query.date_to = filters.dateTo
  }
  
  // Add folder scope filter
  if (filters.folderId) {
    query.folder_id = filters.folderId
  }
  
  return query
}

/**
 * Filter search results locally based on criteria
 * Useful for client-side filtering of cached results
 * @param results - Array of file items
 * @param filters - Filter criteria
 * @returns Filtered array of file items
 */
export function filterSearchResults(
  results: FileItem[],
  filters: {
    fileType?: string | null
    sizeMin?: number | null
    sizeMax?: number | null
    dateFrom?: string | null
    dateTo?: string | null
  }
): FileItem[] {
  let filtered = [...results]
  
  // Filter by file type
  if (filters.fileType) {
    filtered = filtered.filter(item => item.mime_type === filters.fileType)
  }
  
  // Filter by size range
  if (filters.sizeMin !== null && filters.sizeMin !== undefined) {
    filtered = filtered.filter(item => item.size >= filters.sizeMin!)
  }
  if (filters.sizeMax !== null && filters.sizeMax !== undefined) {
    filtered = filtered.filter(item => item.size <= filters.sizeMax!)
  }
  
  // Filter by date range
  if (filters.dateFrom) {
    const fromDate = new Date(filters.dateFrom)
    filtered = filtered.filter(item => {
      const itemDate = new Date(item.updated_at)
      return itemDate >= fromDate
    })
  }
  if (filters.dateTo) {
    const toDate = new Date(filters.dateTo)
    filtered = filtered.filter(item => {
      const itemDate = new Date(item.updated_at)
      return itemDate <= toDate
    })
  }
  
  return filtered
}

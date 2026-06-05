// Search-related TypeScript type definitions

/**
 * Search history record data model
 */
export interface SearchHistory {
  keyword: string;               // Search keyword
  timestamp: string;             // Search time (ISO 8601)
}

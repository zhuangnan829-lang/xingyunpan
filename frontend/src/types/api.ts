// API-related TypeScript type definitions

/**
 * API error response
 */
export interface APIError {
  code: string;
  message: string;
  details?: any;
}

/**
 * Generic API response wrapper
 */
export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: APIError;
  message?: string;
}

/**
 * Network error class
 */
export class NetworkError extends Error {
  constructor(message: string = '网络连接失败，请检查网络设置') {
    super(message);
    this.name = 'NetworkError';
  }
}

/**
 * Server error class
 */
export class ServerError extends Error {
  constructor(message: string = '服务器错误，请稍后重试') {
    super(message);
    this.name = 'ServerError';
  }
}

/**
 * Validation error class
 */
export class ValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ValidationError';
  }
}

/**
 * Pagination parameters
 */
export interface PaginationParams {
  page: number;
  page_size: number;
}

/**
 * Paginated response
 */
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

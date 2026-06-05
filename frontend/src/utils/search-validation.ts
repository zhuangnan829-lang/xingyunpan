// Search validation utilities

/**
 * Validation error class for search operations
 */
export class SearchValidationError extends Error {
  constructor(message: string, public code: string) {
    super(message);
    this.name = 'SearchValidationError';
  }
}

/**
 * Validate search query length
 * @param query - Search query string
 * @returns True if valid
 * @throws SearchValidationError if invalid
 */
export function validateSearchQuery(query: string): boolean {
  // Must be a string
  if (typeof query !== 'string') {
    throw new SearchValidationError(
      '搜索关键词格式无效',
      'INVALID_QUERY_FORMAT'
    );
  }

  // Trim whitespace
  const trimmedQuery = query.trim();

  // Minimum length: 1 character
  if (trimmedQuery.length === 0) {
    throw new SearchValidationError(
      '请输入搜索关键词',
      'QUERY_TOO_SHORT'
    );
  }

  // Maximum length: 100 characters
  const MAX_QUERY_LENGTH = 100;
  if (trimmedQuery.length > MAX_QUERY_LENGTH) {
    throw new SearchValidationError(
      `搜索关键词不能超过${MAX_QUERY_LENGTH}个字符`,
      'QUERY_TOO_LONG'
    );
  }

  return true;
}

/**
 * Validate file type filter
 * @param fileType - File type filter value
 * @returns True if valid
 * @throws SearchValidationError if invalid
 */
export function validateFileTypeFilter(fileType: string | null): boolean {
  // Null means no filter - always valid
  if (fileType === null || fileType === '') {
    return true;
  }

  // Must be a string
  if (typeof fileType !== 'string') {
    throw new SearchValidationError(
      '文件类型过滤器格式无效',
      'INVALID_FILE_TYPE_FORMAT'
    );
  }

  // Valid file types
  const validFileTypes = [
    'image',
    'video',
    'audio',
    'document',
    'archive',
    'code',
    'other',
    'all'
  ];

  if (!validFileTypes.includes(fileType.toLowerCase())) {
    throw new SearchValidationError(
      '不支持的文件类型过滤器',
      'INVALID_FILE_TYPE'
    );
  }

  return true;
}

/**
 * Validate size range filter
 * @param sizeMin - Minimum size in bytes
 * @param sizeMax - Maximum size in bytes
 * @returns True if valid
 * @throws SearchValidationError if invalid
 */
export function validateSizeRange(
  sizeMin: number | null,
  sizeMax: number | null
): boolean {
  // Both null means no filter - always valid
  if (sizeMin === null && sizeMax === null) {
    return true;
  }

  // Validate minimum size
  if (sizeMin !== null) {
    if (typeof sizeMin !== 'number' || sizeMin < 0) {
      throw new SearchValidationError(
        '最小文件大小必须是非负数',
        'INVALID_SIZE_MIN'
      );
    }

    // Maximum size limit: 100GB
    const MAX_SIZE = 100 * 1024 * 1024 * 1024;
    if (sizeMin > MAX_SIZE) {
      throw new SearchValidationError(
        '最小文件大小超过限制',
        'SIZE_MIN_TOO_LARGE'
      );
    }
  }

  // Validate maximum size
  if (sizeMax !== null) {
    if (typeof sizeMax !== 'number' || sizeMax < 0) {
      throw new SearchValidationError(
        '最大文件大小必须是非负数',
        'INVALID_SIZE_MAX'
      );
    }

    // Maximum size limit: 100GB
    const MAX_SIZE = 100 * 1024 * 1024 * 1024;
    if (sizeMax > MAX_SIZE) {
      throw new SearchValidationError(
        '最大文件大小超过限制',
        'SIZE_MAX_TOO_LARGE'
      );
    }
  }

  // Validate range consistency
  if (sizeMin !== null && sizeMax !== null && sizeMin > sizeMax) {
    throw new SearchValidationError(
      '最小文件大小不能大于最大文件大小',
      'INVALID_SIZE_RANGE'
    );
  }

  return true;
}

/**
 * Validate date range filter
 * @param dateFrom - Start date (ISO 8601)
 * @param dateTo - End date (ISO 8601)
 * @returns True if valid
 * @throws SearchValidationError if invalid
 */
export function validateDateRange(
  dateFrom: string | null,
  dateTo: string | null
): boolean {
  // Both null means no filter - always valid
  if (dateFrom === null && dateTo === null) {
    return true;
  }

  // Validate start date
  if (dateFrom !== null) {
    if (typeof dateFrom !== 'string') {
      throw new SearchValidationError(
        '开始日期格式无效',
        'INVALID_DATE_FROM_FORMAT'
      );
    }

    const fromDate = new Date(dateFrom);
    if (isNaN(fromDate.getTime())) {
      throw new SearchValidationError(
        '开始日期格式无效',
        'INVALID_DATE_FROM'
      );
    }

    // Date cannot be in the future
    const now = new Date();
    if (fromDate > now) {
      throw new SearchValidationError(
        '开始日期不能是未来时间',
        'DATE_FROM_IN_FUTURE'
      );
    }
  }

  // Validate end date
  if (dateTo !== null) {
    if (typeof dateTo !== 'string') {
      throw new SearchValidationError(
        '结束日期格式无效',
        'INVALID_DATE_TO_FORMAT'
      );
    }

    const toDate = new Date(dateTo);
    if (isNaN(toDate.getTime())) {
      throw new SearchValidationError(
        '结束日期格式无效',
        'INVALID_DATE_TO'
      );
    }

    // Date cannot be in the future
    const now = new Date();
    if (toDate > now) {
      throw new SearchValidationError(
        '结束日期不能是未来时间',
        'DATE_TO_IN_FUTURE'
      );
    }
  }

  // Validate range consistency
  if (dateFrom !== null && dateTo !== null) {
    const fromDate = new Date(dateFrom);
    const toDate = new Date(dateTo);

    if (fromDate > toDate) {
      throw new SearchValidationError(
        '开始日期不能晚于结束日期',
        'INVALID_DATE_RANGE'
      );
    }
  }

  return true;
}

/**
 * Validate pagination parameters
 * @param page - Page number (1-indexed)
 * @param pageSize - Items per page
 * @returns True if valid
 * @throws SearchValidationError if invalid
 */
export function validatePagination(
  page: number | undefined,
  pageSize: number | undefined
): boolean {
  // Validate page number
  if (page !== undefined) {
    if (typeof page !== 'number' || page < 1) {
      throw new SearchValidationError(
        '页码必须是正整数',
        'INVALID_PAGE'
      );
    }

    // Maximum page number: 10000
    if (page > 10000) {
      throw new SearchValidationError(
        '页码超过限制',
        'PAGE_TOO_LARGE'
      );
    }
  }

  // Validate page size
  if (pageSize !== undefined) {
    if (typeof pageSize !== 'number' || pageSize < 1) {
      throw new SearchValidationError(
        '每页数量必须是正整数',
        'INVALID_PAGE_SIZE'
      );
    }

    // Minimum page size: 1
    // Maximum page size: 100
    if (pageSize < 1 || pageSize > 100) {
      throw new SearchValidationError(
        '每页数量必须在1-100之间',
        'INVALID_PAGE_SIZE_RANGE'
      );
    }
  }

  return true;
}

/**
 * Validate complete search request
 * @param query - Search query
 * @param fileType - File type filter
 * @param sizeMin - Minimum size
 * @param sizeMax - Maximum size
 * @param dateFrom - Start date
 * @param dateTo - End date
 * @param page - Page number
 * @param pageSize - Items per page
 * @returns True if all validations pass
 * @throws SearchValidationError if any validation fails
 */
export function validateSearchRequest(
  query: string,
  fileType: string | null = null,
  sizeMin: number | null = null,
  sizeMax: number | null = null,
  dateFrom: string | null = null,
  dateTo: string | null = null,
  page?: number,
  pageSize?: number
): boolean {
  validateSearchQuery(query);
  validateFileTypeFilter(fileType);
  validateSizeRange(sizeMin, sizeMax);
  validateDateRange(dateFrom, dateTo);
  validatePagination(page, pageSize);
  return true;
}

/**
 * Handle search errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleSearchError(error: any): string {
  if (error instanceof SearchValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '搜索参数错误，请检查输入';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限执行搜索';
      case 429:
        return '搜索过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '搜索服务暂时不可用，请稍后重试';
      case 504:
        return '搜索超时，请缩小搜索范围后重试';
      default:
        return data?.message || '搜索失败，请重试';
    }
  }

  // Network errors
  if (error.message === 'Network Error') {
    return '网络连接失败，请检查网络设置';
  }

  // Timeout errors
  if (error.code === 'ECONNABORTED') {
    return '搜索超时，请缩小搜索范围后重试';
  }

  return error.message || '搜索失败，请重试';
}

/**
 * Sanitize search query
 * @param query - Raw search query
 * @returns Sanitized query
 */
export function sanitizeSearchQuery(query: string): string {
  // Trim whitespace
  let sanitized = query.trim();

  // Remove multiple consecutive spaces
  sanitized = sanitized.replace(/\s+/g, ' ');

  // Remove special characters that might cause issues
  // Keep: letters, numbers, spaces, Chinese characters, common punctuation
  sanitized = sanitized.replace(/[^\w\s\u4e00-\u9fa5\-_.]/g, '');

  return sanitized;
}

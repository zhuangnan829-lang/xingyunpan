// Recycle bin validation utilities

/**
 * Validation error class for recycle bin operations
 */
export class RecycleValidationError extends Error {
  constructor(message: string, public code: string) {
    super(message);
    this.name = 'RecycleValidationError';
  }
}

/**
 * Validate item IDs for recycle bin operations
 * @param itemIds - Array of recycle bin item IDs
 * @returns True if valid
 * @throws RecycleValidationError if invalid
 */
export function validateItemIds(itemIds: string[]): boolean {
  // Must be an array
  if (!Array.isArray(itemIds)) {
    throw new RecycleValidationError(
      '项目ID必须是数组',
      'INVALID_ITEM_IDS_FORMAT'
    );
  }

  // Must not be empty
  if (itemIds.length === 0) {
    throw new RecycleValidationError(
      '请选择要操作的项目',
      'NO_ITEMS_SELECTED'
    );
  }

  // Maximum 100 items per operation
  if (itemIds.length > 100) {
    throw new RecycleValidationError(
      '单次操作最多支持100个项目',
      'TOO_MANY_ITEMS'
    );
  }

  // Each item ID must be a non-empty string
  for (const itemId of itemIds) {
    if (typeof itemId !== 'string' || itemId.trim() === '') {
      throw new RecycleValidationError(
        '项目ID格式无效',
        'INVALID_ITEM_ID'
      );
    }
  }

  return true;
}

/**
 * Validate restore path
 * @param path - Restore path
 * @returns True if valid
 * @throws RecycleValidationError if invalid
 */
export function validateRestorePath(path: string | null): boolean {
  // Null means restore to original location - always valid
  if (path === null) {
    return true;
  }

  // Must be a string
  if (typeof path !== 'string') {
    throw new RecycleValidationError(
      '恢复路径格式无效',
      'INVALID_PATH_FORMAT'
    );
  }

  // Path cannot be empty
  if (path.trim() === '') {
    throw new RecycleValidationError(
      '恢复路径不能为空',
      'EMPTY_PATH'
    );
  }

  // Path must start with /
  if (!path.startsWith('/')) {
    throw new RecycleValidationError(
      '恢复路径必须以/开头',
      'INVALID_PATH_START'
    );
  }

  // Path cannot contain invalid characters
  const invalidChars = /[<>:"|?*\x00-\x1f]/;
  if (invalidChars.test(path)) {
    throw new RecycleValidationError(
      '恢复路径包含非法字符',
      'INVALID_PATH_CHARS'
    );
  }

  // Path cannot end with /
  if (path.length > 1 && path.endsWith('/')) {
    throw new RecycleValidationError(
      '恢复路径不能以/结尾',
      'INVALID_PATH_END'
    );
  }

  // Maximum path length: 1000 characters
  if (path.length > 1000) {
    throw new RecycleValidationError(
      '恢复路径过长',
      'PATH_TOO_LONG'
    );
  }

  return true;
}

/**
 * Check for restore path conflicts
 * @param itemName - Name of item being restored
 * @param targetPath - Target restore path
 * @param existingFiles - List of existing file names at target path
 * @returns Conflict resolution suggestion or null if no conflict
 */
export function checkRestorePathConflict(
  itemName: string,
  targetPath: string,
  existingFiles: string[]
): { hasConflict: boolean; suggestedName?: string } {
  // Check if file with same name exists
  if (!existingFiles.includes(itemName)) {
    return { hasConflict: false };
  }

  // Generate suggested name with suffix
  let suggestedName = itemName;
  let counter = 1;

  // Extract file name and extension
  const lastDotIndex = itemName.lastIndexOf('.');
  const hasExtension = lastDotIndex > 0 && lastDotIndex < itemName.length - 1;
  
  const baseName = hasExtension ? itemName.substring(0, lastDotIndex) : itemName;
  const extension = hasExtension ? itemName.substring(lastDotIndex) : '';

  // Find available name
  while (existingFiles.includes(suggestedName)) {
    suggestedName = `${baseName}(${counter})${extension}`;
    counter++;

    // Prevent infinite loop
    if (counter > 1000) {
      throw new RecycleValidationError(
        '无法生成唯一文件名',
        'CANNOT_GENERATE_UNIQUE_NAME'
      );
    }
  }

  return {
    hasConflict: true,
    suggestedName
  };
}

/**
 * Validate permanent delete confirmation
 * @param confirmed - Whether user confirmed the action
 * @returns True if confirmed
 * @throws RecycleValidationError if not confirmed
 */
export function validatePermanentDeleteConfirmation(confirmed: boolean): boolean {
  if (!confirmed) {
    throw new RecycleValidationError(
      '请确认永久删除操作',
      'DELETE_NOT_CONFIRMED'
    );
  }
  return true;
}

/**
 * Validate empty recycle bin confirmation
 * @param confirmed - Whether user confirmed the action
 * @returns True if confirmed
 * @throws RecycleValidationError if not confirmed
 */
export function validateEmptyRecycleBinConfirmation(confirmed: boolean): boolean {
  if (!confirmed) {
    throw new RecycleValidationError(
      '请确认清空回收站操作',
      'EMPTY_NOT_CONFIRMED'
    );
  }
  return true;
}

/**
 * Handle recycle bin restore errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleRestoreError(error: any): string {
  if (error instanceof RecycleValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '恢复参数错误，请检查输入';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限恢复该文件';
      case 404:
        return '文件不存在或已被永久删除';
      case 409:
        return '目标位置已存在同名文件';
      case 413:
        return '恢复的文件总大小超过存储配额';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '恢复文件失败，请重试';
    }
  }

  // Network errors
  if (error.message === 'Network Error') {
    return '网络连接失败，请检查网络设置';
  }

  // Timeout errors
  if (error.code === 'ECONNABORTED') {
    return '请求超时，请检查网络连接';
  }

  return error.message || '恢复文件失败，请重试';
}

/**
 * Handle permanent delete errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handlePermanentDeleteError(error: any): string {
  if (error instanceof RecycleValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '删除参数错误，请检查输入';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限删除该文件';
      case 404:
        return '文件不存在或已被删除';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '永久删除失败，请重试';
    }
  }

  // Network errors
  if (error.message === 'Network Error') {
    return '网络连接失败，请检查网络设置';
  }

  // Timeout errors
  if (error.code === 'ECONNABORTED') {
    return '请求超时，请检查网络连接';
  }

  return error.message || '永久删除失败，请重试';
}

/**
 * Handle empty recycle bin errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleEmptyRecycleBinError(error: any): string {
  if (error instanceof RecycleValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '清空回收站参数错误';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限清空回收站';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '清空回收站失败，请重试';
    }
  }

  // Network errors
  if (error.message === 'Network Error') {
    return '网络连接失败，请检查网络设置';
  }

  // Timeout errors
  if (error.code === 'ECONNABORTED') {
    return '请求超时，请检查网络连接';
  }

  return error.message || '清空回收站失败，请重试';
}

/**
 * Calculate remaining days before permanent deletion
 * @param deletedAt - Deletion timestamp (ISO 8601)
 * @param retentionDays - Number of days to retain (default: 30)
 * @returns Remaining days (can be negative if expired)
 */
export function calculateRemainingDays(
  deletedAt: string,
  retentionDays: number = 30
): number {
  const deletionDate = new Date(deletedAt);
  const expirationDate = new Date(deletionDate);
  expirationDate.setDate(expirationDate.getDate() + retentionDays);
  
  const now = new Date();
  const remainingMs = expirationDate.getTime() - now.getTime();
  const remainingDays = Math.ceil(remainingMs / (1000 * 60 * 60 * 24));
  
  return remainingDays;
}

/**
 * Check if item is expiring soon
 * @param deletedAt - Deletion timestamp (ISO 8601)
 * @param warningDays - Days before expiration to show warning (default: 3)
 * @returns True if expiring soon
 */
export function isExpiringSoon(
  deletedAt: string,
  warningDays: number = 3
): boolean {
  const remainingDays = calculateRemainingDays(deletedAt);
  return remainingDays > 0 && remainingDays <= warningDays;
}

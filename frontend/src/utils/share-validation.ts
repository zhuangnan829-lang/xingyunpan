// Share validation utilities

/**
 * Validation error class for share operations
 */
export class ShareValidationError extends Error {
  constructor(message: string, public code: string) {
    super(message);
    this.name = 'ShareValidationError';
  }
}

/**
 * Validate share expiration date
 * @param expiresIn - Expiration time in seconds (null for permanent)
 * @returns True if valid
 * @throws ShareValidationError if invalid
 */
export function validateExpirationDate(expiresIn: number | null): boolean {
  // Null means permanent share - always valid
  if (expiresIn === null) {
    return true;
  }

  // Must be a positive number
  if (typeof expiresIn !== 'number' || expiresIn <= 0) {
    throw new ShareValidationError(
      '有效期必须是正数',
      'INVALID_EXPIRATION'
    );
  }

  // Maximum expiration: 365 days (31536000 seconds)
  const MAX_EXPIRATION = 365 * 24 * 60 * 60;
  if (expiresIn > MAX_EXPIRATION) {
    throw new ShareValidationError(
      '有效期不能超过365天',
      'EXPIRATION_TOO_LONG'
    );
  }

  // Minimum expiration: 1 hour (3600 seconds)
  const MIN_EXPIRATION = 60 * 60;
  if (expiresIn < MIN_EXPIRATION) {
    throw new ShareValidationError(
      '有效期不能少于1小时',
      'EXPIRATION_TOO_SHORT'
    );
  }

  return true;
}

/**
 * Validate access code format
 * @param accessCode - Access code/password
 * @returns True if valid
 * @throws ShareValidationError if invalid
 */
export function validateAccessCode(accessCode: string | null): boolean {
  // Null or empty means no password - always valid
  if (!accessCode) {
    return true;
  }

  // Must be a string
  if (typeof accessCode !== 'string') {
    throw new ShareValidationError(
      '访问密码格式无效',
      'INVALID_ACCESS_CODE_FORMAT'
    );
  }

  // Length must be between 4 and 8 characters
  if (accessCode.length < 4 || accessCode.length > 8) {
    throw new ShareValidationError(
      '访问密码长度必须为4-8位字符',
      'INVALID_ACCESS_CODE_LENGTH'
    );
  }

  // Only allow alphanumeric characters
  const alphanumericRegex = /^[a-zA-Z0-9]+$/;
  if (!alphanumericRegex.test(accessCode)) {
    throw new ShareValidationError(
      '访问密码只能包含字母和数字',
      'INVALID_ACCESS_CODE_CHARS'
    );
  }

  return true;
}

/**
 * Validate file IDs for sharing
 * @param fileIds - Array of file IDs
 * @returns True if valid
 * @throws ShareValidationError if invalid
 */
export function validateFileIds(fileIds: string[]): boolean {
  // Must be an array
  if (!Array.isArray(fileIds)) {
    throw new ShareValidationError(
      '文件ID必须是数组',
      'INVALID_FILE_IDS_FORMAT'
    );
  }

  // Must not be empty
  if (fileIds.length === 0) {
    throw new ShareValidationError(
      '至少选择一个文件进行分享',
      'NO_FILES_SELECTED'
    );
  }

  // Maximum 100 files per share
  if (fileIds.length > 100) {
    throw new ShareValidationError(
      '单次分享最多支持100个文件',
      'TOO_MANY_FILES'
    );
  }

  // Each file ID must be a non-empty string
  for (const fileId of fileIds) {
    if (typeof fileId !== 'string' || fileId.trim() === '') {
      throw new ShareValidationError(
        '文件ID格式无效',
        'INVALID_FILE_ID'
      );
    }
  }

  return true;
}

/**
 * Validate share creation request
 * @param fileIds - Array of file IDs
 * @param expiresIn - Expiration time in seconds
 * @param accessCode - Access code/password
 * @returns True if all validations pass
 * @throws ShareValidationError if any validation fails
 */
export function validateShareCreation(
  fileIds: string[],
  expiresIn: number | null,
  accessCode: string | null
): boolean {
  validateFileIds(fileIds);
  validateExpirationDate(expiresIn);
  validateAccessCode(accessCode);
  return true;
}

/**
 * Handle share creation errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleShareCreationError(error: any): string {
  if (error instanceof ShareValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '请求参数错误，请检查输入';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限分享该文件';
      case 404:
        return '文件不存在或已被删除';
      case 409:
        return '分享链接生成冲突，请重试';
      case 413:
        return '分享的文件总大小超过限制';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '创建分享失败，请重试';
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

  return error.message || '创建分享失败，请重试';
}

/**
 * Handle share access errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleShareAccessError(error: any): string {
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '请求参数错误';
      case 401:
        return '访问密码错误，请重新输入';
      case 403:
        return '没有权限访问该分享';
      case 404:
        return '分享链接不存在或已被删除';
      case 410:
        return '分享链接已过期';
      case 429:
        return '访问过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '访问分享失败，请重试';
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

  return error.message || '访问分享失败，请重试';
}

/**
 * Validate share token format
 * @param token - Share token
 * @returns True if valid
 */
export function validateShareToken(token: string): boolean {
  if (!token || typeof token !== 'string') {
    return false;
  }

  // Token should be alphanumeric and at least 16 characters
  const tokenRegex = /^[a-zA-Z0-9_-]{16,}$/;
  return tokenRegex.test(token);
}

/**
 * Check if share has expired
 * @param expiresAt - Expiration timestamp (ISO 8601)
 * @returns True if expired
 */
export function isShareExpired(expiresAt: string | null): boolean {
  if (!expiresAt) {
    return false; // Permanent share
  }

  const expirationDate = new Date(expiresAt);
  const now = new Date();

  return now > expirationDate;
}

// Version management validation utilities

/**
 * Validation error class for version management operations
 */
export class VersionValidationError extends Error {
  constructor(message: string, public code: string) {
    super(message);
    this.name = 'VersionValidationError';
  }
}

/**
 * Maximum number of versions per file
 */
export const MAX_VERSIONS_PER_FILE = 10;

/**
 * Validate file ID for version operations
 * @param fileId - File ID
 * @returns True if valid
 * @throws VersionValidationError if invalid
 */
export function validateFileId(fileId: string): boolean {
  if (!fileId || typeof fileId !== 'string' || fileId.trim() === '') {
    throw new VersionValidationError(
      '文件ID无效',
      'INVALID_FILE_ID'
    );
  }
  return true;
}

/**
 * Validate version ID
 * @param versionId - Version ID
 * @returns True if valid
 * @throws VersionValidationError if invalid
 */
export function validateVersionId(versionId: string): boolean {
  if (!versionId || typeof versionId !== 'string' || versionId.trim() === '') {
    throw new VersionValidationError(
      '版本ID无效',
      'INVALID_VERSION_ID'
    );
  }
  return true;
}

/**
 * Validate version count limit
 * @param currentCount - Current number of versions
 * @param maxVersions - Maximum allowed versions (default: 10)
 * @returns True if within limit
 * @throws VersionValidationError if limit exceeded
 */
export function validateVersionLimit(
  currentCount: number,
  maxVersions: number = MAX_VERSIONS_PER_FILE
): boolean {
  if (typeof currentCount !== 'number' || currentCount < 0) {
    throw new VersionValidationError(
      '版本数量格式无效',
      'INVALID_VERSION_COUNT'
    );
  }

  if (currentCount >= maxVersions) {
    throw new VersionValidationError(
      `文件版本数量已达上限（${maxVersions}个）`,
      'VERSION_LIMIT_EXCEEDED'
    );
  }

  return true;
}

/**
 * Check for version restore conflicts
 * @param versionNumber - Version number being restored
 * @param currentVersionNumber - Current version number
 * @returns Conflict information
 */
export function checkVersionRestoreConflict(
  versionNumber: number,
  currentVersionNumber: number
): { hasConflict: boolean; message?: string } {
  // Restoring the current version is not a conflict, but it's pointless
  if (versionNumber === currentVersionNumber) {
    return {
      hasConflict: true,
      message: '该版本已是当前版本，无需恢复'
    };
  }

  // No conflict for restoring older versions
  return { hasConflict: false };
}

/**
 * Validate version restore operation
 * @param fileId - File ID
 * @param versionId - Version ID to restore
 * @param isCurrentVersion - Whether the version is already current
 * @returns True if valid
 * @throws VersionValidationError if invalid
 */
export function validateVersionRestore(
  fileId: string,
  versionId: string,
  isCurrentVersion: boolean
): boolean {
  validateFileId(fileId);
  validateVersionId(versionId);

  if (isCurrentVersion) {
    throw new VersionValidationError(
      '该版本已是当前版本，无需恢复',
      'VERSION_ALREADY_CURRENT'
    );
  }

  return true;
}

/**
 * Validate version download operation
 * @param fileId - File ID
 * @param versionId - Version ID to download
 * @returns True if valid
 * @throws VersionValidationError if invalid
 */
export function validateVersionDownload(
  fileId: string,
  versionId: string
): boolean {
  validateFileId(fileId);
  validateVersionId(versionId);
  return true;
}

/**
 * Validate version deletion operation
 * @param fileId - File ID
 * @param versionId - Version ID to delete
 * @param isCurrentVersion - Whether the version is current
 * @param totalVersions - Total number of versions
 * @returns True if valid
 * @throws VersionValidationError if invalid
 */
export function validateVersionDeletion(
  fileId: string,
  versionId: string,
  isCurrentVersion: boolean,
  totalVersions: number
): boolean {
  validateFileId(fileId);
  validateVersionId(versionId);

  // Cannot delete the current version
  if (isCurrentVersion) {
    throw new VersionValidationError(
      '不能删除当前版本',
      'CANNOT_DELETE_CURRENT_VERSION'
    );
  }

  // Must have at least 2 versions to delete one
  if (totalVersions <= 1) {
    throw new VersionValidationError(
      '文件只有一个版本，无法删除',
      'CANNOT_DELETE_ONLY_VERSION'
    );
  }

  return true;
}

/**
 * Handle version management errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleVersionError(error: any): string {
  if (error instanceof VersionValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '版本操作参数错误，请检查输入';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限操作该文件版本';
      case 404:
        return '文件或版本不存在';
      case 409:
        return '版本操作冲突，请刷新后重试';
      case 413:
        return '文件大小超过存储配额限制';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '版本操作失败，请重试';
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

  return error.message || '版本操作失败，请重试';
}

/**
 * Handle version restore errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleVersionRestoreError(error: any): string {
  if (error instanceof VersionValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '恢复版本参数错误';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限恢复该版本';
      case 404:
        return '文件或版本不存在';
      case 409:
        return '版本恢复冲突，该版本可能已是当前版本';
      case 413:
        return '恢复版本会超过存储配额限制';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '恢复版本失败，请重试';
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

  return error.message || '恢复版本失败，请重试';
}

/**
 * Handle version download errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleVersionDownloadError(error: any): string {
  if (error instanceof VersionValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '下载版本参数错误';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限下载该版本';
      case 404:
        return '文件或版本不存在';
      case 429:
        return '下载过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '下载版本失败，请重试';
    }
  }

  // Network errors
  if (error.message === 'Network Error') {
    return '网络连接失败，请检查网络设置';
  }

  // Timeout errors
  if (error.code === 'ECONNABORTED') {
    return '下载超时，请检查网络连接';
  }

  return error.message || '下载版本失败，请重试';
}

/**
 * Calculate total storage used by versions
 * @param versions - Array of file versions
 * @returns Total storage in bytes
 */
export function calculateVersionStorage(versions: Array<{ file_size: number }>): number {
  return versions.reduce((total, version) => total + version.file_size, 0);
}

/**
 * Format version number for display
 * @param versionNumber - Version number
 * @returns Formatted version string
 */
export function formatVersionNumber(versionNumber: number): string {
  return `v${versionNumber}`;
}

/**
 * Compare two version numbers
 * @param v1 - First version number
 * @param v2 - Second version number
 * @returns -1 if v1 < v2, 0 if equal, 1 if v1 > v2
 */
export function compareVersions(v1: number, v2: number): number {
  if (v1 < v2) return -1;
  if (v1 > v2) return 1;
  return 0;
}

/**
 * Check if version limit will be exceeded after creating a new version
 * @param currentCount - Current number of versions
 * @param maxVersions - Maximum allowed versions
 * @returns True if limit will be exceeded
 */
export function willExceedVersionLimit(
  currentCount: number,
  maxVersions: number = MAX_VERSIONS_PER_FILE
): boolean {
  return currentCount >= maxVersions;
}

/**
 * Get oldest version that will be auto-deleted
 * @param versions - Array of versions sorted by version number
 * @param maxVersions - Maximum allowed versions
 * @returns Version that will be deleted, or null if no deletion needed
 */
export function getVersionToAutoDelete<T extends { version_number: number }>(
  versions: T[],
  maxVersions: number = MAX_VERSIONS_PER_FILE
): T | null {
  if (versions.length < maxVersions) {
    return null;
  }

  // Find the oldest version (lowest version number)
  return versions.reduce((oldest, current) => {
    return current.version_number < oldest.version_number ? current : oldest;
  });
}

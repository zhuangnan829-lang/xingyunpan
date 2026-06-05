// Collaboration validation utilities

/**
 * Validation error class for collaboration operations
 */
export class CollaborationValidationError extends Error {
  constructor(message: string, public code: string) {
    super(message);
    this.name = 'CollaborationValidationError';
  }
}

/**
 * Valid permission levels
 */
export type PermissionLevel = 'view' | 'download' | 'edit';

/**
 * Permission level hierarchy (higher number = more permissions)
 */
const PERMISSION_HIERARCHY: Record<PermissionLevel, number> = {
  view: 1,
  download: 2,
  edit: 3
};

/**
 * Validate file ID for collaboration operations
 * @param fileId - File ID
 * @returns True if valid
 * @throws CollaborationValidationError if invalid
 */
export function validateFileId(fileId: string): boolean {
  if (!fileId || typeof fileId !== 'string' || fileId.trim() === '') {
    throw new CollaborationValidationError(
      '文件ID无效',
      'INVALID_FILE_ID'
    );
  }
  return true;
}

/**
 * Validate user ID
 * @param userId - User ID
 * @returns True if valid
 * @throws CollaborationValidationError if invalid
 */
export function validateUserId(userId: string): boolean {
  if (!userId || typeof userId !== 'string' || userId.trim() === '') {
    throw new CollaborationValidationError(
      '用户ID无效',
      'INVALID_USER_ID'
    );
  }
  return true;
}

/**
 * Validate username
 * @param username - Username
 * @returns True if valid
 * @throws CollaborationValidationError if invalid
 */
export function validateUsername(username: string): boolean {
  if (!username || typeof username !== 'string') {
    throw new CollaborationValidationError(
      '用户名格式无效',
      'INVALID_USERNAME_FORMAT'
    );
  }

  const trimmedUsername = username.trim();

  // Minimum length: 3 characters
  if (trimmedUsername.length < 3) {
    throw new CollaborationValidationError(
      '用户名长度不能少于3个字符',
      'USERNAME_TOO_SHORT'
    );
  }

  // Maximum length: 50 characters
  if (trimmedUsername.length > 50) {
    throw new CollaborationValidationError(
      '用户名长度不能超过50个字符',
      'USERNAME_TOO_LONG'
    );
  }

  // Valid characters: letters, numbers, underscore, hyphen
  const usernameRegex = /^[a-zA-Z0-9_-]+$/;
  if (!usernameRegex.test(trimmedUsername)) {
    throw new CollaborationValidationError(
      '用户名只能包含字母、数字、下划线和连字符',
      'INVALID_USERNAME_CHARS'
    );
  }

  return true;
}

/**
 * Validate permission level
 * @param permission - Permission level
 * @returns True if valid
 * @throws CollaborationValidationError if invalid
 */
export function validatePermission(permission: string): boolean {
  if (!permission || typeof permission !== 'string') {
    throw new CollaborationValidationError(
      '权限级别格式无效',
      'INVALID_PERMISSION_FORMAT'
    );
  }

  const validPermissions: PermissionLevel[] = ['view', 'download', 'edit'];
  if (!validPermissions.includes(permission as PermissionLevel)) {
    throw new CollaborationValidationError(
      '权限级别无效，必须是 view、download 或 edit',
      'INVALID_PERMISSION_LEVEL'
    );
  }

  return true;
}

/**
 * Validate add collaborator request
 * @param fileId - File ID
 * @param username - Username to add
 * @param permission - Permission level
 * @returns True if valid
 * @throws CollaborationValidationError if invalid
 */
export function validateAddCollaborator(
  fileId: string,
  username: string,
  permission: string
): boolean {
  validateFileId(fileId);
  validateUsername(username);
  validatePermission(permission);
  return true;
}

/**
 * Validate update permission request
 * @param fileId - File ID
 * @param userId - User ID
 * @param permission - New permission level
 * @returns True if valid
 * @throws CollaborationValidationError if invalid
 */
export function validateUpdatePermission(
  fileId: string,
  userId: string,
  permission: string
): boolean {
  validateFileId(fileId);
  validateUserId(userId);
  validatePermission(permission);
  return true;
}

/**
 * Validate remove collaborator request
 * @param fileId - File ID
 * @param userId - User ID to remove
 * @returns True if valid
 * @throws CollaborationValidationError if invalid
 */
export function validateRemoveCollaborator(
  fileId: string,
  userId: string
): boolean {
  validateFileId(fileId);
  validateUserId(userId);
  return true;
}

/**
 * Check for permission conflicts
 * @param currentPermission - Current permission level
 * @param newPermission - New permission level
 * @returns Conflict information
 */
export function checkPermissionConflict(
  currentPermission: PermissionLevel,
  newPermission: PermissionLevel
): { hasConflict: boolean; message?: string } {
  // Same permission is not a conflict, but it's pointless
  if (currentPermission === newPermission) {
    return {
      hasConflict: true,
      message: '该用户已拥有此权限级别'
    };
  }

  // No conflict for changing permissions
  return { hasConflict: false };
}

/**
 * Check if permission level allows an action
 * @param permission - User's permission level
 * @param requiredPermission - Required permission level for action
 * @returns True if allowed
 */
export function hasPermission(
  permission: PermissionLevel,
  requiredPermission: PermissionLevel
): boolean {
  return PERMISSION_HIERARCHY[permission] >= PERMISSION_HIERARCHY[requiredPermission];
}

/**
 * Get permission description
 * @param permission - Permission level
 * @returns Human-readable description
 */
export function getPermissionDescription(permission: PermissionLevel): string {
  const descriptions: Record<PermissionLevel, string> = {
    view: '仅查看：可以查看文件信息和预览',
    download: '下载：可以查看和下载文件',
    edit: '编辑：可以查看、下载、上传新版本和重命名文件'
  };
  return descriptions[permission] || '未知权限';
}

/**
 * Get allowed actions for permission level
 * @param permission - Permission level
 * @returns Array of allowed actions
 */
export function getAllowedActions(permission: PermissionLevel): string[] {
  const actions: Record<PermissionLevel, string[]> = {
    view: ['查看文件信息', '预览文件'],
    download: ['查看文件信息', '预览文件', '下载文件'],
    edit: ['查看文件信息', '预览文件', '下载文件', '上传新版本', '重命名文件']
  };
  return actions[permission] || [];
}

/**
 * Handle collaboration errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleCollaborationError(error: any): string {
  if (error instanceof CollaborationValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '协作操作参数错误，请检查输入';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '没有权限执行此协作操作';
      case 404:
        return '文件或用户不存在';
      case 409:
        return '协作关系已存在或发生冲突';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '协作操作失败，请重试';
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

  return error.message || '协作操作失败，请重试';
}

/**
 * Handle add collaborator errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleAddCollaboratorError(error: any): string {
  if (error instanceof CollaborationValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '添加协作者参数错误';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '只有文件所有者可以添加协作者';
      case 404:
        return data?.message || '用户不存在';
      case 409:
        return '该用户已是协作者';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '添加协作者失败，请重试';
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

  return error.message || '添加协作者失败，请重试';
}

/**
 * Handle update permission errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleUpdatePermissionError(error: any): string {
  if (error instanceof CollaborationValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '更新权限参数错误';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '只有文件所有者可以更改协作者权限';
      case 404:
        return '协作关系不存在';
      case 409:
        return '权限更新冲突，请刷新后重试';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '更新权限失败，请重试';
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

  return error.message || '更新权限失败，请重试';
}

/**
 * Handle remove collaborator errors
 * @param error - Error object
 * @returns User-friendly error message
 */
export function handleRemoveCollaboratorError(error: any): string {
  if (error instanceof CollaborationValidationError) {
    return error.message;
  }

  // Handle API errors
  if (error.response) {
    const status = error.response.status;
    const data = error.response.data;

    switch (status) {
      case 400:
        return data?.message || '移除协作者参数错误';
      case 401:
        return '未登录或登录已过期，请重新登录';
      case 403:
        return '只有文件所有者可以移除协作者';
      case 404:
        return '协作关系不存在';
      case 429:
        return '操作过于频繁，请稍后再试';
      case 500:
        return '服务器错误，请稍后重试';
      case 503:
        return '服务暂时不可用，请稍后重试';
      default:
        return data?.message || '移除协作者失败，请重试';
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

  return error.message || '移除协作者失败，请重试';
}

/**
 * Validate that user is not adding themselves as collaborator
 * @param username - Username to add
 * @param currentUsername - Current user's username
 * @returns True if valid
 * @throws CollaborationValidationError if trying to add self
 */
export function validateNotSelf(username: string, currentUsername: string): boolean {
  if (username.toLowerCase() === currentUsername.toLowerCase()) {
    throw new CollaborationValidationError(
      '不能将自己添加为协作者',
      'CANNOT_ADD_SELF'
    );
  }
  return true;
}

/**
 * Format permission level for display
 * @param permission - Permission level
 * @returns Formatted permission string
 */
export function formatPermissionLevel(permission: PermissionLevel): string {
  const labels: Record<PermissionLevel, string> = {
    view: '查看',
    download: '下载',
    edit: '编辑'
  };
  return labels[permission] || permission;
}

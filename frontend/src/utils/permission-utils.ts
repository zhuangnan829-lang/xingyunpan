/**
 * Permission utility functions
 * Provides helper functions for permission checking and formatting
 */

import type { PermissionCheck } from '@/api/collaboration'

export type PermissionLevel = 'view' | 'download' | 'edit'
export type Action = 'view' | 'download' | 'edit' | 'delete' | 'share' | 'collaborate'

/**
 * Permission matrix defining what each permission level allows
 */
const PERMISSION_MATRIX: Record<PermissionLevel, Set<Action>> = {
  view: new Set(['view']),
  download: new Set(['view', 'download']),
  edit: new Set(['view', 'download', 'edit'])
}

/**
 * Check if a permission level allows a specific action
 * @param permission - The permission level
 * @param action - The action to check
 * @returns true if action is allowed
 */
export function hasPermission(permission: PermissionLevel, action: Action): boolean {
  const allowedActions = PERMISSION_MATRIX[permission]
  return allowedActions.has(action)
}

/**
 * Check if user can perform an action based on permission check result
 * @param permissionCheck - Permission check result from API
 * @param action - The action to check
 * @returns true if action is allowed
 */
export function canPerformAction(permissionCheck: PermissionCheck, action: Action): boolean {
  // Owner can do everything
  if (permissionCheck.is_owner) {
    return true
  }
  
  // Check specific permissions
  switch (action) {
    case 'view':
      return permissionCheck.can_view
    case 'download':
      return permissionCheck.can_download
    case 'edit':
      return permissionCheck.can_edit
    case 'delete':
    case 'share':
    case 'collaborate':
      // Only owner can delete, share, or manage collaborators
      return permissionCheck.is_owner
    default:
      return false
  }
}

/**
 * Format permission level for display
 * @param permission - The permission level
 * @returns Localized permission description
 */
export function formatPermissionLevel(permission: PermissionLevel): string {
  const descriptions: Record<PermissionLevel, string> = {
    view: '查看',
    download: '下载',
    edit: '编辑'
  }
  
  return descriptions[permission] || permission
}

/**
 * Get detailed permission description
 * @param permission - The permission level
 * @returns Detailed description of what the permission allows
 */
export function getPermissionDescription(permission: PermissionLevel): string {
  const descriptions: Record<PermissionLevel, string> = {
    view: '可以查看文件信息和预览内容',
    download: '可以查看和下载文件',
    edit: '可以查看、下载、上传新版本和重命名文件'
  }
  
  return descriptions[permission] || ''
}

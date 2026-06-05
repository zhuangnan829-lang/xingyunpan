/**
 * Recycle bin utility functions
 * Provides helper functions for recycle bin item management and date calculations
 */

/**
 * Calculate remaining days before permanent deletion
 * @param expiresAt - ISO 8601 expiration timestamp
 * @returns Number of days remaining (0 if expired)
 */
export function calculateRemainingDays(expiresAt: string): number {
  const expireDate = new Date(expiresAt)
  const now = new Date()
  
  // Calculate difference in milliseconds
  const diffMs = expireDate.getTime() - now.getTime()
  
  // Convert to days and round up
  const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24))
  
  // Return 0 if already expired
  return Math.max(0, diffDays)
}

/**
 * Check if a recycle bin item is expiring soon (within 3 days)
 * @param expiresAt - ISO 8601 expiration timestamp
 * @returns true if expiring within 3 days
 */
export function isExpiringSoon(expiresAt: string): boolean {
  const remainingDays = calculateRemainingDays(expiresAt)
  return remainingDays > 0 && remainingDays <= 3
}

/**
 * Format deletion date for display
 * @param deletedAt - ISO 8601 deletion timestamp
 * @returns Formatted date string
 */
export function formatDeletionDate(deletedAt: string): string {
  const deleteDate = new Date(deletedAt)
  const now = new Date()
  
  // Calculate time difference
  const diffMs = now.getTime() - deleteDate.getTime()
  const diffMinutes = Math.floor(diffMs / (1000 * 60))
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  
  // Format based on time elapsed
  if (diffMinutes < 1) {
    return '刚刚'
  } else if (diffMinutes < 60) {
    return `${diffMinutes}分钟前`
  } else if (diffHours < 24) {
    return `${diffHours}小时前`
  } else if (diffDays < 7) {
    return `${diffDays}天前`
  } else {
    // Format as date for older items
    const year = deleteDate.getFullYear()
    const month = String(deleteDate.getMonth() + 1).padStart(2, '0')
    const day = String(deleteDate.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }
}

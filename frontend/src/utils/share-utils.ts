/**
 * Share utility functions
 * Provides helper functions for share link generation, validation, and formatting
 */

/**
 * Generate a complete share link URL from share token
 * @param shareToken - The share token from backend
 * @param baseUrl - Optional base URL (defaults to current origin)
 * @returns Complete share link URL
 */
export function generateShareLink(shareToken: string, baseUrl?: string): string {
  const base = baseUrl || window.location.origin
  return `${base}/s/${shareToken}`
}

/**
 * Validate share token format
 * Share tokens should be alphanumeric strings of sufficient length
 * @param token - The share token to validate
 * @returns true if token format is valid
 */
export function validateShareToken(token: string): boolean {
  if (!token || typeof token !== 'string') {
    return false
  }
  
  // Token should be at least 16 characters for security
  if (token.length < 16) {
    return false
  }
  
  // Token should only contain alphanumeric characters and hyphens
  const tokenRegex = /^[a-zA-Z0-9-_]+$/
  return tokenRegex.test(token)
}

/**
 * Format share expiration time for display
 * @param expiresAt - ISO 8601 expiration timestamp or null for permanent
 * @returns Formatted expiration string
 */
export function formatShareExpiration(expiresAt: string | null): string {
  if (!expiresAt) {
    return '永久有效'
  }
  
  const expireDate = new Date(expiresAt)
  const now = new Date()
  
  // Check if expired
  if (expireDate <= now) {
    return '已过期'
  }
  
  // Calculate time difference
  const diffMs = expireDate.getTime() - now.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  const diffHours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const diffMinutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
  
  // Format based on time remaining
  if (diffDays > 0) {
    return `${diffDays}天后过期`
  } else if (diffHours > 0) {
    return `${diffHours}小时后过期`
  } else if (diffMinutes > 0) {
    return `${diffMinutes}分钟后过期`
  } else {
    return '即将过期'
  }
}

/**
 * Check if a share has expired
 * @param expiresAt - ISO 8601 expiration timestamp or null for permanent
 * @returns true if share has expired
 */
export function isShareExpired(expiresAt: string | null): boolean {
  if (!expiresAt) {
    return false // Permanent shares never expire
  }
  
  const expireDate = new Date(expiresAt)
  const now = new Date()
  
  return expireDate <= now
}

/**
 * Validate share access code (password)
 * Access codes should be 4-8 characters
 * @param password - The access code to validate
 * @returns true if password format is valid
 */
export function validatePassword(password: string): boolean {
  if (!password || typeof password !== 'string') {
    return false
  }
  
  // Password should be 4-8 characters
  return password.length >= 4 && password.length <= 8
}

/**
 * Copy text to clipboard
 * @param text - The text to copy
 * @returns Promise<boolean> indicating whether copy succeeded
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text)
    } else {
      // Fallback for older browsers
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      
      try {
        document.execCommand('copy')
        textArea.remove()
      } catch (err) {
        textArea.remove()
        throw new Error('复制失败')
      }
    }
    return true
  } catch (err) {
    throw new Error('复制到剪贴板失败')
  }
}

// Share-related TypeScript type definitions

/**
 * Share link data model
 */
export interface ShareLink {
  id: string;                    // Share ID (nanoid, 12 characters)
  fileIds: string[];             // File ID list
  fileNames: string[];           // File name list (for display)
  createdAt: string;             // Creation time (ISO 8601)
  expiresAt: string | null;      // Expiration time, null means permanent
  password: string | null;       // Access password, null means no password
  accessCount: number;           // Access count
  creatorId: string;             // Creator user ID
}

/**
 * Options for creating a share link
 */
export interface CreateShareOptions {
  fileIds: string[];             // File IDs to share
  expiresIn?: number | null;     // Validity period (seconds), null means permanent
  password?: string | null;      // Access password (4-8 characters)
}

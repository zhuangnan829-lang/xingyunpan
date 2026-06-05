/**
 * Recycle bin item data model
 * Represents a deleted file in the recycle bin
 */
export interface RecycleItem {
  id: string;                    // Recycle bin item ID
  fileId: string;                // Original file ID
  fileName: string;              // File name
  fileSize: number;              // File size in bytes
  fileType: string;              // File type (image/video/document etc.)
  originalPath: string;          // Original path for restoration
  deletedAt: string;             // Deletion time (ISO 8601)
  expiresAt: string;             // Expiration time (deletedAt + 30 days)
  userId: string;                // User ID
}

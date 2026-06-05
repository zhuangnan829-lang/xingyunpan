// Version management API module for file version history

import request from './request';

/**
 * Interface for file version information
 */
export interface FileVersion {
  version_id: string;
  version_number: number;
  file_size: number;
  created_at: string;
  uploader_id: string;
  uploader_name: string;
  is_current: boolean;
}

/**
 * Get version history for a file
 * @param fileId - File ID
 * @returns Array of file versions
 */
export async function getVersionHistory(fileId: string): Promise<FileVersion[]> {
  return request.get(`/api/files/${fileId}/versions`);
}

/**
 * Download a specific version of a file
 * @param fileId - File ID
 * @param versionId - Version ID
 * @returns Blob data for the file version
 */
export async function downloadVersion(fileId: string, versionId: string): Promise<Blob> {
  const response = await request.get(`/api/files/${fileId}/versions/${versionId}/download`, {
    responseType: 'blob'
  });
  return response;
}

/**
 * Restore a previous version (creates a new version)
 * @param fileId - File ID
 * @param versionId - Version ID to restore
 */
export async function restoreVersion(fileId: string, versionId: string): Promise<void> {
  return request.post(`/api/files/${fileId}/versions/${versionId}/restore`);
}

/**
 * Delete a specific version
 * @param fileId - File ID
 * @param versionId - Version ID to delete
 */
export async function deleteVersion(fileId: string, versionId: string): Promise<void> {
  return request.delete(`/api/files/${fileId}/versions/${versionId}`);
}

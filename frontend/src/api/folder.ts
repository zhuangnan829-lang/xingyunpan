// Folder API module

import request from './request';
import { FolderItem, FolderPathItem, CreateFolderRequest } from '@/types/folder';

/**
 * Create folder
 * @param name - Folder name
 * @param parentId - Parent folder ID (null for root directory)
 */
export async function createFolder(name: string, parentId?: number | null): Promise<FolderItem> {
  const data: CreateFolderRequest = { name };
  
  if (parentId !== undefined && parentId !== null) {
    data.parent_id = parentId;
  }
  
  const response = await request.post<any>('/api/v1/folder', data);
  
  // Transform backend response to frontend format
  return {
    id: response.id,
    name: response.name,
    parent_id: response.parent_id,
    created_at: response.created_at,
    updated_at: response.updated_at,
  };
}

/**
 * Rename folder
 * @param folderId - Folder ID
 * @param newName - New folder name
 */
export async function renameFolder(folderId: number, newName: string): Promise<void> {
  await request.put(`/api/v1/folder/${folderId}/rename`, { name: newName });
}

/**
 * Delete folder
 * @param folderId - Folder ID
 */
export async function deleteFolder(folderId: number): Promise<void> {
  await request.delete(`/api/v1/folder/${folderId}`);
}

/**
 * Move folder to another parent folder
 * @param folderId - Folder ID
 * @param targetFolderId - Target parent folder ID (null for root)
 */
export async function moveFolder(folderId: number, targetFolderId: number | null): Promise<void> {
  await request.put(`/api/v1/folder/${folderId}/move`, { parent_id: targetFolderId });
}

/**
 * Copy folder to another parent folder
 * @param folderId - Folder ID
 * @param targetFolderId - Target parent folder ID (null for root)
 * 
 * Note: This endpoint needs to be implemented in the backend
 * Expected endpoint: POST /api/v1/folder/:id/copy with body { parent_id: number | null }
 */
export async function copyFolder(folderId: number, targetFolderId: number | null): Promise<void> {
  await request.post(`/api/v1/folder/${folderId}/copy`, { parent_id: targetFolderId });
}

/**
 * Get folder path (breadcrumb navigation)
 * @param folderId - Folder ID
 * @returns Array of folder path items from root to current folder
 * 
 * Note: This endpoint needs to be implemented in the backend
 * Expected endpoint: GET /api/v1/folder/:id/path
 * Expected response: Array of { id: number, name: string }
 */
export async function getFolderPath(folderId: number): Promise<FolderPathItem[]> {
  const response = await request.get<FolderPathItem[]>(`/api/v1/folder/${folderId}/path`);
  return response;
}

// File API module

import request from './request';
import { FileItem, FileListResponse } from '@/types/file';
import { getToken } from '@/utils/auth';

export interface FileCustomPropertyDefinition {
  id: number;
  key: string;
  name: string;
  icon: string;
  type: 'text' | 'rating' | 'switch' | 'date' | 'tags' | 'multi_select';
  minLength: number | null;
  maxLength: number | null;
  maxValue: number | null;
  options?: string[];
  defaultValue: string;
}

export interface FileCustomPropertyPayload {
  file_id: number;
  definitions: FileCustomPropertyDefinition[];
  values: Record<string, string>;
  last_modified?: string | null;
}

/**
 * Get files and folders list
 * @param folderId - Parent folder ID (null for root directory)
 * @param page - Page number (default: 1)
 * @param pageSize - Page size (default: 20)
 */
export async function getFiles(
  folderId?: number | null,
  page: number = 1,
  pageSize: number = 20
): Promise<FileListResponse> {
  const params: any = { page, page_size: pageSize };
  
  if (folderId !== undefined && folderId !== null) {
    params.parent_id = folderId;
  }
  
  const response = await request.get<any>('/api/v1/file/list', { params });
  
  // Transform backend response to frontend format
  // Backend returns: { list: [...], total, page, page_size }
  // Frontend expects: { files: [...], folders: [...] }
  const items = response.list || [];
  
  const files: FileItem[] = [];
  const folders: any[] = [];
  
  items.forEach((item: any) => {
    if (item.is_folder) {
      folders.push({
        id: item.id,
        name: item.name,
        parent_id: item.parent_id,
        created_at: item.created_at,
        updated_at: item.updated_at,
        directory_stats: item.directory_stats || null,
        display_icon: item.display_icon || '',
        display_icon_tint: item.display_icon_tint || '',
        display_icon_label: item.display_icon_label || '',
        display_icon_source: item.display_icon_source || '',
        encryption_status: item.encryption_status || null,
      });
    } else {
      files.push({
        id: item.id,
        name: item.name,
        size: item.size || 0,
        hash: item.hash || '',
        mime_type: item.mime_type || 'application/octet-stream',
        content_type: item.content_type || item.mime_type || 'application/octet-stream',
        physical_file_id: item.physical_file_id,
        thumbnail_url: item.thumbnail_url,
        folder_id: item.parent_id,
        created_at: item.created_at,
        updated_at: item.updated_at,
        browser_app: item.browser_app || null,
        display_icon: item.display_icon || '',
        display_icon_tint: item.display_icon_tint || '',
        display_icon_label: item.display_icon_label || '',
        display_icon_source: item.display_icon_source || '',
      });
    }
  });
  
  return { files, folders };
}

/**
 * Upload file (for small files < 5MB)
 * @param file - File to upload
 * @param hash - File SHA256 hash
 * @param folderId - Target folder ID (optional)
 */
export async function uploadFile(
  file: File,
  hash: string,
  folderId?: number
): Promise<FileItem> {
  const formData = new FormData();
  formData.append('file', file);
  
  if (folderId !== undefined) {
    formData.append('parent_id', folderId.toString());
  }
  
  const response = await request.post<any>('/api/v1/file/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  
  // Transform backend response to frontend format
  return {
    id: response.id,
    name: response.name,
    size: response.size || 0,
    hash: response.hash || hash,
    mime_type: response.mime_type || 'application/octet-stream',
    content_type: response.content_type || response.mime_type || 'application/octet-stream',
    physical_file_id: response.physical_file_id,
    thumbnail_url: response.thumbnail_url,
    folder_id: response.parent_id,
    created_at: response.created_at,
    updated_at: response.updated_at,
    browser_app: response.browser_app || null,
  };
}

export async function createFile(
  kind: 'file' | 'markdown' | 'text' | 'drawio' | 'dwb' | 'excalidraw',
  name?: string,
  parentId?: number | null,
  content?: string,
): Promise<FileItem> {
  const response = await request.post<any>('/api/v1/file/create', {
    kind,
    name,
    parent_id: parentId ?? null,
    content,
  });

  return {
    id: response.id,
    name: response.name,
    size: response.size || 0,
    hash: response.hash || '',
    mime_type: response.mime_type || 'application/octet-stream',
    content_type: response.content_type || response.mime_type || 'application/octet-stream',
    physical_file_id: response.physical_file_id,
    thumbnail_url: response.thumbnail_url,
    folder_id: response.parent_id,
    created_at: response.created_at,
    updated_at: response.updated_at,
    browser_app: response.browser_app || null,
  };
}

export async function getFileCustomProperties(fileId: number): Promise<FileCustomPropertyPayload> {
  return request.get<FileCustomPropertyPayload>(`/api/v1/file/${fileId}/custom-properties`);
}

export async function updateFileCustomProperties(fileId: number, values: Record<string, string>): Promise<FileCustomPropertyPayload> {
  return request.put<FileCustomPropertyPayload>(`/api/v1/file/${fileId}/custom-properties`, { values });
}

/**
 * Download file
 * @param fileId - File ID
 * @returns Blob containing file data
 * 
 * Note: This endpoint needs to be implemented in the backend
 * Expected endpoint: GET /api/v1/file/download/:id
 */
export function getFileDownloadUrl(fileId: number, inline: boolean = false): string {
  return `/api/v1/file/${fileId}/download${inline ? '?inline=1' : ''}`;
}

export function getAuthenticatedFileDownloadUrl(fileId: number, inline: boolean = false): string {
  const url = new URL(getFileDownloadUrl(fileId, inline), window.location.origin);
  const token = getToken();
  if (token) {
    url.searchParams.set('access_token', token);
  }
  return url.toString();
}

export function getExternalPreviewUrl(file: FileItem): string {
  const sourceUrl = getAuthenticatedFileDownloadUrl(file.id, true);
  const appName = file.browser_app?.name || '';

  if (/google docs|google/i.test(appName)) {
    return `https://docs.google.com/gview?embedded=true&url=${encodeURIComponent(sourceUrl)}`;
  }

  if (/office|microsoft/i.test(appName)) {
    return `https://view.officeapps.live.com/op/embed.aspx?src=${encodeURIComponent(sourceUrl)}`;
  }

  return sourceUrl;
}

export async function downloadFile(fileId: number): Promise<Blob> {
  const response = await request.get(getFileDownloadUrl(fileId), {
    responseType: 'blob',
  });
  
  return response;
}

export async function previewFileAsPdf(fileId: number): Promise<Blob> {
  return request.get(`/api/v1/file/${fileId}/preview-pdf`, {
    responseType: 'blob',
    timeout: 120000,
  });
}

/**
 * Rename file
 * @param fileId - File ID
 * @param newName - New file name
 * 
 * Note: This endpoint needs to be implemented in the backend
 * Expected endpoint: PUT /api/v1/file/:id with body { name: string }
 */
export async function renameFile(fileId: number, newName: string): Promise<void> {
  await request.put(`/api/v1/file/${fileId}`, { name: newName });
}

/**
 * Delete file
 * @param fileId - File ID
 */
export async function deleteFile(fileId: number): Promise<void> {
  await request.delete(`/api/v1/file/${fileId}`);
}

/**
 * Move file to another folder
 * @param fileId - File ID
 * @param targetFolderId - Target folder ID (null for root)
 * 
 * Note: This endpoint needs to be implemented in the backend
 * Expected endpoint: PUT /api/v1/file/:id/move with body { parent_id: number | null }
 */
export async function moveFile(fileId: number, targetFolderId: number | null): Promise<void> {
  await request.put(`/api/v1/file/${fileId}/move`, { parent_id: targetFolderId });
}

/**
 * Copy file to another folder
 * @param fileId - File ID
 * @param targetFolderId - Target folder ID (null for root)
 * 
 * Note: This endpoint needs to be implemented in the backend
 * Expected endpoint: POST /api/v1/file/:id/copy with body { parent_id: number | null }
 */
export async function copyFile(fileId: number, targetFolderId: number | null): Promise<void> {
  await request.post(`/api/v1/file/${fileId}/copy`, { parent_id: targetFolderId });
}

/**
 * Check if file exists (for instant upload)
 * @param hash - File SHA256 hash
 * @returns Object with exists flag and file_id if exists
 */
export async function checkFileExists(hash: string): Promise<{ exists: boolean; file_id?: number }> {
  const response = await request.post<any>('/api/v1/file/check', { hash });
  return response;
}

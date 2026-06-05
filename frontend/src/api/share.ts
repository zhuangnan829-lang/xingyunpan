// Share API module for managing file sharing functionality

import request from './request';
import { resolveAPIBaseURL } from '@/utils/public-url';

/**
 * Request interface for creating a share link
 */
export interface CreateShareRequest {
  file_ids: string[];
  expires_in?: number | null;    // Expiration time in seconds, null for permanent
  access_code?: string | null;   // Access password
  max_downloads?: number | null;
}

/**
 * Response interface for share creation
 */
export interface CreateShareResponse {
  share_id: string;
  share_token: string;
  share_url: string;
  expires_at: string | null;
  max_downloads?: number | null;
}

/**
 * Interface for share information
 */
export interface ShareInfo {
  share_id: string;
  file_ids: string[];
  file_names: string[];
  creator_name: string;
  created_at: string;
  expires_at: string | null;
  has_password: boolean;
  max_downloads?: number | null;
  download_count: number;
}

/**
 * Interface for user's share list item
 */
export interface MyShare {
  share_id: string;
  share_token?: string;
  share_url?: string;
  file_ids: string[];
  file_names: string[];
  created_at: string;
  expires_at: string | null;
  has_password?: boolean;
  max_downloads?: number | null;
  download_count: number;
  access_count: number;
}

/**
 * Request interface for verifying share password
 */
export interface VerifyPasswordRequest {
  access_code: string;
}

/**
 * Response interface for password verification
 */
export interface VerifyPasswordResponse {
  valid: boolean;
  access_token?: string;
}

/**
 * Create a share link for files
 * @param req - Share creation request
 * @returns Share creation response with share URL
 */
export async function createShare(req: CreateShareRequest): Promise<CreateShareResponse> {
  return request.post('/api/v1/shares', req);
}

/**
 * Get share information by share ID
 * @param shareId - Share ID
 * @returns Share information
 */
export async function getShareInfo(shareId: string): Promise<ShareInfo> {
  return request.get(`/api/v1/shares/${shareId}`);
}

/**
 * Verify share password
 * @param shareId - Share ID
 * @param req - Password verification request
 * @returns Verification result
 */
export async function verifySharePassword(
  shareId: string,
  req: VerifyPasswordRequest
): Promise<VerifyPasswordResponse> {
  return request.post(`/api/v1/shares/${shareId}/verify`, req);
}

/**
 * Get current user's share list
 * @returns List of user's shares
 */
export async function getMyShares(): Promise<MyShare[]> {
  return request.get('/api/v1/shares/me');
}

/**
 * Delete a share link
 * @param shareId - Share ID to delete
 */
export async function deleteShare(shareId: string): Promise<void> {
  return request.delete(`/api/v1/shares/${shareId}`);
}

/**
 * Increment download count for a share
 * @param shareId - Share ID
 */
export async function incrementDownloadCount(shareId: string): Promise<void> {
  return request.post(`/api/v1/shares/${shareId}/download`);
}

/**
 * Build a public download URL for a shared file.
 * This endpoint streams the file and does not require the user to be logged in.
 */
export function getShareDownloadUrl(shareId: string): string {
  const baseURL = resolveAPIBaseURL(import.meta.env.VITE_API_BASE_URL);
  const path = `/api/v1/shares/${encodeURIComponent(shareId)}/download`;
  if (!baseURL || baseURL === '/') return path;
  return `${baseURL.replace(/\/$/, '')}${path}`;
}

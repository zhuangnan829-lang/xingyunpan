// Multipart upload API

import request from './request';
import type {
  InitMultipartRequest,
  InitMultipartResponse,
  PresignedUrlsResponse,
  RecordChunkRequest,
  CompletedChunksResponse,
  CompleteMultipartRequest,
  UploadProgressResponse,
} from '@/types/upload';
import type { FileItem } from '@/types/file';

/**
 * Initialize multipart upload
 * @param data - Init multipart upload request data
 * @returns Promise<InitMultipartResponse>
 */
export async function initMultipartUpload(
  data: InitMultipartRequest
): Promise<InitMultipartResponse> {
  return request.post('/api/multipart/init', data);
}

/**
 * Get presigned URLs for all chunks
 * @param uploadId - Upload session ID
 * @returns Promise<PresignedUrlsResponse>
 */
export async function getPresignedUrls(
  uploadId: string
): Promise<PresignedUrlsResponse> {
  return request.get(`/api/multipart/${uploadId}/urls`);
}

/**
 * Record chunk upload completion
 * @param data - Record chunk request data
 * @returns Promise<void>
 */
export async function recordChunkComplete(
  data: RecordChunkRequest
): Promise<void> {
  return request.post('/api/multipart/chunk/complete', data);
}

/**
 * Get completed chunks for an upload session
 * @param uploadId - Upload session ID
 * @returns Promise<CompletedChunksResponse>
 */
export async function getCompletedChunks(
  uploadId: string
): Promise<CompletedChunksResponse> {
  return request.get(`/api/multipart/${uploadId}/chunks`);
}

/**
 * Complete multipart upload
 * @param data - Complete multipart upload request data
 * @returns Promise<FileItem>
 */
export async function completeMultipartUpload(
  data: CompleteMultipartRequest
): Promise<FileItem> {
  return request.post('/api/multipart/complete', data);
}

/**
 * Get upload progress
 * @param uploadId - Upload session ID
 * @returns Promise<UploadProgressResponse>
 */
export async function getUploadProgress(
  uploadId: string
): Promise<UploadProgressResponse> {
  return request.get(`/api/multipart/${uploadId}/progress`);
}

/**
 * Cancel upload task
 * @param uploadId - Upload session ID
 * @returns Promise<void>
 */
export async function cancelUpload(uploadId: string): Promise<void> {
  return request.delete(`/api/multipart/${uploadId}`);
}

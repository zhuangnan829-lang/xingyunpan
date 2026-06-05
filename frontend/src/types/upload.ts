// Upload-related TypeScript type definitions

import type { FileItem } from './file';

export type UploadStatus = 'pending' | 'hashing' | 'uploading' | 'completed' | 'failed' | 'cancelled';

export interface UploadTask {
  id: string;
  file: File;
  hash: string;
  uploadId?: string;
  status: UploadStatus;
  progress: number;
  speed: number;
  error?: string;
  isMultipart: boolean;
  totalChunks?: number;
  completedChunks?: number[];
  startTime?: number;
  folderId?: number;
  instantUpload?: boolean;
  result?: FileItem;
}

export interface ChunkInfo {
  chunkNumber: number;
  start: number;
  end: number;
  blob: Blob;
  url?: string;
  etag?: string;
  uploaded: boolean;
}

export interface InitMultipartRequest {
  filename: string;
  file_size: number;
  hash: string;
  mime_type: string;
  folder_id?: number;
}

export interface InitMultipartResponse {
  upload_id: string;
  chunk_size: number;
  total_chunks: number;
  parallel_chunk_count?: number;
  instant_upload?: boolean;
}

export interface PresignedUrlsResponse {
  urls: Array<{
    chunk_number: number;
    url: string;
  }>;
}

export interface RecordChunkRequest {
  upload_id: string;
  chunk_number: number;
  etag: string;
}

export interface CompletedChunksResponse {
  completed_chunks: number[];
}

export interface CompleteMultipartRequest {
  upload_id: string;
}

export interface UploadProgressResponse {
  upload_id: string;
  total_chunks: number;
  completed_chunks: number;
  status: 'uploading' | 'completed' | 'failed';
}

export interface UploadSession {
  uploadId: string;
  hash: string;
  filename: string;
  fileSize: number;
  chunkSize?: number;
  parallelChunkCount?: number;
  totalChunks: number;
  completedChunks: number[];
  timestamp: number;
}

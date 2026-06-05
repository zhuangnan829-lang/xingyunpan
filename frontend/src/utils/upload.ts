// Upload utility class for managing file uploads

import { calculateFileHash } from './hash';
import {
  initMultipartUpload,
  getPresignedUrls,
  recordChunkComplete,
  getCompletedChunks,
  completeMultipartUpload,
  cancelUpload as cancelUploadAPI,
} from '@/api/upload';
import type {
  UploadTask,
  ChunkInfo,
  UploadSession,
  InitMultipartResponse,
} from '@/types/upload';
import type { FileItem } from '@/types/file';
import { getFileSystemClientSettings, type FileSystemClientSettingsPayload } from '@/api/file-system-settings';

// Constants
const MULTIPART_THRESHOLD = 100 * 1024 * 1024; // 100MB
const MAX_CHUNK_RETRIES = 3;
const MAX_CONCURRENT_CHUNKS = 3;
const UPLOAD_SESSIONS_KEY = 'xingyunpan_upload_sessions';

let uploadSettingsCache: Promise<FileSystemClientSettingsPayload> | null = null;

function getUploadSettings(): Promise<FileSystemClientSettingsPayload> {
  if (!uploadSettingsCache) {
    const fallback: FileSystemClientSettingsPayload = {
      online_editor_size: 50,
      online_editor_unit: 'MB',
      max_page_size: 2000,
      max_chunk_retry: MAX_CHUNK_RETRIES,
      cache_chunks_for_retry: true,
      transfer_parallelism: MAX_CONCURRENT_CHUNKS,
      max_batch_action_size: 3000,
      map_provider: 'osm-leaflet',
      show_encryption_status: true,
      enable_event_push: true,
      debounce_delay: 5,
      file_icon_rules: '[]',
      emoji_options: '{}',
    };
    uploadSettingsCache = getFileSystemClientSettings().catch(() => fallback);
  }
  return uploadSettingsCache as Promise<FileSystemClientSettingsPayload>;
}

/**
 * Upload manager class for handling file uploads
 */
export class UploadManager {
  private task: UploadTask;
  private onProgressUpdate?: (task: UploadTask) => void;
  private onStatusChange?: (task: UploadTask) => void;
  private abortController?: AbortController;

  constructor(
    task: UploadTask,
    onProgressUpdate?: (task: UploadTask) => void,
    onStatusChange?: (task: UploadTask) => void
  ) {
    this.task = task;
    this.onProgressUpdate = onProgressUpdate;
    this.onStatusChange = onStatusChange;
  }

  /**
   * Start the upload process
   */
  async start(): Promise<FileItem> {
    try {
      this.abortController = new AbortController();
      this.task.startTime = Date.now();

      // Calculate file hash if not already calculated
      if (!this.task.hash) {
        this.updateStatus('hashing');
        this.task.hash = await calculateFileHash(this.task.file, (progress) => {
          this.task.progress = progress;
          this.notifyProgress();
        });
      }

      // Determine upload strategy
      const useMultipart = shouldUseMultipart(this.task.file.size);
      this.task.isMultipart = useMultipart;

      let result: FileItem;

      if (useMultipart) {
        // Check for resume upload
        const resumeSession = await this.detectResumeUpload();
        if (resumeSession) {
          result = await this.resumeLargeFile(resumeSession);
        } else {
          result = await this.uploadLargeFile();
        }
      } else {
        result = await this.uploadSmallFile();
      }

      this.task.result = result;
      this.updateStatus('completed');
      this.task.progress = 100;
      this.notifyProgress();

      // Clear upload session from localStorage
      this.clearUploadSession();

      return result;
    } catch (error: any) {
      if (error.name === 'AbortError' || this.task.status === 'cancelled') {
        this.updateStatus('cancelled');
        throw new Error('Upload cancelled');
      }

      this.task.error = error.message || 'Upload failed';
      this.updateStatus('failed');
      throw error;
    }
  }

  /**
   * Cancel the upload
   */
  async cancel(): Promise<void> {
    this.updateStatus('cancelled');
    
    // Abort ongoing requests
    if (this.abortController) {
      this.abortController.abort();
    }

    // Cancel multipart upload on server
    if (this.task.uploadId) {
      try {
        await cancelUploadAPI(this.task.uploadId);
      } catch (error) {
        console.error('Failed to cancel upload on server:', error);
      }
    }

    // Clear upload session
    this.clearUploadSession();
  }

  /**
   * Upload small file (< 5MB)
   */
  private async uploadSmallFile(): Promise<FileItem> {
    this.updateStatus('uploading');
    this.task.progress = 0;
    this.notifyProgress();

    // Use XMLHttpRequest for progress tracking
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('file', this.task.file);
      
      if (this.task.folderId !== undefined) {
        formData.append('parent_id', this.task.folderId.toString());
      }

      // Track upload progress
      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          this.task.progress = Math.floor((e.loaded / e.total) * 100);
          this.task.speed = this.calculateSpeed(e.loaded);
          this.notifyProgress();
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const rawResponse = JSON.parse(xhr.responseText);
            const response = rawResponse && typeof rawResponse === 'object' && 'data' in rawResponse
              ? rawResponse.data
              : rawResponse;
            const fileItem: FileItem = {
              id: response.id,
              name: response.name,
              size: response.size || 0,
              hash: response.hash || this.task.hash,
              mime_type: response.mime_type || 'application/octet-stream',
              folder_id: response.parent_id,
              created_at: response.created_at,
              updated_at: response.updated_at,
            };
            resolve(fileItem);
          } catch (error) {
            reject(new Error('Failed to parse upload response'));
          }
        } else {
          reject(new Error(`Upload failed with status ${xhr.status}`));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Network error during upload'));
      });

      xhr.addEventListener('abort', () => {
        reject(new Error('Upload cancelled'));
      });

      // Get token from localStorage
      const token = localStorage.getItem('xingyunpan_token');
      xhr.open('POST', '/api/v1/file/upload');
      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      }

      xhr.send(formData);

      // Handle abort
      if (this.abortController) {
        this.abortController.signal.addEventListener('abort', () => {
          xhr.abort();
        });
      }
    });
  }

  /**
   * Upload large file with multipart upload
   */
  private async uploadLargeFile(): Promise<FileItem> {
    this.updateStatus('uploading');

    // Initialize multipart upload
    const initResponse = await initMultipartUpload({
      filename: this.task.file.name,
      file_size: this.task.file.size,
      hash: this.task.hash,
      mime_type: this.task.file.type || 'application/octet-stream',
      folder_id: this.task.folderId,
    });

    // Check for instant upload
    if (initResponse.instant_upload) {
      this.task.progress = 100;
      this.task.instantUpload = true;
      this.notifyProgress();
      // Return a mock FileItem (actual file already exists on server)
      return {
        id: 0, // Will be updated by server response
        name: this.task.file.name,
        size: this.task.file.size,
        hash: this.task.hash,
        mime_type: this.task.file.type || 'application/octet-stream',
        folder_id: this.task.folderId || null,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
    }

    this.task.uploadId = initResponse.upload_id;
    this.task.totalChunks = initResponse.total_chunks;
    this.task.completedChunks = [];

    // Save upload session for resume
    const settings = await getUploadSettings();
    if (settings.cache_chunks_for_retry) {
      this.saveUploadSession(initResponse);
    }

    // Get presigned URLs
    const urlsResponse = await getPresignedUrls(initResponse.upload_id);

    // Create chunk info array
    const chunks = this.createChunks(initResponse.chunk_size, urlsResponse.urls);

    // Upload chunks with concurrency control
    await this.uploadChunksWithConcurrency(chunks, initResponse.upload_id, initResponse.parallel_chunk_count);

    // Complete multipart upload
    const result = await completeMultipartUpload({
      upload_id: initResponse.upload_id,
    });

    return result;
  }

  /**
   * Resume large file upload
   */
  private async resumeLargeFile(session: UploadSession): Promise<FileItem> {
    this.updateStatus('uploading');
    this.task.uploadId = session.uploadId;
    this.task.totalChunks = session.totalChunks;
    this.task.completedChunks = session.completedChunks;

    // Get presigned URLs for remaining chunks
    const urlsResponse = await getPresignedUrls(session.uploadId);

    const chunkSize = session.chunkSize || Math.ceil(this.task.file.size / session.totalChunks);

    // Create chunk info array
    const chunks = this.createChunks(chunkSize, urlsResponse.urls);

    // Filter out already completed chunks
    const remainingChunks = chunks.filter(
      (chunk) => !session.completedChunks.includes(chunk.chunkNumber)
    );

    // Upload remaining chunks
    await this.uploadChunksWithConcurrency(remainingChunks, session.uploadId, session.parallelChunkCount);

    // Complete multipart upload
    const result = await completeMultipartUpload({
      upload_id: session.uploadId,
    });

    return result;
  }

  /**
   * Create chunk info array
   */
  private createChunks(
    chunkSize: number,
    urls: Array<{ chunk_number: number; url: string }>
  ): ChunkInfo[] {
    const chunks: ChunkInfo[] = [];
    const totalChunks = Math.ceil(this.task.file.size / chunkSize);

    for (let i = 0; i < totalChunks; i++) {
      const start = i * chunkSize;
      const end = Math.min(start + chunkSize, this.task.file.size);
      const blob = this.task.file.slice(start, end);
      const urlInfo = urls.find((u) => u.chunk_number === i + 1);

      chunks.push({
        chunkNumber: i + 1,
        start,
        end,
        blob,
        url: urlInfo?.url,
        uploaded: false,
      });
    }

    return chunks;
  }

  /**
   * Upload chunks with concurrency control
   */
  private async uploadChunksWithConcurrency(
    chunks: ChunkInfo[],
    uploadId: string,
    preferredConcurrency?: number
  ): Promise<void> {
    const settings = await getUploadSettings();
    const concurrencySource = Number(preferredConcurrency) || Number(settings.transfer_parallelism) || MAX_CONCURRENT_CHUNKS;
    const concurrency = Math.max(1, Math.min(64, concurrencySource));
    const queue = [...chunks];
    const inProgress: Promise<void>[] = [];

    while (queue.length > 0 || inProgress.length > 0) {
      // Check if cancelled
      if (this.task.status === 'cancelled') {
        throw new Error('Upload cancelled');
      }

      // Start new uploads up to max concurrency
      while (inProgress.length < concurrency && queue.length > 0) {
        const chunk = queue.shift()!;
        const promise = this.uploadChunkWithRetry(chunk, uploadId, settings)
          .then(() => {
            // Remove from in-progress
            const index = inProgress.indexOf(promise);
            if (index > -1) {
              inProgress.splice(index, 1);
            }
          })
          .catch((error) => {
            // Remove from in-progress and re-throw
            const index = inProgress.indexOf(promise);
            if (index > -1) {
              inProgress.splice(index, 1);
            }
            throw error;
          });

        inProgress.push(promise);
      }

      // Wait for at least one to complete
      if (inProgress.length > 0) {
        await Promise.race(inProgress);
      }
    }
  }

  /**
   * Upload a single chunk with retry logic
   */
  private async uploadChunkWithRetry(
    chunk: ChunkInfo,
    uploadId: string,
    settings?: FileSystemClientSettingsPayload
  ): Promise<void> {
    let lastError: Error | null = null;
    const maxRetries = Math.max(1, Math.min(50, Number(settings?.max_chunk_retry) || MAX_CHUNK_RETRIES));

    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        // Check if cancelled
        if (this.task.status === 'cancelled') {
          throw new Error('Upload cancelled');
        }

        // Upload chunk to storage
        const etag = await this.uploadChunkToStorage(chunk);

        // Record chunk completion
        await recordChunkComplete({
          upload_id: uploadId,
          chunk_number: chunk.chunkNumber,
          etag,
        });

        // Mark as uploaded
        chunk.uploaded = true;
        chunk.etag = etag;

        // Update completed chunks
        if (!this.task.completedChunks) {
          this.task.completedChunks = [];
        }
        this.task.completedChunks.push(chunk.chunkNumber);

        // Update progress
        this.updateProgress();

        // Update upload session
        if (settings?.cache_chunks_for_retry !== false) {
          this.updateUploadSession();
        }

        return;
      } catch (error: any) {
        lastError = error;

        if (attempt < maxRetries - 1) {
          // Exponential backoff
          const delay = Math.pow(2, attempt) * 1000;
          await new Promise((resolve) => setTimeout(resolve, delay));
        }
      }
    }

    throw lastError || new Error(`Failed to upload chunk ${chunk.chunkNumber}`);
  }

  /**
   * Upload chunk to storage using presigned URL
   */
  private async uploadChunkToStorage(chunk: ChunkInfo): Promise<string> {
    if (!chunk.url) {
      throw new Error('Chunk URL not available');
    }

    const url = chunk.url;

    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          // Get ETag from response headers
          const etag = xhr.getResponseHeader('ETag') || '';
          resolve(etag.replace(/"/g, '')); // Remove quotes from ETag
        } else {
          reject(new Error(`Chunk upload failed with status ${xhr.status}`));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Network error during chunk upload'));
      });

      xhr.addEventListener('abort', () => {
        reject(new Error('Chunk upload cancelled'));
      });

      xhr.open('PUT', url);
      xhr.setRequestHeader('Content-Type', 'application/octet-stream');
      xhr.send(chunk.blob);

      // Handle abort
      if (this.abortController) {
        this.abortController.signal.addEventListener('abort', () => {
          xhr.abort();
        });
      }
    });
  }

  /**
   * Detect if there's an existing upload session for resume
   */
  private async detectResumeUpload(): Promise<UploadSession | null> {
    try {
      const sessionsJson = localStorage.getItem(UPLOAD_SESSIONS_KEY);
      if (!sessionsJson) return null;

      const sessions: Record<string, UploadSession> = JSON.parse(sessionsJson);
      const session = sessions[this.task.hash];

      if (!session) return null;

      // Verify session is still valid on server
      const completedChunks = await getCompletedChunks(session.uploadId);

      return {
        ...session,
        completedChunks: completedChunks.completed_chunks,
      };
    } catch (error) {
      console.error('Failed to detect resume upload:', error);
      // Clear invalid session
      this.clearUploadSession();
      return null;
    }
  }

  /**
   * Save upload session to localStorage
   */
  private saveUploadSession(initResponse: InitMultipartResponse): void {
    try {
      const sessionsJson = localStorage.getItem(UPLOAD_SESSIONS_KEY) || '{}';
      const sessions: Record<string, UploadSession> = JSON.parse(sessionsJson);

      sessions[this.task.hash] = {
        uploadId: initResponse.upload_id,
        hash: this.task.hash,
        filename: this.task.file.name,
        fileSize: this.task.file.size,
        chunkSize: initResponse.chunk_size,
        parallelChunkCount: initResponse.parallel_chunk_count,
        totalChunks: initResponse.total_chunks,
        completedChunks: this.task.completedChunks || [],
        timestamp: Date.now(),
      };

      localStorage.setItem(UPLOAD_SESSIONS_KEY, JSON.stringify(sessions));
    } catch (error) {
      console.error('Failed to save upload session:', error);
    }
  }

  /**
   * Update upload session in localStorage
   */
  private updateUploadSession(): void {
    try {
      const sessionsJson = localStorage.getItem(UPLOAD_SESSIONS_KEY);
      if (!sessionsJson) return;

      const sessions: Record<string, UploadSession> = JSON.parse(sessionsJson);
      const session = sessions[this.task.hash];

      if (session) {
        session.completedChunks = this.task.completedChunks || [];
        localStorage.setItem(UPLOAD_SESSIONS_KEY, JSON.stringify(sessions));
      }
    } catch (error) {
      console.error('Failed to update upload session:', error);
    }
  }

  /**
   * Clear upload session from localStorage
   */
  private clearUploadSession(): void {
    try {
      const sessionsJson = localStorage.getItem(UPLOAD_SESSIONS_KEY);
      if (!sessionsJson) return;

      const sessions: Record<string, UploadSession> = JSON.parse(sessionsJson);
      delete sessions[this.task.hash];

      localStorage.setItem(UPLOAD_SESSIONS_KEY, JSON.stringify(sessions));
    } catch (error) {
      console.error('Failed to clear upload session:', error);
    }
  }

  /**
   * Update upload progress
   */
  private updateProgress(): void {
    if (this.task.totalChunks && this.task.completedChunks) {
      this.task.progress = Math.floor(
        (this.task.completedChunks.length / this.task.totalChunks) * 100
      );
      this.task.speed = this.calculateSpeed(
        (this.task.completedChunks.length / this.task.totalChunks) * this.task.file.size
      );
      this.notifyProgress();
    }
  }

  /**
   * Calculate upload speed
   */
  private calculateSpeed(uploadedBytes: number): number {
    if (!this.task.startTime) return 0;

    const elapsedSeconds = (Date.now() - this.task.startTime) / 1000;
    if (elapsedSeconds === 0) return 0;

    return uploadedBytes / elapsedSeconds;
  }

  /**
   * Update task status
   */
  private updateStatus(status: UploadTask['status']): void {
    this.task.status = status;
    if (this.onStatusChange) {
      this.onStatusChange(this.task);
    }
  }

  /**
   * Notify progress update
   */
  private notifyProgress(): void {
    if (this.onProgressUpdate) {
      this.onProgressUpdate(this.task);
    }
  }
}

/**
 * Determine if file should use multipart upload
 * @param fileSize - File size in bytes
 * @returns true if file should use multipart upload
 */
export function shouldUseMultipart(fileSize: number): boolean {
  return fileSize >= MULTIPART_THRESHOLD;
}


// File-related TypeScript type definitions

/**
 * File item in the file system
 */
export interface FileItem {
  id: number;
  name: string;
  size: number;
  hash: string;
  mime_type: string;
  content_type?: string;
  physical_file_id?: number;
  thumbnail_url?: string;
  folder_id: number | null;
  created_at: string;
  updated_at: string;
  browser_app?: {
    group_id?: number;
    group_name?: string;
    id?: number;
    icon?: string;
    icon_url?: string;
    accent?: string;
    type?: string;
    name?: string;
    platform?: string;
    create_mapping?: string;
    open_in_new_window?: boolean;
    max_size?: number;
    max_size_unit?: string;
    matched_extension?: string;
    source?: string;
  } | null;
  display_icon?: string;
  display_icon_tint?: string;
  display_icon_label?: string;
  display_icon_source?: string;
  encryption_status?: EncryptionStatus | null;
}

export interface EncryptionStatus {
  visible: boolean;
  encrypted: boolean;
  storage_policy_id: number;
  storage_policy_name: string;
  key_id?: string;
}

export interface DirectoryStats {
  child_count: number;
  file_count: number;
  folder_count: number;
  total_size: number;
  cached?: boolean;
  ttl_seconds?: number;
  cache_enabled?: boolean;
  cache_configured?: boolean;
}

/**
 * Folder item reference (to avoid circular dependency)
 */
export interface FolderItem {
  id: number;
  name: string;
  parent_id: number | null;
  created_at: string;
  updated_at: string;
  directory_stats?: DirectoryStats | null;
  display_icon?: string;
  display_icon_tint?: string;
  display_icon_label?: string;
  display_icon_source?: string;
}

/**
 * File list response containing files and folders
 */
export interface FileListResponse {
  files: FileItem[];
  folders: FolderItem[];
}

/**
 * Upload file request (for small files)
 */
export interface UploadFileRequest {
  file: File;
  hash: string;
  folder_id?: number;
}
/**
 * File type for filtering
 */
export type FileItemType = 'image' | 'video' | 'document' | 'archive' | 'other' | 'folder';

/**
 * File size range for filtering
 * - small: < 1MB
 * - medium: 1MB - 10MB
 * - large: 10MB - 100MB
 * - xlarge: > 100MB
 */
export type SizeRange = 'small' | 'medium' | 'large' | 'xlarge';

/**
 * Time range for filtering
 * - today: Today (00:00:00 to current time)
 * - week: Last 7 days
 * - month: Last 30 days
 * - custom: User-specified start and end time
 */
export type TimeRange = 'today' | 'week' | 'month' | 'custom';

/**
 * File filters for search and filtering
 */
export interface FileFilters {
  type: FileItemType | 'all';    // File type filter
  sizeRange: SizeRange | 'all';  // File size filter
  timeRange: TimeRange | 'all';  // Time range filter
  customRange?: {                // Custom time range (when timeRange is 'custom')
    start: Date;
    end: Date;
  };
}

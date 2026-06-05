// Folder-related TypeScript type definitions

/**
 * Folder item in the file system
 */
export interface FolderItem {
  id: number;
  name: string;
  parent_id: number | null;
  created_at: string;
  updated_at: string;
  display_icon?: string;
  display_icon_tint?: string;
  display_icon_label?: string;
  display_icon_source?: string;
  directory_stats?: {
    child_count: number;
    file_count: number;
    folder_count: number;
    total_size: number;
    cached?: boolean;
    ttl_seconds?: number;
    cache_enabled?: boolean;
    cache_configured?: boolean;
  } | null;
}

/**
 * Create folder request
 */
export interface CreateFolderRequest {
  name: string;
  parent_id?: number;
}

/**
 * Folder path item for breadcrumb navigation
 */
export interface FolderPathItem {
  id: number;
  name: string;
}

/**
 * Rename folder request
 */
export interface RenameFolderRequest {
  folder_id: number;
  new_name: string;
}

/**
 * Move folder request
 */
export interface MoveFolderRequest {
  folder_id: number;
  target_folder_id: number;
}

/**
 * Copy folder request
 */
export interface CopyFolderRequest {
  folder_id: number;
  target_folder_id: number;
}

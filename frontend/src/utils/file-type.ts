/**
 * 文件类型识别工具
 * 提供文件类型图标映射和识别功能
 */

/**
 * 文件类型到 Element Plus 图标的映射
 * Requirements: 25.1-25.7
 */
const FILE_TYPE_ICONS: Record<string, string> = {
  // Images
  'image/jpeg': 'Picture',
  'image/png': 'Picture',
  'image/gif': 'Picture',
  'image/webp': 'Picture',
  'image/bmp': 'Picture',
  'image/svg+xml': 'Picture',
  
  // Videos
  'video/mp4': 'VideoCamera',
  'video/avi': 'VideoCamera',
  'video/mkv': 'VideoCamera',
  'video/mov': 'VideoCamera',
  'video/wmv': 'VideoCamera',
  'video/flv': 'VideoCamera',
  
  // Audio
  'audio/mpeg': 'Headset',
  'audio/mp3': 'Headset',
  'audio/wav': 'Headset',
  'audio/ogg': 'Headset',
  'audio/flac': 'Headset',
  
  // Documents
  'application/pdf': 'Document',
  'application/msword': 'Document',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'Document',
  'application/vnd.ms-excel': 'Document',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': 'Document',
  'application/vnd.ms-powerpoint': 'Document',
  'application/vnd.openxmlformats-officedocument.presentationml.presentation': 'Document',
  'text/plain': 'Document',
  'text/html': 'Document',
  'text/css': 'Document',
  'text/javascript': 'Document',
  'application/json': 'Document',
  'application/xml': 'Document',
  
  // Archives
  'application/zip': 'FolderOpened',
  'application/x-rar-compressed': 'FolderOpened',
  'application/x-7z-compressed': 'FolderOpened',
  'application/x-tar': 'FolderOpened',
  'application/gzip': 'FolderOpened',
};

/**
 * 文件扩展名到 MIME 类型的映射
 */
const EXTENSION_TO_MIME: Record<string, string> = {
  // Images
  'jpg': 'image/jpeg',
  'jpeg': 'image/jpeg',
  'png': 'image/png',
  'gif': 'image/gif',
  'webp': 'image/webp',
  'bmp': 'image/bmp',
  'svg': 'image/svg+xml',
  
  // Videos
  'mp4': 'video/mp4',
  'avi': 'video/avi',
  'mkv': 'video/mkv',
  'mov': 'video/mov',
  'wmv': 'video/wmv',
  'flv': 'video/flv',
  
  // Audio
  'mp3': 'audio/mpeg',
  'wav': 'audio/wav',
  'ogg': 'audio/ogg',
  'flac': 'audio/flac',
  
  // Documents
  'pdf': 'application/pdf',
  'doc': 'application/msword',
  'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'xls': 'application/vnd.ms-excel',
  'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'ppt': 'application/vnd.ms-powerpoint',
  'pptx': 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  'txt': 'text/plain',
  'html': 'text/html',
  'htm': 'text/html',
  'css': 'text/css',
  'js': 'text/javascript',
  'json': 'application/json',
  'xml': 'application/xml',
  
  // Archives
  'zip': 'application/zip',
  'rar': 'application/x-rar-compressed',
  '7z': 'application/x-7z-compressed',
  'tar': 'application/x-tar',
  'gz': 'application/gzip',
};

/**
 * 根据 MIME 类型获取文件图标
 * @param mimeType MIME 类型
 * @returns Element Plus 图标名称
 * 
 * Requirements: 25.1-25.7
 * - 图片类型返回图片图标
 * - 视频类型返回视频图标
 * - 音频类型返回音频图标
 * - 文档类型返回文档图标
 * - 压缩包类型返回压缩包图标
 * - 未知类型返回通用文件图标
 */
export function getFileTypeIcon(mimeType: string): string {
  return FILE_TYPE_ICONS[mimeType] || 'Document';
}

/**
 * 根据文件扩展名获取 MIME 类型
 * @param filename 文件名
 * @returns MIME 类型
 * 
 * Requirements: 25.1-25.7
 */
export function getFileTypeByExtension(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase();
  
  if (!ext) {
    return 'application/octet-stream';
  }
  
  return EXTENSION_TO_MIME[ext] || 'application/octet-stream';
}

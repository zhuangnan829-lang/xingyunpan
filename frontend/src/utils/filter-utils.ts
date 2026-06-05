/**
 * 文件筛选工具函数
 * 提供文件搜索、类型筛选、大小筛选、时间筛选等功能
 * 
 * Requirements: 4.3, 5.1, 5.2, 5.3, 5.4, 5.7
 */

import type { FileItem, FileItemType, SizeRange, TimeRange, FileFilters } from '@/types/file';

/**
 * 按关键词过滤文件（不区分大小写）
 * @param files 文件列表
 * @param keyword 搜索关键词
 * @returns 过滤后的文件列表
 * 
 * Requirements: 4.3
 * - 不区分大小写
 * - 匹配文件名中包含关键词的文件
 * - 空关键词返回所有文件
 */
export function filterByKeyword(files: FileItem[], keyword: string): FileItem[] {
  if (!keyword || keyword.trim() === '') {
    return files;
  }
  
  const lowerKeyword = keyword.toLowerCase().trim();
  return files.filter(file => 
    file.name.toLowerCase().includes(lowerKeyword)
  );
}

/**
 * 根据 MIME 类型判断文件类型
 * @param mimeType MIME 类型
 * @returns 文件类型
 */
function getFileTypeFromMime(mimeType: string): FileItemType {
  if (mimeType.startsWith('image/')) {
    return 'image';
  }
  if (mimeType.startsWith('video/')) {
    return 'video';
  }
  if (mimeType.startsWith('audio/')) {
    return 'document'; // 音频归类为文档
  }
  if (
    mimeType.includes('pdf') ||
    mimeType.includes('word') ||
    mimeType.includes('excel') ||
    mimeType.includes('powerpoint') ||
    mimeType.includes('document') ||
    mimeType.includes('spreadsheet') ||
    mimeType.includes('presentation') ||
    mimeType.startsWith('text/')
  ) {
    return 'document';
  }
  if (
    mimeType.includes('zip') ||
    mimeType.includes('rar') ||
    mimeType.includes('7z') ||
    mimeType.includes('tar') ||
    mimeType.includes('gzip') ||
    mimeType.includes('compressed')
  ) {
    return 'archive';
  }
  return 'other';
}

/**
 * 按文件类型过滤
 * @param files 文件列表
 * @param type 文件类型
 * @returns 过滤后的文件列表
 * 
 * Requirements: 5.1
 * - 支持按图片、视频、文档、压缩包、其他类型筛选
 * - 'all' 返回所有文件
 */
export function filterByType(files: FileItem[], type: FileItemType | 'all'): FileItem[] {
  if (type === 'all') {
    return files;
  }
  
  return files.filter(file => {
    const fileType = getFileTypeFromMime(file.mime_type);
    return fileType === type;
  });
}

/**
 * 判断文件大小范围
 * @param fileSize 文件大小（字节）
 * @returns 文件大小范围
 * 
 * Requirements: 5.2
 * - small: < 1MB
 * - medium: 1MB - 10MB
 * - large: 10MB - 100MB
 * - xlarge: > 100MB
 */
export function getSizeRange(fileSize: number): SizeRange {
  const MB = 1024 * 1024;
  
  if (fileSize < MB) {
    return 'small';
  }
  if (fileSize < 10 * MB) {
    return 'medium';
  }
  if (fileSize < 100 * MB) {
    return 'large';
  }
  return 'xlarge';
}

/**
 * 按文件大小过滤
 * @param files 文件列表
 * @param sizeRange 文件大小范围
 * @returns 过滤后的文件列表
 * 
 * Requirements: 5.2
 */
export function filterBySize(files: FileItem[], sizeRange: SizeRange | 'all'): FileItem[] {
  if (sizeRange === 'all') {
    return files;
  }
  
  return files.filter(file => getSizeRange(file.size) === sizeRange);
}

/**
 * 判断文件是否在时间范围内
 * @param fileDate 文件日期（ISO 8601 格式）
 * @param timeRange 时间范围
 * @param customRange 自定义时间范围（当 timeRange 为 'custom' 时使用）
 * @returns 是否在时间范围内
 * 
 * Requirements: 5.3
 * - today: 今天 00:00:00 之后
 * - week: 最近 7 天
 * - month: 最近 30 天
 * - custom: 自定义时间范围
 */
export function isInTimeRange(
  fileDate: string,
  timeRange: TimeRange | 'all',
  customRange?: { start: Date; end: Date }
): boolean {
  if (timeRange === 'all') {
    return true;
  }
  
  const fileTime = new Date(fileDate).getTime();
  const now = new Date();
  
  if (timeRange === 'today') {
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
    return fileTime >= todayStart;
  }
  
  if (timeRange === 'week') {
    const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).getTime();
    return fileTime >= weekAgo;
  }
  
  if (timeRange === 'month') {
    const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000).getTime();
    return fileTime >= monthAgo;
  }
  
  if (timeRange === 'custom' && customRange) {
    const startTime = customRange.start.getTime();
    const endTime = customRange.end.getTime();
    return fileTime >= startTime && fileTime <= endTime;
  }
  
  return true;
}

/**
 * 按时间范围过滤
 * @param files 文件列表
 * @param timeRange 时间范围
 * @param customRange 自定义时间范围（当 timeRange 为 'custom' 时使用）
 * @returns 过滤后的文件列表
 * 
 * Requirements: 5.3
 */
export function filterByTime(
  files: FileItem[],
  timeRange: TimeRange | 'all',
  customRange?: { start: Date; end: Date }
): FileItem[] {
  if (timeRange === 'all') {
    return files;
  }
  
  return files.filter(file => 
    isInTimeRange(file.created_at, timeRange, customRange)
  );
}

/**
 * 组合过滤（应用所有筛选条件，交集逻辑）
 * @param files 文件列表
 * @param keyword 搜索关键词
 * @param filters 筛选条件
 * @returns 过滤后的文件列表
 * 
 * Requirements: 5.4, 5.7
 * - 应用搜索关键词
 * - 应用文件类型筛选
 * - 应用文件大小筛选
 * - 应用时间范围筛选
 * - 所有条件使用交集逻辑（AND）
 */
export function applyFilters(
  files: FileItem[],
  keyword: string,
  filters: FileFilters
): FileItem[] {
  let result = files;
  
  // 应用关键词搜索
  result = filterByKeyword(result, keyword);
  
  // 应用文件类型筛选
  result = filterByType(result, filters.type);
  
  // 应用文件大小筛选
  result = filterBySize(result, filters.sizeRange);
  
  // 应用时间范围筛选
  result = filterByTime(result, filters.timeRange, filters.customRange);
  
  return result;
}

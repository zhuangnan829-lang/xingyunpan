import request from './request';
import { searchFiles } from './search';
import type { FileItem } from '@/types/file';

export interface VideoListRequest {
  keyword?: string;
  sort?: 'recent' | 'name' | 'size';
  page?: number;
  page_size?: number;
}

export interface VideoListResponse {
  files: FileItem[];
  total: number;
  total_size: number;
  page: number;
  page_size: number;
  total_pages: number;
}

const videoExtensions = [
  'mp4',
  'mov',
  'mkv',
  'webm',
  'avi',
  'wmv',
  'flv',
  'm4v',
  'mpeg',
  'mpg',
  '3gp',
  '3g2',
  'ts',
  'mts',
  'm2ts',
  'rm',
  'rmvb',
  'vob',
  'ogv',
  'asf',
  'divx',
];

export async function listVideos(params: VideoListRequest = {}): Promise<VideoListResponse> {
  try {
    const response = await request.get<any>('/api/v1/videos', { params });
    const normalized = normalizeVideoListResponse(response, params);
    if (normalized.files.length > 0) {
      return normalized;
    }
    return listVideosWithSearchFallback(params);
  } catch {
    return listVideosWithSearchFallback(params);
  }
}

function normalizeVideoListResponse(response: any, params: VideoListRequest): VideoListResponse {
  return {
    files: (response.files || []).map((item: any) => ({
      id: Number(item.id),
      name: item.name || '',
      size: item.size || 0,
      hash: item.hash || '',
      mime_type: item.mime_type || 'video/mp4',
      content_type: item.content_type || item.mime_type || 'video/mp4',
      physical_file_id: item.physical_file_id,
      thumbnail_url: item.thumbnail_url,
      folder_id: item.folder_id ?? null,
      created_at: item.created_at,
      updated_at: item.updated_at,
      browser_app: null,
    })),
    total: response.total || 0,
    total_size: response.total_size || 0,
    page: response.page || params.page || 1,
    page_size: response.page_size || params.page_size || 50,
    total_pages: response.total_pages || 0,
  };
}

async function listVideosWithSearchFallback(params: VideoListRequest): Promise<VideoListResponse> {
  const typedResponse = await searchFiles({
    keyword: params.keyword || '',
    file_type: 'video',
    page: params.page || 1,
    page_size: params.page_size || 50,
  });
  const response = typedResponse.files.length > 0
    ? typedResponse
    : await searchFiles({
        keyword: params.keyword || '',
        page: params.page || 1,
        page_size: 2000,
      });

  const files = response.files.filter(isVideoFile).sort((a, b) => {
    if (params.sort === 'name') return a.name.localeCompare(b.name, 'zh-CN');
    if (params.sort === 'size') return (b.size || 0) - (a.size || 0);
    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
  });

  return {
    files,
    total: response.total || files.length,
    total_size: files.reduce((sum, file) => sum + (file.size || 0), 0),
    page: response.page || params.page || 1,
    page_size: response.page_size || params.page_size || 50,
    total_pages: Math.ceil((response.total || files.length) / (response.page_size || params.page_size || 50)),
  };
}

function isVideoFile(file: FileItem) {
  const mime = (file.content_type || file.mime_type || '').toLowerCase();
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return mime === 'video' || mime.startsWith('video/') || videoExtensions.includes(extension);
}

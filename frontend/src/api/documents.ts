import request from './request';
import { searchFiles } from './search';
import type { FileItem } from '@/types/file';

export interface DocumentListRequest {
  keyword?: string;
  sort?: 'recent' | 'name' | 'size';
  page?: number;
  page_size?: number;
}

export interface DocumentListResponse {
  files: FileItem[];
  total: number;
  total_size: number;
  page: number;
  page_size: number;
  total_pages: number;
}

const documentExtensions = [
  'pdf',
  'doc',
  'docx',
  'xls',
  'xlsx',
  'ppt',
  'pptx',
  'txt',
  'md',
  'markdown',
  'csv',
  'json',
  'xml',
  'yaml',
  'yml',
  'rtf',
  'odt',
  'ods',
  'odp',
  'epub',
  'html',
  'htm',
];

export async function listDocuments(params: DocumentListRequest = {}): Promise<DocumentListResponse> {
  try {
    const response = await request.get<any>('/api/v1/documents', { params });
    const normalized = normalizeDocumentListResponse(response, params);
    if (normalized.files.length > 0) {
      return normalized;
    }
    return listDocumentsWithSearchFallback(params);
  } catch {
    return listDocumentsWithSearchFallback(params);
  }
}

function normalizeDocumentListResponse(response: any, params: DocumentListRequest): DocumentListResponse {
  return {
    files: (response.files || []).map((item: any) => ({
      id: Number(item.id),
      name: item.name || '',
      size: item.size || 0,
      hash: item.hash || '',
      mime_type: item.mime_type || 'application/octet-stream',
      content_type: item.content_type || item.mime_type || 'application/octet-stream',
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

async function listDocumentsWithSearchFallback(params: DocumentListRequest): Promise<DocumentListResponse> {
  const typedResponse = await searchFiles({
    keyword: params.keyword || '',
    file_type: 'document',
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

  const files = response.files.filter(isDocumentFile).sort((a, b) => {
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

function isDocumentFile(file: FileItem) {
  const mime = (file.content_type || file.mime_type || '').toLowerCase();
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return (
    mime === 'document' ||
    mime.includes('pdf') ||
    mime.includes('word') ||
    mime.includes('excel') ||
    mime.includes('spreadsheet') ||
    mime.includes('powerpoint') ||
    mime.includes('presentation') ||
    mime.startsWith('text/') ||
    ['application/json', 'application/xml'].includes(mime) ||
    documentExtensions.includes(extension)
  );
}

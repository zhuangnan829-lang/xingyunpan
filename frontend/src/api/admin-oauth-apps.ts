import request from './request';
import type { OAuthApp } from '@/views/admin/oauth-apps/types';

interface PageResponse<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

interface OAuthAppDTO extends Omit<OAuthApp, 'id'> {
  id: number;
  slug: string;
}

export interface OAuthAppListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  status?: 'all' | 'enabled' | 'disabled';
}

const toOAuthApp = (item: OAuthAppDTO): OAuthApp => ({
  ...item,
  id: item.slug || String(item.id),
});

const toPayload = (app: OAuthApp) => ({
  ...app,
  slug: app.id,
});

export async function listOAuthApps(params: OAuthAppListParams = {}): Promise<PageResponse<OAuthApp>> {
  const data = await request.get<PageResponse<OAuthAppDTO>>('/api/v1/admin/oauth-apps', { params });
  return {
    ...data,
    list: data.list.map(toOAuthApp),
  };
}

export async function getOAuthApp(id: string): Promise<OAuthApp> {
  return toOAuthApp(await request.get<OAuthAppDTO>(`/api/v1/admin/oauth-apps/${id}`));
}

export async function createOAuthApp(app: OAuthApp): Promise<OAuthApp> {
  return toOAuthApp(await request.post<OAuthAppDTO>('/api/v1/admin/oauth-apps', toPayload(app)));
}

export async function updateOAuthApp(app: OAuthApp): Promise<OAuthApp> {
  return toOAuthApp(await request.put<OAuthAppDTO>(`/api/v1/admin/oauth-apps/${app.id}`, toPayload(app)));
}

export async function updateOAuthAppStatus(id: string, enabled: boolean): Promise<OAuthApp> {
  return toOAuthApp(await request.put<OAuthAppDTO>(`/api/v1/admin/oauth-apps/${id}/status`, { enabled }));
}

export async function deleteOAuthApp(id: string): Promise<{ deleted: boolean }> {
  return request.delete<{ deleted: boolean }>(`/api/v1/admin/oauth-apps/${id}`);
}

export async function regenerateOAuthAppSecret(id: string): Promise<OAuthApp> {
  return toOAuthApp(await request.post<OAuthAppDTO>(`/api/v1/admin/oauth-apps/${id}/secret`));
}

import request from './request';

export type DavAccountPermission = 'read' | 'write';
export type DavAccountStatus = 'active' | 'disabled';

export interface DavAccount {
  id: number;
  account_token: string;
  name: string;
  root_path: string;
  permission: DavAccountPermission;
  reverse_proxy: boolean;
  status: DavAccountStatus;
  endpoint: string;
  created_at: string;
  updated_at: string;
  last_used_at?: string | null;
  last_used_ip?: string;
  description?: string;
}

export interface DavAccountPayload {
  name?: string;
  root_path?: string;
  permission?: DavAccountPermission;
  reverse_proxy?: boolean;
  status?: DavAccountStatus;
  description?: string;
}

export interface DavAccountWithSecret {
  account: DavAccount;
  secret: string;
}

export function listDavAccounts(): Promise<DavAccount[]> {
  return request.get('/api/v1/dav/accounts');
}

export function createDavAccount(payload: DavAccountPayload): Promise<DavAccountWithSecret> {
  return request.post('/api/v1/dav/accounts', payload);
}

export function updateDavAccount(id: number, payload: DavAccountPayload): Promise<DavAccount> {
  return request.put(`/api/v1/dav/accounts/${id}`, payload);
}

export function deleteDavAccount(id: number): Promise<{ deleted: boolean }> {
  return request.delete(`/api/v1/dav/accounts/${id}`);
}

export function resetDavAccountSecret(id: number): Promise<DavAccountWithSecret> {
  return request.post(`/api/v1/dav/accounts/${id}/secret`);
}

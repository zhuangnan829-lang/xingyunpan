import request from './request';

export type NodeType = 'master' | 'worker';
export type NodeDownloaderType = 'Aria2' | 'qBittorrent' | 'Transmission';

export interface NodeFeaturesPayload {
  create_archive: boolean;
  extract_archive: boolean;
  offline_download: boolean;
}

export interface NodeOfflinePayload {
  downloader: NodeDownloaderType;
  rpc_url: string;
  rpc_secret: string;
  task_options: string;
  temp_dir: string;
  refresh_interval: number;
  wait_for_seeding: boolean;
}

export interface NodePayload {
  id?: number;
  name: string;
  type: NodeType;
  enabled: boolean;
  weight: number;
  is_built_in: boolean;
  health: {
    status: string;
    message: string;
    last_heartbeat_at: string | null;
    last_checked_at: string | null;
  };
  features: NodeFeaturesPayload;
  offline: NodeOfflinePayload;
}

export interface NodeOfflineConnectivityResult {
  success: boolean;
  message: string;
  downloader: NodeDownloaderType;
  rpc_url: string;
  version: string;
  tested_at: string;
}

export function listNodes(): Promise<NodePayload[]> {
  return request.get<NodePayload[]>('/api/v1/admin/nodes');
}

export function getNode(id: number): Promise<NodePayload> {
  return request.get<NodePayload>(`/api/v1/admin/nodes/${id}`);
}

export function createNode(data: NodePayload): Promise<NodePayload> {
  return request.post<NodePayload>('/api/v1/admin/nodes', data);
}

export function updateNode(id: number, data: NodePayload): Promise<NodePayload> {
  return request.put<NodePayload>(`/api/v1/admin/nodes/${id}`, data);
}

export function deleteNode(id: number): Promise<{ deleted: boolean }> {
  return request.delete<{ deleted: boolean }>(`/api/v1/admin/nodes/${id}`);
}

export function testNodeOfflineConnectivity(data: NodeOfflinePayload): Promise<NodeOfflineConnectivityResult> {
  return request.post<NodeOfflineConnectivityResult>('/api/v1/admin/nodes/test-offline-connectivity', data);
}

export function checkNodeHealth(id: number): Promise<NodePayload> {
  return request.post<NodePayload>(`/api/v1/admin/nodes/${id}/check-health`);
}

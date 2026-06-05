import request from './request';

export interface DashboardMetricPayload {
  today_upload_count: number;
  today_upload_bytes: number;
  upload_change_pct: number;
  average_latency_ms: number;
  latency_change_pct: number;
  active_users: number;
  active_users_change_pct: number;
  blob_count: number;
  blob_count_change_pct: number;
  file_count: number;
  folder_count: number;
}

export interface DashboardStorageBreakdownPayload {
  name: string;
  size_bytes: number;
  tone: string;
}

export interface DashboardStoragePayload {
  used_bytes: number;
  total_bytes: number;
  percent: number;
  breakdown: DashboardStorageBreakdownPayload[];
}

export interface DashboardTrafficPayload {
  labels: string[];
  inbound: number[];
  outbound: number[];
  peak_value: number;
  peak_window: string;
}

export interface DashboardOnlineGroupPayload {
  name: string;
  value: number;
  percent: number;
}

export interface DashboardOnlinePayload {
  current_sessions: number;
  groups: DashboardOnlineGroupPayload[];
}

export interface DashboardNodePayload {
  name: string;
  region: string;
  latency_ms: number;
  load: number;
  status: string;
}

export interface DashboardNodesPayload {
  online: number;
  total: number;
  items: DashboardNodePayload[];
}

export interface DashboardTaskPayload {
  name: string;
  detail: string;
  progress: number;
  success: number;
  failed: number;
  processing: number;
  pending: number;
  submitted: number;
}

export interface DashboardOverviewPayload {
  generated_at: string;
  metrics: DashboardMetricPayload;
  storage: DashboardStoragePayload;
  traffic: DashboardTrafficPayload;
  online: DashboardOnlinePayload;
  nodes: DashboardNodesPayload;
  tasks: DashboardTaskPayload[];
}

export function getDashboardOverview(): Promise<DashboardOverviewPayload> {
  return request.get<DashboardOverviewPayload>('/api/v1/admin/dashboard/overview');
}

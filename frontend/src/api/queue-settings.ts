import request from './request';

export interface QueueSettingItemPayload {
  queue_key: 'metadata' | 'blob' | 'io' | 'offline' | 'thumbnail';
  worker_num: number;
  max_execution: number;
  backoff_factor: number;
  max_backoff: number;
  max_retry: number;
  retry_delay: number;
}

export interface QueueStatsItemPayload {
  queue_key: 'metadata' | 'blob' | 'io' | 'offline' | 'thumbnail';
  success: number;
  failed: number;
  processing: number;
  pending: number;
  submitted: number;
}

export interface QueueJobItemPayload {
  id: number;
  queue_key: 'metadata' | 'blob' | 'io' | 'offline' | 'thumbnail';
  job_type: string;
  resource_type: string;
  resource_id: string;
  dispatch_node?: {
    id: number;
    name: string;
    type: string;
  };
  node_capability: string;
  execution_mode: 'unified_runner' | 'runtime_record' | string;
  execution_note: string;
  status: string;
  attempts: number;
  max_attempts: number;
  scheduled_at: string;
  started_at?: string | null;
  finished_at?: string | null;
  last_error: string;
  result: string;
  payload: string;
  created_at: string;
}

export interface QueueJobListPayload {
  list: QueueJobItemPayload[];
  total: number;
  page: number;
  page_size: number;
}

export interface QueueJobMutationPayload {
  deleted: number;
}

export interface QueueJobStaleRecoveryPayload {
  recovered: number;
}

export interface QueueRunnerHeartbeatPayload {
  mode: string;
  process: string;
  host: string;
  pid: number;
  queues: string[];
  started_at: string;
  updated_at: string;
}

export interface QueueRuntimeStatusPayload {
  embedded_runner_enabled: boolean;
  worker_enabled: boolean;
  heartbeat_available: boolean;
  independent_worker_seen: boolean;
  embedded_runner_seen: boolean;
  runner_count: number;
  runners: QueueRunnerHeartbeatPayload[];
  message: string;
}

export function getQueueSettings(): Promise<QueueSettingItemPayload[]> {
  return request.get<QueueSettingItemPayload[]>('/api/v1/admin/queue-settings');
}

export function updateQueueSettings(data: QueueSettingItemPayload[]): Promise<QueueSettingItemPayload[]> {
  return request.put<QueueSettingItemPayload[]>('/api/v1/admin/queue-settings', data);
}

export function getQueueStats(): Promise<QueueStatsItemPayload[]> {
  return request.get<QueueStatsItemPayload[]>('/api/v1/admin/queue-stats');
}

export function getQueueRuntime(): Promise<QueueRuntimeStatusPayload> {
  return request.get<QueueRuntimeStatusPayload>('/api/v1/admin/queue-runtime');
}

export function getQueueJobs(params?: {
  queue_key?: QueueSettingItemPayload['queue_key'] | '';
  status?: string;
  node_id?: number | '';
  page?: number;
  page_size?: number;
}): Promise<QueueJobListPayload> {
  return request.get<QueueJobListPayload>('/api/v1/admin/queue-jobs', { params });
}

export function getQueueJob(id: number): Promise<QueueJobItemPayload> {
  return request.get<QueueJobItemPayload>(`/api/v1/admin/queue-jobs/${id}`);
}

export function retryQueueJob(id: number): Promise<QueueJobItemPayload> {
  return request.post<QueueJobItemPayload>(`/api/v1/admin/queue-jobs/${id}/retry`);
}

export function recoverStaleQueueJobs(params?: {
  queue_key?: QueueSettingItemPayload['queue_key'] | '';
}): Promise<QueueJobStaleRecoveryPayload> {
  return request.post<QueueJobStaleRecoveryPayload>('/api/v1/admin/queue-jobs/recover-stale', undefined, { params });
}

export function deleteQueueJob(id: number): Promise<QueueJobMutationPayload> {
  return request.delete<QueueJobMutationPayload>(`/api/v1/admin/queue-jobs/${id}`);
}

export function batchDeleteQueueJobs(jobIds: number[]): Promise<QueueJobMutationPayload> {
  return request.post<QueueJobMutationPayload>('/api/v1/admin/queue-jobs/batch-delete', {
    job_ids: jobIds,
  });
}

export function clearQueueJobs(data: {
  queue_key?: QueueSettingItemPayload['queue_key'] | '';
  status?: string;
}): Promise<QueueJobMutationPayload> {
  return request.post<QueueJobMutationPayload>('/api/v1/admin/queue-jobs/clear', data);
}

import type { QueueJobItemPayload, QueueSettingItemPayload, QueueStatsItemPayload } from '@/api/queue-settings';

export type QueueKey = QueueSettingItemPayload['queue_key'];
export type TaskStatus = '' | 'pending' | 'processing' | 'completed' | 'failed';

export interface TaskFilters {
  queueKey: QueueKey | '';
  status: TaskStatus;
  nodeId: number | '';
}

export interface QueueOption {
  key: QueueKey;
  label: string;
  description: string;
}

export interface StatusOption {
  key: TaskStatus;
  label: string;
}

export interface TaskMetric {
  key: string;
  label: string;
  value: number;
  hint: string;
  tone: 'sky' | 'mint' | 'rose' | 'amber';
}

export type TaskJob = QueueJobItemPayload;
export type QueueStats = QueueStatsItemPayload;

export const queueOptions: QueueOption[] = [
  { key: 'metadata', label: '元数据', description: '文件信息、索引元数据与属性刷新' },
  { key: 'blob', label: '文件 Blob', description: '对象扫描、回收与物理文件处理' },
  { key: 'io', label: 'I/O', description: '统一 runner 执行 multipart.cleanup 与 fulltext.rebuild；压缩/解压为运行时记录' },
  { key: 'offline', label: '离线下载', description: '离线下载与远端节点接管任务' },
  { key: 'thumbnail', label: '缩略图', description: '图片缩略图与预览生成' },
];

export const statusOptions: StatusOption[] = [
  { key: '', label: '全部状态' },
  { key: 'pending', label: '等待中' },
  { key: 'processing', label: '处理中' },
  { key: 'completed', label: '已完成' },
  { key: 'failed', label: '失败' },
];

export const queueLabelMap = queueOptions.reduce<Record<string, string>>((acc, item) => {
  acc[item.key] = item.label;
  return acc;
}, {});

export const statusLabelMap: Record<string, string> = {
  pending: '等待中',
  processing: '处理中',
  completed: '已完成',
  failed: '失败',
};

export const capabilityLabelMap: Record<string, string> = {
  create_archive: '创建压缩文件',
  extract_archive: '解压缩',
  offline_download: '离线下载',
};

export const executionModeLabelMap: Record<string, string> = {
  unified_runner: '统一 runner 任务',
  runtime_record: '运行时记录任务',
};

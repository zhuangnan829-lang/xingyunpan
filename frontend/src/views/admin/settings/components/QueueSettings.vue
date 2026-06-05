<template>
  <section class="queue-settings">
    <div class="queue-shell">
      <div class="luxury-line luxury-line-left" />
      <div class="luxury-line luxury-line-right" />

      <header class="queue-hero">
        <div class="hero-copy">
          <p class="hero-kicker">参数设置</p>
          <h2>队列</h2>
          <p class="hero-text">管理异步任务队列、并发数、重试机制和后台任务处理策略。</p>
        </div>

        <button class="refresh-button" type="button" :disabled="loading" @click="refreshStats">
          <el-icon class="refresh-icon" :class="{ spinning: loading }"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </header>

      <section class="runtime-panel">
        <div class="runtime-copy">
          <p class="hero-kicker">Runner 模式</p>
          <h3>{{ runtimeModeTitle }}</h3>
          <p class="hero-text">{{ runtime?.message || '正在读取统一队列 runner 运行状态。' }}</p>
        </div>

        <div class="runtime-status-grid">
          <div class="runtime-status">
            <span class="runtime-dot" :class="{ on: runtime?.embedded_runner_enabled }" />
            <div>
              <strong>server 内置 runner</strong>
              <p>{{ runtime?.embedded_runner_enabled ? '配置启用' : '配置关闭' }} · {{ runtime?.embedded_runner_seen ? '已检测到心跳' : '未检测到心跳' }}</p>
            </div>
          </div>
          <div class="runtime-status">
            <span class="runtime-dot" :class="{ on: runtime?.independent_worker_seen }" />
            <div>
              <strong>独立 worker</strong>
              <p>{{ runtime?.worker_enabled === false ? '配置关闭' : '配置允许启动' }} · {{ runtime?.independent_worker_seen ? '已检测到心跳' : '未检测到心跳' }}</p>
            </div>
          </div>
          <div class="runtime-status">
            <span class="runtime-count">{{ runtime?.runner_count ?? 0 }}</span>
            <div>
              <strong>当前 runner 进程</strong>
              <p>{{ runtime?.heartbeat_available ? runnerSummary : '心跳不可用，请查看启动日志' }}</p>
            </div>
          </div>
        </div>
      </section>

      <div class="queue-grid">
        <article v-for="queue in queueCards" :key="queue.key" class="queue-card" :class="`tone-${queue.tone}`">
          <div class="card-topline">
            <div class="card-heading">
              <div class="queue-icon">
                <svg v-if="queue.key === 'metadata'" viewBox="0 0 64 64" aria-hidden="true"><rect x="8" y="16" width="28" height="32" rx="6" /><path d="M18 16v32M26 16v32M8 24h28M8 40h28" /><circle cx="46" cy="32" r="10" /><circle cx="46" cy="32" r="4" /></svg>
                <svg v-else-if="queue.key === 'blob'" viewBox="0 0 64 64" aria-hidden="true"><path d="M18 14h28v10H18zM16 24h32v18a8 8 0 0 1-8 8H24a8 8 0 0 1-8-8z" /><path d="M24 34c3 2 5 5 5 8a5 5 0 1 1-10 0c0-3 2-6 5-8Z" /><path d="M42 29a10 10 0 1 1-2 18" /><path d="m43 26 5 3-4 4" /></svg>
                <svg v-else-if="queue.key === 'io'" viewBox="0 0 64 64" aria-hidden="true"><rect x="18" y="18" width="28" height="28" rx="6" /><rect x="26" y="26" width="12" height="12" rx="2" /><path d="M10 24h8M10 32h8M10 40h8M46 24h8M46 32h8M46 40h8M24 10v8M32 10v8M40 10v8M24 46v8M32 46v8M40 46v8" /></svg>
                <svg v-else-if="queue.key === 'offline'" viewBox="0 0 64 64" aria-hidden="true"><path d="M35 12 22 32l10 3 5 17 5-20 10-7Z" /><path d="M11 38a21 21 0 1 0 42 0" /><path d="M32 23v18m0 0-6-6m6 6 6-6" /></svg>
                <svg v-else viewBox="0 0 64 64" aria-hidden="true"><rect x="14" y="18" width="20" height="16" rx="4" /><path d="m34 23 10-6v18l-10-6" /><rect x="16" y="42" width="8" height="8" rx="2" /><rect x="28" y="42" width="8" height="8" rx="2" /><rect x="40" y="42" width="8" height="8" rx="2" /></svg>
              </div>

              <div class="title-copy">
                <h3>{{ queue.title }}</h3>
                <p>{{ queue.description }}</p>
              </div>
            </div>

            <button class="settings-button" type="button" @click="openDialog(queue.key)">
              <span class="settings-pulse" />
              <el-icon><Setting /></el-icon>
            </button>
          </div>

          <div class="progress-bar" :class="{ active: queue.key === 'blob' && statsByQueueKey[queue.key].submitted > 0 }" />

          <div class="status-grid">
            <div v-for="status in queue.statuses" :key="status.label" class="status-card">
              <div class="status-ring" :style="{ '--ring-color': status.color, '--ring-value': `${status.percent}deg` }">
                <span>{{ status.count }}</span>
              </div>
              <div class="status-copy">
                <strong>{{ status.label }}</strong>
                <span>{{ formatCount(status.count) }}</span>
              </div>
              <svg class="status-trend" viewBox="0 0 64 20" preserveAspectRatio="none" aria-hidden="true">
                <polyline :points="status.trend" />
              </svg>
            </div>
          </div>
        </article>
      </div>

      <section class="jobs-panel">
        <div class="jobs-head">
          <div class="jobs-copy">
            <p class="hero-kicker">队列任务明细</p>
            <h3>统一任务队列</h3>
            <p class="hero-text">直接查看每条 queue_job 的状态、payload、结果和错误信息。</p>
          </div>

          <div class="jobs-filters">
            <select v-model="jobFilterQueue" class="jobs-select" @change="loadJobs(1)">
              <option value="">全部队列</option>
              <option v-for="queue in queueBlueprint" :key="queue.key" :value="queue.key">{{ queue.title }}</option>
            </select>
            <select v-model="jobFilterStatus" class="jobs-select" @change="loadJobs(1)">
              <option value="">全部状态</option>
              <option value="pending">pending</option>
              <option value="processing">processing</option>
              <option value="completed">completed</option>
              <option value="failed">failed</option>
            </select>
            <button class="job-tool-button" type="button" :disabled="operationLoading || selectedDeletableJobIds.length === 0" @click="confirmBatchDelete">
              批量删除
            </button>
            <button class="job-tool-button danger" type="button" :disabled="operationLoading" @click="confirmClearFinishedJobs">
              清理完成/失败
            </button>
            <button class="job-tool-button accent" type="button" :disabled="operationLoading" @click="confirmRecoverStaleJobs">
              恢复超时 processing
            </button>
          </div>
        </div>

        <div class="jobs-table-shell">
          <table class="jobs-table">
            <thead>
              <tr>
                <th>
                  <input class="job-check" type="checkbox" :checked="isCurrentPageSelected" :disabled="operationLoading || !deletableCurrentPageJobs.length" @change="toggleCurrentPageSelection" />
                </th>
                <th>ID</th>
                <th>队列</th>
                <th>任务类型</th>
                <th>执行方式</th>
                <th>状态</th>
                <th>尝试</th>
                <th>计划时间</th>
                <th>Payload</th>
                <th>结果 / 错误</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="job in queueJobs" :key="job.id">
                <td>
                  <input class="job-check" type="checkbox" :checked="selectedJobIds.includes(job.id)" :disabled="operationLoading || job.status === 'processing'" @change="toggleJobSelection(job)" />
                </td>
                <td>#{{ job.id }}</td>
                <td><span class="job-chip">{{ job.queue_key }}</span></td>
                <td><div class="job-stack"><strong>{{ job.job_type }}</strong><small>{{ job.resource_type }} / {{ job.resource_id }}</small></div></td>
                <td><span class="execution-chip" :class="`mode-${job.execution_mode || 'unified_runner'}`">{{ executionModeLabel(job) }}</span></td>
                <td><span class="job-status" :class="`status-${job.status}`">{{ job.status }}</span></td>
                <td>{{ job.attempts }} / {{ job.max_attempts }}</td>
                <td>{{ formatDateTime(job.scheduled_at) }}</td>
                <td><pre class="job-code">{{ formatJson(job.payload) }}</pre></td>
                <td><pre class="job-code">{{ formatJson(job.result || job.last_error || '') }}</pre></td>
                <td>
                  <div class="job-actions">
                    <button class="job-action-button accent" type="button" :disabled="operationLoading || !canRetryJob(job)" @click="confirmRetryJob(job)">重试</button>
                    <button class="job-action-button danger" type="button" :disabled="operationLoading || job.status === 'processing'" @click="confirmDeleteJob(job)">删除</button>
                  </div>
                </td>
              </tr>
              <tr v-if="!queueJobs.length">
                <td colspan="11" class="jobs-empty">暂无任务数据</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="jobs-pagination">
          <button class="pager-button" type="button" :disabled="jobPage <= 1" @click="loadJobs(jobPage - 1)">上一页</button>
          <span>第 {{ jobPage }} 页 / 共 {{ totalJobPages }} 页</span>
          <button class="pager-button" type="button" :disabled="jobPage >= totalJobPages" @click="loadJobs(jobPage + 1)">下一页</button>
        </div>
      </section>
    </div>

    <el-dialog v-model="dialogVisible" :show-close="false" width="780px" class="queue-dialog" append-to-body destroy-on-close>
      <template #header>
        <div class="dialog-header">
          <div>
            <p class="dialog-kicker">队列设置</p>
            <h3>编辑队列设置 - {{ activeQueue?.title }}</h3>
          </div>
          <button class="dialog-close" type="button" @click="closeDialog"><el-icon><Close /></el-icon></button>
        </div>
      </template>

      <div v-if="draftConfig" class="dialog-body">
        <div class="dialog-section-head">
          <p>队列设置</p>
          <span>保持原始字段和默认值，优化为更清晰的双列表单结构。</span>
        </div>

        <label v-for="field in queueFields" :key="field.key" class="field-card">
          <div class="field-head">
            <span class="field-icon">
              <svg v-if="field.key === 'worker_num'" viewBox="0 0 24 24" aria-hidden="true"><path d="M8 6v12M12 4v16M16 7v10M5 9h14" /></svg>
              <svg v-else-if="field.key === 'max_execution'" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="7" /><path d="M12 8v4l3 2" /></svg>
              <svg v-else-if="field.key === 'backoff_factor'" viewBox="0 0 24 24" aria-hidden="true"><path d="M6 16 10 12 13 14 18 8" /><path d="M15 8h3v3" /></svg>
              <svg v-else-if="field.key === 'max_backoff'" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 4v8" /><path d="M12 12 8.5 15.5" /><path d="M6 18h12" /></svg>
              <svg v-else-if="field.key === 'max_retry'" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 7h8a4 4 0 1 1 0 8H9" /><path d="m7 7 2-2M7 7l2 2" /></svg>
              <svg v-else viewBox="0 0 24 24" aria-hidden="true"><path d="M7 12h10" /><path d="M12 7v10" /><rect x="5" y="5" width="14" height="14" rx="4" /></svg>
            </span>
            <div class="field-copy">
              <strong>{{ field.label }}</strong>
              <p>{{ field.description }}</p>
            </div>
          </div>

          <div class="field-control">
            <input v-model.number="draftConfig[field.key]" class="field-input" type="number" min="0" />
            <span class="field-glow" />
          </div>
        </label>
      </div>

      <template #footer>
        <div class="dialog-actions">
          <button class="dialog-button cancel" type="button" @click="closeDialog">取消</button>
          <button class="dialog-button confirm" type="button" @click="confirmDialog">确定</button>
        </div>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Close, Refresh, Setting } from '@element-plus/icons-vue';
import {
  batchDeleteQueueJobs,
  clearQueueJobs,
  deleteQueueJob,
  getQueueJobs,
  getQueueRuntime,
  getQueueSettings,
  getQueueStats,
  recoverStaleQueueJobs,
  retryQueueJob,
  updateQueueSettings,
  type QueueJobItemPayload,
  type QueueRuntimeStatusPayload,
  type QueueSettingItemPayload,
  type QueueStatsItemPayload,
} from '@/api/queue-settings';

type QueueKey = QueueSettingItemPayload['queue_key'];
type QueueConfig = Omit<QueueSettingItemPayload, 'queue_key'>;
type FieldKey = keyof QueueConfig;
type QueueStatus = { label: string; count: number; color: string; trend: string; percent: number };
type QueueCard = { key: QueueKey; title: string; description: string; tone: string; statuses: QueueStatus[] };

const queueFields: { key: FieldKey; label: string; description: string }[] = [
  { key: 'worker_num', label: '工作线程数', description: '每个 runner 进程每轮最多领取的任务数量；如果 server 内置 runner 与独立 worker 同时运行，总并发会按进程叠加。' },
  { key: 'max_execution', label: '最大执行时间', description: '任务最大执行时间（秒），超过此时间的任务将被终止。' },
  { key: 'backoff_factor', label: '退避因子', description: '任务重试时间间隔的增长因子。' },
  { key: 'max_backoff', label: '最大退避时间', description: '任务重试的最大退避时间（秒）。' },
  { key: 'max_retry', label: '最大重试次数', description: '任务失败后的最大重试次数。' },
  { key: 'retry_delay', label: '重试延迟', description: '任务重试的初始延迟时间（秒）。' },
];

const defaults: Record<QueueKey, QueueConfig> = {
  metadata: { worker_num: 30, max_execution: 3600, backoff_factor: 2, max_backoff: 60, max_retry: 1, retry_delay: 0 },
  blob: { worker_num: 5, max_execution: 900, backoff_factor: 2, max_backoff: 60, max_retry: 0, retry_delay: 0 },
  io: { worker_num: 30, max_execution: 2592000, backoff_factor: 2, max_backoff: 600, max_retry: 5, retry_delay: 0 },
  offline: { worker_num: 5, max_execution: 864000, backoff_factor: 2, max_backoff: 600, max_retry: 5, retry_delay: 0 },
  thumbnail: { worker_num: 15, max_execution: 300, backoff_factor: 2, max_backoff: 60, max_retry: 0, retry_delay: 0 },
};

const queueBlueprint = [
  { key: 'metadata' as const, title: '媒体元数据提取', description: '用于提取媒体文件的元数据。', tone: 'sky' },
  { key: 'blob' as const, title: 'Blob 回收', description: '用于删除过期的文件 Blob。', tone: 'mint' },
  { key: 'io' as const, title: 'IO 密集型', description: '统一队列 runner 执行 multipart.cleanup、fulltext.rebuild、压缩打包与解压任务；此处参数会控制每轮领取、超时和重试。', tone: 'violet' },
  { key: 'offline' as const, title: '离线下载', description: '用于处理离线下载任务。', tone: 'amber' },
  { key: 'thumbnail' as const, title: '缩略图生成', description: '用于为文件生成缩略图。', tone: 'rose' },
];

const statusMeta = [
  { label: '成功', color: '#49c7db', trend: '2,14 18,12 34,9 48,8 62,5' },
  { label: '失败', color: '#c4ced9', trend: '2,17 18,17 34,17 48,17 62,17' },
  { label: '处理中', color: '#79b8ff', trend: '2,17 18,16 34,15 48,15 62,14' },
  { label: '挂起', color: '#c7b38f', trend: '2,17 18,17 34,16 48,16 62,15' },
  { label: '已提交', color: '#9dc3e8', trend: '2,15 18,13 34,11 48,8 62,6' },
];

const snapshots = ref(JSON.stringify(defaults));
const configs = reactive<Record<QueueKey, QueueConfig>>(JSON.parse(JSON.stringify(defaults)));
const stats = ref<QueueStatsItemPayload[]>([]);
const runtime = ref<QueueRuntimeStatusPayload | null>(null);
const queueJobs = ref<QueueJobItemPayload[]>([]);
const dialogVisible = ref(false);
const activeKey = ref<QueueKey | null>(null);
const draftConfig = ref<QueueConfig | null>(null);
const loading = ref(false);
const saving = ref(false);
const operationLoading = ref(false);
const jobFilterQueue = ref<QueueKey | ''>('');
const jobFilterStatus = ref('');
const jobPage = ref(1);
const jobPageSize = 10;
const jobTotal = ref(0);
const selectedJobIds = ref<number[]>([]);
const isDirty = computed(() => JSON.stringify(configs) !== snapshots.value);
const totalJobPages = computed(() => Math.max(1, Math.ceil(jobTotal.value / jobPageSize)));
const runtimeModeTitle = computed(() => {
  if (!runtime.value) return '读取中';
  if (runtime.value.embedded_runner_seen && runtime.value.independent_worker_seen) return 'server 内置 runner + 独立 worker 同时运行';
  if (runtime.value.independent_worker_seen) return '独立 worker 模式';
  if (runtime.value.embedded_runner_seen) return 'server 内置 runner 模式';
  if (runtime.value.embedded_runner_enabled) return '内置 runner 已配置启用，等待心跳';
  return '内置 runner 已关闭';
});
const runnerSummary = computed(() => {
  const runners = runtime.value?.runners ?? [];
  if (!runners.length) return '未检测到 runner 心跳';
  return runners.map((item) => `${runnerModeLabel(item.mode)} ${item.host}:${item.pid}`).join('、');
});
const deletableCurrentPageJobs = computed(() => queueJobs.value.filter((job) => job.status !== 'processing'));
const selectedDeletableJobIds = computed(() => selectedJobIds.value.filter((id) => {
  const job = queueJobs.value.find((item) => item.id === id);
  return job && job.status !== 'processing';
}));
const isCurrentPageSelected = computed(() => {
  const ids = deletableCurrentPageJobs.value.map((job) => job.id);
  return ids.length > 0 && ids.every((id) => selectedJobIds.value.includes(id));
});

const statsByQueueKey = computed<Record<QueueKey, QueueStatsItemPayload>>(() => {
  const fallback: Record<QueueKey, QueueStatsItemPayload> = {
    metadata: { queue_key: 'metadata', success: 0, failed: 0, processing: 0, pending: 0, submitted: 0 },
    blob: { queue_key: 'blob', success: 0, failed: 0, processing: 0, pending: 0, submitted: 0 },
    io: { queue_key: 'io', success: 0, failed: 0, processing: 0, pending: 0, submitted: 0 },
    offline: { queue_key: 'offline', success: 0, failed: 0, processing: 0, pending: 0, submitted: 0 },
    thumbnail: { queue_key: 'thumbnail', success: 0, failed: 0, processing: 0, pending: 0, submitted: 0 },
  };

  for (const item of stats.value) fallback[item.queue_key] = item;
  return fallback;
});

const queueCards = computed<QueueCard[]>(() => queueBlueprint.map((queue) => {
  const values = statsByQueueKey.value[queue.key];
  const counts = [values.success, values.failed, values.processing, values.pending, values.submitted];
  const total = Math.max(counts.reduce((sum, count) => sum + count, 0), 1);
  return {
    ...queue,
    statuses: statusMeta.map((status, index) => {
      const count = counts[index];
      const ratio = count > 0 ? count / total : 0;
      return { ...status, count, percent: Math.max(ratio * 360, count > 0 ? 18 : 0) };
    }),
  };
}));

const activeQueue = computed(() => queueBlueprint.find((item) => item.key === activeKey.value) ?? null);

function cloneConfig(config: QueueConfig): QueueConfig { return JSON.parse(JSON.stringify(config)); }
function formatCount(count: number) { return `(${count.toLocaleString('zh-CN')})`; }
function formatDateTime(value?: string | null) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'; }
function formatJson(value?: string | null) {
  if (!value) return '-';
  try { return JSON.stringify(JSON.parse(value), null, 2); } catch { return value; }
}
function executionModeLabel(job: QueueJobItemPayload) {
  return job.execution_mode === 'runtime_record' ? '运行时记录' : '统一 runner';
}
function runnerModeLabel(mode: string) {
  return mode === 'server_embedded' ? 'server' : mode === 'worker' ? 'worker' : mode;
}

function fallbackRuntimeStatus(message: string): QueueRuntimeStatusPayload {
  return {
    embedded_runner_enabled: false,
    worker_enabled: true,
    heartbeat_available: false,
    independent_worker_seen: false,
    embedded_runner_seen: false,
    runner_count: 0,
    runners: [],
    message,
  };
}

async function loadRuntimeStatus() {
  try {
    return await getQueueRuntime();
  } catch (error) {
    console.warn('load queue runtime status failed', error);
    return fallbackRuntimeStatus('Runner 状态接口暂不可用；队列设置和任务列表可继续使用，请检查后端是否已重启到最新版本。');
  }
}

function openDialog(key: QueueKey) {
  activeKey.value = key;
  draftConfig.value = cloneConfig(configs[key]);
  dialogVisible.value = true;
}

function closeDialog() {
  dialogVisible.value = false;
  draftConfig.value = null;
  activeKey.value = null;
}

function confirmDialog() {
  if (!activeKey.value || !draftConfig.value) return;
  configs[activeKey.value] = cloneConfig(draftConfig.value);
  ElMessage.success(`“${activeQueue.value?.title ?? '队列'}”设置已更新，记得点击保存更改`);
  closeDialog();
}

async function loadJobs(page = 1) {
  const data = await getQueueJobs({
    queue_key: jobFilterQueue.value,
    status: jobFilterStatus.value,
    page,
    page_size: jobPageSize,
  });
  queueJobs.value = data.list;
  jobTotal.value = data.total;
  jobPage.value = data.page;
  syncSelectedJobsWithCurrentPage();
}

async function refreshStats() {
  loading.value = true;
  try {
    await refreshQueueData(jobPage.value);
    ElMessage.success('队列状态已刷新');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '刷新队列状态失败');
  } finally {
    loading.value = false;
  }
}

async function refreshQueueData(page = jobPage.value) {
  const [queueStats, runtimeStatus] = await Promise.all([getQueueStats(), loadRuntimeStatus(), loadJobs(page)]);
  stats.value = queueStats;
  runtime.value = runtimeStatus;
}

async function reload() {
  loading.value = true;
  try {
    const [settings, queueStats, runtimeStatus] = await Promise.all([getQueueSettings(), getQueueStats(), loadRuntimeStatus()]);
    applySettings(settings);
    stats.value = queueStats;
    runtime.value = runtimeStatus;
    snapshots.value = createSnapshot(configs);
    await loadJobs(jobPage.value);
    ElMessage.success('队列配置已重新加载');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载队列配置失败');
  } finally {
    loading.value = false;
  }
}

function reset() {
  (Object.keys(defaults) as QueueKey[]).forEach((key) => { configs[key] = cloneConfig(defaults[key]); });
  ElMessage.success('队列配置已恢复默认值，记得保存');
}

function canRetryJob(job: QueueJobItemPayload) {
  return job.status === 'failed' || job.status === 'pending';
}

function syncSelectedJobsWithCurrentPage() {
  const currentIds = new Set(queueJobs.value.map((job) => job.id));
  selectedJobIds.value = selectedJobIds.value.filter((id) => currentIds.has(id));
}

function toggleJobSelection(job: QueueJobItemPayload) {
  if (job.status === 'processing') return;
  selectedJobIds.value = selectedJobIds.value.includes(job.id)
    ? selectedJobIds.value.filter((id) => id !== job.id)
    : [...selectedJobIds.value, job.id];
}

function toggleCurrentPageSelection() {
  const ids = deletableCurrentPageJobs.value.map((job) => job.id);
  if (isCurrentPageSelected.value) {
    selectedJobIds.value = selectedJobIds.value.filter((id) => !ids.includes(id));
    return;
  }
  selectedJobIds.value = Array.from(new Set([...selectedJobIds.value, ...ids]));
}

async function runJobOperation(action: () => Promise<unknown>, successMessage: string) {
  operationLoading.value = true;
  try {
    await action();
    await refreshQueueData(jobPage.value);
    ElMessage.success(successMessage);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '队列任务操作失败');
  } finally {
    operationLoading.value = false;
  }
}

async function askJobConfirm(message: string, title: string, confirmButtonText: string) {
  try {
    await ElMessageBox.confirm(message, title, { type: 'warning', confirmButtonText, cancelButtonText: '取消' });
    return true;
  } catch {
    return false;
  }
}

async function confirmRetryJob(job: QueueJobItemPayload) {
  if (!canRetryJob(job)) return;
  if (!await askJobConfirm(`确定要重试任务 #${job.id} 吗？`, '重试队列任务', '重试')) return;
  await runJobOperation(() => retryQueueJob(job.id), `任务 #${job.id} 已重新加入队列`);
}

async function confirmDeleteJob(job: QueueJobItemPayload) {
  if (job.status === 'processing') {
    ElMessage.warning('processing 任务不能直接删除，请先恢复超时任务或等待执行结束');
    return;
  }
  if (!await askJobConfirm(`确定要删除任务 #${job.id} 吗？此操作不可恢复。`, '删除队列任务', '删除')) return;
  await runJobOperation(() => deleteQueueJob(job.id), `任务 #${job.id} 已删除`);
}

async function confirmBatchDelete() {
  const ids = selectedDeletableJobIds.value;
  if (!ids.length) return;
  if (!await askJobConfirm(`确定要删除选中的 ${ids.length} 个非 processing 任务吗？`, '批量删除队列任务', '批量删除')) return;
  await runJobOperation(() => batchDeleteQueueJobs(ids), `已删除 ${ids.length} 个队列任务`);
  selectedJobIds.value = [];
}

async function confirmClearFinishedJobs() {
  const scope = jobFilterQueue.value ? `队列 ${jobFilterQueue.value}` : '全部队列';
  if (!await askJobConfirm(`确定要清理${scope}中的 completed/failed 任务吗？`, '清理完成任务', '清理')) return;
  await runJobOperation(() => clearQueueJobs({ queue_key: jobFilterQueue.value }), 'completed/failed 任务已清理');
  selectedJobIds.value = [];
}

async function confirmRecoverStaleJobs() {
  const scope = jobFilterQueue.value ? `队列 ${jobFilterQueue.value}` : '全部队列';
  if (!await askJobConfirm(`确定要恢复${scope}中已超时的 processing 任务吗？`, '恢复超时任务', '恢复')) return;
  await runJobOperation(() => recoverStaleQueueJobs({ queue_key: jobFilterQueue.value }), '超时 processing 任务已恢复');
}

async function save() {
  saving.value = true;
  try {
    const saved = await updateQueueSettings(toSettingsPayload(configs));
    applySettings(saved);
    snapshots.value = createSnapshot(configs);
    ElMessage.success('队列配置已保存');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存队列配置失败');
  } finally {
    saving.value = false;
  }
}

defineExpose({ isDirty, loading, saving, reload, reset, save });

function applySettings(items: QueueSettingItemPayload[]) {
  for (const item of items) {
    configs[item.queue_key] = {
      worker_num: item.worker_num,
      max_execution: item.max_execution,
      backoff_factor: item.backoff_factor,
      max_backoff: item.max_backoff,
      max_retry: item.max_retry,
      retry_delay: item.retry_delay,
    };
  }
}

function toSettingsPayload(source: Record<QueueKey, QueueConfig>): QueueSettingItemPayload[] {
  return (Object.keys(source) as QueueKey[]).map((key) => ({ queue_key: key, ...source[key] }));
}

function createSnapshot(source: Record<QueueKey, QueueConfig>) { return JSON.stringify(source); }

onMounted(async () => { await reload(); });
</script>

<style scoped>
.queue-settings{min-height:100%;color:#1f3f64}.queue-shell{position:relative;overflow:hidden;padding:30px;border:1px solid rgba(255,255,255,.92);border-radius:34px;background:radial-gradient(circle at top left,rgba(148,220,255,.26),transparent 22%),radial-gradient(circle at right 12%,rgba(122,201,229,.18),transparent 18%),linear-gradient(180deg,rgba(255,253,248,.98),rgba(252,249,242,.92));box-shadow:inset 0 1px 0 rgba(255,255,255,.94),0 28px 60px rgba(75,113,145,.12),0 16px 36px rgba(117,173,211,.08)}.queue-shell::after{content:'';position:absolute;right:8%;bottom:-80px;width:360px;height:180px;background:radial-gradient(circle,rgba(102,210,235,.22),transparent 68%);pointer-events:none;filter:blur(18px)}.luxury-line{position:absolute;pointer-events:none;height:120px;border-radius:999px;opacity:.6}.luxury-line-left{top:148px;left:-10px;width:240px;border-top:1px solid rgba(214,181,119,.6);transform:rotate(-8deg)}.luxury-line-right{right:44px;bottom:94px;width:260px;border-top:1px solid rgba(86,190,222,.48);transform:rotate(-12deg)}
.queue-hero,.card-topline,.card-heading,.dialog-header,.dialog-actions,.field-head,.jobs-head{display:flex;align-items:center;justify-content:space-between;gap:16px}.queue-hero{margin-bottom:18px}.hero-copy,.title-copy,.status-copy,.field-copy,.jobs-copy{display:grid;gap:8px}.hero-kicker,.dialog-kicker{margin:0;color:#4fa8d0;font-size:12px;font-weight:800;letter-spacing:.18em;text-transform:uppercase}h2,h3,p,strong,span{margin:0}h2{font-size:clamp(2.3rem,4vw,3.3rem);line-height:1;letter-spacing:-.05em;color:#1d3f63}.hero-text,.title-copy p,.status-copy span,.field-card p{color:#70859b;line-height:1.8}

.refresh-button,.settings-button,.dialog-button,.dialog-close,.pager-button{border:none;cursor:pointer;transition:transform .2s ease,box-shadow .2s ease,opacity .2s ease}.refresh-button{display:inline-flex;align-items:center;gap:10px;min-height:52px;padding:0 22px;border-radius:18px;color:#fff;background:linear-gradient(135deg,#4fd7ef 0%,#31b6de 55%,#2aa3d4 100%);box-shadow:0 18px 36px rgba(66,189,225,.28)}.refresh-icon.spinning{animation:spin .8s linear infinite}.queue-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:22px}
.runtime-panel{position:relative;z-index:1;display:grid;grid-template-columns:minmax(0,1fr) minmax(420px,.85fr);gap:20px;margin-bottom:22px;padding:22px;border:1px solid rgba(225,236,244,.92);border-radius:24px;background:linear-gradient(180deg,rgba(255,255,255,.88),rgba(248,251,253,.8));box-shadow:inset 0 1px 0 rgba(255,255,255,.94),0 14px 28px rgba(101,139,166,.08)}.runtime-copy{display:grid;gap:8px}.runtime-status-grid{display:grid;gap:10px}.runtime-status{display:grid;grid-template-columns:auto minmax(0,1fr);align-items:center;gap:12px;padding:12px 14px;border-radius:16px;background:rgba(255,255,255,.72);border:1px solid rgba(225,235,242,.82)}.runtime-status strong{color:#31536f}.runtime-status p{color:#72879d;font-size:13px;line-height:1.6}.runtime-dot{width:12px;height:12px;border-radius:50%;background:#cbd6df;box-shadow:0 0 0 4px rgba(203,214,223,.22)}.runtime-dot.on{background:#49c7db;box-shadow:0 0 0 4px rgba(73,199,219,.16),0 0 16px rgba(73,199,219,.38)}.runtime-count{display:grid;place-items:center;min-width:32px;height:32px;border-radius:12px;background:rgba(73,199,219,.12);color:#218fb7;font-weight:800}.queue-card,.jobs-panel{position:relative;display:grid;gap:18px;padding:24px;border:1px solid rgba(255,255,255,.88);border-radius:28px;background:linear-gradient(180deg,rgba(255,251,244,.88),rgba(255,248,239,.74));box-shadow:inset 0 1px 0 rgba(255,255,255,.95),0 20px 34px rgba(106,132,154,.1),0 8px 20px rgba(124,184,216,.08)}.queue-card::before,.jobs-panel::before{content:'';position:absolute;inset:0;border-radius:inherit;background:linear-gradient(135deg,var(--card-glow,rgba(111,217,242,.22)),transparent 52%);opacity:.38;pointer-events:none}.tone-sky{--card-glow:rgba(111,217,242,.38)}.tone-mint{--card-glow:rgba(138,228,209,.38)}.tone-violet{--card-glow:rgba(187,194,255,.38)}.tone-amber{--card-glow:rgba(231,209,167,.38)}.tone-rose{--card-glow:rgba(238,198,215,.38)}
.queue-icon{display:grid;place-items:center;width:74px;height:74px;border-radius:24px;background:linear-gradient(180deg,rgba(255,255,255,.78),rgba(221,240,247,.78));box-shadow:inset 0 1px 0 rgba(255,255,255,.92),0 14px 24px rgba(121,173,208,.14)}.queue-icon svg{width:40px;height:40px;fill:none;stroke:#2e8dbc;stroke-width:2.2;stroke-linecap:round;stroke-linejoin:round}.title-copy h3,.jobs-copy h3{font-size:1.9rem;color:#2392bf}.settings-button,.dialog-close{position:relative;display:grid;place-items:center;width:44px;height:44px;border-radius:16px;background:linear-gradient(180deg,rgba(255,255,255,.92),rgba(236,246,252,.88));color:#2392bf;box-shadow:0 10px 20px rgba(113,159,188,.14)}.settings-pulse{position:absolute;top:8px;right:8px;width:6px;height:6px;border-radius:50%;background:#54d0e8;box-shadow:0 0 10px rgba(84,208,232,.8)}
.progress-bar{height:10px;border-radius:999px;background:linear-gradient(90deg,rgba(209,220,228,.9),rgba(235,240,244,.9));overflow:hidden}.progress-bar.active{background:linear-gradient(90deg,#57dfd4 0%,#5fcfe6 55%,#81b9ef 100%);box-shadow:0 0 18px rgba(83,207,229,.35)}.status-grid{display:grid;grid-template-columns:repeat(5,minmax(0,1fr));gap:12px}.status-card{display:grid;justify-items:center;gap:10px;padding:14px 10px;border-radius:20px;background:linear-gradient(180deg,rgba(255,255,255,.78),rgba(248,244,236,.72));box-shadow:inset 0 1px 0 rgba(255,255,255,.92),0 10px 20px rgba(148,171,189,.08);text-align:center}
.status-ring{position:relative;width:60px;height:60px;display:grid;place-items:center;border-radius:50%;background:conic-gradient(var(--ring-color) 0deg,var(--ring-color) var(--ring-value),rgba(207,216,224,.6) var(--ring-value),rgba(207,216,224,.6) 360deg);box-shadow:0 0 22px color-mix(in srgb,var(--ring-color) 32%,transparent)}.status-ring::before{content:'';width:44px;height:44px;border-radius:50%;background:rgba(255,251,246,.96);box-shadow:inset 0 1px 0 rgba(255,255,255,.98)}.status-ring span{position:absolute;color:#208fc0;font-size:13px;font-weight:800}.status-copy strong{color:#71869b;font-size:13px;font-weight:700}.status-trend{width:100%;height:20px}.status-trend polyline{fill:none;stroke:var(--ring-color);stroke-width:2.2;stroke-linecap:round;stroke-linejoin:round;opacity:.8}
.jobs-panel{margin-top:24px;--card-glow:rgba(111,217,242,.22)}.jobs-head{align-items:flex-start}.jobs-filters{display:flex;gap:12px;flex-wrap:wrap}.jobs-select,.field-input{min-height:50px;width:100%;padding:0 16px;border:1px solid rgba(212,222,230,.95);border-radius:16px;background:#fff;color:#3c556d;outline:none}.jobs-select{min-width:180px;padding-right:36px}.job-tool-button,.job-action-button{border:none;cursor:pointer;white-space:nowrap;transition:transform .2s ease,box-shadow .2s ease,opacity .2s ease}.job-tool-button{min-height:50px;padding:0 16px;border-radius:16px;background:linear-gradient(180deg,rgba(255,255,255,.94),rgba(231,240,247,.9));color:#4e6a85;font-weight:800;box-shadow:0 10px 20px rgba(111,149,177,.1)}.job-tool-button.accent,.job-action-button.accent{color:#fff;background:linear-gradient(135deg,#4fd7ef 0%,#31b6de 55%,#2aa3d4 100%)}.job-tool-button.danger,.job-action-button.danger{color:#fff;background:linear-gradient(135deg,#f6a6a6 0%,#df7777 100%)}.job-tool-button:disabled,.job-action-button:disabled{opacity:.52;cursor:not-allowed;box-shadow:none}.job-tool-button:not(:disabled):hover,.job-action-button:not(:disabled):hover{transform:translateY(-1px)}.jobs-table-shell{overflow:auto;border-radius:22px;border:1px solid rgba(224,235,242,.95);background:rgba(255,255,255,.8)}.jobs-table{width:100%;min-width:1120px;border-collapse:collapse}.jobs-table th,.jobs-table td{padding:14px 16px;border-bottom:1px solid rgba(228,236,242,.92);text-align:left;vertical-align:top}.jobs-table th{color:#64829d;font-size:12px;letter-spacing:.08em;text-transform:uppercase}.job-check{width:16px;height:16px;accent-color:#35badf;cursor:pointer}.job-check:disabled{cursor:not-allowed;opacity:.45}.job-chip,.job-status,.execution-chip{display:inline-flex;align-items:center;justify-content:center;min-height:30px;padding:0 12px;border-radius:999px;font-size:12px;font-weight:700}.job-chip{background:rgba(73,199,219,.12);color:#248eb9}.execution-chip{background:rgba(214,226,238,.58);color:#526a82}.execution-chip.mode-runtime_record{background:rgba(254,243,199,.86);color:#8a5d13}.job-status{min-width:92px}.status-pending{background:rgba(199,179,143,.18);color:#8d6d33}.status-processing{background:rgba(121,184,255,.18);color:#366cbd}.status-completed{background:rgba(73,199,219,.16);color:#177e94}.status-failed{background:rgba(228,132,132,.18);color:#b34d4d}.job-stack{display:grid;gap:6px}.job-stack strong{color:#274663}.job-stack small{color:#7f93a8}.job-code{margin:0;max-width:320px;white-space:pre-wrap;word-break:break-word;color:#49617a;font-size:12px;line-height:1.65}.job-actions{display:flex;gap:8px;flex-wrap:wrap}.job-action-button{min-height:34px;padding:0 12px;border-radius:12px;font-size:12px;font-weight:800;background:linear-gradient(180deg,rgba(255,255,255,.94),rgba(231,240,247,.9));color:#4e6a85}.jobs-empty{text-align:center;color:#8da0b2}.jobs-pagination{display:flex;align-items:center;justify-content:flex-end;gap:12px;color:#5f7992}.pager-button{min-height:38px;padding:0 16px;border-radius:14px;background:linear-gradient(180deg,rgba(236,242,247,.96),rgba(222,231,239,.92));color:#4d6986}.pager-button:disabled,.refresh-button:disabled{opacity:.72;cursor:not-allowed}
.dialog-header{align-items:flex-start}.dialog-header h3{color:#204364;font-size:2rem;letter-spacing:-.04em}.dialog-body{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:18px}.dialog-section-head{grid-column:1/-1;display:grid;gap:6px;margin-bottom:2px}.dialog-section-head p{color:#2b94c1;font-size:13px;font-weight:800;letter-spacing:.12em;text-transform:uppercase}.dialog-section-head span,.field-copy p{color:#7b8ea2;font-size:13px}.field-card{display:grid;gap:14px;padding:18px 18px 16px;border:1px solid rgba(227,238,246,.95);border-radius:22px;background:linear-gradient(180deg,rgba(255,255,255,.94),rgba(248,251,253,.9));box-shadow:inset 0 1px 0 rgba(255,255,255,.96),0 14px 24px rgba(140,171,193,.08)}.field-head{justify-content:flex-start;align-items:flex-start}.field-icon{display:grid;place-items:center;flex:0 0 40px;width:40px;height:40px;border-radius:14px;background:linear-gradient(180deg,rgba(84,203,230,.14),rgba(84,203,230,.08));color:#2792be;box-shadow:inset 0 1px 0 rgba(255,255,255,.92)}.field-icon svg{width:20px;height:20px;fill:none;stroke:currentColor;stroke-width:1.8;stroke-linecap:round;stroke-linejoin:round}.field-copy strong{color:#41596f;font-size:15px}.field-control{position:relative}.field-glow{position:absolute;right:14px;top:50%;width:10px;height:10px;border-radius:50%;background:rgba(84,208,232,.6);transform:translateY(-50%);box-shadow:0 0 14px rgba(84,208,232,.35);pointer-events:none}.field-input{padding:0 42px 0 16px}.field-input:focus,.jobs-select:focus{border-color:rgba(78,196,225,.8);box-shadow:0 0 0 4px rgba(78,196,225,.12)}.dialog-actions{justify-content:flex-end}.dialog-button{min-height:48px;padding:0 22px;border-radius:16px;font-weight:700}.dialog-button.cancel{background:linear-gradient(180deg,rgba(236,242,247,.96),rgba(222,231,239,.92));color:#90a0b1}.dialog-button.confirm{color:#fff;background:linear-gradient(135deg,#4ed8ef 0%,#2fb5de 55%,#2a9fd3 100%);box-shadow:0 16px 30px rgba(67,188,224,.28)}
.refresh-button:hover,.settings-button:hover,.dialog-button:hover,.dialog-close:hover,.pager-button:hover{transform:translateY(-1px)}:deep(.queue-dialog .el-dialog){border:1px solid rgba(255,255,255,.96);border-radius:32px;background:radial-gradient(circle at top left,rgba(161,222,255,.18),transparent 22%),linear-gradient(180deg,rgba(255,255,255,.98),rgba(252,250,245,.96));box-shadow:0 32px 80px rgba(100,135,163,.18)}:deep(.queue-dialog .el-dialog__header){margin:0;padding:24px 24px 12px}:deep(.queue-dialog .el-dialog__body){padding:12px 24px 20px}:deep(.queue-dialog .el-dialog__footer){padding:0 24px 24px}@keyframes spin{to{transform:rotate(360deg)}}@media (max-width:1280px){.status-grid{grid-template-columns:repeat(3,minmax(0,1fr))}}@media (max-width:1080px){.queue-grid,.dialog-body{grid-template-columns:1fr}.jobs-head{flex-direction:column}}@media (max-width:720px){.queue-shell{padding:18px;border-radius:24px}.queue-hero{align-items:flex-start;flex-direction:column}.status-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.jobs-pagination{justify-content:space-between;width:100%}}
@media (max-width:1280px){.runtime-panel{grid-template-columns:1fr}}
@media (max-width:720px){.runtime-status{grid-template-columns:1fr}.runtime-count{width:32px}}

.queue-settings{position:relative;isolation:isolate;padding:4px 0 28px;color:#193757}.queue-settings::before{content:'';position:fixed;inset:0;z-index:-2;background:radial-gradient(circle at 10% 4%,rgba(112,215,246,.28),transparent 24%),radial-gradient(circle at 84% 12%,rgba(255,187,214,.26),transparent 25%),radial-gradient(circle at 58% 78%,rgba(255,219,161,.2),transparent 30%),linear-gradient(135deg,#f7fbff 0%,#fdf7fb 46%,#f7fbff 100%)}.queue-settings::after{content:'';position:fixed;inset:0;z-index:-1;pointer-events:none;background:linear-gradient(90deg,rgba(255,255,255,.34),transparent 12%,rgba(191,226,241,.18) 50%,transparent 88%),radial-gradient(circle at 45% 20%,rgba(255,255,255,.55),transparent 34%);backdrop-filter:blur(2px)}.queue-shell{z-index:1;max-width:1580px;margin:0 auto;padding:34px 38px;border-radius:34px;border:1px solid rgba(255,255,255,.72);background:linear-gradient(135deg,rgba(255,255,255,.62),rgba(255,246,252,.42) 46%,rgba(236,249,255,.52));box-shadow:inset 0 1px 0 rgba(255,255,255,.9),inset 0 -1px 0 rgba(126,207,236,.14),0 24px 70px rgba(75,122,161,.16);backdrop-filter:blur(22px) saturate(150%)}.queue-shell::before{content:'';position:absolute;inset:14px;border-radius:28px;border:1px solid rgba(255,255,255,.42);pointer-events:none}.queue-shell::after{right:2%;bottom:auto;top:8%;width:520px;height:420px;background:radial-gradient(circle at 35% 35%,rgba(80,213,239,.24),transparent 58%),radial-gradient(circle at 70% 72%,rgba(255,178,215,.2),transparent 62%);filter:blur(22px);opacity:.9}.luxury-line{display:none}.queue-hero{position:relative;z-index:2;margin-bottom:18px;padding:8px 0 4px}.hero-kicker,.dialog-kicker{color:#34a9d5;letter-spacing:.22em}.queue-hero h2{color:#173b5e;text-shadow:0 1px 0 rgba(255,255,255,.8);letter-spacing:0;font-weight:900}.hero-text{color:#6d86a0}.refresh-button,.job-tool-button,.pager-button,.dialog-button,.settings-button,.dialog-close{border:1px solid rgba(255,255,255,.74);background:linear-gradient(180deg,rgba(255,255,255,.86),rgba(244,248,255,.58));color:#173b5e;box-shadow:inset 0 1px 0 rgba(255,255,255,.92),0 12px 28px rgba(92,141,179,.12);backdrop-filter:blur(14px)}.refresh-button,.job-tool-button.accent,.job-action-button.accent,.dialog-button.confirm{color:#fff;background:linear-gradient(135deg,#56d8ee 0%,#34b6dd 54%,#6f8df0 100%);box-shadow:0 16px 34px rgba(70,180,224,.26),inset 0 1px 0 rgba(255,255,255,.48)}.runtime-panel,.queue-card,.jobs-panel{scroll-margin-top:110px;border-radius:30px;border:1px solid rgba(255,255,255,.76);background:linear-gradient(145deg,rgba(255,255,255,.62),rgba(255,246,252,.34) 48%,rgba(240,253,255,.48));box-shadow:inset 0 1px 0 rgba(255,255,255,.92),inset 0 -1px 0 rgba(141,208,232,.16),0 18px 40px rgba(96,130,161,.13);backdrop-filter:blur(18px) saturate(142%)}.runtime-panel::after,.queue-card::after,.jobs-panel::after{content:'';position:absolute;inset:1px;border-radius:inherit;pointer-events:none;background:linear-gradient(135deg,rgba(255,255,255,.54),transparent 34%,rgba(112,215,246,.12) 78%,rgba(255,190,220,.18))}.runtime-panel{grid-template-columns:minmax(0,1.25fr) minmax(390px,.8fr);padding:26px 30px}.runtime-copy h3{font-size:2rem;color:#173b5e;letter-spacing:0}.runtime-status{border-radius:18px;border-color:rgba(219,236,247,.82);background:rgba(255,255,255,.58);box-shadow:inset 0 1px 0 rgba(255,255,255,.82)}.queue-grid{scroll-margin-top:110px;gap:24px}.queue-card{min-height:310px;padding:30px}.queue-icon{width:82px;height:82px;border-radius:26px;background:linear-gradient(145deg,rgba(231,250,255,.84),rgba(255,246,251,.66));box-shadow:inset 0 1px 0 rgba(255,255,255,.92),0 16px 34px rgba(91,171,204,.14)}.queue-icon svg{stroke:#2aa5d3}.title-copy h3,.jobs-copy h3{color:#219dcc;font-size:2.05rem;letter-spacing:0;font-weight:900}.title-copy p{font-size:15px;color:#6e879f}.settings-button{width:50px;height:50px;border-radius:18px}.settings-pulse{background:#4bd7ef;box-shadow:0 0 0 5px rgba(75,215,239,.12),0 0 18px rgba(75,215,239,.85)}.progress-bar{height:12px;background:rgba(212,231,240,.62);box-shadow:inset 0 1px 2px rgba(80,120,150,.08)}.status-card{border:1px solid rgba(255,255,255,.62);border-radius:22px;background:linear-gradient(180deg,rgba(255,255,255,.62),rgba(255,250,244,.42));box-shadow:inset 0 1px 0 rgba(255,255,255,.88),0 12px 28px rgba(106,142,167,.1)}.status-ring{width:66px}.status-ring::before{width:48px;height:48px}.jobs-panel{padding:28px}.jobs-table-shell{border-radius:24px;border-color:rgba(221,238,248,.86);background:rgba(255,255,255,.48);box-shadow:inset 0 1px 0 rgba(255,255,255,.74)}.jobs-table th{background:rgba(247,252,255,.72);color:#5b7894}.jobs-table td{background:rgba(255,255,255,.32)}.jobs-table tr:hover td{background:rgba(238,250,255,.55)}.job-action-button{border:1px solid rgba(255,255,255,.7);box-shadow:inset 0 1px 0 rgba(255,255,255,.8),0 8px 18px rgba(104,142,170,.1)}.job-tool-button.danger,.job-action-button.danger{background:linear-gradient(135deg,#ffb6bd 0%,#f27786 100%)}:deep(.queue-dialog .el-dialog){border-radius:30px;border-color:rgba(255,255,255,.78);background:linear-gradient(145deg,rgba(255,255,255,.82),rgba(255,247,253,.68) 48%,rgba(241,253,255,.72));backdrop-filter:blur(22px) saturate(150%)}@media (max-width:1080px){.queue-shell{padding:22px}.runtime-panel{grid-template-columns:1fr}.queue-card{min-height:auto}}@media (max-width:720px){.queue-card,.runtime-panel,.jobs-panel{padding:20px}.title-copy h3,.jobs-copy h3{font-size:1.55rem}}
</style>



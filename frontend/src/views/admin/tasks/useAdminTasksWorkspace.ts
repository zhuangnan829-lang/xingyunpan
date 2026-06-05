import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import {
  batchDeleteQueueJobs,
  clearQueueJobs,
  deleteQueueJob,
  getQueueJobs,
  getQueueStats,
} from '@/api/queue-settings';
import { listNodes, type NodePayload } from '@/api/nodes';
import {
  queueLabelMap,
  queueOptions,
  statusLabelMap,
  type QueueStats,
  type TaskFilters,
  type TaskJob,
  type TaskMetric,
} from './types';

const defaultStats = (): QueueStats[] => queueOptions.map((queue) => ({
  queue_key: queue.key,
  success: 0,
  failed: 0,
  processing: 0,
  pending: 0,
  submitted: 0,
}));

export function useAdminTasksWorkspace() {
  const loading = ref(false);
  const statsLoading = ref(false);
  const jobs = ref<TaskJob[]>([]);
  const nodes = ref<NodePayload[]>([]);
  const stats = ref<QueueStats[]>(defaultStats());
  const selectedJobIds = ref<number[]>([]);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  const selectedJob = ref<TaskJob | null>(null);
  const detailVisible = ref(false);
  const lastUpdatedAt = ref('');
  const filters = reactive<TaskFilters>({
    queueKey: '',
    status: '',
    nodeId: '',
  });

  const totals = computed(() => stats.value.reduce(
    (acc, item) => {
      acc.submitted += item.submitted;
      acc.processing += item.processing;
      acc.pending += item.pending;
      acc.failed += item.failed;
      acc.success += item.success;
      return acc;
    },
    { submitted: 0, processing: 0, pending: 0, failed: 0, success: 0 },
  ));

  const metrics = computed<TaskMetric[]>(() => [
    { key: 'submitted', label: '任务总数', value: totals.value.submitted, hint: '当前队列累计提交', tone: 'sky' },
    { key: 'processing', label: '处理中', value: totals.value.processing, hint: '正在被 Worker 执行', tone: 'mint' },
    { key: 'pending', label: '等待中', value: totals.value.pending, hint: '已入队等待调度', tone: 'amber' },
    { key: 'failed', label: '失败', value: totals.value.failed, hint: '需要排查或重试', tone: 'rose' },
  ]);

  const queueCards = computed(() => stats.value.map((item) => ({
    ...item,
    label: queueLabelMap[item.queue_key] || item.queue_key,
  })));

  const activeFilterText = computed(() => {
    const queue = filters.queueKey ? queueLabelMap[filters.queueKey] : '全部队列';
    const status = filters.status ? statusLabelMap[filters.status] : '全部状态';
    const node = filters.nodeId ? nodes.value.find((item) => item.id === filters.nodeId)?.name || `节点 #${filters.nodeId}` : '全部节点';
    return `${queue} / ${status} / ${node}`;
  });

  async function loadStats() {
    statsLoading.value = true;
    try {
      const data = await getQueueStats();
      stats.value = mergeStats(data);
    } finally {
      statsLoading.value = false;
    }
  }

  async function loadNodes() {
    nodes.value = await listNodes();
  }

  async function loadJobs(nextPage = page.value) {
    loading.value = true;
    try {
      const data = await getQueueJobs({
        queue_key: filters.queueKey,
        status: filters.status,
        node_id: filters.nodeId,
        page: nextPage,
        page_size: pageSize.value,
      });
      jobs.value = data.list || [];
      selectedJobIds.value = [];
      total.value = data.total || 0;
      page.value = data.page || nextPage;
      pageSize.value = data.page_size || pageSize.value;
      lastUpdatedAt.value = formatTime(new Date());
    } finally {
      loading.value = false;
    }
  }

  async function refresh() {
    try {
      await Promise.all([loadStats(), loadJobs(page.value)]);
      ElMessage.success('后台任务已刷新');
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '刷新失败');
    }
  }

  async function applyFilters() {
    await loadJobs(1);
  }

  async function resetFilters() {
    filters.queueKey = '';
    filters.status = '';
    filters.nodeId = '';
    await loadJobs(1);
  }

  async function clearCurrentView() {
    if (selectedJobIds.value.length === 0 && total.value === 0) {
      ElMessage.info('暂无可清理的后台任务');
      return;
    }

    if (selectedJobIds.value.length === 0 && (filters.status === 'pending' || filters.status === 'processing')) {
      ElMessage.warning('等待中或处理中的任务不能按筛选清理，请选择具体任务或切换到已完成/失败');
      return;
    }

    try {
      if (selectedJobIds.value.length > 0) {
        const result = await batchDeleteQueueJobs(selectedJobIds.value);
        ElMessage.success(`已删除 ${result.deleted} 条后台任务`);
      } else {
        const result = await clearQueueJobs({
          queue_key: filters.queueKey,
          status: filters.status,
        });
        ElMessage.success(`已清理 ${result.deleted} 条已完成/失败任务`);
      }
      await Promise.all([loadStats(), loadJobs(1)]);
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '清理失败，请确认后端服务已重启');
    }
  }

  async function removeJob(job: TaskJob) {
    try {
      const result = await deleteQueueJob(job.id);
      ElMessage.success(`已删除 ${result.deleted} 条后台任务`);
      await Promise.all([loadStats(), loadJobs(page.value)]);
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '删除失败');
    }
  }

  async function changePage(nextPage: number) {
    await loadJobs(nextPage);
  }

  async function changePageSize(nextPageSize: number) {
    pageSize.value = nextPageSize;
    await loadJobs(1);
  }

  function openDetail(job: TaskJob) {
    selectedJob.value = job;
    detailVisible.value = true;
  }

  onMounted(async () => {
    try {
      await loadNodes();
      await Promise.all([loadStats(), loadJobs(1)]);
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '后台任务加载失败');
    }
  });

  return {
    activeFilterText,
    changePage,
    changePageSize,
    clearCurrentView,
    detailVisible,
    filters,
    jobs,
    nodes,
    lastUpdatedAt,
    loading,
    metrics,
    openDetail,
    page,
    pageSize,
    queueCards,
    refresh,
    resetFilters,
    removeJob,
    selectedJob,
    selectedJobIds,
    statsLoading,
    total,
    applyFilters,
  };
}

function mergeStats(data: QueueStats[]): QueueStats[] {
  const byKey = new Map(data.map((item) => [item.queue_key, item]));
  return defaultStats().map((item) => byKey.get(item.queue_key) || item);
}

function formatTime(value: Date): string {
  return value.toLocaleTimeString('zh-CN', { hour12: false });
}

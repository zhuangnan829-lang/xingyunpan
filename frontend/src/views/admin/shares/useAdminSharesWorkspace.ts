import { computed, reactive, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import {
  batchDeleteAdminShares,
  deleteAdminShare,
  getAdminShareMetrics,
  listAdminShares,
  type AdminShareMetricsPayload,
  type AdminSharePayload,
} from '@/api/admin-shares';
import { resolveShareURL } from '@/utils/public-url';
import { copyToClipboard } from '@/utils/share-utils';
import type { ShareDisplayRecord, ShareFilters, ShareMetric } from './types';

export function useAdminSharesWorkspace() {
  const shareRows = ref<AdminSharePayload[]>([]);
  const metricsPayload = ref<AdminShareMetricsPayload | null>(null);
  const isLoading = ref(false);
  const selectedRecords = ref<ShareDisplayRecord[]>([]);
  const filters = reactive<ShareFilters>({
    keyword: '',
    status: 'all',
    minDownloads: null,
    expiringOnly: false,
  });
  const pagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0,
    mode: 'cursor',
    cursor: '',
    nextCursor: '',
    cursorHistory: [] as string[],
  });

  const records = computed<ShareDisplayRecord[]>(() =>
    shareRows.value.map((share, index) => {
      const timeExpired = share.is_expired || share.status_reason === 'time_expired' || isShareExpired(share.expires_at);
      const isUnavailable = share.is_unavailable || timeExpired || share.status_reason === 'download_limit_reached';
      const sourceLabel = share.file_names.length ? share.file_names.join('、') : '未命名分享';
      const score = share.access_count + share.download_count * 3;

      return {
        ...share,
        index: (pagination.page - 1) * pagination.pageSize + index + 1,
        sourceLabel,
        ownerName: share.owner_username || '未知用户',
        statusText: statusTextFor(share.status_reason, isUnavailable),
        expiryText: formatExpiry(share.expires_at),
        createdText: formatDate(share.created_at),
        score,
        isExpired: timeExpired,
        isUnavailable,
        isProtected: share.has_password,
      };
    }),
  );

  const filteredRecords = computed(() => records.value);

  const metrics = computed<ShareMetric[]>(() => {
    const data = metricsPayload.value;
    const total = data?.total_shares ?? 0;
    const active = data?.active_shares ?? 0;
    const visits = data?.total_access_count ?? 0;
    const downloads = data?.total_download_count ?? 0;
    const expiringSoon = data?.expiring_soon_count ?? 0;

    return [
      { label: '分享总量', value: formatNumber(total), detail: `${active} 条仍在生效`, tone: 'blue' },
      { label: '总浏览', value: formatNumber(visits), detail: '所有分享访问累计', tone: 'cyan' },
      { label: '总下载', value: formatNumber(downloads), detail: '下载触达与转化', tone: 'coral' },
      { label: '即将过期', value: formatNumber(expiringSoon), detail: '未来 3 天内失效', tone: 'violet' },
    ];
  });

  const totalPages = computed(() => Math.max(1, Math.ceil(pagination.total / pagination.pageSize)));
  const canGoPrevious = computed(() => pagination.page > 1);
  const canGoNext = computed(() => {
    if (pagination.mode === 'cursor') {
      return Boolean(pagination.nextCursor);
    }
    return pagination.page < totalPages.value;
  });

  async function loadShares(options: { resetCursor?: boolean } = {}) {
    if (options.resetCursor) {
      resetPaginationCursor();
    }

    isLoading.value = true;
    try {
      const data = await listAdminShares({
        page: pagination.page,
        page_size: pagination.pageSize,
        cursor: pagination.cursor || undefined,
        keyword: filters.keyword.trim() || undefined,
        status: filters.status === 'all' ? undefined : filters.status,
        min_downloads: filters.minDownloads ?? undefined,
        expiring_within_days: filters.expiringOnly ? 3 : undefined,
      });
      shareRows.value = data.list;
      pagination.total = data.total;
      pagination.mode = data.pagination_mode || pagination.mode;
      pagination.nextCursor = data.next_cursor || '';
      pagination.pageSize = data.page_size || pagination.pageSize;
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载分享记录失败');
    } finally {
      isLoading.value = false;
    }
  }

  async function loadMetrics() {
    try {
      metricsPayload.value = await getAdminShareMetrics(3);
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '加载分享统计失败');
    }
  }

  async function refreshAll(options: { resetCursor?: boolean } = {}) {
    await Promise.all([loadShares(options), loadMetrics()]);
  }

  async function copyLink(record: ShareDisplayRecord) {
    const success = await copyToClipboard(normalizeShareLink(record));
    ElMessage[success ? 'success' : 'error'](success ? '分享链接已复制' : '复制失败，请手动复制');
  }

  async function deleteRecord(record: ShareDisplayRecord) {
    try {
      await deleteAdminShare(record.share_id);
      shareRows.value = shareRows.value.filter((share) => share.share_id !== record.share_id);
      pagination.total = Math.max(0, pagination.total - 1);
      void loadMetrics();
      ElMessage.success('分享记录已删除');
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '删除分享失败');
    }
  }

  async function deleteSelected() {
    if (!selectedRecords.value.length) return;
    try {
      const shareIds = selectedRecords.value.map((record) => record.share_id);
      await batchDeleteAdminShares(shareIds);
      shareRows.value = shareRows.value.filter((share) => !shareIds.includes(share.share_id));
      pagination.total = Math.max(0, pagination.total - shareIds.length);
      void loadMetrics();
      ElMessage.success(`已删除 ${selectedRecords.value.length} 条分享记录`);
      selectedRecords.value = [];
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : '批量删除失败');
    }
  }

  function updateSelection(records: ShareDisplayRecord[]) {
    selectedRecords.value = records;
  }

  function resetFilters() {
    filters.keyword = '';
    filters.status = 'all';
    filters.minDownloads = null;
    filters.expiringOnly = false;
    void loadShares({ resetCursor: true });
  }

  function goNextPage() {
    if (!canGoNext.value) return;
    if (pagination.mode === 'cursor') {
      pagination.cursorHistory.push(pagination.cursor);
      pagination.cursor = pagination.nextCursor;
    }
    pagination.page += 1;
    void loadShares();
  }

  function goPreviousPage() {
    if (!canGoPrevious.value) return;
    if (pagination.mode === 'cursor') {
      pagination.cursor = pagination.cursorHistory.pop() || '';
    }
    pagination.page -= 1;
    void loadShares();
  }

  let filterTimer: ReturnType<typeof window.setTimeout> | undefined;
  watch(
    () => [filters.keyword, filters.status, filters.minDownloads, filters.expiringOnly] as const,
    () => {
      if (filterTimer) {
        window.clearTimeout(filterTimer);
      }
      filterTimer = window.setTimeout(() => {
        void loadShares({ resetCursor: true });
      }, 260);
    },
  );

  function resetPaginationCursor() {
    pagination.page = 1;
    pagination.cursor = '';
    pagination.nextCursor = '';
    pagination.cursorHistory = [];
  }

  return {
    filters,
    filteredRecords,
    isLoading: computed(() => isLoading.value),
    metrics,
    pagination,
    selectedRecords,
    canGoNext,
    canGoPrevious,
    copyLink,
    deleteRecord,
    deleteSelected,
    goNextPage,
    goPreviousPage,
    loadShares: refreshAll,
    resetFilters,
    updateSelection,
  };
}

function normalizeShareLink(share: AdminSharePayload) {
  return resolveShareURL(share.share_url, share.share_token);
}

function statusTextFor(reason: string, unavailable: boolean) {
  switch (reason) {
    case 'time_expired':
      return '已过期';
    case 'download_limit_reached':
      return '达上限';
    case 'active':
      return '有效';
    default:
      return unavailable ? '已失效' : '有效';
  }
}

function formatDate(value: string) {
  if (!value) return '--';
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}

function formatExpiry(value: string | null) {
  return value ? formatDate(value) : '永久有效';
}

function isShareExpired(value: string | null) {
  return !!value && new Date(value).getTime() <= Date.now();
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('zh-CN', {
    notation: value >= 10000 ? 'compact' : 'standard',
  }).format(value);
}

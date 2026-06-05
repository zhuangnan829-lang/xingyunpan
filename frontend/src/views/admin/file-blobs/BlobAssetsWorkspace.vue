<template>
  <section class="blob-page">
    <div class="blob-shell">
      <header class="page-header">
        <div class="title-block">
          <span class="page-kicker">Xingyunpan V2 / Files Blob</span>
          <div class="title-row">
            <h1>文件 Blob</h1>
            <p>把底层二进制数据块整理成可巡检、可追踪、可控权的智慧资产库。</p>
          </div>
        </div>

        <div class="header-actions">
          <button class="toolbar-button primary" type="button" :disabled="loading" @click="reloadPage">
            <el-icon :class="{ spinning: loading }"><RefreshRight /></el-icon>
            <span>刷新</span>
          </button>
          <button class="toolbar-button" type="button" :disabled="scanLoading" @click="runScan">
            <el-icon :class="{ spinning: scanLoading }"><Search /></el-icon>
            <span>巡检 Blob</span>
          </button>
          <button class="toolbar-button danger-lite" type="button" :disabled="batchDeleting || !deletableOrphanIds.length" @click="batchDeleteOrphans">
            <el-icon><Delete /></el-icon>
            <span>批量删除孤儿</span>
          </button>
          <button class="toolbar-button" type="button" @click="resetFilters">
            <el-icon><RefreshLeft /></el-icon>
            <span>重置筛选</span>
          </button>
        </div>      </header>

      <section class="stats-grid">
        <article class="stat-card">
          <span class="stat-label">Blob 总数</span>
          <strong>{{ stats.total }}</strong>
          <small>当前后端筛选结果中的资产记录</small>
        </article>
        <article class="stat-card">
          <span class="stat-label">资产体积</span>
          <strong>{{ formatSize(stats.totalSize) }}</strong>
          <small>已纳入管理的数据块体积</small>
        </article>
        <article class="stat-card">
          <span class="stat-label">引用计数</span>
          <strong>{{ stats.referenceTotal }}</strong>
          <small>文件、版本、缓存与任务引用总量</small>
        </article>
        <article class="stat-card">
          <span class="stat-label">加密 / 孤儿</span>
          <strong>{{ stats.encryptedCount }} / {{ stats.orphanCount }}</strong>
          <small>加密 Blob 与孤儿 Blob 数量</small>
        </article>
      </section>
      <section class="filter-bar">
        <label class="filter-field wide">
          <span>关键词</span>
          <input v-model.trim="filters.keyword" class="field-input" type="search" placeholder="路径 / hash / MIME / 文件名" @keyup.enter="applyFilters" />
        </label>

        <label class="filter-field compact">
          <span>创建者 ID</span>
          <input v-model.trim="filters.ownerId" class="field-input" type="text" placeholder="例如 1" @keyup.enter="applyFilters" />
        </label>

        <label class="filter-field">
          <span>类型</span>
          <select v-model="filters.kind" class="field-input">
            <option value="all">全部</option>
            <option value="file">文件 Blob</option>
            <option value="thumbnail">缩略图缓存</option>
            <option value="version">版本 Blob</option>
            <option value="live-photo">Live Photo</option>
            <option value="orphan">孤儿 Blob</option>
          </select>
        </label>

        <label class="filter-field">
          <span>存储策略</span>
          <select v-model="filters.storagePolicyId" class="field-input">
            <option value="all">全部</option>
            <option v-for="policy in policyOptions" :key="policy.id" :value="policy.id">
              {{ policy.name }}
            </option>
          </select>
        </label>

        <label class="filter-field compact">
          <span>引用数下限</span>
          <input v-model.trim="filters.refCountMin" class="field-input" type="number" min="0" placeholder="min" />
        </label>
        <label class="filter-field compact">
          <span>引用数上限</span>
          <input v-model.trim="filters.refCountMax" class="field-input" type="number" min="0" placeholder="max" />
        </label>
        <label class="filter-field compact">
          <span>大小下限(B)</span>
          <input v-model.trim="filters.minSize" class="field-input" type="number" min="0" placeholder="min" />
        </label>
        <label class="filter-field compact">
          <span>大小上限(B)</span>
          <input v-model.trim="filters.maxSize" class="field-input" type="number" min="0" placeholder="max" />
        </label>

        <label class="filter-field">
          <span>加密状态</span>
          <select v-model="filters.encrypted" class="field-input">
            <option value="all">全部</option>
            <option value="true">已加密</option>
            <option value="false">未加密</option>
          </select>
        </label>

        <label class="filter-field">
          <span>创建时间起</span>
          <input v-model="filters.createdFrom" class="field-input" type="date" />
        </label>
        <label class="filter-field">
          <span>创建时间止</span>
          <input v-model="filters.createdTo" class="field-input" type="date" />
        </label>

        <div class="filter-actions">
          <button class="toolbar-button primary" type="button" :disabled="loading" @click="applyFilters">应用筛选</button>
          <button class="toolbar-button" type="button" @click="resetFilters">清空</button>
        </div>
      </section>
      <section class="table-card">
        <div class="table-toolbar">
          <div class="table-title">
            <strong>Blob 智慧资产库</strong>
            <small>底层源路径、策略、引用计数和生命周期状态保持同屏可见。</small>
          </div>
          <div class="table-meta">
            <span>{{ stats.total }} 条</span>
            <span>{{ formatSize(stats.totalSize) }}</span>
            <span v-if="latestScan">最近巡检：{{ latestScan.status }} / {{ latestScan.progress }}%</span>
          </div>
        </div>

        <div v-if="loading" class="state-shell">正在加载 Blob 数据...</div>
        <div v-else-if="!rows.length" class="state-shell">当前筛选条件下没有 Blob。</div>

        <div v-else class="list-shell">
          <div class="table-head">
            <button class="col-id sortable" type="button" @click="setSort('id')">
              <span>#</span>
              <el-icon class="sort-icon" :class="{ reverse: sortOrder === 'asc' }"><Sort /></el-icon>
            </button>
            <div class="col-type">Type</div>
            <div class="col-source">Source</div>
            <div class="col-size">Size</div>
            <div class="col-policy">Policy</div>
            <div class="col-ref">Reference Count</div>
            <div class="col-created">Time</div>
            <div class="col-actions">Tools</div>
          </div>

          <article v-for="blob in rows" :key="blob.id" class="blob-row">
            <div class="col-id row-id">#{{ blob.id }}</div>

            <div class="col-type">
              <div class="type-cell">
                <span class="blob-glyph" :class="`kind-${blob.kind}`">
                  <el-icon v-if="blob.kind === 'live-photo'"><VideoCameraFilled /></el-icon>
                  <span v-else-if="blob.kind === 'thumbnail'" class="thumb-preview"></span>
                  <el-icon v-else-if="blob.kind === 'version'"><Collection /></el-icon>
                  <el-icon v-else><Files /></el-icon>
                </span>
                <div class="type-copy">
                  <strong>{{ blob.kindLabel }}</strong>
                  <small>{{ blob.encrypted ? '安全锁已启用' : '标准存储' }}</small>
                </div>
              </div>
            </div>

            <div class="col-source">
              <button class="source-button" type="button" @click="showMetadata(blob)">
                {{ blob.source }}
              </button>
            </div>

            <div class="col-size">
              <span class="size-chip">{{ formatSize(blob.sizeBytes) }}</span>
            </div>

            <div class="col-policy">
              <div class="policy-copy">
                <strong>{{ blob.storagePolicyName || '未知策略' }}</strong>
                <small>{{ blob.storagePolicySubtitle || 'local' }}</small>
              </div>
            </div>

            <div class="col-ref">
              <span class="ref-pill">{{ blob.referenceCount }}</span>
            </div>

            <div class="col-created">
              <div class="created-copy">
                <strong>{{ formatDate(blob.createdAt) }}</strong>
                <small>{{ blob.creatorName || 'system' }} #{{ blob.creatorId || 0 }}</small>
              </div>
            </div>

            <div class="col-actions">
              <button class="action-button" type="button" :title="blob.locked ? '解锁 Blob' : '锁定 Blob'" @click="toggleBlobLock(blob)">
                <el-icon><Unlock v-if="blob.locked" /><Lock v-else /></el-icon>
              </button>
              <button class="action-button" type="button" title="元数据" @click="showMetadata(blob)">
                <el-icon><Document /></el-icon>
              </button>
              <button class="action-button danger" type="button" :disabled="!blob.canDelete" :title="deleteTooltip(blob)" @click="removeBlob(blob)">
                <el-icon><Delete /></el-icon>
              </button>
            </div>
          </article>
        </div>
      </section>

      <footer class="footer-bar">
        <div class="pager">
          <button class="pager-button" type="button" :disabled="page <= 1 || loading" @click="goPage(page - 1)">
            <el-icon><ArrowLeft /></el-icon>
          </button>
          <span class="pager-current">{{ page }} / {{ totalPages }}</span>
          <button class="pager-button" type="button" :disabled="page >= totalPages || loading" @click="goPage(page + 1)">
            <el-icon><ArrowRight /></el-icon>
          </button>
        </div>

        <select v-model.number="pageSize" class="page-size-select" @change="changePageSize">
          <option :value="10">每页 10 条</option>
          <option :value="25">每页 25 条</option>
          <option :value="50">每页 50 条</option>
          <option :value="100">每页 100 条</option>
        </select>
      </footer>
    </div>

    <el-dialog
      :model-value="detailVisible"
      width="860px"
      append-to-body
      destroy-on-close
      @close="closeDetail"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-title">
            <h2>{{ activeBlob?.kindLabel || 'Blob 详情' }}</h2>
            <small>{{ activeBlob ? `Blob #${activeBlob.id}` : 'Blob 详情' }}</small>
          </div>
        </div>
      </template>

      <div v-if="activeBlob" class="dialog-body">
        <section class="detail-grid">
          <article class="detail-card">
            <span>底层源路径</span>
            <strong class="wrap">{{ activeBlob.source }}</strong>
          </article>
          <article class="detail-card">
            <span>Hash</span>
            <strong class="wrap">{{ activeBlob.hash || '-' }}</strong>
          </article>
          <article class="detail-card">
            <span>MIME</span>
            <strong>{{ activeBlob.contentType || '-' }}</strong>
          </article>
          <article class="detail-card">
            <span>健康状态</span>
            <strong>{{ healthLabel(activeBlob.healthStatus) }}</strong>
          </article>
          <article class="detail-card">
            <span>是否可删除</span>
            <strong>{{ activeBlob.canDelete ? '可删除' : (activeBlob.deleteBlockedReasons?.join('；') || '不可删除') }}</strong>
          </article>
          <article class="detail-card">
            <span>大小</span>
            <strong>{{ formatSize(activeBlob.sizeBytes) }}</strong>
          </article>
          <article class="detail-card">
            <span>策略</span>
            <strong>{{ activeBlob.storagePolicyName || '未知策略' }}</strong>
          </article>
          <article class="detail-card">
            <span>引用计数</span>
            <strong>{{ activeBlob.referenceCount }}</strong>
          </article>
          <article class="detail-card">
            <span>创建者</span>
            <strong>{{ activeBlob.creatorName || 'system' }} #{{ activeBlob.creatorId || 0 }}</strong>
          </article>
          <article class="detail-card">
            <span>时间</span>
            <strong>{{ formatDate(activeBlob.createdAt) }}</strong>
          </article>
          <article class="detail-card">
            <span>安全锁</span>
            <strong>{{ activeBlob.encrypted ? '已加密' : '未加密' }}</strong>
          </article>
          <article class="detail-card">
            <span>上传会话</span>
            <strong>{{ activeBlob.uploadSessionId || '-' }}</strong>
          </article>
        </section>

        <section class="linked-panel">
          <div class="linked-title">
            <strong>引用来源</strong>
            <small>{{ activeBlob.referenceSources?.length || 0 }} 条</small>
          </div>

          <div v-if="activeBlob.referenceSources?.length" class="linked-list">
            <article v-for="source in activeBlob.referenceSources" :key="`${source.type}-${source.id}`" class="linked-item">
              <span class="linked-badge">{{ source.type }}</span>
              <div class="linked-copy">
                <strong>{{ source.name || source.id }}</strong>
                <small>{{ source.type }} / {{ source.id }}</small>
              </div>
            </article>
          </div>

          <div v-if="!activeBlob.linkedFiles.length" class="empty-linked">没有关联文件，可能是孤儿 Blob。</div>

          <div v-else class="linked-list">
            <article v-for="file in activeBlob.linkedFiles" :key="`${file.id}-${file.name}`" class="linked-item">
              <span class="linked-badge">{{ file.extension || 'bin' }}</span>
              <div class="linked-copy">
                <strong>{{ file.name }}</strong>
                <small>#{{ file.id }} / {{ file.ownerName }} / {{ formatSize(file.sizeBytes) }} / {{ formatDate(file.createdAt) }}</small>
              </div>
            </article>
          </div>
        </section>
      </div>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  ArrowLeft,
  ArrowRight,
  Collection,
  Delete,
  Document,
  Files,
  Lock,
  RefreshLeft,
  RefreshRight,
  Search,
  Sort,
  Unlock,
  VideoCameraFilled,
} from '@element-plus/icons-vue';
import {
  batchDeleteAdminBlobs,
  deleteAdminBlob,
  getAdminBlob,
  getLatestAdminBlobScan,
  listAdminBlobsPage,
  lockAdminBlob,
  scanAdminBlobs,
  unlockAdminBlob,
  type AdminBlobListParams,
  type AdminBlobScanTask,
} from '@/api/admin-blobs';
import { listStoragePolicies } from '@/api/storage-policy';
import type { BlobFilterState, BlobRecord } from './types';

const detailVisible = ref(false);
const activeBlob = ref<BlobRecord | null>(null);
const loading = ref(false);
const scanLoading = ref(false);
const batchDeleting = ref(false);
const rows = ref<BlobRecord[]>([]);
const page = ref(1);
const pageSize = ref(10);
const sortBy = ref<'id' | 'size' | 'reference_count' | 'created_at'>('id');
const sortOrder = ref<'desc' | 'asc'>('desc');
const total = ref(0);
const stats = ref({
  total: 0,
  totalSize: 0,
  referenceTotal: 0,
  encryptedCount: 0,
  orphanCount: 0,
});
const latestScan = ref<AdminBlobScanTask | null>(null);
const policyOptions = ref<Array<{ id: number; name: string }>>([]);

const createDefaultFilters = (): BlobFilterState => ({
  ownerId: '',
  kind: 'all',
  storagePolicyId: 'all',
  keyword: '',
  minSize: '',
  maxSize: '',
  refCountMin: '',
  refCountMax: '',
  encrypted: 'all',
  createdFrom: '',
  createdTo: '',
});

const filters = ref<BlobFilterState>(createDefaultFilters());

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)));
const deletableOrphanIds = computed(() => rows.value.filter((blob) => blob.kind === 'orphan' && blob.canDelete).map((blob) => blob.id));

function formatSize(value: number): string {
  if (!value) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = value;
  let index = 0;
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024;
    index += 1;
  }
  const digits = size >= 10 || index === 0 ? 0 : 1;
  return `${size.toFixed(digits)} ${units[index]}`;
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
}

async function fetchRows() {
  loading.value = true;
  try {
    const data = await listAdminBlobsPage(buildListParams());
    rows.value = data.list;
    total.value = data.total;
    page.value = data.page;
    pageSize.value = data.pageSize;
    stats.value = {
      total: data.total,
      totalSize: data.totalSize,
      referenceTotal: data.referenceTotal,
      encryptedCount: data.encryptedCount,
      orphanCount: data.orphanCount,
    };
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载 Blob 列表失败');
  } finally {
    loading.value = false;
  }
}

async function reloadPage() {
  await fetchRows();
  ElMessage.success('Blob 列表已刷新');
}

function buildListParams(): AdminBlobListParams {
  const params: AdminBlobListParams = {
    page: page.value,
    page_size: pageSize.value,
    kind: filters.value.kind,
    sort_by: sortBy.value,
    sort_order: sortOrder.value,
  };
  const ownerID = parseOptionalNumber(filters.value.ownerId);
  const storagePolicyID = filters.value.storagePolicyId === 'all' ? undefined : parseOptionalNumber(String(filters.value.storagePolicyId));
  const minSize = parseOptionalNumber(filters.value.minSize);
  const maxSize = parseOptionalNumber(filters.value.maxSize);
  const refCountMin = parseOptionalNumber(filters.value.refCountMin);
  const refCountMax = parseOptionalNumber(filters.value.refCountMax);
  if (filters.value.keyword) params.keyword = filters.value.keyword;
  if (ownerID !== undefined) params.owner_id = ownerID;
  if (storagePolicyID !== undefined) params.storage_policy_id = storagePolicyID;
  if (minSize !== undefined) params.min_size = minSize;
  if (maxSize !== undefined) params.max_size = maxSize;
  if (refCountMin !== undefined) params.ref_count_min = refCountMin;
  if (refCountMax !== undefined) params.ref_count_max = refCountMax;
  if (filters.value.encrypted !== 'all') params.encrypted = filters.value.encrypted === 'true';
  if (filters.value.createdFrom) params.created_from = filters.value.createdFrom;
  if (filters.value.createdTo) params.created_to = filters.value.createdTo;
  return params;
}

function parseOptionalNumber(value: string): number | undefined {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  return Number.isFinite(parsed) ? parsed : undefined;
}

async function applyFilters() {
  page.value = 1;
  await fetchRows();
}

function resetFilters() {
  filters.value = createDefaultFilters();
  page.value = 1;
  void fetchRows();
}

async function setSort(nextSortBy: typeof sortBy.value) {
  if (sortBy.value === nextSortBy) {
    sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc';
  } else {
    sortBy.value = nextSortBy;
    sortOrder.value = 'desc';
  }
  await fetchRows();
}

async function goPage(nextPage: number) {
  page.value = Math.min(Math.max(1, nextPage), totalPages.value);
  await fetchRows();
}

async function changePageSize() {
  page.value = 1;
  await fetchRows();
}

function closeDetail() {
  detailVisible.value = false;
  activeBlob.value = null;
}

async function toggleBlobLock(blob: BlobRecord) {
  try {
    let updated: BlobRecord;
    if (blob.locked) {
      await ElMessageBox.confirm(`确认解锁 Blob #${blob.id} 吗？`, '解锁 Blob', {
        type: 'warning',
        confirmButtonText: '解锁',
        cancelButtonText: '取消',
      });
      updated = await unlockAdminBlob(blob.id);
      ElMessage.success('Blob 已解锁');
    } else {
      const result = await ElMessageBox.prompt('请输入锁定原因（可选）', `锁定 Blob #${blob.id}`, {
        confirmButtonText: '锁定',
        cancelButtonText: '取消',
      });
      updated = await lockAdminBlob(blob.id, String(result.value || ''));
      ElMessage.success('Blob 已锁定');
    }

    const index = rows.value.findIndex((item) => item.id === updated.id);
    if (index >= 0) {
      rows.value[index] = updated;
    }
    if (activeBlob.value?.id === updated.id) {
      activeBlob.value = updated;
    }
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : 'Blob 锁定状态更新失败');
  }
}

function deleteTooltip(blob: BlobRecord): string {
  if (blob.canDelete) return '删除孤儿 Blob';
  return blob.deleteBlockedReasons?.join('；') || '仍有引用或已锁定，不能删除';
}

function healthLabel(status?: string): string {
  switch (status) {
    case 'missing':
      return '存储缺失';
    case 'ref_mismatch':
      return '引用不一致';
    case 'ok':
      return '正常';
    default:
      return status || '-';
  }
}

async function showMetadata(blob: BlobRecord) {
  try {
    activeBlob.value = await getAdminBlob(blob.id);
    detailVisible.value = true;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载 Blob 详情失败');
  }
}

async function removeBlob(blob: BlobRecord) {
  if (!blob.canDelete) {
    ElMessage.warning(deleteTooltip(blob));
    return;
  }
  try {
    await ElMessageBox.confirm(`确认删除 Blob #${blob.id} 吗？只有未被引用的孤儿 Blob 才能删除。`, '删除 Blob', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    });
    await deleteAdminBlob(blob.id);
    if (activeBlob.value?.id === blob.id) {
      closeDetail();
    }
    await fetchRows();
    ElMessage.success('Blob 已删除');
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : '删除 Blob 失败');
  }
}

async function batchDeleteOrphans() {
  if (!deletableOrphanIds.value.length) {
    ElMessage.warning('当前页没有可删除的孤儿 Blob');
    return;
  }
  try {
    await ElMessageBox.confirm(`确认批量删除当前页 ${deletableOrphanIds.value.length} 个可删除孤儿 Blob 吗？`, '批量删除孤儿 Blob', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    });
    batchDeleting.value = true;
    const result = await batchDeleteAdminBlobs(deletableOrphanIds.value);
    await fetchRows();
    ElMessage.success(`已删除 ${result.deleted.length} 个，跳过 ${result.skipped.length} 个，失败 ${result.failed.length} 个`);
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : '批量删除失败');
  } finally {
    batchDeleting.value = false;
  }
}

async function runScan() {
  scanLoading.value = true;
  try {
    latestScan.value = await scanAdminBlobs();
    ElMessage.success(`巡检完成：孤儿 ${latestScan.value.orphan_count}，缺失 ${latestScan.value.missing_on_storage}，引用不一致 ${latestScan.value.ref_count_mismatch}`);
    await fetchRows();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Blob 巡检失败');
  } finally {
    scanLoading.value = false;
  }
}

async function loadLatestScan() {
  try {
    latestScan.value = await getLatestAdminBlobScan();
  } catch {
    latestScan.value = null;
  }
}

async function loadPolicies() {
  try {
    const policies = await listStoragePolicies();
    policyOptions.value = policies
      .filter((policy) => typeof policy.id === 'number')
      .map((policy) => ({ id: Number(policy.id), name: policy.name }));
  } catch {
    policyOptions.value = [];
  }
}

onMounted(async () => {
  await Promise.all([loadPolicies(), loadLatestScan()]);
  await fetchRows();
});
</script>

<style scoped>
.blob-page {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  min-height: calc(100vh - 96px);
  padding: 16px;
  overflow-x: hidden;
  color: #122033;
}

.blob-shell {
  position: relative;
  display: grid;
  gap: 18px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  min-height: calc(100vh - 128px);
  overflow: visible;
  padding: 24px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 28px;
  background:
    radial-gradient(circle at 12% 8%, rgba(125, 223, 232, 0.48), transparent 28%),
    radial-gradient(circle at 84% 4%, rgba(255, 178, 165, 0.55), transparent 30%),
    linear-gradient(135deg, #f9ffff 0%, #fff7f3 52%, #f8fbff 100%);
  box-shadow: 0 28px 72px rgba(100, 116, 139, 0.16);
}

.blob-shell::before {
  position: absolute;
  inset: 0;
  pointer-events: none;
  content: '';
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.34) 1px, transparent 1px),
    linear-gradient(180deg, rgba(255, 255, 255, 0.24) 1px, transparent 1px);
  background-size: 42px 42px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.72), transparent 72%);
}

.page-header,
.title-row,
.header-actions,
.stats-grid,
.filter-bar,
.table-toolbar,
.table-meta,
.footer-bar,
.pager,
.col-actions {
  display: flex;
}

.page-header,
.stats-grid,
.filter-bar,
.table-card,
.footer-bar {
  position: relative;
  z-index: 1;
}

.page-header {
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.title-block {
  display: grid;
  gap: 8px;
  flex: 1 1 560px;
  min-width: 0;
}

.page-kicker {
  color: #2f6f91;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.title-row {
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
  min-width: 0;
}

.title-row h1 {
  margin: 0;
  color: #0f172a;
  font-size: 52px;
  line-height: 1;
  font-weight: 900;
}

.title-row p {
  min-width: 0;
  max-width: 680px;
  margin: 0 0 6px;
  color: #526273;
  font-size: 18px;
}

.header-actions,
.table-meta,
.pager,
.col-actions {
  align-items: center;
  gap: 10px;
}

.header-actions {
  flex: 0 1 auto;
  flex-wrap: wrap;
  justify-content: flex-end;
  max-width: 100%;
  min-width: 0;
}

.toolbar-button,
.pager-button,
.page-size-select {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 42px;
  padding: 0 16px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.62);
  color: #213045;
  font-size: 14px;
  font-weight: 800;
  white-space: nowrap;
  box-shadow: 0 12px 28px rgba(116, 134, 156, 0.12);
  backdrop-filter: blur(18px);
  cursor: pointer;
}

.toolbar-button.primary {
  border-color: rgba(89, 158, 203, 0.45);
  background: linear-gradient(135deg, rgba(79, 190, 203, 0.94), rgba(255, 146, 135, 0.9));
  color: #fff;
}

.toolbar-button.danger-lite {
  border-color: rgba(244, 114, 100, 0.42);
  color: #9f2a20;
}

.toolbar-button:disabled,
.pager-button:disabled {
  opacity: 0.52;
  cursor: not-allowed;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  min-width: 0;
}

.stat-card {
  display: grid;
  gap: 8px;
  min-width: 0;
  padding: 18px 20px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.55);
  box-shadow: 0 18px 46px rgba(100, 116, 139, 0.12);
  backdrop-filter: blur(18px);
}

.stat-label {
  color: #627083;
  font-size: 13px;
  font-weight: 800;
}

.stat-card strong {
  color: #0f172a;
  font-size: 34px;
  line-height: 1.1;
  font-weight: 900;
}

.stat-card small {
  color: #65758a;
  font-size: 13px;
}

.filter-bar {
  align-items: end;
  gap: 14px;
  flex-wrap: wrap;
  min-width: 0;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.74);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.44);
  backdrop-filter: blur(18px);
}

.filter-field {
  min-width: 220px;
  display: grid;
  gap: 8px;
}

.filter-field.wide {
  min-width: min(420px, 100%);
  flex: 1 1 360px;
}

.filter-field.compact {
  min-width: 150px;
  flex: 0 1 170px;
}

.filter-actions {
  display: flex;
  align-items: end;
  gap: 10px;
  min-height: 68px;
}

.filter-field span {
  color: #334155;
  font-size: 13px;
  font-weight: 800;
}

.field-input,
.page-size-select {
  min-height: 42px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.7);
  color: #0f172a;
  font-size: 14px;
  outline: none;
}

.table-card {
  --blob-table-min-width: 1480px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  overflow: visible;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.43);
  box-shadow: 0 24px 60px rgba(100, 116, 139, 0.14);
  backdrop-filter: blur(22px);
}

.table-toolbar {
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-width: 0;
  flex-wrap: wrap;
  padding: 18px 20px;
}

.table-title {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.table-meta {
  min-width: 0;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.table-meta span {
  white-space: nowrap;
}

.table-title strong {
  color: #0f172a;
  font-size: 18px;
}

.table-title small,
.table-meta span,
.type-copy small,
.policy-copy small,
.created-copy small,
.empty-linked,
.linked-copy small,
.dialog-title small,
.detail-card span {
  color: #64748b;
  font-size: 13px;
}

.state-shell {
  display: grid;
  place-items: center;
  min-height: 220px;
  color: #64748b;
}

.list-shell {
  width: 100%;
  box-sizing: border-box;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 0 14px 20px;
}

.table-head,
.blob-row {
  display: grid;
  box-sizing: border-box;
  width: max(100%, var(--blob-table-min-width));
  min-width: var(--blob-table-min-width);
  grid-template-columns: 82px 206px minmax(360px, 1fr) 118px 188px 142px 208px 148px;
  gap: 14px;
  align-items: center;
}

.col-created,
.col-actions {
  position: static;
  min-width: 0;
}

.table-head {
  padding: 12px 18px;
  color: #526173;
  font-size: 12px;
  font-weight: 900;
  text-transform: uppercase;
}

.blob-row {
  min-height: 76px;
  margin-top: 10px;
  padding: 14px 18px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.56);
  box-shadow: 0 12px 34px rgba(100, 116, 139, 0.12), inset 0 0 0 1px rgba(255, 255, 255, 0.42);
  backdrop-filter: blur(18px);
}

.sortable {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.sort-icon {
  transition: transform 0.2s ease;
}

.sort-icon.reverse {
  transform: rotate(180deg);
}

.row-id {
  color: #0f172a;
  font-weight: 900;
}

.type-cell {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
}

.blob-glyph {
  position: relative;
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 14px;
  color: #fff;
  box-shadow: 0 12px 24px rgba(107, 114, 128, 0.18);
}

.kind-file {
  background: linear-gradient(135deg, #66c7e5, #6b8cff);
}

.kind-thumbnail {
  background: linear-gradient(135deg, #cbf3ff, #ffb6ad);
}

.kind-version {
  background: linear-gradient(135deg, #9c8cff, #7bd5e8);
}

.kind-live-photo {
  background: linear-gradient(135deg, #c7a7ff, #ff9eb3);
}

.thumb-preview {
  width: 28px;
  height: 22px;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 7px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.18)),
    linear-gradient(145deg, #78d6ed 0 38%, #ffc4b7 39% 66%, #8f98ff 67%);
  box-shadow: 0 6px 12px rgba(71, 85, 105, 0.16);
}

.type-copy,
.created-copy,
.policy-copy,
.linked-copy {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.type-copy strong,
.policy-copy strong,
.created-copy strong,
.linked-copy strong,
.detail-card strong,
.dialog-title h2 {
  color: #0f172a;
}

.source-button {
  max-width: 100%;
  overflow: hidden;
  padding: 0;
  border: 0;
  background: transparent;
  color: #237a9d;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.size-chip,
.ref-pill {
  display: inline-grid;
  place-items: center;
  min-width: 44px;
  min-height: 32px;
  padding: 0 12px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.62);
  color: #25425f;
  font-size: 13px;
  font-weight: 900;
}

.ref-pill {
  color: #206985;
  background: rgba(214, 249, 255, 0.66);
}

.action-button {
  display: inline-grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border: 1px solid rgba(255, 255, 255, 0.74);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.52);
  color: #31546d;
  box-shadow: 0 10px 20px rgba(100, 116, 139, 0.12);
  backdrop-filter: blur(14px);
  cursor: pointer;
  transition: transform 0.18s ease, background 0.18s ease, color 0.18s ease;
}

.action-button:disabled {
  opacity: 0.42;
  cursor: not-allowed;
  transform: none;
}

.action-button:hover {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.82);
  color: #0f6b87;
}

.action-button.danger {
  color: #b42318;
}

.footer-bar {
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  width: 100%;
  box-sizing: border-box;
  padding: 0 4px;
}

.pager-current {
  display: grid;
  place-items: center;
  min-width: 72px;
  height: 42px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.58);
  color: #1f5674;
  font-size: 14px;
  font-weight: 900;
  backdrop-filter: blur(14px);
}

.dialog-header,
.dialog-title,
.detail-grid,
.dialog-body,
.linked-panel,
.linked-list,
.linked-item {
  display: grid;
}

.dialog-title {
  gap: 4px;
}

.dialog-title h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 900;
}

.dialog-body {
  gap: 18px;
}

.detail-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.detail-card,
.linked-panel {
  padding: 18px;
  border: 1px solid #e4eaf4;
  border-radius: 18px;
  background: #fff;
}

.detail-card {
  gap: 8px;
}

.wrap {
  word-break: break-all;
}

.linked-panel {
  gap: 14px;
}

.linked-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.linked-list {
  gap: 10px;
}

.linked-item {
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  padding-top: 10px;
  border-top: 1px solid #edf2f7;
}

.linked-item:first-child {
  padding-top: 0;
  border-top: 0;
}

.linked-badge {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: #eef4ff;
  color: #2563eb;
  font-size: 13px;
  font-weight: 800;
}

.spinning {
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1440px) {
  .table-head,
  .blob-row {
    min-width: var(--blob-table-min-width);
  }
}

@media (max-width: 960px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .title-row h1 {
    font-size: 40px;
  }

  .title-row p {
    font-size: 16px;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .blob-page {
    padding: 8px;
  }

  .blob-shell {
    padding: 16px;
  }

  .stats-grid,
  .filter-bar {
    display: grid;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .filter-field {
    min-width: 0;
  }
}
</style>




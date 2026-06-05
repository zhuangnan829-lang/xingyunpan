<template>
  <main class="recycle-workspace">
    <section class="recycle-shell">
      <div class="recycle-pathbar">
        <div class="path-title">
          <el-icon><DeleteFilled /></el-icon>
          <strong>回收站</strong>
        </div>

        <div class="path-actions">
          <button class="tool-button" type="button" title="刷新" :disabled="recycleStore.isLoading" @click.stop="reloadCurrentPage">
            <el-icon><Refresh /></el-icon>
          </button>

          <div class="view-control">
            <button class="tool-button labeled" type="button" @click.stop="toggleViewPanel">
              <el-icon><component :is="viewModeIcon" /></el-icon>
              <span>视图</span>
            </button>

            <div v-if="viewPanelVisible" class="view-panel" @click.stop>
              <p class="view-panel-title">布局</p>
              <div class="view-segment">
                <button
                  v-for="option in viewOptions"
                  :key="option.mode"
                  class="view-option"
                  :class="{ active: activeViewMode === option.mode }"
                  type="button"
                  @click="setViewMode(option.mode)"
                >
                  <el-icon><component :is="option.icon" /></el-icon>
                  <span>{{ option.label }}</span>
                </button>
              </div>

              <template v-if="activeViewMode === 'table'">
                <p class="view-panel-title">列设置</p>
                <button class="panel-wide-button" type="button" @click="columnDialogVisible = true">
                  <el-icon><Setting /></el-icon>
                  <span>列设置</span>
                </button>
              </template>

              <template v-if="activeViewMode === 'grid'">
                <label class="slider-field">
                  <span>卡片大小</span>
                  <input v-model.number="gridItemSize" type="range" min="240" max="340" step="10" />
                  <small><span>240</span><span>340</span></small>
                </label>
              </template>

              <template v-if="activeViewMode === 'gallery'">
                <label class="slider-field">
                  <span>图片尺寸</span>
                  <input v-model.number="galleryImageSize" type="range" min="90" max="260" step="10" />
                  <small><span>90</span><span>260</span></small>
                </label>
              </template>

              <label class="slider-field">
                <span>分页大小</span>
                <input v-model.number="pageSizeDraft" type="range" min="10" max="100" step="10" @change="applyPageSize" />
                <small><span>10</span><span>100</span></small>
              </label>
            </div>
          </div>

          <div class="sort-control">
            <button class="tool-button labeled" type="button" @click.stop="toggleSortPanel">
              <el-icon><Sort /></el-icon>
              <span>排序</span>
            </button>

            <div v-if="sortPanelVisible" class="sort-panel" @click.stop>
              <p class="view-panel-title">排序方式</p>
              <button
                v-for="option in sortOptions"
                :key="option.value"
                class="sort-option"
                :class="{ active: sortMode === option.value }"
                type="button"
                @click="setSortMode(option.value)"
              >
                <span>{{ option.label }}</span>
                <el-icon v-if="sortMode === option.value"><Check /></el-icon>
              </button>
            </div>
          </div>

          <button
            class="tool-button danger"
            type="button"
            title="清空回收站"
            :disabled="recycleStore.items.length === 0"
            @click.stop="openEmptyDialog"
          >
            <el-icon><Delete /></el-icon>
          </button>
        </div>
      </div>

      <div class="recycle-info-strip">
        <div>
          <strong>{{ recycleStore.total }}</strong>
          <span>个对象</span>
        </div>
        <div>
          <strong>{{ formattedTotalSize }}</strong>
          <span>占用空间</span>
        </div>
        <div>
          <strong>{{ selectedIds.length }}</strong>
          <span>已选择</span>
        </div>
        <p>
          <el-icon><InfoFilled /></el-icon>
          回收站文件将在 30 天后自动永久删除，请及时还原重要文件。
        </p>
      </div>

      <div v-if="selectedIds.length > 0" class="selection-toolbar">
        <button class="selection-icon-button" type="button" title="取消选择" @click="clearSelection">
          <el-icon><Close /></el-icon>
        </button>
        <button class="select-all-button" type="button" @click="toggleSelectAll">
          <el-icon><Check /></el-icon>
          <span>{{ allDisplayedSelected ? '取消全选' : '全选' }}</span>
        </button>
        <span>已选择 {{ selectedIds.length }} 个对象</span>
        <div class="selection-actions">
          <button class="selection-button" type="button" @click="handleBatchRestore">
            <el-icon><RefreshRight /></el-icon>
            <span>还原</span>
          </button>
          <button class="selection-button danger" type="button" @click="openBatchDeleteDialog">
            <el-icon><Delete /></el-icon>
            <span>永久删除</span>
          </button>
        </div>
      </div>

      <div class="recycle-content">
        <div v-if="recycleStore.isLoading && recycleStore.items.length === 0" class="recycle-state">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>正在加载回收站...</span>
        </div>

        <div v-else-if="displayedItems.length === 0" class="recycle-state">
          <div class="empty-icon"><el-icon><DeleteFilled /></el-icon></div>
          <strong>回收站为空</strong>
          <span>删除的文件将会在这里保留 30 天</span>
        </div>

        <div v-else-if="activeViewMode === 'table'" class="recycle-table">
          <div class="table-head" :style="tableGridStyle">
            <button
              class="table-select-all-cell"
              type="button"
              :title="allDisplayedSelected ? '取消全选' : '全选当前页'"
              @click="toggleSelectAll"
            >
              <span class="head-check" :class="{ active: allDisplayedSelected }">
                <el-icon v-if="allDisplayedSelected"><Check /></el-icon>
              </span>
              <span>全选</span>
            </button>
            <span v-if="visibleColumns.name">名称</span>
            <span v-if="visibleColumns.size">大小</span>
            <span v-if="visibleColumns.expiresAt">过期时间</span>
            <span v-if="visibleColumns.originalPath">原始位置</span>
            <span class="table-actions-head">操作</span>
          </div>

          <article
            v-for="item in displayedItems"
            :key="item.id"
            class="table-row"
            :class="{ selected: selectedIds.includes(item.id) }"
            :style="tableGridStyle"
            @click="toggleSelection(item.id)"
          >
            <div class="select-cell">
              <button class="select-dot" type="button" @click.stop="toggleSelection(item.id)">
                <el-icon v-if="selectedIds.includes(item.id)"><Check /></el-icon>
              </button>
            </div>
            <div v-if="visibleColumns.name" class="name-cell">
              <el-icon class="file-icon" :class="normalizeFileType(item.fileType)">
                <component :is="getItemIcon(item)" />
              </el-icon>
              <span :title="item.fileName" class="file-name-text">{{ item.fileName }}</span>
            </div>
            <span v-if="visibleColumns.size">{{ formatFileSize(item.fileSize) }}</span>
            <span v-if="visibleColumns.expiresAt">{{ recycleStore.getRemainingDays(item) }} 天后</span>
            <span v-if="visibleColumns.originalPath" class="path-cell" :title="item.originalPath || '/'">
              <el-icon><House /></el-icon>
              {{ formatOriginalPath(item.originalPath) }}
            </span>
            <div class="row-actions">
              <button class="row-action restore" type="button" title="还原" @click.stop="handleRestore(item.id)">
                <el-icon><RefreshRight /></el-icon>
              </button>
              <button class="row-action danger" type="button" title="永久删除" @click.stop="openDeleteDialog(item.id)">
                <el-icon><Delete /></el-icon>
              </button>
            </div>
          </article>
        </div>

        <div
          v-else-if="activeViewMode === 'grid'"
          class="recycle-grid"
          :style="{ '--recycle-card-min': `${gridItemSize}px` }"
        >
          <RecycleItem
            v-for="item in displayedItems"
            :key="item.id"
            :item="item"
            :is-selected="selectedIds.includes(item.id)"
            @restore="handleRestore"
            @permanent-delete="openDeleteDialog"
            @selection-change="handleSelectionChange"
          />
        </div>

        <div
          v-else
          class="recycle-gallery"
          :style="{ '--gallery-size': `${galleryImageSize}px` }"
        >
          <button
            v-for="item in displayedItems"
            :key="item.id"
            class="gallery-card"
            :class="{ selected: selectedIds.includes(item.id) }"
            type="button"
            @click="toggleSelection(item.id)"
          >
            <span class="gallery-select">
              <el-icon v-if="selectedIds.includes(item.id)"><Check /></el-icon>
            </span>
            <span class="gallery-preview">
              <el-icon :size="Math.min(galleryImageSize * 0.48, 86)" :class="normalizeFileType(item.fileType)">
                <component :is="getItemIcon(item)" />
              </el-icon>
            </span>
            <span class="gallery-caption" :title="item.fileName">{{ item.fileName }}</span>
            <span class="gallery-meta">剩余 {{ recycleStore.getRemainingDays(item) }} 天</span>
          </button>
        </div>
      </div>

      <div v-if="recycleStore.total > 0" class="recycle-pagination">
        <el-pagination
          v-model:current-page="recycleStore.currentPage"
          v-model:page-size="recycleStore.pageSize"
          :total="recycleStore.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          background
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </section>

    <el-dialog v-model="columnDialogVisible" title="列设置" width="420px">
      <div class="column-settings">
        <label v-for="column in columnOptions" :key="column.key">
          <span>{{ column.label }}</span>
          <el-switch v-model="visibleColumns[column.key]" />
        </label>
      </div>
    </el-dialog>

    <el-dialog v-model="deleteDialogVisible" title="永久删除" width="420px" :close-on-click-modal="false">
      <div class="recycle-dialog-content">
        <el-icon class="warning-icon"><WarningFilled /></el-icon>
        <p>确定永久删除 <strong>{{ deleteTarget?.fileName }}</strong> 吗？</p>
        <span>此操作无法撤销，文件记录和存储内容都会被删除。</span>
      </div>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="recycleStore.isLoading" @click="confirmDelete">永久删除</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchDeleteDialogVisible" title="批量永久删除" width="420px" :close-on-click-modal="false">
      <div class="recycle-dialog-content">
        <el-icon class="warning-icon"><WarningFilled /></el-icon>
        <p>确定永久删除选中的 <strong>{{ selectedIds.length }}</strong> 个文件吗？</p>
        <span>此操作无法撤销。</span>
      </div>
      <template #footer>
        <el-button @click="batchDeleteDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="recycleStore.isLoading" @click="confirmBatchDelete">永久删除</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="emptyDialogVisible" title="清空回收站" width="420px" :close-on-click-modal="false">
      <div class="recycle-dialog-content">
        <el-icon class="warning-icon"><WarningFilled /></el-icon>
        <p>确定清空回收站中的 <strong>{{ recycleStore.total }}</strong> 个文件吗？</p>
        <span>预计释放 {{ formattedTotalSize }}，此操作无法撤销。</span>
      </div>
      <template #footer>
        <el-button @click="emptyDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="recycleStore.isLoading" @click="confirmEmpty">清空回收站</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import {
  Check,
  Close,
  Delete,
  DeleteFilled,
  Document as DocumentIcon,
  Files,
  Grid,
  House,
  InfoFilled,
  List,
  Loading,
  Picture,
  Refresh,
  RefreshRight,
  Setting,
  Sort,
  VideoCamera,
  WarningFilled,
} from '@element-plus/icons-vue';
import RecycleItem from '@/components/RecycleItem/index.vue';
import { useRecycleStore } from '@/stores/recycle';
import type { RecycleItem as RecycleItemType } from '@/types/recycle';
import { formatFileSize } from '@/utils/format';

type ViewMode = 'grid' | 'table' | 'gallery';
type SortMode = 'deleted-desc' | 'deleted-asc' | 'name-asc' | 'size-desc' | 'expires-asc';

const recycleStore = useRecycleStore();
const selectedIds = ref<string[]>([]);
const deleteDialogVisible = ref(false);
const batchDeleteDialogVisible = ref(false);
const emptyDialogVisible = ref(false);
const columnDialogVisible = ref(false);
const viewPanelVisible = ref(false);
const sortPanelVisible = ref(false);
const deleteTarget = ref<RecycleItemType | null>(null);
const activeViewMode = ref<ViewMode>('table');
const sortMode = ref<SortMode>('deleted-desc');
const gridItemSize = ref(280);
const galleryImageSize = ref(160);
const pageSizeDraft = ref(20);
const visibleColumns = reactive({
  name: true,
  size: true,
  expiresAt: true,
  originalPath: true,
});

const viewOptions = [
  { mode: 'grid' as const, label: '网格', icon: Grid },
  { mode: 'table' as const, label: '列表', icon: List },
  { mode: 'gallery' as const, label: '画廊', icon: Picture },
];
const sortOptions = [
  { value: 'deleted-desc' as const, label: '删除时间从近到远' },
  { value: 'deleted-asc' as const, label: '删除时间从远到近' },
  { value: 'name-asc' as const, label: '名称 A-Z' },
  { value: 'size-desc' as const, label: '大小从大到小' },
  { value: 'expires-asc' as const, label: '过期时间最近' },
];
const columnOptions: { key: keyof typeof visibleColumns; label: string }[] = [
  { key: 'name', label: '名称' },
  { key: 'size', label: '大小' },
  { key: 'expiresAt', label: '过期时间' },
  { key: 'originalPath', label: '原始位置' },
];

const formattedTotalSize = computed(() => formatFileSize(recycleStore.getTotalSize()));
const viewModeIcon = computed(() => viewOptions.find((item) => item.mode === activeViewMode.value)?.icon || List);
const displayedItems = computed(() => {
  const items = [...recycleStore.items];
  return items.sort((a, b) => {
    if (sortMode.value === 'deleted-asc') return Date.parse(a.deletedAt) - Date.parse(b.deletedAt);
    if (sortMode.value === 'name-asc') return a.fileName.localeCompare(b.fileName, 'zh-CN');
    if (sortMode.value === 'size-desc') return b.fileSize - a.fileSize;
    if (sortMode.value === 'expires-asc') return Date.parse(a.expiresAt) - Date.parse(b.expiresAt);
    return Date.parse(b.deletedAt) - Date.parse(a.deletedAt);
  });
});
const allDisplayedSelected = computed(() => (
  displayedItems.value.length > 0 && displayedItems.value.every((item) => selectedIds.value.includes(item.id))
));
const tableGridStyle = computed(() => {
  const columns = ['112px'];
  if (visibleColumns.name) columns.push('minmax(280px, 1.8fr)');
  if (visibleColumns.size) columns.push('120px');
  if (visibleColumns.expiresAt) columns.push('150px');
  if (visibleColumns.originalPath) columns.push('minmax(180px, 1fr)');
  columns.push('116px');
  return { gridTemplateColumns: columns.join(' ') };
});

onMounted(() => {
  pageSizeDraft.value = recycleStore.pageSize;
  document.addEventListener('click', closeFloatingPanels);
  reloadCurrentPage();
});

onBeforeUnmount(() => {
  document.removeEventListener('click', closeFloatingPanels);
});

async function reloadCurrentPage() {
  try {
    await recycleStore.loadItems(recycleStore.currentPage, recycleStore.pageSize, recycleQueryParams());
    ElMessage.success('回收站已刷新');
  } catch (error: any) {
    ElMessage.error(error.message || '加载回收站失败');
  }
}

function closeFloatingPanels() {
  viewPanelVisible.value = false;
  sortPanelVisible.value = false;
}

function toggleViewPanel() {
  viewPanelVisible.value = !viewPanelVisible.value;
  if (viewPanelVisible.value) {
    sortPanelVisible.value = false;
  }
}

function toggleSortPanel() {
  sortPanelVisible.value = !sortPanelVisible.value;
  if (sortPanelVisible.value) {
    viewPanelVisible.value = false;
  }
}

function setViewMode(mode: ViewMode) {
  activeViewMode.value = mode;
}

function setSortMode(mode: SortMode) {
  sortMode.value = mode;
  sortPanelVisible.value = false;
  void handlePageChange(1);
}

async function applyPageSize() {
  await handleSizeChange(pageSizeDraft.value);
}

function clearSelection() {
  selectedIds.value = [];
}

function toggleSelectAll() {
  if (displayedItems.value.length === 0) return;

  const displayedIds = displayedItems.value.map((item) => item.id);
  if (allDisplayedSelected.value) {
    const displayedIdSet = new Set(displayedIds);
    selectedIds.value = selectedIds.value.filter((id) => !displayedIdSet.has(id));
    return;
  }

  selectedIds.value = Array.from(new Set([...selectedIds.value, ...displayedIds]));
}

function toggleSelection(itemId: string) {
  if (selectedIds.value.includes(itemId)) {
    selectedIds.value = selectedIds.value.filter((id) => id !== itemId);
  } else {
    selectedIds.value.push(itemId);
  }
}

function handleSelectionChange(itemId: string, selected: boolean) {
  if (selected && !selectedIds.value.includes(itemId)) {
    selectedIds.value.push(itemId);
    return;
  }

  if (!selected) {
    selectedIds.value = selectedIds.value.filter((id) => id !== itemId);
  }
}

async function handleRestore(itemId: string) {
  try {
    const item = recycleStore.items.find((entry) => entry.id === itemId);
    await recycleStore.restoreFiles([itemId]);
    selectedIds.value = selectedIds.value.filter((id) => id !== itemId);
    ElMessage.success(`已还原 ${item?.fileName || '文件'}`);
  } catch (error: any) {
    ElMessage.error(error.message || '还原失败');
  }
}

async function handleBatchRestore() {
  if (selectedIds.value.length === 0) return;

  try {
    const count = selectedIds.value.length;
    await recycleStore.restoreFiles([...selectedIds.value]);
    selectedIds.value = [];
    ElMessage.success(`已还原 ${count} 个文件`);
  } catch (error: any) {
    ElMessage.error(error.message || '批量还原失败');
  }
}

function openDeleteDialog(itemId: string) {
  const item = recycleStore.items.find((entry) => entry.id === itemId);
  if (!item) return;

  deleteTarget.value = item;
  deleteDialogVisible.value = true;
}

async function confirmDelete() {
  if (!deleteTarget.value) return;

  try {
    const { id, fileName } = deleteTarget.value;
    await recycleStore.permanentDelete([id]);
    selectedIds.value = selectedIds.value.filter((selectedId) => selectedId !== id);
    deleteDialogVisible.value = false;
    deleteTarget.value = null;
    ElMessage.success(`已永久删除 ${fileName}`);
  } catch (error: any) {
    ElMessage.error(error.message || '永久删除失败');
  }
}

function openBatchDeleteDialog() {
  if (selectedIds.value.length > 0) {
    batchDeleteDialogVisible.value = true;
  }
}

async function confirmBatchDelete() {
  if (selectedIds.value.length === 0) return;

  try {
    const count = selectedIds.value.length;
    await recycleStore.permanentDelete([...selectedIds.value]);
    selectedIds.value = [];
    batchDeleteDialogVisible.value = false;
    ElMessage.success(`已永久删除 ${count} 个文件`);
  } catch (error: any) {
    ElMessage.error(error.message || '批量删除失败');
  }
}

function openEmptyDialog() {
  if (recycleStore.items.length > 0) {
    emptyDialogVisible.value = true;
  }
}

async function confirmEmpty() {
  try {
    await recycleStore.emptyRecycleBin();
    selectedIds.value = [];
    emptyDialogVisible.value = false;
    ElMessage.success('已清空回收站');
  } catch (error: any) {
    ElMessage.error(error.message || '清空回收站失败');
  }
}

async function handlePageChange(page: number) {
  selectedIds.value = [];
  try {
    await recycleStore.loadItems(page, recycleStore.pageSize, recycleQueryParams());
  } catch (error: any) {
    ElMessage.error(error.message || '加载页面失败');
  }
}

async function handleSizeChange(pageSize: number) {
  selectedIds.value = [];
  pageSizeDraft.value = pageSize;
  try {
    await recycleStore.loadItems(1, pageSize, recycleQueryParams());
  } catch (error: any) {
    ElMessage.error(error.message || '加载页面失败');
  }
}

function recycleQueryParams() {
  return {
    sort: sortMode.value,
  };
}

function normalizeFileType(type: string) {
  const value = (type || '').toLowerCase();
  if (value.includes('image')) return 'image';
  if (value.includes('video')) return 'video';
  if (value.includes('folder')) return 'folder';
  if (value.includes('pdf')) return 'pdf';
  if (value.includes('zip') || value.includes('rar') || value.includes('archive')) return 'archive';
  return 'document';
}

function getItemIcon(item: RecycleItemType) {
  const type = normalizeFileType(item.fileType);
  if (type === 'image') return Picture;
  if (type === 'video') return VideoCamera;
  if (type === 'folder') return Files;
  if (type === 'archive') return Files;
  return DocumentIcon;
}

function formatOriginalPath(path: string) {
  if (!path || path === '/') return '我的文件';
  return path.replace(/^\/+/, '') || '我的文件';
}
</script>

<style>
.recycle-workspace {
  min-height: 100%;
  padding: 12px 18px 24px;
  color: #172033;
}

.recycle-shell {
  min-height: calc(100vh - 132px);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.recycle-pathbar,
.recycle-info-strip,
.selection-toolbar,
.recycle-content,
.view-panel,
.sort-panel {
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.62);
  box-shadow: 0 16px 40px rgba(84, 123, 169, 0.12), inset 0 1px rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(18px) saturate(1.1);
}

.recycle-pathbar {
  position: relative;
  z-index: 60;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  min-height: 58px;
  padding: 0 14px 0 18px;
}

.path-title,
.path-actions,
.selection-toolbar,
.selection-actions,
.name-cell,
.path-cell,
.row-actions {
  display: flex;
  align-items: center;
}

.path-title {
  gap: 10px;
  color: #15213a;
  font-size: 18px;
}

.path-title .el-icon {
  color: #1f78ea;
}

.path-actions {
  gap: 10px;
}

.tool-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-width: 52px;
  height: 46px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: 8px;
  color: #1d2b42;
  background: rgba(255, 255, 255, 0.68);
  box-shadow: 0 10px 24px rgba(87, 126, 170, 0.1), inset 0 1px rgba(255, 255, 255, 0.9);
  cursor: pointer;
  font-size: 15px;
  font-weight: 800;
}

.tool-button:hover {
  color: #1677ff;
  background: rgba(240, 248, 255, 0.92);
}

.tool-button.danger {
  color: #e44961;
}

.tool-button:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.view-control,
.sort-control {
  position: relative;
}

.view-panel,
.sort-panel {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  z-index: 100;
  width: 360px;
  padding: 18px;
}

.sort-panel {
  width: 250px;
}

.view-panel-title {
  margin: 0 0 10px;
  color: #707887;
  font-size: 14px;
  font-weight: 850;
}

.view-segment {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  padding: 4px;
  border: 1px solid rgba(218, 226, 236, 0.88);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.58);
}

.view-option,
.panel-wide-button,
.sort-option {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 44px;
  border: 0;
  border-radius: 8px;
  color: #687385;
  background: transparent;
  cursor: pointer;
  font-weight: 800;
}

.view-option.active,
.view-option:hover,
.panel-wide-button:hover,
.sort-option.active,
.sort-option:hover {
  color: #1677ff;
  background: rgba(226, 241, 255, 0.88);
}

.panel-wide-button {
  width: 100%;
  margin-bottom: 14px;
  border: 1px solid rgba(218, 226, 236, 0.88);
  background: rgba(255, 255, 255, 0.58);
}

.sort-option {
  width: 100%;
  justify-content: space-between;
  padding: 0 12px;
}

.slider-field {
  display: grid;
  gap: 9px;
  margin-top: 16px;
  color: #707887;
  font-size: 14px;
  font-weight: 850;
}

.slider-field input {
  width: 100%;
  accent-color: #1677ff;
}

.slider-field small {
  display: flex;
  justify-content: space-between;
  color: #8b93a1;
}

.recycle-info-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(120px, 160px)) minmax(260px, 1fr);
  align-items: center;
  gap: 12px;
  padding: 12px 18px;
}

.recycle-info-strip div {
  display: grid;
  gap: 2px;
}

.recycle-info-strip strong {
  color: #10213f;
  font-size: 18px;
}

.recycle-info-strip span,
.recycle-info-strip p {
  color: #718097;
}

.recycle-info-strip p {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin: 0;
}

.selection-toolbar {
  min-height: 64px;
  justify-content: space-between;
  padding: 0 16px;
  color: #10213f;
  font-weight: 850;
}

.selection-icon-button,
.selection-button,
.row-action,
.select-dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  cursor: pointer;
}

.selection-icon-button,
.row-action,
.select-dot {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.7);
  color: #1d2b42;
}

.selection-actions {
  gap: 8px;
}

.select-all-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 8px;
  color: #1677ff;
  background: rgba(229, 242, 255, 0.78);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.9);
  cursor: pointer;
  font-weight: 850;
}

.select-all-button:hover {
  background: rgba(211, 234, 255, 0.92);
}

.selection-button {
  gap: 8px;
  min-height: 38px;
  padding: 0 14px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3389ff, #20c9bd);
  color: #fff;
  font-weight: 800;
}

.selection-button.danger {
  background: rgba(255, 93, 118, 0.14);
  color: #e44961;
}

.recycle-content {
  flex: 1;
  min-height: 460px;
  overflow: auto;
}

.recycle-state {
  min-height: 430px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #7b8798;
  text-align: center;
}

.recycle-state .is-loading {
  color: #1677ff;
  font-size: 42px;
}

.empty-icon {
  width: 78px;
  height: 78px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  color: rgba(107, 119, 138, 0.52);
  background: rgba(255, 255, 255, 0.58);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.86), 0 16px 34px rgba(109, 132, 160, 0.12);
  font-size: 42px;
}

.recycle-table {
  min-width: 860px;
}

.table-head,
.table-row {
  display: grid;
  align-items: center;
  gap: 18px;
  padding: 0 14px;
}

.table-head {
  min-height: 54px;
  border-bottom: 1px solid rgba(217, 225, 236, 0.86);
  color: #111827;
  font-size: 15px;
  font-weight: 900;
}

.table-select-all-cell {
  appearance: none;
  -webkit-appearance: none;
  display: grid;
  grid-template-columns: 34px max-content;
  align-items: center;
  justify-content: start;
  gap: 12px;
  min-width: 0;
  height: auto;
  padding: 0;
  border: 0 !important;
  outline: 0;
  color: inherit;
  background: transparent !important;
  box-shadow: none !important;
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.table-select-all-cell:hover {
  color: #10213f;
}

.head-check {
  width: 34px;
  height: 34px;
  display: inline-grid;
  place-items: center;
  flex: 0 0 auto;
  border: 1px solid rgba(203, 216, 231, 0.92);
  border-radius: 8px;
  color: #fff;
  background: rgba(255, 255, 255, 0.72);
  box-shadow:
    inset 0 1px rgba(255, 255, 255, 0.96),
    0 8px 18px rgba(77, 127, 180, 0.08);
  transition: all 0.16s ease;
}

.table-select-all-cell:hover .head-check {
  border-color: rgba(203, 216, 231, 0.92);
  background: rgba(255, 255, 255, 0.72);
}

.head-check.active {
  border-color: transparent;
  background: linear-gradient(135deg, #2588ff, #1cc8c0);
  box-shadow:
    0 12px 22px rgba(37, 136, 255, 0.18),
    inset 0 1px rgba(255, 255, 255, 0.34);
}

.table-row {
  min-height: 58px;
  border-bottom: 1px solid rgba(225, 232, 241, 0.72);
  color: #223047;
  cursor: pointer;
}

.select-cell {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.table-row:hover,
.table-row.selected {
  background: rgba(226, 241, 255, 0.66);
}

.name-cell {
  min-width: 0;
  gap: 12px;
}

.name-cell span,
.path-cell {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-name-text {
  min-width: 0;
}

.select-dot {
  flex: 0 0 auto;
  border: 1px solid rgba(207, 218, 230, 0.9);
}

.table-row.selected .select-dot {
  color: #fff;
  border-color: transparent;
  background: #1677ff;
}

.select-dot.active {
  color: #fff;
  border-color: transparent;
  background: #1677ff;
}

.file-icon {
  flex: 0 0 auto;
  color: #697483;
  font-size: 22px;
}

.file-icon.image,
.gallery-preview .image {
  color: #e13b45;
}

.file-icon.video,
.gallery-preview .video {
  color: #8a68f1;
}

.file-icon.folder,
.gallery-preview .folder {
  color: #f2a72f;
}

.path-cell {
  gap: 8px;
  color: #687385;
}

.row-actions {
  justify-content: flex-end;
  gap: 8px;
}

.row-action.restore {
  color: #1677ff;
}

.row-action.danger {
  color: #e44961;
}

.recycle-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--recycle-card-min, 280px), 1fr));
  gap: 16px;
  padding: 18px;
}

.recycle-gallery {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(calc(var(--gallery-size, 160px) + 44px), max-content));
  align-items: start;
  gap: 18px;
  padding: 20px;
}

.gallery-card {
  position: relative;
  width: calc(var(--gallery-size, 160px) + 28px);
  padding: 12px 12px 14px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: #172033;
  cursor: pointer;
}

.gallery-card:hover,
.gallery-card.selected {
  border-color: rgba(76, 154, 255, 0.54);
  background: rgba(226, 241, 255, 0.58);
}

.gallery-select {
  position: absolute;
  top: 16px;
  left: 16px;
  width: 26px;
  height: 26px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  color: #fff;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: inset 0 0 0 1px rgba(194, 210, 228, 0.88);
}

.gallery-card.selected .gallery-select {
  background: #1677ff;
  box-shadow: none;
}

.gallery-preview {
  width: var(--gallery-size, 160px);
  height: var(--gallery-size, 160px);
  display: grid;
  place-items: center;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.68);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.9), 0 14px 28px rgba(77, 127, 180, 0.1);
}

.gallery-caption,
.gallery-meta {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: center;
}

.gallery-caption {
  margin-top: 10px;
  font-weight: 850;
}

.gallery-meta {
  margin-top: 4px;
  color: #718097;
  font-size: 12px;
}

.recycle-pagination {
  display: flex;
  justify-content: center;
  padding: 8px 0 0;
}

.column-settings {
  display: grid;
  gap: 12px;
}

.column-settings label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 44px;
  padding: 0 12px;
  border-radius: 8px;
  background: rgba(245, 248, 252, 0.9);
  color: #243047;
  font-weight: 800;
}

.recycle-dialog-content {
  text-align: center;
}

.recycle-dialog-content p {
  margin: 12px 0 8px;
  color: #243047;
}

.recycle-dialog-content span {
  color: #778397;
  font-size: 13px;
}

.warning-icon {
  font-size: 46px;
  color: #f05b70;
}

@media (max-width: 900px) {
  .recycle-pathbar,
  .selection-toolbar {
    align-items: stretch;
    flex-direction: column;
    padding: 14px;
  }

  .path-actions {
    flex-wrap: wrap;
  }

  .recycle-info-strip {
    grid-template-columns: repeat(3, 1fr);
  }

  .recycle-info-strip p {
    grid-column: 1 / -1;
    justify-content: flex-start;
  }
}
</style>

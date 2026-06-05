<template>
  <main class="my-shares-page" @click="closeFloatingLayers" @contextmenu="showBlankContextMenu">
    <section class="filter-panel glass-panel">
      <label class="search-field">
        <el-icon><Search /></el-icon>
        <input v-model="keyword" type="search" placeholder="筛选分享文件名称、链接或状态" />
      </label>
      <div class="drive-actions">
        <button class="tool-button" type="button" title="刷新" @click.stop="refreshMyShares">
          <el-icon><Refresh /></el-icon>
        </button>
        <button class="tool-button" type="button" title="复制选中链接" @click.stop="copySelectedLinks">
          <el-icon><CopyDocument /></el-icon>
        </button>
        <div class="view-control" @click.stop>
          <button class="view-button" type="button" :class="{ active: viewPanelVisible }" @click="toggleViewPanel">
            <el-icon><Grid /></el-icon>
            <span>视图</span>
          </button>
        </div>
        <button class="sort-button" type="button" :class="{ active: sortPanelVisible }" @click.stop="toggleSortPanel">
          <el-icon><Sort /></el-icon>
          <span>{{ currentSortLabel }}</span>
        </button>
        <span class="result-count">{{ filteredItems.length }} / {{ items.length }}</span>
      </div>
    </section>

    <section class="drive-bar glass-panel">
      <div class="breadcrumb">
        <el-icon><Share /></el-icon>
        <span>我的分享</span>
      </div>
      <div class="drive-bar-actions">
        <button class="ghost-action" type="button" @click.stop="refreshMyShares">
          <el-icon><Refresh /></el-icon>
          刷新
        </button>
        <button class="ghost-action" type="button" :class="{ active: sortPanelVisible }" @click.stop="toggleSortPanel">
          <el-icon><Sort /></el-icon>
          排序
        </button>
      </div>
    </section>

    <section v-if="selectedItems.length" class="selection-strip glass-panel">
      <span>已选择 {{ selectedItems.length }} 条分享</span>
      <div>
        <button v-if="selectedItems.length === 1" type="button" @click="openSelectedShare">打开</button>
        <button type="button" @click="copySelectedLinks">复制链接</button>
        <button type="button" class="danger" @click="deleteSelectedShares">删除</button>
        <button type="button" @click="clearSelection">取消</button>
      </div>
    </section>

    <section
      class="share-board glass-panel"
      :class="`is-${viewMode}`"
      :style="{ '--card-size': `${cardSize}px` }"
    >
      <div v-if="loading" class="empty-state">正在加载分享...</div>
      <div v-else-if="!filteredItems.length" class="empty-state">
        <div class="empty-visual">
          <el-icon><Share /></el-icon>
        </div>
        <strong>还没有分享记录</strong>
        <span>在文件、图片、视频、音乐或文档界面创建分享后，会在这里集中管理。</span>
      </div>
      <article
        v-for="item in filteredItems"
        v-else
        :key="item.id"
        class="share-card"
        :class="{ selected: selectedIds.includes(item.id), expired: item.expired }"
        @click.stop="toggleSelect(item)"
        @dblclick.stop="openShare(item)"
        @contextmenu.prevent.stop="showItemContextMenu($event, item)"
      >
        <button class="select-dot" type="button" :aria-label="`选择 ${item.display_name}`" @click.stop="toggleSelect(item)">
          <el-icon v-if="selectedIds.includes(item.id)"><Check /></el-icon>
        </button>

        <button class="share-cover" type="button" @click.stop="toggleSelect(item)" @dblclick.stop="openShare(item)">
          <span class="status-chip" :class="{ expired: item.expired }">{{ item.expired ? '已过期' : '有效' }}</span>
          <el-icon><component :is="fileIcon(item)" /></el-icon>
        </button>

        <div class="share-info">
          <strong :title="item.display_name">{{ item.display_name }}</strong>
          <span>{{ item.share_link }}</span>
          <dl class="share-meta">
            <div>
              <dt>分享时间</dt>
              <dd>{{ formatDate(item.created_at) }}</dd>
            </div>
            <div>
              <dt>状态</dt>
              <dd :class="{ danger: item.expired }">{{ item.expired ? '已过期' : '有效' }}</dd>
            </div>
            <div>
              <dt>访问</dt>
              <dd>{{ item.access_count || 0 }} 次</dd>
            </div>
            <div>
              <dt>下载</dt>
              <dd>{{ item.download_count || 0 }} 次</dd>
            </div>
            <div>
              <dt>过期时间</dt>
              <dd>{{ expiryLabel(item) }}</dd>
            </div>
          </dl>
        </div>

        <div class="card-actions">
          <button type="button" title="打开" @click.stop="openShare(item)">
            <el-icon><View /></el-icon>
          </button>
          <button type="button" title="复制链接" @click.stop="copyShareLink(item)">
            <el-icon><CopyDocument /></el-icon>
          </button>
          <button type="button" title="更多" @click.stop="showItemContextMenu($event, item)">
            <el-icon><MoreFilled /></el-icon>
          </button>
        </div>
      </article>
    </section>

    <Teleport to="body">
      <div
        v-if="viewPanelVisible"
        class="view-panel"
        :style="{ left: `${viewPanelPosition.x}px`, top: `${viewPanelPosition.y}px` }"
        @click.stop
      >
        <p>布局</p>
        <div class="segmented">
          <button type="button" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'">
            <el-icon><Grid /></el-icon>
            网格
          </button>
          <button type="button" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">
            <el-icon><List /></el-icon>
            列表
          </button>
          <button type="button" :class="{ active: viewMode === 'compact' }" @click="viewMode = 'compact'">
            <el-icon><Share /></el-icon>
            紧凑
          </button>
        </div>
        <label class="slider-row">
          <span>卡片大小</span>
          <input v-model.number="cardSize" type="range" min="190" max="380" step="10" />
        </label>
        <div class="slider-scale">
          <span>190</span>
          <span>380</span>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="sortPanelVisible"
        class="sort-panel"
        :style="{ left: `${sortPanelPosition.x}px`, top: `${sortPanelPosition.y}px` }"
        @click.stop
      >
        <p>排序方式</p>
        <button
          v-for="option in sortOptions"
          :key="option.value"
          type="button"
          :class="{ active: sortMode === option.value }"
          @click="setSortMode(option.value)"
        >
          <el-icon><component :is="option.icon" /></el-icon>
          <span>{{ option.label }}</span>
        </button>
      </div>
    </Teleport>

    <MyShareContextMenu
      :visible="itemMenuVisible"
      :x="itemMenuPosition.x"
      :y="itemMenuPosition.y"
      @open="openContextShare"
      @copy="copyContextShareLink"
      @edit="editContextShare"
      @details="showContextDetails"
      @delete="deleteContextShare"
    />

    <MyShareBlankContextMenu
      :visible="blankMenuVisible"
      :x="blankMenuPosition.x"
      :y="blankMenuPosition.y"
      @refresh="refreshFromMenu"
      @select-all="selectAll"
      @clear="clearSelection"
      @invert="invertSelection"
    />
  </main>
</template>

<script setup lang="ts">
import { computed, markRaw, reactive, ref } from 'vue';
import {
  Calendar,
  Check,
  CopyDocument,
  Document,
  Files,
  Grid,
  Link,
  List,
  Lock,
  MoreFilled,
  Refresh,
  Search,
  Share,
  Sort,
  View,
} from '@element-plus/icons-vue';
import type { Component } from 'vue';
import { ElIcon, ElMessage, ElMessageBox } from 'element-plus';
import { deleteShare } from '@/api/share';
import { copyToClipboard } from '@/utils/share-utils';
import MyShareBlankContextMenu from './components/MyShareBlankContextMenu.vue';
import MyShareContextMenu from './components/MyShareContextMenu.vue';
import { useMySharesWorkspace, type MyShareItem, type MyShareSortMode } from './useMySharesWorkspace';

type SortOption = {
  value: MyShareSortMode;
  label: string;
  icon: Component;
};

const {
  cardSize,
  clearSelection,
  filteredItems,
  items,
  keyword,
  loading,
  refreshMyShares,
  selectedIds,
  selectedItems,
  selectOnly,
  sortMode,
  toggleSelect,
  viewMode,
} = useMySharesWorkspace();

const viewPanelVisible = ref(false);
const viewPanelPosition = reactive({ x: 0, y: 0 });
const sortPanelVisible = ref(false);
const sortPanelPosition = reactive({ x: 0, y: 0 });
const itemMenuVisible = ref(false);
const itemMenuPosition = reactive({ x: 0, y: 0 });
const itemMenuTarget = ref<MyShareItem | null>(null);
const blankMenuVisible = ref(false);
const blankMenuPosition = reactive({ x: 0, y: 0 });
const sortOptions: SortOption[] = [
  { value: 'recent', label: '最新分享', icon: Calendar },
  { value: 'name', label: '文件名称', icon: Document },
  { value: 'expires', label: '过期时间', icon: Sort },
  { value: 'visits', label: '访问次数', icon: View },
];

const currentSortLabel = computed(() => sortOptions.find((option) => option.value === sortMode.value)?.label || '排序');

function closeViewPanel() {
  viewPanelVisible.value = false;
}

function closeSortPanel() {
  sortPanelVisible.value = false;
}

function closeMenus() {
  itemMenuVisible.value = false;
  blankMenuVisible.value = false;
}

function closeFloatingLayers() {
  closeViewPanel();
  closeSortPanel();
  closeMenus();
}

function toggleViewPanel(event: MouseEvent) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
  const width = 492;
  const height = 232;
  viewPanelPosition.x = Math.max(12, Math.min(rect.right - width, window.innerWidth - width - 12));
  viewPanelPosition.y = Math.max(12, Math.min(rect.bottom + 12, window.innerHeight - height - 12));
  itemMenuVisible.value = false;
  blankMenuVisible.value = false;
  sortPanelVisible.value = false;
  viewPanelVisible.value = !viewPanelVisible.value;
}

function toggleSortPanel(event: MouseEvent) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
  const width = 244;
  const height = 246;
  sortPanelPosition.x = Math.max(12, Math.min(rect.right - width, window.innerWidth - width - 12));
  sortPanelPosition.y = Math.max(12, Math.min(rect.bottom + 12, window.innerHeight - height - 12));
  closeViewPanel();
  closeMenus();
  sortPanelVisible.value = !sortPanelVisible.value;
}

function setSortMode(mode: MyShareSortMode) {
  sortMode.value = mode;
  closeSortPanel();
}

function fileExtension(item: MyShareItem) {
  return item.display_name.split('.').pop()?.toLowerCase() || '';
}

function fileIcon(item: MyShareItem) {
  if (item.file_ids.length > 1) return markRaw(Files);
  if (item.has_password) return markRaw(Lock);
  const ext = fileExtension(item);
  if (['url', 'link'].includes(ext)) return markRaw(Link);
  return markRaw(Document);
}

function formatDate(value: string) {
  if (!value) return '未知时间';
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(
    new Date(value),
  );
}

function expiryLabel(item: MyShareItem) {
  if (!item.expires_at) return '永久有效';
  if (item.expired) return '已过期';
  return `到期 ${formatDate(item.expires_at)}`;
}

function showItemContextMenu(event: MouseEvent, item: MyShareItem) {
  event.preventDefault();
  event.stopPropagation();
  closeViewPanel();
  selectOnly(item);
  itemMenuTarget.value = item;
  blankMenuVisible.value = false;
  itemMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 250));
  itemMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 300));
  itemMenuVisible.value = true;
}

function showBlankContextMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  if (
    target?.closest('.share-card') ||
    target?.closest('.el-overlay') ||
    target?.closest('input') ||
    target?.closest('select') ||
    target?.closest('button') ||
    target?.closest('.view-panel') ||
    target?.closest('.sort-panel')
  ) {
    return;
  }

  event.preventDefault();
  closeViewPanel();
  closeSortPanel();
  itemMenuVisible.value = false;
  blankMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 230));
  blankMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 250));
  blankMenuVisible.value = true;
}

function getContextShare() {
  return itemMenuTarget.value;
}

function openShare(item: MyShareItem) {
  selectOnly(item);
  closeMenus();
  window.open(item.share_link, '_blank', 'noopener,noreferrer');
}

function openSelectedShare() {
  const item = selectedItems.value[0];
  if (item) openShare(item);
}

function openContextShare() {
  const item = getContextShare();
  closeMenus();
  if (item) openShare(item);
}

async function copyShareLink(item: MyShareItem) {
  try {
    await copyToClipboard(item.share_link);
    ElMessage.success('分享链接已复制');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '复制失败');
  }
}

async function copySelectedLinks() {
  const links = selectedItems.value.map((item) => item.share_link);
  if (!links.length) {
    ElMessage.info('请先选择分享');
    return;
  }
  try {
    await copyToClipboard(links.join('\n'));
    ElMessage.success(`已复制 ${links.length} 条分享链接`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '复制失败');
  }
}

async function copyContextShareLink() {
  const item = getContextShare();
  closeMenus();
  if (item) await copyShareLink(item);
}

function editContextShare() {
  closeMenus();
  ElMessage.info('分享编辑功能将沿用创建分享的配置弹窗');
}

function showContextDetails() {
  const item = getContextShare();
  closeMenus();
  if (!item) return;
  ElMessageBox.alert(
    `文件：${item.display_name}\n链接：${item.share_link}\n状态：${item.expired ? '已过期' : '有效'}\n访问：${item.access_count || 0}\n下载：${item.download_count || 0}`,
    '分享详情',
    { confirmButtonText: '知道了' },
  );
}

async function deleteShares(targets: MyShareItem[]) {
  if (!targets.length) return;
  await ElMessageBox.confirm(`确定删除 ${targets.length} 条分享吗？删除后原分享链接将失效。`, '删除分享', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  });
  await Promise.all(targets.map((item) => deleteShare(item.share_id)));
  ElMessage.success('分享已删除');
  clearSelection();
  await refreshMyShares();
}

async function deleteSelectedShares() {
  try {
    await deleteShares(selectedItems.value);
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

async function deleteContextShare() {
  const item = getContextShare();
  closeMenus();
  if (!item) return;
  try {
    await deleteShares([item]);
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

async function refreshFromMenu() {
  closeMenus();
  await refreshMyShares();
}

function selectAll() {
  selectedIds.value = filteredItems.value.map((item) => item.id);
  closeMenus();
}

function invertSelection() {
  const visible = new Set(filteredItems.value.map((item) => item.id));
  const selected = new Set(selectedIds.value);
  selectedIds.value = [
    ...selectedIds.value.filter((id) => !visible.has(id)),
    ...filteredItems.value.filter((item) => !selected.has(item.id)).map((item) => item.id),
  ];
  closeMenus();
}
</script>

<style scoped>
.my-shares-page {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: calc(100vh - 112px);
  padding: 14px 26px 28px;
  isolation: isolate;
  overflow: visible;
}

.my-shares-page::before,
.my-shares-page::after {
  content: '';
  position: fixed;
  z-index: -1;
  pointer-events: none;
  border-radius: 999px;
  filter: blur(8px);
}

.my-shares-page::before {
  width: 42vw;
  height: 42vw;
  right: 8vw;
  top: 2vh;
  background: radial-gradient(circle, rgba(191, 219, 254, 0.42), rgba(252, 231, 243, 0.18) 54%, transparent 70%);
}

.my-shares-page::after {
  width: 38vw;
  height: 30vw;
  left: 24vw;
  bottom: 2vh;
  background: radial-gradient(circle, rgba(186, 230, 253, 0.34), rgba(255, 214, 226, 0.2) 58%, transparent 72%);
}

.glass-panel {
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 28px;
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.56), transparent 38%),
    radial-gradient(circle at 100% 10%, rgba(252, 231, 243, 0.48), transparent 38%),
    rgba(255, 255, 255, 0.62);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.86), 0 20px 60px rgba(122, 154, 204, 0.12);
  backdrop-filter: blur(22px);
}

.filter-panel {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) auto;
  align-items: center;
  gap: 18px;
  min-height: 92px;
  padding: 18px 34px;
}

.search-field {
  display: flex;
  align-items: center;
  gap: 18px;
  min-width: 0;
  color: #2f7df5;
}

.search-field input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #10203d;
  font-size: 18px;
  font-weight: 760;
}

.search-field input::placeholder {
  color: rgba(71, 85, 105, 0.72);
}

.drive-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 14px;
  min-width: 0;
}

.tool-button,
.view-button,
.sort-button {
  min-height: 54px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 18px 46px rgba(113, 139, 182, 0.16);
  color: #14223d;
  font-size: 17px;
  font-weight: 820;
}

.tool-button {
  display: inline-grid;
  width: 54px;
  place-items: center;
  cursor: pointer;
}

.view-button {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 0 24px;
  cursor: pointer;
}

.sort-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 150px;
  padding: 0 18px;
  cursor: pointer;
}

.view-button.active,
.sort-button.active,
.tool-button:hover,
.view-button:hover,
.sort-button:hover {
  border-color: rgba(47, 125, 245, 0.48);
  color: #1d72ed;
}

.result-count {
  white-space: nowrap;
  color: #64748b;
  font-size: 18px;
  font-weight: 820;
}

.drive-bar,
.selection-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 76px;
  padding: 0 34px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 16px;
  color: #10203d;
  font-size: 22px;
  font-weight: 900;
}

.drive-bar-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.ghost-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 42px;
  padding: 0 16px;
  border: 1px solid rgba(255, 255, 255, 0.68);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.58);
  color: #172642;
  font-weight: 820;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.78);
  cursor: pointer;
}

.ghost-action:hover,
.ghost-action.active {
  color: #1d72ed;
  border-color: rgba(47, 125, 245, 0.38);
}

.selection-strip {
  min-height: 80px;
}

.selection-strip span {
  color: #172642;
  font-size: 17px;
  font-weight: 820;
}

.selection-strip div {
  display: flex;
  gap: 10px;
}

.selection-strip button {
  min-height: 40px;
  border: 0;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.78);
  color: #172642;
  font-weight: 800;
  cursor: pointer;
}

.selection-strip button:hover {
  color: #1d72ed;
}

.selection-strip .danger:hover {
  color: #ef4444;
}

.share-board {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(var(--card-size), 100%), 1fr));
  align-content: start;
  gap: 18px;
  min-height: 430px;
  padding: 32px;
}

.share-card {
  position: relative;
  display: grid;
  grid-template-rows: minmax(132px, 1fr) auto auto;
  gap: 14px;
  min-height: 376px;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 24px;
  background:
    linear-gradient(160deg, rgba(255, 255, 255, 0.66), rgba(255, 255, 255, 0.42)),
    radial-gradient(circle at 0% 0%, rgba(191, 219, 254, 0.24), transparent 48%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.22), transparent 48%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.88),
    0 18px 42px rgba(105, 133, 178, 0.08);
  backdrop-filter: blur(18px);
  cursor: default;
  transition: border-color 0.18s ease, transform 0.18s ease, box-shadow 0.18s ease;
}

.share-card:hover,
.share-card.selected {
  border-color: rgba(47, 125, 245, 0.52);
  box-shadow: 0 18px 48px rgba(77, 129, 225, 0.16);
  transform: translateY(-1px);
}

.share-card.expired {
  opacity: 0.72;
}

.select-dot {
  position: absolute;
  top: 14px;
  left: 14px;
  z-index: 2;
  display: grid;
  width: 26px;
  height: 26px;
  place-items: center;
  border: 1px solid rgba(47, 125, 245, 0.42);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.78);
  color: #1d72ed;
  cursor: pointer;
}

.share-cover {
  position: relative;
  display: grid;
  min-height: 150px;
  place-items: center;
  border: 0;
  border-radius: 18px;
  background:
    radial-gradient(circle at 25% 20%, rgba(186, 230, 253, 0.64), transparent 44%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 46%),
    rgba(255, 255, 255, 0.62);
  color: #2f7df5;
  cursor: pointer;
}

.share-cover .el-icon {
  width: 68px;
  height: 68px;
}

.status-chip {
  position: absolute;
  top: 12px;
  right: 12px;
  border-radius: 999px;
  padding: 6px 12px;
  background: rgba(34, 197, 94, 0.13);
  color: #15803d;
  font-size: 12px;
  font-weight: 900;
}

.status-chip.expired {
  background: rgba(100, 116, 139, 0.12);
  color: #64748b;
}

.share-info {
  min-width: 0;
}

.share-info strong,
.share-info span {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.share-info strong {
  color: #10203d;
  font-size: 18px;
  font-weight: 900;
}

.share-info span,
dt,
dd {
  margin-top: 6px;
  color: #64748b;
  font-size: 13px;
  font-weight: 720;
}

.share-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin: 14px 0 0;
}

.share-meta div {
  min-width: 0;
  padding: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.48);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.62);
}

.share-meta div:last-child {
  grid-column: 1 / -1;
}

.share-meta dt,
.share-meta dd {
  overflow: hidden;
  margin: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.share-meta dt {
  color: #8b9ab0;
  font-size: 12px;
  font-weight: 760;
}

.share-meta dd {
  margin-top: 4px;
  color: #172642;
  font-size: 13px;
  font-weight: 860;
}

.share-meta dd.danger {
  color: #ef4444;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.card-actions button {
  display: grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border: 0;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.74);
  color: #172642;
  cursor: pointer;
}

.card-actions button:hover {
  color: #1d72ed;
}

.share-board.is-list,
.share-board.is-compact {
  grid-template-columns: 1fr;
}

.share-board.is-list .share-card,
.share-board.is-compact .share-card {
  grid-template-columns: 76px minmax(0, 1fr) auto;
  grid-template-rows: auto;
  align-items: center;
  min-height: 112px;
}

.share-board.is-list .share-meta,
.share-board.is-compact .share-meta {
  grid-template-columns: repeat(5, minmax(94px, 1fr));
}

.share-board.is-list .share-meta div:last-child,
.share-board.is-compact .share-meta div:last-child {
  grid-column: auto;
}

.share-board.is-list .share-cover,
.share-board.is-compact .share-cover {
  min-height: 76px;
}

.share-board.is-list .share-cover .el-icon,
.share-board.is-compact .share-cover .el-icon {
  width: 36px;
  height: 36px;
}

.share-board.is-compact .share-card {
  min-height: 84px;
}

.share-board.is-compact .share-cover {
  min-height: 54px;
}

.share-board.is-list .status-chip,
.share-board.is-compact .status-chip {
  display: none;
}

.empty-state {
  grid-column: 1 / -1;
  display: grid;
  min-height: 340px;
  place-items: center;
  align-content: center;
  gap: 14px;
  color: #64748b;
  text-align: center;
}

.empty-visual {
  display: grid;
  width: 92px;
  height: 92px;
  place-items: center;
  border-radius: 26px;
  background: rgba(255, 255, 255, 0.68);
  color: #2f7df5;
}

.empty-visual .el-icon {
  width: 48px;
  height: 48px;
}

.empty-state strong {
  color: #10203d;
  font-size: 24px;
  font-weight: 920;
}

.view-panel {
  position: fixed;
  z-index: 5000;
  width: 492px;
  padding: 26px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 28px;
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.54), transparent 42%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.52), transparent 44%),
    rgba(255, 255, 255, 0.88);
  box-shadow: 0 24px 72px rgba(92, 120, 166, 0.28);
  backdrop-filter: blur(24px);
}

.sort-panel {
  position: fixed;
  z-index: 5100;
  display: grid;
  gap: 8px;
  width: 244px;
  padding: 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 24px;
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.54), transparent 42%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.52), transparent 44%),
    rgba(255, 255, 255, 0.9);
  box-shadow: 0 24px 72px rgba(92, 120, 166, 0.28);
  backdrop-filter: blur(24px);
}

.sort-panel p {
  margin: 4px 8px 8px;
  color: #64748b;
  font-size: 14px;
  font-weight: 900;
}

.sort-panel button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  border: 0;
  border-radius: 14px;
  background: transparent;
  color: #172642;
  font-weight: 820;
  cursor: pointer;
}

.sort-panel button:hover,
.sort-panel button.active {
  background: rgba(255, 255, 255, 0.72);
  color: #1d72ed;
}

.view-panel p {
  margin: 0 0 18px;
  color: #64748b;
  font-size: 18px;
  font-weight: 900;
}

.segmented {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid rgba(193, 217, 246, 0.78);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.46);
}

.segmented button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 48px;
  border: 0;
  border-right: 1px solid rgba(193, 217, 246, 0.62);
  background: transparent;
  color: #64748b;
  font-weight: 820;
  cursor: pointer;
}

.segmented button:last-child {
  border-right: 0;
}

.segmented button.active {
  background: rgba(47, 125, 245, 0.12);
  color: #1d72ed;
}

.slider-row {
  display: grid;
  gap: 12px;
  margin-top: 22px;
  color: #64748b;
  font-weight: 900;
}

.slider-row input {
  width: 100%;
  accent-color: #2f7df5;
}

.slider-scale {
  display: flex;
  justify-content: space-between;
  color: #64748b;
  font-size: 13px;
  font-weight: 760;
}

@media (max-width: 980px) {
  .filter-panel {
    grid-template-columns: 1fr;
    padding: 26px;
  }

  .drive-actions {
    justify-content: flex-start;
    flex-wrap: wrap;
  }
}

@media (max-width: 640px) {
  .my-shares-page {
    padding: 12px;
  }

  .filter-panel,
  .drive-bar,
  .selection-strip,
  .share-board {
    border-radius: 22px;
    padding: 18px;
  }

  .tool-button,
  .view-button,
  .sort-button {
    min-height: 54px;
  }

  .view-panel {
    left: 12px !important;
    width: calc(100vw - 24px);
  }
}
</style>

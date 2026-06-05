<template>
  <main class="shared-page" @click="closeFloatingLayers" @contextmenu="showBlankContextMenu">
    <section class="filter-panel glass-panel">
      <label class="search-field">
        <el-icon><Search /></el-icon>
        <input v-model="keyword" type="search" placeholder="筛选分享文件名称、分享人或权限" />
      </label>
      <div class="drive-actions">
        <button class="tool-button" type="button" title="刷新" @click="refreshSharedWithMe">
          <el-icon><Refresh /></el-icon>
        </button>
        <button class="tool-button" type="button" title="保存分享为快捷方式" @click="pinShareShortcut">
          <el-icon><FolderAdd /></el-icon>
        </button>
        <div class="view-control" @click.stop>
          <button class="view-button" type="button" :class="{ active: viewPanelVisible }" @click="viewPanelVisible = !viewPanelVisible">
            <el-icon><Grid /></el-icon>
            <span>视图</span>
          </button>
          <div v-if="viewPanelVisible" class="view-panel">
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
              <input v-model.number="cardSize" type="range" min="180" max="380" step="10" />
            </label>
            <div class="slider-scale">
              <span>180</span>
              <span>380</span>
            </div>
          </div>
        </div>
        <select v-model="sortMode" class="sort-select" aria-label="排序">
          <option value="recent">最近分享</option>
          <option value="name">名称</option>
          <option value="owner">分享人</option>
        </select>
        <span class="result-count">{{ filteredItems.length }} / {{ items.length }}</span>
      </div>
    </section>

    <section class="drive-bar glass-panel">
      <div class="breadcrumb">
        <el-icon><Share /></el-icon>
        <span>与我分享</span>
      </div>
    </section>

    <section v-if="selectedItems.length" class="selection-strip glass-panel">
      <span>已选择 {{ selectedItems.length }} 个 · {{ formatFileSize(selectedSize) }}</span>
      <div>
        <button v-if="selectedItems.length === 1" type="button" @click="openSelectedItem">打开</button>
        <button type="button" @click="downloadSelectedItems">下载</button>
        <button type="button" @click="clearSelection">取消</button>
      </div>
    </section>

    <section
      class="shared-board glass-panel"
      :class="`is-${viewMode}`"
      :style="{ '--card-size': `${cardSize}px` }"
    >
      <div v-if="loading" class="empty-state">正在加载分享...</div>
      <div v-else-if="!filteredItems.length" class="empty-state">
        <div class="empty-visual">
          <el-icon><Share /></el-icon>
        </div>
        <strong>没有找到别人的分享</strong>
        <span>访问别人给你的分享链接后，可以把它保存为快捷方式，后续会在这里集中管理。</span>
      </div>
      <article
        v-for="item in filteredItems"
        v-else
        :key="item.id"
        class="shared-card"
        :class="{ selected: selectedIds.includes(item.id) }"
        @click.stop="toggleSelect(item)"
        @dblclick.stop="previewItem(item)"
        @contextmenu.prevent.stop="showItemContextMenu($event, item)"
      >
        <button class="select-dot" type="button" :aria-label="`选择 ${item.file_name}`" @click.stop="toggleSelect(item)">
          <el-icon v-if="selectedIds.includes(item.id)"><Check /></el-icon>
        </button>
        <button class="shared-cover" type="button" @click.stop="toggleSelect(item)" @dblclick.stop="previewItem(item)">
          <span class="permission-chip">{{ permissionLabel(item.permission) }}</span>
          <el-icon><component :is="fileIcon(item)" /></el-icon>
        </button>
        <div class="shared-info">
          <strong :title="item.file_name">{{ item.file_name }}</strong>
          <span>{{ item.owner_name }} · {{ formatDate(item.shared_at) }} · {{ formatFileSize(item.file_size) }}</span>
        </div>
        <div class="card-actions">
          <button type="button" title="预览" @click.stop="previewItem(item)">
            <el-icon><View /></el-icon>
          </button>
          <button type="button" title="下载" @click.stop="downloadOneItem(item)">
            <el-icon><Download /></el-icon>
          </button>
          <button type="button" title="更多" @click.stop="showItemContextMenu($event, item)">
            <el-icon><MoreFilled /></el-icon>
          </button>
        </div>
      </article>
    </section>

    <SharedItemContextMenu
      :visible="itemMenuVisible"
      :x="itemMenuPosition.x"
      :y="itemMenuPosition.y"
      @open="openContextItem"
      @download="downloadContextItem"
      @copy-name="copyContextName"
      @details="showContextDetails"
      @locate="locateContextItem"
      @unpin="unpinContextItem"
    />

    <SharedBlankContextMenu
      :visible="blankMenuVisible"
      :x="blankMenuPosition.x"
      :y="blankMenuPosition.y"
      @pin="pinShareShortcut"
      @refresh="refreshFromMenu"
      @select-all="selectAll"
      @clear="clearSelection"
      @invert="invertSelection"
    />

    <el-dialog
      v-model="previewVisible"
      width="min(980px, 92vw)"
      top="6vh"
      class="shared-preview-dialog"
      :show-close="false"
      destroy-on-close
      @closed="clearPreview"
    >
      <template #header>
        <div class="preview-heading">
          <div>
            <p>分享预览</p>
            <strong>{{ previewState.title }}</strong>
          </div>
          <span>{{ previewState.meta }}</span>
          <button class="preview-close" type="button" title="关闭预览" aria-label="关闭预览" @click.stop="closePreview">
            <el-icon><Close /></el-icon>
          </button>
        </div>
      </template>
      <div class="preview-shell">
        <iframe v-if="previewState.url" class="frame-preview" :src="previewState.url" title="分享预览"></iframe>
        <div v-else class="preview-empty">这个分享暂时无法预览</div>
      </div>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { markRaw, reactive, ref } from 'vue';
import {
  Check,
  Close,
  Document,
  Download,
  Files,
  FolderAdd,
  Grid,
  Headset,
  List,
  MoreFilled,
  Picture,
  Refresh,
  Search,
  Share,
  VideoCameraFilled,
  View,
} from '@element-plus/icons-vue';
import { ElDialog, ElIcon, ElMessage, ElMessageBox } from 'element-plus';
import { downloadFile as apiDownloadFile, getAuthenticatedFileDownloadUrl } from '@/api/file';
import type { SharedWithMeItem } from '@/api/shared-with-me';
import { formatFileSize } from '@/utils/format';
import SharedBlankContextMenu from './components/SharedBlankContextMenu.vue';
import SharedItemContextMenu from './components/SharedItemContextMenu.vue';
import { useSharedWithMeWorkspace } from './useSharedWithMeWorkspace';

const {
  cardSize,
  clearSelection,
  filteredItems,
  items,
  keyword,
  loading,
  refreshSharedWithMe,
  selectedIds,
  selectedItems,
  selectedSize,
  selectOnly,
  sortMode,
  toggleSelect,
  viewMode,
} = useSharedWithMeWorkspace();

const viewPanelVisible = ref(false);
const itemMenuVisible = ref(false);
const itemMenuPosition = reactive({ x: 0, y: 0 });
const itemMenuTarget = ref<SharedWithMeItem | null>(null);
const blankMenuVisible = ref(false);
const blankMenuPosition = reactive({ x: 0, y: 0 });
const previewVisible = ref(false);
const previewState = reactive({ title: '', meta: '', url: '' });

function closeViewPanel() {
  viewPanelVisible.value = false;
}

function closeMenus() {
  itemMenuVisible.value = false;
  blankMenuVisible.value = false;
}

function closeFloatingLayers() {
  closeViewPanel();
  closeMenus();
}

function fileExtension(item: SharedWithMeItem) {
  return item.file_name.split('.').pop()?.toLowerCase() || '';
}

function fileIcon(item: SharedWithMeItem) {
  const ext = fileExtension(item);
  const type = item.file_type.toLowerCase();
  if (type.startsWith('image/') || ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(ext)) return markRaw(Picture);
  if (type.startsWith('video/') || ['mp4', 'mov', 'mkv', 'webm', 'avi'].includes(ext)) return markRaw(VideoCameraFilled);
  if (type.startsWith('audio/') || ['mp3', 'wav', 'flac', 'm4a'].includes(ext)) return markRaw(Headset);
  if (['xls', 'xlsx', 'csv'].includes(ext)) return markRaw(Files);
  return markRaw(Document);
}

function formatDate(value: string) {
  if (!value) return '未知时间';
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit' }).format(new Date(value));
}

function permissionLabel(permission: string) {
  if (permission === 'edit') return '可编辑';
  if (permission === 'download') return '可下载';
  return '仅查看';
}

function showItemContextMenu(event: MouseEvent, item: SharedWithMeItem) {
  event.preventDefault();
  event.stopPropagation();
  closeViewPanel();
  selectOnly(item);
  itemMenuTarget.value = item;
  blankMenuVisible.value = false;
  itemMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 230));
  itemMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 370));
  itemMenuVisible.value = true;
}

function showBlankContextMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  if (
    target?.closest('.shared-card') ||
    target?.closest('.el-dialog') ||
    target?.closest('.el-overlay') ||
    target?.closest('input') ||
    target?.closest('select') ||
    target?.closest('button')
  ) {
    return;
  }

  event.preventDefault();
  closeViewPanel();
  itemMenuVisible.value = false;
  blankMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 260));
  blankMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 330));
  blankMenuVisible.value = true;
}

function getContextItem() {
  return itemMenuTarget.value;
}

function previewItem(item: SharedWithMeItem) {
  selectOnly(item);
  closeMenus();
  previewState.title = item.file_name;
  previewState.meta = `${item.owner_name} · ${permissionLabel(item.permission)} · ${formatFileSize(item.file_size)}`;
  previewState.url = getAuthenticatedFileDownloadUrl(Number(item.file_id), true);
  previewVisible.value = true;
}

function clearPreview() {
  previewState.title = '';
  previewState.meta = '';
  previewState.url = '';
}

function closePreview() {
  previewVisible.value = false;
}

function openSelectedItem() {
  const item = selectedItems.value[0];
  if (item) previewItem(item);
}

function openContextItem() {
  const item = getContextItem();
  closeMenus();
  if (item) previewItem(item);
}

async function downloadOneItem(item: SharedWithMeItem) {
  if (item.permission === 'view') {
    ElMessage.warning('该分享只有查看权限，不能下载');
    return;
  }
  try {
    const blob = await apiDownloadFile(Number(item.file_id));
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = item.file_name;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(link.href);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载失败');
  }
}

async function downloadSelectedItems() {
  for (const item of selectedItems.value) {
    await downloadOneItem(item);
  }
}

async function downloadContextItem() {
  const item = getContextItem();
  closeMenus();
  if (item) await downloadOneItem(item);
}

async function copyContextName() {
  const item = getContextItem();
  closeMenus();
  if (!item) return;
  await navigator.clipboard?.writeText(item.file_name);
  ElMessage.success('名称已复制');
}

function showContextDetails() {
  const item = getContextItem();
  closeMenus();
  if (!item) return;
  ElMessageBox.alert(
    `分享人：${item.owner_name}\n权限：${permissionLabel(item.permission)}\n分享时间：${formatDate(item.shared_at)}\n大小：${formatFileSize(item.file_size)}`,
    item.file_name,
    { confirmButtonText: '知道了' },
  );
}

function locateContextItem() {
  closeMenus();
  ElMessage.info('转到源文件功能即将开放');
}

function unpinContextItem() {
  closeMenus();
  ElMessage.info('移除快捷方式功能即将开放');
}

function pinShareShortcut() {
  closeMenus();
  ElMessage.info('保存分享为快捷方式功能即将开放');
}

async function refreshFromMenu() {
  closeMenus();
  await refreshSharedWithMe();
}

function selectAll() {
  closeMenus();
  selectedIds.value = filteredItems.value.map((item) => item.id);
}

function invertSelection() {
  closeMenus();
  const current = new Set(selectedIds.value);
  selectedIds.value = filteredItems.value.filter((item) => !current.has(item.id)).map((item) => item.id);
}
</script>

<style scoped>
.shared-page {
  display: grid;
  gap: 12px;
  min-height: calc(100vh - 96px);
  padding: 0 4px 24px 22px;
  color: #10213f;
}

.glass-panel {
  border: 1px solid rgba(255, 255, 255, 0.82);
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.5), transparent 38%),
    radial-gradient(circle at 100% 0%, rgba(252, 231, 243, 0.42), transparent 36%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.74), rgba(246, 251, 255, 0.46));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 18px 45px rgba(88, 124, 170, 0.12);
  backdrop-filter: blur(22px);
}

.filter-panel,
.drive-bar,
.selection-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 78px;
  border-radius: 28px;
  padding: 16px 24px;
}

.filter-panel {
  position: relative;
  z-index: 60;
  overflow: visible;
}

.drive-bar,
.selection-strip,
.shared-board {
  position: relative;
  z-index: 1;
}

.search-field {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 240px;
  gap: 14px;
  color: #2f7df5;
}

.search-field input {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: #10213f;
  font-size: 17px;
  font-weight: 760;
}

.search-field input::placeholder {
  color: rgba(16, 33, 63, 0.56);
}

.drive-actions,
.selection-strip > div {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tool-button,
.view-button,
.sort-select,
.selection-strip button,
.card-actions button {
  border: 1px solid rgba(255, 255, 255, 0.8);
  background: rgba(255, 255, 255, 0.72);
  color: #10213f;
  box-shadow: 0 12px 34px rgba(83, 117, 165, 0.12);
  cursor: pointer;
}

.tool-button,
.view-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 52px;
  border-radius: 20px;
  font-size: 18px;
}

.tool-button {
  width: 52px;
}

.view-button {
  gap: 8px;
  min-width: 116px;
  padding: 0 18px;
  font-weight: 820;
}

.tool-button:hover,
.view-button:hover,
.view-button.active,
.selection-strip button:hover,
.card-actions button:hover {
  border-color: rgba(47, 125, 245, 0.45);
  background: rgba(235, 246, 255, 0.88);
  color: #2473e6;
  transform: translateY(-1px);
}

.view-control {
  position: relative;
}

.view-panel {
  position: absolute;
  top: calc(100% + 14px);
  right: 0;
  z-index: 120;
  width: 360px;
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 26px;
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.62), transparent 45%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.48), transparent 42%),
    rgba(255, 255, 255, 0.84);
  box-shadow: 0 24px 60px rgba(82, 112, 160, 0.22);
  backdrop-filter: blur(24px);
}

.view-panel p,
.slider-row span {
  margin: 0 0 10px;
  color: #65758f;
  font-size: 14px;
  font-weight: 820;
}

.segmented {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  overflow: hidden;
  margin-bottom: 18px;
  border: 1px solid rgba(204, 219, 236, 0.76);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.46);
}

.segmented button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 44px;
  border: 0;
  background: transparent;
  color: #65758f;
  font-weight: 780;
  cursor: pointer;
}

.segmented button.active {
  background: rgba(222, 241, 255, 0.82);
  color: #2473e6;
}

.slider-row {
  display: grid;
  gap: 8px;
}

.slider-row input {
  width: 100%;
  accent-color: #2f7df5;
}

.slider-scale {
  display: flex;
  justify-content: space-between;
  color: #71819a;
  font-size: 13px;
  font-weight: 760;
}

.sort-select {
  height: 52px;
  border-radius: 20px;
  padding: 0 42px 0 18px;
  outline: 0;
  color: #10213f;
  font-size: 16px;
  font-weight: 820;
}

.result-count {
  min-width: 48px;
  color: #60708a;
  font-size: 16px;
  font-weight: 820;
  text-align: right;
}

.drive-bar {
  justify-content: flex-start;
}

.breadcrumb {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 900;
}

.selection-strip {
  min-height: 66px;
  color: #10213f;
  font-weight: 820;
}

.selection-strip button {
  min-height: 38px;
  border-radius: 14px;
  padding: 0 14px;
  font-weight: 760;
}

.shared-board {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(var(--card-size), 100%), 1fr));
  align-content: start;
  gap: 18px;
  min-height: 560px;
  border-radius: 30px;
  padding: 28px;
}

.shared-board.is-list,
.shared-board.is-compact {
  grid-template-columns: 1fr;
}

.shared-card {
  position: relative;
  display: grid;
  grid-template-rows: 1fr auto auto;
  gap: 14px;
  min-height: 252px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 28px;
  background:
    radial-gradient(circle at 25% 8%, rgba(147, 197, 253, 0.36), transparent 40%),
    radial-gradient(circle at 100% 82%, rgba(251, 207, 232, 0.28), transparent 44%),
    rgba(255, 255, 255, 0.62);
  box-shadow: 0 18px 45px rgba(80, 113, 162, 0.14);
  cursor: default;
}

.shared-card:hover,
.shared-card.selected {
  border-color: rgba(47, 125, 245, 0.5);
  box-shadow: 0 22px 54px rgba(77, 124, 190, 0.2);
  transform: translateY(-2px);
}

.select-dot {
  position: absolute;
  top: 14px;
  right: 14px;
  z-index: 2;
  display: grid;
  width: 28px;
  height: 28px;
  place-items: center;
  border: 1px solid rgba(191, 208, 232, 0.8);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.78);
  color: #2f7df5;
}

.shared-cover {
  position: relative;
  display: grid;
  min-height: 142px;
  place-items: center;
  overflow: hidden;
  border: 0;
  border-radius: 24px;
  background:
    radial-gradient(circle at 20% 10%, rgba(255, 255, 255, 0.72), transparent 38%),
    linear-gradient(135deg, rgba(221, 241, 255, 0.88), rgba(252, 231, 243, 0.74));
  color: #2f7df5;
  cursor: pointer;
}

.shared-cover .el-icon {
  width: 58px;
  height: 58px;
}

.permission-chip {
  position: absolute;
  top: 16px;
  left: 16px;
  border-radius: 999px;
  padding: 6px 10px;
  background: rgba(47, 125, 245, 0.92);
  color: #fff;
  font-size: 12px;
  font-weight: 900;
}

.shared-info {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.shared-info strong,
.shared-info span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shared-info strong {
  color: #10213f;
  font-size: 17px;
  font-weight: 900;
}

.shared-info span {
  color: #61708a;
  font-size: 13px;
  font-weight: 760;
}

.card-actions {
  display: flex;
  gap: 8px;
}

.card-actions button {
  display: grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border-radius: 14px;
}

.shared-board.is-list .shared-card,
.shared-board.is-compact .shared-card {
  grid-template-columns: 90px minmax(0, 1fr) auto;
  grid-template-rows: auto;
  align-items: center;
  min-height: 126px;
}

.shared-board.is-list .shared-cover,
.shared-board.is-compact .shared-cover {
  min-height: 90px;
}

.shared-board.is-compact .shared-card {
  min-height: 92px;
  padding: 12px 16px;
}

.shared-board.is-compact .shared-cover {
  min-height: 66px;
}

.empty-state {
  grid-column: 1 / -1;
  display: grid;
  min-height: 420px;
  place-items: center;
  align-content: center;
  gap: 12px;
  color: #61708a;
  text-align: center;
  font-weight: 780;
}

.empty-visual {
  display: grid;
  width: 92px;
  height: 92px;
  place-items: center;
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.72);
  color: #2f7df5;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9), 0 18px 40px rgba(74, 116, 177, 0.14);
}

.empty-visual .el-icon {
  width: 48px;
  height: 48px;
}

.empty-state strong {
  color: #14233d;
  font-size: 22px;
}

.empty-state span {
  max-width: 520px;
  line-height: 1.8;
}

:deep(.shared-preview-dialog .el-dialog) {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 30px;
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.42), transparent 38%),
    radial-gradient(circle at 100% 20%, rgba(252, 231, 243, 0.46), transparent 40%),
    rgba(255, 255, 255, 0.94);
  box-shadow: 0 30px 80px rgba(73, 112, 160, 0.24);
  backdrop-filter: blur(22px);
}

:deep(.shared-preview-dialog .el-dialog__header) {
  margin: 0;
  padding: 22px 24px 12px;
}

:deep(.shared-preview-dialog .el-dialog__body) {
  padding: 0 24px 24px;
}

.preview-heading {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding-right: 44px;
}

.preview-heading p {
  margin: 0 0 6px;
  color: #2684e8;
  font-size: 12px;
  font-weight: 820;
  letter-spacing: 0.12em;
}

.preview-heading strong {
  color: #10213f;
  font-size: 20px;
}

.preview-heading span {
  color: #61708a;
  font-size: 13px;
  font-weight: 780;
}

.preview-close {
  position: absolute;
  top: -4px;
  right: 0;
  z-index: 5;
  display: inline-grid;
  width: 36px;
  height: 36px;
  place-items: center;
  border: 1px solid rgba(203, 219, 238, 0.86);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.9);
  color: #61708a;
  cursor: pointer;
}

.preview-close:hover {
  border-color: rgba(47, 125, 245, 0.38);
  background: rgba(235, 246, 255, 0.96);
  color: #2f7df5;
}

.preview-shell {
  display: grid;
  overflow: hidden;
  min-height: min(620px, calc(100vh - 210px));
  border: 1px solid rgba(219, 234, 249, 0.82);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.62);
}

.frame-preview {
  width: 100%;
  min-height: min(620px, calc(100vh - 210px));
  border: 0;
  background: #fff;
}

.preview-empty {
  display: grid;
  min-height: 320px;
  place-items: center;
  color: #657892;
  font-weight: 800;
}

@media (max-width: 860px) {
  .filter-panel,
  .drive-bar,
  .selection-strip {
    align-items: stretch;
    flex-direction: column;
  }

  .drive-actions,
  .selection-strip > div {
    width: 100%;
    flex-wrap: wrap;
  }

  .shared-board.is-list .shared-card,
  .shared-board.is-compact .shared-card {
    grid-template-columns: 1fr;
  }
}
</style>

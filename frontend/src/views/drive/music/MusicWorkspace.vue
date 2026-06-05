<template>
  <main class="music-page" @click="closeFloatingLayers" @contextmenu="showMusicContextMenu">
    <input ref="fileInputRef" class="hidden-file-input" type="file" multiple accept="audio/*" @change="handleFileSelect" />
    <input
      ref="folderInputRef"
      class="hidden-file-input"
      type="file"
      multiple
      webkitdirectory
      directory
      @change="handleFileSelect"
    />

    <section class="filter-panel glass-panel">
      <label class="search-field">
        <el-icon><Search /></el-icon>
        <input v-model="keyword" type="search" placeholder="筛选音乐名称" />
      </label>
      <div class="drive-actions">
        <button class="tool-button" type="button" title="刷新" @click="refreshMusic">
          <el-icon><Refresh /></el-icon>
        </button>
        <button class="tool-button" type="button" title="上传音乐" @click="triggerFileUpload">
          <el-icon><Upload /></el-icon>
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
              <button type="button" :class="{ active: viewMode === 'album' }" @click="viewMode = 'album'">
                <el-icon><Headset /></el-icon>
                专辑
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
          <option value="recent">最近更新</option>
          <option value="name">名称</option>
          <option value="size">大小</option>
        </select>
        <span class="result-count">{{ filteredTracks.length }} / {{ tracks.length }}</span>
      </div>
    </section>

    <section class="drive-bar glass-panel">
      <div class="breadcrumb">
        <el-icon><Headset /></el-icon>
        <span>音乐</span>
      </div>
    </section>

    <section v-if="selectedTracks.length" class="selection-strip glass-panel">
      <span>已选择 {{ selectedTracks.length }} 首 · {{ formatFileSize(selectedSize) }}</span>
      <div>
        <button v-if="selectedTracks.length === 1" type="button" @click="openRenameDialog">重命名</button>
        <button type="button" @click="downloadSelectedTracks">下载</button>
        <button v-if="selectedTracks.length === 1" type="button" @click="openShareDialog">分享</button>
        <button type="button" class="danger" @click="deleteSelectedTracks">删除</button>
        <button type="button" @click="clearSelection">取消</button>
      </div>
    </section>

    <section
      class="music-board glass-panel"
      :class="`is-${viewMode}`"
      :style="{ '--card-size': `${cardSize}px` }"
    >
      <div v-if="loading" class="empty-state">正在加载音乐...</div>
      <div v-else-if="!filteredTracks.length" class="empty-state">
        <el-icon><Headset /></el-icon>
        <strong>这里还没有音乐</strong>
      </div>
      <article
        v-for="track in filteredTracks"
        v-else
        :key="track.id"
        class="music-card"
        :class="{ selected: selectedIds.includes(track.id), playing: previewState.id === track.id && previewVisible }"
        @click.stop="toggleSelect(track)"
        @dblclick.stop="previewTrack(track)"
        @contextmenu.prevent.stop="showMusicFileContextMenu($event, track)"
      >
        <button class="select-dot" type="button" :aria-label="`选择 ${track.name}`" @click.stop="toggleSelect(track)">
          <el-icon v-if="selectedIds.includes(track.id)"><Check /></el-icon>
        </button>
        <button class="album-art" type="button" @click.stop="toggleSelect(track)" @dblclick.stop="previewTrack(track)">
          <span class="vinyl-ring"></span>
          <el-icon><Headset /></el-icon>
        </button>
        <div class="track-info">
          <strong :title="track.name">{{ track.name }}</strong>
          <span>{{ formatFileSize(track.size || 0) }} · {{ formatDate(track.updated_at) }}</span>
        </div>
        <div class="card-actions">
          <button type="button" title="播放" @click.stop="previewTrack(track)">
            <el-icon><VideoPlay /></el-icon>
          </button>
          <button type="button" title="下载" @click.stop="downloadOneTrack(track)">
            <el-icon><Download /></el-icon>
          </button>
          <button type="button" title="更多" @click.stop="showMusicFileContextMenu($event, track)">
            <el-icon><MoreFilled /></el-icon>
          </button>
        </div>
      </article>
    </section>

    <MusicFileContextMenu
      :visible="musicFileMenuVisible"
      :x="musicFileMenuPosition.x"
      :y="musicFileMenuPosition.y"
      @rename="renameContextTrack"
      @download="downloadContextTrack"
      @share="shareContextTrack"
      @history="showContextPlaceholder('版本历史')"
      @collaboration="showContextPlaceholder('协作管理')"
      @move="openFolderOperation('move')"
      @copy="openFolderOperation('copy')"
      @delete="deleteContextTrack"
    />

    <ImageContextMenu
      :visible="contextMenuVisible"
      :x="contextMenuPosition.x"
      :y="contextMenuPosition.y"
      @upload-file="triggerFileUpload"
      @upload-folder="triggerFolderUpload"
      @upload-clipboard="uploadFromClipboard"
      @offline-download="openOfflineDownload"
      @create-folder="openCreateFolderDialog"
      @create-file="createWorkspaceFile"
      @refresh="refreshFromContextMenu"
    />

    <CreateFolderDialog v-model:visible="createFolderVisible" @confirm="handleCreateFolder" />
    <FolderSelectDialog
      v-model:visible="folderSelectVisible"
      :title="folderSelectMode === 'move' ? '移动到' : '复制到'"
      @confirm="confirmFolderOperation"
    />
    <RenameDialog
      v-model:visible="renameDialogVisible"
      :item="selectedTracks[0] || null"
      @confirm="confirmRename"
    />
    <ShareDialog
      v-model:visible="shareDialogVisible"
      :file-ids="selectedTracks.map((track) => track.id.toString())"
    />

    <el-dialog
      v-model="previewVisible"
      width="min(760px, 92vw)"
      top="12vh"
      class="music-preview-dialog"
      :show-close="false"
      destroy-on-close
      @closed="clearPreview"
    >
      <template #header>
        <div class="preview-heading">
          <div>
            <p>音乐预览</p>
            <strong>{{ previewState.title }}</strong>
          </div>
          <span>{{ previewState.size > 0 ? formatFileSize(previewState.size) : '' }}</span>
          <button class="preview-close" type="button" title="关闭预览" aria-label="关闭预览" @click.stop="closePreview">
            <el-icon><Close /></el-icon>
          </button>
        </div>
      </template>
      <div class="audio-shell">
        <div class="audio-cover">
          <span class="cover-glow"></span>
          <el-icon><Headset /></el-icon>
        </div>
        <audio v-if="previewState.url" class="audio-preview" :src="previewState.url" controls autoplay></audio>
        <div v-else class="preview-empty">这个音乐暂时无法预览</div>
      </div>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import {
  Check,
  Close,
  Download,
  Grid,
  Headset,
  List,
  MoreFilled,
  Refresh,
  Search,
  Upload,
  VideoPlay,
} from '@element-plus/icons-vue';
import { ElDialog, ElIcon, ElMessage, ElMessageBox } from 'element-plus';
import {
  copyFile as apiCopyFile,
  createFile as apiCreateFile,
  deleteFile as apiDeleteFile,
  downloadFile as apiDownloadFile,
  getAuthenticatedFileDownloadUrl,
  moveFile as apiMoveFile,
  renameFile as apiRenameFile,
} from '@/api/file';
import { createFolder as apiCreateFolder } from '@/api/folder';
import CreateFolderDialog from '@/components/CreateFolderDialog/index.vue';
import FolderSelectDialog from '@/components/FolderSelectDialog/index.vue';
import RenameDialog from '@/components/RenameDialog/index.vue';
import ShareDialog from '@/components/ShareDialog/index.vue';
import { useUploadStore } from '@/stores/upload';
import type { FileItem } from '@/types/file';
import { formatFileSize } from '@/utils/format';
import ImageContextMenu from '../images/components/ImageContextMenu.vue';
import type { WorkspaceFileKind } from '../images/useImagesWorkspace';
import MusicFileContextMenu from './components/MusicFileContextMenu.vue';
import { useMusicWorkspace } from './useMusicWorkspace';

const router = useRouter();
const uploadStore = useUploadStore();
const fileInputRef = ref<HTMLInputElement | null>(null);
const folderInputRef = ref<HTMLInputElement | null>(null);
const viewPanelVisible = ref(false);
const createFolderVisible = ref(false);
const renameDialogVisible = ref(false);
const shareDialogVisible = ref(false);
const previewVisible = ref(false);
const contextMenuVisible = ref(false);
const contextMenuPosition = reactive({ x: 0, y: 0 });
const musicFileMenuVisible = ref(false);
const musicFileMenuPosition = reactive({ x: 0, y: 0 });
const musicFileMenuTarget = ref<FileItem | null>(null);
const folderSelectVisible = ref(false);
const folderSelectMode = ref<'move' | 'copy'>('move');
const folderSelectTargets = ref<FileItem[]>([]);
const refreshedUploadTaskIds = new Set<string>();

const {
  cardSize,
  clearSelection,
  filteredTracks,
  keyword,
  loading,
  refreshMusic,
  removeTracksByIds,
  selectedIds,
  selectedSize,
  selectedTracks,
  selectOnly,
  sortMode,
  toggleSelect,
  tracks,
  upsertTracks,
  viewMode,
} = useMusicWorkspace();

const previewState = reactive({
  id: 0,
  title: '',
  size: 0,
  url: '',
});

function closeViewPanel() {
  viewPanelVisible.value = false;
}

function closeContextMenus() {
  contextMenuVisible.value = false;
  musicFileMenuVisible.value = false;
}

function closeFloatingLayers() {
  closeViewPanel();
  closeContextMenus();
}

function triggerFileUpload() {
  closeContextMenus();
  fileInputRef.value?.click();
}

function triggerFolderUpload() {
  closeContextMenus();
  folderInputRef.value?.click();
}

function audioUrl(file: FileItem) {
  const url = new URL(getAuthenticatedFileDownloadUrl(file.id, true));
  url.searchParams.set('v', file.updated_at || String(file.id));
  return url.toString();
}

function formatDate(value: string) {
  if (!value) return '未知时间';
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit' }).format(new Date(value));
}

function clearPreview() {
  previewState.id = 0;
  previewState.title = '';
  previewState.size = 0;
  previewState.url = '';
}

function closePreview() {
  previewVisible.value = false;
}

function previewTrack(file: FileItem) {
  selectOnly(file);
  closeContextMenus();
  previewState.id = file.id;
  previewState.title = file.name;
  previewState.size = file.size || 0;
  previewState.url = audioUrl(file);
  previewVisible.value = true;
}

function showMusicFileContextMenu(event: MouseEvent, file: FileItem) {
  event.preventDefault();
  event.stopPropagation();
  closeViewPanel();
  selectOnly(file);
  musicFileMenuTarget.value = file;
  contextMenuVisible.value = false;
  musicFileMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 220));
  musicFileMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 430));
  musicFileMenuVisible.value = true;
}

function getContextTrack() {
  return musicFileMenuTarget.value;
}

function renameContextTrack() {
  const track = getContextTrack();
  closeContextMenus();
  if (!track) return;
  selectOnly(track);
  renameDialogVisible.value = true;
}

async function downloadContextTrack() {
  const track = getContextTrack();
  closeContextMenus();
  if (!track) return;
  await downloadOneTrack(track);
}

function shareContextTrack() {
  const track = getContextTrack();
  closeContextMenus();
  if (!track) return;
  selectOnly(track);
  shareDialogVisible.value = true;
}

function showContextPlaceholder(label: string) {
  const track = getContextTrack();
  closeContextMenus();
  if (!track) return;
  selectOnly(track);
  ElMessage.info(`${label}功能即将开放`);
}

function openFolderOperation(mode: 'move' | 'copy') {
  const track = getContextTrack();
  closeContextMenus();
  if (!track) return;
  selectOnly(track);
  folderSelectMode.value = mode;
  folderSelectTargets.value = [track];
  folderSelectVisible.value = true;
}

async function confirmFolderOperation(folderId: number | null) {
  const targets = folderSelectTargets.value;
  if (!targets.length) return;
  try {
    for (const track of targets) {
      if (folderSelectMode.value === 'move') {
        await apiMoveFile(track.id, folderId);
      } else {
        await apiCopyFile(track.id, folderId);
      }
    }
    folderSelectVisible.value = false;
    folderSelectTargets.value = [];
    await refreshMusic();
    ElMessage.success(folderSelectMode.value === 'move' ? '移动成功' : '复制成功');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '操作失败');
  }
}

function openRenameDialog() {
  if (!selectedTracks.value.length) return;
  renameDialogVisible.value = true;
}

function openShareDialog() {
  if (!selectedTracks.value.length) return;
  shareDialogVisible.value = true;
}

async function confirmRename(name: string) {
  const track = selectedTracks.value[0];
  if (!track) return;
  try {
    await apiRenameFile(track.id, name);
    renameDialogVisible.value = false;
    await refreshMusic();
    ElMessage.success('重命名成功');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '重命名失败');
  }
}

async function downloadOneTrack(track: FileItem) {
  try {
    const blob = await apiDownloadFile(track.id);
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = track.name;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(link.href);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载失败');
  }
}

async function downloadSelectedTracks() {
  for (const track of selectedTracks.value) {
    await downloadOneTrack(track);
  }
}

async function deleteContextTrack() {
  const track = getContextTrack();
  closeContextMenus();
  if (!track) return;
  selectOnly(track);
  await deleteSelectedTracks();
}

async function deleteSelectedTracks() {
  const targets = [...selectedTracks.value];
  if (!targets.length) return;

  try {
    await ElMessageBox.confirm(
      `确定删除 ${targets.length} 首音乐吗？此操作不可撤销。`,
      '删除音乐',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        confirmButtonClass: 'el-button--danger',
      },
    );

    for (const track of targets) {
      await apiDeleteFile(track.id);
    }
    removeTracksByIds(targets.map((track) => track.id));
    clearSelection();
    void refreshMusic();
    ElMessage.success('删除成功');
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

async function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  const files = Array.from(target.files || []);
  target.value = '';
  if (!files.length) return;

  try {
    for (const file of files) {
      await uploadStore.addTask(file);
    }
    scheduleMusicRefresh();
    ElMessage.success(`已添加 ${files.length} 个音乐到上传队列`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '添加上传任务失败');
  }
}

async function uploadFromClipboard() {
  closeContextMenus();
  try {
    if (!navigator.clipboard?.read) {
      ElMessage.warning('当前浏览器不支持读取剪贴板文件');
      return;
    }

    const clipboardItems = await navigator.clipboard.read();
    const files: File[] = [];
    for (const item of clipboardItems) {
      for (const type of item.types) {
        const blob = await item.getType(type);
        if (!blob.size) continue;
        files.push(new File([blob], uniqueTrackName(`剪贴板音乐.${clipboardExtension(type)}`), { type }));
      }
    }

    if (!files.length) {
      ElMessage.warning('剪贴板中没有可上传的文件内容');
      return;
    }

    for (const file of files) {
      await uploadStore.addTask(file);
    }
    scheduleMusicRefresh();
    ElMessage.success(`已添加 ${files.length} 个剪贴板文件到上传队列`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '读取剪贴板失败');
  }
}

function clipboardExtension(type: string) {
  if (type.includes('mpeg')) return 'mp3';
  if (type.includes('wav')) return 'wav';
  if (type.includes('ogg')) return 'ogg';
  if (type.includes('flac')) return 'flac';
  if (type.includes('mp4')) return 'm4a';
  if (type.includes('plain')) return 'txt';
  return 'bin';
}

function uniqueTrackName(baseName: string) {
  const existingNames = new Set(tracks.value.map((item) => item.name));
  if (!existingNames.has(baseName)) return baseName;

  const dotIndex = baseName.lastIndexOf('.');
  const stem = dotIndex > 0 ? baseName.slice(0, dotIndex) : baseName;
  const extension = dotIndex > 0 ? baseName.slice(dotIndex) : '';
  let index = 2;
  let nextName = `${stem} ${index}${extension}`;
  while (existingNames.has(nextName)) {
    index += 1;
    nextName = `${stem} ${index}${extension}`;
  }
  return nextName;
}

async function openOfflineDownload() {
  closeContextMenus();
  await router.push('/drive/offline-downloads');
}

function openCreateFolderDialog() {
  closeContextMenus();
  createFolderVisible.value = true;
}

async function handleCreateFolder(name: string) {
  try {
    await apiCreateFolder(name, null);
    createFolderVisible.value = false;
    ElMessage.success('文件夹已创建');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建文件夹失败');
  }
}

async function createWorkspaceFile(kind: WorkspaceFileKind) {
  closeContextMenus();
  const defaultNames: Record<WorkspaceFileKind, string> = {
    file: '新建文件.txt',
    markdown: '新建 Markdown.md',
    text: '新建文本.txt',
    drawio: '新建图表.drawio',
    dwb: '新建白板.dwb',
    excalidraw: '新建 Excalidraw.excalidraw',
  };

  try {
    await apiCreateFile(kind, uniqueTrackName(defaultNames[kind]), null);
    await refreshMusic();
    ElMessage.success('文件已创建');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建文件失败');
  }
}

async function refreshFromContextMenu() {
  closeContextMenus();
  await refreshMusic();
}

function showMusicContextMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  if (
    target?.closest('.music-card') ||
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
  musicFileMenuVisible.value = false;
  contextMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 330));
  contextMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 560));
  contextMenuVisible.value = true;
}

function isAudioUploadTask(file: File) {
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return file.type.startsWith('audio/') || [
    'mp3',
    'wav',
    'flac',
    'aac',
    'ogg',
    'oga',
    'm4a',
    'wma',
    'ape',
    'opus',
    'amr',
    'mid',
    'midi',
  ].includes(extension);
}

function scheduleMusicRefresh() {
  [250, 900, 2200].forEach((delay) => {
    window.setTimeout(() => {
      void refreshMusic();
    }, delay);
  });
}

onMounted(() => {
  document.addEventListener('click', closeContextMenus);
  window.addEventListener('resize', closeFloatingLayers);
});

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenus);
  window.removeEventListener('resize', closeFloatingLayers);
});

watch(
  () => uploadStore.tasks.map((task) => `${task.id}:${task.status}`).join('|'),
  () => {
    const completedAudioTasks = uploadStore.tasks.filter(
      (task) => task.status === 'completed' && isAudioUploadTask(task.file) && !refreshedUploadTaskIds.has(task.id),
    );
    if (!completedAudioTasks.length) return;

    completedAudioTasks.forEach((task) => refreshedUploadTaskIds.add(task.id));
    upsertTracks(completedAudioTasks.map((task) => task.result).filter((file): file is FileItem => Boolean(file)));
    scheduleMusicRefresh();
  },
);
</script>

<style scoped>
.music-page {
  display: grid;
  gap: 12px;
  min-height: calc(100vh - 96px);
  padding: 0 4px 24px 22px;
  color: #10213f;
}

.hidden-file-input {
  display: none;
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
.music-board {
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
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
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

.selection-strip .danger {
  color: #ef4444;
}

.music-board {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(var(--card-size), 100%), 1fr));
  align-content: start;
  gap: 18px;
  min-height: 560px;
  border-radius: 30px;
  padding: 28px;
}

.music-board.is-list {
  grid-template-columns: 1fr;
}

.music-board.is-album {
  grid-template-columns: repeat(auto-fill, minmax(min(calc(var(--card-size) + 80px), 100%), 1fr));
}

.music-card {
  position: relative;
  display: grid;
  grid-template-rows: 1fr auto auto;
  gap: 14px;
  min-height: 252px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 28px;
  background:
    radial-gradient(circle at 25% 8%, rgba(147, 197, 253, 0.42), transparent 40%),
    radial-gradient(circle at 100% 82%, rgba(251, 207, 232, 0.34), transparent 44%),
    rgba(255, 255, 255, 0.62);
  box-shadow: 0 18px 45px rgba(80, 113, 162, 0.14);
  cursor: default;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.music-card:hover,
.music-card.selected,
.music-card.playing {
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

.album-art {
  position: relative;
  display: grid;
  min-height: 142px;
  place-items: center;
  overflow: hidden;
  border: 0;
  border-radius: 24px;
  background:
    radial-gradient(circle, rgba(255, 255, 255, 0.92) 0 11%, rgba(47, 125, 245, 0.88) 12% 14%, transparent 15%),
    conic-gradient(from 140deg, #93c5fd, #fbcfe8, #a7f3d0, #bfdbfe, #93c5fd);
  color: #ffffff;
  cursor: pointer;
}

.album-art .el-icon {
  z-index: 1;
  width: 56px;
  height: 56px;
  filter: drop-shadow(0 8px 16px rgba(15, 23, 42, 0.22));
}

.vinyl-ring {
  position: absolute;
  inset: 18px;
  border: 18px solid rgba(255, 255, 255, 0.36);
  border-radius: 999px;
}

.track-info {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.track-info strong {
  overflow: hidden;
  color: #10213f;
  font-size: 17px;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.track-info span {
  overflow: hidden;
  color: #61708a;
  font-size: 13px;
  font-weight: 760;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.music-board.is-list .music-card {
  grid-template-columns: 90px minmax(0, 1fr) auto;
  grid-template-rows: auto;
  align-items: center;
  min-height: 126px;
}

.music-board.is-list .album-art {
  min-height: 90px;
}

.empty-state {
  grid-column: 1 / -1;
  display: grid;
  min-height: 360px;
  place-items: center;
  align-content: center;
  gap: 12px;
  color: #61708a;
  text-align: center;
  font-weight: 780;
}

.empty-state .el-icon {
  width: 64px;
  height: 64px;
  color: #2f7df5;
}

.empty-state strong {
  color: #14233d;
  font-size: 22px;
}

:deep(.music-preview-dialog .el-dialog) {
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

:deep(.music-preview-dialog .el-dialog__header) {
  margin: 0;
  padding: 22px 24px 12px;
}

:deep(.music-preview-dialog .el-dialog__body) {
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

.audio-shell {
  display: grid;
  gap: 22px;
  justify-items: center;
  padding: 24px;
  border: 1px solid rgba(219, 234, 249, 0.82);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.52);
}

.audio-cover {
  position: relative;
  display: grid;
  width: min(280px, 70vw);
  aspect-ratio: 1;
  place-items: center;
  overflow: hidden;
  border-radius: 42px;
  background:
    radial-gradient(circle, rgba(255, 255, 255, 0.92) 0 11%, rgba(47, 125, 245, 0.9) 12% 14%, transparent 15%),
    conic-gradient(from 140deg, #93c5fd, #fbcfe8, #a7f3d0, #bfdbfe, #93c5fd);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72), 0 24px 52px rgba(74, 116, 177, 0.18);
  color: #ffffff;
}

.audio-cover .el-icon {
  z-index: 1;
  width: 76px;
  height: 76px;
}

.cover-glow {
  position: absolute;
  inset: 32px;
  border: 24px solid rgba(255, 255, 255, 0.36);
  border-radius: 999px;
}

.audio-preview {
  width: min(560px, 100%);
}

.preview-empty {
  min-height: 80px;
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

  .music-board.is-list .music-card {
    grid-template-columns: 1fr;
  }
}
</style>

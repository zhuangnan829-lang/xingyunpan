<template>
  <main class="videos-page" @click="closeFloatingLayers" @contextmenu="showVideosContextMenu">
    <input ref="fileInputRef" class="hidden-file-input" type="file" multiple accept="video/*" @change="handleFileSelect" />
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
        <input v-model="keyword" type="search" placeholder="筛选视频名称" />
      </label>
      <div class="drive-actions">
        <button class="tool-button" type="button" title="刷新" @click="refreshVideos">
          <el-icon><Refresh /></el-icon>
        </button>
        <button class="tool-button" type="button" title="上传视频" @click="triggerFileUpload">
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
              <button type="button" :class="{ active: viewMode === 'cinema' }" @click="viewMode = 'cinema'">
                <el-icon><Picture /></el-icon>
                画廊
              </button>
            </div>
            <p>缩略图</p>
            <div class="segmented two">
              <button type="button" :class="{ active: posterEnabled }" @click="posterEnabled = true">
                <el-icon><PictureFilled /></el-icon>
                开启
              </button>
              <button type="button" :class="{ active: !posterEnabled }" @click="posterEnabled = false">
                <el-icon><Close /></el-icon>
                关闭
              </button>
            </div>
            <label class="slider-row">
              <span>分页大小</span>
              <input v-model.number="tileSize" type="range" min="180" max="420" step="10" />
            </label>
            <div class="slider-scale">
              <span>180</span>
              <span>420</span>
            </div>
          </div>
        </div>
        <select v-model="sortMode" class="sort-select" aria-label="排序">
          <option value="recent">最近更新</option>
          <option value="name">名称</option>
          <option value="size">大小</option>
        </select>
        <span class="result-count">{{ filteredVideos.length }} / {{ videos.length }}</span>
      </div>
    </section>

    <section class="drive-bar glass-panel">
      <div class="breadcrumb">
        <el-icon><VideoPlay /></el-icon>
        <span>视频</span>
      </div>
    </section>

    <section v-if="selectedVideos.length" class="selection-strip glass-panel">
      <span>已选择 {{ selectedVideos.length }} 个 · {{ formatFileSize(selectedSize) }}</span>
      <div>
        <button v-if="selectedVideos.length === 1" type="button" @click="openRenameDialog">重命名</button>
        <button type="button" @click="downloadSelectedVideos">下载</button>
        <button v-if="selectedVideos.length === 1" type="button" @click="openShareDialog">分享</button>
        <button type="button" class="danger" @click="deleteSelectedVideos">删除</button>
        <button type="button" @click="clearSelection">取消</button>
      </div>
    </section>


    <section
      class="video-board glass-panel"
      :class="[`is-${viewMode}`, { 'poster-disabled': !posterEnabled }]"
      :style="{ '--tile-size': `${tileSize}px` }"
    >
      <div v-if="loading" class="empty-state">正在加载视频...</div>
      <div v-else-if="!filteredVideos.length" class="empty-state">
        <el-icon><VideoCameraFilled /></el-icon>
        <strong>这里还没有视频</strong>
      </div>
      <article
        v-for="video in filteredVideos"
        v-else
        :key="video.id"
        class="video-card"
        :class="{ selected: selectedIds.includes(video.id) }"
        @click.stop="toggleSelect(video)"
        @dblclick.stop="previewVideo(video)"
        @contextmenu.prevent.stop="showVideoFileContextMenu($event, video)"
      >
        <button class="select-dot" type="button" :aria-label="`选择 ${video.name}`" @click.stop="toggleSelect(video)">
          <el-icon v-if="selectedIds.includes(video.id)"><Check /></el-icon>
        </button>
        <button class="poster" type="button" @click.stop="toggleSelect(video)" @dblclick.stop="previewVideo(video)">
          <video v-if="posterEnabled" muted preload="metadata" :src="videoUrl(video)"></video>
          <el-icon class="poster-icon"><VideoCameraFilled /></el-icon>
          <span class="play-chip"><el-icon><VideoPlay /></el-icon></span>
        </button>
        <div class="video-info">
          <strong :title="video.name">{{ video.name }}</strong>
          <span>{{ formatFileSize(video.size || 0) }} 路 {{ formatDate(video.updated_at) }}</span>
        </div>
        <div class="card-actions">
          <button type="button" title="预览" @click.stop="previewVideo(video)">
            <el-icon><VideoPlay /></el-icon>
          </button>
          <button type="button" title="下载" @click.stop="downloadOneVideo(video)">
            <el-icon><Download /></el-icon>
          </button>
          <button type="button" title="更多" @click.stop="showVideoFileContextMenu($event, video)">
            <el-icon><MoreFilled /></el-icon>
          </button>
        </div>
      </article>
    </section>

    <VideoFileContextMenu
      :visible="videoFileMenuVisible"
      :x="videoFileMenuPosition.x"
      :y="videoFileMenuPosition.y"
      @rename="renameContextVideo"
      @download="downloadContextVideo"
      @share="shareContextVideo"
      @history="showContextPlaceholder('版本历史')"
      @collaboration="showContextPlaceholder('协作管理')"
      @move="openFolderOperation('move')"
      @copy="openFolderOperation('copy')"
      @delete="deleteContextVideo"
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
      :item="selectedVideos[0] || null"
      @confirm="confirmRename"
    />
    <ShareDialog
      v-model:visible="shareDialogVisible"
      :file-ids="selectedVideos.map((video) => video.id.toString())"
    />

    <el-dialog
      v-model="previewVisible"
      width="min(1120px, 92vw)"
      top="5vh"
      class="video-preview-dialog"
      :show-close="false"
      destroy-on-close
      @closed="clearPreview"
    >
      <template #header>
        <div class="preview-heading">
          <div>
            <p>视频预览</p>
            <strong>{{ previewState.title }}</strong>
          </div>
          <span>{{ previewState.size > 0 ? formatFileSize(previewState.size) : '' }}</span>
          <button class="preview-close" type="button" title="关闭预览" aria-label="关闭预览" @click.stop="closePreview">
            <el-icon><Close /></el-icon>
          </button>
        </div>
      </template>
      <div class="preview-shell">
        <video v-if="previewState.url" class="media-preview" :src="previewState.url" controls autoplay></video>
        <div v-else class="preview-empty">这个视频暂时无法预览</div>
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
  List,
  MoreFilled,
  Picture,
  PictureFilled,
  Refresh,
  Search,
  Upload,
  VideoCameraFilled,
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
import VideoFileContextMenu from './components/VideoFileContextMenu.vue';
import { useVideosWorkspace } from './useVideosWorkspace';

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
const videoFileMenuVisible = ref(false);
const videoFileMenuPosition = reactive({ x: 0, y: 0 });
const videoFileMenuTarget = ref<FileItem | null>(null);
const folderSelectVisible = ref(false);
const folderSelectMode = ref<'move' | 'copy'>('move');
const folderSelectTargets = ref<FileItem[]>([]);
const refreshedUploadTaskIds = new Set<string>();

const {
  clearSelection,
  filteredVideos,
  keyword,
  loading,
  posterEnabled,
  refreshVideos,
  removeVideosByIds,
  selectedIds,
  selectedSize,
  selectedVideos,
  selectOnly,
  sortMode,
  tileSize,
  toggleSelect,
  upsertVideos,
  videos,
  viewMode,
} = useVideosWorkspace();

const previewState = reactive({
  title: '',
  size: 0,
  url: '',
});

function closeViewPanel() {
  viewPanelVisible.value = false;
}

function closeContextMenus() {
  contextMenuVisible.value = false;
  videoFileMenuVisible.value = false;
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

function videoUrl(file: FileItem) {
  const url = new URL(getAuthenticatedFileDownloadUrl(file.id, true));
  url.searchParams.set('v', file.updated_at || String(file.id));
  return url.toString();
}

function formatDate(value: string) {
  if (!value) return '未知时间';
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit' }).format(new Date(value));
}

function clearPreview() {
  previewState.title = '';
  previewState.size = 0;
  previewState.url = '';
}

function closePreview() {
  previewVisible.value = false;
}

function previewVideo(file: FileItem) {
  selectOnly(file);
  closeContextMenus();
  previewState.title = file.name;
  previewState.size = file.size || 0;
  previewState.url = videoUrl(file);
  previewVisible.value = true;
}

function showVideoFileContextMenu(event: MouseEvent, file: FileItem) {
  event.preventDefault();
  event.stopPropagation();
  closeViewPanel();
  selectOnly(file);
  videoFileMenuTarget.value = file;
  contextMenuVisible.value = false;
  videoFileMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 220));
  videoFileMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 430));
  videoFileMenuVisible.value = true;
}

function getContextVideo() {
  return videoFileMenuTarget.value;
}

function renameContextVideo() {
  const video = getContextVideo();
  closeContextMenus();
  if (!video) return;
  selectOnly(video);
  openRenameDialog();
}

async function downloadContextVideo() {
  const video = getContextVideo();
  closeContextMenus();
  if (!video) return;
  await downloadOneVideo(video);
}

function shareContextVideo() {
  const video = getContextVideo();
  closeContextMenus();
  if (!video) return;
  selectOnly(video);
  openShareDialog();
}

function deleteContextVideo() {
  const video = getContextVideo();
  closeContextMenus();
  if (!video) return;
  selectOnly(video);
  void deleteSelectedVideos();
}

function showContextPlaceholder(label: string) {
  closeContextMenus();
  ElMessage.info(`${label}功能待接入`);
}

function openFolderOperation(mode: 'move' | 'copy') {
  const video = getContextVideo();
  closeContextMenus();
  if (!video) return;
  selectOnly(video);
  folderSelectMode.value = mode;
  folderSelectTargets.value = [video];
  folderSelectVisible.value = true;
}

async function confirmFolderOperation(folderId: number | null) {
  const targets = [...folderSelectTargets.value];
  if (!targets.length) return;

  try {
    for (const video of targets) {
      if (folderSelectMode.value === 'move') {
        await apiMoveFile(video.id, folderId);
      } else {
        await apiCopyFile(video.id, folderId);
      }
    }

    folderSelectTargets.value = [];
    await refreshVideos();
    clearSelection();
    ElMessage.success(folderSelectMode.value === 'move' ? '移动成功' : '复制成功');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : folderSelectMode.value === 'move' ? '移动失败' : '复制失败');
  }
}

function showVideosContextMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;

  if (
    target?.closest('.video-card') ||
    target?.closest('.el-dialog') ||
    target?.closest('.el-overlay') ||
    target?.closest('input') ||
    target?.closest('select') ||
    target?.closest('button')
  ) {
    return;
  }

  event.preventDefault();
  event.stopPropagation();
  closeViewPanel();
  clearSelection();
  videoFileMenuVisible.value = false;
  contextMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 450));
  contextMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 470));
  contextMenuVisible.value = true;
}

function openRenameDialog() {
  if (selectedVideos.value.length !== 1) return;
  renameDialogVisible.value = true;
}

function openShareDialog() {
  if (selectedVideos.value.length !== 1) return;
  shareDialogVisible.value = true;
}

async function confirmRename(name: string) {
  const video = selectedVideos.value[0];
  if (!video) return;

  try {
    await apiRenameFile(video.id, name);
    renameDialogVisible.value = false;
    await refreshVideos();
    ElMessage.success('重命名成功');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '重命名失败');
  }
}

async function downloadOneVideo(video: FileItem) {
  try {
    const blob = await apiDownloadFile(video.id);
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = video.name;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载失败');
  }
}

async function downloadSelectedVideos() {
  for (const video of selectedVideos.value) {
    await downloadOneVideo(video);
  }
}

async function deleteSelectedVideos() {
  const targets = [...selectedVideos.value];
  if (!targets.length) return;

  try {
    await ElMessageBox.confirm(
      targets.length === 1 ? `确定删除“${targets[0].name}”吗？` : `确定删除选中的 ${targets.length} 个视频吗？`,
      '删除视频',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
      },
    );

    for (const video of targets) {
      await apiDeleteFile(video.id);
    }
    removeVideosByIds(targets.map((video) => video.id));
    clearSelection();
    void refreshVideos();
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
    window.setTimeout(refreshVideos, 600);
    window.setTimeout(refreshVideos, 2200);
    ElMessage.success(`已添加 ${files.length} 个视频到上传队列`);
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
        files.push(new File([blob], uniqueVideoName(`剪贴板文件.${clipboardExtension(type)}`), { type }));
      }
    }

    if (!files.length) {
      ElMessage.warning('剪贴板中没有可上传的文件内容');
      return;
    }

    for (const file of files) {
      await uploadStore.addTask(file);
    }
    window.setTimeout(refreshVideos, 600);
    window.setTimeout(refreshVideos, 2200);
    ElMessage.success(`已添加 ${files.length} 个剪贴板文件到上传队列`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '读取剪贴板失败');
  }
}

function clipboardExtension(type: string) {
  if (type.includes('mp4')) return 'mp4';
  if (type.includes('webm')) return 'webm';
  if (type.includes('quicktime')) return 'mov';
  if (type.includes('mpeg')) return 'mpeg';
  if (type.includes('html')) return 'html';
  if (type.includes('plain')) return 'txt';
  return 'bin';
}

function uniqueVideoName(baseName: string) {
  const existingNames = new Set(videos.value.map((item) => item.name));
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
    await apiCreateFile(kind, uniqueVideoName(defaultNames[kind]), null);
    await refreshVideos();
    ElMessage.success('文件已创建');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建文件失败');
  }
}

async function refreshFromContextMenu() {
  closeContextMenus();
  await refreshVideos();
}

function isVideoUploadTask(file: File) {
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return file.type.startsWith('video/') || [
    'mp4',
    'mov',
    'mkv',
    'webm',
    'avi',
    'wmv',
    'flv',
    'm4v',
    'mpeg',
    'mpg',
    '3gp',
    '3g2',
    'ts',
    'mts',
    'm2ts',
    'rm',
    'rmvb',
    'vob',
    'ogv',
    'asf',
    'divx',
  ].includes(extension);
}

function scheduleVideoRefresh() {
  [250, 900, 2200].forEach((delay) => {
    window.setTimeout(() => {
      void refreshVideos();
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
    const completedVideoTasks = uploadStore.tasks.filter(
      (task) => task.status === 'completed' && isVideoUploadTask(task.file) && !refreshedUploadTaskIds.has(task.id),
    );
    if (!completedVideoTasks.length) return;

    completedVideoTasks.forEach((task) => refreshedUploadTaskIds.add(task.id));
    upsertVideos(completedVideoTasks.map((task) => task.result).filter((file): file is FileItem => Boolean(file)));
    scheduleVideoRefresh();
  },
);
</script>

<style scoped>
.videos-page {
  display: grid;
  gap: 12px;
  min-height: calc(100vh - 96px);
  padding: 0 4px 24px 22px;
  color: #10213f;
}

.hidden-file-input {
  position: fixed;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.glass-panel {
  border: 1px solid rgba(255, 255, 255, 0.82);
  background:
    radial-gradient(circle at 0% 0%, rgba(119, 205, 255, 0.15), transparent 36%),
    rgba(255, 255, 255, 0.58);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.94),
    0 18px 44px rgba(88, 128, 176, 0.11);
  backdrop-filter: blur(20px);
}

.drive-bar,
.filter-panel,
.selection-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  min-height: 58px;
  padding: 10px 16px;
  border-radius: 18px;
}

.filter-panel {
  position: relative;
  z-index: 30;
  min-height: 56px;
  padding: 8px 18px;
  border-radius: 18px;
  overflow: visible;
}

.drive-bar {
  position: relative;
  z-index: 10;
  min-height: 60px;
  justify-content: flex-start;
  padding: 0 24px;
  border-radius: 18px;
  background:
    radial-gradient(circle at 4% 0%, rgba(119, 205, 255, 0.2), transparent 38%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.68), rgba(247, 252, 255, 0.5));
}

.breadcrumb,
.drive-actions,
.selection-strip div,
.search-field,
.card-actions {
  display: flex;
  align-items: center;
}

.breadcrumb {
  gap: 10px;
  color: #13213a;
  font-size: 17px;
  font-weight: 850;
}

.breadcrumb .el-icon {
  width: 22px;
  height: 22px;
  color: #111827;
}

.drive-actions {
  position: relative;
  gap: 10px;
}

.tool-button,
.view-button,
.sort-select,
.selection-strip button,
.card-actions button {
  border: 1px solid rgba(255, 255, 255, 0.78);
  background: rgba(255, 255, 255, 0.68);
  color: #172846;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 10px 24px rgba(64, 119, 180, 0.1);
  cursor: pointer;
}

.tool-button,
.card-actions button {
  display: inline-grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border-radius: 12px;
}

.view-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 0 14px;
  border-radius: 12px;
  font-weight: 820;
}

.tool-button:hover,
.view-button:hover,
.view-button.active,
.card-actions button:hover,
.selection-strip button:hover {
  border-color: rgba(111, 194, 255, 0.82);
  background: linear-gradient(135deg, rgba(229, 244, 255, 0.95), rgba(255, 232, 241, 0.82));
  color: #1d70da;
}

.sort-select {
  min-height: 38px;
  padding: 0 12px;
  border-radius: 12px;
  outline: none;
  font-weight: 780;
}

.view-panel {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  z-index: 80;
  width: 332px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 20px;
  background:
    radial-gradient(circle at 0 0, rgba(117, 196, 255, 0.18), transparent 34%),
    rgba(255, 255, 255, 0.86);
  box-shadow: 0 24px 58px rgba(66, 92, 130, 0.24);
  backdrop-filter: blur(22px);
}

.view-panel p {
  margin: 0 0 10px;
  color: #68768a;
  font-size: 13px;
  font-weight: 820;
}

.segmented {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  overflow: hidden;
  margin-bottom: 16px;
  border: 1px solid rgba(221, 231, 243, 0.9);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
}

.segmented.two {
  grid-template-columns: repeat(2, 1fr);
}

.segmented button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-height: 44px;
  border: 0;
  background: transparent;
  color: #6b7280;
  font-weight: 820;
  cursor: pointer;
}

.segmented button.active {
  color: #1d70da;
  background: rgba(226, 241, 255, 0.96);
}

.slider-row {
  display: grid;
  gap: 8px;
  color: #68768a;
  font-size: 13px;
  font-weight: 820;
}

.slider-row input {
  accent-color: #2384e8;
}

.slider-scale {
  display: flex;
  justify-content: space-between;
  color: #7b8492;
  font-size: 12px;
  font-weight: 760;
}

.filter-panel > span,
.result-count,
.selection-strip > span {
  color: #61708a;
  font-size: 13px;
  font-weight: 800;
}

.selection-strip div {
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.selection-strip button {
  min-height: 38px;
  padding: 0 14px;
  border-radius: 12px;
  font-weight: 820;
}

.selection-strip button.danger:hover {
  color: #fff;
  background: linear-gradient(135deg, #ef4444, #f97316);
}

.search-field {
  flex: 1;
  gap: 10px;
  min-width: min(360px, 100%);
}

.search-field .el-icon {
  color: #2f7df5;
}

.search-field input {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: #172846;
  font-size: 15px;
  font-weight: 760;
}

.video-board {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(var(--tile-size), 100%), 1fr));
  align-content: start;
  gap: 18px;
  min-height: 380px;
  padding: 20px;
  border-radius: 24px;
}

.video-board.is-list {
  grid-template-columns: 1fr;
}

.video-board.is-cinema {
  grid-template-columns: repeat(auto-fill, minmax(min(calc(var(--tile-size) + 90px), 100%), 1fr));
}

.video-card {
  position: relative;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 14px;
  min-width: 0;
  padding: 10px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 20px;
  background:
    radial-gradient(circle at 100% 0%, rgba(120, 216, 255, 0.13), transparent 36%),
    rgba(255, 255, 255, 0.62);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 18px 38px rgba(79, 118, 168, 0.12);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.video-card:hover,
.video-card.selected {
  transform: translateY(-2px);
  border-color: rgba(74, 156, 255, 0.58);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 22px 46px rgba(63, 134, 214, 0.18);
}

.poster {
  position: relative;
  grid-column: 1 / -1;
  display: grid;
  overflow: hidden;
  width: 100%;
  aspect-ratio: 16 / 10;
  place-items: center;
  border: 0;
  border-radius: 16px;
  background:
    radial-gradient(circle at 28% 18%, rgba(137, 215, 255, 0.4), transparent 30%),
    radial-gradient(circle at 70% 20%, rgba(255, 193, 218, 0.4), transparent 26%),
    linear-gradient(135deg, #f8fbff, #eef7ff 50%, #fff2f7);
  cursor: pointer;
}

.poster video {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0.82;
}

.poster-icon {
  z-index: 1;
  width: 52px;
  height: 52px;
  color: #e00000;
  filter: drop-shadow(0 10px 22px rgba(224, 0, 0, 0.18));
}

.play-chip {
  position: absolute;
  right: 14px;
  bottom: 14px;
  z-index: 2;
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.82);
  color: #1d70da;
  box-shadow: 0 12px 24px rgba(30, 89, 150, 0.18);
}

.select-dot {
  position: absolute;
  top: 18px;
  left: 18px;
  z-index: 3;
  display: grid;
  width: 24px;
  height: 24px;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.85);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.76);
  color: #1d70da;
  cursor: pointer;
}

.video-info {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.video-info strong {
  overflow: hidden;
  color: #172846;
  font-size: 15px;
  font-weight: 840;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-info span {
  color: #6b7b92;
  font-size: 12px;
  font-weight: 760;
}

.card-actions {
  gap: 6px;
}

.card-actions button {
  width: 34px;
  height: 34px;
  border-radius: 11px;
}

.video-board.is-list .video-card {
  grid-template-columns: 196px minmax(0, 1fr) auto;
  align-items: center;
}

.video-board.is-list .poster {
  grid-column: auto;
}

.video-board.poster-disabled .poster video {
  display: none;
}

.empty-state {
  grid-column: 1 / -1;
  display: grid;
  min-height: 320px;
  place-items: center;
  align-content: center;
  gap: 12px;
  color: #61708a;
  text-align: center;
  font-weight: 780;
}

.empty-state .el-icon {
  width: 58px;
  height: 58px;
  color: #e00000;
}

.empty-state strong {
  color: #14233d;
  font-size: 20px;
}

.empty-state span {
  max-width: 520px;
  line-height: 1.8;
}

:deep(.video-preview-dialog .el-dialog) {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 28px;
  background:
    radial-gradient(circle at 100% 0%, rgba(116, 220, 255, 0.15), transparent 34%),
    rgba(255, 255, 255, 0.94);
  box-shadow: 0 30px 80px rgba(73, 112, 160, 0.24);
  backdrop-filter: blur(20px);
}

:deep(.video-preview-dialog .el-dialog__header) {
  margin: 0;
  padding: 22px 24px 12px;
}

:deep(.video-preview-dialog .el-dialog__body) {
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
  transition:
    color 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease,
    transform 0.18s ease;
}

.preview-close:hover {
  border-color: rgba(47, 125, 245, 0.38);
  background: rgba(235, 246, 255, 0.96);
  color: #2f7df5;
  transform: translateY(-1px);
}

.preview-shell {
  display: grid;
  overflow: hidden;
  min-height: min(560px, calc(100vh - 210px));
  border: 1px solid rgba(219, 234, 249, 0.82);
  border-radius: 22px;
  background: rgba(15, 23, 42, 0.92);
}

.media-preview {
  display: block;
  width: 100%;
  max-height: 72vh;
  background: #0f172a;
}

.preview-empty {
  display: grid;
  min-height: 520px;
  place-items: center;
  color: #657892;
  font-weight: 800;
}

@media (max-width: 860px) {
  .drive-bar,
  .filter-panel,
  .selection-strip {
    align-items: stretch;
    flex-direction: column;
  }

  .drive-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .view-panel {
    right: 0;
    width: min(332px, calc(100vw - 48px));
  }

  .video-board.is-list .video-card {
    grid-template-columns: 1fr;
  }

  .video-board.is-list .poster {
    grid-column: 1 / -1;
  }
}
</style>

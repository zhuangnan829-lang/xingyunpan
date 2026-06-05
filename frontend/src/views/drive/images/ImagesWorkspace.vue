<template>
  <main class="images-page" @contextmenu="showImagesContextMenu">
    <input ref="fileInputRef" class="hidden-file-input" type="file" multiple @change="handleFileSelect" />
    <input
      ref="folderInputRef"
      class="hidden-file-input"
      type="file"
      multiple
      webkitdirectory
      directory
      @change="handleFileSelect"
    />

    <section v-if="selectedImages.length > 0" class="selection-strip">
      <span>已选择 {{ selectedImages.length }} 张 · {{ formatFileSize(selectedSize) }}</span>
      <div class="selection-actions">
        <button v-if="selectedImages.length === 1" type="button" @click="openRenameDialog">重命名</button>
        <button type="button" @click="downloadSelectedImages">下载</button>
        <button v-if="selectedImages.length === 1" type="button" @click="openShareDialog">分享</button>
        <button type="button" class="danger" @click="deleteSelectedImages">删除</button>
        <button type="button" @click="clearSelection">取消选择</button>
      </div>
    </section>

    <div class="filter-panel">
      <label class="search-field">
        <Search class="search-icon" />
        <input v-model="keyword" type="search" placeholder="筛选图片名称" />
      </label>
      <span>{{ filteredImages.length }} / {{ images.length }}</span>
    </div>

    <ImageToolbar
      v-model:card-size="cardSize"
      v-model:sort-mode="sortMode"
      :loading="loading"
      @open-menu="showImagesContextMenu"
      @refresh="refreshImages"
    />

    <ImageGrid
      :card-size="cardSize"
      :ensure-thumbnail="ensureThumbnail"
      :images="filteredImages"
      :loading="loading"
      :selected-ids="selectedIds"
      :thumbnail-url="thumbnailUrl"
      @clear-selection="clearSelection"
      @image-context="showImageFileContextMenu"
      @preview="previewImage"
      @toggle-select="toggleSelect"
    />

    <ImageFileContextMenu
      :visible="imageFileMenuVisible"
      :x="imageFileMenuPosition.x"
      :y="imageFileMenuPosition.y"
      @choose-app="showComingSoon('选择应用')"
      @copy="copySelectedImageName"
      @copy-link="copySelectedImageLink"
      @custom-icon="showComingSoon('自定义图标')"
      @delete="deleteContextImage"
      @details="showImageDetails"
      @download="downloadContextImage"
      @locate="showComingSoon('转到所在目录')"
      @manage-tags="showComingSoon('管理标签')"
      @move="showComingSoon('移动')"
      @open="openContextImage"
      @open-photopea="openContextImageInPhotopea"
      @rename="renameContextImage"
      @share="shareContextImage"
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
    <RenameDialog
      v-model:visible="renameDialogVisible"
      :item="selectedImages[0] || null"
      @confirm="confirmRename"
    />
    <ShareDialog
      v-model:visible="shareDialogVisible"
      :file-ids="selectedImages.map((image) => image.id.toString())"
    />

    <el-dialog
      v-model="previewVisible"
      width="min(1080px, 92vw)"
      top="5vh"
      class="image-preview-dialog"
      destroy-on-close
      @closed="clearPreview"
    >
      <template #header>
        <div class="preview-heading">
          <div>
            <p>文件预览</p>
            <strong>{{ previewState.title }}</strong>
          </div>
          <span>{{ previewState.size > 0 ? formatFileSize(previewState.size) : '' }}</span>
        </div>
      </template>

      <div class="preview-shell">
        <div v-if="previewState.loading" class="preview-empty">正在加载预览...</div>
        <div v-else-if="previewState.error" class="preview-empty">{{ previewState.error }}</div>
        <img v-else-if="previewState.url" class="media-preview image-preview" :src="previewState.url" alt="" />
        <div v-else class="preview-empty">此图片暂不可预览，请重新上传原图。</div>
      </div>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { Search } from '@element-plus/icons-vue';
import { ElDialog, ElMessage, ElMessageBox } from 'element-plus';
import {
  createFile as apiCreateFile,
  deleteFile as apiDeleteFile,
  downloadFile as apiDownloadFile,
  getAuthenticatedFileDownloadUrl,
  renameFile as apiRenameFile,
} from '@/api/file';
import { createFolder as apiCreateFolder } from '@/api/folder';
import CreateFolderDialog from '@/components/CreateFolderDialog/index.vue';
import RenameDialog from '@/components/RenameDialog/index.vue';
import ShareDialog from '@/components/ShareDialog/index.vue';
import { useUploadStore } from '@/stores/upload';
import type { FileItem } from '@/types/file';
import { formatFileSize } from '@/utils/format';
import ImageContextMenu from './components/ImageContextMenu.vue';
import ImageFileContextMenu from './components/ImageFileContextMenu.vue';
import ImageGrid from './components/ImageGrid.vue';
import ImageToolbar from './components/ImageToolbar.vue';
import { useImagesWorkspace, type WorkspaceFileKind } from './useImagesWorkspace';

const router = useRouter();
const uploadStore = useUploadStore();

const fileInputRef = ref<HTMLInputElement | null>(null);
const folderInputRef = ref<HTMLInputElement | null>(null);
const createFolderVisible = ref(false);
const renameDialogVisible = ref(false);
const shareDialogVisible = ref(false);
const contextMenuVisible = ref(false);
const contextMenuPosition = reactive({ x: 0, y: 0 });
const imageFileMenuVisible = ref(false);
const imageFileMenuPosition = reactive({ x: 0, y: 0 });
const imageFileMenuTarget = ref<FileItem | null>(null);

const {
  cardSize,
  clearSelection,
  ensureThumbnail,
  filteredImages,
  imageUrl,
  images,
  keyword,
  loading,
  refreshImages,
  removeImagesByIds,
  selectedIds,
  selectedImages,
  selectedSize,
  selectOnly,
  sortMode,
  thumbnailUrl,
  toggleSelect,
  uniqueImageName,
} = useImagesWorkspace();

const previewVisible = ref(false);
const previewState = reactive({
  title: '',
  size: 0,
  loading: false,
  url: '',
  error: '',
});
const refreshedUploadTaskIds = new Set<string>();
const uploadRefreshTimers = new Set<number>();

function scheduleImagesRefresh() {
  [250, 1000, 2500].forEach((delay) => {
    const timer = window.setTimeout(async () => {
      uploadRefreshTimers.delete(timer);
      await refreshImages();
    }, delay);
    uploadRefreshTimers.add(timer);
  });
}

async function previewImage(file: FileItem) {
  closeImagesContextMenu();
  clearPreview();
  previewState.title = file.name;
  previewState.size = file.size || 0;
  previewState.loading = true;
  previewVisible.value = true;

  try {
    const blob = await apiDownloadFile(file.id);
    if (blob.size <= 0) {
      previewState.error = '这个图片实际下载内容为空，暂无可预览内容。请重新上传一次原图。';
      return;
    }

    file.size = blob.size;
    previewState.size = blob.size;

    const mime = blob.type || file.content_type || file.mime_type || '';
    if (!mime.startsWith('image/') && !isImageFileName(file.name)) {
      previewState.error = '此文件不是可预览的图片格式。';
      return;
    }

    previewState.url = URL.createObjectURL(blob);
  } catch (error) {
    previewState.error = error instanceof Error ? error.message : '加载预览失败';
  } finally {
    previewState.loading = false;
  }
}

function clearPreview() {
  if (previewState.url) {
    URL.revokeObjectURL(previewState.url);
  }
  previewState.title = '';
  previewState.size = 0;
  previewState.loading = false;
  previewState.url = '';
  previewState.error = '';
}

function isImageFileName(fileName: string) {
  const extension = fileName.split('.').pop()?.toLowerCase() || '';
  return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg', 'avif'].includes(extension);
}

function openRenameDialog() {
  if (selectedImages.value.length !== 1) return;
  renameDialogVisible.value = true;
}

function openShareDialog() {
  if (selectedImages.value.length !== 1) return;
  shareDialogVisible.value = true;
}

async function confirmRename(name: string) {
  const image = selectedImages.value[0];
  if (!image) return;

  try {
    await apiRenameFile(image.id, name);
    renameDialogVisible.value = false;
    await refreshImages();
    ElMessage.success('重命名成功');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '重命名失败');
  }
}

async function downloadSelectedImages() {
  try {
    for (const image of selectedImages.value) {
      const blob = await apiDownloadFile(image.id);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = image.name;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载失败');
  }
}

async function downloadOneImage(image: FileItem) {
  const blob = await apiDownloadFile(image.id);
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = image.name;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}

async function deleteSelectedImages() {
  const targets = [...selectedImages.value];
  if (!targets.length) return;

  try {
    await ElMessageBox.confirm(
      targets.length === 1 ? `确定删除“${targets[0].name}”吗？` : `确定删除选中的 ${targets.length} 张图片吗？`,
      '删除图片',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
      },
    );

    for (const image of targets) {
      await apiDeleteFile(image.id);
    }
    removeImagesByIds(targets.map((image) => image.id));
    clearSelection();
    void refreshImages();
    ElMessage.success('删除成功');
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

function closeImagesContextMenu() {
  contextMenuVisible.value = false;
  imageFileMenuVisible.value = false;
}

function closeImageFileContextMenu() {
  imageFileMenuVisible.value = false;
}

function showImageFileContextMenu(event: MouseEvent, image: FileItem) {
  event.preventDefault();
  event.stopPropagation();
  closeImagesContextMenu();
  selectOnly(image);
  imageFileMenuTarget.value = image;
  imageFileMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 472));
  imageFileMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 548));
  imageFileMenuVisible.value = true;
}

function getContextImage() {
  return imageFileMenuTarget.value;
}

function openContextImage() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (image) previewImage(image);
}

function openContextImageInPhotopea() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (!image) return;
  window.open(`https://www.photopea.com#${encodeURIComponent(getAuthenticatedFileDownloadUrl(image.id, true))}`, '_blank');
}

async function downloadContextImage() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (!image) return;

  try {
    await downloadOneImage(image);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载失败');
  }
}

function renameContextImage() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (!image) return;
  selectOnly(image);
  openRenameDialog();
}

function shareContextImage() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (!image) return;
  selectOnly(image);
  openShareDialog();
}

async function copySelectedImageName() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (!image) return;

  try {
    await navigator.clipboard.writeText(image.name);
    ElMessage.success('已复制文件名');
  } catch {
    ElMessage.warning('复制失败');
  }
}

async function copySelectedImageLink() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (!image) return;

  try {
    await navigator.clipboard.writeText(getAuthenticatedFileDownloadUrl(image.id, true));
    ElMessage.success('已复制直链');
  } catch {
    ElMessage.warning('复制失败');
  }
}

function showImageDetails() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (!image) return;
  ElMessage.info(`${image.name} · ${formatFileSize(image.size || 0)}`);
}

function deleteContextImage() {
  const image = getContextImage();
  closeImageFileContextMenu();
  if (!image) return;
  selectOnly(image);
  void deleteSelectedImages();
}

function showComingSoon(label: string) {
  closeImageFileContextMenu();
  ElMessage.info(`${label}功能待接入`);
}

function showImagesContextMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  const isToolbarMoreButton = target?.closest('.image-toolbar .icon-action[title="更多"]');

  if (
    !isToolbarMoreButton &&
    (target?.closest('.image-card') ||
      target?.closest('.el-dialog') ||
      target?.closest('.el-overlay') ||
      target?.closest('input') ||
      target?.closest('select') ||
      target?.closest('button'))
  ) {
    return;
  }

  event.preventDefault();
  event.stopPropagation();
  clearSelection();
  contextMenuPosition.x = event.clientX;
  contextMenuPosition.y = event.clientY;
  contextMenuVisible.value = true;
}

function triggerFileUpload() {
  closeImagesContextMenu();
  fileInputRef.value?.click();
}

function triggerFolderUpload() {
  closeImagesContextMenu();
  folderInputRef.value?.click();
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
    scheduleImagesRefresh();
    ElMessage.success(`已添加 ${files.length} 个文件到上传队列`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '添加上传任务失败');
  }
}

async function uploadFromClipboard() {
  closeImagesContextMenu();
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
        files.push(new File([blob], uniqueImageName(`剪贴板文件.${clipboardExtension(type)}`), { type }));
      }
    }

    if (!files.length) {
      ElMessage.warning('剪贴板中没有可上传的文件内容');
      return;
    }

    for (const file of files) {
      await uploadStore.addTask(file);
    }
    ElMessage.success(`已添加 ${files.length} 个剪贴板文件到上传队列`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '读取剪贴板失败');
  }
}

function clipboardExtension(type: string) {
  if (type.includes('png')) return 'png';
  if (type.includes('jpeg') || type.includes('jpg')) return 'jpg';
  if (type.includes('gif')) return 'gif';
  if (type.includes('webp')) return 'webp';
  if (type.includes('svg')) return 'svg';
  if (type.includes('html')) return 'html';
  if (type.includes('plain')) return 'txt';
  return 'bin';
}

function isImageUploadTask(file: File) {
  const mime = file.type || '';
  const extension = file.name.split('.').pop()?.toLowerCase() || '';
  return mime.startsWith('image/') || ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg', 'avif'].includes(extension);
}

async function openOfflineDownload() {
  closeImagesContextMenu();
  await router.push('/drive/offline-downloads');
}

function openCreateFolderDialog() {
  closeImagesContextMenu();
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
  closeImagesContextMenu();
  const defaultNames: Record<WorkspaceFileKind, string> = {
    file: '新建文件.txt',
    markdown: '新建 Markdown.md',
    text: '新建文本.txt',
    drawio: '新建图表.drawio',
    dwb: '新建白板.dwb',
    excalidraw: '新建 Excalidraw.excalidraw',
  };

  try {
    await apiCreateFile(kind, uniqueImageName(defaultNames[kind]), null);
    await refreshImages();
    ElMessage.success('文件已创建');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建文件失败');
  }
}

async function refreshFromContextMenu() {
  closeImagesContextMenu();
  await refreshImages();
}

onMounted(() => {
  document.addEventListener('click', closeImagesContextMenu);
  window.addEventListener('resize', closeImagesContextMenu);
});

onUnmounted(() => {
  document.removeEventListener('click', closeImagesContextMenu);
  window.removeEventListener('resize', closeImagesContextMenu);
  uploadRefreshTimers.forEach((timer) => window.clearTimeout(timer));
  uploadRefreshTimers.clear();
});

watch(
  () => uploadStore.tasks.map((task) => `${task.id}:${task.status}`).join('|'),
  async () => {
    const completedImageTasks = uploadStore.tasks.filter(
      (task) => task.status === 'completed' && isImageUploadTask(task.file) && !refreshedUploadTaskIds.has(task.id),
    );
    if (!completedImageTasks.length) return;

    completedImageTasks.forEach((task) => refreshedUploadTaskIds.add(task.id));
    scheduleImagesRefresh();
  },
);
</script>

<style scoped>
.images-page {
  display: grid;
  gap: 12px;
  min-height: calc(100vh - 96px);
  padding: 0 4px 24px;
  color: #10213f;
}

.hidden-file-input {
  position: fixed;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.filter-panel,
.selection-strip {
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.58);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 14px 34px rgba(88, 128, 176, 0.1);
  backdrop-filter: blur(18px);
}

.filter-panel span {
  color: #61708a;
  font-size: 12px;
  font-weight: 780;
}

.selection-strip,
.filter-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  min-height: 58px;
  padding: 10px 16px;
}

.selection-strip > span {
  color: #13213a;
  font-size: 14px;
  font-weight: 820;
}

.selection-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.selection-actions button {
  min-height: 38px;
  border: 0;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.78);
  color: #172846;
  font-weight: 820;
  padding: 0 14px;
  cursor: pointer;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 8px 18px rgba(64, 119, 180, 0.1);
}

.selection-actions button:hover {
  background: linear-gradient(135deg, #2d70ff, #19aeea);
  color: #fff;
}

.selection-actions button.danger:hover {
  background: linear-gradient(135deg, #ef4444, #f97316);
}

.search-field {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: min(420px, 100%);
}

.search-field input {
  width: 100%;
  border: 0;
  background: transparent;
  color: #172846;
  font-size: 15px;
  font-weight: 760;
  outline: none;
}

.search-icon {
  width: 18px;
  height: 18px;
  color: #2f7df5;
}

:deep(.image-preview-dialog .el-dialog) {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 28px;
  background:
    radial-gradient(circle at 100% 0%, rgba(116, 220, 255, 0.15), transparent 34%),
    rgba(255, 255, 255, 0.94);
  box-shadow: 0 30px 80px rgba(73, 112, 160, 0.24);
  backdrop-filter: blur(20px);
}

:deep(.image-preview-dialog .el-dialog__header) {
  margin: 0;
  padding: 22px 24px 12px;
}

:deep(.image-preview-dialog .el-dialog__body) {
  min-height: 0;
  padding: 0 24px 24px;
}

.preview-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
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

.preview-shell {
  display: grid;
  overflow: hidden;
  min-height: min(520px, calc(100vh - 210px));
  max-height: calc(100vh - 210px);
  border: 1px solid rgba(219, 234, 249, 0.82);
  border-radius: 22px;
  background: rgba(15, 23, 42, 0.92);
}

.preview-empty {
  display: grid;
  min-height: 520px;
  place-items: center;
  padding: 32px;
  background: rgba(255, 255, 255, 0.78);
  color: #657892;
  font-weight: 800;
  text-align: center;
}

.media-preview {
  display: block;
  width: 100%;
  min-height: 520px;
  border: 0;
  background: #0f172a;
}

.image-preview {
  min-height: 0;
  max-height: 72vh;
  object-fit: contain;
}

@media (max-width: 640px) {
  .filter-panel,
  .selection-strip {
    display: grid;
    grid-template-columns: 1fr;
  }

  .selection-actions {
    justify-content: flex-start;
  }
}
</style>

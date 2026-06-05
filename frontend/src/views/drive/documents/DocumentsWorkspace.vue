<template>
  <main class="documents-page" @click="closeFloatingLayers" @contextmenu="showDocumentsContextMenu">
    <input ref="fileInputRef" class="hidden-file-input" type="file" multiple :accept="documentAccept" @change="handleFileSelect" />
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
        <input v-model="keyword" type="search" placeholder="筛选文档名称" />
      </label>
      <div class="drive-actions">
        <button class="tool-button" type="button" title="刷新" @click="refreshDocuments">
          <el-icon><Refresh /></el-icon>
        </button>
        <button class="tool-button" type="button" title="上传文档" @click="triggerFileUpload">
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
              <button type="button" :class="{ active: viewMode === 'compact' }" @click="viewMode = 'compact'">
                <el-icon><Memo /></el-icon>
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
          <option value="recent">最近更新</option>
          <option value="name">名称</option>
          <option value="size">大小</option>
        </select>
        <span class="result-count">{{ filteredDocuments.length }} / {{ documents.length }}</span>
      </div>
    </section>

    <section class="drive-bar glass-panel">
      <div class="breadcrumb">
        <el-icon><Document /></el-icon>
        <span>文档</span>
      </div>
    </section>

    <section v-if="selectedDocuments.length" class="selection-strip glass-panel">
      <span>已选择 {{ selectedDocuments.length }} 个 · {{ formatFileSize(selectedSize) }}</span>
      <div>
        <button v-if="selectedDocuments.length === 1" type="button" @click="openRenameDialog">重命名</button>
        <button type="button" @click="downloadSelectedDocuments">下载</button>
        <button v-if="selectedDocuments.length === 1" type="button" @click="openShareDialog">分享</button>
        <button type="button" class="danger" @click="deleteSelectedDocuments">删除</button>
        <button type="button" @click="clearSelection">取消</button>
      </div>
    </section>

    <section
      class="documents-board glass-panel"
      :class="`is-${viewMode}`"
      :style="{ '--card-size': `${cardSize}px` }"
    >
      <div v-if="loading" class="empty-state">正在加载文档...</div>
      <div v-else-if="!filteredDocuments.length" class="empty-state">
        <el-icon><Document /></el-icon>
        <strong>这里还没有文档</strong>
      </div>
      <article
        v-for="documentFile in filteredDocuments"
        v-else
        :key="documentFile.id"
        class="document-card"
        :class="{ selected: selectedIds.includes(documentFile.id) }"
        @click.stop="toggleSelect(documentFile)"
        @dblclick.stop="previewDocument(documentFile)"
        @contextmenu.prevent.stop="showDocumentFileContextMenu($event, documentFile)"
      >
        <button class="select-dot" type="button" :aria-label="`选择 ${documentFile.name}`" @click.stop="toggleSelect(documentFile)">
          <el-icon v-if="selectedIds.includes(documentFile.id)"><Check /></el-icon>
        </button>
        <button class="document-cover" type="button" @click.stop="toggleSelect(documentFile)" @dblclick.stop="previewDocument(documentFile)">
          <span class="file-badge" :class="documentKind(documentFile).className">{{ documentKind(documentFile).label }}</span>
          <el-icon><component :is="documentKind(documentFile).icon" /></el-icon>
        </button>
        <div class="document-info">
          <strong :title="documentFile.name">{{ documentFile.name }}</strong>
          <span>{{ formatFileSize(documentFile.size || 0) }} · {{ formatDate(documentFile.updated_at) }}</span>
        </div>
        <div class="card-actions">
          <button type="button" title="预览" @click.stop="previewDocument(documentFile)">
            <el-icon><View /></el-icon>
          </button>
          <button type="button" title="下载" @click.stop="downloadOneDocument(documentFile)">
            <el-icon><Download /></el-icon>
          </button>
          <button type="button" title="更多" @click.stop="showDocumentFileContextMenu($event, documentFile)">
            <el-icon><MoreFilled /></el-icon>
          </button>
        </div>
      </article>
    </section>

    <DocumentFileContextMenu
      :visible="documentFileMenuVisible"
      :x="documentFileMenuPosition.x"
      :y="documentFileMenuPosition.y"
      @open="openContextDocument"
      @rename="renameContextDocument"
      @download="downloadContextDocument"
      @share="shareContextDocument"
      @history="showContextPlaceholder('版本历史')"
      @collaboration="showContextPlaceholder('协作管理')"
      @move="openFolderOperation('move')"
      @copy="openFolderOperation('copy')"
      @delete="deleteContextDocument"
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
      :item="selectedDocuments[0] || null"
      @confirm="confirmRename"
    />
    <ShareDialog
      v-model:visible="shareDialogVisible"
      :file-ids="selectedDocuments.map((item) => item.id.toString())"
    />

    <el-dialog
      v-model="previewVisible"
      width="min(1060px, 92vw)"
      top="5vh"
      class="document-preview-dialog"
      :show-close="false"
      destroy-on-close
      @closed="clearPreview"
    >
      <template #header>
        <div class="preview-heading">
          <div>
            <p>文档预览</p>
            <strong>{{ previewState.title }}</strong>
          </div>
          <span>{{ previewState.size > 0 ? formatFileSize(previewState.size) : '' }}</span>
          <button class="preview-close" type="button" title="关闭预览" aria-label="关闭预览" @click.stop="closePreview">
            <el-icon><Close /></el-icon>
          </button>
        </div>
      </template>
      <div class="preview-shell">
        <div v-if="previewState.loading" class="preview-empty">正在加载预览...</div>
        <pre v-else-if="previewState.kind === 'text'" class="text-preview">{{ previewState.text }}</pre>
        <iframe v-else-if="previewState.url" class="frame-preview" :src="previewState.url" title="文档预览"></iframe>
        <div v-else class="preview-empty">{{ previewState.error || '这个文档暂时无法预览' }}</div>
      </div>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { markRaw, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import {
  Check,
  Close,
  CopyDocument,
  Document,
  Download,
  Files,
  Grid,
  List,
  Memo,
  MoreFilled,
  Refresh,
  Search,
  Tickets,
  Upload,
  View,
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
import DocumentFileContextMenu from './components/DocumentFileContextMenu.vue';
import { useDocumentsWorkspace } from './useDocumentsWorkspace';

const documentAccept = '.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.md,.markdown,.csv,.json,.xml,.yaml,.yml,.rtf,.odt,.ods,.odp,.epub,.html,.htm';
const textExtensions = ['txt', 'md', 'markdown', 'csv', 'json', 'xml', 'yaml', 'yml', 'html', 'htm', 'css', 'js', 'ts', 'log', 'ini', 'sql'];
const documentExtensions = ['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt', 'md', 'markdown', 'csv', 'json', 'xml', 'yaml', 'yml', 'rtf', 'odt', 'ods', 'odp', 'epub', 'html', 'htm'];

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
const documentFileMenuVisible = ref(false);
const documentFileMenuPosition = reactive({ x: 0, y: 0 });
const documentFileMenuTarget = ref<FileItem | null>(null);
const folderSelectVisible = ref(false);
const folderSelectMode = ref<'move' | 'copy'>('move');
const folderSelectTargets = ref<FileItem[]>([]);
const refreshedUploadTaskIds = new Set<string>();

const {
  cardSize,
  clearSelection,
  documents,
  filteredDocuments,
  keyword,
  loading,
  refreshDocuments,
  removeDocumentsByIds,
  selectedDocuments,
  selectedIds,
  selectedSize,
  selectOnly,
  sortMode,
  toggleSelect,
  upsertDocuments,
  viewMode,
} = useDocumentsWorkspace();

const previewState = reactive({
  title: '',
  size: 0,
  url: '',
  text: '',
  kind: 'empty' as 'empty' | 'text' | 'frame',
  loading: false,
  error: '',
});

function closeViewPanel() {
  viewPanelVisible.value = false;
}

function closeContextMenus() {
  contextMenuVisible.value = false;
  documentFileMenuVisible.value = false;
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

function fileExtension(file: FileItem | File) {
  return file.name.split('.').pop()?.toLowerCase() || '';
}

function formatDate(value: string) {
  if (!value) return '未知时间';
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit' }).format(new Date(value));
}

function documentKind(file: FileItem) {
  const ext = fileExtension(file);
  if (ext === 'pdf') return { label: 'PDF', className: 'pdf', icon: markRaw(Tickets) };
  if (['doc', 'docx', 'odt'].includes(ext)) return { label: 'W', className: 'word', icon: markRaw(Document) };
  if (['xls', 'xlsx', 'ods', 'csv'].includes(ext)) return { label: 'X', className: 'sheet', icon: markRaw(Files) };
  if (['ppt', 'pptx', 'odp'].includes(ext)) return { label: 'P', className: 'slide', icon: markRaw(CopyDocument) };
  if (['md', 'markdown'].includes(ext)) return { label: 'MD', className: 'markdown', icon: markRaw(Memo) };
  if (textExtensions.includes(ext)) return { label: 'TXT', className: 'text', icon: markRaw(Memo) };
  return { label: ext ? ext.toUpperCase().slice(0, 4) : 'DOC', className: 'doc', icon: markRaw(Document) };
}

function clearPreview() {
  if (previewState.url.startsWith('blob:')) {
    URL.revokeObjectURL(previewState.url);
  }
  previewState.title = '';
  previewState.size = 0;
  previewState.url = '';
  previewState.text = '';
  previewState.kind = 'empty';
  previewState.loading = false;
  previewState.error = '';
}

function closePreview() {
  previewVisible.value = false;
}

async function previewDocument(file: FileItem) {
  selectOnly(file);
  closeContextMenus();
  clearPreview();
  previewState.title = file.name;
  previewState.size = file.size || 0;
  previewState.loading = true;
  previewVisible.value = true;

  try {
    const ext = fileExtension(file);
    if (textExtensions.includes(ext)) {
      const blob = await apiDownloadFile(file.id);
      previewState.text = await blob.text();
      previewState.kind = 'text';
    } else {
      previewState.url = getAuthenticatedFileDownloadUrl(file.id, true);
      previewState.kind = 'frame';
    }
  } catch (error) {
    previewState.error = error instanceof Error ? error.message : '加载预览失败';
  } finally {
    previewState.loading = false;
  }
}

function showDocumentFileContextMenu(event: MouseEvent, file: FileItem) {
  event.preventDefault();
  event.stopPropagation();
  closeViewPanel();
  selectOnly(file);
  documentFileMenuTarget.value = file;
  contextMenuVisible.value = false;
  documentFileMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 230));
  documentFileMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 470));
  documentFileMenuVisible.value = true;
}

function getContextDocument() {
  return documentFileMenuTarget.value;
}

function openContextDocument() {
  const documentFile = getContextDocument();
  closeContextMenus();
  if (documentFile) void previewDocument(documentFile);
}

function renameContextDocument() {
  const documentFile = getContextDocument();
  closeContextMenus();
  if (!documentFile) return;
  selectOnly(documentFile);
  renameDialogVisible.value = true;
}

async function downloadContextDocument() {
  const documentFile = getContextDocument();
  closeContextMenus();
  if (documentFile) await downloadOneDocument(documentFile);
}

function shareContextDocument() {
  const documentFile = getContextDocument();
  closeContextMenus();
  if (!documentFile) return;
  selectOnly(documentFile);
  shareDialogVisible.value = true;
}

function showContextPlaceholder(label: string) {
  const documentFile = getContextDocument();
  closeContextMenus();
  if (!documentFile) return;
  selectOnly(documentFile);
  ElMessage.info(`${label}功能即将开放`);
}

function openFolderOperation(mode: 'move' | 'copy') {
  const documentFile = getContextDocument();
  closeContextMenus();
  if (!documentFile) return;
  selectOnly(documentFile);
  folderSelectMode.value = mode;
  folderSelectTargets.value = [documentFile];
  folderSelectVisible.value = true;
}

async function confirmFolderOperation(folderId: number | null) {
  const targets = folderSelectTargets.value;
  if (!targets.length) return;
  try {
    for (const documentFile of targets) {
      if (folderSelectMode.value === 'move') {
        await apiMoveFile(documentFile.id, folderId);
      } else {
        await apiCopyFile(documentFile.id, folderId);
      }
    }
    folderSelectVisible.value = false;
    folderSelectTargets.value = [];
    await refreshDocuments();
    ElMessage.success(folderSelectMode.value === 'move' ? '移动成功' : '复制成功');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '操作失败');
  }
}

function openRenameDialog() {
  if (selectedDocuments.value.length) renameDialogVisible.value = true;
}

function openShareDialog() {
  if (selectedDocuments.value.length) shareDialogVisible.value = true;
}

async function confirmRename(name: string) {
  const documentFile = selectedDocuments.value[0];
  if (!documentFile) return;
  try {
    await apiRenameFile(documentFile.id, name);
    renameDialogVisible.value = false;
    await refreshDocuments();
    ElMessage.success('重命名成功');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '重命名失败');
  }
}

async function downloadOneDocument(documentFile: FileItem) {
  try {
    const blob = await apiDownloadFile(documentFile.id);
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = documentFile.name;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(link.href);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载失败');
  }
}

async function downloadSelectedDocuments() {
  for (const documentFile of selectedDocuments.value) {
    await downloadOneDocument(documentFile);
  }
}

async function deleteContextDocument() {
  const documentFile = getContextDocument();
  closeContextMenus();
  if (!documentFile) return;
  selectOnly(documentFile);
  await deleteSelectedDocuments();
}

async function deleteSelectedDocuments() {
  const targets = [...selectedDocuments.value];
  if (!targets.length) return;

  try {
    await ElMessageBox.confirm(`确定删除 ${targets.length} 个文档吗？此操作不可撤销。`, '删除文档', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger',
    });

    for (const documentFile of targets) {
      await apiDeleteFile(documentFile.id);
    }
    removeDocumentsByIds(targets.map((item) => item.id));
    clearSelection();
    void refreshDocuments();
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
    scheduleDocumentRefresh();
    ElMessage.success(`已添加 ${files.length} 个文档到上传队列`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '添加上传任务失败');
  }
}

async function uploadFromClipboard() {
  closeContextMenus();
  try {
    const text = await navigator.clipboard?.readText?.();
    if (!text) {
      ElMessage.warning('剪贴板中没有可上传的文本内容');
      return;
    }
    const file = new File([text], uniqueDocumentName('剪贴板文本.txt'), { type: 'text/plain' });
    await uploadStore.addTask(file);
    scheduleDocumentRefresh();
    ElMessage.success('已添加剪贴板文本到上传队列');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '读取剪贴板失败');
  }
}

function uniqueDocumentName(baseName: string) {
  const existingNames = new Set(documents.value.map((item) => item.name));
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
    await apiCreateFile(kind, uniqueDocumentName(defaultNames[kind]), null);
    await refreshDocuments();
    ElMessage.success('文件已创建');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建文件失败');
  }
}

async function refreshFromContextMenu() {
  closeContextMenus();
  await refreshDocuments();
}

function showDocumentsContextMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  if (
    target?.closest('.document-card') ||
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
  documentFileMenuVisible.value = false;
  contextMenuPosition.x = Math.max(12, Math.min(event.clientX, window.innerWidth - 330));
  contextMenuPosition.y = Math.max(12, Math.min(event.clientY, window.innerHeight - 560));
  contextMenuVisible.value = true;
}

function isDocumentUploadTask(file: File) {
  const extension = fileExtension(file);
  return file.type.startsWith('text/') || file.type.includes('pdf') || file.type.includes('document') || documentExtensions.includes(extension);
}

function scheduleDocumentRefresh() {
  [250, 900, 2200].forEach((delay) => {
    window.setTimeout(() => {
      void refreshDocuments();
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
    const completedDocumentTasks = uploadStore.tasks.filter(
      (task) => task.status === 'completed' && isDocumentUploadTask(task.file) && !refreshedUploadTaskIds.has(task.id),
    );
    if (!completedDocumentTasks.length) return;

    completedDocumentTasks.forEach((task) => refreshedUploadTaskIds.add(task.id));
    upsertDocuments(completedDocumentTasks.map((task) => task.result).filter((file): file is FileItem => Boolean(file)));
    scheduleDocumentRefresh();
  },
);
</script>

<style scoped>
.documents-page {
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
.documents-board {
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

.selection-strip .danger {
  color: #ef4444;
}

.documents-board {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(var(--card-size), 100%), 1fr));
  align-content: start;
  gap: 18px;
  min-height: 560px;
  border-radius: 30px;
  padding: 28px;
}

.documents-board.is-list,
.documents-board.is-compact {
  grid-template-columns: 1fr;
}

.document-card {
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

.document-card:hover,
.document-card.selected {
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

.document-cover {
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

.document-cover .el-icon {
  width: 58px;
  height: 58px;
}

.file-badge {
  position: absolute;
  top: 16px;
  left: 16px;
  min-width: 44px;
  border-radius: 12px;
  padding: 6px 9px;
  background: #4f8ded;
  color: #fff;
  font-size: 12px;
  font-weight: 900;
  text-align: center;
}

.file-badge.pdf {
  background: #ef4444;
}

.file-badge.word {
  background: #4f8ded;
}

.file-badge.sheet {
  background: #16a34a;
}

.file-badge.slide {
  background: #f97316;
}

.file-badge.markdown,
.file-badge.text {
  background: #3f3f46;
}

.document-info {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.document-info strong,
.document-info span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.document-info strong {
  color: #10213f;
  font-size: 17px;
  font-weight: 900;
}

.document-info span {
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

.documents-board.is-list .document-card,
.documents-board.is-compact .document-card {
  grid-template-columns: 90px minmax(0, 1fr) auto;
  grid-template-rows: auto;
  align-items: center;
  min-height: 126px;
}

.documents-board.is-list .document-cover,
.documents-board.is-compact .document-cover {
  min-height: 90px;
}

.documents-board.is-compact .document-card {
  min-height: 92px;
  padding: 12px 16px;
}

.documents-board.is-compact .document-cover {
  min-height: 66px;
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

:deep(.document-preview-dialog .el-dialog) {
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

:deep(.document-preview-dialog .el-dialog__header) {
  margin: 0;
  padding: 22px 24px 12px;
}

:deep(.document-preview-dialog .el-dialog__body) {
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

.text-preview {
  overflow: auto;
  min-height: min(620px, calc(100vh - 210px));
  margin: 0;
  padding: 24px;
  color: #14233d;
  font-family: Consolas, Monaco, monospace;
  font-size: 14px;
  line-height: 1.75;
  white-space: pre-wrap;
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

  .documents-board.is-list .document-card,
  .documents-board.is-compact .document-card {
    grid-template-columns: 1fr;
  }
}
</style>

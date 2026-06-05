<template>
  <main class="offline-page" @click="closeFloatingLayers" @contextmenu="showBlankMenu">
    <section class="offline-hero glass-panel">
      <div>
        <p>远程资源接管</p>
        <h1>离线下载</h1>
      </div>
      <div class="hero-actions">
        <label class="auto-refresh">
          <input v-model="autoRefresh" type="checkbox" />
          <span></span>
          自动刷新
        </label>
        <button class="primary-action" type="button" @click.stop="createTask">
          <el-icon><Plus /></el-icon>
          新建任务
        </button>
      </div>
    </section>

    <section class="offline-toolbar glass-panel">
      <label class="search-field">
        <el-icon><Search /></el-icon>
        <input v-model="keyword" type="search" placeholder="筛选任务名称、链接、保存路径或状态" />
      </label>
      <div class="toolbar-actions">
        <button class="tool-button" type="button" title="刷新" @click.stop="refreshTasks">
          <el-icon><Refresh /></el-icon>
        </button>
        <button class="view-button" type="button" :class="{ active: viewMenuVisible }" @click.stop="toggleViewMenu">
          <el-icon><Grid /></el-icon>
          <span>{{ viewLabel }}</span>
        </button>
        <button class="sort-button" type="button" :class="{ active: sortMenuVisible }" @click.stop="toggleSortMenu">
          <el-icon><Sort /></el-icon>
          <span>{{ sortLabel }}</span>
        </button>
        <span class="result-count">{{ filteredTasks.length }} / {{ tasks.length }}</span>
      </div>
    </section>

    <section v-if="selectedTasks.length" class="selection-strip glass-panel">
      <span>已选择 {{ selectedTasks.length }} 个离线任务</span>
      <div>
        <button v-if="selectedTasks.length === 1" type="button" @click.stop="openTask(selectedTasks[0])">打开</button>
        <button type="button" @click.stop="copySelectedLinks">复制链接</button>
        <button type="button" @click.stop="retrySelected">重试</button>
        <button type="button" class="danger" @click.stop="deleteSelected">删除</button>
        <button type="button" @click.stop="clearSelection">取消</button>
      </div>
    </section>

    <section class="task-board glass-panel" :class="`is-${viewMode}`">
      <header class="board-head">
        <div class="section-title">
          <el-icon><Download /></el-icon>
          <span>进行中</span>
          <small>{{ activeTasks.length }}</small>
        </div>
        <div class="board-actions">
          <button type="button" @click.stop="refreshTasks">
            <el-icon><Refresh /></el-icon>
            刷新
          </button>
          <button type="button" @click.stop="toggleSortMenu">
            <el-icon><Sort /></el-icon>
            排序
          </button>
        </div>
      </header>

      <div v-if="!activeTasks.length" class="empty-state compact">
        <div class="empty-visual">
          <el-icon><Box /></el-icon>
        </div>
        <strong>没有进行中的任务</strong>
        <span>添加下载链接后，任务会在这里显示进度、速度和保存路径。</span>
      </div>

      <article
        v-for="task in activeTasks"
        :key="task.id"
        class="task-card"
        :class="{ selected: selectedIds.includes(task.id), failed: task.status === 'failed' }"
        @click.stop="toggleSelect(task)"
        @dblclick.stop="openTask(task)"
        @contextmenu.prevent.stop="showItemMenu($event, task)"
      >
        <button class="select-dot" type="button" :aria-label="`选择 ${task.name}`" @click.stop="toggleSelect(task)">
          <el-icon v-if="selectedIds.includes(task.id)"><Check /></el-icon>
        </button>
        <div class="task-icon">
          <span class="status-chip" :class="task.status">{{ statusLabel(task.status) }}</span>
          <el-icon><component :is="taskIcon(task)" /></el-icon>
        </div>
        <div class="task-info">
          <strong :title="task.name">{{ task.name }}</strong>
          <span>{{ task.url }}</span>
          <div class="progress-track">
            <i :style="{ width: `${task.progress}%` }"></i>
          </div>
          <dl>
            <div>
              <dt>进度</dt>
              <dd>{{ task.progress }}%</dd>
            </div>
            <div>
              <dt>速度</dt>
              <dd>{{ task.speed }}</dd>
            </div>
            <div>
              <dt>大小</dt>
              <dd>{{ task.size }}</dd>
            </div>
            <div>
              <dt>保存到</dt>
              <dd>{{ task.savePath }}</dd>
            </div>
          </dl>
        </div>
        <div class="card-actions">
          <button type="button" title="打开" @click.stop="openTask(task)">
            <el-icon><View /></el-icon>
          </button>
          <button type="button" title="暂停/继续" @click.stop="togglePause(task)">
            <el-icon><component :is="task.status === 'paused' ? VideoPlay : VideoPause" /></el-icon>
          </button>
          <button type="button" title="更多" @click.stop="showItemMenu($event, task)">
            <el-icon><MoreFilled /></el-icon>
          </button>
        </div>
      </article>
    </section>

    <section class="task-board glass-panel" :class="`is-${viewMode}`">
      <header class="board-head">
        <div class="section-title">
          <el-icon><CircleCheck /></el-icon>
          <span>已完成</span>
          <small>{{ completedTasks.length }}</small>
        </div>
      </header>

      <div v-if="!completedTasks.length" class="empty-state compact">
        <div class="empty-visual">
          <el-icon><Box /></el-icon>
        </div>
        <strong>没有已完成任务</strong>
        <span>完成的离线下载会保留在这里，方便打开保存位置或重新复制来源链接。</span>
      </div>

      <article
        v-for="task in completedTasks"
        :key="task.id"
        class="task-card"
        :class="{ selected: selectedIds.includes(task.id) }"
        @click.stop="toggleSelect(task)"
        @dblclick.stop="openTask(task)"
        @contextmenu.prevent.stop="showItemMenu($event, task)"
      >
        <button class="select-dot" type="button" :aria-label="`选择 ${task.name}`" @click.stop="toggleSelect(task)">
          <el-icon v-if="selectedIds.includes(task.id)"><Check /></el-icon>
        </button>
        <div class="task-icon">
          <span class="status-chip completed">已完成</span>
          <el-icon><CircleCheck /></el-icon>
        </div>
        <div class="task-info">
          <strong :title="task.name">{{ task.name }}</strong>
          <span>{{ task.url }}</span>
          <dl>
            <div>
              <dt>完成时间</dt>
              <dd>{{ formatDate(task.updatedAt) }}</dd>
            </div>
            <div>
              <dt>大小</dt>
              <dd>{{ task.size }}</dd>
            </div>
            <div>
              <dt>保存到</dt>
              <dd>{{ task.savePath }}</dd>
            </div>
          </dl>
        </div>
        <div class="card-actions">
          <button type="button" title="打开" @click.stop="openTask(task)">
            <el-icon><View /></el-icon>
          </button>
          <button type="button" title="复制链接" @click.stop="copyTaskLink(task)">
            <el-icon><CopyDocument /></el-icon>
          </button>
          <button type="button" title="更多" @click.stop="showItemMenu($event, task)">
            <el-icon><MoreFilled /></el-icon>
          </button>
        </div>
      </article>
    </section>

    <Teleport to="body">
      <div v-if="createDialogVisible" class="dialog-mask" @click.self="closeCreateDialog">
        <form class="create-dialog" @submit.prevent="submitTask">
          <header>
            <strong>新建离线下载</strong>
            <button type="button" @click="closeCreateDialog">
              <el-icon><Close /></el-icon>
            </button>
          </header>
          <label>
            <span>下载链接</span>
            <input v-model="draft.url" required type="url" placeholder="https://example.com/file.zip 或 magnet:..." />
          </label>
          <label>
            <span>任务名称</span>
            <input v-model="draft.name" type="text" placeholder="自动识别或手动填写" />
          </label>
          <label>
            <span>保存位置</span>
            <input
              :value="draft.savePath"
              type="text"
              placeholder="/离线下载"
              @input="draft.savePath = ($event.target as HTMLInputElement).value"
            />
          </label>
          <footer>
            <button type="button" @click="closeCreateDialog">取消</button>
            <button class="primary-action" type="submit">创建任务</button>
          </footer>
        </form>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="viewMenuVisible" class="floating-menu" :style="menuStyle(viewMenuPosition)" @click.stop>
        <p>显示方式</p>
        <button type="button" :class="{ active: viewMode === 'grid' }" @click="setViewMode('grid')">
          <el-icon><Grid /></el-icon>
          卡片视图
        </button>
        <button type="button" :class="{ active: viewMode === 'list' }" @click="setViewMode('list')">
          <el-icon><List /></el-icon>
          列表视图
        </button>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="sortMenuVisible" class="floating-menu" :style="menuStyle(sortMenuPosition)" @click.stop>
        <p>排序方式</p>
        <button
          v-for="option in sortOptions"
          :key="option.value"
          type="button"
          :class="{ active: sortMode === option.value }"
          @click="setSortMode(option.value)"
        >
          <el-icon><component :is="option.icon" /></el-icon>
          {{ option.label }}
        </button>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="itemMenuVisible" class="floating-menu context-menu" :style="menuStyle(itemMenuPosition)" @click.stop @contextmenu.prevent.stop>
        <button type="button" @click="openContextTask">
          <el-icon><View /></el-icon>
          打开任务
        </button>
        <button type="button" @click="toggleContextPause">
          <el-icon><VideoPause /></el-icon>
          暂停/继续
        </button>
        <button type="button" @click="retryContextTask">
          <el-icon><Refresh /></el-icon>
          重试任务
        </button>
        <button type="button" @click="copyContextLink">
          <el-icon><CopyDocument /></el-icon>
          复制来源链接
        </button>
        <button type="button" @click="showContextDetails">
          <el-icon><InfoFilled /></el-icon>
          任务详情
        </button>
        <hr />
        <button type="button" class="danger" @click="deleteContextTask">
          <el-icon><Delete /></el-icon>
          删除任务
        </button>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="blankMenuVisible" class="floating-menu context-menu" :style="menuStyle(blankMenuPosition)" @click.stop @contextmenu.prevent.stop>
        <button type="button" @click="createFromMenu">
          <el-icon><Plus /></el-icon>
          新建任务
        </button>
        <button type="button" @click="refreshFromMenu">
          <el-icon><Refresh /></el-icon>
          刷新列表
        </button>
        <button type="button" @click="selectAll">
          <el-icon><Grid /></el-icon>
          全选
        </button>
        <button type="button" @click="invertSelection">
          <el-icon><Operation /></el-icon>
          反选
        </button>
        <button type="button" @click="clearSelection">
          <el-icon><Close /></el-icon>
          取消选择
        </button>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="previewVisible" class="preview-mask" @click.self="closePreview">
        <section class="preview-dialog glass-panel" @click.stop>
          <header>
            <div>
              <span>离线下载预览</span>
              <strong>{{ previewState.title }}</strong>
            </div>
            <button type="button" @click="closePreview">
              <el-icon><Close /></el-icon>
            </button>
          </header>
          <div class="preview-body">
            <div v-if="previewState.loading" class="preview-empty">正在加载预览...</div>
            <div v-else-if="previewState.error" class="preview-empty">{{ previewState.error }}</div>
            <img v-else-if="previewState.kind === 'image'" class="preview-media image" :src="previewState.url" alt="" />
            <video v-else-if="previewState.kind === 'video'" class="preview-media" :src="previewState.url" controls autoplay></video>
            <audio v-else-if="previewState.kind === 'audio'" class="preview-audio" :src="previewState.url" controls autoplay></audio>
            <div v-else-if="previewState.kind === 'docx'" class="preview-docx" v-html="previewState.html"></div>
            <pre v-else-if="previewState.kind === 'text'" class="preview-text">{{ previewState.text }}</pre>
            <div v-else-if="previewState.kind === 'pdf'" class="preview-pdf">
              <img v-for="(page, index) in previewState.pdfPages" :key="index" :src="page" :alt="`PDF page ${index + 1}`" />
            </div>
            <div v-else class="preview-empty">此文件类型暂不支持站内预览，可先下载后查看。</div>
          </div>
        </section>
      </div>
    </Teleport>
  </main>
</template>

<script setup lang="ts">
import { computed, markRaw, onMounted, onUnmounted, reactive, ref } from 'vue';
import {
  Box,
  Check,
  CircleCheck,
  Close,
  CopyDocument,
  Delete,
  Document,
  Download,
  Files,
  Grid,
  InfoFilled,
  Link,
  List,
  MoreFilled,
  Operation,
  Plus,
  Refresh,
  Search,
  Sort,
  VideoPause,
  VideoPlay,
  View,
} from '@element-plus/icons-vue';
import type { Component } from 'vue';
import { ElIcon, ElMessage, ElMessageBox } from 'element-plus';
import {
  batchDeleteOfflineDownloads,
  createOfflineDownload,
  deleteOfflineDownload,
  listOfflineDownloads,
  pauseOfflineDownload,
  refreshOfflineDownloads,
  resumeOfflineDownload,
  retryOfflineDownload,
  type OfflineDownloadTask as ApiOfflineDownloadTask,
} from '@/api/offline-download';
import { downloadFile as apiDownloadFile, previewFileAsPdf } from '@/api/file';

type TaskStatus = 'downloading' | 'paused' | 'queued' | 'completed' | 'failed';
type ViewMode = 'grid' | 'list';
type SortMode = 'recent' | 'name' | 'progress' | 'status';

type OfflineTask = {
  id: number;
  name: string;
  url: string;
  savePath: string;
  status: TaskStatus;
  progress: number;
  speed: string;
  size: string;
  createdAt: string;
  updatedAt: string;
  savedFileId?: number | null;
  savedFolderId?: number | null;
};

const autoRefresh = ref(true);
const keyword = ref('');
const viewMode = ref<ViewMode>('grid');
const sortMode = ref<SortMode>('recent');
const selectedIds = ref<number[]>([]);
const tasks = ref<OfflineTask[]>([]);
const itemMenuTarget = ref<OfflineTask | null>(null);
const createDialogVisible = ref(false);
const draft = reactive({ url: '', name: '', savePath: '/离线下载' });

const viewMenuVisible = ref(false);
const sortMenuVisible = ref(false);
const itemMenuVisible = ref(false);
const blankMenuVisible = ref(false);
const viewMenuPosition = reactive({ x: 0, y: 0 });
const sortMenuPosition = reactive({ x: 0, y: 0 });
const itemMenuPosition = reactive({ x: 0, y: 0 });
const blankMenuPosition = reactive({ x: 0, y: 0 });
const previewVisible = ref(false);
const previewState = reactive<{
  title: string;
  loading: boolean;
  kind: 'empty' | 'image' | 'video' | 'audio' | 'docx' | 'text' | 'pdf' | 'unsupported';
  url: string;
  html: string;
  text: string;
  pdfPages: string[];
  error: string;
}>({
  title: '',
  loading: false,
  kind: 'empty',
  url: '',
  html: '',
  text: '',
  pdfPages: [],
  error: '',
});

let refreshTimer: number | undefined;

const sortOptions: Array<{ value: SortMode; label: string; icon: Component }> = [
  { value: 'recent', label: '最近更新', icon: Refresh },
  { value: 'name', label: '任务名称', icon: Document },
  { value: 'progress', label: '下载进度', icon: Download },
  { value: 'status', label: '任务状态', icon: CircleCheck },
];

const viewLabel = computed(() => (viewMode.value === 'grid' ? '卡片' : '列表'));
const sortLabel = computed(() => sortOptions.find((option) => option.value === sortMode.value)?.label || '排序');
const selectedTasks = computed(() => tasks.value.filter((task) => selectedIds.value.includes(task.id)));
const filteredTasks = computed(() => {
  const query = keyword.value.trim().toLowerCase();
  const source = query
    ? tasks.value.filter((task) => `${task.name} ${task.url} ${task.savePath} ${task.status}`.toLowerCase().includes(query))
    : tasks.value;

  return [...source].sort((a, b) => {
    if (sortMode.value === 'name') return a.name.localeCompare(b.name, 'zh-CN');
    if (sortMode.value === 'progress') return b.progress - a.progress;
    if (sortMode.value === 'status') return a.status.localeCompare(b.status);
    return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
  });
});
const activeTasks = computed(() => filteredTasks.value.filter((task) => task.status !== 'completed'));
const completedTasks = computed(() => filteredTasks.value.filter((task) => task.status === 'completed'));

function mapTask(task: ApiOfflineDownloadTask): OfflineTask {
  return {
    id: task.id,
    name: task.name,
    url: task.url,
    savePath: task.save_path,
    status: task.status,
    progress: task.progress,
    speed: task.speed,
    size: task.size,
    createdAt: task.created_at,
    updatedAt: task.updated_at,
    savedFileId: task.saved_file_id ?? null,
    savedFolderId: task.saved_folder_id ?? null,
  };
}

function syncTasks(items: ApiOfflineDownloadTask[]) {
  tasks.value = items.map(mapTask);
  selectedIds.value = selectedIds.value.filter((id) => tasks.value.some((task) => task.id === id));
}

function upsertTask(task: OfflineTask) {
  const found = tasks.value.some((item) => item.id === task.id);
  tasks.value = found ? tasks.value.map((item) => (item.id === task.id ? task : item)) : [task, ...tasks.value];
}

async function loadTasks(showToast = false) {
  try {
    const items = await listOfflineDownloads();
    syncTasks(items);
    if (showToast) ElMessage.success('任务列表已刷新');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载离线下载任务失败');
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(
    new Date(value),
  );
}

function menuStyle(position: { x: number; y: number }) {
  return { left: `${position.x}px`, top: `${position.y}px` };
}

function clampMenu(event: MouseEvent, position: { x: number; y: number }, width: number, height: number) {
  position.x = Math.max(12, Math.min(event.clientX, window.innerWidth - width - 12));
  position.y = Math.max(12, Math.min(event.clientY, window.innerHeight - height - 12));
}

function closeFloatingLayers() {
  viewMenuVisible.value = false;
  sortMenuVisible.value = false;
  itemMenuVisible.value = false;
  blankMenuVisible.value = false;
}

function toggleViewMenu(event: MouseEvent) {
  closeFloatingLayers();
  clampMenu(event, viewMenuPosition, 220, 156);
  viewMenuVisible.value = true;
}

function toggleSortMenu(event: MouseEvent) {
  closeFloatingLayers();
  clampMenu(event, sortMenuPosition, 230, 238);
  sortMenuVisible.value = true;
}

function setViewMode(mode: ViewMode) {
  viewMode.value = mode;
  viewMenuVisible.value = false;
}

function setSortMode(mode: SortMode) {
  sortMode.value = mode;
  sortMenuVisible.value = false;
}

function statusLabel(status: TaskStatus) {
  return {
    downloading: '下载中',
    paused: '已暂停',
    queued: '排队中',
    completed: '已完成',
    failed: '失败',
  }[status];
}

function taskIcon(task: OfflineTask) {
  if (task.url.startsWith('magnet:')) return markRaw(Link);
  if (task.name.includes('.')) return markRaw(Document);
  return markRaw(Files);
}

function fileExtension(fileName: string) {
  const dotIndex = fileName.lastIndexOf('.');
  return dotIndex >= 0 ? fileName.slice(dotIndex + 1).toLowerCase() : '';
}

function isPlainTextTask(task: OfflineTask, mime = '') {
  const ext = fileExtension(task.name);
  return (
    mime.startsWith('text/') ||
    ['md', 'markdown', 'txt', 'json', 'xml', 'yaml', 'yml', 'csv', 'log', 'ini', 'css', 'js', 'ts', 'html', 'htm', 'go', 'py', 'java', 'c', 'cpp', 'h', 'sql', 'sh'].includes(ext)
  );
}

function resetPreviewState() {
  if (previewState.url) {
    URL.revokeObjectURL(previewState.url);
  }
  previewState.title = '';
  previewState.loading = false;
  previewState.kind = 'empty';
  previewState.url = '';
  previewState.html = '';
  previewState.text = '';
  previewState.pdfPages = [];
  previewState.error = '';
}

function closePreview() {
  previewVisible.value = false;
  resetPreviewState();
}

async function renderPdfPages(blob: Blob) {
  const [pdfjsLib, pdfWorker] = await Promise.all([
    import('pdfjs-dist/legacy/build/pdf.mjs'),
    import('pdfjs-dist/legacy/build/pdf.worker.mjs?url'),
  ]);
  pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorker.default;

  const pdf = await pdfjsLib.getDocument({ data: await blob.arrayBuffer() }).promise;
  const pages: string[] = [];
  for (let index = 1; index <= pdf.numPages; index += 1) {
    const page = await pdf.getPage(index);
    const viewport = page.getViewport({ scale: 1.45 });
    const canvas = document.createElement('canvas');
    const context = canvas.getContext('2d');
    if (!context) continue;

    canvas.width = Math.ceil(viewport.width);
    canvas.height = Math.ceil(viewport.height);
    await page.render({ canvas, canvasContext: context, viewport }).promise;
    pages.push(canvas.toDataURL('image/png'));
  }
  return pages;
}

async function previewTask(task: OfflineTask) {
  selectOnly(task);
  closeFloatingLayers();
  if (task.status !== 'completed') {
    window.open(task.url, '_blank', 'noopener,noreferrer');
    return;
  }
  if (!task.savedFileId) {
    ElMessage.warning('还没有定位到已保存文件，请先刷新任务列表后再试');
    return;
  }

  resetPreviewState();
  previewState.title = task.name;
  previewState.loading = true;
  previewVisible.value = true;

  try {
    const ext = fileExtension(task.name);
    if (['ppt', 'pptx', 'pps', 'ppsx'].includes(ext)) {
      const pdfBlob = await previewFileAsPdf(task.savedFileId);
      previewState.kind = 'pdf';
      previewState.pdfPages = await renderPdfPages(pdfBlob);
      return;
    }

    const blob = await apiDownloadFile(task.savedFileId);
    const mime = blob.type || '';
    if (blob.size <= 0) {
      previewState.kind = 'unsupported';
      previewState.error = '这个文件实际内容为空，暂无可预览内容。';
    } else if (ext === 'docx') {
      const mammoth = await import('mammoth');
      const arrayBuffer = await blob.arrayBuffer();
      const result = await mammoth.convertToHtml({ arrayBuffer });
      previewState.kind = 'docx';
      previewState.html = result.value || '<p>文档没有可显示的正文内容。</p>';
    } else if (mime === 'application/pdf' || ext === 'pdf') {
      previewState.kind = 'pdf';
      previewState.pdfPages = await renderPdfPages(blob);
    } else if (mime.startsWith('image/') || ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg', 'avif'].includes(ext)) {
      previewState.kind = 'image';
      previewState.url = URL.createObjectURL(blob);
    } else if (mime.startsWith('video/')) {
      previewState.kind = 'video';
      previewState.url = URL.createObjectURL(blob);
    } else if (mime.startsWith('audio/')) {
      previewState.kind = 'audio';
      previewState.url = URL.createObjectURL(blob);
    } else if (isPlainTextTask(task, mime)) {
      previewState.kind = 'text';
      previewState.text = await blob.text();
    } else {
      previewState.kind = 'unsupported';
      previewState.error = '此文件类型暂不支持站内预览，可先下载后查看。';
    }
  } catch (error) {
    previewState.kind = 'unsupported';
    previewState.error = error instanceof Error ? error.message : '加载预览失败';
  } finally {
    previewState.loading = false;
  }
}

function toggleSelect(task: OfflineTask) {
  selectedIds.value = selectedIds.value.includes(task.id)
    ? selectedIds.value.filter((id) => id !== task.id)
    : [...selectedIds.value, task.id];
}

function selectOnly(task: OfflineTask) {
  selectedIds.value = [task.id];
}

function clearSelection() {
  selectedIds.value = [];
  closeFloatingLayers();
}

function createTask() {
  closeFloatingLayers();
  createDialogVisible.value = true;
}

function closeCreateDialog() {
  createDialogVisible.value = false;
}

async function submitTask() {
  try {
    const created = await createOfflineDownload({
      url: draft.url.trim(),
      name: draft.name.trim(),
      save_path: draft.savePath.trim() || '/离线下载',
    });
    const task = mapTask(created);
    upsertTask(task);
    selectedIds.value = [task.id];
    draft.url = '';
    draft.name = '';
    draft.savePath = '/离线下载';
    closeCreateDialog();
    ElMessage.success('离线下载任务已创建');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建离线下载任务失败');
  }
}

async function refreshTasks() {
  closeFloatingLayers();
  try {
    const items = await refreshOfflineDownloads();
    syncTasks(items);
    ElMessage.success('任务列表已刷新');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '刷新离线下载任务失败');
  }
}

function openTask(task: OfflineTask) {
  void previewTask(task);
}

async function copyText(text: string, message: string) {
  await navigator.clipboard.writeText(text);
  ElMessage.success(message);
}

async function copyTaskLink(task: OfflineTask) {
  await copyText(task.url, '来源链接已复制');
}

async function copySelectedLinks() {
  if (!selectedTasks.value.length) {
    ElMessage.info('请先选择离线任务');
    return;
  }
  await copyText(selectedTasks.value.map((task) => task.url).join('\n'), '已复制选中任务链接');
}

async function togglePause(task: OfflineTask) {
  if (task.status === 'completed') {
    ElMessage.info('已完成任务无需暂停');
    return;
  }
  try {
    const updated = task.status === 'paused' ? await resumeOfflineDownload(task.id) : await pauseOfflineDownload(task.id);
    upsertTask(mapTask(updated));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '更新任务状态失败');
  }
}

async function retryTasks(targets: OfflineTask[]) {
  try {
    const updated = await Promise.all(targets.map((task) => retryOfflineDownload(task.id)));
    updated.map(mapTask).forEach(upsertTask);
    ElMessage.success('任务已重新加入队列');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '重试任务失败');
  }
}

function retrySelected() {
  void retryTasks(selectedTasks.value);
}

async function deleteTasks(targets: OfflineTask[]) {
  if (!targets.length) return;
  await ElMessageBox.confirm(`确定删除 ${targets.length} 个离线下载任务吗？`, '删除离线任务', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  });
  const ids = targets.map((task) => task.id);
  if (ids.length === 1) {
    await deleteOfflineDownload(ids[0]);
  } else {
    await batchDeleteOfflineDownloads(ids);
  }
  const idSet = new Set(ids);
  tasks.value = tasks.value.filter((task) => !idSet.has(task.id));
  clearSelection();
  ElMessage.success('离线任务已删除');
}

async function deleteSelected() {
  try {
    await deleteTasks(selectedTasks.value);
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

function showItemMenu(event: MouseEvent, task: OfflineTask) {
  closeFloatingLayers();
  selectOnly(task);
  itemMenuTarget.value = task;
  clampMenu(event, itemMenuPosition, 236, 330);
  itemMenuVisible.value = true;
}

function showBlankMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null;
  if (
    target?.closest('.task-card') ||
    target?.closest('.floating-menu') ||
    target?.closest('.dialog-mask') ||
    target?.closest('button') ||
    target?.closest('input') ||
    target?.closest('.el-overlay')
  ) {
    return;
  }
  event.preventDefault();
  closeFloatingLayers();
  clampMenu(event, blankMenuPosition, 220, 264);
  blankMenuVisible.value = true;
}

function getContextTask() {
  return itemMenuTarget.value;
}

function openContextTask() {
  const task = getContextTask();
  if (task) openTask(task);
}

function toggleContextPause() {
  const task = getContextTask();
  closeFloatingLayers();
  if (task) void togglePause(task);
}

function retryContextTask() {
  const task = getContextTask();
  closeFloatingLayers();
  if (task) void retryTasks([task]);
}

async function copyContextLink() {
  const task = getContextTask();
  closeFloatingLayers();
  if (task) await copyTaskLink(task);
}

function showContextDetails() {
  const task = getContextTask();
  closeFloatingLayers();
  if (!task) return;
  ElMessageBox.alert(
    `任务：${task.name}\n链接：${task.url}\n状态：${statusLabel(task.status)}\n进度：${task.progress}%\n速度：${task.speed}\n大小：${task.size}\n保存到：${task.savePath}`,
    '离线任务详情',
    { confirmButtonText: '知道了' },
  );
}

async function deleteContextTask() {
  const task = getContextTask();
  closeFloatingLayers();
  if (!task) return;
  try {
    await deleteTasks([task]);
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

function createFromMenu() {
  createTask();
}

function refreshFromMenu() {
  void refreshTasks();
}

function selectAll() {
  selectedIds.value = filteredTasks.value.map((task) => task.id);
  closeFloatingLayers();
}

function invertSelection() {
  const visible = new Set(filteredTasks.value.map((task) => task.id));
  const selected = new Set(selectedIds.value);
  selectedIds.value = [
    ...selectedIds.value.filter((id) => !visible.has(id)),
    ...filteredTasks.value.filter((task) => !selected.has(task.id)).map((task) => task.id),
  ];
  closeFloatingLayers();
}

onMounted(() => {
  void loadTasks();
  refreshTimer = window.setInterval(() => {
    if (autoRefresh.value) {
      void refreshTasks();
    }
  }, 30000);
});

onUnmounted(() => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer);
  }
  resetPreviewState();
});
</script>

<style scoped>
.offline-page {
  position: relative;
  isolation: isolate;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: calc(100vh - 112px);
  padding: 14px 26px 28px;
  overflow: visible;
}

.offline-page::before,
.offline-page::after {
  content: '';
  position: fixed;
  z-index: -1;
  pointer-events: none;
  border-radius: 999px;
  filter: blur(10px);
}

.offline-page::before {
  width: 48vw;
  height: 40vw;
  right: 2vw;
  top: 0;
  background: radial-gradient(circle, rgba(191, 219, 254, 0.44), rgba(252, 231, 243, 0.2) 58%, transparent 72%);
}

.offline-page::after {
  width: 40vw;
  height: 34vw;
  left: 18vw;
  bottom: 0;
  background: radial-gradient(circle, rgba(186, 230, 253, 0.34), rgba(255, 214, 226, 0.2) 56%, transparent 72%);
}

.glass-panel {
  border: 1px solid rgba(255, 255, 255, 0.74);
  border-radius: 28px;
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.5), transparent 42%),
    radial-gradient(circle at 100% 8%, rgba(252, 231, 243, 0.48), transparent 42%),
    rgba(255, 255, 255, 0.6);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.88), 0 22px 64px rgba(115, 145, 190, 0.13);
  backdrop-filter: blur(24px);
}

.offline-hero,
.offline-toolbar,
.selection-strip,
.board-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.offline-hero {
  min-height: 110px;
  padding: 24px 34px;
}

.offline-hero p {
  margin: 0 0 8px;
  color: #64748b;
  font-weight: 820;
}

.offline-hero h1 {
  margin: 0;
  color: #10203d;
  font-size: 38px;
  line-height: 1.1;
  font-weight: 920;
}

.hero-actions,
.toolbar-actions,
.board-actions,
.selection-strip div,
.card-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.auto-refresh {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #172642;
  font-weight: 820;
}

.auto-refresh input {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.auto-refresh span {
  position: relative;
  width: 44px;
  height: 24px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.32);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.auto-refresh span::after {
  content: '';
  position: absolute;
  top: 3px;
  left: 4px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 6px 14px rgba(75, 95, 130, 0.18);
  transition: transform 0.18s ease;
}

.auto-refresh input:checked + span {
  background: linear-gradient(90deg, #60a5fa, #22c7b8);
}

.auto-refresh input:checked + span::after {
  transform: translateX(18px);
}

.primary-action,
.tool-button,
.view-button,
.sort-button,
.board-actions button,
.selection-strip button,
.empty-state button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 48px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  color: #172642;
  font-weight: 820;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82), 0 14px 34px rgba(99, 132, 174, 0.1);
  cursor: pointer;
}

.primary-action {
  padding: 0 22px;
  border-color: rgba(47, 125, 245, 0.58);
  background: linear-gradient(135deg, #2f72ff 0%, #1bb6e8 100%);
  color: #fff;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.28), 0 18px 34px rgba(45, 127, 240, 0.26);
}

.offline-toolbar {
  min-height: 92px;
  padding: 18px 34px;
}

.search-field {
  display: flex;
  align-items: center;
  flex: 1;
  gap: 16px;
  min-width: 240px;
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
  color: rgba(71, 85, 105, 0.7);
}

.tool-button {
  width: 54px;
  padding: 0;
}

.view-button,
.sort-button {
  padding: 0 18px;
}

.view-button.active,
.sort-button.active,
.tool-button:hover,
.board-actions button:hover {
  background: rgba(255, 255, 255, 0.76);
  color: #1d72ed;
}

.result-count {
  color: #64748b;
  font-size: 17px;
  font-weight: 840;
  white-space: nowrap;
}

.selection-strip {
  min-height: 74px;
  padding: 0 28px;
}

.selection-strip span {
  color: #172642;
  font-weight: 840;
}

.task-board {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(286px, 1fr));
  align-content: start;
  gap: 18px;
  min-height: 420px;
  padding: 28px;
}

.board-head,
.empty-state {
  grid-column: 1 / -1;
}

.board-head {
  min-height: 64px;
  padding: 0 6px 12px;
}

.section-title {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  color: #10203d;
  font-size: 24px;
  font-weight: 920;
}

.section-title small {
  display: inline-grid;
  min-width: 30px;
  height: 30px;
  place-items: center;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.72);
  color: #64748b;
  font-size: 14px;
}

.task-card {
  position: relative;
  display: grid;
  grid-template-rows: 132px auto auto;
  gap: 14px;
  min-height: 370px;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.64);
  border-radius: 24px;
  background:
    linear-gradient(160deg, rgba(255, 255, 255, 0.66), rgba(255, 255, 255, 0.42)),
    radial-gradient(circle at 0% 0%, rgba(191, 219, 254, 0.24), transparent 48%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.22), transparent 48%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.88), 0 18px 42px rgba(105, 133, 178, 0.08);
  backdrop-filter: blur(18px);
  cursor: default;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.task-card:hover,
.task-card.selected {
  border-color: rgba(47, 125, 245, 0.52);
  box-shadow: 0 20px 52px rgba(77, 129, 225, 0.16);
  transform: translateY(-1px);
}

.task-card.failed {
  border-color: rgba(248, 113, 113, 0.42);
}

.select-dot {
  position: absolute;
  z-index: 2;
  top: 14px;
  left: 14px;
  display: grid;
  width: 28px;
  height: 28px;
  place-items: center;
  border: 1px solid rgba(47, 125, 245, 0.42);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.78);
  color: #1d72ed;
  cursor: pointer;
}

.task-icon {
  position: relative;
  display: grid;
  min-height: 132px;
  place-items: center;
  border-radius: 18px;
  background:
    radial-gradient(circle at 22% 20%, rgba(186, 230, 253, 0.64), transparent 44%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 46%),
    rgba(255, 255, 255, 0.62);
  color: #2f7df5;
}

.task-icon .el-icon {
  width: 66px;
  height: 66px;
}

.status-chip {
  position: absolute;
  top: 12px;
  right: 12px;
  border-radius: 999px;
  padding: 6px 12px;
  background: rgba(59, 130, 246, 0.12);
  color: #1d4ed8;
  font-size: 12px;
  font-weight: 900;
}

.status-chip.completed {
  background: rgba(34, 197, 94, 0.13);
  color: #15803d;
}

.status-chip.paused,
.status-chip.queued {
  background: rgba(100, 116, 139, 0.12);
  color: #64748b;
}

.status-chip.failed {
  background: rgba(248, 113, 113, 0.14);
  color: #dc2626;
}

.task-info {
  min-width: 0;
}

.task-info strong,
.task-info span,
.task-info dt,
.task-info dd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-info strong,
.task-info span {
  display: block;
}

.task-info strong {
  color: #10203d;
  font-size: 18px;
  font-weight: 920;
}

.task-info span {
  margin-top: 6px;
  color: #64748b;
  font-size: 13px;
  font-weight: 720;
}

.progress-track {
  overflow: hidden;
  height: 8px;
  margin-top: 14px;
  border-radius: 999px;
  background: rgba(219, 231, 244, 0.82);
}

.progress-track i {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #2f7df5, #61d7ff 58%, #ffc0d7);
  box-shadow: 0 0 18px rgba(96, 165, 250, 0.48);
}

.task-info dl {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin: 14px 0 0;
}

.task-info dl div {
  min-width: 0;
  padding: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.48);
}

.task-info dl div:last-child:nth-child(odd) {
  grid-column: 1 / -1;
}

.task-info dt,
.task-info dd {
  margin: 0;
}

.task-info dt {
  color: #8b9ab0;
  font-size: 12px;
  font-weight: 760;
}

.task-info dd {
  margin-top: 4px;
  color: #172642;
  font-size: 13px;
  font-weight: 860;
}

.card-actions {
  justify-content: flex-end;
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

.preview-mask {
  position: fixed;
  inset: 0;
  z-index: 6000;
  display: grid;
  place-items: center;
  padding: 28px;
  background: rgba(18, 31, 55, 0.22);
  backdrop-filter: blur(12px);
}

.preview-dialog {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  width: min(1120px, calc(100vw - 56px));
  height: min(820px, calc(100vh - 56px));
  overflow: hidden;
}

.preview-dialog header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 18px 22px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.62);
}

.preview-dialog header span,
.preview-dialog header strong {
  display: block;
}

.preview-dialog header span {
  color: #64748b;
  font-size: 13px;
  font-weight: 820;
}

.preview-dialog header strong {
  max-width: min(760px, 70vw);
  overflow: hidden;
  color: #10203d;
  font-size: 18px;
  font-weight: 920;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-dialog header button {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
  color: #172642;
  cursor: pointer;
}

.preview-body {
  min-height: 0;
  overflow: auto;
  padding: 22px;
}

.preview-empty {
  display: grid;
  min-height: 360px;
  place-items: center;
  color: #64748b;
  font-weight: 820;
  text-align: center;
}

.preview-media {
  display: block;
  width: 100%;
  max-height: calc(100vh - 190px);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.45);
  object-fit: contain;
}

.preview-media.image {
  margin: 0 auto;
  width: auto;
  max-width: 100%;
}

.preview-audio {
  width: 100%;
  margin-top: 160px;
}

.preview-text {
  min-height: 100%;
  margin: 0;
  padding: 20px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.62);
  color: #10203d;
  font: 14px/1.7 Consolas, Monaco, 'Courier New', monospace;
  white-space: pre-wrap;
  word-break: break-word;
}

.preview-docx {
  min-height: 100%;
  padding: 28px;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  color: #10203d;
  font-size: 15px;
  line-height: 1.75;
}

.preview-docx :deep(p) {
  margin: 0 0 12px;
}

.preview-docx :deep(img) {
  max-width: 100%;
  height: auto;
}

.preview-pdf {
  display: grid;
  gap: 18px;
  justify-items: center;
}

.preview-pdf img {
  width: min(100%, 920px);
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 18px 44px rgba(80, 100, 135, 0.16);
}

.task-board.is-list {
  grid-template-columns: 1fr;
}

.task-board.is-list .task-card {
  grid-template-columns: 76px minmax(0, 1fr) auto;
  grid-template-rows: auto;
  align-items: center;
  min-height: 126px;
}

.task-board.is-list .task-icon {
  min-height: 76px;
}

.task-board.is-list .task-icon .el-icon {
  width: 36px;
  height: 36px;
}

.task-board.is-list .status-chip {
  display: none;
}

.task-board.is-list .task-info dl {
  grid-template-columns: repeat(4, minmax(92px, 1fr));
}

.empty-state {
  display: grid;
  min-height: 300px;
  place-items: center;
  align-content: center;
  gap: 14px;
  color: #64748b;
  text-align: center;
}

.empty-visual {
  display: grid;
  width: 104px;
  height: 104px;
  place-items: center;
  border-radius: 24px;
  background:
    radial-gradient(circle at 25% 20%, rgba(186, 230, 253, 0.62), transparent 48%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 48%),
    rgba(255, 255, 255, 0.68);
  color: #2f7df5;
}

.empty-visual .el-icon {
  width: 52px;
  height: 52px;
}

.empty-state strong {
  color: #10203d;
  font-size: 26px;
  font-weight: 920;
}

.floating-menu,
.create-dialog {
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 24px;
  background:
    radial-gradient(circle at 8% 0%, rgba(186, 230, 253, 0.58), transparent 42%),
    radial-gradient(circle at 100% 100%, rgba(252, 231, 243, 0.54), transparent 44%),
    rgba(255, 255, 255, 0.9);
  box-shadow: 0 24px 72px rgba(92, 120, 166, 0.28);
  backdrop-filter: blur(24px);
}

.floating-menu {
  position: fixed;
  z-index: 5200;
  display: grid;
  gap: 8px;
  width: 230px;
  padding: 14px 12px;
}

.floating-menu p {
  margin: 4px 8px 8px;
  color: #64748b;
  font-size: 14px;
  font-weight: 900;
}

.floating-menu button {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  min-height: 42px;
  border: 0;
  border-radius: 14px;
  background: transparent;
  color: #172642;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  text-align: left;
}

.floating-menu button:hover,
.floating-menu button.active {
  background: rgba(255, 255, 255, 0.72);
  color: #2f7df5;
}

.floating-menu hr {
  width: 100%;
  height: 1px;
  margin: 8px 0;
  border: 0;
  background: rgba(170, 190, 215, 0.48);
}

.danger:hover,
.floating-menu .danger:hover {
  color: #ef4444;
}

.dialog-mask {
  position: fixed;
  inset: 0;
  z-index: 5300;
  display: grid;
  place-items: center;
  padding: 24px;
  background: rgba(219, 234, 254, 0.28);
  backdrop-filter: blur(10px);
}

.create-dialog {
  display: grid;
  gap: 18px;
  width: min(560px, calc(100vw - 48px));
  padding: 22px;
}

.create-dialog header,
.create-dialog footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.create-dialog header strong {
  color: #10203d;
  font-size: 22px;
  font-weight: 920;
}

.create-dialog header button,
.create-dialog footer button {
  min-height: 42px;
  border: 0;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.7);
  color: #172642;
  font-weight: 820;
  cursor: pointer;
}

.create-dialog label {
  display: grid;
  gap: 8px;
  color: #64748b;
  font-weight: 820;
}

.create-dialog input {
  min-height: 48px;
  border: 1px solid rgba(255, 255, 255, 0.74);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.68);
  color: #10203d;
  font: inherit;
  outline: 0;
  padding: 0 14px;
}

@media (max-width: 980px) {
  .offline-hero,
  .offline-toolbar,
  .selection-strip {
    align-items: stretch;
    flex-direction: column;
  }

  .hero-actions,
  .toolbar-actions,
  .board-actions,
  .selection-strip div {
    flex-wrap: wrap;
  }
}

@media (max-width: 640px) {
  .offline-page {
    padding: 12px;
  }

  .glass-panel,
  .task-card {
    border-radius: 22px;
  }

  .offline-hero,
  .offline-toolbar,
  .task-board {
    padding: 18px;
  }

  .offline-hero h1 {
    font-size: 30px;
  }
}
</style>


<template>
  <div class="file-list-container" @click="handleBlankClick">
    <div v-if="fileStore.loading" class="loading-state">
      <el-skeleton :rows="5" animated />
    </div>

    <el-result
      v-else-if="fileStore.error"
      icon="error"
      :sub-title="fileStore.error"
      class="error-state"
    >
      <template #extra>
        <el-button type="primary" @click="handleRetry">{{ t('retry') }}</el-button>
      </template>
    </el-result>

    <el-empty
      v-else-if="fileStore.isEmpty && !props.useFiltered"
      :description="t('emptyFolder')"
      class="empty-state"
    >
      <template #image>
        <el-icon :size="100" color="#909399"><Folder /></el-icon>
      </template>
    </el-empty>

    <el-empty
      v-else-if="props.useFiltered && allItems.length === 0"
      :description="t('noMatchingFiles')"
      class="empty-state"
    >
      <template #image>
        <el-icon :size="100" color="#909399"><Folder /></el-icon>
      </template>
    </el-empty>

    <div v-else class="file-list" :class="`view-mode-${viewMode}`">
      <el-table
        v-if="viewMode === 'table'"
        :data="displayedItems"
        style="width: 100%"
        @row-click="handleRowClick"
        @row-dblclick="handleRowDoubleClick"
        @row-contextmenu="handleRowContextMenu"
      >
        <el-table-column v-if="visibleColumns.name" :label="t('name')" min-width="240">
          <template #default="{ row }">
            <div class="table-name-cell" :class="{ 'is-selected': isSelected(row) }">
              <button
                class="table-select-button"
                type="button"
                :aria-label="isSelected(row) ? t('cancelSelection') : t('select')"
                :title="isSelected(row) ? t('cancelSelection') : t('select')"
                @click.stop="toggleItemSelection(row)"
              >
                <el-icon v-if="isSelected(row)" :size="14"><Check /></el-icon>
              </button>
              <img v-if="getThumbnailSrc(row)" class="table-thumbnail" :src="getThumbnailSrc(row)" :alt="row.name" @error="handleThumbnailError(getThumbnailSrc(row))" />
              <span
                v-else-if="getDisplayIcon(row)"
                class="custom-file-icon table-custom-icon"
                :style="{ '--custom-icon-tint': getDisplayIconTint(row) }"
                :title="row.display_icon_label || row.name"
              >
                {{ getDisplayIcon(row) }}
              </span>
              <el-icon v-else :size="20" class="table-icon">
                <component :is="getItemIcon(row)" />
              </el-icon>
              <div class="table-name-stack">
                <span class="table-name" v-html="highlightText(row.name)"></span>
                <small v-if="!row.isFolder && (row.content_type || row.mime_type)" class="table-meta">
                  {{ row.content_type || row.mime_type }}
                </small>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column v-if="visibleColumns.size && !isMobile()" :label="'\u5927\u5c0f'" width="120">
          <template #default="{ row }">
            <span v-if="!row.isFolder">{{ formatFileSize(row.size) }}</span>
            <span v-else class="folder-indicator">-</span>
          </template>
        </el-table-column>

        <el-table-column v-if="visibleColumns.updatedAt && !isMobile()" :label="'\u4fee\u6539\u65f6\u95f4'" width="180">
          <template #default="{ row }">
            {{ formatTimestamp(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('actions')" :width="isMobile() ? 60 : 80" align="center">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, row)">
              <el-button :icon="More" circle size="small" />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="rename" :icon="Edit">{{ t('rename') }}</el-dropdown-item>
                  <el-dropdown-item v-if="!row.isFolder" command="download" :icon="Download">{{ t('download') }}</el-dropdown-item>
                  <el-dropdown-item command="move" :icon="FolderIcon">{{ t('moveTo') }}</el-dropdown-item>
                  <el-dropdown-item command="copy" :icon="CopyDocument">{{ t('copyTo') }}</el-dropdown-item>
                  <el-dropdown-item v-if="!row.isFolder" command="share" :icon="Share">{{ t('share') }}</el-dropdown-item>
                  <el-dropdown-item v-if="!row.isFolder" command="version-history" :icon="Clock">{{ t('versionHistory') }}</el-dropdown-item>
                  <el-dropdown-item v-if="!row.isFolder" command="collaborate" :icon="User">{{ t('collaboration') }}</el-dropdown-item>
                  <el-dropdown-item command="delete" :icon="Delete" divided>{{ t('delete') }}</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <div
        v-else
        class="grid-view"
        :style="{ '--file-card-min': `${itemSize}px`, '--gallery-size': `${imageSize}px` }"
      >
        <template v-if="viewMode === 'grid'">
          <FileItem
            v-for="item in displayedItems"
            :key="item.id"
            :file="item.isFolder ? undefined : item"
            :folder="item.isFolder ? item : undefined"
            :icon-size="48"
            :show-thumbnails="showThumbnails"
            :selected="isSelected(item)"
            @rename="handleRename"
            @delete="handleDelete"
            @move="handleMove"
            @copy="handleCopy"
            @download="handleDownload"
            @share="handleShare"
            @version-history="handleVersionHistory"
            @collaborate="handleCollaborate"
            @file-click="handleFileClick"
            @file-open="handleFileOpen"
            @folder-click="handleFolderClick"
            @folder-preview="handleFolderPreview"
            @select="toggleItemSelection"
            @context-select="selectSingleItem"
            @modifier-select="toggleItemSelection"
          />
        </template>

        <button
          v-else
          v-for="item in displayedItems"
          :key="item.id"
          class="gallery-item"
          :class="{ 'is-selected': isSelected(item) }"
          type="button"
          :style="{ '--gallery-size': `${imageSize}px` }"
          @click="handleItemClick(item, $event)"
          @dblclick="handleItemDoubleClick(item)"
          @contextmenu.prevent="selectSingleItem(item)"
        >
          <span
            class="gallery-select-button"
            role="button"
            tabindex="0"
            :aria-label="isSelected(item) ? t('cancelSelection') : t('select')"
            :title="isSelected(item) ? t('cancelSelection') : t('select')"
            @click.stop="toggleItemSelection(item)"
            @keydown.enter.stop.prevent="toggleItemSelection(item)"
          >
            <el-icon v-if="isSelected(item)" :size="14"><Check /></el-icon>
          </span>
          <span class="gallery-preview">
            <img v-if="getThumbnailSrc(item)" :src="getThumbnailSrc(item)" :alt="item.name" @error="handleThumbnailError(getThumbnailSrc(item))" />
            <span
              v-else-if="getDisplayIcon(item)"
              class="custom-file-icon gallery-custom-icon"
              :style="{ '--custom-icon-tint': getDisplayIconTint(item) }"
              :title="item.display_icon_label || item.name"
            >
              {{ getDisplayIcon(item) }}
            </span>
            <el-icon v-else :size="Math.min(imageSize * 0.46, 72)" class="gallery-icon">
              <component :is="getItemIcon(item)" />
            </el-icon>
          </span>
          <span class="gallery-caption" v-html="highlightText(item.name)"></span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { t } from '@/utils/language';
import {
  Folder as FolderIcon,
  More,
  Edit,
  Download,
  CopyDocument,
  Delete,
  FolderOpened,
  Picture,
  VideoCamera,
  Headset,
  Document as DocumentIcon,
  Share,
  Clock,
  User,
  Check
} from '@element-plus/icons-vue';
import { useFileStore } from '@/stores/file';
import { useBreakpoint } from '@/composables/useBreakpoint';
import FileItem from '@/components/FileItem/index.vue';
import type { FileItem as FileItemType } from '@/types/file';
import type { FolderItem } from '@/types/folder';
import { getFileTypeByExtension, getFileTypeIcon } from '@/utils/file-type';
import { formatFileSize, formatTimestamp } from '@/utils/format';

interface Props {
  folderId?: number | null;
  viewMode?: 'table' | 'grid' | 'gallery';
  filteredFiles?: FileItemType[];
  useFiltered?: boolean;
  files?: FileItemType[];
  highlightKeyword?: string;
  showThumbnails?: boolean;
  itemSize?: number;
  imageSize?: number;
  pageSize?: number;
  visibleColumns?: {
    name: boolean;
    size: boolean;
    updatedAt: boolean;
  };
  selectedKeys?: string[];
}

interface Emits {
  (e: 'file-click', file: FileItemType): void;
  (e: 'file-open', file: FileItemType): void;
  (e: 'folder-click', folder: FolderItem): void;
  (e: 'folder-preview', folder: FolderItem): void;
  (e: 'refresh'): void;
  (e: 'rename', item: FileItemType | FolderItem): void;
  (e: 'delete', item: FileItemType | FolderItem): void;
  (e: 'move', item: FileItemType | FolderItem): void;
  (e: 'copy', item: FileItemType | FolderItem): void;
  (e: 'download', file: FileItemType): void;
  (e: 'share', file: FileItemType): void;
  (e: 'version-history', file: FileItemType): void;
  (e: 'collaborate', file: FileItemType): void;
  (e: 'toggle-select', item: FileItemType | FolderItem): void;
  (e: 'select-one', item: FileItemType | FolderItem): void;
  (e: 'blank-click'): void;
}

const props = withDefaults(defineProps<Props>(), {
  folderId: null,
  viewMode: 'table',
  filteredFiles: undefined,
  useFiltered: false,
  files: undefined,
  highlightKeyword: '',
  showThumbnails: true,
  itemSize: 216,
  imageSize: 180,
  pageSize: 2000,
  visibleColumns: () => ({
    name: true,
    size: true,
    updatedAt: true,
  }),
  selectedKeys: () => [],
});

const emit = defineEmits<Emits>();
const fileStore = useFileStore();
const { isMobile } = useBreakpoint();
const failedThumbnails = ref(new Set<string>());
const rowClickTimer = ref<number | null>(null);
const ROW_CLICK_DELAY = 220;

type DisplayItem =
  | (FileItemType & { isFolder: false })
  | (FolderItem & { isFolder: true });

const allItems = computed<DisplayItem[]>(() => {
  if (props.files) {
    return props.files.map((file: FileItemType): DisplayItem => ({ ...file, isFolder: false }));
  }

  const folderItems = fileStore.folders.map((folder: FolderItem): DisplayItem => ({ ...folder, isFolder: true }));
  const sourceFiles = props.useFiltered ? fileStore.filteredItems : fileStore.files;
  const fileItems = sourceFiles.map((file: FileItemType): DisplayItem => ({ ...file, isFolder: false }));
  return [...folderItems, ...fileItems];
});

const displayedItems = computed<DisplayItem[]>(() => allItems.value.slice(0, props.pageSize));

const itemKey = (item: DisplayItem | FileItemType | FolderItem) => {
  const isFolder = 'isFolder' in item ? item.isFolder : !('hash' in item);
  return `${isFolder ? 'folder' : 'file'}:${item.id}`;
};

const isSelected = (item: DisplayItem) => props.selectedKeys.includes(itemKey(item));

const toggleItemSelection = (item: DisplayItem | FileItemType | FolderItem) => {
  emit('toggle-select', item);
};

const selectSingleItem = (item: DisplayItem | FileItemType | FolderItem) => {
  emit('select-one', item);
};

const getItemIcon = (item: any) => {
  if (item.isFolder) return FolderOpened;
  const fileType = item.content_type || item.mime_type || getFileTypeByExtension(item.name);
  const iconName = getFileTypeIcon(fileType);
  const iconMap: Record<string, any> = {
    Picture,
    VideoCamera,
    Headset,
    Document: DocumentIcon,
    FolderOpened,
  };
  return iconMap[iconName] || DocumentIcon;
};

const getThumbnailSrc = (item: any) => {
  if (!props.showThumbnails || item.isFolder || !item.thumbnail_url) return '';
  return failedThumbnails.value.has(item.thumbnail_url) ? '' : item.thumbnail_url;
};

const getDisplayIcon = (item: any) => item.display_icon || '';
const getDisplayIconTint = (item: any) => item.display_icon_tint || '#64748b';

const handleThumbnailError = (src: string) => {
  if (!src) return;
  failedThumbnails.value = new Set([...failedThumbnails.value, src]);
};

const highlightText = (text: string): string => {
  if (!props.highlightKeyword || !text) return text;
  const keyword = props.highlightKeyword.trim();
  if (!keyword) return text;
  const regex = new RegExp(`(${keyword})`, 'gi');
  return text.replace(regex, '<mark class="search-highlight">$1</mark>');
};

const handleRowDoubleClick = (row: any) => {
  if (rowClickTimer.value) {
    clearTimeout(rowClickTimer.value);
    rowClickTimer.value = null;
  }
  if (row.isFolder) handleFolderClick(row);
  else handleFileOpen(row as FileItemType);
};

const handleItemClick = (item: DisplayItem, event?: MouseEvent) => {
  if (event?.ctrlKey || event?.metaKey) {
    toggleItemSelection(item);
    return;
  }
  if (rowClickTimer.value) clearTimeout(rowClickTimer.value);
  rowClickTimer.value = window.setTimeout(() => {
    rowClickTimer.value = null;
    if (item.isFolder) {
      handleFolderPreview(item);
      return;
    }
    handleFileClick(item);
  }, ROW_CLICK_DELAY);
};

const handleItemDoubleClick = (item: DisplayItem) => {
  if (rowClickTimer.value) {
    clearTimeout(rowClickTimer.value);
    rowClickTimer.value = null;
  }
  if (item.isFolder) handleFolderClick(item);
  else handleFileOpen(item);
};

const handleFolderClick = (folder: FolderItem) => emit('folder-click', folder);
const handleFolderPreview = (folder: FolderItem) => emit('folder-preview', folder);
const handleFileClick = (file: FileItemType) => emit('file-click', file);
const handleFileOpen = (file: FileItemType) => emit('file-open', file);
const handleRename = (item: FileItemType | FolderItem) => emit('rename', item);
const handleDelete = (item: FileItemType | FolderItem) => emit('delete', item);
const handleMove = (item: FileItemType | FolderItem) => emit('move', item);
const handleCopy = (item: FileItemType | FolderItem) => emit('copy', item);
const handleDownload = (file: FileItemType) => emit('download', file);
const handleShare = (file: FileItemType) => emit('share', file);
const handleVersionHistory = (file: FileItemType) => emit('version-history', file);
const handleCollaborate = (file: FileItemType) => emit('collaborate', file);
const handleRetry = () => emit('refresh');

const handleBlankClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement | null;
  if (!target) return;
  if (
    target.closest('.file-item') ||
    target.closest('.gallery-item') ||
    target.closest('.el-table__row') ||
    target.closest('.el-dropdown') ||
    target.closest('button')
  ) {
    return;
  }
  emit('blank-click');
};

const handleRowClick = (row: any, _column?: unknown, event?: MouseEvent) => {
  if ((event?.ctrlKey || event?.metaKey) && row) {
    toggleItemSelection(row as DisplayItem);
    return;
  }
  if (rowClickTimer.value) clearTimeout(rowClickTimer.value);
  rowClickTimer.value = window.setTimeout(() => {
    rowClickTimer.value = null;
    if (row.isFolder) {
      handleFolderPreview(row as FolderItem);
      return;
    }
    handleFileClick(row as FileItemType);
  }, ROW_CLICK_DELAY);
};

const handleRowContextMenu = (row: any, _column: unknown, event: MouseEvent) => {
  event.preventDefault();
  selectSingleItem(row as DisplayItem);
};

const handleCommand = (command: string, item: any) => {
  const actualItem = item.isFolder
    ? fileStore.folders.find((f: FolderItem) => f.id === item.id)
    : fileStore.files.find((f: FileItemType) => f.id === item.id);

  if (!actualItem) return;

  switch (command) {
    case 'rename':
      handleRename(actualItem);
      break;
    case 'delete':
      handleDelete(actualItem);
      break;
    case 'move':
      handleMove(actualItem);
      break;
    case 'copy':
      handleCopy(actualItem);
      break;
    case 'download':
      if (!item.isFolder) handleDownload(actualItem as FileItemType);
      break;
    case 'share':
      if (!item.isFolder) handleShare(actualItem as FileItemType);
      break;
    case 'version-history':
      if (!item.isFolder) handleVersionHistory(actualItem as FileItemType);
      break;
    case 'collaborate':
      if (!item.isFolder) handleCollaborate(actualItem as FileItemType);
      break;
  }
};
</script>

<style scoped>
.file-list-container {
  min-height: 420px;
}

.loading-state {
  padding: 22px;
}

.empty-state {
  display: grid;
  place-items: center;
  min-height: 430px;
  padding: 72px 0;
  color: #7d8da6;
}

.file-list {
  width: 100%;
}

.table-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.table-select-button,
.gallery-select-button {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border: 1px solid rgba(179, 207, 238, 0.9);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.78);
  color: #fff;
  cursor: pointer;
  flex: 0 0 auto;
  opacity: 0;
  pointer-events: none;
}

.table-name-cell.is-selected .table-select-button,
.gallery-item.is-selected .gallery-select-button {
  border-color: rgba(29, 112, 218, 0.8);
  background: #1d70da;
  opacity: 1;
  pointer-events: auto;
}

.table-thumbnail {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  object-fit: cover;
  flex-shrink: 0;
  border: 1px solid rgba(218, 231, 246, 0.9);
  box-shadow: 0 10px 22px rgba(68, 124, 180, 0.1);
}

.table-icon {
  flex-shrink: 0;
  width: 38px;
  height: 38px;
  padding: 8px;
  border-radius: 12px;
  color: #2d70ff;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(229, 243, 255, 0.72));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 10px 22px rgba(68, 124, 180, 0.1);
}

.custom-file-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--custom-icon-tint, #64748b);
  background: color-mix(in srgb, var(--custom-icon-tint, #64748b) 12%, #ffffff);
  border: 1px solid color-mix(in srgb, var(--custom-icon-tint, #64748b) 26%, #ffffff);
  font-weight: 800;
  line-height: 1;
}

.table-custom-icon {
  flex-shrink: 0;
  width: 38px;
  height: 38px;
  border-radius: 12px;
  font-size: 18px;
}

.table-name-stack {
  display: grid;
  min-width: 0;
  flex: 1;
}

.table-name {
  overflow: hidden;
  color: #14233e;
  font-weight: 760;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.table-meta {
  margin-top: 3px;
  color: #7d8da6;
  font-size: 11px;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.folder-indicator {
  color: #9aa8bc;
}

:deep(.el-table) {
  --el-table-border-color: rgba(219, 234, 249, 0.72);
  --el-table-header-bg-color: rgba(245, 250, 255, 0.76);
  --el-table-tr-bg-color: rgba(255, 255, 255, 0.44);
  --el-table-row-hover-bg-color: rgba(231, 244, 255, 0.72);
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.42);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

:deep(.el-table th.el-table__cell) {
  color: #647892;
  font-weight: 820;
}

:deep(.el-table__cell) {
  border-bottom-color: rgba(219, 234, 249, 0.72);
}

:deep(.el-table__row) {
  cursor: pointer;
}

.grid-view {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--file-card-min, 216px), 1fr));
  gap: 20px;
  padding: 4px 0 16px;
}

.view-mode-gallery .grid-view {
  grid-template-columns: repeat(auto-fill, minmax(calc(var(--gallery-size, 180px) + 44px), max-content));
  align-items: start;
  gap: 28px 34px;
  padding: 26px 18px 34px;
}

.gallery-item {
  position: relative;
  display: grid;
  justify-items: center;
  gap: 12px;
  width: calc(var(--gallery-size, 180px) + 28px);
  padding: 12px;
  border: 1px solid transparent;
  border-radius: 22px;
  background: transparent;
  color: #14233e;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    background 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.gallery-item.is-selected {
  border-color: rgba(45, 112, 255, 0.56);
  background: rgba(220, 237, 255, 0.78);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 18px 36px rgba(45, 112, 255, 0.14);
}

.gallery-select-button {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 2;
}

.gallery-item:hover {
  transform: translateY(-2px);
  border-color: rgba(142, 212, 255, 0.5);
  background: rgba(255, 255, 255, 0.42);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 18px 36px rgba(91, 132, 181, 0.12);
}

.gallery-preview {
  display: grid;
  place-items: center;
  width: var(--gallery-size, 180px);
  height: var(--gallery-size, 180px);
  border-radius: 24px;
  overflow: hidden;
  background:
    radial-gradient(circle at 50% 40%, rgba(255, 255, 255, 0.92), rgba(255, 255, 255, 0.62) 42%, transparent 43%),
    linear-gradient(145deg, rgba(240, 248, 255, 0.82), rgba(255, 255, 255, 0.56));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.94),
    0 18px 34px rgba(91, 132, 181, 0.1);
}

.gallery-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.gallery-icon {
  color: #2d70ff;
}

.gallery-custom-icon {
  min-width: 72px;
  height: 72px;
  padding: 0 12px;
  border-radius: 20px;
  font-size: 34px;
}

.gallery-caption {
  max-width: 100%;
  color: #24344f;
  font-size: 14px;
  font-weight: 760;
  overflow: hidden;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.el-button.is-circle) {
  border-color: rgba(201, 221, 244, 0.9);
  background: rgba(255, 255, 255, 0.72);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 10px 20px rgba(82, 128, 178, 0.1);
}

:deep(.el-button.is-circle:hover) {
  border-color: rgba(91, 184, 246, 0.72);
  color: #1d70da;
}

:deep(.el-empty__description p) {
  color: #7d8da6;
  font-weight: 680;
}

:deep(.el-empty__image svg) {
  filter: drop-shadow(0 14px 24px rgba(114, 143, 176, 0.12));
}

@media (max-width: 768px) {
  .grid-view {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 12px;
    padding: 12px;
  }
}
@media (max-width: 480px) {
  .grid-view {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 8px;
    padding: 8px;
  }
}
:deep(.search-highlight) {
  background-color: #fff566;
  color: #000;
  font-weight: 500;
  padding: 0 2px;
  border-radius: 2px;
}
</style>

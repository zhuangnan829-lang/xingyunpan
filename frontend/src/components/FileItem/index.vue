<template>
  <div
    class="file-item"
    :class="{ 'is-folder': isFolder, 'is-long-pressing': isLongPressing, 'is-selected': selected }"
    @click="handleClick"
    @dblclick="handleDoubleClick"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
    @touchmove="handleTouchMove"
    @contextmenu.prevent="handleContextMenu"
  >
    <button
      class="file-select-button"
      type="button"
      :aria-label="selected ? '取消选择' : '选择'"
      :title="selected ? '取消选择' : '选择'"
      @click.stop="emitSelect"
    >
      <el-icon v-if="selected" :size="16"><Check /></el-icon>
    </button>

    <div class="file-icon">
      <img v-if="thumbnailSrc" class="file-thumbnail" :src="thumbnailSrc" :alt="displayName" @error="handleThumbnailError" />
      <span
        v-else-if="displayIcon"
        class="custom-file-icon"
        :style="{ '--custom-icon-tint': displayIconTint }"
        :title="displayIconLabel"
      >
        {{ displayIcon }}
      </span>
      <el-icon v-else :size="iconSize">
        <component :is="iconComponent" />
      </el-icon>
    </div>

    <div class="file-info">
      <div class="file-name-row">
        <span class="file-name" :title="displayName">{{ displayName }}</span>
        <el-icon v-if="!isFolder && isShared" class="share-indicator" :size="14" title="已分享">
          <Share />
        </el-icon>
      </div>

      <div class="file-meta">
        <span v-if="!isFolder" class="file-size">{{ formattedSize }}</span>
        <span v-else class="folder-count">{{ itemCount }} 项</span>
        <span class="file-time">{{ formattedTime }}</span>
      </div>

      <div v-if="!isFolder && metadataLabel" class="file-metadata">
        {{ metadataLabel }}
      </div>
    </div>

    <div
      class="file-actions"
      @click.stop
      @dblclick.stop
      @mousedown.stop
      @contextmenu.stop.prevent
      @touchstart.stop
    >
      <el-dropdown ref="dropdownRef" trigger="click" @command="handleCommand">
        <el-button :icon="More" circle size="small" @click.stop />
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="rename" :icon="Edit">重命名</el-dropdown-item>
            <el-dropdown-item v-if="!isFolder" command="download" :icon="Download">下载</el-dropdown-item>
            <el-dropdown-item v-if="!isFolder" command="share" :icon="Share">分享</el-dropdown-item>
            <el-dropdown-item v-if="!isFolder" command="version-history" :icon="Clock">版本历史</el-dropdown-item>
            <el-dropdown-item v-if="!isFolder" command="collaborate" :icon="User">协作管理</el-dropdown-item>
            <el-dropdown-item command="move" :icon="Folder">移动到</el-dropdown-item>
            <el-dropdown-item command="copy" :icon="CopyDocument">复制到</el-dropdown-item>
            <el-dropdown-item command="delete" :icon="Delete" divided>删除</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import {
  More,
  Edit,
  Download,
  Folder,
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
import type { FileItem } from '@/types/file';
import type { FolderItem } from '@/types/folder';
import { getFileTypeByExtension, getFileTypeIcon } from '@/utils/file-type';
import { formatFileSize, formatTimestamp } from '@/utils/format';

interface Props {
  file?: FileItem;
  folder?: FolderItem;
  iconSize?: number;
  isShared?: boolean;
  showThumbnails?: boolean;
  selected?: boolean;
}

interface Emits {
  (e: 'rename', item: FileItem | FolderItem): void;
  (e: 'delete', item: FileItem | FolderItem): void;
  (e: 'move', item: FileItem | FolderItem): void;
  (e: 'copy', item: FileItem | FolderItem): void;
  (e: 'download', file: FileItem): void;
  (e: 'share', file: FileItem): void;
  (e: 'version-history', file: FileItem): void;
  (e: 'collaborate', file: FileItem): void;
  (e: 'file-click', file: FileItem): void;
  (e: 'file-open', file: FileItem): void;
  (e: 'folder-click', folder: FolderItem): void;
  (e: 'folder-preview', folder: FolderItem): void;
  (e: 'select', item: FileItem | FolderItem): void;
  (e: 'context-select', item: FileItem | FolderItem): void;
  (e: 'modifier-select', item: FileItem | FolderItem): void;
}

const props = withDefaults(defineProps<Props>(), {
  iconSize: 24,
  isShared: false,
  showThumbnails: true,
  selected: false
});

const emit = defineEmits<Emits>();

const isLongPressing = ref(false);
const longPressTimer = ref<number | null>(null);
const clickTimer = ref<number | null>(null);
const touchStartPos = ref<{ x: number; y: number } | null>(null);
const dropdownRef = ref();
const thumbnailFailed = ref(false);

const LONG_PRESS_DURATION = 500;
const TOUCH_MOVE_THRESHOLD = 10;
const CLICK_DELAY = 220;

const isFolder = computed(() => !!props.folder);
const currentItem = computed(() => props.folder || props.file);
const displayName = computed(() => currentItem.value?.name || '');
const rawThumbnailSrc = computed(() => (props.showThumbnails && !isFolder.value ? props.file?.thumbnail_url || '' : ''));
const thumbnailSrc = computed(() => (thumbnailFailed.value ? '' : rawThumbnailSrc.value));
const displayIcon = computed(() => currentItem.value?.display_icon || '');
const displayIconTint = computed(() => currentItem.value?.display_icon_tint || '#64748b');
const displayIconLabel = computed(() => currentItem.value?.display_icon_label || displayName.value);

const formattedSize = computed(() => (props.file ? formatFileSize(props.file.size) : ''));
const itemCount = computed(() => 0);
const formattedTime = computed(() => (currentItem.value ? formatTimestamp(currentItem.value.created_at) : ''));
const metadataLabel = computed(() => {
  if (!props.file) return '';
  return props.file.content_type || props.file.mime_type || '';
});

const iconComponent = computed(() => {
  if (isFolder.value) return FolderOpened;
  if (!props.file) return DocumentIcon;

  const fileType = props.file.content_type || props.file.mime_type || getFileTypeByExtension(props.file.name);
  const iconName = getFileTypeIcon(fileType);
  const iconMap: Record<string, any> = {
    Picture,
    VideoCamera,
    Headset,
    Document: DocumentIcon,
    FolderOpened,
  };

  return iconMap[iconName] || DocumentIcon;
});

watch(rawThumbnailSrc, () => {
  thumbnailFailed.value = false;
});

const handleThumbnailError = () => {
  thumbnailFailed.value = true;
};

const handleDoubleClick = () => {
  if (clickTimer.value) {
    clearTimeout(clickTimer.value);
    clickTimer.value = null;
  }
  if (isFolder.value && props.folder) emit('folder-click', props.folder);
  if (!isFolder.value && props.file) emit('file-open', props.file);
};

const handleClick = (event: MouseEvent) => {
  if ((event.ctrlKey || event.metaKey) && currentItem.value) {
    emit('modifier-select', currentItem.value);
    return;
  }
  if (clickTimer.value) clearTimeout(clickTimer.value);
  clickTimer.value = window.setTimeout(() => {
    clickTimer.value = null;
    if (isFolder.value && props.folder) {
      emit('folder-preview', props.folder);
      return;
    }
    if (!isFolder.value && props.file) emit('file-click', props.file);
  }, CLICK_DELAY);
};

const handleTouchStart = (event: TouchEvent) => {
  const touch = event.touches[0];
  touchStartPos.value = { x: touch.clientX, y: touch.clientY };
  longPressTimer.value = window.setTimeout(() => {
    isLongPressing.value = true;
    if (navigator.vibrate) navigator.vibrate(50);
    openDropdown();
  }, LONG_PRESS_DURATION);
};

const handleTouchEnd = () => {
  if (longPressTimer.value) {
    clearTimeout(longPressTimer.value);
    longPressTimer.value = null;
  }
  setTimeout(() => {
    isLongPressing.value = false;
  }, 100);
  touchStartPos.value = null;
};

const handleTouchMove = (event: TouchEvent) => {
  if (!touchStartPos.value || !longPressTimer.value) return;
  const touch = event.touches[0];
  const deltaX = Math.abs(touch.clientX - touchStartPos.value.x);
  const deltaY = Math.abs(touch.clientY - touchStartPos.value.y);
  if (deltaX > TOUCH_MOVE_THRESHOLD || deltaY > TOUCH_MOVE_THRESHOLD) {
    clearTimeout(longPressTimer.value);
    longPressTimer.value = null;
    touchStartPos.value = null;
  }
};

const handleContextMenu = (event: MouseEvent) => {
  event.preventDefault();
  const item = currentItem.value;
  if (item) emit('context-select', item);
  openDropdown();
};

const emitSelect = () => {
  const item = currentItem.value;
  if (item) emit('select', item);
};

const openDropdown = () => {
  if (!dropdownRef.value) return;
  const buttonEl = dropdownRef.value.$el.querySelector('button');
  if (buttonEl) buttonEl.click();
};

const handleCommand = (command: string) => {
  const item = currentItem.value;
  if (!item) return;

  switch (command) {
    case 'rename':
      emit('rename', item);
      break;
    case 'delete':
      emit('delete', item);
      break;
    case 'move':
      emit('move', item);
      break;
    case 'copy':
      emit('copy', item);
      break;
    case 'download':
      if (props.file) emit('download', props.file);
      break;
    case 'share':
      if (props.file) emit('share', props.file);
      break;
    case 'version-history':
      if (props.file) emit('version-history', props.file);
      break;
    case 'collaborate':
      if (props.file) emit('collaborate', props.file);
      break;
  }
};
</script>

<style scoped>
.file-item {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 16px;
  min-height: 250px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 24px;
  background:
    radial-gradient(circle at 100% 0%, rgba(114, 218, 255, 0.14), transparent 34%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.68), rgba(247, 251, 255, 0.52));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.88),
    0 18px 36px rgba(91, 132, 181, 0.1);
  backdrop-filter: blur(16px);
  transition:
    background 0.22s ease,
    border-color 0.22s ease,
    box-shadow 0.22s ease,
    transform 0.18s ease;
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
  -webkit-touch-callout: none;
}

.file-item:hover {
  transform: translateY(-3px);
  border-color: rgba(121, 205, 255, 0.72);
  background:
    radial-gradient(circle at 100% 0%, rgba(114, 218, 255, 0.22), transparent 36%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(245, 251, 255, 0.68));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 24px 46px rgba(71, 141, 215, 0.16);
}

.file-item.is-selected {
  border-color: rgba(45, 112, 255, 0.56);
  background:
    radial-gradient(circle at 100% 0%, rgba(114, 218, 255, 0.22), transparent 36%),
    linear-gradient(180deg, rgba(226, 241, 255, 0.94), rgba(210, 230, 250, 0.82));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 24px 46px rgba(45, 112, 255, 0.18);
}

.file-select-button {
  position: absolute;
  top: 96px;
  left: 18px;
  z-index: 2;
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid rgba(179, 207, 238, 0.9);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.78);
  color: #fff;
  cursor: pointer;
  opacity: 0;
  pointer-events: none;
  box-shadow: 0 10px 20px rgba(82, 128, 178, 0.12);
}

.file-item.is-selected .file-select-button {
  border-color: rgba(29, 112, 218, 0.8);
  background: #1d70da;
  opacity: 1;
  pointer-events: auto;
}

.file-item.is-long-pressing {
  background: rgba(231, 245, 255, 0.86);
  transform: scale(0.98);
}

.file-item:hover .file-actions {
  opacity: 1;
}

.file-item.is-folder {
  cursor: pointer;
}

.file-icon {
  flex-shrink: 0;
  order: 2;
  width: 100%;
  height: 156px;
  display: grid;
  place-items: center;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 18px;
  color: #2f72ff;
  background:
    radial-gradient(circle at 48% 40%, rgba(255, 255, 255, 0.96), rgba(255, 255, 255, 0.62) 40%, transparent 41%),
    linear-gradient(145deg, rgba(240, 248, 255, 0.9), rgba(255, 255, 255, 0.74));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.95),
    inset 0 -16px 34px rgba(210, 228, 245, 0.22);
}

.file-item.is-folder .file-icon {
  color: #19aeea;
}

.file-thumbnail {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.custom-file-icon {
  min-width: 58px;
  height: 58px;
  padding: 0 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 16px;
  color: var(--custom-icon-tint, #64748b);
  background: color-mix(in srgb, var(--custom-icon-tint, #64748b) 12%, #ffffff);
  border: 1px solid color-mix(in srgb, var(--custom-icon-tint, #64748b) 26%, #ffffff);
  font-size: 28px;
  font-weight: 800;
  line-height: 1;
}

.file-info {
  order: 1;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.file-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 32px;
  margin-bottom: 6px;
}

.file-name {
  font-size: 15px;
  color: #14233e;
  font-weight: 780;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  min-width: 0;
}

.share-indicator {
  flex-shrink: 0;
  color: #1d70da;
}

.file-meta {
  font-size: 12px;
  color: #7d8da6;
  display: flex;
  gap: 12px;
}

.file-metadata {
  margin-top: 4px;
  font-size: 11px;
  color: #8da0b8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-actions {
  position: absolute;
  top: 14px;
  right: 14px;
  flex-shrink: 0;
  opacity: 0;
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
  transform: translateY(-4px);
}

.file-item:hover .file-actions {
  transform: translateY(0);
}

:deep(.el-button.is-circle) {
  border-color: rgba(201, 221, 244, 0.9);
  background: rgba(255, 255, 255, 0.72);
  color: #536987;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 12px 22px rgba(82, 128, 178, 0.12);
  backdrop-filter: blur(10px);
}

:deep(.el-button.is-circle:hover) {
  border-color: rgba(91, 184, 246, 0.74);
  color: #1d70da;
  background: rgba(244, 251, 255, 0.9);
}

@media (max-width: 768px), (hover: none) {
  .file-actions {
    opacity: 1;
  }

  .file-item {
    padding: 16px;
  }

  .file-name {
    font-size: 15px;
  }

  .file-meta {
    font-size: 13px;
  }
}
</style>

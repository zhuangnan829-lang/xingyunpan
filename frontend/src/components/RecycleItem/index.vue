<template>
  <article
    class="recycle-card"
    :class="{ selected: isSelected, expiring: isExpiringSoon }"
    @contextmenu.prevent="showMenu"
  >
    <div class="card-top">
      <el-checkbox :model-value="isSelected" size="large" @change="handleSelectionChange" />
      <button class="icon-only" type="button" title="更多" @click.stop="showMenu">
        <el-icon><MoreFilled /></el-icon>
      </button>
    </div>

    <div class="file-preview" :class="previewClass">
      <el-icon>
        <component :is="iconComponent" />
      </el-icon>
    </div>

    <div class="file-info">
      <h3 :title="item.fileName">{{ item.fileName }}</h3>
      <p class="path" :title="item.originalPath">{{ normalizedPath }}</p>
    </div>

    <dl class="meta-grid">
      <div>
        <dt>大小</dt>
        <dd>{{ formattedSize }}</dd>
      </div>
      <div>
        <dt>删除时间</dt>
        <dd>{{ formattedDeletedTime }}</dd>
      </div>
    </dl>

    <div class="card-bottom">
      <span class="days" :class="{ warn: isExpiringSoon }">
        <el-icon v-if="isExpiringSoon"><WarningFilled /></el-icon>
        剩余 {{ remainingDays }} 天
      </span>
      <div class="quick-actions">
        <button class="round-action restore" type="button" title="还原" @click="handleRestore">
          <el-icon><RefreshRight /></el-icon>
        </button>
        <button class="round-action delete" type="button" title="永久删除" @click="handlePermanentDelete">
          <el-icon><Delete /></el-icon>
        </button>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="menuVisible" class="menu-backdrop" @click="hideMenu" @contextmenu.prevent="hideMenu">
        <div class="context-menu" :style="menuStyle" @click.stop>
          <button type="button" @click="copyName">
            <el-icon><DocumentCopy /></el-icon>
            <span>复制名称</span>
          </button>
          <button type="button" @click="handleRestore">
            <el-icon><RefreshRight /></el-icon>
            <span>还原</span>
          </button>
          <button class="danger" type="button" @click="handlePermanentDelete">
            <el-icon><Delete /></el-icon>
            <span>永久删除</span>
          </button>
        </div>
      </div>
    </Teleport>
  </article>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import type { CheckboxValueType } from 'element-plus';
import { ElMessage } from 'element-plus';
import {
  Delete,
  Document as DocumentIcon,
  DocumentCopy,
  Files,
  Folder,
  MoreFilled,
  Picture,
  RefreshRight,
  VideoCamera,
  WarningFilled,
} from '@element-plus/icons-vue';
import type { RecycleItem } from '@/types/recycle';
import { useRecycleStore } from '@/stores/recycle';
import { formatFileSize } from '@/utils/format';

interface Props {
  item: RecycleItem;
  isSelected?: boolean;
}

interface Emits {
  (event: 'restore', itemId: string): void;
  (event: 'permanent-delete', itemId: string): void;
  (event: 'selection-change', itemId: string, selected: boolean): void;
}

const props = withDefaults(defineProps<Props>(), {
  isSelected: false,
});
const emit = defineEmits<Emits>();
const recycleStore = useRecycleStore();

const menuVisible = ref(false);
const menuPosition = ref({ x: 0, y: 0 });

const formattedSize = computed(() => formatFileSize(props.item.fileSize));
const formattedDeletedTime = computed(() => formatDateTime(props.item.deletedAt));
const remainingDays = computed(() => recycleStore.getRemainingDays(props.item));
const isExpiringSoon = computed(() => recycleStore.isExpiringSoon(props.item));
const normalizedPath = computed(() => props.item.originalPath || '/');
const previewClass = computed(() => normalizeFileType(props.item.fileType));
const menuStyle = computed(() => ({
  left: `${menuPosition.value.x}px`,
  top: `${menuPosition.value.y}px`,
}));

const iconComponent = computed(() => {
  const iconMap: Record<string, any> = {
    archive: Files,
    document: DocumentIcon,
    excel: DocumentIcon,
    folder: Folder,
    image: Picture,
    pdf: DocumentIcon,
    powerpoint: DocumentIcon,
    text: DocumentIcon,
    video: VideoCamera,
    word: DocumentIcon,
  };

  return iconMap[previewClass.value] || DocumentIcon;
});

function normalizeFileType(type: string) {
  const value = (type || '').toLowerCase();
  if (value.includes('image')) return 'image';
  if (value.includes('video')) return 'video';
  if (value.includes('folder')) return 'folder';
  if (value.includes('pdf')) return 'pdf';
  if (value.includes('word') || value.includes('doc')) return 'word';
  if (value.includes('excel') || value.includes('sheet') || value.includes('xls')) return 'excel';
  if (value.includes('powerpoint') || value.includes('ppt')) return 'powerpoint';
  if (value.includes('zip') || value.includes('rar') || value.includes('archive')) return 'archive';
  if (value.includes('text') || value.includes('txt')) return 'text';
  return value || 'document';
}

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '-';

  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);
  const thisYear = new Date(now.getFullYear(), 0, 1);
  const time = `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;

  if (date >= today) return time;
  if (date >= yesterday) return `昨天 ${time}`;
  if (date >= thisYear) {
    return `${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${time}`;
  }

  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

function handleSelectionChange(selected: CheckboxValueType) {
  emit('selection-change', props.item.id, Boolean(selected));
}

function handleRestore() {
  hideMenu();
  emit('restore', props.item.id);
}

function handlePermanentDelete() {
  hideMenu();
  emit('permanent-delete', props.item.id);
}

function showMenu(event: MouseEvent) {
  const menuWidth = 174;
  const menuHeight = 144;
  menuPosition.value = {
    x: Math.max(12, Math.min(event.clientX, window.innerWidth - menuWidth - 12)),
    y: Math.max(12, Math.min(event.clientY, window.innerHeight - menuHeight - 12)),
  };
  menuVisible.value = true;
}

function hideMenu() {
  menuVisible.value = false;
}

async function copyName() {
  try {
    await navigator.clipboard.writeText(props.item.fileName);
    ElMessage.success('已复制文件名');
  } catch {
    ElMessage.warning('复制失败');
  } finally {
    hideMenu();
  }
}
</script>

<style scoped>
.recycle-card {
  position: relative;
  min-height: 282px;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 8px;
  background:
    radial-gradient(circle at 16% 10%, rgba(108, 190, 255, 0.12), transparent 35%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.78), rgba(246, 252, 255, 0.52));
  box-shadow:
    0 18px 38px rgba(86, 122, 162, 0.13),
    inset 0 1px rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(18px) saturate(1.12);
  transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease;
}

.recycle-card:hover {
  transform: translateY(-2px);
  border-color: rgba(90, 160, 255, 0.48);
  box-shadow:
    0 22px 44px rgba(65, 109, 160, 0.17),
    inset 0 1px rgba(255, 255, 255, 0.96);
}

.recycle-card.selected {
  border-color: rgba(47, 132, 255, 0.76);
  background:
    radial-gradient(circle at 18% 0%, rgba(87, 177, 255, 0.22), transparent 36%),
    linear-gradient(145deg, rgba(224, 241, 255, 0.86), rgba(255, 235, 244, 0.5));
  box-shadow:
    0 22px 48px rgba(57, 132, 255, 0.18),
    inset 0 0 0 1px rgba(255, 255, 255, 0.72);
}

.recycle-card.expiring {
  border-color: rgba(255, 129, 150, 0.68);
}

.card-top,
.card-bottom,
.quick-actions {
  display: flex;
  align-items: center;
}

.card-top,
.card-bottom {
  justify-content: space-between;
}

.icon-only,
.round-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  cursor: pointer;
}

.icon-only {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  color: #657184;
  background: rgba(255, 255, 255, 0.58);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.82);
}

.icon-only:hover {
  color: #1677ff;
  background: rgba(229, 242, 255, 0.9);
}

.file-preview {
  height: 96px;
  margin: 16px 0 14px;
  display: grid;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 8px;
  color: #2f84ff;
  font-size: 44px;
  background:
    radial-gradient(circle at 50% 35%, rgba(255, 255, 255, 0.9), transparent 34%),
    linear-gradient(135deg, rgba(232, 245, 255, 0.82), rgba(255, 249, 252, 0.72));
  box-shadow:
    inset 0 1px rgba(255, 255, 255, 0.9),
    0 12px 26px rgba(77, 127, 180, 0.1);
}

.file-preview.image {
  color: #1977ff;
}

.file-preview.video {
  color: #8a68f1;
}

.file-preview.folder {
  color: #f2a72f;
}

.file-preview.archive {
  color: #18a999;
}

.file-info h3 {
  margin: 0;
  overflow: hidden;
  color: #1d2b42;
  font-size: 16px;
  font-weight: 850;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.path {
  min-height: 20px;
  margin: 7px 0 14px;
  overflow: hidden;
  color: #748197;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin: 0 0 16px;
}

.meta-grid div {
  min-width: 0;
  padding: 9px 10px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.48);
}

.meta-grid dt {
  margin: 0 0 3px;
  color: #8a96a8;
  font-size: 12px;
}

.meta-grid dd {
  margin: 0;
  overflow: hidden;
  color: #46566e;
  font-size: 13px;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.days {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-height: 32px;
  padding: 6px 10px;
  border-radius: 8px;
  color: #4f6178;
  font-size: 13px;
  font-weight: 760;
  background: rgba(255, 255, 255, 0.58);
}

.days.warn {
  color: #e24a62;
  background: rgba(255, 96, 124, 0.12);
}

.quick-actions {
  gap: 8px;
}

.round-action {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.64);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.82);
}

.round-action.restore {
  color: #1677ff;
}

.round-action.delete {
  color: #e24a62;
}

.round-action:hover {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.92);
}

.menu-backdrop {
  position: fixed;
  inset: 0;
  z-index: 3000;
}

.context-menu {
  position: fixed;
  width: 174px;
  padding: 7px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.84);
  box-shadow:
    0 22px 46px rgba(47, 70, 104, 0.2),
    inset 0 1px rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(18px) saturate(1.1);
}

.context-menu button {
  width: 100%;
  min-height: 38px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  border: 0;
  border-radius: 8px;
  color: #243047;
  background: transparent;
  cursor: pointer;
  font-weight: 650;
  text-align: left;
}

.context-menu button:hover {
  background: rgba(47, 132, 255, 0.1);
}

.context-menu button.danger {
  color: #e24a62;
}
</style>

<template>
  <el-dialog
    v-model="dialogVisible"
    :title="title"
    width="600px"
    @close="handleClose"
  >
    <div class="folder-select-content">
      <!-- Current Path -->
      <div class="current-path">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item>
            <a @click.prevent="navigateToFolder(null)">我的文件</a>
          </el-breadcrumb-item>
          <el-breadcrumb-item
            v-for="item in currentPath"
            :key="item.id"
          >
            <a @click.prevent="navigateToFolder(item.id)">{{ item.name }}</a>
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>

      <!-- Folder List -->
      <div class="folder-list">
        <el-scrollbar height="300px">
          <div v-if="loading" class="loading-state">
            <el-skeleton :rows="3" animated />
          </div>

          <el-empty
            v-else-if="folders.length === 0"
            description="此文件夹下没有子文件夹"
            :image-size="80"
          />

          <div v-else class="folder-items">
            <div
              v-for="folder in folders"
              :key="folder.id"
              class="folder-item"
              :class="{ 'is-selected': selectedFolderId === folder.id, 'is-disabled': isDisabled(folder.id) }"
              @click="handleFolderClick(folder)"
              @dblclick="navigateToFolder(folder.id)"
            >
              <el-icon :size="20" class="folder-icon">
                <Folder />
              </el-icon>
              <span class="folder-name">{{ folder.name }}</span>
              <el-icon v-if="selectedFolderId === folder.id" :size="16" class="check-icon">
                <Check />
              </el-icon>
            </div>
          </div>
        </el-scrollbar>
      </div>

      <!-- Selected Folder Info -->
      <div class="selected-info">
        <span class="label">目标位置：</span>
        <span class="value">{{ selectedFolderPath }}</span>
      </div>
    </div>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button
        type="primary"
        :disabled="!canConfirm"
        @click="handleConfirm"
      >
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { Folder, Check } from '@element-plus/icons-vue';
import type { FolderItem, FolderPathItem } from '@/types/folder';
import * as fileAPI from '@/api/file';
import * as folderAPI from '@/api/folder';

interface Props {
  visible: boolean;
  title?: string;
  currentFolderId?: number | null;
  excludeFolderId?: number | null; // Folder to exclude (e.g., the folder being moved)
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'confirm', folderId: number | null): void;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '选择目标文件夹',
  currentFolderId: null,
  excludeFolderId: null
});

const emit = defineEmits<Emits>();

const dialogVisible = ref(props.visible);
const loading = ref(false);
const folders = ref<FolderItem[]>([]);
const currentPath = ref<FolderPathItem[]>([]);
const currentFolderId = ref<number | null>(null);
const selectedFolderId = ref<number | null>(null);

// Computed properties
const selectedFolderPath = computed(() => {
  if (selectedFolderId.value === null) {
    return '我的文件（根目录）';
  }

  const folder = folders.value.find(f => f.id === selectedFolderId.value);
  if (folder) {
    const pathNames = currentPath.value.map(p => p.name);
    return [...pathNames, folder.name].join(' / ');
  }

  return '未选择';
});

const canConfirm = computed(() => {
  // Can't select the same folder as current
  if (selectedFolderId.value === props.currentFolderId) {
    return false;
  }

  // Can't select the excluded folder
  if (selectedFolderId.value === props.excludeFolderId) {
    return false;
  }

  return true;
});

// Check if a folder should be disabled
const isDisabled = (folderId: number) => {
  return folderId === props.excludeFolderId;
};

// Watch for visibility changes
watch(() => props.visible, async (newVal) => {
  dialogVisible.value = newVal;
  if (newVal) {
    // Reset state and load root folders
    selectedFolderId.value = null;
    currentFolderId.value = null;
    currentPath.value = [];
    await loadFolders(null);
  }
});

watch(dialogVisible, (newVal) => {
  emit('update:visible', newVal);
});

// Load folders for a given parent folder
const loadFolders = async (folderId: number | null) => {
  loading.value = true;
  try {
    const response = await fileAPI.getFiles(folderId);
    folders.value = response.folders;
    currentFolderId.value = folderId;

    // Update path
    if (folderId === null) {
      currentPath.value = [];
    } else {
      currentPath.value = await folderAPI.getFolderPath(folderId);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载文件夹列表失败');
  } finally {
    loading.value = false;
  }
};

// Navigate to a folder
const navigateToFolder = async (folderId: number | null) => {
  await loadFolders(folderId);
  selectedFolderId.value = null; // Reset selection when navigating
};

// Handle folder click (select)
const handleFolderClick = (folder: FolderItem) => {
  if (isDisabled(folder.id)) {
    return;
  }
  selectedFolderId.value = folder.id;
};

// Handle close
const handleClose = () => {
  dialogVisible.value = false;
};

// Handle confirm
const handleConfirm = () => {
  if (!canConfirm.value) {
    return;
  }

  // If no folder is selected, use current folder (or root if at root)
  const targetFolderId = selectedFolderId.value !== null 
    ? selectedFolderId.value 
    : currentFolderId.value;

  emit('confirm', targetFolderId);
  dialogVisible.value = false;
};
</script>

<style scoped>
.folder-select-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.current-path {
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.folder-list {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.loading-state {
  padding: 20px;
}

.folder-items {
  padding: 8px;
}

.folder-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
  gap: 8px;
}

.folder-item:hover:not(.is-disabled) {
  background-color: #f5f7fa;
}

.folder-item.is-selected {
  background-color: #ecf5ff;
}

.folder-item.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.folder-icon {
  flex-shrink: 0;
  color: #409eff;
}

.folder-name {
  flex: 1;
  font-size: 14px;
  color: #303133;
}

.check-icon {
  flex-shrink: 0;
  color: #409eff;
}

.selected-info {
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-size: 14px;
}

.selected-info .label {
  color: #606266;
  font-weight: 500;
}

.selected-info .value {
  color: #303133;
}

:deep(.el-dialog__body) {
  padding: 20px;
}

:deep(.el-breadcrumb__item a) {
  color: #409eff;
  transition: color 0.3s;
}

:deep(.el-breadcrumb__item a:hover) {
  color: #66b1ff;
}
</style>

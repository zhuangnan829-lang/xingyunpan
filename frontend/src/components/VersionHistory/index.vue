<template>
  <div class="version-history">
    <a-spin :loading="versionStore.isLoading" class="version-history-spin">
      <div class="version-header">
        <div class="version-title">
          <icon-history />
          <span>版本历史</span>
        </div>
        <div class="version-storage">
          <span class="storage-label">版本占用:</span>
          <span class="storage-value">{{ formatFileSize(totalStorage) }}</span>
        </div>
      </div>

      <div v-if="versionList.length === 0" class="version-empty">
        <a-empty description="暂无版本历史" />
      </div>

      <div v-else class="version-list">
        <div
          v-for="version in versionList"
          :key="version.version_id"
          class="version-item"
          :class="{ 'is-current': version.is_current }"
        >
          <div class="version-info">
            <div class="version-number">
              <span class="version-label">版本 {{ version.version_number }}</span>
              <a-tag v-if="version.is_current" color="arcoblue" size="small">
                当前版本
              </a-tag>
            </div>
            <div class="version-meta">
              <span class="meta-item">
                <icon-user />
                {{ version.uploader_name }}
              </span>
              <span class="meta-item">
                <icon-clock-circle />
                {{ formatTimestamp(version.created_at) }}
              </span>
              <span class="meta-item">
                <icon-file />
                {{ formatFileSize(version.file_size) }}
              </span>
            </div>
          </div>

          <div class="version-actions">
            <a-button
              type="text"
              size="small"
              @click="handleDownload(version)"
            >
              <template #icon>
                <icon-download />
              </template>
              下载
            </a-button>

            <a-button
              v-if="!version.is_current"
              type="text"
              size="small"
              @click="handleRestore(version)"
            >
              <template #icon>
                <icon-undo />
              </template>
              恢复
            </a-button>

            <a-button
              v-if="!version.is_current && versionList.length > 1"
              type="text"
              size="small"
              status="danger"
              @click="handleDelete(version)"
            >
              <template #icon>
                <icon-delete />
              </template>
              删除
            </a-button>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="totalVersions > pageSize" class="version-pagination">
          <a-pagination
            v-model:current="currentPage"
            :total="totalVersions"
            :page-size="pageSize"
            show-total
            show-jumper
            @change="handlePageChange"
          />
        </div>
      </div>
    </a-spin>

    <!-- Restore Confirmation Modal -->
    <a-modal
      v-model:visible="restoreModalVisible"
      title="恢复版本"
      @ok="confirmRestore"
      @cancel="cancelRestore"
    >
      <p>确定要恢复到版本 {{ restoreTarget?.version_number }} 吗？</p>
      <p class="restore-hint">恢复操作会创建一个新版本，不会删除当前版本。</p>
    </a-modal>

    <!-- Delete Confirmation Modal -->
    <a-modal
      v-model:visible="deleteModalVisible"
      title="删除版本"
      @ok="confirmDelete"
      @cancel="cancelDelete"
    >
      <p>确定要删除版本 {{ deleteTarget?.version_number }} 吗？</p>
      <p class="delete-warning">此操作不可恢复！</p>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Message } from '@arco-design/web-vue';
import {
  IconHistory,
  IconUser,
  IconClockCircle,
  IconFile,
  IconDownload,
  IconUndo,
  IconDelete
} from '@arco-design/web-vue/es/icon';
import { useVersionStore } from '@/stores/version';
import type { FileVersion } from '@/api/version';
import { formatFileSize, formatTimestamp } from '@/utils/format';

// Props
interface Props {
  fileId: string;
}

const props = defineProps<Props>();

// Emits
const emit = defineEmits<{
  versionRestored: [];
  versionDeleted: [];
}>();

// Store
const versionStore = useVersionStore();

// State
const restoreModalVisible = ref(false);
const restoreTarget = ref<FileVersion | null>(null);
const deleteModalVisible = ref(false);
const deleteTarget = ref<FileVersion | null>(null);
const currentPage = ref(1);
const pageSize = ref(10);

// Computed
const paginatedData = computed(() => {
  return versionStore.getVersionsPaginated(props.fileId, currentPage.value, pageSize.value);
});

const versionList = computed(() => {
  return paginatedData.value.versions;
});

const totalVersions = computed(() => {
  return paginatedData.value.total;
});

const hasMore = computed(() => {
  return paginatedData.value.hasMore;
});

const totalStorage = computed(() => {
  return versionStore.getVersionsSize(props.fileId);
});

// Watch for fileId changes and load version history (lazy loading)
watch(
  () => props.fileId,
  async (newFileId) => {
    if (newFileId) {
      // Reset pagination when file changes
      currentPage.value = 1;
      
      try {
        // Lazy loading: only load if not already loaded
        await versionStore.loadVersionHistory(newFileId, false);
      } catch (error) {
        Message.error('加载版本历史失败');
        console.error('Failed to load version history:', error);
      }
    }
  },
  { immediate: true }
);

// Methods
function handlePageChange(page: number) {
  currentPage.value = page;
}

async function handleDownload(version: FileVersion) {
  try {
    await versionStore.downloadVersion(props.fileId, version.version_id);
    Message.success('版本下载成功');
  } catch (error) {
    Message.error('版本下载失败');
    console.error('Failed to download version:', error);
  }
}

function handleRestore(version: FileVersion) {
  restoreTarget.value = version;
  restoreModalVisible.value = true;
}

async function confirmRestore() {
  if (!restoreTarget.value) return;

  try {
    await versionStore.restoreVersion(props.fileId, restoreTarget.value.version_id);
    Message.success('版本恢复成功');
    emit('versionRestored');
    restoreModalVisible.value = false;
    restoreTarget.value = null;
    // Reset to first page after restore
    currentPage.value = 1;
  } catch (error) {
    Message.error('版本恢复失败');
    console.error('Failed to restore version:', error);
  }
}

function cancelRestore() {
  restoreModalVisible.value = false;
  restoreTarget.value = null;
}

function handleDelete(version: FileVersion) {
  deleteTarget.value = version;
  deleteModalVisible.value = true;
}

async function confirmDelete() {
  if (!deleteTarget.value) return;

  try {
    await versionStore.deleteVersion(props.fileId, deleteTarget.value.version_id);
    Message.success('版本删除成功');
    emit('versionDeleted');
    deleteModalVisible.value = false;
    deleteTarget.value = null;
  } catch (error) {
    Message.error('版本删除失败');
    console.error('Failed to delete version:', error);
  }
}

function cancelDelete() {
  deleteModalVisible.value = false;
  deleteTarget.value = null;
}
</script>

<style scoped>
.version-history {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.version-history-spin {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.version-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--color-border-2);
}

.version-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
}

.version-storage {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.storage-label {
  color: var(--color-text-3);
}

.storage-value {
  color: var(--color-text-1);
  font-weight: 500;
}

.version-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.version-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
}

.version-pagination {
  display: flex;
  justify-content: center;
  padding: 16px 0;
  margin-top: auto;
  border-top: 1px solid var(--color-border-2);
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  margin-bottom: 12px;
  background-color: var(--color-fill-2);
  border-radius: 4px;
  transition: all 0.3s;
}

.version-item:hover {
  background-color: var(--color-fill-3);
}

.version-item.is-current {
  background-color: rgb(var(--arcoblue-1));
  border: 1px solid rgb(var(--arcoblue-3));
}

.version-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.version-number {
  display: flex;
  align-items: center;
  gap: 8px;
}

.version-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-1);
}

.version-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: var(--color-text-3);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.version-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.restore-hint {
  margin-top: 8px;
  font-size: 13px;
  color: var(--color-text-3);
}

.delete-warning {
  margin-top: 8px;
  font-size: 13px;
  color: var(--color-danger-6);
  font-weight: 500;
}

/* Responsive design */
@media (max-width: 768px) {
  .version-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 12px;
  }

  .version-title {
    font-size: 15px;
  }

  .version-storage {
    font-size: 13px;
  }

  .version-list {
    padding: 12px;
  }

  .version-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 12px;
  }

  .version-info {
    width: 100%;
    gap: 6px;
  }

  .version-label {
    font-size: 13px;
  }

  .version-meta {
    flex-wrap: wrap;
    gap: 12px;
    font-size: 12px;
  }

  .version-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .version-actions .arco-btn {
    font-size: 13px;
  }

  .restore-hint,
  .delete-warning {
    font-size: 12px;
  }
}

@media (max-width: 480px) {
  .version-header {
    padding: 10px;
  }

  .version-title {
    font-size: 14px;
  }

  .version-storage {
    font-size: 12px;
  }

  .version-list {
    padding: 10px;
  }

  .version-item {
    padding: 10px;
    margin-bottom: 10px;
  }

  .version-label {
    font-size: 12px;
  }

  .version-meta {
    font-size: 11px;
    gap: 10px;
  }

  .version-actions {
    flex-direction: column;
    width: 100%;
  }

  .version-actions .arco-btn {
    width: 100%;
    font-size: 12px;
  }

  .version-pagination {
    padding: 12px 0;
  }

  :deep(.arco-pagination) {
    font-size: 12px;
  }
}
</style>

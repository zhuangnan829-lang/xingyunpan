<template>
  <div class="collaborations-page">
    <!-- Page Header -->
    <div class="page-header">
      <h2>协作文件</h2>
      <p class="page-description">查看其他用户与您共享的文件</p>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-container">
      <a-spin size="large" />
      <p>加载协作文件...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="collaborations.length === 0" class="empty-state">
      <a-empty description="暂无协作文件">
        <template #image>
          <icon-folder />
        </template>
      </a-empty>
    </div>

    <!-- Collaborations List -->
    <div v-else class="collaborations-list">
      <a-table
        :columns="columns"
        :data="collaborations"
        :pagination="pagination"
        :loading="isLoading"
        @page-change="handlePageChange"
      >
        <!-- File Name Column -->
        <template #fileName="{ record }">
          <div class="file-name-cell">
            <icon-file :style="{ color: getFileTypeColor(record.file_type) }" />
            <span class="file-name">{{ record.file_name }}</span>
          </div>
        </template>

        <!-- Owner Column -->
        <template #owner="{ record }">
          <div class="owner-cell">
            <icon-user />
            <span>{{ record.owner_name }}</span>
          </div>
        </template>

        <!-- Permission Column -->
        <template #permission="{ record }">
          <a-tag :color="getPermissionColor(record.permission)">
            {{ getPermissionLabel(record.permission) }}
          </a-tag>
        </template>

        <!-- Shared Date Column -->
        <template #sharedAt="{ record }">
          {{ formatDate(record.shared_at) }}
        </template>

        <!-- File Size Column -->
        <template #fileSize="{ record }">
          {{ formatFileSize(record.file_size) }}
        </template>

        <!-- Actions Column -->
        <template #actions="{ record }">
          <a-space>
            <a-button
              type="text"
              size="small"
              @click="handleViewFile(record)"
            >
              <template #icon>
                <icon-eye />
              </template>
              查看
            </a-button>
            <a-button
              v-if="canDownload(record)"
              type="text"
              size="small"
              @click="handleDownloadFile(record)"
            >
              <template #icon>
                <icon-download />
              </template>
              下载
            </a-button>
            <a-button
              v-if="canEdit(record)"
              type="text"
              size="small"
              @click="handleEditFile(record)"
            >
              <template #icon>
                <icon-edit />
              </template>
              编辑
            </a-button>
          </a-space>
        </template>
      </a-table>
    </div>

    <!-- Permission Info -->
    <div v-if="collaborations.length > 0" class="permission-info">
      <a-alert type="info" :show-icon="true">
        <template #icon>
          <icon-info-circle />
        </template>
        <div class="permission-info-content">
          <strong>权限说明：</strong>
          <span class="permission-desc">
            仅查看 - 只能查看文件信息；
            下载 - 可以查看和下载；
            编辑 - 可以查看、下载和编辑文件
          </span>
        </div>
      </a-alert>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { Message } from '@arco-design/web-vue';
import {
  IconFolder,
  IconFile,
  IconUser,
  IconEye,
  IconDownload,
  IconEdit,
  IconInfoCircle
} from '@arco-design/web-vue/es/icon';
import { useCollaborationStore } from '@/stores/collaboration';
import type { CollaborationFile, PermissionLevel } from '@/api/collaboration';

const collaborationStore = useCollaborationStore();

// State
const isLoading = ref(false);
const currentPage = ref(1);
const pageSize = ref(20);

// Collaborations list
const collaborations = computed(() => collaborationStore.myCollaborations);

// Pagination
const pagination = computed(() => ({
  current: currentPage.value,
  pageSize: pageSize.value,
  total: collaborations.value.length,
  showTotal: true,
  showPageSize: true
}));

// Table columns
const columns = [
  {
    title: '文件名',
    dataIndex: 'file_name',
    slotName: 'fileName',
    width: 300
  },
  {
    title: '所有者',
    dataIndex: 'owner_name',
    slotName: 'owner',
    width: 150
  },
  {
    title: '权限',
    dataIndex: 'permission',
    slotName: 'permission',
    width: 120
  },
  {
    title: '共享时间',
    dataIndex: 'shared_at',
    slotName: 'sharedAt',
    width: 150
  },
  {
    title: '文件大小',
    dataIndex: 'file_size',
    slotName: 'fileSize',
    width: 120
  },
  {
    title: '操作',
    slotName: 'actions',
    width: 200,
    fixed: 'right' as const
  }
];

// Load collaborations on mount
onMounted(async () => {
  await loadCollaborations();
});

/**
 * Load collaborations from server
 */
async function loadCollaborations() {
  isLoading.value = true;
  try {
    await collaborationStore.loadMyCollaborations();
  } catch (error: any) {
    Message.error(error.message || '加载协作文件失败');
  } finally {
    isLoading.value = false;
  }
}

/**
 * Handle page change
 */
function handlePageChange(page: number) {
  currentPage.value = page;
}

/**
 * Check if user can download file
 */
function canDownload(file: CollaborationFile): boolean {
  return file.permission === 'download' || file.permission === 'edit';
}

/**
 * Check if user can edit file
 */
function canEdit(file: CollaborationFile): boolean {
  return file.permission === 'edit';
}

/**
 * Handle view file
 */
function handleViewFile(file: CollaborationFile) {
  Message.info(`查看文件: ${file.file_name}`);
  // TODO: Implement file preview/view functionality
}

/**
 * Handle download file
 */
function handleDownloadFile(file: CollaborationFile) {
  if (!canDownload(file)) {
    Message.warning('您没有下载权限');
    return;
  }
  
  Message.info(`下载文件: ${file.file_name}`);
  // TODO: Implement file download functionality
}

/**
 * Handle edit file
 */
function handleEditFile(file: CollaborationFile) {
  if (!canEdit(file)) {
    Message.warning('您没有编辑权限');
    return;
  }
  
  Message.info(`编辑文件: ${file.file_name}`);
  // TODO: Implement file edit functionality
}

/**
 * Get permission label
 */
function getPermissionLabel(permission: PermissionLevel): string {
  const labels: Record<PermissionLevel, string> = {
    view: '仅查看',
    download: '下载',
    edit: '编辑'
  };
  return labels[permission] || '未知';
}

/**
 * Get permission color
 */
function getPermissionColor(permission: PermissionLevel): string {
  const colors: Record<PermissionLevel, string> = {
    view: 'gray',
    download: 'orange',
    edit: 'green'
  };
  return colors[permission] || 'gray';
}

/**
 * Get file type color
 */
function getFileTypeColor(fileType: string): string {
  if (fileType.startsWith('image/')) return '#00b42a';
  if (fileType.startsWith('video/')) return '#ff7d00';
  if (fileType.startsWith('audio/')) return '#722ed1';
  if (fileType.includes('pdf')) return '#f53f3f';
  if (fileType.includes('word') || fileType.includes('document')) return '#165dff';
  if (fileType.includes('excel') || fileType.includes('sheet')) return '#14c9c9';
  return '#86909c';
}

/**
 * Format date string
 */
function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
}

/**
 * Format file size
 */
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
}
</script>

<style scoped>
.collaborations-page {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.page-description {
  margin: 0;
  color: var(--color-text-3);
  font-size: 14px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  gap: 16px;
}

.loading-container p {
  color: var(--color-text-3);
  font-size: 14px;
}

.empty-state {
  display: flex;
  justify-content: center;
  padding: 80px 20px;
}

.empty-state :deep(.arco-empty-icon) {
  font-size: 64px;
  color: var(--color-text-4);
}

.collaborations-list {
  background-color: var(--color-bg-2);
  border-radius: 4px;
  padding: 16px;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-name {
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.owner-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.permission-info {
  margin-top: 16px;
}

.permission-info-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.permission-desc {
  font-size: 13px;
  color: var(--color-text-2);
}

/* Responsive design */
@media (max-width: 768px) {
  .collaborations-page {
    padding: 12px;
  }

  .page-header {
    margin-bottom: 16px;
  }

  .page-header h2 {
    font-size: 18px;
  }

  .page-description {
    font-size: 13px;
  }

  .collaborations-list {
    padding: 8px;
    overflow-x: auto;
  }

  .permission-info-content {
    font-size: 12px;
  }

  .file-name-cell {
    gap: 6px;
  }

  .file-name {
    font-size: 13px;
  }

  .owner-cell {
    font-size: 13px;
  }

  .loading-container {
    padding: 60px 16px;
  }

  .empty-state {
    padding: 60px 16px;
  }

  /* Hide less important columns on mobile */
  :deep(.arco-table-th:nth-child(5)),
  :deep(.arco-table-td:nth-child(5)) {
    display: none;
  }

  :deep(.arco-table) {
    font-size: 13px;
  }

  :deep(.arco-table-th),
  :deep(.arco-table-td) {
    padding: 8px 6px;
  }

  :deep(.arco-btn) {
    font-size: 13px;
    padding: 4px 10px;
  }

  :deep(.arco-tag) {
    font-size: 11px;
  }
}

@media (max-width: 480px) {
  .collaborations-page {
    padding: 10px;
  }

  .page-header h2 {
    font-size: 16px;
  }

  .page-description {
    font-size: 12px;
  }

  .collaborations-list {
    padding: 6px;
  }

  .file-name {
    font-size: 12px;
  }

  .owner-cell {
    font-size: 12px;
  }

  .permission-info-content {
    font-size: 11px;
  }

  .loading-container,
  .empty-state {
    padding: 40px 12px;
  }

  /* Hide owner column on very small screens */
  :deep(.arco-table-th:nth-child(2)),
  :deep(.arco-table-td:nth-child(2)) {
    display: none;
  }

  :deep(.arco-table) {
    font-size: 12px;
  }

  :deep(.arco-table-th),
  :deep(.arco-table-td) {
    padding: 6px 4px;
  }

  :deep(.arco-btn) {
    font-size: 12px;
    padding: 3px 8px;
  }

  :deep(.arco-pagination) {
    font-size: 12px;
  }
}
</style>

<template>
  <div class="shares-container">
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="toolbar-left">
        <h2 class="page-title">我的分享</h2>
      </div>
      <div class="toolbar-right">
        <!-- Search box -->
        <el-input
          v-model="searchKeyword"
          placeholder="按文件名搜索"
          clearable
          :prefix-icon="Search"
          style="width: 240px"
          @input="handleSearch"
          @clear="handleSearch"
        />
        <!-- Batch delete button -->
        <el-button
          v-if="selectedShares.length > 0"
          type="danger"
          :icon="Delete"
          @click="handleBatchDelete"
        >
          批量删除 ({{ selectedShares.length }})
        </el-button>
      </div>
    </div>

    <!-- Share list table -->
    <div class="table-wrapper">
      <el-table
        v-loading="shareStore.isLoading"
        :data="filteredShares"
        @selection-change="handleSelectionChange"
        empty-text=""
        row-key="share_id"
      >
        <!-- Selection column -->
        <el-table-column type="selection" width="50" />

        <!-- File name column -->
        <el-table-column label="文件名" min-width="180">
          <template #default="{ row }">
            <div class="file-names">
              <el-icon class="file-icon"><Document /></el-icon>
              <span :class="{ 'expired-text': shareStore.isExpired(row) }">
                {{ row.file_names.join(', ') || '未知文件' }}
              </span>
            </div>
          </template>
        </el-table-column>

        <!-- Share link column -->
        <el-table-column label="分享链接" min-width="220">
          <template #default="{ row }">
            <div class="link-cell">
              <span
                class="link-text"
                :class="{ 'expired-text': shareStore.isExpired(row) }"
              >
                {{ getShareLink(row.share_id) }}
              </span>
              <el-button
                type="primary"
                link
                :icon="CopyDocument"
                :disabled="shareStore.isExpired(row)"
                @click="handleCopyLink(row)"
              >
                复制
              </el-button>
            </div>
          </template>
        </el-table-column>

        <!-- Created time column -->
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            <span :class="{ 'expired-text': shareStore.isExpired(row) }">
              {{ formatDate(row.created_at) }}
            </span>
          </template>
        </el-table-column>

        <!-- Expiration column -->
        <el-table-column label="有效期" width="160">
          <template #default="{ row }">
            <span :class="{ 'expired-text': shareStore.isExpired(row) }">
              {{ formatExpiry(row.expires_at) }}
            </span>
          </template>
        </el-table-column>

        <!-- Download count column -->
        <el-table-column label="下载次数" width="100" align="center">
          <template #default="{ row }">
            <span :class="{ 'expired-text': shareStore.isExpired(row) }">
              {{ row.download_count }}
            </span>
          </template>
        </el-table-column>

        <!-- Status column -->
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="shareStore.isExpired(row) ? 'info' : 'success'"
              size="small"
            >
              {{ shareStore.isExpired(row) ? '已过期' : '有效' }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- Actions column -->
        <el-table-column label="操作" width="80" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              type="danger"
              link
              :icon="Delete"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Empty state -->
      <div v-if="!shareStore.isLoading && filteredShares.length === 0" class="empty-state">
        <el-empty
          :description="searchKeyword ? '未找到匹配的分享记录' : '暂无分享记录'"
        >
          <el-button v-if="searchKeyword" @click="searchKeyword = ''; handleSearch()">
            清除搜索
          </el-button>
        </el-empty>
      </div>
    </div>

    <!-- Confirm delete dialog -->
    <el-dialog
      v-model="deleteDialogVisible"
      title="确认删除"
      width="400px"
    >
      <p>{{ deleteDialogMessage }}</p>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDelete">确认删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Delete, Document, CopyDocument } from '@element-plus/icons-vue';
import { useShareStore } from '@/stores/share';
import { copyToClipboard } from '@/utils/share-utils';
import type { ShareLink } from '@/stores/share';

const shareStore = useShareStore();

const searchKeyword = ref('');
const selectedShares = ref<ShareLink[]>([]);
const deleteDialogVisible = ref(false);
const deleteTarget = ref<ShareLink | null>(null);
const isBatchDelete = ref(false);

const deleteDialogMessage = computed(() => {
  if (isBatchDelete.value) {
    return `确定要删除选中的 ${selectedShares.value.length} 条分享记录吗？`;
  }
  if (deleteTarget.value) {
    return `确定要删除 "${deleteTarget.value.file_names.join(', ')}" 的分享链接吗？`;
  }
  return '';
});

// Filter shares based on search keyword
const filteredShares = computed(() => {
  if (!searchKeyword.value.trim()) {
    return shareStore.shares;
  }
  const keyword = searchKeyword.value.toLowerCase();
  return shareStore.shares.filter(share =>
    share.file_names.some(name => name.toLowerCase().includes(keyword))
  );
});

function getShareLink(shareId: string): string {
  const share = shareStore.shares.find(s => s.share_id === shareId);
  return share?.share_url || '';
}

function formatDate(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatExpiry(expiresAt: string | null): string {
  if (!expiresAt) return '永久有效';
  const date = new Date(expiresAt);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function handleSearch() {
  // Search is handled by computed property
}

function handleSelectionChange(rows: ShareLink[]) {
  selectedShares.value = rows;
}

async function handleCopyLink(share: ShareLink) {
  const link = share.share_url;
  const success = await copyToClipboard(link);
  if (success) {
    ElMessage.success('链接已复制到剪贴板');
  } else {
    ElMessage.error('复制失败，请手动复制');
  }
}

function handleDelete(share: ShareLink) {
  deleteTarget.value = share;
  isBatchDelete.value = false;
  deleteDialogVisible.value = true;
}

function handleBatchDelete() {
  if (selectedShares.value.length === 0) return;
  isBatchDelete.value = true;
  deleteTarget.value = null;
  deleteDialogVisible.value = true;
}

async function confirmDelete() {
  try {
    if (isBatchDelete.value) {
      const shareIds = selectedShares.value.map(s => s.share_id);
      await shareStore.deleteShares(shareIds);
      ElMessage.success(`已删除 ${shareIds.length} 条分享记录`);
      selectedShares.value = [];
    } else if (deleteTarget.value) {
      await shareStore.deleteShare(deleteTarget.value.share_id);
      ElMessage.success('分享链接已删除');
      deleteTarget.value = null;
    }
    deleteDialogVisible.value = false;
  } catch (error: any) {
    ElMessage.error(error.message || '删除失败');
  }
}

onMounted(async () => {
  try {
    await shareStore.loadMyShares();
  } catch (error: any) {
    ElMessage.error(error.message || '加载分享列表失败');
  }
});
</script>

<style scoped>
.shares-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #f5f7fa;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  flex-wrap: wrap;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.table-wrapper {
  flex: 1;
  padding: 16px 24px;
  overflow: auto;
}

.file-names {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
}

.file-icon {
  flex-shrink: 0;
  color: var(--el-color-primary);
}

.file-names span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
}

.link-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.expired-text {
  color: var(--el-text-color-placeholder);
}

.empty-state {
  padding: 60px 0;
  text-align: center;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0 0;
}

/* Mobile responsive */
@media (max-width: 767px) {
  .toolbar {
    padding: 12px 16px;
  }

  .table-wrapper {
    padding: 12px 16px;
  }

  .toolbar-right {
    width: 100%;
  }

  .toolbar-right .el-input {
    flex: 1;
  }

  .page-title {
    font-size: 16px;
  }

  /* Hide less important columns on mobile */
  :deep(.el-table__header-wrapper),
  :deep(.el-table__body-wrapper) {
    .el-table-column--selection {
      width: 40px !important;
    }
  }

  /* Stack table cells vertically on very small screens */
  :deep(.el-table__body .el-table__row) {
    display: flex;
    flex-direction: column;
    border-bottom: 2px solid #e4e7ed;
    padding: 12px 0;
  }

  :deep(.el-table__body .el-table__cell) {
    border-bottom: none !important;
    padding: 4px 8px;
  }

  :deep(.el-table__header) {
    display: none;
  }

  .link-text {
    font-size: 11px;
  }
}

@media (min-width: 768px) and (max-width: 1024px) {
  /* Tablet view - hide download count column */
  :deep(.el-table__header-wrapper),
  :deep(.el-table__body-wrapper) {
    .el-table__cell:nth-child(6) {
      display: none;
    }
  }
}

@media (max-width: 480px) {
  .toolbar {
    padding: 10px 12px;
  }

  .table-wrapper {
    padding: 10px 12px;
  }

  .page-title {
    font-size: 15px;
  }

  .toolbar-right .el-button {
    font-size: 13px;
    padding: 8px 12px;
  }
}
</style>

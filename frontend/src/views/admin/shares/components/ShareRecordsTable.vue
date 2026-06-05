<template>
  <section class="table-panel">
    <el-table
      v-loading="loading"
      :data="records"
      row-key="share_id"
      empty-text="暂无分享记录"
      class="share-table"
      @selection-change="$emit('selection-change', $event)"
    >
      <el-table-column type="selection" width="52" />
      <el-table-column label="#" width="64">
        <template #default="{ row }">
          <span class="muted">#{{ row.index }}</span>
        </template>
      </el-table-column>
      <el-table-column label="源文件" min-width="260">
        <template #default="{ row }">
          <div class="source-cell">
            <span class="file-mark"><Document /></span>
            <div>
              <strong>{{ row.sourceLabel }}</strong>
              <small>ID {{ row.share_id }}</small>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="浏览" width="100" align="center" prop="access_count" />
      <el-table-column label="下载" width="100" align="center" prop="download_count" />
      <el-table-column label="积分" width="100" align="center" prop="score" />
      <el-table-column label="自动过期" width="180">
        <template #default="{ row }">
          <span :class="['expiry-pill', { expired: row.isUnavailable }]">{{ row.expiryText }}</span>
        </template>
      </el-table-column>
      <el-table-column label="分享者" width="140" prop="ownerName" />
      <el-table-column label="分享于" width="180" prop="createdText" />
      <el-table-column label="状态" width="116" align="center">
        <template #default="{ row }">
          <span :class="['status-pill', row.isUnavailable ? 'expired' : 'active']">{{ row.statusText }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="132" fixed="right" align="right">
        <template #default="{ row }">
          <button class="icon-action" type="button" title="复制链接" :disabled="row.isUnavailable" @click="$emit('copy', row)">
            <CopyDocument />
          </button>
          <button class="icon-action danger" type="button" title="删除" @click="$emit('delete', row)">
            <Delete />
          </button>
        </template>
      </el-table-column>
    </el-table>
  </section>
</template>

<script setup lang="ts">
import { CopyDocument, Delete, Document } from '@element-plus/icons-vue';
import type { ShareDisplayRecord } from '../types';

defineProps<{
  loading: boolean;
  records: ShareDisplayRecord[];
}>();

defineEmits<{
  copy: [record: ShareDisplayRecord];
  delete: [record: ShareDisplayRecord];
  'selection-change': [records: ShareDisplayRecord[]];
}>();
</script>

<style scoped>
.table-panel {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 22px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.74), rgba(255, 255, 255, 0.4));
  box-shadow: 0 18px 40px rgba(67, 93, 121, 0.11), inset 0 1px 0 rgba(255, 255, 255, 0.86);
  backdrop-filter: blur(16px);
}

.share-table {
  --el-table-border-color: rgba(214, 226, 238, 0.58);
  --el-table-header-bg-color: rgba(255, 255, 255, 0.34);
  --el-table-row-hover-bg-color: rgba(228, 246, 255, 0.46);
  background: transparent;
}

:deep(.el-table__inner-wrapper::before) {
  display: none;
}

:deep(.el-table th.el-table__cell),
:deep(.el-table tr),
:deep(.el-table td.el-table__cell) {
  background: transparent;
}

:deep(.el-table th.el-table__cell) {
  color: #1f2d3d;
  font-weight: 820;
}

.muted,
.source-cell small {
  color: #718096;
}

.source-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.source-cell div {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.source-cell strong {
  overflow: hidden;
  color: #152238;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-mark {
  display: inline-grid;
  place-items: center;
  width: 38px;
  height: 38px;
  flex: 0 0 auto;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(125, 211, 252, 0.42), rgba(255, 192, 203, 0.34));
  color: #1687d9;
}

.expiry-pill,
.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 780;
  white-space: nowrap;
}

.expiry-pill {
  background: rgba(255, 255, 255, 0.52);
  color: #4b6076;
}

.expiry-pill.expired,
.status-pill.expired {
  background: rgba(148, 163, 184, 0.18);
  color: #64748b;
}

.status-pill.active {
  background: rgba(209, 250, 229, 0.62);
  color: #059669;
}

.icon-action {
  display: inline-grid;
  place-items: center;
  width: 34px;
  height: 34px;
  margin-left: 6px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.55);
  color: #1687d9;
  cursor: pointer;
}

.icon-action:disabled {
  cursor: not-allowed;
  opacity: 0.42;
}

.icon-action.danger {
  color: #d94a5f;
}

.icon-action svg {
  width: 17px;
  height: 17px;
}
</style>

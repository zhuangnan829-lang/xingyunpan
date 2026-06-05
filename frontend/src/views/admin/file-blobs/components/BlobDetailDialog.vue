<template>
  <div v-if="visible && blob" class="dialog-overlay" @click.self="$emit('close')">
    <div class="dialog-card">
      <div class="dialog-head">
        <div>
          <h2>Blob 详情</h2>
        </div>
        <button class="close-button" type="button" @click="$emit('close')">
          <el-icon><Close /></el-icon>
        </button>
      </div>

      <section class="summary-grid">
        <div class="summary-item">
          <span>ID</span>
          <strong>{{ blob.id }}</strong>
        </div>
        <div class="summary-item">
          <span>大小</span>
          <strong>{{ formatSize(blob.sizeBytes) }}</strong>
        </div>
        <div class="summary-item">
          <span>类型</span>
          <strong>{{ blob.kindLabel }}</strong>
        </div>
        <div class="summary-item">
          <span>引用次数</span>
          <strong>{{ blob.referenceCount }}</strong>
        </div>
        <div class="summary-item">
          <span>创建者</span>
          <strong class="creator-line">
            <em>{{ blob.creatorBadge }}</em>
            <span>{{ blob.creatorName }}</span>
          </strong>
        </div>
      </section>

      <section class="detail-grid">
        <div class="detail-item">
          <span>上传会话 ID</span>
          <strong>{{ blob.uploadSessionId || '-' }}</strong>
        </div>
        <div class="detail-item">
          <span>源</span>
          <strong class="wrap">{{ blob.source }}</strong>
        </div>
        <div class="detail-item">
          <span>存储策略</span>
          <strong>{{ blob.storagePolicyName }}</strong>
        </div>
        <div class="detail-item">
          <span>加密</span>
          <strong>{{ blob.encrypted ? '已加密' : '未加密' }}</strong>
        </div>
      </section>

      <section class="linked-section">
        <h3>关联文件</h3>
        <div class="linked-table">
          <div class="linked-head linked-grid">
            <span>#</span>
            <span>文件名</span>
            <span>大小</span>
            <span>所有者</span>
            <span>创建于</span>
          </div>

          <div v-for="file in blob.linkedFiles" :key="file.id" class="linked-row linked-grid">
            <span>{{ file.id }}</span>
            <span class="filename">
              <em>{{ file.extension }}</em>
              <strong>{{ file.name }}</strong>
            </span>
            <span>{{ formatSize(file.sizeBytes) }}</span>
            <span class="owner">
              <i>{{ blob.creatorBadge }}</i>
              <strong>{{ file.ownerName }}</strong>
            </span>
            <span>{{ formatDate(file.createdAt) }}</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Close } from '@element-plus/icons-vue';
import type { BlobRecord } from '../types';

defineProps<{
  visible: boolean;
  blob: BlobRecord | null;
  formatSize: (value: number) => string;
  formatDate: (value: string) => string;
}>();

defineEmits<{
  close: [];
}>();
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  inset: 0;
  z-index: 70;
  display: grid;
  place-items: center;
  padding: 28px;
  background: rgba(15, 23, 42, 0.38);
  backdrop-filter: blur(2px);
}

.dialog-card {
  width: min(1120px, calc(100vw - 56px));
  max-height: calc(100vh - 56px);
  overflow: auto;
  padding: 32px 34px;
  border-radius: 26px;
  background: #ffffff;
  box-shadow: 0 28px 60px rgba(15, 23, 42, 0.26);
}

.dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 26px;
}

.dialog-head h2 {
  margin: 0;
  color: #20232d;
  font-size: 28px;
  font-weight: 800;
}

.close-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border: 0;
  border-radius: 14px;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
}

.summary-grid,
.detail-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 16px;
}

.detail-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
  margin-top: 26px;
}

.summary-item,
.detail-item {
  display: grid;
  gap: 10px;
}

.summary-item span,
.detail-item span {
  color: #1f2937;
  font-size: 14px;
  font-weight: 700;
}

.summary-item strong,
.detail-item strong {
  color: #4b5563;
  font-size: 18px;
  font-weight: 500;
}

.wrap {
  line-height: 1.45;
  word-break: break-all;
}

.creator-line,
.owner,
.filename {
  display: inline-flex;
  align-items: center;
  gap: 12px;
}

.creator-line em,
.owner i,
.filename em {
  display: inline-grid;
  place-items: center;
  min-width: 34px;
  height: 34px;
  padding: 0 10px;
  border-radius: 999px;
  background: #14c8a8;
  color: #ffffff;
  font-style: normal;
  font-size: 18px;
  font-weight: 800;
}

.filename em {
  min-width: 28px;
  height: 28px;
  border-radius: 8px;
  background: #5e8ce8;
  font-size: 16px;
}

.linked-section {
  margin-top: 34px;
}

.linked-section h3 {
  margin: 0 0 16px;
  color: #20232d;
  font-size: 18px;
}

.linked-table {
  overflow: hidden;
  border: 1px solid #dfe6ef;
  border-radius: 20px;
}

.linked-grid {
  display: grid;
  grid-template-columns: 90px minmax(220px, 2.1fr) 160px 190px 220px;
  gap: 14px;
  align-items: center;
}

.linked-head,
.linked-row {
  padding: 16px 22px;
}

.linked-head {
  background: #ffffff;
  color: #1f2937;
  font-weight: 800;
}

.linked-row {
  border-top: 1px solid #edf2f7;
  color: #4b5563;
  font-size: 15px;
}

@media (max-width: 1100px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .detail-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .linked-table {
    overflow-x: auto;
  }

  .linked-grid {
    min-width: 860px;
  }
}

@media (max-width: 720px) {
  .dialog-overlay {
    padding: 14px;
  }

  .dialog-card {
    width: calc(100vw - 28px);
    padding: 24px 18px;
  }

  .summary-grid,
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>

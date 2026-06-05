<template>
  <div class="upload-task-item">
    <div class="task-header">
      <div class="file-info">
        <el-icon class="file-icon"><Document /></el-icon>
        <div class="file-details">
          <div class="file-name">{{ task.file.name }}</div>
          <div class="file-size">{{ formatFileSize(task.file.size) }}</div>
        </div>
      </div>
      <div class="task-actions">
        <el-button
          v-if="task.status === 'uploading' || task.status === 'hashing'"
          type="danger"
          size="small"
          text
          @click="$emit('cancel', task.id)"
        >
          取消
        </el-button>
        <el-button
          v-if="task.status === 'failed'"
          type="primary"
          size="small"
          text
          @click="$emit('retry', task.id)"
        >
          重试
        </el-button>
      </div>
    </div>

    <div class="task-progress">
      <ProgressBar :progress="task.progress" :status="task.status" />
    </div>

    <div class="task-status">
      <div class="status-message">
        <span v-if="task.status === 'pending'">等待上传...</span>
        <span v-else-if="task.status === 'hashing'">计算文件哈希中...</span>
        <span v-else-if="task.status === 'uploading' && task.isMultipart">
          上传中 {{ task.completedChunks?.length || 0 }}/{{ task.totalChunks }} 分片
        </span>
        <span v-else-if="task.status === 'uploading'">上传中...</span>
        <span v-else-if="task.status === 'completed'" class="success">上传成功</span>
        <span v-else-if="task.status === 'failed'" class="error">
          上传失败{{ task.error ? `: ${task.error}` : '' }}
        </span>
        <span v-else-if="task.status === 'cancelled'" class="warning">已取消</span>
      </div>
      <div v-if="task.status === 'uploading' && task.speed > 0" class="upload-info">
        <span class="upload-speed">{{ formatSpeed(task.speed) }}</span>
        <span class="remaining-time">剩余 {{ formatRemainingTime(task) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Document } from '@element-plus/icons-vue';
import ProgressBar from '../ProgressBar/index.vue';
import type { UploadTask } from '@/types/upload';
import { formatFileSize } from '@/utils/format';

interface Props {
  task: UploadTask;
}

interface Emits {
  (e: 'cancel', taskId: string): void;
  (e: 'retry', taskId: string): void;
}

const props = defineProps<Props>();
defineEmits<Emits>();

function formatSpeed(bytesPerSecond: number): string {
  if (bytesPerSecond < 1024) {
    return `${bytesPerSecond.toFixed(0)} B/s`;
  } else if (bytesPerSecond < 1024 * 1024) {
    return `${(bytesPerSecond / 1024).toFixed(2)} KB/s`;
  } else if (bytesPerSecond < 1024 * 1024 * 1024) {
    return `${(bytesPerSecond / (1024 * 1024)).toFixed(2)} MB/s`;
  } else {
    return `${(bytesPerSecond / (1024 * 1024 * 1024)).toFixed(2)} GB/s`;
  }
}

function formatRemainingTime(task: UploadTask): string {
  if (!task.speed || task.speed === 0 || task.progress >= 100) {
    return '--';
  }

  const remainingBytes = task.file.size * (1 - task.progress / 100);
  const remainingSeconds = Math.ceil(remainingBytes / task.speed);

  if (remainingSeconds < 60) {
    return `${remainingSeconds}秒`;
  } else if (remainingSeconds < 3600) {
    const minutes = Math.floor(remainingSeconds / 60);
    return `${minutes}分钟`;
  } else {
    const hours = Math.floor(remainingSeconds / 3600);
    const minutes = Math.floor((remainingSeconds % 3600) / 60);
    return `${hours}小时${minutes}分钟`;
  }
}
</script>

<style scoped>
.upload-task-item {
  padding: 16px;
  border-bottom: 1px solid rgba(219, 234, 249, 0.76);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.5), rgba(247, 252, 255, 0.36));
}

.upload-task-item:last-child {
  border-bottom: none;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.file-info {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.file-icon {
  font-size: 24px;
  color: #2d70ff;
  margin-right: 12px;
  flex-shrink: 0;
  filter: drop-shadow(0 8px 14px rgba(45, 112, 255, 0.16));
}

.file-details {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 14px;
  font-weight: 780;
  color: #14233e;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 12px;
  color: #7d8da6;
  margin-top: 4px;
}

.task-actions {
  margin-left: 12px;
  flex-shrink: 0;
}

.task-progress {
  margin-bottom: 8px;
}

.task-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}

.status-message {
  color: #536987;
}

.status-message .success {
  color: #1f9d55;
  font-weight: 760;
}

.status-message .error {
  color: #f56c6c;
}

.status-message .warning {
  color: #e6a23c;
}

.upload-info {
  display: flex;
  gap: 12px;
  color: #7d8da6;
}

.upload-speed {
  font-weight: 500;
}
</style>

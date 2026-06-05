<template>
  <el-dialog
    v-model="uploadStore.dialogVisible"
    title="上传管理"
    width="600px"
    :close-on-click-modal="false"
    class="upload-dialog"
  >
    <div class="upload-container">
      <div
        class="upload-area"
        :class="{ 'is-dragover': isDragover }"
        @drop.prevent="handleDrop"
        @dragover.prevent="isDragover = true"
        @dragleave.prevent="isDragover = false"
      >
        <el-icon class="upload-icon"><Upload /></el-icon>
        <div class="upload-text">
          <p>拖拽文件到此处上传</p>
          <p class="upload-hint">或</p>
        </div>
        <el-button type="primary" @click="triggerFileSelect">选择文件</el-button>
        <input ref="fileInputRef" type="file" multiple style="display: none" @change="handleFileSelect" />
      </div>

      <div v-if="uploadStore.activeTasks.length > 0" class="task-section">
        <div class="section-header">
          <span class="section-title">上传中 ({{ uploadStore.activeTasks.length }})</span>
        </div>
        <div class="task-list">
          <UploadTaskItem
            v-for="task in uploadStore.activeTasks"
            :key="task.id"
            :task="task"
            @cancel="handleCancel"
            @retry="handleRetry"
          />
        </div>
      </div>

      <div v-if="uploadStore.completedTasks.length > 0" class="task-section">
        <div class="section-header">
          <span class="section-title">已完成 ({{ uploadStore.completedTasks.length }})</span>
          <el-button type="danger" size="small" text @click="handleClearCompleted">清除已完成</el-button>
        </div>
        <div class="task-list">
          <UploadTaskItem
            v-for="task in uploadStore.completedTasks"
            :key="task.id"
            :task="task"
            @cancel="handleCancel"
            @retry="handleRetry"
          />
        </div>
      </div>

      <div v-if="uploadStore.tasks.length === 0" class="empty-state">
        <el-icon class="empty-icon"><FolderOpened /></el-icon>
        <p>暂无上传任务</p>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { FolderOpened, Upload } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import UploadTaskItem from '../UploadTaskItem/index.vue';
import { useUploadStore } from '@/stores/upload';
import { useFileStore } from '@/stores/file';

interface Props {
  folderId?: number;
}

const props = defineProps<Props>();

const uploadStore = useUploadStore();
const fileStore = useFileStore();

const fileInputRef = ref<HTMLInputElement>();
const isDragover = ref(false);

function triggerFileSelect(): void {
  fileInputRef.value?.click();
}

async function handleFileSelect(event: Event): Promise<void> {
  const target = event.target as HTMLInputElement;
  const files = target.files;

  if (!files || files.length === 0) {
    return;
  }

  await addFiles(Array.from(files));
  target.value = '';
}

async function handleDrop(event: DragEvent): Promise<void> {
  isDragover.value = false;

  const files = event.dataTransfer?.files;
  if (!files || files.length === 0) {
    return;
  }

  await addFiles(Array.from(files));
}

async function addFiles(files: File[]): Promise<void> {
  if (files.length === 0) {
    return;
  }

  try {
    const targetFolderId = props.folderId ?? fileStore.currentFolderId ?? undefined;

    for (const file of files) {
      await uploadStore.addTask(file, targetFolderId);
    }

    ElMessage.success(`已添加 ${files.length} 个文件到上传队列`);
  } catch (error: any) {
    console.error('Failed to add files:', error);
    ElMessage.error(error.message || '添加文件失败');
  }
}

async function handleCancel(taskId: string): Promise<void> {
  try {
    await uploadStore.cancelUpload(taskId);
    ElMessage.info('已取消上传');
  } catch (error: any) {
    console.error('Failed to cancel upload:', error);
    ElMessage.error(error.message || '取消上传失败');
  }
}

async function handleRetry(taskId: string): Promise<void> {
  try {
    await uploadStore.retryUpload(taskId);
    ElMessage.info('正在重试上传');
  } catch (error: any) {
    console.error('Failed to retry upload:', error);
    ElMessage.error(error.message || '重试上传失败');
  }
}

function handleClearCompleted(): void {
  uploadStore.clearCompleted();
  ElMessage.success('已清除完成任务');
}
</script>

<style scoped>
.upload-dialog {
  max-height: 80vh;
}

:deep(.upload-dialog .el-dialog) {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 28px;
  background:
    radial-gradient(circle at 100% 0%, rgba(100, 215, 255, 0.18), transparent 28%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(247, 252, 255, 0.84));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.94),
    0 32px 80px rgba(73, 112, 160, 0.2);
  backdrop-filter: blur(22px);
}

:deep(.upload-dialog .el-dialog__header) {
  margin: 0;
  padding: 22px 24px 14px;
}

:deep(.upload-dialog .el-dialog__title) {
  color: #10213f;
  font-weight: 820;
}

:deep(.upload-dialog .el-dialog__body) {
  padding: 12px 24px 24px;
}

.upload-container {
  max-height: 60vh;
  overflow-y: auto;
}

.upload-area {
  border: 1px dashed rgba(96, 165, 250, 0.58);
  border-radius: 24px;
  padding: 40px 20px;
  text-align: center;
  transition: all 0.3s;
  cursor: pointer;
  background:
    radial-gradient(circle at 50% 0%, rgba(116, 220, 255, 0.16), transparent 42%),
    rgba(255, 255, 255, 0.58);
  margin-bottom: 20px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.86);
}

.upload-area:hover {
  border-color: rgba(45, 112, 255, 0.74);
  background:
    radial-gradient(circle at 50% 0%, rgba(116, 220, 255, 0.24), transparent 42%),
    rgba(245, 251, 255, 0.82);
}

.upload-area.is-dragover {
  border-color: rgba(45, 112, 255, 0.9);
  background: rgba(231, 245, 255, 0.86);
  transform: scale(1.02);
}

.upload-icon {
  font-size: 48px;
  color: #2d70ff;
  margin-bottom: 16px;
  filter: drop-shadow(0 12px 18px rgba(45, 112, 255, 0.18));
}

.upload-text {
  margin-bottom: 16px;
  color: #536987;
}

.upload-text p {
  margin: 4px 0;
}

.upload-hint {
  font-size: 12px;
  color: #8da0b8;
}

.task-section {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid rgba(255, 255, 255, 0.74);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.58);
  margin-bottom: 8px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.84);
}

.section-title {
  font-size: 14px;
  font-weight: 820;
  color: #14233e;
}

.task-list {
  overflow: hidden;
  border: 1px solid rgba(219, 234, 249, 0.76);
  border-radius: 20px;
  background-color: rgba(255, 255, 255, 0.58);
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #7d8da6;
}

.empty-icon {
  font-size: 64px;
  color: #cbd5e1;
  margin-bottom: 16px;
}

.empty-state p {
  font-size: 14px;
  margin: 0;
}
</style>

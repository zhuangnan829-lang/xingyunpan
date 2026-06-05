import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import type { UploadTask } from '@/types/upload';
import { UploadManager } from '@/utils/upload';

function generateTaskId(): string {
  return `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}

export const useUploadStore = defineStore('upload', () => {
  const tasks = ref<UploadTask[]>([]);
  const dialogVisible = ref(false);
  const maxConcurrent = ref(3);

  const activeTasks = computed(() =>
    tasks.value.filter((task) => task.status === 'pending' || task.status === 'hashing' || task.status === 'uploading'),
  );

  const completedTasks = computed(() =>
    tasks.value.filter((task) => task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled'),
  );

  const uploadingCount = computed(() => tasks.value.filter((task) => task.status === 'uploading').length);

  async function addTask(file: File, folderId?: number): Promise<string> {
    const taskId = generateTaskId();

    const task: UploadTask = {
      id: taskId,
      file,
      hash: '',
      status: 'pending',
      progress: 0,
      speed: 0,
      isMultipart: false,
      folderId,
    };

    tasks.value.push(task);
    dialogVisible.value = true;

    if (uploadingCount.value < maxConcurrent.value) {
      void startUpload(taskId);
    }

    return taskId;
  }

  async function startUpload(taskId: string): Promise<void> {
    const task = tasks.value.find((item) => item.id === taskId);
    if (!task) {
      throw new Error('Task not found');
    }

    if (task.status !== 'pending' && task.status !== 'failed') {
      return;
    }

    try {
      const manager = new UploadManager(
        task,
        (updatedTask) => {
          updateTask(taskId, updatedTask);
        },
        (updatedTask) => {
          updateTask(taskId, updatedTask);

          if (updatedTask.status === 'completed' || updatedTask.status === 'failed' || updatedTask.status === 'cancelled') {
            startNextPendingTask();
          }
        },
      );

      await manager.start();
    } catch (error: any) {
      console.error('Upload failed:', error);
      task.error = error.message || '上传失败';
      task.status = 'failed';
    }
  }

  async function cancelUpload(taskId: string): Promise<void> {
    const task = tasks.value.find((item) => item.id === taskId);
    if (!task) {
      throw new Error('Task not found');
    }

    if (task.status !== 'hashing' && task.status !== 'uploading' && task.status !== 'pending') {
      return;
    }

    const manager = new UploadManager(task);
    await manager.cancel();

    task.status = 'cancelled';
    task.error = '用户取消上传';
    startNextPendingTask();
  }

  async function retryUpload(taskId: string): Promise<void> {
    const task = tasks.value.find((item) => item.id === taskId);
    if (!task) {
      throw new Error('Task not found');
    }

    if (task.status !== 'failed') {
      return;
    }

    task.status = 'pending';
    task.progress = 0;
    task.speed = 0;
    task.error = undefined;
    task.startTime = undefined;
    task.completedChunks = [];
    task.result = undefined;

    await startUpload(taskId);
  }

  function clearCompleted(): void {
    tasks.value = tasks.value.filter(
      (task) => task.status !== 'completed' && task.status !== 'failed' && task.status !== 'cancelled',
    );
  }

  function removeTask(taskId: string): void {
    const index = tasks.value.findIndex((task) => task.id === taskId);
    if (index > -1) {
      tasks.value.splice(index, 1);
    }
  }

  function updateTask(taskId: string, updates: Partial<UploadTask>): void {
    const task = tasks.value.find((item) => item.id === taskId);
    if (task) {
      Object.assign(task, updates);
    }
  }

  function startNextPendingTask(): void {
    if (uploadingCount.value >= maxConcurrent.value) {
      return;
    }

    const pendingTask = tasks.value.find((task) => task.status === 'pending');
    if (pendingTask) {
      void startUpload(pendingTask.id);
    }
  }

  function showDialog(): void {
    dialogVisible.value = true;
  }

  function hideDialog(): void {
    dialogVisible.value = false;
  }

  function getTask(taskId: string): UploadTask | undefined {
    return tasks.value.find((task) => task.id === taskId);
  }

  function getStatusMessage(task: UploadTask): string {
    switch (task.status) {
      case 'pending':
        return '等待上传';
      case 'hashing':
        return '计算文件哈希中...';
      case 'uploading':
        if (task.isMultipart && task.totalChunks && task.completedChunks) {
          return `上传中 ${task.completedChunks.length}/${task.totalChunks} 分片`;
        }
        return '上传中...';
      case 'completed':
        return '上传完成';
      case 'failed':
        return task.error || '上传失败';
      case 'cancelled':
        return '已取消';
      default:
        return '';
    }
  }

  function formatSpeed(speed: number): string {
    if (speed === 0) return '0 B/s';

    const units = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
    let unitIndex = 0;
    let value = speed;

    while (value >= 1024 && unitIndex < units.length - 1) {
      value /= 1024;
      unitIndex += 1;
    }

    return `${value.toFixed(2)} ${units[unitIndex]}`;
  }

  function getRemainingTime(task: UploadTask): string {
    if (task.status !== 'uploading' || task.speed === 0) {
      return '--';
    }

    const remainingBytes = task.file.size * (1 - task.progress / 100);
    const remainingSeconds = Math.ceil(remainingBytes / task.speed);

    if (remainingSeconds < 60) {
      return `${remainingSeconds}秒`;
    }

    if (remainingSeconds < 3600) {
      const minutes = Math.floor(remainingSeconds / 60);
      return `${minutes}分钟`;
    }

    const hours = Math.floor(remainingSeconds / 3600);
    const minutes = Math.floor((remainingSeconds % 3600) / 60);
    return `${hours}小时${minutes}分钟`;
  }

  return {
    tasks,
    dialogVisible,
    maxConcurrent,
    activeTasks,
    completedTasks,
    uploadingCount,
    addTask,
    startUpload,
    cancelUpload,
    retryUpload,
    clearCompleted,
    removeTask,
    updateTask,
    showDialog,
    hideDialog,
    getTask,
    getStatusMessage,
    formatSpeed,
    getRemainingTime,
  };
});

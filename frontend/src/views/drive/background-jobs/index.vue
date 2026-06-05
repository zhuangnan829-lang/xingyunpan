<template>
  <section class="background-jobs-page" @click="clearSelection">
    <div class="aurora aurora-left" />
    <div class="aurora aurora-right" />

    <div class="jobs-shell glass-panel">
      <header class="jobs-header">
        <div>
          <p class="eyebrow">DRIVE CONSOLE</p>
          <div class="title-row">
            <h1>后台任务</h1>
            <button class="icon-button" type="button" title="刷新" @click.stop="refresh">
              <el-icon :class="{ spinning: refreshing }"><RefreshRight /></el-icon>
            </button>
          </div>
          <p class="subtitle">集中查看上传、转码和清理任务，单击选择，双击查看任务详情。</p>
        </div>

        <label class="auto-refresh" @click.stop>
          <span>自动刷新</span>
          <input v-model="autoRefresh" type="checkbox" />
          <i />
        </label>
      </header>

      <div class="summary-grid">
        <article v-for="item in summary" :key="item.key" class="summary-card" :class="`tone-${item.tone}`">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </article>
      </div>

      <section class="toolbar glass-panel compact">
        <div class="filter-box">
          <el-icon><Search /></el-icon>
          <input v-model="keyword" type="search" placeholder="筛选任务名称、状态或文件类型" @click.stop />
        </div>

        <div class="toolbar-actions" @click.stop>
          <select v-model="statusFilter">
            <option value="">全部状态</option>
            <option value="pending">等待中</option>
            <option value="hashing">校验中</option>
            <option value="uploading">上传中</option>
            <option value="completed">已完成</option>
            <option value="failed">失败</option>
            <option value="cancelled">已取消</option>
          </select>
          <button type="button" @click="retrySelected" :disabled="!selectedFailedTasks.length">重试</button>
          <button type="button" @click="clearCompleted" :disabled="!completedCount">清理完成</button>
        </div>
      </section>

      <section class="task-board glass-panel">
        <template v-if="filteredTasks.length">
          <article
            v-for="task in filteredTasks"
            :key="task.id"
            class="task-card"
            :class="[`status-${task.status}`, { selected: selectedIds.includes(task.id) }]"
            @click.stop="toggleSelect(task.id)"
            @dblclick.stop="openDetail(task)"
          >
            <div class="task-icon">
              <el-icon><UploadFilled /></el-icon>
            </div>
            <div class="task-main">
              <div class="task-title">
                <strong>{{ task.file.name }}</strong>
                <span>{{ statusText(task.status) }}</span>
              </div>
              <p>{{ formatSize(task.file.size) }} · {{ task.file.type || '未知类型' }}</p>
              <div class="progress-track">
                <i :style="{ width: `${Math.max(0, Math.min(100, task.progress || 0))}%` }" />
              </div>
            </div>
            <div class="task-meta">
              <strong>{{ Math.round(task.progress || 0) }}%</strong>
              <span>{{ uploadStore.formatSpeed(task.speed || 0) }}</span>
            </div>
          </article>
        </template>

        <div v-else class="empty-state">
          <el-icon><Box /></el-icon>
          <strong>没有记录</strong>
          <span>上传文件后，任务会在这里以卡片方式呈现。</span>
        </div>
      </section>
    </div>

    <el-drawer v-model="detailVisible" size="420px" class="job-detail-drawer" append-to-body>
      <template #header>
        <div class="drawer-head">
          <p>任务详情</p>
          <h2>{{ selectedTask?.file.name || '未选择任务' }}</h2>
        </div>
      </template>

      <div v-if="selectedTask" class="detail-stack">
        <div class="detail-row">
          <span>状态</span>
          <strong>{{ statusText(selectedTask.status) }}</strong>
        </div>
        <div class="detail-row">
          <span>大小</span>
          <strong>{{ formatSize(selectedTask.file.size) }}</strong>
        </div>
        <div class="detail-row">
          <span>速度</span>
          <strong>{{ uploadStore.formatSpeed(selectedTask.speed || 0) }}</strong>
        </div>
        <div class="detail-row">
          <span>剩余时间</span>
          <strong>{{ uploadStore.getRemainingTime(selectedTask) }}</strong>
        </div>
        <div class="detail-row">
          <span>分片</span>
          <strong>{{ selectedTask.completedChunks?.length || 0 }} / {{ selectedTask.totalChunks || 0 }}</strong>
        </div>
        <div class="detail-row">
          <span>任务 ID</span>
          <strong>{{ selectedTask.id }}</strong>
        </div>
        <div v-if="selectedTask.error" class="detail-error">{{ selectedTask.error }}</div>
      </div>
    </el-drawer>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { ElDrawer, ElIcon, ElMessage } from 'element-plus';
import { Box, RefreshRight, Search, UploadFilled } from '@element-plus/icons-vue';
import { useUploadStore } from '@/stores/upload';
import type { UploadStatus, UploadTask } from '@/types/upload';

type Tone = 'sky' | 'mint' | 'amber' | 'rose';

const uploadStore = useUploadStore();
const keyword = ref('');
const statusFilter = ref<UploadStatus | ''>('');
const selectedIds = ref<string[]>([]);
const selectedTask = ref<UploadTask | null>(null);
const detailVisible = ref(false);
const refreshing = ref(false);
const autoRefresh = ref(true);
let refreshTimer: ReturnType<typeof window.setInterval> | undefined;

const filteredTasks = computed(() => {
  const query = keyword.value.trim().toLowerCase();
  return uploadStore.tasks.filter((task) => {
    const matchesStatus = !statusFilter.value || task.status === statusFilter.value;
    const haystack = `${task.file.name} ${task.file.type} ${task.status}`.toLowerCase();
    return matchesStatus && (!query || haystack.includes(query));
  });
});

const selectedFailedTasks = computed(() =>
  uploadStore.tasks.filter((task) => selectedIds.value.includes(task.id) && task.status === 'failed'),
);

const completedCount = computed(() =>
  uploadStore.tasks.filter((task) => ['completed', 'failed', 'cancelled'].includes(task.status)).length,
);

const summary = computed<Array<{ key: string; label: string; value: number; tone: Tone }>>(() => [
  { key: 'all', label: '任务总数', value: uploadStore.tasks.length, tone: 'sky' },
  { key: 'active', label: '执行中', value: uploadStore.activeTasks.length, tone: 'mint' },
  { key: 'done', label: '已完成', value: uploadStore.completedTasks.length, tone: 'amber' },
  { key: 'failed', label: '失败', value: uploadStore.tasks.filter((task) => task.status === 'failed').length, tone: 'rose' },
]);

watch(autoRefresh, (enabled) => {
  if (enabled) startAutoRefresh();
  else stopAutoRefresh();
}, { immediate: true });

function toggleSelect(taskId: string) {
  selectedIds.value = selectedIds.value.includes(taskId)
    ? selectedIds.value.filter((id) => id !== taskId)
    : [...selectedIds.value, taskId];
}

function clearSelection() {
  selectedIds.value = [];
}

function openDetail(task: UploadTask) {
  selectedTask.value = task;
  detailVisible.value = true;
  if (!selectedIds.value.includes(task.id)) {
    selectedIds.value = [task.id];
  }
}

function refresh() {
  refreshing.value = true;
  window.setTimeout(() => {
    refreshing.value = false;
    ElMessage.success('后台任务已刷新');
  }, 360);
}

async function retrySelected() {
  for (const task of selectedFailedTasks.value) {
    await uploadStore.retryUpload(task.id);
  }
  ElMessage.success('已重新加入上传队列');
}

function clearCompleted() {
  uploadStore.clearCompleted();
  selectedIds.value = [];
  ElMessage.success('已清理完成任务');
}

function startAutoRefresh() {
  stopAutoRefresh();
  refreshTimer = window.setInterval(() => {
    refreshing.value = true;
    window.setTimeout(() => {
      refreshing.value = false;
    }, 220);
  }, 5000);
}

function stopAutoRefresh() {
  if (refreshTimer) {
    window.clearInterval(refreshTimer);
    refreshTimer = undefined;
  }
}

function statusText(status: UploadStatus): string {
  const map: Record<UploadStatus, string> = {
    pending: '等待中',
    hashing: '校验中',
    uploading: '上传中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
  };
  return map[status];
}

function formatSize(size: number): string {
  if (!size) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = size;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${value.toFixed(index === 0 ? 0 : 2)} ${units[index]}`;
}

onBeforeUnmount(stopAutoRefresh);
</script>

<style scoped>
.background-jobs-page {
  position: relative;
  min-height: 100%;
  padding: 24px;
  overflow: hidden;
  color: #10213b;
  background:
    radial-gradient(circle at 7% 0%, rgba(186, 230, 253, 0.48), transparent 28%),
    radial-gradient(circle at 92% 4%, rgba(252, 231, 243, 0.58), transparent 34%),
    linear-gradient(135deg, #eef8ff 0%, #f8fbff 52%, #fff7fb 100%);
}

.aurora {
  position: absolute;
  pointer-events: none;
  filter: blur(20px);
  opacity: 0.58;
}

.aurora-left {
  left: -80px;
  top: 120px;
  width: 260px;
  height: 260px;
  border-radius: 50%;
  background: rgba(125, 211, 252, 0.42);
}

.aurora-right {
  right: -90px;
  top: 40px;
  width: 320px;
  height: 240px;
  border-radius: 50%;
  background: rgba(251, 207, 232, 0.45);
}

.glass-panel {
  border: 1px solid rgba(255, 255, 255, 0.76);
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.72), rgba(240, 249, 255, 0.48)),
    rgba(255, 255, 255, 0.48);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.94), 0 18px 46px rgba(96, 165, 250, 0.12);
  backdrop-filter: blur(20px);
}

.jobs-shell {
  position: relative;
  z-index: 1;
  min-height: calc(100vh - 132px);
  padding: 34px;
  border-radius: 28px;
}

.jobs-header,
.title-row,
.toolbar,
.toolbar-actions,
.task-card,
.task-title,
.task-meta,
.auto-refresh {
  display: flex;
  align-items: center;
}

.jobs-header {
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 24px;
}

.eyebrow {
  margin: 0 0 12px;
  color: #2f7df6;
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 0.16em;
}

.title-row {
  gap: 14px;
}

h1 {
  margin: 0;
  color: #10213b;
  font-size: clamp(36px, 4vw, 54px);
  line-height: 1;
  letter-spacing: 0;
}

.subtitle {
  margin: 14px 0 0;
  color: #62748e;
  font-size: 16px;
}

.icon-button,
.toolbar-actions button {
  border: 0;
  cursor: pointer;
}

.icon-button {
  display: grid;
  width: 46px;
  height: 46px;
  place-items: center;
  border-radius: 16px;
  color: #31516e;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 12px 26px rgba(30, 64, 175, 0.12);
}

.spinning {
  animation: spin 0.8s linear infinite;
}

.auto-refresh {
  position: relative;
  gap: 10px;
  color: #64748b;
  font-weight: 700;
  cursor: pointer;
}

.auto-refresh input {
  position: absolute;
  opacity: 0;
}

.auto-refresh i {
  width: 44px;
  height: 24px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.32);
  box-shadow: inset 0 1px 2px rgba(15, 23, 42, 0.12);
}

.auto-refresh i::after {
  content: '';
  display: block;
  width: 18px;
  height: 18px;
  margin: 3px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 4px 10px rgba(15, 23, 42, 0.18);
  transition: transform 0.2s ease;
}

.auto-refresh input:checked + i {
  background: linear-gradient(135deg, #5ab5ff, #24c7b7);
}

.auto-refresh input:checked + i::after {
  transform: translateX(20px);
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  margin-bottom: 18px;
}

.summary-card {
  min-height: 92px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.55);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9), 0 14px 28px rgba(59, 130, 246, 0.08);
}

.summary-card span {
  color: #64748b;
  font-size: 13px;
  font-weight: 800;
}

.summary-card strong {
  display: block;
  margin-top: 10px;
  color: #10213b;
  font-size: 28px;
}

.tone-sky { background: linear-gradient(135deg, rgba(219, 234, 254, 0.72), rgba(255, 255, 255, 0.5)); }
.tone-mint { background: linear-gradient(135deg, rgba(204, 251, 241, 0.68), rgba(255, 255, 255, 0.5)); }
.tone-amber { background: linear-gradient(135deg, rgba(254, 243, 199, 0.62), rgba(255, 255, 255, 0.5)); }
.tone-rose { background: linear-gradient(135deg, rgba(252, 231, 243, 0.68), rgba(255, 255, 255, 0.5)); }

.toolbar {
  justify-content: space-between;
  gap: 18px;
  min-height: 74px;
  margin-bottom: 18px;
  padding: 12px 16px;
  border-radius: 22px;
}

.filter-box {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 12px;
  min-width: 240px;
  color: #2f7df6;
}

.filter-box input {
  width: 100%;
  border: 0;
  outline: none;
  color: #10213b;
  background: transparent;
  font-size: 16px;
  font-weight: 700;
}

.toolbar-actions {
  gap: 10px;
}

.toolbar-actions select,
.toolbar-actions button {
  height: 42px;
  border: 0;
  border-radius: 14px;
  color: #10213b;
  background: rgba(255, 255, 255, 0.78);
  font-weight: 800;
}

.toolbar-actions select {
  min-width: 130px;
  padding: 0 12px;
}

.toolbar-actions button {
  padding: 0 16px;
}

.toolbar-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.task-board {
  min-height: 420px;
  padding: 18px;
  border-radius: 24px;
}

.task-card {
  gap: 16px;
  min-height: 86px;
  margin-bottom: 12px;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.58);
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.task-card:hover,
.task-card.selected {
  border-color: rgba(47, 125, 246, 0.58);
  box-shadow: 0 18px 34px rgba(47, 125, 246, 0.14);
  transform: translateY(-1px);
}

.task-icon {
  display: grid;
  flex: 0 0 54px;
  width: 54px;
  height: 54px;
  place-items: center;
  border-radius: 18px;
  color: #2f7df6;
  background: rgba(255, 255, 255, 0.78);
}

.task-main {
  flex: 1;
  min-width: 0;
}

.task-title {
  justify-content: space-between;
  gap: 12px;
}

.task-title strong {
  overflow: hidden;
  color: #10213b;
  font-size: 16px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-title span {
  flex: 0 0 auto;
  padding: 4px 10px;
  border-radius: 999px;
  color: #2563eb;
  background: rgba(219, 234, 254, 0.8);
  font-size: 12px;
  font-weight: 900;
}

.task-main p,
.task-meta span {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 13px;
}

.progress-track {
  overflow: hidden;
  height: 7px;
  margin-top: 12px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.18);
}

.progress-track i {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #2f7df6, #24c7b7);
}

.task-meta {
  flex-direction: column;
  align-items: flex-end;
  min-width: 90px;
}

.task-meta strong {
  color: #10213b;
  font-size: 20px;
}

.empty-state {
  display: grid;
  min-height: 360px;
  place-items: center;
  align-content: center;
  color: #94a3b8;
  text-align: center;
}

.empty-state .el-icon {
  margin-bottom: 12px;
  font-size: 68px;
}

.empty-state strong {
  color: #64748b;
  font-size: 28px;
}

.empty-state span {
  margin-top: 8px;
}

:deep(.job-detail-drawer .el-drawer) {
  border-left: 1px solid rgba(255, 255, 255, 0.72);
  background:
    radial-gradient(circle at 0% 0%, rgba(186, 230, 253, 0.42), transparent 34%),
    rgba(255, 255, 255, 0.86);
  backdrop-filter: blur(24px);
}

.drawer-head p {
  margin: 0 0 6px;
  color: #2f7df6;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.12em;
}

.drawer-head h2 {
  overflow: hidden;
  margin: 0;
  color: #10213b;
  font-size: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-stack {
  display: grid;
  gap: 12px;
}

.detail-row,
.detail-error {
  padding: 14px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.detail-row span {
  color: #64748b;
}

.detail-row strong {
  overflow-wrap: anywhere;
  color: #10213b;
  text-align: right;
}

.detail-error {
  color: #b91c1c;
  background: rgba(254, 226, 226, 0.7);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 1100px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .toolbar {
    align-items: stretch;
    flex-direction: column;
  }
}

@media (max-width: 720px) {
  .background-jobs-page {
    padding: 14px;
  }

  .jobs-shell {
    padding: 20px;
    border-radius: 22px;
  }

  .jobs-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .summary-grid {
    grid-template-columns: 1fr;
  }

  .task-card {
    align-items: flex-start;
  }
}
</style>

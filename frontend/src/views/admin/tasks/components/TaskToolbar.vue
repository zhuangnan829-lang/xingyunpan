<template>
  <section class="task-toolbar">
    <div class="toolbar-actions">
      <button class="action-button primary" type="button" :disabled="loading" @click="$emit('refresh')">
        <Refresh class="button-icon" :class="{ spinning: loading }" />
        <span>刷新</span>
      </button>
      <button class="action-button" type="button" :disabled="loading" @click="$emit('apply')">
        <Filter class="button-icon" />
        <span>过滤</span>
      </button>
      <button class="action-button" type="button" :disabled="loading" @click="$emit('reset')">
        <Delete class="button-icon" />
        <span>清理</span>
      </button>
    </div>

    <div class="toolbar-filters">
      <label>
        <span>队列</span>
        <select :value="queueKey" @change="updateQueue">
          <option value="">全部队列</option>
          <option v-for="queue in queueOptions" :key="queue.key" :value="queue.key">{{ queue.label }}</option>
        </select>
      </label>
      <label>
        <span>状态</span>
        <select :value="status" @change="updateStatus">
          <option v-for="item in statusOptions" :key="item.key" :value="item.key">{{ item.label }}</option>
        </select>
      </label>
      <label>
        <span>节点</span>
        <select :value="nodeId" @change="updateNode">
          <option value="">全部节点</option>
          <option v-for="node in nodes" :key="node.id" :value="node.id">{{ node.name }}</option>
        </select>
      </label>
    </div>

    <div class="toolbar-meta">
      <strong>{{ activeFilterText }}</strong>
      <span>{{ lastUpdatedAt ? `更新于 ${lastUpdatedAt}` : '等待首次刷新' }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { Delete, Filter, Refresh } from '@element-plus/icons-vue';
import type { NodePayload } from '@/api/nodes';
import { queueOptions, statusOptions, type QueueKey, type TaskStatus } from '../types';

defineProps<{
  queueKey: QueueKey | '';
  status: TaskStatus;
  nodeId: number | '';
  nodes: NodePayload[];
  loading: boolean;
  activeFilterText: string;
  lastUpdatedAt: string;
}>();

const emit = defineEmits<{
  (event: 'update:queueKey', value: QueueKey | ''): void;
  (event: 'update:status', value: TaskStatus): void;
  (event: 'update:nodeId', value: number | ''): void;
  (event: 'refresh'): void;
  (event: 'apply'): void;
  (event: 'reset'): void;
}>();

function updateQueue(event: Event) {
  emit('update:queueKey', (event.target as HTMLSelectElement).value as QueueKey | '');
}

function updateStatus(event: Event) {
  emit('update:status', (event.target as HTMLSelectElement).value as TaskStatus);
}

function updateNode(event: Event) {
  const value = (event.target as HTMLSelectElement).value;
  emit('update:nodeId', value ? Number(value) : '');
}
</script>

<style scoped>
.task-toolbar {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: auto minmax(320px, 1fr) auto;
  align-items: center;
  gap: 16px;
  padding: 14px;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 22px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.74), rgba(244, 250, 255, 0.6));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(16px);
}

.toolbar-actions,
.toolbar-filters,
.toolbar-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.action-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 42px;
  padding: 0 16px;
  border: 1px solid rgba(211, 224, 236, 0.78);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.7);
  color: #42556a;
  font-weight: 750;
  cursor: pointer;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.action-button.primary {
  border-color: transparent;
  background: linear-gradient(135deg, #5ac8fa, #35b8df 52%, #f2a6ba);
  color: #ffffff;
  box-shadow: 0 14px 26px rgba(79, 184, 235, 0.22);
}

.action-button:hover {
  transform: translateY(-1px);
}

.action-button:disabled {
  cursor: wait;
  opacity: 0.72;
  transform: none;
}

.button-icon {
  width: 17px;
  height: 17px;
}

.spinning {
  animation: spin 0.8s linear infinite;
}

.toolbar-filters label {
  display: grid;
  gap: 6px;
  min-width: 150px;
}

.toolbar-filters span,
.toolbar-meta span {
  color: #738497;
  font-size: 12px;
  font-weight: 700;
}

.toolbar-filters select {
  height: 42px;
  padding: 0 34px 0 12px;
  border: 1px solid rgba(205, 219, 232, 0.82);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.78);
  color: #25374a;
  font-weight: 720;
}

.toolbar-meta {
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.toolbar-meta strong {
  color: #25445f;
  font-size: 14px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 1180px) {
  .task-toolbar {
    grid-template-columns: 1fr;
    align-items: stretch;
  }

  .toolbar-actions,
  .toolbar-filters {
    flex-wrap: wrap;
  }

  .toolbar-meta {
    align-items: flex-start;
  }
}
</style>

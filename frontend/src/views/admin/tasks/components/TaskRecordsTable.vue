<template>
  <section class="records-panel">
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th><input type="checkbox" aria-label="閫夋嫨鍏ㄩ儴浠诲姟" :checked="allSelected" @change="toggleAll" /></th>
            <th>#</th>
            <th>任务类型</th>
            <th>使用节点</th>
            <th>节点能力</th>
            <th>执行方式</th>
            <th>当前状态</th>
            <th>错误原因</th>
            <th>创建时间</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="10" class="empty-cell">正在读取后台任务...</td>
          </tr>
          <tr v-else-if="!jobs.length">
            <td colspan="10" class="empty-cell">暂无后台任务</td>
          </tr>
          <tr v-for="job in jobs" v-else :key="job.id">
            <td>
              <input
                type="checkbox"
                :aria-label="`閫夋嫨浠诲姟 ${job.id}`"
                :checked="selectedIds.includes(job.id)"
                @change="toggleJob(job.id)"
              />
            </td>
            <td class="id-cell">{{ job.id }}</td>
            <td>
              <button class="content-button" type="button" @click="$emit('detail', job)">
                <strong>{{ jobTypeLabel(job) }}</strong>
                <span>{{ contentSummary(job) }}</span>
              </button>
            </td>
            <td>
              <button class="node-button" type="button" @click="$emit('detail', job)">
                <strong>{{ nodeName(job) }}</strong>
                <span>{{ nodeMeta(job) }}</span>
              </button>
            </td>
            <td>
              <span class="capability-chip">{{ capabilityLabel(job) }}</span>
            </td>
            <td>
              <span class="execution-chip" :class="`mode-${job.execution_mode || 'unified_runner'}`">{{ executionModeLabel(job) }}</span>
            </td>
            <td>
              <span class="status-chip" :class="`is-${job.status}`">
                <i></i>{{ statusLabelMap[job.status] || job.status }}
              </span>
            </td>
            <td class="error-cell" :title="job.last_error || ''">{{ errorSummary(job) }}</td>
            <td>{{ formatDate(job.created_at) }}</td>
            <td>
              <button class="icon-action" type="button" title="鏌ョ湅璇︽儏" @click="$emit('detail', job)">
                <View class="row-icon" />
              </button>
              <button class="icon-action danger" type="button" title="鍒犻櫎浠诲姟" @click="$emit('delete', job)">
                <Delete class="row-icon" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="table-footer">
      <div class="pager">
        <button type="button" :disabled="page <= 1 || loading" @click="$emit('page-change', page - 1)">
          <ArrowLeft class="pager-icon" />
        </button>
        <span>{{ page }} / {{ totalPages }}</span>
        <button type="button" :disabled="page >= totalPages || loading" @click="$emit('page-change', page + 1)">
          <ArrowRight class="pager-icon" />
        </button>
      </div>

      <select :value="pageSize" :disabled="loading" @change="changePageSize">
        <option :value="10">每页 10 条</option>
        <option :value="20">每页 20 条</option>
        <option :value="50">每页 50 条</option>
      </select>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { ArrowLeft, ArrowRight, Delete, View } from '@element-plus/icons-vue';
import { capabilityLabelMap, executionModeLabelMap, statusLabelMap, type TaskJob } from '../types';

const props = defineProps<{
  jobs: TaskJob[];
  loading: boolean;
  total: number;
  page: number;
  pageSize: number;
  selectedIds: number[];
}>();

const emit = defineEmits<{
  (event: 'update:selectedIds', ids: number[]): void;
  (event: 'detail', job: TaskJob): void;
  (event: 'delete', job: TaskJob): void;
  (event: 'page-change', page: number): void;
  (event: 'page-size-change', pageSize: number): void;
}>();

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)));
const allSelected = computed(() => props.jobs.length > 0 && props.jobs.every((job) => props.selectedIds.includes(job.id)));

function changePageSize(event: Event) {
  emit('page-size-change', Number((event.target as HTMLSelectElement).value));
}

function toggleAll(event: Event) {
  const checked = (event.target as HTMLInputElement).checked;
  emit('update:selectedIds', checked ? props.jobs.map((job) => job.id) : []);
}

function toggleJob(jobId: number) {
  const ids = props.selectedIds.includes(jobId)
    ? props.selectedIds.filter((id) => id !== jobId)
    : [...props.selectedIds, jobId];
  emit('update:selectedIds', ids);
}

function jobTypeLabel(job: TaskJob): string {
  const map: Record<string, string> = {
    'archive.create': '创建压缩文件',
    'archive.extract': '解压缩',
    'offline.download': '离线下载任务',
  };
  return map[job.job_type] || job.job_type || '闃熷垪浠诲姟';
}

function contentSummary(job: TaskJob): string {
  const resource = [job.resource_type, job.resource_id].filter(Boolean).join(' #');
  return resource || '系统后台任务';
}

function nodeName(job: TaskJob): string {
  return job.dispatch_node?.name || '未分配节点';
}

function nodeMeta(job: TaskJob): string {
  if (!job.dispatch_node) return '节点选择失败或尚未调度';
  return `${job.dispatch_node.type || 'node'} #${job.dispatch_node.id}`;
}

function capabilityLabel(job: TaskJob): string {
  return capabilityLabelMap[job.node_capability] || job.node_capability || '-';
}

function executionModeLabel(job: TaskJob): string {
  const mode = job.execution_mode || 'unified_runner';
  return executionModeLabelMap[mode] || mode;
}

function errorSummary(job: TaskJob): string {
  return job.last_error?.trim() || '-';
}

function formatDate(value: string): string {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString('zh-CN', { hour12: false });
}
</script>

<style scoped>
.records-panel {
  position: relative;
  z-index: 1;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.86);
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(247, 251, 255, 0.64));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.96),
    0 20px 44px rgba(83, 130, 170, 0.12);
  backdrop-filter: blur(16px);
}

.table-wrap {
  overflow: auto;
}

table {
  width: 100%;
  min-width: 1120px;
  border-collapse: collapse;
}

th,
td {
  padding: 14px 16px;
  border-bottom: 1px solid rgba(213, 224, 235, 0.72);
  color: #26384c;
  text-align: left;
  white-space: nowrap;
}

th {
  color: #66788a;
  font-size: 13px;
  font-weight: 850;
  background: rgba(255, 255, 255, 0.36);
}

td {
  font-size: 14px;
  font-weight: 620;
}

tbody tr:hover {
  background: rgba(227, 245, 255, 0.36);
}

input[type='checkbox'] {
  width: 18px;
  height: 18px;
  accent-color: #39b9df;
}

.id-cell {
  font-weight: 850;
}

.content-button,
.node-button,
.icon-action,
.pager button {
  border: 0;
  background: transparent;
  cursor: pointer;
}

.content-button,
.node-button {
  display: grid;
  gap: 5px;
  max-width: 280px;
  padding: 0;
  color: inherit;
  text-align: left;
}

.content-button strong,
.node-button strong {
  overflow: hidden;
  font-size: 14px;
  text-overflow: ellipsis;
}

.content-button span,
.node-button span {
  overflow: hidden;
  color: #758698;
  font-size: 12px;
  text-overflow: ellipsis;
}

.capability-chip,
.execution-chip,
.status-chip {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 850;
}

.capability-chip {
  color: #236987;
  background: rgba(213, 242, 255, 0.72);
}

.execution-chip {
  color: #526a82;
  background: rgba(214, 226, 238, 0.58);
}

.execution-chip.mode-runtime_record {
  color: #8a5d13;
  background: rgba(254, 243, 199, 0.86);
}

.status-chip {
  gap: 7px;
  color: #67788a;
  background: rgba(238, 244, 249, 0.86);
}

.status-chip i {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: currentColor;
}

.status-chip.is-completed { color: #25864a; background: rgba(220, 252, 231, 0.76); }
.status-chip.is-processing { color: #147a9e; background: rgba(207, 246, 255, 0.76); }
.status-chip.is-pending { color: #a06009; background: rgba(254, 243, 199, 0.78); }
.status-chip.is-failed { color: #c24155; background: rgba(255, 228, 230, 0.82); }

.error-cell {
  max-width: 260px;
  overflow: hidden;
  color: #be123c;
  font-size: 13px;
  text-overflow: ellipsis;
}

.icon-action {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: 12px;
  color: #647789;
}

.icon-action:hover {
  background: rgba(219, 241, 252, 0.78);
  color: #208db5;
}

.row-icon,
.pager-icon {
  width: 17px;
  height: 17px;
}

.empty-cell {
  height: 180px;
  color: #7b8da0;
  text-align: center;
}

.table-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
}

.pager {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.pager button {
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: 12px;
  color: #3d556d;
  background: rgba(255, 255, 255, 0.7);
}

.pager button:disabled {
  cursor: not-allowed;
  opacity: 0.42;
}

.pager span {
  color: #4e6276;
  font-weight: 850;
}

.table-footer select {
  height: 40px;
  padding: 0 32px 0 12px;
  border: 1px solid rgba(205, 219, 232, 0.82);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.78);
  color: #25374a;
  font-weight: 720;
}

@media (max-width: 760px) {
  .table-footer {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>

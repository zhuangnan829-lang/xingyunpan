<template>
  <el-drawer
    :model-value="modelValue"
    title="任务详情"
    size="520px"
    class="task-detail-drawer"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <template v-if="job">
      <div class="detail-head">
        <span class="status-chip" :class="`is-${job.status}`">{{ statusLabelMap[job.status] || job.status }}</span>
        <strong>#{{ job.id }} {{ jobTypeLabel }}</strong>
        <p>{{ queueLabelMap[job.queue_key] || job.queue_key }} / {{ job.resource_type || 'resource' }}</p>
      </div>

      <dl class="detail-list">
        <div>
          <dt>使用节点</dt>
          <dd>{{ job.dispatch_node?.name || '未分配节点' }}</dd>
        </div>
        <div>
          <dt>节点编号</dt>
          <dd>{{ job.dispatch_node ? `#${job.dispatch_node.id} ${job.dispatch_node.type || ''}` : '-' }}</dd>
        </div>
        <div>
          <dt>节点能力</dt>
          <dd>{{ capabilityLabel }}</dd>
        </div>
        <div>
          <dt>执行方式</dt>
          <dd>{{ executionModeLabel }}</dd>
        </div>
        <div v-if="job.execution_note">
          <dt>执行说明</dt>
          <dd>{{ job.execution_note }}</dd>
        </div>
        <div>
          <dt>资源编号</dt>
          <dd>{{ job.resource_id || '-' }}</dd>
        </div>
        <div>
          <dt>尝试次数</dt>
          <dd>{{ job.attempts }} / {{ job.max_attempts }}</dd>
        </div>
        <div>
          <dt>计划执行</dt>
          <dd>{{ formatDate(job.scheduled_at) }}</dd>
        </div>
        <div>
          <dt>开始时间</dt>
          <dd>{{ formatDate(job.started_at) }}</dd>
        </div>
        <div>
          <dt>结束时间</dt>
          <dd>{{ formatDate(job.finished_at) }}</dd>
        </div>
        <div>
          <dt>创建时间</dt>
          <dd>{{ formatDate(job.created_at) }}</dd>
        </div>
      </dl>

      <section v-if="job.last_error" class="detail-block error">
        <h3>错误原因</h3>
        <pre>{{ job.last_error }}</pre>
      </section>

      <section class="detail-block">
        <h3>Payload</h3>
        <pre>{{ prettyText(job.payload) }}</pre>
      </section>

      <section v-if="job.result" class="detail-block">
        <h3>Result</h3>
        <pre>{{ prettyText(job.result) }}</pre>
      </section>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { capabilityLabelMap, executionModeLabelMap, queueLabelMap, statusLabelMap, type TaskJob } from '../types';

const props = defineProps<{
  modelValue: boolean;
  job: TaskJob | null;
}>();

defineEmits<{
  (event: 'update:modelValue', value: boolean): void;
}>();

const jobTypeLabel = computed(() => {
  const job = props.job;
  if (!job) return '闃熷垪浠诲姟';
  const map: Record<string, string> = {
    'archive.create': '创建压缩文件',
    'archive.extract': '解压缩',
    'offline.download': '离线下载任务',
  };
  return map[job.job_type] || job.job_type || '闃熷垪浠诲姟';
});

const capabilityLabel = computed(() => {
  const capability = props.job?.node_capability || '';
  return capabilityLabelMap[capability] || capability || '-';
});

const executionModeLabel = computed(() => {
  const mode = props.job?.execution_mode || 'unified_runner';
  return executionModeLabelMap[mode] || mode;
});

function formatDate(value?: string | null): string {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString('zh-CN', { hour12: false });
}

function prettyText(value: string): string {
  if (!value) return '-';
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
}
</script>

<style scoped>
.detail-head {
  display: grid;
  gap: 10px;
  padding: 18px;
  border: 1px solid rgba(218, 230, 240, 0.86);
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.84), rgba(239, 249, 255, 0.74));
}

.detail-head strong {
  color: #203246;
  font-size: 22px;
}

.detail-head p {
  margin: 0;
  color: #6f8294;
}

.status-chip {
  justify-self: start;
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  color: #67788a;
  background: rgba(238, 244, 249, 0.86);
  font-size: 12px;
  font-weight: 850;
  line-height: 28px;
}

.status-chip.is-completed { color: #25864a; background: rgba(220, 252, 231, 0.76); }
.status-chip.is-processing { color: #147a9e; background: rgba(207, 246, 255, 0.76); }
.status-chip.is-pending { color: #a06009; background: rgba(254, 243, 199, 0.78); }
.status-chip.is-failed { color: #c24155; background: rgba(255, 228, 230, 0.82); }

.detail-list {
  display: grid;
  gap: 10px;
  margin: 18px 0;
}

.detail-list div {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr);
  gap: 14px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(220, 230, 240, 0.78);
}

.detail-list dt,
.detail-list dd {
  margin: 0;
}

.detail-list dt {
  color: #7a8b9d;
  font-weight: 800;
}

.detail-list dd {
  overflow-wrap: anywhere;
  color: #26384c;
  font-weight: 650;
}

.detail-block {
  display: grid;
  gap: 10px;
  margin-top: 16px;
}

.detail-block h3 {
  margin: 0;
  color: #294156;
  font-size: 15px;
}

.detail-block pre {
  overflow: auto;
  max-height: 260px;
  margin: 0;
  padding: 14px;
  border: 1px solid rgba(219, 230, 240, 0.86);
  border-radius: 14px;
  background: rgba(247, 251, 255, 0.88);
  color: #31455a;
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
}

.detail-block.error pre {
  border-color: rgba(254, 205, 211, 0.96);
  background: rgba(255, 241, 242, 0.82);
  color: #be123c;
}
</style>

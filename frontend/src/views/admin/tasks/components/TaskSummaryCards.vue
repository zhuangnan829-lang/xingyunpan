<template>
  <div class="summary-area">
    <div class="metric-grid">
      <article v-for="metric in metrics" :key="metric.key" class="metric-card" :class="`tone-${metric.tone}`">
        <span>{{ metric.label }}</span>
        <strong>{{ metric.value }}</strong>
        <small>{{ metric.hint }}</small>
      </article>
    </div>

    <div class="queue-strip" :class="{ loading }">
      <article v-for="queue in queueCards" :key="queue.queue_key" class="queue-card">
        <div>
          <strong>{{ queue.label }}</strong>
          <span>{{ queue.submitted }} 个任务</span>
        </div>
        <div class="queue-bars" aria-hidden="true">
          <i :style="{ width: barWidth(queue.success, queue.submitted) }"></i>
          <i :style="{ width: barWidth(queue.processing, queue.submitted) }"></i>
          <i :style="{ width: barWidth(queue.failed, queue.submitted) }"></i>
        </div>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TaskMetric } from '../types';

defineProps<{
  metrics: TaskMetric[];
  queueCards: Array<{
    queue_key: string;
    label: string;
    submitted: number;
    success: number;
    processing: number;
    failed: number;
  }>;
  loading: boolean;
}>();

function barWidth(value: number, total: number): string {
  if (!total || value <= 0) return '3%';
  return `${Math.max(8, Math.round((value / total) * 100))}%`;
}
</script>

<style scoped>
.summary-area {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 16px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metric-card,
.queue-card {
  border: 1px solid rgba(255, 255, 255, 0.82);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.74), rgba(243, 250, 255, 0.56));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    0 16px 30px rgba(104, 151, 189, 0.1);
  backdrop-filter: blur(16px);
}

.metric-card {
  display: grid;
  gap: 10px;
  min-height: 116px;
  padding: 18px;
  border-radius: 20px;
}

.metric-card span,
.metric-card small,
.queue-card span {
  color: #6d7f91;
}

.metric-card span {
  font-size: 14px;
  font-weight: 760;
}

.metric-card strong {
  color: #1d3044;
  font-size: 34px;
  line-height: 1;
}

.metric-card small {
  font-size: 12px;
  font-weight: 650;
}

.tone-sky { box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.92), 0 16px 34px rgba(79, 184, 235, 0.16); }
.tone-mint { box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.92), 0 16px 34px rgba(47, 190, 159, 0.14); }
.tone-amber { box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.92), 0 16px 34px rgba(235, 179, 85, 0.16); }
.tone-rose { box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.92), 0 16px 34px rgba(230, 104, 132, 0.16); }

.queue-strip {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  opacity: 1;
  transition: opacity 0.2s ease;
}

.queue-strip.loading {
  opacity: 0.68;
}

.queue-card {
  display: grid;
  gap: 14px;
  padding: 14px;
  border-radius: 18px;
}

.queue-card strong {
  display: block;
  color: #26394e;
  font-size: 15px;
}

.queue-card span {
  display: block;
  margin-top: 6px;
  font-size: 12px;
}

.queue-bars {
  display: grid;
  gap: 5px;
}

.queue-bars i {
  display: block;
  height: 6px;
  max-width: 100%;
  border-radius: 999px;
}

.queue-bars i:nth-child(1) { background: linear-gradient(90deg, #5ac8fa, #7dd3fc); }
.queue-bars i:nth-child(2) { background: linear-gradient(90deg, #34d399, #99f6e4); }
.queue-bars i:nth-child(3) { background: linear-gradient(90deg, #fb7185, #fecdd3); }

@media (max-width: 1280px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .queue-strip {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .metric-grid,
  .queue-strip {
    grid-template-columns: 1fr;
  }
}
</style>

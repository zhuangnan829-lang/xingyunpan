<template>
  <section class="summary-grid" aria-label="分享概览">
    <article v-for="metric in metrics" :key="metric.label" class="summary-card" :class="`tone-${metric.tone}`">
      <span>{{ metric.label }}</span>
      <strong>{{ metric.value }}</strong>
      <small>{{ metric.detail }}</small>
    </article>
  </section>
</template>

<script setup lang="ts">
import type { ShareMetric } from '../types';

defineProps<{
  metrics: ShareMetric[];
}>();
</script>

<style scoped>
.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.summary-card {
  display: grid;
  gap: 8px;
  min-height: 128px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 20px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.76), rgba(255, 255, 255, 0.42));
  box-shadow: 0 18px 38px rgba(67, 93, 121, 0.11), inset 0 1px 0 rgba(255, 255, 255, 0.86);
  backdrop-filter: blur(16px);
}

.summary-card span {
  color: #60748a;
  font-size: 13px;
  font-weight: 760;
}

.summary-card strong {
  color: #132238;
  font-size: 34px;
  line-height: 1;
  font-weight: 860;
}

.summary-card small {
  color: #718096;
  font-size: 13px;
}

.tone-blue { color: #1687d9; }
.tone-cyan { color: #0891b2; }
.tone-coral { color: #fb7185; }
.tone-violet { color: #7c3aed; }

@media (max-width: 1160px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>

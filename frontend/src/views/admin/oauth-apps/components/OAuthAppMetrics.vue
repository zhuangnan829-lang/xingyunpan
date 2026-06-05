<template>
  <section class="metrics-grid" aria-label="OAuth 应用概览">
    <article v-for="metric in metrics" :key="metric.label" class="metric-card" :class="`tone-${metric.tone}`">
      <span>{{ metric.label }}</span>
      <strong>{{ metric.value }}</strong>
      <small>{{ metric.detail }}</small>
    </article>
  </section>
</template>

<script setup lang="ts">
import type { OAuthMetric } from '../types';

defineProps<{
  metrics: OAuthMetric[];
}>();
</script>

<style scoped>
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metric-card {
  display: grid;
  gap: 8px;
  min-height: 112px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.46);
  box-shadow: 0 18px 40px rgba(76, 113, 144, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.84);
  backdrop-filter: blur(16px);
}

.metric-card span {
  color: #6b7d91;
  font-size: 13px;
  font-weight: 800;
}

.metric-card strong {
  color: #162336;
  font-size: 30px;
  line-height: 1;
}

.metric-card small {
  color: #738397;
  font-size: 12px;
}

.tone-sky {
  background: linear-gradient(145deg, rgba(232, 247, 255, 0.78), rgba(255, 255, 255, 0.46));
}

.tone-mint {
  background: linear-gradient(145deg, rgba(225, 255, 248, 0.74), rgba(255, 255, 255, 0.46));
}

.tone-coral {
  background: linear-gradient(145deg, rgba(255, 235, 239, 0.78), rgba(255, 255, 255, 0.46));
}

.tone-violet {
  background: linear-gradient(145deg, rgba(241, 235, 255, 0.76), rgba(255, 255, 255, 0.46));
}

@media (max-width: 1100px) {
  .metrics-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }
}
</style>

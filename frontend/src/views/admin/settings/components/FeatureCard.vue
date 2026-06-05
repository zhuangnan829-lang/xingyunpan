<template>
  <article class="feature-card" :class="[`tone-${tone}`, { 'is-dark': dark }]">
    <div class="feature-head">
      <div class="feature-icon">
        <el-icon>
          <component :is="icon" />
        </el-icon>
      </div>

      <div>
        <p class="feature-eyebrow">{{ eyebrow }}</p>
        <h3 class="feature-title">{{ title }}</h3>
      </div>
    </div>

    <p class="feature-desc">{{ description }}</p>

    <div v-if="badges.length" class="badge-row">
      <span
        v-for="badge in badges"
        :key="badge.label"
        class="badge"
        :class="{ 'is-active': badge.active }"
      >
        {{ badge.label }}
      </span>
    </div>

    <div v-if="metrics.length" class="metric-stack">
      <div
        v-for="metric in metrics"
        :key="metric.label"
        class="metric-item"
      >
        <div class="metric-meta">
          <span>{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
        </div>
        <div class="metric-track">
          <div class="metric-bar" :style="{ width: metric.width }" />
        </div>
      </div>
    </div>

    <div v-if="items.length" class="item-grid" :class="{ compact }">
      <div
        v-for="item in items"
        :key="item.title"
        class="item-card"
      >
        <div class="item-icon">
          <el-icon>
            <component :is="item.icon" />
          </el-icon>
        </div>
        <div>
          <p class="item-title">{{ item.title }}</p>
          <p class="item-desc">{{ item.desc }}</p>
        </div>
      </div>
    </div>

    <div class="feature-content">
      <slot />
    </div>
  </article>
</template>

<script setup lang="ts">
import type { Component, PropType } from 'vue';
import type { FeatureCardBadge, FeatureCardItem, FeatureCardMetric, FeatureCardTone } from './media-center.types';

defineProps({
  eyebrow: {
    type: String,
    required: true,
  },
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    required: true,
  },
  icon: {
    type: Object as PropType<Component>,
    required: true,
  },
  tone: {
    type: String as PropType<FeatureCardTone>,
    default: 'sky',
  },
  badges: {
    type: Array as PropType<FeatureCardBadge[]>,
    default: () => [],
  },
  metrics: {
    type: Array as PropType<FeatureCardMetric[]>,
    default: () => [],
  },
  items: {
    type: Array as PropType<FeatureCardItem[]>,
    default: () => [],
  },
  compact: {
    type: Boolean,
    default: false,
  },
  dark: {
    type: Boolean,
    default: false,
  },
});
</script>

<style scoped>
.feature-card {
  display: grid;
  gap: 18px;
  padding: 24px;
  border: 1px solid rgba(226, 232, 240, 0.88);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 12px 36px rgba(15, 23, 42, 0.08);
  transition:
    transform 0.24s ease,
    box-shadow 0.24s ease;
}

.feature-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 22px 40px rgba(15, 23, 42, 0.1);
}

.feature-card.is-dark {
  color: #fff;
  border-color: rgba(255, 255, 255, 0.08);
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 45%, #334155 100%);
}

.feature-head {
  display: flex;
  align-items: center;
  gap: 14px;
}

.feature-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 18px;
  font-size: 20px;
}

.tone-sky .feature-icon {
  background: #e0f2fe;
  color: #0369a1;
}

.tone-violet .feature-icon {
  background: #ede9fe;
  color: #6d28d9;
}

.tone-amber .feature-icon {
  background: #fef3c7;
  color: #b45309;
}

.tone-emerald .feature-icon {
  background: #d1fae5;
  color: #047857;
}

.feature-card.is-dark .feature-icon {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}

.feature-eyebrow,
.feature-title,
.feature-desc,
.item-title,
.item-desc,
.metric-meta span,
.metric-meta strong {
  margin: 0;
}

.feature-eyebrow {
  color: #94a3b8;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
}

.feature-card.is-dark .feature-eyebrow {
  color: rgba(255, 255, 255, 0.56);
}

.feature-title {
  margin-top: 4px;
  color: #0f172a;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.feature-card.is-dark .feature-title {
  color: #fff;
}

.feature-desc {
  color: #475569;
  font-size: 14px;
  line-height: 1.85;
}

.feature-card.is-dark .feature-desc {
  color: rgba(255, 255, 255, 0.76);
}

.badge-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.badge {
  padding: 8px 12px;
  border-radius: 999px;
  background: #f8fafc;
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
}

.badge.is-active {
  background: #0f172a;
  color: #fff;
}

.feature-card.is-dark .badge {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.74);
}

.feature-card.is-dark .badge.is-active {
  background: #fff;
  color: #0f172a;
}

.metric-stack,
.feature-content {
  display: grid;
  gap: 14px;
}

.metric-item {
  display: grid;
  gap: 8px;
}

.metric-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: #64748b;
  font-size: 12px;
}

.metric-meta strong {
  color: #0f172a;
  font-weight: 700;
}

.feature-card.is-dark .metric-meta {
  color: rgba(255, 255, 255, 0.66);
}

.feature-card.is-dark .metric-meta strong {
  color: #fff;
}

.metric-track {
  height: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: #e2e8f0;
}

.feature-card.is-dark .metric-track {
  background: rgba(255, 255, 255, 0.12);
}

.metric-bar {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #8b5cf6, #0ea5e9);
}

.item-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.item-grid.compact {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.item-card {
  display: flex;
  gap: 12px;
  padding: 16px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  border-radius: 20px;
  background: #fff;
}

.item-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 14px;
  background: #f8fafc;
  color: #334155;
  font-size: 18px;
}

.item-title {
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
}

.item-desc {
  margin-top: 6px;
  color: #64748b;
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 960px) {
  .item-grid,
  .item-grid.compact {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .feature-card {
    padding: 20px;
  }

  .feature-title {
    font-size: 22px;
  }
}
</style>

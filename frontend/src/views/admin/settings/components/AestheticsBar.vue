<template>
  <header class="aesthetics-bar">
    <div class="copy-block">
      <p class="eyebrow">{{ eyebrow }}</p>
      <h2>{{ title }}</h2>
      <p class="description">{{ description }}</p>
    </div>

    <div class="engine-grid">
      <article
        v-for="engine in engines"
        :key="engine.name"
        class="engine-card"
      >
        <div class="engine-head">
          <div>
            <p class="engine-name">{{ engine.name }}</p>
            <p class="engine-label">{{ engine.label }}</p>
          </div>
          <span
            class="engine-dot"
            :class="{ 'is-ready': engine.ready, 'is-waiting': !engine.ready }"
          />
        </div>

        <div class="engine-track">
          <div
            class="engine-progress"
            :class="{ 'is-ready': engine.ready, 'is-waiting': !engine.ready }"
            :style="{ width: `${engine.progress}%` }"
          />
        </div>
      </article>
    </div>
  </header>
</template>

<script setup lang="ts">
import type { PropType } from 'vue';
import type { EngineStatus } from './media-center.types';

defineProps({
  eyebrow: {
    type: String,
    default: 'Nebula Drive V2',
  },
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    required: true,
  },
  engines: {
    type: Array as PropType<EngineStatus[]>,
    required: true,
  },
});
</script>

<style scoped>
.aesthetics-bar {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.9fr);
  gap: 24px;
  align-items: center;
  padding: 30px;
  border: 1px solid rgba(226, 232, 240, 0.85);
  border-radius: 32px;
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.14), transparent 34%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(255, 255, 255, 0.9));
  box-shadow: 0 24px 80px rgba(15, 23, 42, 0.08);
}

.copy-block,
.engine-grid {
  display: grid;
  gap: 16px;
}

.eyebrow,
.engine-name,
.engine-label,
.description,
h2 {
  margin: 0;
}

.eyebrow {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  padding: 7px 12px;
  border: 1px solid rgba(186, 230, 253, 0.9);
  border-radius: 999px;
  background: rgba(240, 249, 255, 0.9);
  color: #0369a1;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.copy-block h2 {
  font-size: 42px;
  line-height: 1.08;
  letter-spacing: -0.04em;
  color: #0f172a;
}

.description {
  max-width: 680px;
  color: #475569;
  font-size: 16px;
  line-height: 1.9;
}

.engine-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.engine-card {
  padding: 18px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
  transition:
    transform 0.24s ease,
    box-shadow 0.24s ease;
}

.engine-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 16px 32px rgba(15, 23, 42, 0.08);
}

.engine-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.engine-name {
  color: #0f172a;
  font-size: 15px;
  font-weight: 700;
}

.engine-label {
  margin-top: 6px;
  color: #64748b;
  font-size: 12px;
}

.engine-dot {
  width: 12px;
  height: 12px;
  margin-top: 4px;
  border-radius: 50%;
}

.engine-dot.is-ready {
  background: #10b981;
  box-shadow: 0 0 0 6px rgba(16, 185, 129, 0.14);
}

.engine-dot.is-waiting {
  background: #f59e0b;
  box-shadow: 0 0 0 6px rgba(245, 158, 11, 0.16);
}

.engine-track {
  height: 8px;
  margin-top: 16px;
  overflow: hidden;
  border-radius: 999px;
  background: #e2e8f0;
}

.engine-progress {
  height: 100%;
  border-radius: inherit;
}

.engine-progress.is-ready {
  background: linear-gradient(90deg, #10b981, #34d399);
}

.engine-progress.is-waiting {
  background: linear-gradient(90deg, #f59e0b, #fbbf24);
}

@media (max-width: 1080px) {
  .aesthetics-bar {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .aesthetics-bar {
    padding: 22px;
  }

  .copy-block h2 {
    font-size: 34px;
  }

  .engine-grid {
    grid-template-columns: 1fr;
  }
}
</style>

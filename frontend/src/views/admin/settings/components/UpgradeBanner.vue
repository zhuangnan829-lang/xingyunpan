<template>
  <footer class="upgrade-banner">
    <div class="copy-block">
      <p class="eyebrow">{{ eyebrow }}</p>
      <h3>{{ title }}</h3>
      <p class="description">{{ description }}</p>
    </div>

    <div class="action-group">
      <button type="button" class="primary-button">
        {{ primaryAction }}
      </button>
      <button type="button" class="secondary-button">
        {{ secondaryAction }}
      </button>
    </div>
  </footer>
</template>

<script setup lang="ts">
defineProps({
  eyebrow: {
    type: String,
    default: 'Professional Upgrade',
  },
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    required: true,
  },
  primaryAction: {
    type: String,
    default: '立即升级',
  },
  secondaryAction: {
    type: String,
    default: '查看媒体能力',
  },
});
</script>

<style scoped>
.upgrade-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 28px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 30px;
  background: linear-gradient(135deg, #0f172a 0%, #111827 50%, #1e293b 100%);
  box-shadow: 0 20px 60px rgba(15, 23, 42, 0.2);
}

.copy-block {
  display: grid;
  gap: 10px;
  max-width: 760px;
}

.eyebrow,
.description,
h3 {
  margin: 0;
}

.eyebrow {
  color: #7dd3fc;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

h3 {
  color: #fff;
  font-size: 30px;
  line-height: 1.18;
  letter-spacing: -0.03em;
}

.description {
  color: rgba(226, 232, 240, 0.82);
  font-size: 14px;
  line-height: 1.9;
}

.action-group {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.primary-button,
.secondary-button {
  min-height: 48px;
  padding: 0 18px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    opacity 0.2s ease,
    background 0.2s ease;
}

.primary-button {
  border: none;
  background: #fff;
  color: #0f172a;
}

.secondary-button {
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.92);
}

.primary-button:hover,
.secondary-button:hover {
  transform: translateY(-1px);
}

@media (max-width: 900px) {
  .upgrade-banner {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 720px) {
  .upgrade-banner {
    padding: 22px;
  }

  h3 {
    font-size: 24px;
  }
}
</style>

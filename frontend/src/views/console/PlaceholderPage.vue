<template>
  <section class="placeholder-page" :class="{ 'is-admin': section === 'Admin' }">
    <div class="placeholder-card">
      <p class="eyebrow">{{ section === 'Admin' ? 'Admin Console' : 'Drive Console' }}</p>
      <h1 class="title">{{ title }}</h1>
      <p class="description">{{ description }}</p>

      <div class="status-panel">
        <span class="status-label">页面骨架已就绪</span>
        <p>这个入口已经放进新的控制台层级里了，后续可以继续在这里接入真实的数据表格、筛选器、操作按钮和后端接口。</p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  section: string;
  title: string;
  description: string;
}>();
</script>

<style scoped>
.placeholder-page {
  min-height: calc(100vh - 96px);
}

.placeholder-card {
  padding: 30px;
  border: 1px solid #dfe4ea;
  border-radius: 20px;
  background: #ffffff;
}

.eyebrow {
  margin: 0 0 10px;
  color: #4f8fe9;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.title {
  margin: 0 0 14px;
  color: #1f2937;
  font-size: 42px;
  line-height: 1.08;
}

.description {
  max-width: 780px;
  margin: 0;
  color: #6b7280;
  font-size: 16px;
  line-height: 1.85;
}

.status-panel {
  margin-top: 30px;
  padding: 22px 24px;
  border: 1px solid #e5e7eb;
  border-radius: 18px;
  background: #f9fafb;
}

.status-label {
  display: inline-flex;
  margin-bottom: 10px;
  color: #2f7ed8;
  font-size: 13px;
  font-weight: 700;
}

.status-panel p {
  margin: 0;
  color: #4b5563;
  line-height: 1.8;
}

.placeholder-page.is-admin .eyebrow {
  color: #6b7280;
}

@media (max-width: 768px) {
  .placeholder-card {
    padding: 18px;
  }

  .title {
    font-size: 30px;
  }
}
</style>

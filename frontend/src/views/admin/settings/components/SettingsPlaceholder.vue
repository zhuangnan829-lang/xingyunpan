<template>
  <div class="settings-placeholder">
    <section class="panel-card">
      <p class="eyebrow">{{ eyebrow }}</p>
      <h2>{{ title }}</h2>
      <p class="description">{{ description }}</p>
    </section>

    <section class="panel-card">
      <p class="eyebrow">Coming Soon</p>
      <div class="placeholder-box">
        <strong>{{ title }}</strong>
        <span>{{ hint }}</span>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';

const props = withDefaults(
  defineProps<{
    eyebrow?: string;
    title: string;
    description: string;
    hint?: string;
  }>(),
  {
    eyebrow: 'Settings Module',
    hint: '该面板已完成占位，后续可直接补充真实表单、接口与校验逻辑。',
  },
);

const loading = ref(false);
const saving = ref(false);
const isDirty = computed(() => false);

async function reload() {
  ElMessage.info(`“${props.title}”暂未接入重新加载逻辑`);
}

function reset() {
  ElMessage.info(`“${props.title}”暂未接入恢复默认逻辑`);
}

async function save() {
  ElMessage.info(`“${props.title}”暂未接入保存逻辑`);
}

defineExpose({
  isDirty,
  loading,
  saving,
  reload,
  reset,
  save,
});
</script>

<style scoped>
.settings-placeholder {
  display: grid;
  gap: 20px;
}

.panel-card,
.placeholder-box {
  border: 1px solid rgba(226, 232, 240, 0.95);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.05);
}

.panel-card {
  padding: 24px;
  border-radius: 24px;
}

.placeholder-box {
  display: grid;
  gap: 8px;
  padding: 20px;
  border-radius: 20px;
  background: linear-gradient(180deg, #fff 0%, #f8fbff 100%);
}

.eyebrow {
  margin: 0 0 10px;
  color: #2563eb;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

h2,
p,
strong,
span {
  margin: 0;
}

h2,
strong {
  color: #0f172a;
}

.description,
.placeholder-box span {
  color: #64748b;
  line-height: 1.8;
}

@media (max-width: 720px) {
  .panel-card {
    padding: 18px;
  }
}
</style>
